package postgresql

import (
	"context"
	"errors"

	"github.com/Karzoug/loyalty_program/internal/model/user"
	"github.com/Karzoug/loyalty_program/internal/repository/storage"
	"github.com/shopspring/decimal"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ storage.User = (*userStorage)(nil)

type userStorage struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func NewUserStorage(pool *pgxpool.Pool) *userStorage {
	return &userStorage{
		pool: pool,
	}
}

func newUserTxStorage(tx pgx.Tx) *userStorage {
	return &userStorage{
		tx: tx,
	}
}

func (s userStorage) connection() pgConnecter {
	if s.tx == nil {
		return s.pool
	}
	return s.tx
}

func (s userStorage) Create(ctx context.Context, user user.User) error {
	_, err := s.connection().Exec(ctx, `INSERT INTO users(login, encrypted_password, balance) VALUES($1, $2, $3)`,
		user.Login, user.EncryptedPassword, user.Balance)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicateKeyErrorCode {
			return storage.ErrRecordAlreadyExists
		}
		return err
	}

	return nil
}

func (s userStorage) UpdateBalance(ctx context.Context, login user.Login, delta decimal.Decimal) (*decimal.Decimal, error) {
	var balance decimal.Decimal
	err := s.connection().QueryRow(ctx,
		`UPDATE users SET balance = balance + ($1) WHERE login = $2 RETURNING balance`,
		delta, login).Scan(&balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}

	return &balance, nil
}

func (s userStorage) Get(ctx context.Context, login user.Login) (*user.User, error) {
	user := user.User{Login: login}
	err := s.connection().QueryRow(ctx,
		`SELECT encrypted_password, balance FROM users WHERE login = $1`, login).
		Scan(&user.EncryptedPassword, &user.Balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}
