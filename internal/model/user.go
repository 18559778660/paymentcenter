package model

import "time"

const (
	UserStatusDisabled = 0
	UserStatusEnabled  = 1
)

// User 是后台用户模型，对应 users 表。
type User struct {
	ID           uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"column:username;type:varchar(64);not null;uniqueIndex" json:"username"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Role         string     `gorm:"column:role;type:varchar(32);not null;default:'admin'" json:"role"`
	Status       int        `gorm:"column:status;not null;default:1;index" json:"status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
