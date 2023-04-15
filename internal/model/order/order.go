package order

import (
	"time"

	"github.com/shopspring/decimal"
)

type Number string

type Order struct {
	Number    Number
	UserLogin string
	Status    status
	Accrual   decimal.Decimal

	UploadedAt time.Time
}

// New creates a new Order, ready to be processed and inserted into repository.
func New(number Number, userLogin string) (*Order, error) {
	order := Order{
		Number:    number,
		UserLogin: userLogin,
		Status:    StatusNew,
		Accrual:   decimal.Decimal{},

		UploadedAt: time.Now().UTC(),
	}

	return &order, nil
}
