package memory

import (
	"context"
	"sync"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/shopspring/decimal"
)

var _ storage.User = (*userStorage)(nil)

type userStorage struct {
	users map[user.Login]user.User
	mu    *sync.RWMutex
}

func NewUserStorage() *userStorage {
	return &userStorage{
		users: make(map[user.Login]user.User),
		mu:    &sync.RWMutex{},
	}
}

func (s userStorage) Create(ctx context.Context, user user.User) error {
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

func (s userStorage) UpdateBalance(ctx context.Context, login user.Login, delta decimal.Decimal) (*decimal.Decimal, error) {
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

func (s userStorage) Get(ctx context.Context, login user.Login) (*user.User, error) {
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
