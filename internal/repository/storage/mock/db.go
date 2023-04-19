package mock

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

const (
	duplicateKeyErrorCode = "1555"
)

type configPostgreSQLStorage interface {
	DatabaseURI() string
}

// NewDBInMemory creates connection to sqlite database in memory. Intended for testing only!
func NewDBInMemory(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("unable to create db connection: %w", err)
	}
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to apply migrations: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "sqlite", driver)
	if err != nil {
		return nil, fmt.Errorf("unable to apply migrations: %w", err)
	}
	err = m.Up()
	if err != nil {
		return nil, fmt.Errorf("unable to apply migrations: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database error: %w", err)
	}

	return db, nil
}

type sqliteConnecter interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
