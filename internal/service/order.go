package service

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
)

func (s *Service) CreateOrder(ctx context.Context, userLogin string, orderNumber order.Number) (*order.Order, bool, error) {
	o, err := order.New(orderNumber, userLogin)
	if err != nil {
		if errors.Is(err, order.ErrInvalidOrderNumber) {
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
			if existedOrder.UserLogin != userLogin {
				return nil, true, ErrAnotherUserOrderNumber
			}
			return existedOrder, true, nil
		}
		return nil, false, err
	}

	// TODO: add interaction with the system for calculating bonuses

	return o, false, nil
}

func (s *Service) ListUserOrders(ctx context.Context, userLogin string) ([]order.Order, error) {
	ws, err := s.storages.Order().GetByUser(ctx, userLogin)
	if err != nil {
		return nil, err
	}

	return ws, nil
}
