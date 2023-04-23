package service

import (
	"errors"
	"fmt"

	"github.com/Karzoug/loyalty_program/internal/model/user"
)

var (
	ErrLoginAlreadyExists    = errors.New("login already exists")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInvalidLoginFormat    = fmt.Errorf("invalid login format: must have (0; %d] UTF-8 characters count", user.MaxRuneCountInLogin)
	ErrInvalidPasswordFormat = errors.New("invalid password format: must have (0; 72] bytes UTF-8 characters")
	ErrInvalidAuthData       = errors.New("invalid login/password/token")

	ErrInvalidOrderNumber     = errors.New("invalid order number")
	ErrAnotherUserOrderNumber = errors.New("invalid order number: another user's order")
	ErrReAttemptWithdraw      = errors.New("re-attempt to withdraw")
)
