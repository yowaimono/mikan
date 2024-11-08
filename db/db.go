package db

import (
	"fmt"
	"mikan/pkg/logger"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Keyable interface {
	GetKey() string
	SetKey(key string)
	String() string
	From(s string) error
}

type TableNamer interface {
	TableName() string
}

type MinDB struct {
	tables      map[string]*Table
	types       map[string]reflect.Type
	constraints map[string]map[string]struct{}
	defaults    map[string]map[string]interface{}
	mu          sync.Mutex
	cfg         *Config
}

var (
	DB   *MinDB
	once sync.Once
)

func GetDB(cfg ...*Config) *MinDB {
	once.Do(func() {
		if len(cfg) == 0 {
			if err := Init(); err != nil {
				logger.Error("init db error %v", err)
			}
		} else {
			if err := Init(cfg...); err != nil {
				logger.Error("init db error %v", err)
			}
		}

	})
	return DB
}

// Init 初始化数据库
func Init(cfg ...*Config) error {
	if len(cfg) == 0 {
		DB = &MinDB{
			cfg: &Config{
				LogLevel: "INFO",
			},
			tables:      make(map[string]*Table),
			types:       make(map[string]reflect.Type),
			constraints: make(map[string]map[string]struct{}),
			defaults:    make(map[string]map[string]interface{}),
		}
	} else {
		DB = &MinDB{
			cfg:         cfg[0],
			tables:      make(map[string]*Table),
			types:       make(map[string]reflect.Type),
			constraints: make(map[string]map[string]struct{}),
			defaults:    make(map[string]map[string]interface{}),
		}
	}

	logger.GetLogger().SetLevel(DB.cfg.LogLevel) // 设置日志等级

	return nil
}
func (m *MinDB) AutoCreate(data Keyable) {
	tableName := getTableName(data)
	m.types[tableName] = reflect.TypeOf(data).Elem()

	typ := reflect.TypeOf(data).Elem()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("min")
		if strings.Contains(tag, "unique") {
			if m.constraints[tableName] == nil {
				m.constraints[tableName] = make(map[string]struct{})
			}
			m.constraints[tableName][field.Name] = struct{}{}
		}
		if strings.Contains(tag, "default") {
			if m.defaults[tableName] == nil {
				m.defaults[tableName] = make(map[string]interface{})
			}
			defaultValue := strings.Split(tag, "default ")[1]
			switch field.Type.Kind() {
			case reflect.String:
				m.defaults[tableName][field.Name] = defaultValue
			case reflect.Int:
				intValue, err := strconv.Atoi(defaultValue)
				if err != nil {
					logger.Error("invalid default value for int field: %s", defaultValue)
					continue
				}
				m.defaults[tableName][field.Name] = intValue
			default:
				logger.Error("unsupported field type for default value: %s", field.Type.Kind())
			}
		}
	}
}

func (m *MinDB) getOrCreateTable(tableName string) (*Table, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if table, exists := m.tables[tableName]; exists {
		return table, nil
	}

	file, err := os.OpenFile(tableName+".wal", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("create file error: %w", err)
	}

	table := &Table{wal: file, typ: m.types[tableName]}
	m.tables[tableName] = table
	table.Recover(m.types[tableName])
	return table, nil
}

func (m *MinDB) Create(data Keyable) error {
	tableName := getTableName(data)
	table, err := m.getOrCreateTable(tableName)
	if err != nil {
		return err
	}

	if constraints, exists := m.constraints[tableName]; exists {
		for field := range constraints {
			fieldValue := reflect.ValueOf(data).Elem().FieldByName(field).Interface()
			for _, existingData := range table.data {
				if reflect.DeepEqual(reflect.ValueOf(existingData).Elem().FieldByName(field).Interface(), fieldValue) {
					return fmt.Errorf("unique constraint violation for field %s", field)
				}
			}
		}
	}

	if defaults, exists := m.defaults[tableName]; exists {
		for field, defaultValue := range defaults {
			fieldValue := reflect.ValueOf(data).Elem().FieldByName(field)
			if fieldValue.IsValid() && fieldValue.IsZero() {
				fieldValue.Set(reflect.ValueOf(defaultValue))
			}
		}
	}

	// 调用hooks的BeforeAdd方法
	if beforeAdd, ok := data.(BeforeCreate); ok {
		logger.Debug("beforeAdd for %s ,record is %v", tableName, data)

		if err := beforeAdd.BeforeCreate(data); err != nil {
			return err
		}

	}

	key := generateUniqueKey()
	data.SetKey(key)

	if err := table.writeWAL("ADD", data.String()); err != nil {
		return err
	}

	table.data = append(table.data, data)

	if afterAdd, ok := data.(AfterCreate); ok {
		logger.Debug("afterAdd for %s,record is %v", tableName, data)
		if err := afterAdd.AfterCreate(data); err != nil {
			return err
		}
	}

	return nil
}
func generateUniqueKey() string {
	return uuid.New().String()
}

func (m *MinDB) Delete(data Keyable) error {
	tableName := getTableName(data)
	table, err := m.getOrCreateTable(tableName)
	if err != nil {
		return err
	}

	// call the hooks

	if beforeDelete, ok := data.(BeforeDelete); ok {
		logger.Debug("beforeDelete for %s,record is %v", tableName, data)

		if err := beforeDelete.BeforeDelete(data); err != nil {
			return err
		}
	}

	key := data.GetKey()
	for i, v := range table.data {
		if v.GetKey() == key {
			table.data = append(table.data[:i], table.data[i+1:]...)
			if err := table.writeWAL("DELETE", key); err != nil {
				return err
			}

			if afterDelete, ok := data.(AfterDelete); ok {
				logger.Debug("afterDelete for %s,record is %v", tableName, data)
				if err := afterDelete.AfterDelete(data); err != nil {
					return err
				}
			}

			return nil
		}
	}

	return nil
}

func (m *MinDB) GetAll(data Keyable) ([]Keyable, error) {
	tableName := getTableName(data)
	table, err := m.getOrCreateTable(tableName)
	if err != nil {
		return nil, err
	}
	return table.data, nil
}

func (m *MinDB) GetByKey(data Keyable) (Keyable, error) {
	tableName := getTableName(data)
	table, err := m.getOrCreateTable(tableName)
	if err != nil {
		return nil, err
	}

	for _, v := range table.data {
		if v.GetKey() == data.GetKey() {
			return v, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *MinDB) Update(data Keyable) error {
	tableName := getTableName(data)
	table, err := m.getOrCreateTable(tableName)
	if err != nil {
		return err
	}

	// call the hooks

	if beforeUpdate, ok := data.(BeforeUpdate); ok {
		logger.Debug("beforeUpdate for %s,record is %v", tableName, data)
		if err := beforeUpdate.BeforeUpdate(data); err != nil {
			return err
		}
	}

	key := data.GetKey()
	for i, v := range table.data {
		if v.GetKey() == key {
			table.data[i] = data
			if err := table.writeWAL("UPDATE", data.String()); err != nil {
				return err
			}

			if afterUpdate, ok := data.(AfterUpdate); ok {
				logger.Debug("afterUpdate for %s,record is %v", tableName, data)
				if err := afterUpdate.AfterUpdate(data); err != nil {
					return err
				}
			}
			return nil
		}
	}

	return fmt.Errorf("not found")
}

func (m *MinDB) Close() error {
	for _, table := range m.tables {
		if table.wal != nil {
			if err := table.wal.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
