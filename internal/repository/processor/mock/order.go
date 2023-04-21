package mock

import (
	"context"

	morder "github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
)

var _ processor.Order = (*order)(nil)

type order struct {
}

func NewOrderProcessor() *order {
	return &order{}
}

func (p *order) Process(ctx context.Context, o morder.Order) <-chan processor.AcrualOrderResult {
	c := make(chan processor.AcrualOrderResult)
	go func() {
		defer close(c)

		// TODO: add some logic to test
	}()
	return c
}
