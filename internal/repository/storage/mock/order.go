package mock

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
)

var _ storage.Order = (*orderStorage)(nil)

type orderStorage struct {
	db *sql.DB
	tx *sql.Tx
}

func NewOrderStorage(db *sql.DB) *orderStorage {
	return &orderStorage{
		db: db,
	}
}

func newOrderTxStorage(tx *sql.Tx) *orderStorage {
	return &orderStorage{
		tx: tx,
	}
}

func (s orderStorage) connection() sqliteConnecter {
	if s.tx == nil {
		return s.db
	}
	return s.tx
}

func (s orderStorage) Create(ctx context.Context, order order.Order) error {
	res, err := s.connection().ExecContext(ctx, `INSERT INTO orders(number, user_login, status, accrual, uploaded_at) VALUES(?, ?, ?, ?, ?)`,
		order.Number, order.UserLogin, order.Status, order.Accrual, order.UploadedAt)
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

func (s orderStorage) Get(ctx context.Context, number order.Number) (*order.Order, error) {
	order := order.Order{Number: number}
	err := s.connection().QueryRowContext(ctx,
		`SELECT user_login, status, accrual, uploaded_at FROM orders WHERE number = ?`, number).
		Scan(&order.UserLogin, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (s orderStorage) GetByUser(ctx context.Context, login user.Login) ([]order.Order, error) {
	rows, err := s.connection().QueryContext(ctx,
		`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_login = ?`, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]order.Order, 0)
	for rows.Next() {
		order := order.Order{UserLogin: login}
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s orderStorage) Update(ctx context.Context, order order.Order) error {
	res, err := s.connection().ExecContext(ctx,
		`UPDATE orders SET user_login = ?, status = ?, accrual = ?, uploaded_at = ? WHERE number = ?`,
		order.UserLogin, order.Status, order.Accrual, order.UploadedAt, order.Number)
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

func (s orderStorage) Delete(ctx context.Context, number order.Number) error {
	res, err := s.connection().ExecContext(ctx, `DELETE FROM orders WHERE number = ?`, number)
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
