package mock

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Karzoug/loyalty_program/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

const (
	duplicateKeyErrorCode = "1555"
)

// newDBInMemory creates connection to sqlite database in memory. Intended for testing only!
func newDBInMemory(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("unable to create db connection: %w", err)
	}

	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("unable to apply migrations: %w", err)
	}
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to apply migrations: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", driver)
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
