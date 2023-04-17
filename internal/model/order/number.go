package order

import (
	"errors"

	"github.com/Karzoug/loyalty_program/pkg/luhn"
)

var (
	ErrInvalidNumber = errors.New("invalid order number")
)

type Number int64

func (n Number) Valid() bool {
	return luhn.Valid(int64(n))
}
