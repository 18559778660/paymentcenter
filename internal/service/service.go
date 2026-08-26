package service

import (
	"strings"
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
	GetUsersByIDs(ids []uint) ([]model.User, error)
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
	GetMerchantsByIDs(ids []uint) ([]model.Merchant, error)
	CreateMerchantGroup(g *model.MerchantGroup) error
	SaveMerchantGroup(g *model.MerchantGroup) error
	GetMerchantGroupByID(id uint) (*model.MerchantGroup, error)
	FindMerchantGroupByName(name string) (*model.MerchantGroup, error)
	ListMerchantGroups(filter store.MerchantGroupListFilter) ([]model.MerchantGroup, error)
	DeleteMerchantGroup(id uint) error
	ListMerchantGroupMembers(groupIDs []uint) ([]model.MerchantGroupMember, error)
	ReplaceMerchantGroupMembers(groupID uint, merchantIDs []uint) error
	CreateCardType(item *model.CardType) error
	SaveCardType(item *model.CardType) error
	GetCardTypeByID(id uint) (*model.CardType, error)
	FindCardTypeByCode(code string) (*model.CardType, error)
	ListCardTypes(filter store.CardTypeListFilter) ([]model.CardType, error)
	CountCardTypes() (int64, error)
	CreateCurrency(item *model.Currency) error
	SaveCurrency(item *model.Currency) error
	GetCurrencyByID(id uint) (*model.Currency, error)
	FindCurrencyByCode(code string) (*model.Currency, error)
	ListCurrencies(filter store.CurrencyListFilter) ([]model.Currency, error)
	CountCurrencies() (int64, error)
	DeleteCurrency(id uint) error
	CreateCountry(item *model.Country) error
	SaveCountry(item *model.Country) error
	GetCountryByID(id uint) (*model.Country, error)
	FindCountryByCode(code string) (*model.Country, error)
	ListCountries(filter store.CountryListFilter) ([]model.Country, error)
	CountCountries() (int64, error)
	DeleteCountry(id uint) error
	CreateChannel(item *model.Channel) error
	SaveChannel(item *model.Channel) error
	GetChannelByID(id uint) (*model.Channel, error)
	FindChannelByName(name string) (*model.Channel, error)
	ListChannels(filter store.ChannelListFilter) ([]model.Channel, error)
	CreateSiteA(item *model.SiteA) error
	GetSiteAByID(id uint) (*model.SiteA, error)
	FindSiteAByDomain(domain string) (*model.SiteA, error)
	ListSiteAs(filter store.SiteAListFilter) ([]model.SiteA, error)
	BatchUpdateSiteAStatus(ids []uint, status, operator string) error
}

// App 业务层入口。各业务方法拆在同包的 auth.go / menu.go / order.go / seed.go。
type App struct {
	store          Store
	authSecret     string
	tokenTTL       time.Duration
	gatewayBaseURL string
}

// NewApp 创建业务层。store 一般传入 *store.MySQLStore。
func NewApp(st Store, authSecret string, tokenTTL time.Duration, gatewayBaseURL string) *App {
	return &App{
		store:          st,
		authSecret:     authSecret,
		tokenTTL:       tokenTTL,
		gatewayBaseURL: strings.TrimRight(gatewayBaseURL, "/"),
	}
}
