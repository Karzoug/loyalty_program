package storage

import (
	"context"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/shopspring/decimal"
)

type User interface {
	Create(context.Context, user.User) error
	Get(ctx context.Context, login string) (*user.User, error)
	UpdateBalance(ctx context.Context, login string, deltaBalance decimal.Decimal) (*decimal.Decimal, error)
}

type Order interface {
	Create(context.Context, order.Order) error
	Get(context.Context, order.Number) (*order.Order, error)
	GetByUser(context context.Context, login string) ([]order.Order, error)
	Update(context.Context, order.Order) error
}

type Withdraw interface {
	Create(context.Context, withdraw.Withdraw) error
	//Get(ctx context.Context, orderNumber order.Number) error
	GetByUser(ctx context.Context, login string) ([]withdraw.Withdraw, error)
	Update(context.Context, withdraw.Withdraw) error
}
