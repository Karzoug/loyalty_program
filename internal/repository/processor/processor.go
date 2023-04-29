package processor

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/order"
)

var (
	ErrOrderNotRegistered = errors.New("order is not registered")
	ErrServerNotRespond   = errors.New("server not respond")
)

type Order interface {
	Process(context.Context, order.Order) (*order.Order, error)
}
