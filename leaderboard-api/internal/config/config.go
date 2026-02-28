package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	TrustedProxies  []string
	AllowOrigins    []string
	WSPingInterval  int // seconds
	RefreshInterval int // seconds between WS leaderboard pushes
}

type DatabaseConfig struct {
	DSN      string
	PoolMax  int
	PoolMin  int
	ReadOnly bool
}

// Load reads config from env variables / config file.
func Load() *Config {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.allow_origins", []string{"*"})
	viper.SetDefault("server.ws_ping_interval", 30)
	viper.SetDefault("server.refresh_interval", 30)

	viper.SetDefault("database.pool_max", 20)
	viper.SetDefault("database.pool_min", 2)

	viper.SetEnvPrefix("API")
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/leaderboard-api")
	_ = viper.ReadInConfig()

	dsn := viper.GetString("database.dsn")
	if dsn == "" {
		dsn = viper.GetString("API_DATABASE_DSN")
	}
	if dsn == "" {
		log.Warn().Msg("No DATABASE_DSN set; using default local postgres")
		dsn = "postgres://postgres:postgres@localhost:5432/geokrety?sslmode=disable"
	}

	return &Config{
		Server: ServerConfig{
			Host:            viper.GetString("server.host"),
			Port:            viper.GetInt("server.port"),
			AllowOrigins:    viper.GetStringSlice("server.allow_origins"),
			WSPingInterval:  viper.GetInt("server.ws_ping_interval"),
			RefreshInterval: viper.GetInt("server.refresh_interval"),
		},
		Database: DatabaseConfig{
			DSN:     dsn,
			PoolMax: viper.GetInt("database.pool_max"),
			PoolMin: viper.GetInt("database.pool_min"),
		},
	}
}
