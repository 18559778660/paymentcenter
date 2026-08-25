package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringList JSON 字符串数组，存 MySQL JSON 列。
type StringList []string

func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return errors.New("invalid StringList value")
	}
	if len(raw) == 0 {
		*s = StringList{}
		return nil
	}
	out := StringList{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	*s = out
	return nil
}
