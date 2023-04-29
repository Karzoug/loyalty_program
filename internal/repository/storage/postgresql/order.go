package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ storage.Order = (*orderStorage)(nil)

type orderStorage struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func newOrderStorage(pool *pgxpool.Pool) *orderStorage {
	return &orderStorage{
		pool: pool,
	}
}

func newOrderTxStorage(tx pgx.Tx) *orderStorage {
	return &orderStorage{
		tx: tx,
	}
}

func (s orderStorage) connection() pgConnecter {
	if s.tx == nil {
		return s.pool
	}
	return s.tx
}

func (s orderStorage) Create(ctx context.Context, order order.Order) error {
	tag, err := s.connection().Exec(ctx, `INSERT INTO orders(number, user_login, status, accrual, uploaded_at) VALUES($1, $2, $3, $4, $5)`,
		order.Number, order.UserLogin, order.Status, order.Accrual, order.UploadedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicateKeyErrorCode {
			return storage.ErrRecordAlreadyExists
		}
		return err
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrNoRecordAffected
	}

	return nil
}

func (s orderStorage) Get(ctx context.Context, number order.Number) (*order.Order, error) {
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

func (s orderStorage) GetByUser(ctx context.Context, login user.Login) ([]order.Order, error) {
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
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s orderStorage) ListUnprocessed(ctx context.Context, limit, offset int, uploadedEarlierThan time.Time) ([]order.Order, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if limit == -1 {
		rows, err = s.connection().Query(ctx, `SELECT number, user_login, status, accrual, uploaded_at FROM orders WHERE status NOT IN ($1, $2) AND uploaded_at < $3 ORDER BY uploaded_at OFFSET $4`, order.StatusInvalid, order.StatusProcessed, uploadedEarlierThan, offset)
	} else {
		rows, err = s.connection().Query(ctx, `SELECT number, user_login, status, accrual, uploaded_at FROM orders WHERE status NOT IN ($1, $2) AND uploaded_at < $3 ORDER BY uploaded_at LIMIT $4 OFFSET $5`, order.StatusInvalid, order.StatusProcessed, uploadedEarlierThan, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (order.Order, error) {
		var order order.Order
		err := rows.Scan(&order.Number, &order.UserLogin, &order.Status, &order.Accrual, &order.UploadedAt)
		return order, err
	})
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s orderStorage) Update(ctx context.Context, order order.Order) error {
	tag, err := s.connection().Exec(ctx,
		`UPDATE orders SET user_login = $1, status = $2, accrual = $3, uploaded_at = $4 WHERE number = $5`,
		order.UserLogin, order.Status, order.Accrual, order.UploadedAt, order.Number)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrNoRecordAffected
	}

	return nil
}

func (s orderStorage) Delete(ctx context.Context, number order.Number) error {
	tag, err := s.connection().Exec(ctx, `DELETE FROM orders WHERE number = $1`, number)
	if err != nil {
		if err == pgx.ErrNoRows {
			return storage.ErrRecordNotFound
		}
		return err
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrNoRecordAffected
	}

	return nil
}
