package order

import (
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/shopspring/decimal"
)

type Order struct {
	Number    Number
	UserLogin user.Login
	Status    status
	Accrual   decimal.Decimal

	UploadedAt time.Time
}

// New creates a new Order, ready to be processed and inserted into repository.
func New(number Number, login user.Login) (*Order, error) {
	if !login.Valid() {
		return nil, user.ErrInvalidLogin
	}

	order := Order{
		Number:    number,
		UserLogin: login,
		Status:    StatusNew,
		Accrual:   decimal.Decimal{},

		UploadedAt: time.Now().UTC(),
	}

	if !order.Number.Valid() {
		return nil, ErrInvalidNumber
	}

	return &order, nil
}
