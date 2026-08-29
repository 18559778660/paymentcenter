package model

import "time"

const (
	StripeWordBankConfigWebhook  = "webhook链接"
	StripeWordBankConfigCallback   = "回调路径"
	StripeWordBankConfigDirectory  = "目录"
)

// StripeWordBank Stripe 路径单词库。
type StripeWordBank struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement;comment:ID" json:"id"`
	Name       string    `gorm:"column:name;type:varchar(128);not null;uniqueIndex;comment:路径名称" json:"name"`
	UsageCount int       `gorm:"column:usage_count;not null;default:0;comment:使用次数" json:"usage_count"`
	ConfigItem string    `gorm:"column:config_item;type:varchar(32);not null;index;comment:配置项" json:"config_item"`
	CreatedBy  string    `gorm:"column:created_by;type:varchar(64);not null;default:'';comment:创建人" json:"created_by"`
	UpdatedBy  string    `gorm:"column:updated_by;type:varchar(64);not null;default:'';comment:更新人" json:"updated_by"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (StripeWordBank) TableName() string {
	return "stripe_word_banks"
}
