package db

import (
	"bufio"
	"mikan/pkg/logger"
	"os"
	"reflect"
	"strings"
	"unicode"
)

type Table struct {
	wal  *os.File
	data []Keyable
	typ  reflect.Type
}

func getTableName(data Keyable) string {
	if tn, ok := data.(TableNamer); ok {
		return tn.TableName()
	}
	return convertToSnakeCase(reflect.TypeOf(data).Elem().Name())
}

func convertToSnakeCase(name string) string {
	var result strings.Builder
	for i, char := range name {
		if unicode.IsUpper(char) {
			if i != 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(char))
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

func (t *Table) writeWAL(opType, payload string) error {
	_, err := t.wal.Write([]byte(opType + "," + payload + "\n"))
	if err != nil {
		return err
	}
	return t.wal.Sync()
}

func (t *Table) Recover(typ reflect.Type) {
	s := bufio.NewScanner(t.wal)
	logger.Info("Recovering data...")
	for s.Scan() {
		line := s.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			logger.Error("Invalid WAL entry: %s", line)
			continue
		}

		opType := parts[0]
		payload := parts[1]

		switch opType {
		case "ADD":
			data := reflect.New(typ).Interface().(Keyable)
			if err := data.From(payload); err != nil {
				logger.Error("Error parsing string: %v", err)
				continue
			}
			t.data = append(t.data, data)
		case "UPDATE":
			data := reflect.New(typ).Interface().(Keyable)
			if err := data.From(payload); err != nil {
				logger.Error("Error parsing string: %v", err)
				continue
			}
			for i, v := range t.data {
				if v.GetKey() == data.GetKey() {
					t.data[i] = data
					break
				}
			}
		case "DELETE":
			key := payload
			for i, v := range t.data {
				if v.GetKey() == key {
					t.data = append(t.data[:i], t.data[i+1:]...)
					break
				}
			}
		default:
			logger.Error("Unknown WAL operation type: %s", opType)
		}
	}
	logger.Info("Recovered data")
}
