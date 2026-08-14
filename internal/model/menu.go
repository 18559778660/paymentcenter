package model

import "time"

const (
	MenuTypeDir    = 0 // 目录
	MenuTypeMenu   = 1 // 菜单
	MenuTypeButton = 2 // 按钮/权限码

	MenuStatusDisabled = 0 // 禁用
	MenuStatusEnabled  = 1 // 启用
)

// Menu 模型层：菜单与权限点，对应 menus 表。后续可给动态菜单和权限管理页使用。
type Menu struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                         // 菜单ID
	ParentID  uint      `gorm:"column:parent_id;not null;default:0;index" json:"parent_id"`           // 父菜单ID，0 表示顶级
	Name      string    `gorm:"column:name;type:varchar(64);not null;uniqueIndex" json:"name"`        // 路由 Name，需唯一
	Title     string    `gorm:"column:title;type:varchar(64);not null" json:"title"`                  // 显示标题
	Path      string    `gorm:"column:path;type:varchar(128);not null;default:''" json:"path"`        // 前端路由路径
	Component string    `gorm:"column:component;type:varchar(255);not null;default:''" json:"component"` // 前端组件路径
	Icon      string    `gorm:"column:icon;type:varchar(64);not null;default:''" json:"icon"`         // 图标
	AuthCode  string    `gorm:"column:auth_code;type:varchar(64);not null;default:''" json:"auth_code"` // 权限码，给 /auth/codes
	Type      int       `gorm:"column:type;not null;default:1" json:"type"`                           // 0目录 1菜单 2按钮
	Sort      int       `gorm:"column:sort;not null;default:0" json:"sort"`                           // 排序，越小越靠前
	Status    int       `gorm:"column:status;not null;default:1;index" json:"status"`                 // 1启用 0禁用
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`                                  // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`                                  // 更新时间
}

// TableName 指定表名。
func (Menu) TableName() string {
	return "menus"
}
