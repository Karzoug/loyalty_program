package withdraw

import (
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/shopspring/decimal"
)

type Withdraw struct {
	OrderNumber order.Number
	UserLogin   user.Login
	Sum         decimal.Decimal

	ProcessedAt time.Time
}

func New(login user.Login, orderNumber order.Number, sum decimal.Decimal) (*Withdraw, error) {
	w := Withdraw{
		OrderNumber: orderNumber,
		Sum:         sum,
		UserLogin:   login,
	}

	if !w.OrderNumber.Valid() {
		return nil, order.ErrInvalidNumber
	}

	return &w, nil
}
