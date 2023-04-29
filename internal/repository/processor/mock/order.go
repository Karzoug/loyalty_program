package mock

import (
	"context"

	"github.com/Karzoug/loyalty_program/internal/model/order"
)

type Order struct {
	order *order.Order
	err   error
}

func NewOrder() *Order {
	return &Order{}
}

func (m *Order) SetResult(o *order.Order, err error) {
	m.order = o
	m.err = err
}

func (m *Order) Process(_ context.Context, _ order.Order) (*order.Order, error) {
	return m.order, m.err
}
