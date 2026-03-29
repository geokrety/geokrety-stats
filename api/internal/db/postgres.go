package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
}

func Open(cfg config.Config) (*Store, error) {
	database, err := sqlx.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}
	database.SetMaxOpenConns(cfg.DBMaxOpenConns)
	database.SetMaxIdleConns(cfg.DBMaxIdleConns)
	database.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Store{db: database}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func pictureURLSQL(alias string) string {
	return fmt.Sprintf(`
CASE
	WHEN %[1]s.bucket IS NOT NULL AND %[1]s.key IS NOT NULL THEN 'https://minio.geokrety.org/' || %[1]s.bucket || '/' || %[1]s.key
	WHEN %[1]s.filename IS NOT NULL THEN 'https://cdn.geokrety.org/images/obrazki/' || %[1]s.filename
	ELSE NULL
END`, alias)
}

func countryFlag(code string) string {
	if len(code) != 2 {
		return ""
	}
	code = strings.ToUpper(code)
	return string([]rune{rune(code[0]) - 'A' + 0x1F1E6, rune(code[1]) - 'A' + 0x1F1E6})
}
