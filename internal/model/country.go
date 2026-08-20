package model

import "time"

// Country 国家列表。code 为 2 位编码，name 为展示名称。
type Country struct {
	ID           uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code         string    `gorm:"column:code;type:varchar(8);not null;uniqueIndex" json:"code"`
	Name         string    `gorm:"column:name;type:varchar(128);not null;index" json:"name"`
	CardBinRatio float64   `gorm:"column:card_bin_ratio;type:decimal(8,2);not null" json:"card_bin_ratio"`
	CreatedBy    string    `gorm:"column:created_by;type:varchar(64);not null;default:''" json:"created_by"`
	UpdatedBy    string    `gorm:"column:updated_by;type:varchar(64);not null;default:''" json:"updated_by"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Country) TableName() string {
	return "countries"
}
