package store

import (
	"gorm.io/driver/mysql"

	"gorm.io/gorm"
	"paymentcenter/internal/model"
)

// MySQL 存储
type MySQLStore struct {
	db *gorm.DB
}

// 创建 MySQL 存储
func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.Order{}, &model.User{}); err != nil {
		return nil, err
	}
	return &MySQLStore{db: db}, nil
}

// 关闭 MySQL 连接
func (s *MySQLStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// 保存订单
func (s *MySQLStore) SaveOrder(order *model.Order) error {
	return s.db.Save(order).Error
}

// 获取订单
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

// 创建用户
func (s *MySQLStore) CreateUser(user *model.User) error {
	return s.db.Create(user).Error
}

// 保存用户
func (s *MySQLStore) SaveUser(user *model.User) error {
	return s.db.Save(user).Error
}

// 根据 ID 获取用户
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

// 根据用户名获取用户
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

// 获取订单列表
func (s *MySQLStore) ListOrders() ([]*model.Order, error) {
	var orders []*model.Order
	if err := s.db.Order("created_at DESC").Limit(100).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
