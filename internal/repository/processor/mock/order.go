package mock

import (
	"context"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
)

type Order struct {
	ch chan processor.AcrualOrderResult
}

func NewOrder(ch chan processor.AcrualOrderResult) *Order {
	return &Order{
		ch: ch,
	}
}

func (m *Order) Process(ctx context.Context, o order.Order) <-chan processor.AcrualOrderResult {
	return m.ch
}
