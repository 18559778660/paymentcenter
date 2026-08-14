package service

import (
	"time"

	"paymentcenter/internal/model"
)

// Store 数据访问接口。service 只依赖这个抽象，不直接碰数据库。
type Store interface {
	SaveOrder(order *model.Order) error
	GetOrder(id string) (*model.Order, error)
	ListOrders() ([]*model.Order, error)
	CreateUser(user *model.User) error
	SaveUser(user *model.User) error
	GetUserByID(id uint) (*model.User, error)
	FindUserByUsername(username string) (*model.User, error)
	CreateRole(role *model.Role) error
	FindRoleByCode(code string) (*model.Role, error)
	ListRolesByUserID(userID uint) ([]model.Role, error)
	EnsureUserRole(userID, roleID uint) error
	CreateMenu(menu *model.Menu) error
	FindMenuByName(name string) (*model.Menu, error)
	ListMenus() ([]model.Menu, error)
	SaveMenu(menu *model.Menu) error
	EnsureRoleMenu(roleID, menuID uint) error
	ListMenusByUserID(userID uint) ([]model.Menu, error)
}

// App 业务层入口。各业务方法拆在同包的 auth.go / menu.go / order.go / seed.go。
type App struct {
	store      Store
	authSecret string
	tokenTTL   time.Duration
}

// NewApp 创建业务层。store 一般传入 *store.MySQLStore。
func NewApp(st Store, authSecret string, tokenTTL time.Duration) *App {
	return &App{store: st, authSecret: authSecret, tokenTTL: tokenTTL}
}
