package store

import (
	"errors"
	"sync"

	"paymentcenter/internal/domain"
)

var ErrNotFound = errors.New("order not found")

type MemoryStore struct {
	mu     sync.RWMutex
	orders  map[string]*domain.Order
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		orders: make(map[string]*domain.Order),
	}
}

func (s *MemoryStore) Save(order *domain.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
}

func (s *MemoryStore) Get(id string) (*domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	return order, nil
}

func (s *MemoryStore) List() []*domain.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*domain.Order, 0, len(s.orders))
	for _, order := range s.orders {
		out = append(out, order)
	}
	return out
}
