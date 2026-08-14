package model

import "time"

const (
	RoleStatusDisabled = 0 // 禁用
	RoleStatusEnabled  = 1 // 启用
)

// Role 模型层：角色，对应 roles 表。Code 给前端用，例如 super / admin。
type Role struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                      // 角色ID
	Code      string    `gorm:"column:code;type:varchar(64);not null;uniqueIndex" json:"code"`     // 角色编码
	Name      string    `gorm:"column:name;type:varchar(64);not null" json:"name"`                 // 角色名称
	Remark    string    `gorm:"column:remark;type:varchar(255);not null;default:''" json:"remark"` // 备注
	Status    int       `gorm:"column:status;not null;default:1;index" json:"status"`              // 1启用 0禁用
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`                               // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`                               // 更新时间
}

// TableName 指定表名。
func (Role) TableName() string {
	return "roles"
}
