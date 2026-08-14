package store

import (
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

// ListMenus 查询全部启用中的菜单。
func (s *MySQLStore) ListMenus() ([]model.Menu, error) {
	var menus []model.Menu
	err := s.db.Where("status = ?", model.MenuStatusEnabled).
		Order("sort ASC, id ASC").
		Find(&menus).Error
	return menus, err
}

// SaveMenu 更新菜单字段。
func (s *MySQLStore) SaveMenu(menu *model.Menu) error {
	return s.db.Save(menu).Error
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
