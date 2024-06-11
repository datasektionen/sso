package database

import (
	"context"
	"database/sql"
	"embed"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:generate sqlc generate

//go:embed migrations/*.sql
var migrations embed.FS

func Connect(ctx context.Context) (*Queries, func() *sql.DB, error) {
	pool, err := pgxpool.New(ctx, config.Config.DatabaseURL.String())
	if err != nil {
		return nil, nil, err
	}

	return New(pool), func() *sql.DB { return stdlib.OpenDBFromPool(pool) }, nil
}

func ConnectAndMigrate(ctx context.Context) (*Queries, error) {
	q, db, err := Connect(ctx)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.UpContext(ctx, db(), "migrations"); err != nil {
		return nil, err
	}

	return q, nil
}

func (q *Queries) Tx(ctx context.Context, f func(db *Queries) error) error {
	pool := q.db.(*pgxpool.Pool)
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := f(q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
