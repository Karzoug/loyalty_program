package service

import "errors"

var (
	ErrLoginAlreadyExists    = errors.New("login already exists")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrInvalidPasswordFormat = errors.New("invalid password format: must not be longer than 72 ASCII characters")
	ErrInvalidAuthData       = errors.New("invalid login and/or password and/or token")

	ErrInvalidOrderNumber     = errors.New("invalid order number")
	ErrAnotherUserOrderNumber = errors.New("invalid order number: another user's order")
)
