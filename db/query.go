package db

import (
	"fmt"
	"mikan/pkg/logger"

	"reflect"
	"regexp"
	"sort"
	"strings"
)

type Queryable interface {
	Where(field string, value interface{}) Queryable
	OrderBy(field string, desc bool) Queryable
	Limit(limit int) Queryable
	GroupBy(field string) Queryable
	Having(field string, value interface{}) Queryable
	Join(table string, on string) Queryable
	Distinct() Queryable
	Count() (int, error)
	Sum(field string) (int, error)
	Avg(field string) (float64, error)
	Max(field string) (interface{}, error)
	Min(field string) (interface{}, error)
	Find() Finder
	First(out Keyable) error
	Like(field string, pattern string) Queryable
	Update(field string, value interface{}, conditionFields ...interface{}) error
	Delete(conditionFields ...interface{}) error
	Select(fields ...string) Queryable
	In(field string, values ...interface{}) Queryable
	IsNull(field string) Queryable
	IsNotNull(field string) Queryable
}

type Query struct {
	table    *Table
	filter   func(Keyable) bool
	sortBy   func(i, j Keyable) bool
	limit    int
	groupBy  string
	having   func(Keyable) bool
	join     *Join
	distinct bool
	fields   []string
}

type Result struct {
	Data  Keyable
	Error error
}

func (m *MinDB) Query(data Keyable) Queryable {
	tableName := getTableName(data)
	table, err := m.getOrCreateTable(tableName)
	if err != nil {
		logger.Error("get or create table error %v", err)
		return nil
	}
	return &Query{table: table, filter: func(data Keyable) bool { return true }}
}

func (q *Query) Where(field string, value interface{}) Queryable {
	q.filter = func(data Keyable) bool {
		v := reflect.ValueOf(data).Elem().FieldByName(field)
		return v.IsValid() && v.Interface() == value
	}
	return q
}

func (q *Query) OrderBy(field string, desc bool) Queryable {
	q.sortBy = func(i, j Keyable) bool {
		vi := reflect.ValueOf(i).Elem().FieldByName(field)
		vj := reflect.ValueOf(j).Elem().FieldByName(field)
		if !vi.IsValid() || !vj.IsValid() {
			return false
		}
		if desc {
			return vi.Interface().(int) > vj.Interface().(int)
		}
		return vi.Interface().(int) < vj.Interface().(int)
	}
	return q
}

func (q *Query) Limit(limit int) Queryable {
	q.limit = limit
	return q
}

func (q *Query) GroupBy(field string) Queryable {
	q.groupBy = field
	return q
}

func (q *Query) Having(field string, value interface{}) Queryable {
	q.having = func(data Keyable) bool {
		v := reflect.ValueOf(data).Elem().FieldByName(field)
		return v.IsValid() && v.Interface() == value
	}
	return q
}

func (q *Query) Join(table string, on string) Queryable {
	joinTable, err := DB.getOrCreateTable(table)
	if err != nil {
		logger.Error("get or create table error %v", err)
		return nil
	}
	q.join = &Join{table: joinTable, on: on}
	return q
}

func (q *Query) Distinct() Queryable {
	q.distinct = true
	return q
}

func (q *Query) First(out Keyable) error {
	for _, data := range q.table.data {
		if q.filter == nil || q.filter(data) {
			return out.From(data.String())
		}
	}
	return fmt.Errorf("not found")
}

func (q *Query) Count() (int, error) {
	results, err := q.Find().List()
	if err != nil {
		return 0, err
	}
	return len(results), nil
}

func (q *Query) Sum(field string) (int, error) {
	res, err := aggregate(q, field, func(sum interface{}, v int) interface{} {
		if sum == nil {
			return v
		}
		return sum.(int) + v
	})
	if err != nil {
		return 0, err
	}
	return res.(int), nil
}

func (q *Query) Avg(field string) (float64, error) {
	results, err := q.Find().List()
	if err != nil {
		return 0, err
	}
	var sum int
	for _, data := range results {
		v := getFieldValue(data, field)
		if v == nil {
			return 0, fmt.Errorf("field not found")
		}
		sum += v.(int)
	}
	return float64(sum) / float64(len(results)), nil
}

