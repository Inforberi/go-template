package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Inforberi/go-template/internal/infra/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

const connectTimeout = 5 * time.Second

func New(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, errors.New("parse PostgreSQL configuration")
	}
	poolConfig.MaxConns = cfg.MaxConns

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, errors.New("create PostgreSQL pool")
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, errors.New("connect to PostgreSQL")
	}

	return pool, nil
}
