package storage

import "context"

type Storages interface {
	User() User
	Order() Order
	Withdraw() Withdraw
}

type TxStorages interface {
	Storages
	BeginTx(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Storages
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
