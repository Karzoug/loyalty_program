package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	duplicateKeyErrorCode = "23505"
	poolCreationTimeout   = 3 * time.Second
)

type configPostgreSQLStorage interface {
	DatabaseURI() string
}

func newDBPool(ctx context.Context, cfg configPostgreSQLStorage) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURI())
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, poolCreationTimeout)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database error: %w", err)
	}

	return pool, nil
}

type pgConnecter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
