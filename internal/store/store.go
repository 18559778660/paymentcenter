package store

import (
	"fmt"

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
		&model.Channel{},
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
	ID   *uint
	Name string
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
	var list []model.Channel
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}
