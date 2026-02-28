package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Database    DatabaseConfig    `mapstructure:"database"`
	AMQP        AMQPConfig        `mapstructure:"amqp"`
	Log         LogConfig         `mapstructure:"log"`
	Replay      ReplayConfig      `mapstructure:"replay"`
	Stats       StatsConfig       `mapstructure:"stats"`
	Maintenance MaintenanceConfig `mapstructure:"maintenance"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	// Connection pool settings
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// PGXURL returns the pgx-compatible URL format.
func (d DatabaseConfig) PGXURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// AMQPConfig holds RabbitMQ connection settings.
type AMQPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
	// Queue name for this service
	QueueName string `mapstructure:"queue_name"`
	// Exchange to bind to
	Exchange string `mapstructure:"exchange"`
	// Reconnect delay if connection fails
	ReconnectDelay time.Duration `mapstructure:"reconnect_delay"`
	// Max reconnect delay
	MaxReconnectDelay time.Duration `mapstructure:"max_reconnect_delay"`
}

// URL returns the AMQP connection URL.
func (a AMQPConfig) URL() string {
	vhost := a.VHost
	if vhost == "" {
		vhost = "/"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		a.User, a.Password, a.Host, a.Port, strings.TrimPrefix(vhost, "/"),
	)
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // "json" or "console"
}

// ReplayConfig holds settings for historical replay mode.
type ReplayConfig struct {
	// Batch size for reading historical moves
	BatchSize int `mapstructure:"batch_size"`
	// Delay between batches (to avoid overloading DB)
	BatchDelay time.Duration `mapstructure:"batch_delay"`
	// Concurrency (number of parallel pipeline workers)
	Concurrency int `mapstructure:"concurrency"`
}

// StatsConfig holds scoring rule parameters.
// All tuneable parameters for the gamification system are here.
type StatsConfig struct {
	// Base points for a first move
	BaseMovePoints float64 `mapstructure:"base_move_points"`

	// Max distinct GKs per owner a user can earn base points from (all time)
	MaxGKsPerOwner int `mapstructure:"max_gks_per_owner"`

	// Waypoint penalty tiers (0-indexed: 1st GK = 100%, 2nd = 50%, etc.)
	WaypointPenaltyTiers []float64 `mapstructure:"waypoint_penalty_tiers"`

	// Country crossing bonus: non-owner moving standard GK to new country
	CountryCrossingActorBonus float64 `mapstructure:"country_crossing_actor_bonus"`
	// Country crossing bonus: owner moving their own standard GK to new country
	CountryCrossingOwnerSelfBonus float64 `mapstructure:"country_crossing_owner_self_bonus"`
	// Country crossing bonus: owner moving non-transferable GK to new country
	CountryCrossingNonTransferableBonus float64 `mapstructure:"country_crossing_non_transferable_bonus"`
	// Country crossing bonus to owner (given by non-owner)
	CountryCrossingOwnerBonus float64 `mapstructure:"country_crossing_owner_bonus"`

	// Relay bonus: to the new grabber (mover)
	RelayMoverBonus float64 `mapstructure:"relay_mover_bonus"`
	// Relay bonus: to the previous dropper
	RelayDropperBonus float64 `mapstructure:"relay_dropper_bonus"`
	// Relay window in hours (default: 168 = 7 days)
	RelayWindowHours int `mapstructure:"relay_window_hours"`

	// Rescuer bonus: to the grabber
	RescuerGrabberBonus float64 `mapstructure:"rescuer_grabber_bonus"`
	// Rescuer bonus: to the GK owner
	RescuerOwnerBonus float64 `mapstructure:"rescuer_owner_bonus"`
	// Dormancy threshold in months to qualify for rescue bonus
	RescuerDormancyMonths int `mapstructure:"rescuer_dormancy_months"`

	// Handover bonus: to the GK owner
	HandoverOwnerBonus float64 `mapstructure:"handover_owner_bonus"`

	// Reach bonus: to owner at milestone
	ReachOwnerBonus float64 `mapstructure:"reach_owner_bonus"`
	// Reach milestone: number of distinct users in 6-month window
	ReachMilestoneUsers int `mapstructure:"reach_milestone_users"`
	// Rolling window for reach milestone in months
	ReachWindowMonths int `mapstructure:"reach_window_months"`

	// Chain timeout in days (chain expires after this many days of inactivity)
	ChainTimeoutDays int `mapstructure:"chain_timeout_days"`
	// Minimum chain length to earn chain bonus
	ChainMinLength int `mapstructure:"chain_min_length"`
	// Chain anti-farming cooldown in months
	ChainAntiFarmingMonths int `mapstructure:"chain_anti_farming_months"`
	// Owner's share of chain bonus (fraction, default 0.25)
	ChainOwnerShareFraction float64 `mapstructure:"chain_owner_share_fraction"`

	// Diversity bonus: 5 drops in a month
	DiversityDropsBonus float64 `mapstructure:"diversity_drops_bonus"`
	DiversityDropsMilestone int `mapstructure:"diversity_drops_milestone"`
	// Diversity bonus: 10 distinct owners in a month
	DiversityOwnersBonus float64 `mapstructure:"diversity_owners_bonus"`
	DiversityOwnersMilestone int `mapstructure:"diversity_owners_milestone"`
	// Diversity bonus: new country (first time per country per actor per month)
	DiversityCountryBonus float64 `mapstructure:"diversity_country_bonus"`

	// GK multiplier settings
	MultiplierMin            float64 `mapstructure:"multiplier_min"`
	MultiplierMax            float64 `mapstructure:"multiplier_max"`
	MultiplierFirstMoveInc   float64 `mapstructure:"multiplier_first_move_inc"`
	MultiplierCountryInc     float64 `mapstructure:"multiplier_country_inc"`
	MultiplierInHandDecayPerDay  float64 `mapstructure:"multiplier_in_hand_decay_per_day"`
	MultiplierInCacheDecayPerWeek float64 `mapstructure:"multiplier_in_cache_decay_per_week"`

	// GK types considered non-transferable (owner can earn points on their own moves)
	NonTransferableGKTypes []int `mapstructure:"non_transferable_gk_types"`

	// First-finder window in hours (moves within this window qualify as first-finders)
	FirstFinderWindowHours int `mapstructure:"first_finder_window_hours"`
}

// IsNonTransferable returns true if the given GK type is non-transferable.
func (s StatsConfig) IsNonTransferable(gkType int) bool {
	for _, t := range s.NonTransferableGKTypes {
		if t == gkType {
			return true
		}
	}
	return false
}

// MaintenanceConfig holds settings for background maintenance jobs.
type MaintenanceConfig struct {
	// Cron schedule for chain expiry check
	ChainExpirySchedule string `mapstructure:"chain_expiry_schedule"`
	// Cron schedule for multiplier decay update (for GKs with no recent events)
	MultiplierDecaySchedule string `mapstructure:"multiplier_decay_schedule"`
	// Cron schedule for pruning old processed events log
	PruneSchedule string `mapstructure:"prune_schedule"`
	// How long to keep processed events log (days)
	ProcessedEventRetentionDays int `mapstructure:"processed_event_retention_days"`
}

// Load reads configuration from file and environment variables.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.AddConfigPath("$HOME/.geokrety-stats")
	}

	// Environment variables (GK_STATS_ prefix)
	v.SetEnvPrefix("GK_STATS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Also support plain GK_ prefix env vars (compatibility with website workers)
	v.BindEnv("database.host", "GK_DB_HOST", "GK_STATS_DATABASE_HOST")       //nolint:errcheck
	v.BindEnv("database.port", "GK_DB_PORT", "GK_STATS_DATABASE_PORT")       //nolint:errcheck
	v.BindEnv("database.user", "GK_DB_USER", "GK_STATS_DATABASE_USER")       //nolint:errcheck
	v.BindEnv("database.password", "GK_DB_PASS", "GK_STATS_DATABASE_PASSWORD") //nolint:errcheck
	v.BindEnv("database.dbname", "GK_DB_NAME", "GK_STATS_DATABASE_DBNAME")   //nolint:errcheck
	v.BindEnv("amqp.host", "GK_RABBITMQ_HOST", "GK_STATS_AMQP_HOST")         //nolint:errcheck
	v.BindEnv("amqp.port", "GK_RABBITMQ_PORT", "GK_STATS_AMQP_PORT")         //nolint:errcheck
	v.BindEnv("amqp.user", "GK_RABBITMQ_USER", "GK_STATS_AMQP_USER")         //nolint:errcheck
	v.BindEnv("amqp.password", "GK_RABBITMQ_PASS", "GK_STATS_AMQP_PASSWORD") //nolint:errcheck
	v.BindEnv("amqp.vhost", "GK_RABBITMQ_VHOST", "GK_STATS_AMQP_VHOST")     //nolint:errcheck

	// Support full URL env vars (parsed into individual fields)
	v.BindEnv("db_url", "GK_STATS_DB_URL", "GK_STATS_DATABASE_URL")           //nolint:errcheck
	v.BindEnv("amqp_url", "GK_STATS_AMQP_URL")                               //nolint:errcheck

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		// Config file not found is OK — use defaults + env vars
	}

	// Parse GK_STATS_DB_URL or GK_STATS_DATABASE_URL if provided
	// Check environment variables directly (viper binding may not work reliably)
	dbURL := os.Getenv("GK_STATS_DB_URL")
	if dbURL == "" {
		dbURL = os.Getenv("GK_STATS_DATABASE_URL")
	}
	if dbURL != "" {
		if err := parsePostgresURL(v, dbURL); err != nil {
			return nil, fmt.Errorf("parsing GK_STATS_DB_URL: %w", err)
		}
	}

	// Parse GK_STATS_AMQP_URL if provided
	amqpURL := os.Getenv("GK_STATS_AMQP_URL")
	if amqpURL != "" {
		if err := parseAMQPURL(v, amqpURL); err != nil {
			return nil, fmt.Errorf("parsing GK_STATS_AMQP_URL: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets all default configuration values.
func setDefaults(v *viper.Viper) {
	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "geokrety")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "geokrety")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// AMQP defaults
	v.SetDefault("amqp.host", "localhost")
	v.SetDefault("amqp.port", 5672)
	v.SetDefault("amqp.user", "guest")
	v.SetDefault("amqp.password", "guest")
	v.SetDefault("amqp.vhost", "/")
	v.SetDefault("amqp.queue_name", "geokrety_stats_worker")
	v.SetDefault("amqp.exchange", "geokrety")
	v.SetDefault("amqp.reconnect_delay", "5s")
	v.SetDefault("amqp.max_reconnect_delay", "60s")

	// Logging defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")

	// Replay defaults
	v.SetDefault("replay.batch_size", 500)
	v.SetDefault("replay.batch_delay", "0")
	v.SetDefault("replay.concurrency", 1)

	// Stats / scoring parameters
	v.SetDefault("stats.base_move_points", 3.0)
	v.SetDefault("stats.max_gks_per_owner", 10)
	v.SetDefault("stats.waypoint_penalty_tiers", []float64{1.0, 0.5, 0.25, 0.0})
	v.SetDefault("stats.country_crossing_actor_bonus", 3.0)
	v.SetDefault("stats.country_crossing_owner_self_bonus", 2.0)
	v.SetDefault("stats.country_crossing_non_transferable_bonus", 4.0)
	v.SetDefault("stats.country_crossing_owner_bonus", 1.0)
	v.SetDefault("stats.relay_mover_bonus", 2.0)
	v.SetDefault("stats.relay_dropper_bonus", 1.0)
	v.SetDefault("stats.relay_window_hours", 168) // 7 days
	v.SetDefault("stats.rescuer_grabber_bonus", 2.0)
	v.SetDefault("stats.rescuer_owner_bonus", 1.0)
	v.SetDefault("stats.rescuer_dormancy_months", 6)
	v.SetDefault("stats.handover_owner_bonus", 1.0)
	v.SetDefault("stats.reach_owner_bonus", 5.0)
	v.SetDefault("stats.reach_milestone_users", 10)
	v.SetDefault("stats.reach_window_months", 6)
	v.SetDefault("stats.chain_timeout_days", 14)
	v.SetDefault("stats.chain_min_length", 3)
	v.SetDefault("stats.chain_anti_farming_months", 6)
	v.SetDefault("stats.chain_owner_share_fraction", 0.25)
	v.SetDefault("stats.diversity_drops_bonus", 3.0)
	v.SetDefault("stats.diversity_drops_milestone", 5)
	v.SetDefault("stats.diversity_owners_bonus", 7.0)
	v.SetDefault("stats.diversity_owners_milestone", 10)
	v.SetDefault("stats.diversity_country_bonus", 5.0)
	v.SetDefault("stats.multiplier_min", 1.0)
	v.SetDefault("stats.multiplier_max", 2.0)
	v.SetDefault("stats.multiplier_first_move_inc", 0.01)
	v.SetDefault("stats.multiplier_country_inc", 0.05)
	v.SetDefault("stats.multiplier_in_hand_decay_per_day", 0.008)
	v.SetDefault("stats.multiplier_in_cache_decay_per_week", 0.02)
	// GK type 4 = KretyPost (non-transferable)
	v.SetDefault("stats.non_transferable_gk_types", []int{4})
	v.SetDefault("stats.first_finder_window_hours", 168) // 7 days

	// Maintenance defaults
	v.SetDefault("maintenance.chain_expiry_schedule", "0 */15 * * * *")    // every 15 min
	v.SetDefault("maintenance.multiplier_decay_schedule", "0 0 2 * * *")   // daily at 2am
	v.SetDefault("maintenance.prune_schedule", "0 0 3 * * *")              // daily at 3am
	v.SetDefault("maintenance.processed_event_retention_days", 30)
}

// parsePostgresURL parses a PostgreSQL connection URL and sets database config values.
func parsePostgresURL(v *viper.Viper, dbURL string) error {
	// Handle both postgresql:// and postgres:// schemes
	if !strings.HasPrefix(dbURL, "postgresql://") && !strings.HasPrefix(dbURL, "postgres://") {
		return fmt.Errorf("invalid PostgreSQL URL scheme (must start with postgresql:// or postgres://)")
	}

	// Replace postgres:// with postgresql:// for net/url parsing
	if strings.HasPrefix(dbURL, "postgres://") {
		dbURL = "postgresql://" + dbURL[11:]
	}

	u, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	// Extract host and port
	host := u.Hostname()
	if host == "" {
		host = "localhost"
	}
	v.Set("database.host", host)

	port := 5432 // PostgreSQL default
	if p := u.Port(); p != "" {
		if portNum, err := strconv.Atoi(p); err == nil {
			port = portNum
		}
	}
	v.Set("database.port", port)

	// Extract credentials
	if u.User != nil {
		v.Set("database.user", u.User.Username())
		if password, ok := u.User.Password(); ok {
			v.Set("database.password", password)
		}
	}

	// Extract database name from path (e.g., /geokrety)
	if path := strings.TrimPrefix(u.Path, "/"); path != "" {
		v.Set("database.dbname", path)
	}

	// Extract sslmode from query params
	if q := u.Query().Get("sslmode"); q != "" {
		v.Set("database.sslmode", q)
	}

	return nil
}

// parseAMQPURL parses an AMQP connection URL and sets amqp config values.
func parseAMQPURL(v *viper.Viper, amqpURL string) error {
	if !strings.HasPrefix(amqpURL, "amqp://") && !strings.HasPrefix(amqpURL, "amqps://") {
		return fmt.Errorf("invalid AMQP URL scheme (must start with amqp:// or amqps://)")
	}

	u, err := url.Parse(amqpURL)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	// Extract host and port
	host := u.Hostname()
	if host == "" {
		host = "localhost"
	}
	v.Set("amqp.host", host)

	port := 5672 // AMQP default (5671 for amqps)
	if p := u.Port(); p != "" {
		if portNum, err := strconv.Atoi(p); err == nil {
			port = portNum
		}
	} else if strings.HasPrefix(amqpURL, "amqps://") {
		port = 5671
	}
	v.Set("amqp.port", port)

	// Extract credentials
	if u.User != nil {
		v.Set("amqp.user", u.User.Username())
		if password, ok := u.User.Password(); ok {
			v.Set("amqp.password", password)
		}
	}

	// Extract vhost from path (e.g., /myvhost)
	if path := u.Path; path != "" && path != "/" {
		v.Set("amqp.vhost", path)
	}

	return nil
}
