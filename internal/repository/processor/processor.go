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

type AcrualOrderResult struct {
	Order *order.Order
	Err   error
}

type Order interface {
	Process(context.Context, order.Order) <-chan AcrualOrderResult
}
