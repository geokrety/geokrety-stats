package database

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
)

// DB wraps the pgxpool connection pool.
type DB struct {
	Pool *pgxpool.Pool
	cfg  config.DatabaseConfig
}

// New creates a new database connection pool and runs migrations.
func New(ctx context.Context, cfg config.DatabaseConfig, migrationPath string) (*DB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.PGXURL())
	if err != nil {
		return nil, fmt.Errorf("parsing db config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating db pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	db := &DB{Pool: pool, cfg: cfg}

	// Run migrations if path provided
	if migrationPath != "" {
		if err := db.Migrate(migrationPath); err != nil {
			pool.Close()
			return nil, fmt.Errorf("running migrations: %w", err)
		}
	}

	log.Info().Str("host", cfg.Host).Int("port", cfg.Port).Str("dbname", cfg.DBName).
		Msg("database connected")

	return db, nil
}

// Migrate runs database migrations.
func (db *DB) Migrate(migrationPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		db.cfg.PGXURL(),
	)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("applying migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("getting migration version: %w", err)
	}

	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("migrations applied")
	return nil
}

// MigrateDown rolls back all migrations (used for testing/reset).
func (db *DB) MigrateDown(migrationPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		db.cfg.PGXURL(),
	)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("rolling back migrations: %w", err)
	}

	log.Info().Msg("migrations rolled back")
	return nil
}

// Close closes the database connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// Exec is a convenience wrapper for executing a query.
func (db *DB) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := db.Pool.Exec(ctx, sql, args...)
	return err
}

// QueryRow is a convenience wrapper for querying a single row.
func (db *DB) QueryRow(ctx context.Context, sql string, args ...any) pgxRow {
	return db.Pool.QueryRow(ctx, sql, args...)
}

// pgxRow is an interface for pgx.Row to allow testing.
type pgxRow interface {
	Scan(dest ...any) error
}

// TruncateStatsSchema truncates all tables in the geokrety_stats schema.
// Used for replay initialization.
func (db *DB) TruncateStatsSchema(ctx context.Context) error {
	tables := []string{
		"geokrety_stats.user_points_log",
		"geokrety_stats.gk_points_log",
		"geokrety_stats.user_points_totals",
		"geokrety_stats.gk_multiplier_state",
		"geokrety_stats.gk_countries_visited",
		"geokrety_stats.user_move_history",
		"geokrety_stats.user_owner_gk_counts",
		"geokrety_stats.user_waypoint_monthly_counts",
		"geokrety_stats.user_monthly_diversity",
		"geokrety_stats.gk_chains",
		"geokrety_stats.gk_chain_members",
		"geokrety_stats.gk_chain_completions",
		"geokrety_stats.processed_events",
	}

	for _, t := range tables {
		log.Info().Str("table", t).Msg("truncating stats table")
		if err := db.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", t)); err != nil {
			return fmt.Errorf("truncating %s: %w", t, err)
		}
	}

	log.Info().Msg("stats schema truncated")
	return nil
}
