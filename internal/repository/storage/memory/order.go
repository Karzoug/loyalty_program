package memory

import (
	"context"
	"sync"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
)

var _ storage.Order = (*orderStorage)(nil)

type orderStorage struct {
	orders map[order.Number]order.Order
	mu     *sync.RWMutex
}

func NewOrderStorage() *orderStorage {
	return &orderStorage{
		orders: make(map[order.Number]order.Order),
		mu:     &sync.RWMutex{},
	}
}

func (s orderStorage) Create(ctx context.Context, order order.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.orders[order.Number]
	if exists {
		return storage.ErrRecordAlreadyExists
	}
	s.orders[order.Number] = order

	return nil
}
func (s orderStorage) Get(ctx context.Context, number order.Number) (*order.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	o, exists := s.orders[number]
	if !exists {
		return nil, storage.ErrRecordNotFound
	}

	return &o, nil
}
func (s orderStorage) GetByUser(ctx context.Context, login user.Login) ([]order.Order, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]order.Order, 0)
	for _, v := range s.orders {
		if v.UserLogin == login {
			orders = append(orders, v)
		}
	}

	return orders, nil
}
func (s orderStorage) Update(ctx context.Context, order order.Order) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.orders[order.Number]
	if !exists {
		return storage.ErrRecordNotFound
	}
	s.orders[order.Number] = order

	return nil
}
