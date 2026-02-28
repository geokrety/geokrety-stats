package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Connect creates and validates a pgx connection pool.
func Connect(ctx context.Context, dsn string, maxConns, minConns int) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.MaxConns = int32(maxConns)
	cfg.MinConns = int32(minConns)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.ConnConfig.Host).
		Str("database", cfg.ConnConfig.Database).
		Int32("max_conns", cfg.MaxConns).
		Msg("database pool connected")
	return pool, nil
}
