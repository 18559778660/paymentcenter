package store

import (
	"errors"
	"sync"

	"paymentcenter/internal/domain"
)

var ErrNotFound = errors.New("order not found")

// 内存存储
type MemoryStore struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

// 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		orders: make(map[string]*domain.Order),
	}
}

// 保存订单
func (s *MemoryStore) Save(order *domain.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
}

// 获取订单
func (s *MemoryStore) Get(id string) (*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	return order, nil
}

// 获取订单列表
func (s *MemoryStore) List() []*domain.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*domain.Order, 0, len(s.orders))
	for _, order := range s.orders {
		out = append(out, order)
	}
	return out
}