func (q *Query) Max(field string) (interface{}, error) {
	return aggregate(q, field, func(max interface{}, v int) interface{} {
		if max == nil || v > max.(int) {
			return v
		}
		return max
	})
}

func (q *Query) Min(field string) (interface{}, error) {
	return aggregate(q, field, func(min interface{}, v int) interface{} {
		if min == nil || v < min.(int) {
			return v
		}
		return min
	})
}

func (q *Query) Find() Finder {
	results, err := q.List()
	return Finder{query: q, results: results, err: err}
}

func (q *Query) Select(fields ...string) Queryable {
	q.fields = fields
	return q
}

func (q *Query) In(field string, values ...interface{}) Queryable {
	q.filter = func(data Keyable) bool {
		v := reflect.ValueOf(data).Elem().FieldByName(field)
		if !v.IsValid() {
			return false
		}
		for _, value := range values {
			if v.Interface() == value {
				return true
			}
		}
		return false
	}
	return q
}

func (q *Query) IsNull(field string) Queryable {
	q.filter = func(data Keyable) bool {
		v := reflect.ValueOf(data).Elem().FieldByName(field)
		if !v.IsValid() {
			return false
		}
		return v.IsZero()
	}
	return q
}

func (q *Query) IsNotNull(field string) Queryable {
	q.filter = func(data Keyable) bool {
		v := reflect.ValueOf(data).Elem().FieldByName(field)
		if !v.IsValid() {
			return false
		}
		return !v.IsZero()
	}
	return q
}

func (q *Query) Like(field string, pattern string) Queryable {
	regexPattern := convertToRegex(pattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		logger.Error("invalid pattern for Like: %v", err)
		return q
	}
	q.filter = func(data Keyable) bool {
		v := reflect.ValueOf(data).Elem().FieldByName(field)
		if !v.IsValid() {
			return false
		}
		return re.MatchString(v.String())
	}
	return q
}

func (q *Query) Update(field string, value interface{}, conditionFields ...interface{}) error {
	tableName := getTableName(q.table.data[0])
	table, err := DB.getOrCreateTable(tableName)
	if err != nil {
		return err
	}

	filter := q.filter
	if len(conditionFields) > 0 {
		if len(conditionFields)%2 != 0 {
			return fmt.Errorf("condition fields must be in pairs")
		}
		filter = func(data Keyable) bool {
			for i := 0; i < len(conditionFields); i += 2 {
				condField := conditionFields[i].(string)
				condValue := conditionFields[i+1]
				v := reflect.ValueOf(data).Elem().FieldByName(condField)
				if !v.IsValid() || v.Interface() != condValue {
					return false
				}
			}
			return true
		}
	}

	for i, data := range table.data {
		if filter(data) {
			reflect.ValueOf(data).Elem().FieldByName(field).Set(reflect.ValueOf(value))
			if err := table.writeWAL("UPDATE", data.String()); err != nil {
				return err
			}
			table.data[i] = data
		}
	}
	return nil
}

func (q *Query) Delete(conditionFields ...interface{}) error {
	tableName := getTableName(q.table.data[0])
	table, err := DB.getOrCreateTable(tableName)
	if err != nil {
		return err
	}

	filter := q.filter
	if len(conditionFields) > 0 {
		if len(conditionFields)%2 != 0 {
			return fmt.Errorf("condition fields must be in pairs")
		}
		filter = func(data Keyable) bool {
			for i := 0; i < len(conditionFields); i += 2 {
				condField := conditionFields[i].(string)
				condValue := conditionFields[i+1]
				v := reflect.ValueOf(data).Elem().FieldByName(condField)
				if !v.IsValid() || v.Interface() != condValue {
					return false
				}
			}
			return true
		}
	}

	newData := make([]Keyable, 0, len(table.data))
	for _, data := range table.data {
		if !filter(data) {
			newData = append(newData, data)
		} else {
			if err := table.writeWAL("DELETE", data.GetKey()); err != nil {
				return err
			}
		}
	}
	table.data = newData
	return nil
}

