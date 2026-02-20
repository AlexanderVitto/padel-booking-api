package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(databaseURL string) (*pgxpool.Pool, error) {
	// ctx ini untuk timeout saat connect; kalau DB down, tidak nge-hang lama.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Atur batas maksimum koneksi dari aplikasi Anda ke Postgres
	cfg.MaxConns = 5

	// Opsional (nanti saja kalau sudah paham):
	// cfg.MinConns = 2
	// cfg.MaxConnLifetime = 30 * time.Minute
	// cfg.MaxConnIdleTime = 5 * time.Minute
	// cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Test koneksi sekali saat startup (fail fast).
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
