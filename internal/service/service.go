package service

import (
	"time"

	"paymentcenter/internal/model"
	"paymentcenter/internal/store"
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
	GetMenuByID(id uint) (*model.Menu, error)
	ListMenus() ([]model.Menu, error)
	ListAllMenus() ([]model.Menu, error)
	SaveMenu(menu *model.Menu) error
	DeleteMenu(id uint) error
	CountMenusByParentID(parentID uint) (int64, error)
	DeleteRoleMenusByMenuID(menuID uint) error
	MenuNameExists(name string, excludeID uint) (bool, error)
	MenuPathExists(path string, excludeID uint) (bool, error)
	EnsureRoleMenu(roleID, menuID uint) error
	ListMenusByUserID(userID uint) ([]model.Menu, error)
	ListRoles() ([]model.Role, error)
	CreateMerchant(m *model.Merchant) error
	SaveMerchant(m *model.Merchant) error
	GetMerchantByID(id uint) (*model.Merchant, error)
	FindMerchantByName(name string) (*model.Merchant, error)
	FindMerchantByAccount(account string) (*model.Merchant, error)
	MaxWINMerchantAccountSeq() (int, error)
	ListMerchants(filter store.MerchantListFilter) ([]model.Merchant, error)
	ListMerchantOptions() ([]model.Merchant, error)
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
