package store

import (
	"errors"

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
	if err := db.AutoMigrate(&model.Order{}); err != nil {
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
func (s *MySQLStore) Save(order *model.Order) error {
	return s.db.Save(order).Error
}

// 获取订单
func (s *MySQLStore) Get(id string) (*model.Order, error) {
	var order model.Order
	if err := s.db.First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

// 获取订单列表
func (s *MySQLStore) List() ([]*model.Order, error) {
	var orders []*model.Order
	if err := s.db.Order("created_at DESC").Limit(100).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
