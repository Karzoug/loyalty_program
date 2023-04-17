package service

import (
	"errors"
	"fmt"

	"github.com/Karzoug/loyalty_program/internal/model/user"
)

var (
	ErrLoginAlreadyExists    = errors.New("login already exists")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInvalidLoginFormat    = fmt.Errorf("invalid login format: must be not more than %d UTF-8 characters", user.MaxRuneCountInLogin)
	ErrInvalidPasswordFormat = errors.New("invalid password format: must not exceed 72 bytes UTF-8 characters")
	ErrInvalidAuthData       = errors.New("invalid login and/or password and/or token")

	ErrInvalidOrderNumber     = errors.New("invalid order number")
	ErrAnotherUserOrderNumber = errors.New("invalid order number: another user's order")
)
