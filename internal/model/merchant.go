package model

import "time"

const (
	MerchantStatusDisabled = 0
	MerchantStatusEnabled  = 1

	MerchantLimitModeUnified = "统一配置" // 限制模式暂时写死
)

// Merchant 商户资料。登录账号在 users 表（type=merchant），本表存商户业务字段。
type Merchant struct {
	ID               uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID           uint      `gorm:"column:user_id;not null;uniqueIndex" json:"user_id"` // 关联 users.id
	Name             string    `gorm:"column:name;type:varchar(64);not null;uniqueIndex" json:"name"` // 商户名
	Account          string    `gorm:"column:account;type:varchar(64);not null;uniqueIndex" json:"account"` // 登录账号（冗余，同 username）
	PasswordPlain    string    `gorm:"column:password_plain;type:varchar(64);not null;default:''" json:"password_plain"` // 明文密码，列表展示用
	Contact          string    `gorm:"column:contact;type:varchar(128);not null;default:''" json:"contact"` // 联系方式
	ParentID         *uint     `gorm:"column:parent_id;index" json:"parent_id"` // 上级商户ID
	AutoShip         bool      `gorm:"column:auto_ship;not null" json:"auto_ship"`                         // 自动发货；不要加 default，否则 false 会被 GORM 当成空值跳过
	ConfirmEmail     bool      `gorm:"column:confirm_email;not null" json:"confirm_email"`                 // 确认邮件；同上，false 必须能写入
	Status           int       `gorm:"column:status;not null;default:1;index" json:"status"` // 1启用 0禁用
	LimitMode        string    `gorm:"column:limit_mode;type:varchar(32);not null;default:'统一配置'" json:"limit_mode"` // 限制模式
	RateDiff         int       `gorm:"column:rate_diff;not null;default:0" json:"rate_diff"` // 汇率偏差 0-100
	HoldRate         int       `gorm:"column:hold_rate;not null;default:0" json:"hold_rate"` // 扣单
	MutualHoldRate   int       `gorm:"column:mutual_hold_rate;not null;default:0" json:"mutual_hold_rate"` // 互抛扣单
	HoldStatus       int       `gorm:"column:hold_status;not null;default:0" json:"hold_status"` // 扣单状态 0关1开
	MutualHoldStatus int       `gorm:"column:mutual_hold_status;not null;default:0" json:"mutual_hold_status"` // 互抛扣单状态
	SecretKey        string    `gorm:"column:secret_key;type:varchar(64);not null;default:''" json:"secret_key"` // 密钥
	AuditSiteA       string    `gorm:"column:audit_site_a;type:varchar(16);not null;default:'manual'" json:"audit_site_a"` // manual|auto
	Starred          bool      `gorm:"column:starred;not null;default:false" json:"starred"` // 星标
	CreatedBy        string    `gorm:"column:created_by;type:varchar(64);not null;default:''" json:"created_by"`
	UpdatedBy        string    `gorm:"column:updated_by;type:varchar(64);not null;default:''" json:"updated_by"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Merchant) TableName() string {
	return "merchants"
}
