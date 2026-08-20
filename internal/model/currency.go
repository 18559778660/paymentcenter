package model

import "time"

// Currency 货币列表。code 为 ISO 编码，name 为展示名称。
type Currency struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"column:code;type:varchar(16);not null;uniqueIndex" json:"code"`
	Name      string    `gorm:"column:name;type:varchar(64);not null;index" json:"name"`
	Rate      float64   `gorm:"column:rate;type:decimal(18,8);not null" json:"rate"`
	CreatedBy string    `gorm:"column:created_by;type:varchar(64);not null;default:''" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by;type:varchar(64);not null;default:''" json:"updated_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Currency) TableName() string {
	return "currencies"
}
