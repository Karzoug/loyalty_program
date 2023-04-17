package order

import (
	"errors"
	"strconv"

	"github.com/Karzoug/loyalty_program/pkg/luhn"
)

var (
	ErrInvalidNumber = errors.New("invalid order number")
)

type Number string

func (n Number) Valid() bool {
	number, err := strconv.ParseInt(string(n), 10, 64)
	if err != nil {
		return false
	}
	return luhn.Valid(number)
}
