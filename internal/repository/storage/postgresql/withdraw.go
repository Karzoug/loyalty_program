package postgresql

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ storage.Withdraw = (*WithdrawStorage)(nil)

type WithdrawStorage struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func NewWithdrawStorage(pool *pgxpool.Pool) *WithdrawStorage {
	return &WithdrawStorage{
		pool: pool,
	}
}

func newWithdrawTxStorage(tx pgx.Tx) *WithdrawStorage {
	return &WithdrawStorage{
		tx: tx,
	}
}

func (s WithdrawStorage) connection() pgConnecter {
	if s.tx == nil {
		return s.pool
	}
	return s.tx
}

func (s WithdrawStorage) Create(ctx context.Context, withdraw withdraw.Withdraw) error {
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

func (s WithdrawStorage) GetByUser(ctx context.Context, login user.Login) ([]withdraw.Withdraw, error) {
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

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (s WithdrawStorage) CountByUser(ctx context.Context, login user.Login) (int, error) {
	var count int
	err := s.connection().QueryRow(ctx,
		`SELECT COUNT(*) FROM withdrawals WHERE user_login = $1`, login).Scan(&count)

	return count, err
}

func (s WithdrawStorage) Update(ctx context.Context, withdraw withdraw.Withdraw) error {
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
