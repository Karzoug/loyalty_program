package memory

import (
	"context"
	"sync"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
)

var _ storage.Withdraw = (*WithdrawStorage)(nil)

type WithdrawStorage struct {
	Withdrawals map[order.Number]withdraw.Withdraw
	mu          *sync.RWMutex
}

func NewWithdrawStorage() *WithdrawStorage {
	return &WithdrawStorage{
		Withdrawals: make(map[order.Number]withdraw.Withdraw),
		mu:          &sync.RWMutex{},
	}
}

func (s WithdrawStorage) Create(ctx context.Context, withdraw withdraw.Withdraw) error {
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

func (s WithdrawStorage) GetByUser(ctx context.Context, login string) ([]withdraw.Withdraw, error) {
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

func (s WithdrawStorage) Update(ctx context.Context, withdraw withdraw.Withdraw) error {
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
