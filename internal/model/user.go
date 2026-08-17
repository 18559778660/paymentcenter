package model

import "time"

const (
	UserStatusDisabled = 0 // 禁用
	UserStatusEnabled  = 1 // 启用

	UserTypeAdmin    = "admin"    // 后台管理员
	UserTypeMerchant = "merchant" // 商户账号，可登录同一后台
)

// User 模型层：后台用户，对应 users 表。
type User struct {
	ID           uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                          // 用户ID
	Username     string     `gorm:"column:username;type:varchar(64);not null;uniqueIndex" json:"username"`                 // 登录账号
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`                               // 密码哈希，不返回给前端
	RealName     string     `gorm:"column:real_name;type:varchar(64);not null;default:''" json:"real_name"`                 // 显示名
	Avatar       string     `gorm:"column:avatar;type:varchar(255);not null;default:''" json:"avatar"`                      // 头像地址
	HomePath     string     `gorm:"column:home_path;type:varchar(128);not null;default:'/dashboard/analytics'" json:"home_path"` // 登录后跳转页
	Type         string     `gorm:"column:type;type:varchar(32);not null;default:'admin';index" json:"type"`                // admin|merchant
	Status       int        `gorm:"column:status;not null;default:1;index" json:"status"`                                    // 1启用 0禁用
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`                                     // 最后登录时间
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`                                                     // 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`                                                     // 更新时间
}

// TableName 指定表名。
func (User) TableName() string {
	return "users"
}
