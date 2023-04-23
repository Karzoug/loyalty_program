package storage

import (
	"context"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/shopspring/decimal"
)

type User interface {
	Create(context.Context, user.User) error
	Get(context.Context, user.Login) (*user.User, error)
	UpdateBalance(ctx context.Context, login user.Login, deltaBalance decimal.Decimal) (*decimal.Decimal, error)
}

type Order interface {
	Create(context.Context, order.Order) error
	Get(context.Context, order.Number) (*order.Order, error)
	GetByUser(context.Context, user.Login) ([]order.Order, error)
	// ListUnprocessed returns limit (-1 is a special value: no limit) orders not yet processed.
	ListUnprocessed(ctx context.Context, limit, offset int, uploadedEarlierThan time.Time) ([]order.Order, error)
	Update(context.Context, order.Order) error
	Delete(context.Context, order.Number) error
}

type Withdraw interface {
	Create(context.Context, withdraw.Withdraw) error
	//Get(ctx context.Context, orderNumber order.Number) error
	GetByUser(context.Context, user.Login) ([]withdraw.Withdraw, error)
	CountByUser(context.Context, user.Login) (int, error)
	SumByUser(context.Context, user.Login) (*decimal.Decimal, error)
	Update(context.Context, withdraw.Withdraw) error
}
