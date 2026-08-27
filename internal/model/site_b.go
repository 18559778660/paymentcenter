package model

import "time"

const (
	SiteBFrameworkOther   = "其他"
	SiteBFrameworkShopyy  = "shopyy"
	SiteBFrameworkShopify = "shopify"

	SiteBDefaultRunDirectory = "deal"
)

// SiteB B 站管理。
type SiteB struct {
	ID             uint      `gorm:"column:id;primaryKey;autoIncrement;comment:B站ID" json:"id"`
	Domain         string    `gorm:"column:domain;type:varchar(191);not null;uniqueIndex;comment:域名" json:"domain"`
	PlatformID     uint      `gorm:"column:platform_id;not null;index;comment:通道平台ID" json:"platform_id"`
	Framework      string    `gorm:"column:framework;type:varchar(32);not null;comment:框架" json:"framework"`
	Status         bool      `gorm:"column:status;not null;default:1;index;comment:状态 1启用 0停用" json:"status"`
	ChannelEnabled bool      `gorm:"column:channel_enabled;not null;default:1;comment:通道启用 1启用 0停用" json:"channel_enabled"`
	IsFtp          bool      `gorm:"column:is_ftp;not null;default:1;comment:是否FTP 1是 0否" json:"is_ftp"`
	Host           string    `gorm:"column:host;type:varchar(255);not null;default:'';comment:主机" json:"host"`
	Account        string    `gorm:"column:account;type:varchar(128);not null;default:'';comment:账号" json:"account"`
	Password       string    `gorm:"column:password;type:varchar(128);not null;default:'';comment:密码" json:"password"`
	RunDirectory   string    `gorm:"column:run_directory;type:varchar(64);not null;default:'deal';comment:运行目录" json:"run_directory"`
	Remark         string    `gorm:"column:remark;type:varchar(512);not null;default:'';comment:备注" json:"remark"`
	CreatedBy      string    `gorm:"column:created_by;type:varchar(64);not null;default:'';comment:创建人" json:"created_by"`
	UpdatedBy      string    `gorm:"column:updated_by;type:varchar(64);not null;default:'';comment:更新人" json:"updated_by"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (SiteB) TableName() string {
	return "site_bs"
}
