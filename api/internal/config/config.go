package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds server configuration loaded from environment variables.
type Config struct {
	Port                string
	LogLevel            string
	EnableSwagger       bool
	WSBroadcastInterval int
	DatabaseURL         string
	PGHost              string
	PGPort              int
	PGUser              string
	PGPassword          string
	PGDatabase          string
	DBMaxOpenConns      int
	DBMaxIdleConns      int
}

func Load() (Config, error) {
	cfg := Config{
		Port:                getEnv("PORT", "3001"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		EnableSwagger:       getEnvBool("ENABLE_SWAGGER", true),
		WSBroadcastInterval: getEnvInt("WS_BROADCAST_INTERVAL", 15000),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		PGHost:              getEnv("PGHOST", "localhost"),
		PGPort:              getEnvInt("PGPORT", 5432),
		PGUser:              getEnv("PGUSER", "geokrety"),
		PGPassword:          os.Getenv("PGPASSWORD"),
		PGDatabase:          getEnv("PGDATABASE", "geokrety"),
		DBMaxOpenConns:      getEnvInt("DB_MAX_OPEN_CONNS", 20),
		DBMaxIdleConns:      getEnvInt("DB_MAX_IDLE_CONNS", 5),
	}

	if cfg.Port == "" {
		return Config{}, fmt.Errorf("PORT cannot be empty")
	}

	return cfg, nil
}

func (c Config) DSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.PGHost,
		c.PGPort,
		c.PGUser,
		c.PGPassword,
		c.PGDatabase,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
