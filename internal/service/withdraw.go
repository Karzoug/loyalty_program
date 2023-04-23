package service

import (
	"context"
	"errors"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/shopspring/decimal"
)

func (s *Service) CreateWithdraw(ctx context.Context, login user.Login, orderNumber order.Number, sum decimal.Decimal) (*withdraw.Withdraw, error) {
	w, err := withdraw.New(login, orderNumber, sum)
	if err != nil {
		if errors.Is(err, order.ErrInvalidNumber) {
			return nil, ErrInvalidOrderNumber
		}
		return nil, err
	}

	// since checks from external services are not required,
	// we can set the processed time to the current
	w.ProcessedAt = time.Now().UTC()

	tx, err := s.storages.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result, err := tx.User().UpdateBalance(ctx, login, sum.Neg())
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return nil, ErrInvalidAuthData
		}
		return nil, err
	}
	// balance went negative
	if result.LessThan(decimal.Decimal{}) {
		return nil, ErrInsufficientBalance
	}

	err = tx.Withdraw().Create(ctx, *w)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			existedWithdraw, err := s.storages.Withdraw().Get(ctx, orderNumber)
			if err != nil {
				return nil, err
			}
			if existedWithdraw.UserLogin != login {
				return nil, ErrAnotherUserOrderNumber
			}
			return existedWithdraw, ErrReAttemptWithdraw
		}
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Service) ListUserWithdrawals(ctx context.Context, login user.Login) ([]withdraw.Withdraw, error) {
	ws, err := s.storages.Withdraw().GetByUser(ctx, login)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *Service) SumUserWithdrawals(ctx context.Context, login user.Login) (*decimal.Decimal, error) {
	sum, err := s.storages.Withdraw().SumByUser(ctx, login)
	if err != nil {
		return nil, err
	}

	return sum, nil
}
