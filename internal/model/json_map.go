package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONMap 存任意前端 meta 字段，例如 hideInMenu、affixTab、keepAlive。
type JSONMap map[string]interface{}

// Value 写入数据库。
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan 从数据库读出。
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported meta type: %T", value)
	}
	if len(raw) == 0 {
		*m = JSONMap{}
		return nil
	}
	return json.Unmarshal(raw, m)
}
