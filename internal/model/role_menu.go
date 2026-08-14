package model

// RoleMenu 模型层：角色和菜单的多对多关联，对应 role_menus 表。
type RoleMenu struct {
	RoleID uint `gorm:"column:role_id;primaryKey"` // 角色ID
	MenuID uint `gorm:"column:menu_id;primaryKey"` // 菜单ID
}

// TableName 指定表名。
func (RoleMenu) TableName() string {
	return "role_menus"
}
