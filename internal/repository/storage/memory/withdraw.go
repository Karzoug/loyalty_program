package memory

import (
	"context"
	"sync"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
)

var _ storage.Withdraw = (*withdrawStorage)(nil)

type withdrawStorage struct {
	Withdrawals map[order.Number]withdraw.Withdraw
	mu          *sync.RWMutex
}

func NewWithdrawStorage() *withdrawStorage {
	return &withdrawStorage{
		Withdrawals: make(map[order.Number]withdraw.Withdraw),
		mu:          &sync.RWMutex{},
	}
}

func (s withdrawStorage) Create(ctx context.Context, withdraw withdraw.Withdraw) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.Withdrawals[withdraw.OrderNumber]
	if exists {
		return storage.ErrRecordAlreadyExists
	}
	s.Withdrawals[withdraw.OrderNumber] = withdraw

	return nil
}

func (s withdrawStorage) GetByUser(ctx context.Context, login user.Login) ([]withdraw.Withdraw, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	withdrawals := make([]withdraw.Withdraw, 0)
	for _, v := range s.Withdrawals {
		if v.UserLogin == login {
			withdrawals = append(withdrawals, v)
		}
	}

	return withdrawals, nil
}

func (s withdrawStorage) CountByUser(ctx context.Context, login user.Login) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	for _, v := range s.Withdrawals {
		if v.UserLogin == login {
			count++
		}
	}

	return count, nil
}

func (s withdrawStorage) Update(ctx context.Context, withdraw withdraw.Withdraw) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.Withdrawals[withdraw.OrderNumber]
	if !exists {
		return storage.ErrRecordNotFound
	}
	s.Withdrawals[withdraw.OrderNumber] = withdraw

	return nil
}
