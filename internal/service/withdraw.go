package service

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/shopspring/decimal"
)

func (s *Service) CreateWithdraw(ctx context.Context, userLogin string, orderNumber order.Number, sum decimal.Decimal) (*withdraw.Withdraw, error) {
	w, err := withdraw.New(userLogin, orderNumber, sum)
	if err != nil {
		if errors.Is(err, order.ErrInvalidOrderNumber) {
			return nil, ErrInvalidOrderNumber
		}
		return nil, err
	}

	// TODO: add transaction
	result, err := s.storages.User().UpdateBalance(ctx, userLogin, sum.Neg())
	if err != nil {
		return nil, err
	}
	// balance went negative
	if result.LessThan(decimal.Decimal{}) {
		// TODO: rollback transaction
		return nil, ErrInsufficientBalance
	}

	err = s.storages.Withdraw().Create(ctx, *w)
	if err != nil {
		// TODO: rollback transaction
		return nil, err
	}

	// TODO: commit transaction
	return w, nil
}

func (s *Service) ListUserWithdrawals(ctx context.Context, userLogin string) ([]withdraw.Withdraw, error) {
	ws, err := s.storages.Withdraw().GetByUser(ctx, userLogin)
	if err != nil {
		return nil, err
	}

	return ws, nil
}
