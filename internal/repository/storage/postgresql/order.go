package postgresql

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ storage.Order = (*OrderStorage)(nil)

type OrderStorage struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func NewOrderStorage(pool *pgxpool.Pool) *OrderStorage {
	return &OrderStorage{
		pool: pool,
	}
}

func newOrderTxStorage(tx pgx.Tx) *OrderStorage {
	return &OrderStorage{
		tx: tx,
	}
}

func (s OrderStorage) connection() pgConnecter {
	if s.tx == nil {
		return s.pool
	}
	return s.tx
}

func (s OrderStorage) Create(ctx context.Context, order order.Order) error {
	_, err := s.connection().Exec(ctx, `INSERT INTO orders(number, user_login, status, accrual, uploaded_at) VALUES($1, $2, $3, $4, $5)`,
		order.Number, order.UserLogin, order.Status, order.Accrual, order.UploadedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicateKeyErrorCode {
			return storage.ErrRecordAlreadyExists
		}
		return err
	}

	return nil
}

func (s OrderStorage) Get(ctx context.Context, number order.Number) (*order.Order, error) {
	order := order.Order{Number: number}
	err := s.connection().QueryRow(ctx,
		`SELECT user_login, status, accrual, uploaded_at FROM orders WHERE number = $1`, number).
		Scan(&order.UserLogin, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (s OrderStorage) GetByUser(ctx context.Context, login string) ([]order.Order, error) {
	rows, err := s.connection().Query(ctx,
		`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_login = $1`, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (order.Order, error) {
		order := order.Order{UserLogin: login}
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		return order, err
	})

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s OrderStorage) Update(ctx context.Context, order order.Order) error {
	_, err := s.connection().Exec(ctx,
		`UPDATE orders SET user_login = $1, status = $2, accrual = $3, uploaded_at = $4 WHERE number = $5`,
		order.UserLogin, order.Status, order.Accrual, order.UploadedAt, order.Number)
	if err != nil {
		if err == pgx.ErrNoRows {
			return storage.ErrRecordNotFound
		}
		return err
	}

	return nil
}
