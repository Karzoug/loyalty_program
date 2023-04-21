package postgresql

import (
	"context"

	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/Karzoug/loyalty_program/pkg/e"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type storages struct {
	pool *pgxpool.Pool

	userStorage     storage.User
	orderStorage    storage.Order
	withdrawStorage storage.Withdraw
}

// NewStorages returns a set of storages for the service to work with data.
func NewStorages(ctx context.Context, cfg configPostgreSQLStorage) (*storages, error) {
	pool, err := newDBPool(ctx, cfg)
	if err != nil {
		return nil, e.Wrap("open postgresql db connection", err)
	}

	return &storages{
		pool:            pool,
		userStorage:     NewUserStorage(pool),
		orderStorage:    NewOrderStorage(pool),
		withdrawStorage: NewWithdrawStorage(pool),
	}, nil
}

func (r *storages) BeginTx(ctx context.Context) (storage.Transaction, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
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
	tx pgx.Tx

	userStorage     storage.User
	orderStorage    storage.Order
	withdrawStorage storage.Withdraw
}

func (t *transaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}
func (t *transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
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
