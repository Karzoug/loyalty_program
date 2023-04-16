package memory

import "github.com/Karzoug/loyalty_program/internal/repository/storage"

var _ storage.Storages = (*storages)(nil)

type storages struct {
	userStorage     storage.User
	orderStorage    storage.Order
	withdrawStorage storage.Withdraw
}

func NewStorages() *storages {
	return &storages{
		userStorage:     NewUserStorage(),
		orderStorage:    NewOrderStorage(),
		withdrawStorage: NewWithdrawStorage(),
	}
}

// User return user storage.
func (r *storages) User() storage.User {
	return r.userStorage
}

// Order return order storage.
func (r *storages) Order() storage.Order {
	return r.orderStorage
}

// Withdraw return withdraw storage.
func (r *storages) Withdraw() storage.Withdraw {
	return r.withdrawStorage
}
