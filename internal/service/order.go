package service

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
)

func (s *Service) CreateOrder(ctx context.Context, login user.Login, orderNumber order.Number) (*order.Order, bool, error) {
	o, err := order.New(orderNumber, login)
	if err != nil {
		if errors.Is(err, order.ErrInvalidNumber) {
			return nil, false, ErrInvalidOrderNumber
		}
		return nil, false, err
	}

	err = s.storages.Order().Create(ctx, *o)
	if err != nil {
		if errors.Is(err, storage.ErrRecordAlreadyExists) {
			existedOrder, err := s.storages.Order().Get(ctx, orderNumber)
			if err != nil {
				return nil, false, err
			}
			if existedOrder.UserLogin != login {
				return nil, true, ErrAnotherUserOrderNumber
			}
			return existedOrder, true, nil
		}
		return nil, false, err
	}

	go s.processOrder(*o)

	return o, false, nil
}

func (s *Service) ListUserOrders(ctx context.Context, login user.Login) ([]order.Order, error) {
	ws, err := s.storages.Order().GetByUser(ctx, login)
	if err != nil {
		return nil, err
	}

	return ws, nil
}
