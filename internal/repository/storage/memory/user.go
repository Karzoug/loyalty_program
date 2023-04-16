package memory

import (
	"context"
	"sync"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/shopspring/decimal"
)

var _ storage.User = (*UserStorage)(nil)

type UserStorage struct {
	users map[string]user.User
	mu    *sync.RWMutex
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		users: make(map[string]user.User),
		mu:    &sync.RWMutex{},
	}
}

func (s UserStorage) Create(ctx context.Context, user user.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[user.Login]
	if exists {
		return storage.ErrRecordAlreadyExists
	}
	s.users[user.Login] = user

	return nil
}

func (s UserStorage) UpdateBalance(ctx context.Context, login string, delta decimal.Decimal) (*decimal.Decimal, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	u, exists := s.users[login]
	if !exists {
		return nil, storage.ErrRecordNotFound
	}
	u.Balance = u.Balance.Add(delta)
	s.users[login] = u

	return &u.Balance, nil
}

func (s UserStorage) Get(ctx context.Context, login string) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	u, exists := s.users[login]
	if !exists {
		return nil, storage.ErrRecordNotFound
	}

	return &u, nil
}
