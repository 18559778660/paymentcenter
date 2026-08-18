package model

import "time"

// MerchantGroup 商户分组，只做统筹归类。
type MerchantGroup struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(64);not null;uniqueIndex" json:"name"`
	CreatedBy string    `gorm:"column:created_by;type:varchar(64);not null;default:''" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by;type:varchar(64);not null;default:''" json:"updated_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (MerchantGroup) TableName() string {
	return "merchant_groups"
}

// MerchantGroupMember 分组与商户的多对多关系。
type MerchantGroupMember struct {
	GroupID    uint `gorm:"column:group_id;primaryKey" json:"group_id"`
	MerchantID uint `gorm:"column:merchant_id;primaryKey;index" json:"merchant_id"`
}

func (MerchantGroupMember) TableName() string {
	return "merchant_group_members"
}
