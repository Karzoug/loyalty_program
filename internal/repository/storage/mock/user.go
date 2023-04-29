package mock

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/shopspring/decimal"
)

var _ storage.User = (*userStorage)(nil)

type userStorage struct {
	db *sql.DB
	tx *sql.Tx
}

func NewUserStorage(db *sql.DB) *userStorage {
	return &userStorage{
		db: db,
	}
}

func newUserTxStorage(tx *sql.Tx) *userStorage {
	return &userStorage{
		tx: tx,
	}
}

func (s userStorage) connection() sqliteConnecter {
	if s.tx == nil {
		return s.db
	}
	return s.tx
}

func (s userStorage) Create(ctx context.Context, user user.User) error {
	res, err := s.connection().ExecContext(ctx, `INSERT INTO users(login, encrypted_password, balance) VALUES(?, ?, ?)`,
		user.Login, user.EncryptedPassword, user.Balance)
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

func (s userStorage) UpdateBalance(ctx context.Context, login user.Login, delta decimal.Decimal) (*decimal.Decimal, error) {
	var balance decimal.Decimal
	err := s.connection().QueryRowContext(ctx,
		`UPDATE users SET balance = balance + (?) WHERE login = ? RETURNING balance`,
		delta, login).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}

	return &balance, nil
}

func (s userStorage) Get(ctx context.Context, login user.Login) (*user.User, error) {
	user := user.User{Login: login}
	err := s.connection().QueryRowContext(ctx,
		`SELECT encrypted_password, balance FROM users WHERE login = ?`, login).
		Scan(&user.EncryptedPassword, &user.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}
