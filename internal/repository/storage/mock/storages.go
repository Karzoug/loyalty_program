package mock

import (
	"context"
	"database/sql"

	"github.com/Karzoug/loyalty_program/internal/repository/storage"
)

var _ storage.TxStorages = (*storages)(nil)

type storages struct {
	db *sql.DB

	userStorage     storage.User
	orderStorage    storage.Order
	withdrawStorage storage.Withdraw
}

// NewStorages creates a mock storage for testing tasks, does not support real transactions.
func NewStorages(db *sql.DB) *storages {
	return &storages{
		db:              db,
		userStorage:     NewUserStorage(db),
		orderStorage:    NewOrderStorage(db),
		withdrawStorage: NewWithdrawStorage(db),
	}
}

func (r *storages) BeginTx(ctx context.Context) (storage.Transaction, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &transaction{
		tx:              tx,
		userStorage:     newUserTxStorage(tx),
		orderStorage:    newOrderTxStorage(tx),
		withdrawStorage: newWithdrawTxStorage(tx),
	}, nil
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

type transaction struct {
	tx *sql.Tx

	userStorage     storage.User
	orderStorage    storage.Order
	withdrawStorage storage.Withdraw
}

func (t *transaction) Commit(ctx context.Context) error {
	return t.tx.Commit()
}
func (t *transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// User return user storage with transaction.
func (t *transaction) User() storage.User {
	return t.userStorage
}

// Order return order storage with transaction.
func (t *transaction) Order() storage.Order {
	return t.orderStorage
}

// Withdraw return withdraw storage with transaction.
func (t *transaction) Withdraw() storage.Withdraw {
	return t.withdrawStorage
}
