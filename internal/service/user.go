package service

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/shopspring/decimal"
)

func (s *Service) RegisterUser(ctx context.Context, login, password string) (*user.User, error) {
	u, err := user.New(login, password)
	if err != nil {
		if errors.Is(err, user.ErrBadPassword) {
			return nil, ErrInvalidPasswordFormat
		}
		return nil, err
	}

	err = s.storages.User().Create(ctx, *u)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			return nil, ErrLoginAlreadyExists
		}
		return nil, err
	}
	return u, nil
}

func (s *Service) LoginUser(ctx context.Context, login, password string) (*user.User, error) {
	u, err := s.storages.User().Get(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return nil, ErrInvalidAuthData
		}
		return nil, err
	}
	if !u.VerifyPassword(password) {
		return nil, ErrInvalidAuthData
	}

	return u, nil
}

func (s *Service) GetUserBalance(ctx context.Context, login string) (*decimal.Decimal, error) {
	u, err := s.storages.User().Get(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return nil, ErrInvalidAuthData
		}
		return nil, err
	}

	return &u.Balance, nil
}
