package db

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMissingDatabaseURL = errors.New("DATABASE_URL is not set")

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, ErrMissingDatabaseURL
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// sensible defaults for dev
	cfg.MinConns = 1
	cfg.MaxConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute

	return pgxpool.NewWithConfig(ctx, cfg)
}
