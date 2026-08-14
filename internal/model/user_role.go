package model

// UserRole 模型层：用户和角色的多对多关联，对应 user_roles 表。
type UserRole struct {
	UserID uint `gorm:"column:user_id;primaryKey"` // 用户ID
	RoleID uint `gorm:"column:role_id;primaryKey"` // 角色ID
}

// TableName 指定表名。
func (UserRole) TableName() string {
	return "user_roles"
}
