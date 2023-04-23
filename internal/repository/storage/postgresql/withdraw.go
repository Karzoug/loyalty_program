package postgresql

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var _ storage.Withdraw = (*withdrawStorage)(nil)

type withdrawStorage struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func newWithdrawStorage(pool *pgxpool.Pool) *withdrawStorage {
	return &withdrawStorage{
		pool: pool,
	}
}

func newWithdrawTxStorage(tx pgx.Tx) *withdrawStorage {
	return &withdrawStorage{
		tx: tx,
	}
}

func (s withdrawStorage) connection() pgConnecter {
	if s.tx == nil {
		return s.pool
	}
	return s.tx
}

func (s withdrawStorage) Create(ctx context.Context, withdraw withdraw.Withdraw) error {
	_, err := s.connection().Exec(ctx, `INSERT INTO withdrawals(order_number, user_login, sum, processed_at) VALUES($1, $2, $3, $4)`,
		withdraw.OrderNumber, withdraw.UserLogin, withdraw.Sum, withdraw.ProcessedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicateKeyErrorCode {
			return storage.ErrRecordAlreadyExists
		}
		return err
	}

	return nil
}

func (s withdrawStorage) Get(ctx context.Context, number order.Number) (*withdraw.Withdraw, error) {
	w := withdraw.Withdraw{OrderNumber: number}
	err := s.connection().QueryRow(ctx,
		`SELECT sum, processed_at, user_login FROM withdrawals WHERE order_number = $1`, number).
		Scan(&w.Sum, &w.ProcessedAt, &w.UserLogin)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}

	return &w, nil
}

func (s withdrawStorage) GetByUser(ctx context.Context, login user.Login) ([]withdraw.Withdraw, error) {
	rows, err := s.connection().Query(ctx,
		`SELECT order_number, sum, processed_at FROM withdrawals WHERE user_login = $1`, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (withdraw.Withdraw, error) {
		withdraw := withdraw.Withdraw{UserLogin: login}
		err := rows.Scan(&withdraw.OrderNumber, &withdraw.Sum, &withdraw.ProcessedAt)
		return withdraw, err
	})
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (s withdrawStorage) CountByUser(ctx context.Context, login user.Login) (int, error) {
	var count int
	err := s.connection().QueryRow(ctx,
		`SELECT COUNT(*) FROM withdrawals WHERE user_login = $1`, login).Scan(&count)

	return count, err
}

func (s withdrawStorage) SumByUser(ctx context.Context, login user.Login) (*decimal.Decimal, error) {
	var sum decimal.NullDecimal
	err := s.connection().QueryRow(ctx,
		`SELECT SUM(sum) FROM withdrawals WHERE user_login = $1`, login).Scan(&sum)

	if !sum.Valid {
		return &decimal.Zero, nil
	}

	return &sum.Decimal, err
}

func (s withdrawStorage) Update(ctx context.Context, withdraw withdraw.Withdraw) error {
	_, err := s.connection().Exec(ctx,
		`UPDATE withdrawals SET user_login = $1, sum = $2, processed_at = $3 WHERE order_number = $4`,
		withdraw.UserLogin, withdraw.Sum, withdraw.ProcessedAt, withdraw.OrderNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return storage.ErrRecordNotFound
		}
		return err
	}

	return nil
}
