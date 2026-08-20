package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// CardBinPrefix 卡头规则：只填起始，或填起始到结束的一段。
type CardBinPrefix struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CardBinLengths 卡号允许位数，例如 [16,17,18,19]。
type CardBinLengths []int

func (l CardBinLengths) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (l *CardBinLengths) Scan(value interface{}) error {
	return scanJSON(value, l, CardBinLengths{})
}

// CardBinPrefixes 卡头规则列表。
type CardBinPrefixes []CardBinPrefix

func (p CardBinPrefixes) Value() (driver.Value, error) {
	if p == nil {
		return "[]", nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (p *CardBinPrefixes) Scan(value interface{}) error {
	return scanJSON(value, p, CardBinPrefixes{})
}

func scanJSON[T any](value interface{}, dest *T, empty T) error {
	if value == nil {
		*dest = empty
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported json type: %T", value)
	}
	if len(raw) == 0 {
		*dest = empty
		return nil
	}
	return json.Unmarshal(raw, dest)
}

// CardType 卡类型 / 卡头验证。name 存品牌编码，例如 unionpay。
type CardType struct {
	ID        uint            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code      string          `gorm:"column:code;type:varchar(32);not null;uniqueIndex" json:"code"`
	Name      string          `gorm:"column:name;type:varchar(32);not null;index" json:"name"`
	Lengths   CardBinLengths  `gorm:"column:lengths;type:json;not null" json:"lengths"`
	Prefixes  CardBinPrefixes `gorm:"column:prefixes;type:json;not null" json:"prefixes"`
	CreatedBy string          `gorm:"column:created_by;type:varchar(64);not null;default:''" json:"created_by"`
	UpdatedBy string          `gorm:"column:updated_by;type:varchar(64);not null;default:''" json:"updated_by"`
	CreatedAt time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at" json:"updated_at"`
}

func (CardType) TableName() string {
	return "card_types"
}
