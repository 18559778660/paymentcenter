package store

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"paymentcenter/internal/model"
)

// MySQLStore 数据库层：用 GORM 访问 MySQL。
type MySQLStore struct {
	db *gorm.DB
}

// NewMySQLStore 连接 MySQL，并自动迁移订单、用户、角色、菜单相关表。
func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&model.Order{},
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.UserRole{},
		&model.RoleMenu{},
		&model.Merchant{},
		&model.MerchantGroup{},
		&model.MerchantGroupMember{},
		&model.CardType{},
		&model.Currency{},
		&model.Country{},
		&model.Platform{},
		&model.Channel{},
		&model.SiteA{},
		&model.SiteB{},
		&model.ChannelAccount{},
		&model.ChannelGroup{},
		&model.ChannelGroupMember{},
		&model.StripeWordBank{},
	); err != nil {
		return nil, err
	}
	return &MySQLStore{db: db}, nil
}

// Close 关闭数据库连接。
func (s *MySQLStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// SaveOrder 新增或更新订单。
func (s *MySQLStore) SaveOrder(order *model.Order) error {
	return s.db.Save(order).Error
}

// GetOrder 按支付中心订单号查询一笔订单。
func (s *MySQLStore) GetOrder(id string) (*model.Order, error) {
	var order model.Order
	tx := s.db.Where("id = ?", id).Limit(1).Find(&order)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &order, nil
}

// ListOrders 查询最近 100 笔订单，按创建时间倒序。
func (s *MySQLStore) ListOrders() ([]*model.Order, error) {
	var orders []*model.Order
	if err := s.db.Order("created_at DESC").Limit(100).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// CreateUser 插入一条后台用户。
func (s *MySQLStore) CreateUser(user *model.User) error {
	return s.db.Create(user).Error
}

// SaveUser 更新后台用户，例如写入最后登录时间。
func (s *MySQLStore) SaveUser(user *model.User) error {
	return s.db.Save(user).Error
}

// GetUserByID 按主键查询用户。
func (s *MySQLStore) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	tx := s.db.Where("id = ?", id).Limit(1).Find(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &user, nil
}

// GetUsersByIDs 按主键批量查用户。
func (s *MySQLStore) GetUsersByIDs(ids []uint) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	err := s.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// FindUserByUsername 按登录账号查询用户，登录时使用。
func (s *MySQLStore) FindUserByUsername(username string) (*model.User, error) {
	var user model.User
	tx := s.db.Where("username = ?", username).Limit(1).Find(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &user, nil
}

// CreateRole 插入一条角色。
func (s *MySQLStore) CreateRole(role *model.Role) error {
	return s.db.Create(role).Error
}

// FindRoleByCode 按角色编码查询，例如 super、admin。
func (s *MySQLStore) FindRoleByCode(code string) (*model.Role, error) {
	var role model.Role
	tx := s.db.Where("code = ?", code).Limit(1).Find(&role)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &role, nil
}

// ListRolesByUserID 查询某用户已绑定且启用的角色。
func (s *MySQLStore) ListRolesByUserID(userID uint) ([]model.Role, error) {
	var roles []model.Role
	err := s.db.Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, model.RoleStatusEnabled).
		Find(&roles).Error
	return roles, err
}

// EnsureUserRole 给用户绑定角色，已存在则不重复插入。
func (s *MySQLStore) EnsureUserRole(userID, roleID uint) error {
	rel := model.UserRole{UserID: userID, RoleID: roleID}
	return s.db.Where(rel).FirstOrCreate(&rel).Error
}

// CreateMenu 插入一条菜单或权限点。
func (s *MySQLStore) CreateMenu(menu *model.Menu) error {
	return s.db.Create(menu).Error
}

// FindMenuByName 按路由 Name 查询菜单，种子数据用来判断是否已存在。
func (s *MySQLStore) FindMenuByName(name string) (*model.Menu, error) {
	var menu model.Menu
	tx := s.db.Where("name = ?", name).Limit(1).Find(&menu)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &menu, nil
}

// GetMenuByID 按主键查询菜单。
func (s *MySQLStore) GetMenuByID(id uint) (*model.Menu, error) {
	var menu model.Menu
	tx := s.db.Where("id = ?", id).Limit(1).Find(&menu)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &menu, nil
}

// ListMenus 查询全部启用中的菜单。
func (s *MySQLStore) ListMenus() ([]model.Menu, error) {
	var menus []model.Menu
	err := s.db.Where("status = ?", model.MenuStatusEnabled).
		Order("sort ASC, id ASC").
		Find(&menus).Error
	return menus, err
}

// ListAllMenus 查询全部菜单（含禁用），菜单管理页用。
func (s *MySQLStore) ListAllMenus() ([]model.Menu, error) {
	var menus []model.Menu
	err := s.db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

// SaveMenu 更新菜单字段。
func (s *MySQLStore) SaveMenu(menu *model.Menu) error {
	return s.db.Save(menu).Error
}

// DeleteMenu 按主键删除菜单。
func (s *MySQLStore) DeleteMenu(id uint) error {
	return s.db.Delete(&model.Menu{}, id).Error
}

// CountMenusByParentID 统计某父菜单下的子菜单数量。
func (s *MySQLStore) CountMenusByParentID(parentID uint) (int64, error) {
	var count int64
	err := s.db.Model(&model.Menu{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// DeleteRoleMenusByMenuID 删除某菜单的全部角色绑定。
func (s *MySQLStore) DeleteRoleMenusByMenuID(menuID uint) error {
	return s.db.Where("menu_id = ?", menuID).Delete(&model.RoleMenu{}).Error
}

// MenuNameExists 判断路由 Name 是否已被占用。excludeID>0 时排除自身（编辑用）。
func (s *MySQLStore) MenuNameExists(name string, excludeID uint) (bool, error) {
	q := s.db.Model(&model.Menu{}).Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// MenuPathExists 判断路径是否已被占用。空路径不算；excludeID>0 时排除自身。
func (s *MySQLStore) MenuPathExists(path string, excludeID uint) (bool, error) {
	if path == "" {
		return false, nil
	}
	q := s.db.Model(&model.Menu{}).Where("path = ?", path)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// EnsureRoleMenu 给角色绑定菜单，已存在则不重复插入。
func (s *MySQLStore) EnsureRoleMenu(roleID, menuID uint) error {
	rel := model.RoleMenu{RoleID: roleID, MenuID: menuID}
	return s.db.Where(rel).FirstOrCreate(&rel).Error
}

// ListMenusByUserID 按用户角色查出该用户能访问的菜单和权限点。
func (s *MySQLStore) ListMenusByUserID(userID uint) ([]model.Menu, error) {
	var menus []model.Menu
	err := s.db.Table("menus").
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_menus.role_id").
		Where("user_roles.user_id = ? AND menus.status = ?", userID, model.MenuStatusEnabled).
		Distinct().
		Order("menus.sort ASC, menus.id ASC").
		Find(&menus).Error
	return menus, err
}

// ListRoles 查询全部角色。
func (s *MySQLStore) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := s.db.Order("id ASC").Find(&roles).Error
	return roles, err
}

// CreateMerchant 插入商户资料。
func (s *MySQLStore) CreateMerchant(m *model.Merchant) error {
	return s.db.Create(m).Error
}

// SaveMerchant 更新商户资料。
func (s *MySQLStore) SaveMerchant(m *model.Merchant) error {
	return s.db.Save(m).Error
}

// GetMerchantByID 按主键查商户。
func (s *MySQLStore) GetMerchantByID(id uint) (*model.Merchant, error) {
	var m model.Merchant
	tx := s.db.Where("id = ?", id).Limit(1).Find(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &m, nil
}

// FindMerchantByName 按商户名查询。
func (s *MySQLStore) FindMerchantByName(name string) (*model.Merchant, error) {
	var m model.Merchant
	tx := s.db.Where("name = ?", name).Limit(1).Find(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &m, nil
}

// FindMerchantByAccount 按登录账号查询。
func (s *MySQLStore) FindMerchantByAccount(account string) (*model.Merchant, error) {
	var m model.Merchant
	tx := s.db.Where("account = ?", account).Limit(1).Find(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &m, nil
}

// MaxWINMerchantAccountSeq 取已有 WIN##### 账号中的最大序号；没有则返回 -1。
// 账号定长递增，按 account 倒序取一条即可。
func (s *MySQLStore) MaxWINMerchantAccountSeq() (int, error) {
	var account string
	tx := s.db.Model(&model.Merchant{}).
		Select("account").
		Where("account LIKE ?", "WIN_____").
		Order("account DESC").
		Limit(1).
		Scan(&account)
	if tx.Error != nil {
		return -1, tx.Error
	}
	if account == "" {
		return -1, nil
	}
	var n int
	if _, err := fmt.Sscanf(account[3:], "%d", &n); err != nil {
		return -1, nil
	}
	return n, nil
}

// MerchantListFilter 商户列表筛选条件。
type MerchantListFilter struct {
	Name             string
	ParentID         *uint
	Status           *int
	HoldStatus       *int
	MutualHoldStatus *int
}

// ListMerchants 按条件查询商户，按 id 倒序。
func (s *MySQLStore) ListMerchants(filter MerchantListFilter) ([]model.Merchant, error) {
	q := s.db.Model(&model.Merchant{})
	if filter.Name != "" {
		q = q.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.ParentID != nil {
		q = q.Where("parent_id = ?", *filter.ParentID)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.HoldStatus != nil {
		q = q.Where("hold_status = ?", *filter.HoldStatus)
	}
	if filter.MutualHoldStatus != nil {
		q = q.Where("mutual_hold_status = ?", *filter.MutualHoldStatus)
	}
	var list []model.Merchant
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// ListMerchantOptions 上级下拉用：返回全部商户简要信息。
func (s *MySQLStore) ListMerchantOptions() ([]model.Merchant, error) {
	var list []model.Merchant
	err := s.db.Select("id", "name", "account").Order("id DESC").Find(&list).Error
	return list, err
}

// GetMerchantsByIDs 按主键批量查商户。
func (s *MySQLStore) GetMerchantsByIDs(ids []uint) ([]model.Merchant, error) {
	if len(ids) == 0 {
		return []model.Merchant{}, nil
	}
	var list []model.Merchant
	err := s.db.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

// MerchantGroupListFilter 分组列表筛选。
type MerchantGroupListFilter struct {
	ID   *uint
	Name string
}

// CreateMerchantGroup 插入分组。
func (s *MySQLStore) CreateMerchantGroup(g *model.MerchantGroup) error {
	return s.db.Create(g).Error
}

// SaveMerchantGroup 更新分组。
func (s *MySQLStore) SaveMerchantGroup(g *model.MerchantGroup) error {
	return s.db.Save(g).Error
}

// GetMerchantGroupByID 按主键查分组。
func (s *MySQLStore) GetMerchantGroupByID(id uint) (*model.MerchantGroup, error) {
	var g model.MerchantGroup
	tx := s.db.Where("id = ?", id).Limit(1).Find(&g)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &g, nil
}

// FindMerchantGroupByName 按分组名查询。
func (s *MySQLStore) FindMerchantGroupByName(name string) (*model.MerchantGroup, error) {
	var g model.MerchantGroup
	tx := s.db.Where("name = ?", name).Limit(1).Find(&g)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &g, nil
}

// ListMerchantGroups 按条件查询分组，按 id 倒序。
func (s *MySQLStore) ListMerchantGroups(filter MerchantGroupListFilter) ([]model.MerchantGroup, error) {
	q := s.db.Model(&model.MerchantGroup{})
	if filter.ID != nil {
		q = q.Where("id = ?", *filter.ID)
	}
	if filter.Name != "" {
		q = q.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	var list []model.MerchantGroup
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// DeleteMerchantGroup 删除分组及其成员关系。
func (s *MySQLStore) DeleteMerchantGroup(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", id).Delete(&model.MerchantGroupMember{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.MerchantGroup{}).Error
	})
}

// ListMerchantGroupMembers 查出若干分组下的商户 ID。
func (s *MySQLStore) ListMerchantGroupMembers(groupIDs []uint) ([]model.MerchantGroupMember, error) {
	if len(groupIDs) == 0 {
		return []model.MerchantGroupMember{}, nil
	}
	var list []model.MerchantGroupMember
	err := s.db.Where("group_id IN ?", groupIDs).Find(&list).Error
	return list, err
}

// ReplaceMerchantGroupMembers 用新列表覆盖分组成员。
func (s *MySQLStore) ReplaceMerchantGroupMembers(groupID uint, merchantIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&model.MerchantGroupMember{}).Error; err != nil {
			return err
		}
		if len(merchantIDs) == 0 {
			return nil
		}
		rows := make([]model.MerchantGroupMember, 0, len(merchantIDs))
		for _, merchantID := range merchantIDs {
			rows = append(rows, model.MerchantGroupMember{
				GroupID:    groupID,
				MerchantID: merchantID,
			})
		}
		return tx.Create(&rows).Error
	})
}

// CardTypeListFilter 卡类型列表筛选。
type CardTypeListFilter struct {
	Field   string
	Keyword string
	Names   []string
}

// CreateCardType 插入卡类型。
func (s *MySQLStore) CreateCardType(item *model.CardType) error {
	return s.db.Create(item).Error
}

// SaveCardType 更新卡类型。
func (s *MySQLStore) SaveCardType(item *model.CardType) error {
	return s.db.Save(item).Error
}

// GetCardTypeByID 按主键查卡类型。
func (s *MySQLStore) GetCardTypeByID(id uint) (*model.CardType, error) {
	var item model.CardType
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindCardTypeByCode 按缩写查卡类型。
func (s *MySQLStore) FindCardTypeByCode(code string) (*model.CardType, error) {
	var item model.CardType
	tx := s.db.Where("code = ?", code).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListCardTypes 按条件查询卡类型，按 id 倒序。
func (s *MySQLStore) ListCardTypes(filter CardTypeListFilter) ([]model.CardType, error) {
	q := s.db.Model(&model.CardType{})
	keyword := filter.Keyword
	if keyword != "" {
		like := "%" + keyword + "%"
		switch filter.Field {
		case "code":
			q = q.Where("code LIKE ?", like)
		case "name":
			if len(filter.Names) > 0 {
				q = q.Where("name IN ?", filter.Names)
			} else {
				q = q.Where("name LIKE ?", like)
			}
		default:
			if len(filter.Names) > 0 {
				q = q.Where("code LIKE ? OR name IN ?", like, filter.Names)
			} else {
				q = q.Where("code LIKE ? OR name LIKE ?", like, like)
			}
		}
	}
	var list []model.CardType
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// CountCardTypes 统计卡类型数量。
func (s *MySQLStore) CountCardTypes() (int64, error) {
	var n int64
	err := s.db.Model(&model.CardType{}).Count(&n).Error
	return n, err
}

// CurrencyListFilter 货币列表筛选。
type CurrencyListFilter struct {
	Field   string
	Keyword string
}

// CreateCurrency 插入货币。
func (s *MySQLStore) CreateCurrency(item *model.Currency) error {
	return s.db.Create(item).Error
}

// SaveCurrency 更新货币。
func (s *MySQLStore) SaveCurrency(item *model.Currency) error {
	return s.db.Save(item).Error
}

// GetCurrencyByID 按主键查货币。
func (s *MySQLStore) GetCurrencyByID(id uint) (*model.Currency, error) {
	var item model.Currency
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindCurrencyByCode 按编码查货币。
func (s *MySQLStore) FindCurrencyByCode(code string) (*model.Currency, error) {
	var item model.Currency
	tx := s.db.Where("code = ?", code).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListCurrencies 按条件查询货币，按 id 倒序。
func (s *MySQLStore) ListCurrencies(filter CurrencyListFilter) ([]model.Currency, error) {
	q := s.db.Model(&model.Currency{})
	keyword := filter.Keyword
	if keyword != "" {
		like := "%" + keyword + "%"
		switch filter.Field {
		case "code":
			q = q.Where("code LIKE ?", like)
		case "name":
			q = q.Where("name LIKE ?", like)
		default:
			q = q.Where("code LIKE ? OR name LIKE ?", like, like)
		}
	}
	var list []model.Currency
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// CountCurrencies 统计货币数量。
func (s *MySQLStore) CountCurrencies() (int64, error) {
	var n int64
	err := s.db.Model(&model.Currency{}).Count(&n).Error
	return n, err
}

// DeleteCurrency 删除货币。
func (s *MySQLStore) DeleteCurrency(id uint) error {
	return s.db.Delete(&model.Currency{}, id).Error
}

// CountryListFilter 国家列表筛选。
type CountryListFilter struct {
	Field   string
	Keyword string
}

// CreateCountry 插入国家。
func (s *MySQLStore) CreateCountry(item *model.Country) error {
	return s.db.Create(item).Error
}

// SaveCountry 更新国家。
func (s *MySQLStore) SaveCountry(item *model.Country) error {
	return s.db.Save(item).Error
}

// GetCountryByID 按主键查国家。
func (s *MySQLStore) GetCountryByID(id uint) (*model.Country, error) {
	var item model.Country
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindCountryByCode 按编码查国家。
func (s *MySQLStore) FindCountryByCode(code string) (*model.Country, error) {
	var item model.Country
	tx := s.db.Where("code = ?", code).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListCountries 按条件查询国家，按 id 倒序。
func (s *MySQLStore) ListCountries(filter CountryListFilter) ([]model.Country, error) {
	q := s.db.Model(&model.Country{})
	keyword := filter.Keyword
	if keyword != "" {
		like := "%" + keyword + "%"
		switch filter.Field {
		case "code":
			q = q.Where("code LIKE ?", like)
		case "name":
			q = q.Where("name LIKE ?", like)
		default:
			q = q.Where("code LIKE ? OR name LIKE ?", like, like)
		}
	}
	var list []model.Country
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// CountCountries 统计国家数量。
func (s *MySQLStore) CountCountries() (int64, error) {
	var n int64
	err := s.db.Model(&model.Country{}).Count(&n).Error
	return n, err
}

// DeleteCountry 删除国家。
func (s *MySQLStore) DeleteCountry(id uint) error {
	return s.db.Delete(&model.Country{}, id).Error
}

// ChannelListFilter 通道列表筛选。
type ChannelListFilter struct {
	ID         *uint
	Name       string
	PlatformID *uint
}

// CreateChannel 插入通道。
func (s *MySQLStore) CreateChannel(item *model.Channel) error {
	return s.db.Create(item).Error
}

// SaveChannel 更新通道。
func (s *MySQLStore) SaveChannel(item *model.Channel) error {
	return s.db.Save(item).Error
}

// GetChannelByID 按主键查通道。
func (s *MySQLStore) GetChannelByID(id uint) (*model.Channel, error) {
	var item model.Channel
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindChannelByName 按通道名查询。
func (s *MySQLStore) FindChannelByName(name string) (*model.Channel, error) {
	var item model.Channel
	tx := s.db.Where("name = ?", name).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListChannels 按条件查询通道，按 id 倒序。
func (s *MySQLStore) ListChannels(filter ChannelListFilter) ([]model.Channel, error) {
	q := s.db.Model(&model.Channel{})
	if filter.ID != nil {
		q = q.Where("id = ?", *filter.ID)
	}
	if filter.Name != "" {
		q = q.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.PlatformID != nil {
		q = q.Where("platform_id = ?", *filter.PlatformID)
	}
	var list []model.Channel
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// SiteAListFilter A 站列表筛选。
type SiteAListFilter struct {
	MerchantID *uint
	Domain     string
	Status     string
}

// CreateSiteA 插入 A 站。
func (s *MySQLStore) CreateSiteA(item *model.SiteA) error {
	return s.db.Create(item).Error
}

// GetSiteAByID 按主键查 A 站。
func (s *MySQLStore) GetSiteAByID(id uint) (*model.SiteA, error) {
	var item model.SiteA
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindSiteAByDomain 按域名查 A 站。
func (s *MySQLStore) FindSiteAByDomain(domain string) (*model.SiteA, error) {
	var item model.SiteA
	tx := s.db.Where("domain = ?", domain).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListSiteAs 按条件查询 A 站，按 id 倒序。
func (s *MySQLStore) ListSiteAs(filter SiteAListFilter) ([]model.SiteA, error) {
	q := s.db.Model(&model.SiteA{})
	if filter.MerchantID != nil {
		q = q.Where("merchant_id = ?", *filter.MerchantID)
	}
	if filter.Domain != "" {
		q = q.Where("domain LIKE ?", "%"+filter.Domain+"%")
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	var list []model.SiteA
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// BatchUpdateSiteAStatus 批量更新 A 站状态。
func (s *MySQLStore) BatchUpdateSiteAStatus(ids []uint, status, operator string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.db.Model(&model.SiteA{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"status":     status,
		"updated_by": operator,
	}).Error
}

// SiteBListFilter B 站列表筛选。
type SiteBListFilter struct {
	ID       *uint
	Domain   string
	Remark   string
	Status   *bool
	PlatformID *uint
}

// CreateSiteB 插入 B 站。
func (s *MySQLStore) CreateSiteB(item *model.SiteB) error {
	return s.db.Create(item).Error
}

// SaveSiteB 更新 B 站。
func (s *MySQLStore) SaveSiteB(item *model.SiteB) error {
	return s.db.Save(item).Error
}

// GetSiteBByID 按主键查 B 站。
func (s *MySQLStore) GetSiteBByID(id uint) (*model.SiteB, error) {
	var item model.SiteB
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindSiteBByDomain 按域名查 B 站。
func (s *MySQLStore) FindSiteBByDomain(domain string) (*model.SiteB, error) {
	var item model.SiteB
	tx := s.db.Where("domain = ?", domain).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListSiteBs 按条件查询 B 站，按 id 倒序。
func (s *MySQLStore) ListSiteBs(filter SiteBListFilter) ([]model.SiteB, error) {
	q := s.db.Model(&model.SiteB{})
	if filter.ID != nil {
		q = q.Where("id = ?", *filter.ID)
	}
	if filter.Domain != "" {
		q = q.Where("domain LIKE ?", "%"+filter.Domain+"%")
	}
	if filter.Remark != "" {
		q = q.Where("remark LIKE ?", "%"+filter.Remark+"%")
	}
	if filter.PlatformID != nil {
		q = q.Where("platform_id = ?", *filter.PlatformID)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	var list []model.SiteB
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// CreatePlatform 插入平台。
func (s *MySQLStore) CreatePlatform(item *model.Platform) error {
	return s.db.Create(item).Error
}

// FindPlatformByCode 按编码查平台。
func (s *MySQLStore) FindPlatformByCode(code string) (*model.Platform, error) {
	var item model.Platform
	tx := s.db.Where("code = ?", code).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// GetPlatformByID 按主键查平台。
func (s *MySQLStore) GetPlatformByID(id uint) (*model.Platform, error) {
	var item model.Platform
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListPlatforms 查询全部平台，按 sort、id 升序。
func (s *MySQLStore) ListPlatforms() ([]model.Platform, error) {
	var list []model.Platform
	err := s.db.Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// ChannelAccountListFilter 通道账号列表筛选。
type ChannelAccountListFilter struct {
	ID             *uint
	ChannelID      *uint
	SiteBID        *uint
	ChannelName    string
	Alias        string
	Remark       string
	CreatedFrom  string
	CreatedTo    string
	GroupID      *uint
	AssignedUserID *uint
	ListFilter     string
}

// AssignUserListFilter 分配子账号列表筛选。
type AssignUserListFilter struct {
	Field   string
	Keyword string
}

// ListUsersByType 按用户类型查询用户列表。
func (s *MySQLStore) ListUsersByType(userType string, filter AssignUserListFilter) ([]model.User, error) {
	q := s.db.Model(&model.User{}).Where("type = ?", userType)
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		if filter.Field == "nickname" {
			q = q.Where("real_name LIKE ?", "%"+keyword+"%")
		} else {
			q = q.Where("username LIKE ?", "%"+keyword+"%")
		}
	}
	var list []model.User
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// CountChannelAccountsByAssignedUserIDs 批量统计各子账号已分配通道账号数。
func (s *MySQLStore) CountChannelAccountsByAssignedUserIDs(userIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	type row struct {
		AssignedUserID uint
		Count          int64
	}
	var rows []row
	err := s.db.Model(&model.ChannelAccount{}).
		Select("assigned_user_id, COUNT(*) AS count").
		Where("assigned_user_id IN ?", userIDs).
		Group("assigned_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.AssignedUserID] = item.Count
	}
	return result, nil
}

// CreateChannelAccount 插入通道账号。
func (s *MySQLStore) CreateChannelAccount(item *model.ChannelAccount) error {
	return s.db.Create(item).Error
}

// SaveChannelAccount 更新通道账号。
func (s *MySQLStore) SaveChannelAccount(item *model.ChannelAccount) error {
	return s.db.Save(item).Error
}

// GetChannelAccountByID 按主键查通道账号。
func (s *MySQLStore) GetChannelAccountByID(id uint) (*model.ChannelAccount, error) {
	var item model.ChannelAccount
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindChannelAccountByChannelAndSiteB 按通道和 B 站查询。
func (s *MySQLStore) FindChannelAccountByChannelAndSiteB(channelID, siteBID uint) (*model.ChannelAccount, error) {
	var item model.ChannelAccount
	tx := s.db.Where("channel_id = ? AND site_b_id = ?", channelID, siteBID).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListChannelAccounts 按条件查询通道账号，按 id 倒序。
func (s *MySQLStore) ListChannelAccounts(filter ChannelAccountListFilter) ([]model.ChannelAccount, error) {
	q := s.db.Model(&model.ChannelAccount{})
	if filter.ID != nil {
		q = q.Where("channel_accounts.id = ?", *filter.ID)
	}
	if filter.Alias != "" {
		q = q.Where("channel_accounts.alias LIKE ?", "%"+filter.Alias+"%")
	}
	if filter.Remark != "" {
		q = q.Where("channel_accounts.remark LIKE ?", "%"+filter.Remark+"%")
	}
	if filter.ChannelID != nil {
		q = q.Where("channel_accounts.channel_id = ?", *filter.ChannelID)
	}
	if filter.SiteBID != nil {
		q = q.Where("channel_accounts.site_b_id = ?", *filter.SiteBID)
	}
	if filter.CreatedFrom != "" {
		q = q.Where("DATE(channel_accounts.created_at) >= ?", filter.CreatedFrom)
	}
	if filter.CreatedTo != "" {
		q = q.Where("DATE(channel_accounts.created_at) <= ?", filter.CreatedTo)
	}
	if filter.GroupID != nil && *filter.GroupID > 0 {
		q = q.Joins(`
			INNER JOIN channel_group_members cgm
				ON cgm.channel_account_id = channel_accounts.id
				AND cgm.group_id = ?
		`, *filter.GroupID)
	}
	if filter.AssignedUserID != nil {
		q = q.Where("channel_accounts.assigned_user_id = ?", *filter.AssignedUserID)
	}
	switch filter.ListFilter {
	case "unpaid":
		q = q.Where("channel_accounts.unpaid_closed = ?", true)
	case "restricted":
		q = q.Where("channel_accounts.restricted_closed = ?", true)
	case "closed8":
		q = q.Where("channel_accounts.cannot_open_at8 = ?", true)
	}
	if filter.ChannelName != "" {
		q = q.Joins("LEFT JOIN channels ON channels.id = channel_accounts.channel_id").
			Where("channels.name LIKE ?", "%"+filter.ChannelName+"%")
	}
	var list []model.ChannelAccount
	err := q.Order("channel_accounts.id DESC").Find(&list).Error
	return list, err
}

// ChannelGroupListFilter 通道分组列表筛选。
type ChannelGroupListFilter struct {
	ID   *uint
	Code string
}

// CreateChannelGroup 插入通道分组。
func (s *MySQLStore) CreateChannelGroup(item *model.ChannelGroup) error {
	return s.db.Create(item).Error
}

// SaveChannelGroup 更新通道分组。
func (s *MySQLStore) SaveChannelGroup(item *model.ChannelGroup) error {
	return s.db.Save(item).Error
}

// GetChannelGroupByID 按主键查通道分组。
func (s *MySQLStore) GetChannelGroupByID(id uint) (*model.ChannelGroup, error) {
	var item model.ChannelGroup
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindChannelGroupByCode 按分组 CODE 查询。
func (s *MySQLStore) FindChannelGroupByCode(code string) (*model.ChannelGroup, error) {
	var item model.ChannelGroup
	tx := s.db.Where("code = ?", code).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindChannelGroupByName 按分组名查询。
func (s *MySQLStore) FindChannelGroupByName(name string) (*model.ChannelGroup, error) {
	var item model.ChannelGroup
	tx := s.db.Where("name = ?", name).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListChannelGroups 按条件查询通道分组，按 id 倒序。
func (s *MySQLStore) ListChannelGroups(filter ChannelGroupListFilter) ([]model.ChannelGroup, error) {
	q := s.db.Model(&model.ChannelGroup{})
	if filter.ID != nil {
		q = q.Where("id = ?", *filter.ID)
	}
	if filter.Code != "" {
		q = q.Where("code LIKE ?", "%"+filter.Code+"%")
	}
	var list []model.ChannelGroup
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// CountEnabledChannelAccountsByGroupID 统计分组下启用的通道账号数。
func (s *MySQLStore) CountEnabledChannelAccountsByGroupID(groupID uint) (int64, error) {
	if groupID == 0 {
		return 0, nil
	}
	var count int64
	err := s.db.Model(&model.ChannelAccount{}).
		Joins(`
			INNER JOIN channel_group_members cgm
				ON cgm.channel_account_id = channel_accounts.id
				AND cgm.group_id = ?
		`, groupID).
		Where("channel_accounts.status = ?", model.ChannelAccountStatusEnabled).
		Count(&count).Error
	return count, err
}

// AddChannelGroupMember 添加分组与账号关系。
func (s *MySQLStore) AddChannelGroupMember(groupID, accountID uint) error {
	row := model.ChannelGroupMember{
		GroupID:          groupID,
		ChannelAccountID: accountID,
	}
	return s.db.Where("group_id = ? AND channel_account_id = ?", groupID, accountID).
		FirstOrCreate(&row).Error
}

// RemoveChannelGroupMember 移除分组与账号关系。
func (s *MySQLStore) RemoveChannelGroupMember(groupID, accountID uint) error {
	return s.db.Where("group_id = ? AND channel_account_id = ?", groupID, accountID).
		Delete(&model.ChannelGroupMember{}).Error
}

// ListChannelGroupMemberAccountIDs 查询分组下的账号 ID。
func (s *MySQLStore) ListChannelGroupMemberAccountIDs(groupID uint) ([]uint, error) {
	var accountIDs []uint
	err := s.db.Model(&model.ChannelGroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("channel_account_id", &accountIDs).Error
	return accountIDs, err
}

// ListChannelGroupMembersByAccountIDs 查询若干账号所属的分组关系。
func (s *MySQLStore) ListChannelGroupMembersByAccountIDs(accountIDs []uint) ([]model.ChannelGroupMember, error) {
	if len(accountIDs) == 0 {
		return []model.ChannelGroupMember{}, nil
	}
	var list []model.ChannelGroupMember
	err := s.db.Where("channel_account_id IN ?", accountIDs).Find(&list).Error
	return list, err
}

// StripeWordBankListFilter Stripe 单词库列表筛选。
type StripeWordBankListFilter struct {
	ConfigItem string
}

// CreateStripeWordBank 插入 Stripe 单词。
func (s *MySQLStore) CreateStripeWordBank(item *model.StripeWordBank) error {
	return s.db.Create(item).Error
}

// SaveStripeWordBank 更新 Stripe 单词。
func (s *MySQLStore) SaveStripeWordBank(item *model.StripeWordBank) error {
	return s.db.Save(item).Error
}

// GetStripeWordBankByID 按主键查 Stripe 单词。
func (s *MySQLStore) GetStripeWordBankByID(id uint) (*model.StripeWordBank, error) {
	var item model.StripeWordBank
	tx := s.db.Where("id = ?", id).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// FindStripeWordBankByName 按路径名查询。
func (s *MySQLStore) FindStripeWordBankByName(name string) (*model.StripeWordBank, error) {
	var item model.StripeWordBank
	tx := s.db.Where("name = ?", name).Limit(1).Find(&item)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return &item, nil
}

// ListStripeWordBanks 查询 Stripe 单词库，按 id 倒序。
func (s *MySQLStore) ListStripeWordBanks(filter StripeWordBankListFilter) ([]model.StripeWordBank, error) {
	q := s.db.Model(&model.StripeWordBank{})
	if filter.ConfigItem != "" {
		q = q.Where("config_item = ?", filter.ConfigItem)
	}
	var list []model.StripeWordBank
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// DeleteStripeWordBank 删除 Stripe 单词。
func (s *MySQLStore) DeleteStripeWordBank(id uint) error {
	return s.db.Delete(&model.StripeWordBank{}, id).Error
}
