package model

import "time"

const (
	SiteAStatusPending  = "pending"
	SiteAStatusAudited  = "audited"
	SiteAStatusDisabled = "disabled"

	SiteAFrameworkWooCommerce = "woocommerce"
	SiteAFrameworkShopyy       = "shopyy"
)

// SiteA A 站管理。
type SiteA struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement;comment:A站ID" json:"id"`
	MerchantID uint      `gorm:"column:merchant_id;not null;index;comment:商户ID" json:"merchant_id"`
	Domain     string    `gorm:"column:domain;type:varchar(191);not null;uniqueIndex;comment:域名" json:"domain"`
	Framework  string    `gorm:"column:framework;type:varchar(32);not null;comment:框架 woocommerce shopyy" json:"framework"`
	Status     string    `gorm:"column:status;type:varchar(16);not null;default:'pending';index;comment:状态 pending待审核 audited已审核 disabled禁用" json:"status"`
	CreatedBy  string    `gorm:"column:created_by;type:varchar(64);not null;default:'';comment:创建人" json:"created_by"`
	UpdatedBy  string    `gorm:"column:updated_by;type:varchar(64);not null;default:'';comment:更新人" json:"updated_by"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (SiteA) TableName() string {
	return "site_as"
}
