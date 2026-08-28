package model

import "time"

const (
	PlatformStatusDisabled = 0
	PlatformStatusEnabled  = 1

	PlatformCodeStripe = "stripe"
	PlatformCodePaypal = "paypal"
)

// Platform 支付通道平台分类。
type Platform struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement;comment:平台ID" json:"id"`
	Code      string    `gorm:"column:code;type:varchar(32);not null;uniqueIndex;comment:平台编码 stripe paypal" json:"code"`
	Name      string    `gorm:"column:name;type:varchar(64);not null;comment:平台名称" json:"name"`
	Sort      int       `gorm:"column:sort;not null;default:0;comment:排序" json:"sort"`
	Status    int       `gorm:"column:status;not null;default:1;index;comment:状态 1启用 0禁用" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (Platform) TableName() string {
	return "platforms"
}