func (q *Query) List() ([]Keyable, error) {
	var result []Keyable
	for _, data := range q.table.data {
		if q.filter == nil || q.filter(data) {
			result = append(result, data)
		}
	}

	if q.sortBy != nil {
		sort.Slice(result, func(i, j int) bool {
			return q.sortBy(result[i], result[j])
		})
	}

	if q.limit > 0 && len(result) > q.limit {
		result = result[:q.limit]
	}

	if q.groupBy != "" {
		grouped := make(map[string][]Keyable)
		for _, data := range result {
			v := getFieldValue(data, q.groupBy)
			if v == nil {
				return nil, fmt.Errorf("field not found")
			}
			key := v.(string)
			grouped[key] = append(grouped[key], data)
		}
		result = nil
		for _, group := range grouped {
			if q.having == nil || q.having(group[0]) {
				result = append(result, group...)
			}
		}
	}

	if q.distinct {
		seen := make(map[string]bool)
		var distinctResult []Keyable
		for _, data := range result {
			key := data.GetKey()
			if !seen[key] {
				seen[key] = true
				distinctResult = append(distinctResult, data)
			}
		}
		result = distinctResult
	}

	if q.join != nil {
		var joinedResult []Keyable
		for _, data := range result {
			for _, joinData := range q.join.table.data {
				if q.join.on == "" || reflect.ValueOf(data).Elem().FieldByName(q.join.on).Interface() == reflect.ValueOf(joinData).Elem().FieldByName(q.join.on).Interface() {
					joinedResult = append(joinedResult, data)
				}
			}
		}
		result = joinedResult
	}

	if len(q.fields) > 0 {
		var selectedResult []Keyable
		for _, data := range result {
			newData := reflect.New(q.table.typ).Interface().(Keyable)
			for _, field := range q.fields {
				v := reflect.ValueOf(data).Elem().FieldByName(field)
				if v.IsValid() {
					reflect.ValueOf(newData).Elem().FieldByName(field).Set(v)
				}
			}
			selectedResult = append(selectedResult, newData)
		}
		result = selectedResult
	}

	return result, nil
}

func aggregate(q *Query, field string, fn func(interface{}, int) interface{}) (interface{}, error) {
	results, err := q.Find().List()
	if err != nil {
		return nil, err
	}
	var result interface{}
	for _, data := range results {
		v := getFieldValue(data, field)
		if v == nil {
			return nil, fmt.Errorf("field not found")
		}
		result = fn(result, v.(int))
	}
	return result, nil
}

func getFieldValue(data Keyable, field string) interface{} {
	v := reflect.ValueOf(data).Elem().FieldByName(field)
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}

func convertToRegex(pattern string) string {
	regexPattern := strings.ReplaceAll(pattern, "%", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "_", ".")
	return regexPattern
}

type Finder struct {
	query   *Query
	results []Keyable
	err     error
}

func (f Finder) List() ([]Keyable, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.results, nil
}

func (f Finder) Unwrap(out interface{}) error {
	if f.err != nil {
		return f.err
	}

	outValue := reflect.ValueOf(out)
	if outValue.Kind() != reflect.Ptr || outValue.IsNil() {
		return fmt.Errorf("out must be a non-nil pointer to a slice")
	}

	outValue = outValue.Elem()
	if outValue.Kind() != reflect.Slice {
		return fmt.Errorf("out must be a pointer to a slice")
	}

	outValue.Set(reflect.MakeSlice(outValue.Type(), 0, 0))

	for _, result := range f.results {
		newItem := reflect.New(f.query.table.typ).Interface().(Keyable)
		if err := newItem.From(result.String()); err != nil {
			return err
		}
		outValue.Set(reflect.Append(outValue, reflect.ValueOf(newItem).Elem()))
	}

	return nil
}

type Join struct {
	table *Table
	on    string
}
