package store

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"paymentcenter/internal/domain"
)

type MySQLStore struct {
	db *gorm.DB
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&domain.Order{}); err != nil {
		return nil, err
	}
	return &MySQLStore{db: db}, nil
}

func (s *MySQLStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *MySQLStore) Save(order *domain.Order) error {
	return s.db.Save(order).Error
}

func (s *MySQLStore) Get(id string) (*domain.Order, error) {
	var order domain.Order
	if err := s.db.First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (s *MySQLStore) List() ([]*domain.Order, error) {
	var orders []*domain.Order
	if err := s.db.Order("created_at DESC").Limit(100).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}
