package mock

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/model/withdraw"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/shopspring/decimal"
)

var _ storage.Withdraw = (*withdrawStorage)(nil)

type withdrawStorage struct {
	db *sql.DB
	tx *sql.Tx
}

func NewWithdrawStorage(db *sql.DB) *withdrawStorage {
	return &withdrawStorage{
		db: db,
	}
}

func newWithdrawTxStorage(tx *sql.Tx) *withdrawStorage {
	return &withdrawStorage{
		tx: tx,
	}
}

func (s withdrawStorage) connection() sqliteConnecter {
	if s.tx == nil {
		return s.db
	}
	return s.tx
}

func (s withdrawStorage) Create(ctx context.Context, withdraw withdraw.Withdraw) error {
	res, err := s.connection().ExecContext(ctx, `INSERT INTO withdrawals(order_number, user_login, sum, processed_at) VALUES(?, ?, ?, ?)`,
		withdraw.OrderNumber, withdraw.UserLogin, withdraw.Sum, withdraw.ProcessedAt)
	if err != nil {
		if strings.Contains(err.Error(), duplicateKeyErrorCode) {
			return storage.ErrRecordAlreadyExists
		}
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return storage.ErrNoRecordAffected
	}

	return nil
}

func (s withdrawStorage) GetByUser(ctx context.Context, login user.Login) ([]withdraw.Withdraw, error) {
	rows, err := s.connection().QueryContext(ctx,
		`SELECT order_number, sum, processed_at FROM withdrawals WHERE user_login = ?`, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]withdraw.Withdraw, 0)
	for rows.Next() {
		withdraw := withdraw.Withdraw{UserLogin: login}
		err := rows.Scan(&withdraw.OrderNumber, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdraw)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (s withdrawStorage) CountByUser(ctx context.Context, login user.Login) (int, error) {
	var count int
	err := s.connection().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM withdrawals WHERE user_login = ?`, login).Scan(&count)

	return count, err
}

func (s withdrawStorage) SumByUser(ctx context.Context, login user.Login) (*decimal.Decimal, error) {
	var sum decimal.NullDecimal
	err := s.connection().QueryRowContext(ctx,
		`SELECT SUM(sum) FROM withdrawals WHERE user_login = ?`, login).Scan(&sum)

	if !sum.Valid {
		return &decimal.Zero, nil
	}

	return &sum.Decimal, err
}

func (s withdrawStorage) Update(ctx context.Context, withdraw withdraw.Withdraw) error {
	res, err := s.connection().ExecContext(ctx,
		`UPDATE withdrawals SET user_login = ?, sum = ?, processed_at = ? WHERE order_number = ?`,
		withdraw.UserLogin, withdraw.Sum, withdraw.ProcessedAt, withdraw.OrderNumber)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return storage.ErrNoRecordAffected
	}

	return nil
}
