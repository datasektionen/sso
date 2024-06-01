package database

import (
	"context"
	"embed"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:generate sqlc generate

//go:embed migrations/*.sql
var migrations embed.FS

func ConnectAndMigrate(ctx context.Context) (*Queries, error) {
	pool, err := pgxpool.New(ctx, config.Config.DatabaseURL.String())
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.UpContext(ctx, stdlib.OpenDBFromPool(pool), "migrations"); err != nil {
		return nil, err
	}

	return New(pool), nil
}
