package withdraw

import (
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/shopspring/decimal"
)

type Withdraw struct {
	OrderNumber order.Number
	UserLogin   string
	Sum         decimal.Decimal

	ProcessedAt time.Time
}

func New(userLogin string, orderNumber order.Number, sum decimal.Decimal) (*Withdraw, error) {
	w := Withdraw{
		OrderNumber: orderNumber,
		Sum:         sum,
		UserLogin:   userLogin,
	}

	return &w, nil
}
