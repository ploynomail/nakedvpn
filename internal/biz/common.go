package biz

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type StringArr []string

// 实现 driver.Valuer 接口
func (s StringArr) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// 实现 sql.Scanner 接口
func (s *StringArr) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	if len(bytes) > 0 {
		return json.Unmarshal([]byte(bytes), s)
	}
	*s = make([]string, 0)
	return nil
}

// 实现 driver.Valuer 接口
func (s *QuotaConfig) Value() (driver.Value, error) {
	if s == nil {
		return "{}", nil
	}
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// 实现 sql.Scanner 接口
func (s *QuotaConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	if len(bytes) > 0 {
		return json.Unmarshal([]byte(bytes), s)
	}
	s = new(QuotaConfig)
	return nil
}

type MapAny map[string]any

// 实现 driver.Valuer 接口
func (s *MapAny) Value() (driver.Value, error) {
	if s == nil {
		return "{}", nil
	}
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// 实现 sql.Scanner 接口
func (s *MapAny) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	if len(bytes) > 0 {
		return json.Unmarshal([]byte(bytes), s)
	}
	s = new(MapAny)
	return nil
}
