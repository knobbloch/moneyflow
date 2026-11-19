package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDSN = "postgres://moneyflow:moneyflow@localhost:5432/moneyflow?sslmode=disable"

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDSN
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}