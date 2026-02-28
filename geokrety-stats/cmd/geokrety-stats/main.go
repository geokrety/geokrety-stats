// Command geokrety-stats is the GeoKrety scoring statistics daemon.
// It subscribes to the GeoKrety RabbitMQ exchange, scores incoming moves,
// and provides a historical replay mode for backfilling statistics.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/database"
	"github.com/geokrety/geokrety-points-system/internal/engine"
	"github.com/geokrety/geokrety-points-system/internal/maintenance"
	"github.com/geokrety/geokrety-points-system/internal/mqclient"
	"github.com/geokrety/geokrety-points-system/internal/replay"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

func main() {
	// ── Flags ────────────────────────────────────────────────────────────────
	var (
		cfgFile = flag.String("config", "", "Path to config file (optional; defaults to env vars)")
		replayMode    = flag.Bool("replay", false, "Run historical replay instead of daemon mode")
		replayYear    = flag.Int("year", 0, "Replay all moves from this year (e.g. 2017)")
		replayStart   = flag.Int64("start-id", 0, "Replay moves with id >= start-id")
		replayEnd     = flag.Int64("end-id", 0, "Replay moves with id <= end-id")
		replayTruncate = flag.Bool("truncate", false, "Truncate stats schema before replay")
		replayExpireDuringRun = flag.Bool("replay-expire-during-run", false, "During replay, periodically expire chains using historical as-of timestamps (slower; default false)")
		migrationUp   = flag.Bool("migration-up", false, "Apply all pending migrations and exit")
		migrationDown = flag.Bool("migration-down", false, "Rollback all migrations and exit")
	)
	flag.Parse()

	// ── Configuration ────────────────────────────────────────────────────────
	cfg, err := config.Load(*cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	setupLogging(cfg.Log)

	// ── Database ─────────────────────────────────────────────────────────────
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer initCancel()

	db, err := database.New(initCtx, cfg.Database, "migrations")
	if err != nil {
		log.Fatal().Err(err).Msg("database init failed")
	}
	defer db.Close()

	// ── Migration mode ───────────────────────────────────────────────────────
	if *migrationUp {
		if err := db.Migrate("migrations"); err != nil {
			log.Fatal().Err(err).Msg("migration up failed")
		}
		log.Info().Msg("migrations applied successfully")
		return
	}
	if *migrationDown {
		if err := db.MigrateDown("migrations"); err != nil {
			log.Fatal().Err(err).Msg("migration down failed")
		}
		log.Info().Msg("migrations rolled back successfully")
		return
	}

	s := store.New(db.Pool)

	// ── Engine (pipeline + all computers) ────────────────────────────────────
	eng := engine.New(s, cfg.Stats)

	// ── Replay mode ──────────────────────────────────────────────────────────
	if *replayMode {
		runReplay(*cfg, db, s, eng, *replayYear, *replayStart, *replayEnd, *replayTruncate, *replayExpireDuringRun)
		return
	}

	// ── Daemon mode ──────────────────────────────────────────────────────────
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Maintenance scheduler.
	sched := maintenance.New(s, eng.ChainBonus(), *cfg)
	sched.Start()
	defer sched.Stop()

	// AMQP consumer — blocks until ctx is cancelled.
	log.Info().Str("exchange", cfg.AMQP.Exchange).Msg("starting AMQP consumer")
	mq := mqclient.New(*cfg, func(ctx context.Context, moveID int64) error {
		return eng.ProcessMove(ctx, moveID)
	})

	if err := mq.Start(ctx); err != nil {
		log.Error().Err(err).Msg("AMQP consumer exited with error")
	}

	log.Info().Msg("geokrety-stats daemon stopped")
}

// runReplay executes historical replay and exits.
func runReplay(
	cfg config.Config,
	db *database.DB,
	s store.Store,
	eng *engine.Engine,
	year int, startID, endID int64, truncate bool, expireDuringRun bool,
) {
	opts := replay.Options{
		StartID:       startID,
		EndID:         endID,
		Year:          year,
		TruncateFirst: truncate,
	}

	log.Info().
		Int("year", year).
		Int64("start_id", startID).
		Int64("end_id", endID).
		Bool("truncate", truncate).
		Msg("starting historical replay")

	// Create progress tracker
	tracker := replay.NewProgressTracker(500 * time.Millisecond)

	// Configure engine for replay mode: suppress verbose logs and set progress callback
	eng.SetSuppressVerboseLog(true)
	eng.SetProgressCallback(func(result *engine.MoveResult) {
		tracker.RecordMoveResult(result.MoveID, result.GKID, result.Awards, result.LoggedAt, result.LogType)
	})
	eng.SetSkipCallback(func(logType int) {
		tracker.RecordSkipped(logType)
	})

	runner := replay.New(s, cfg.Replay, func(ctx context.Context, moveID int64) error {
		return eng.ProcessMove(ctx, moveID)
	})

	lastPrunedMonth := ""
	runner.SetReplayAsOfHook(func(ctx context.Context, asOf time.Time) error {
		currentMonth := asOf.UTC().Format("2006-01")
		if currentMonth != lastPrunedMonth {
			if _, err := db.Pool.Exec(ctx, `
				DELETE FROM geokrety_stats.user_monthly_diversity
				WHERE year_month < $1
			`, currentMonth); err != nil {
				return fmt.Errorf("pruning user_monthly_diversity: %w", err)
			}
			if _, err := db.Pool.Exec(ctx, `
				DELETE FROM geokrety_stats.user_monthly_diversity_countries
				WHERE year_month < $1
			`, currentMonth); err != nil {
				return fmt.Errorf("pruning user_monthly_diversity_countries: %w", err)
			}
			if _, err := db.Pool.Exec(ctx, `
				DELETE FROM geokrety_stats.user_monthly_diversity_drops
				WHERE year_month < $1
			`, currentMonth); err != nil {
				return fmt.Errorf("pruning user_monthly_diversity_drops: %w", err)
			}
			if _, err := db.Pool.Exec(ctx, `
				DELETE FROM geokrety_stats.user_monthly_diversity_owners
				WHERE year_month < $1
			`, currentMonth); err != nil {
				return fmt.Errorf("pruning user_monthly_diversity_owners: %w", err)
			}

			if err := s.RotateWaypointMonthlyPartitions(ctx, asOf, cfg.Maintenance.WaypointPartitionRetentionMonths, cfg.Maintenance.WaypointPartitionFutureMonths); err != nil {
				return fmt.Errorf("rotating waypoint partitions: %w", err)
			}

			endedBefore := asOf.AddDate(0, 0, -cfg.Maintenance.EndedChainRetentionDays)
			if _, err := s.PruneEndedChainsOlderThan(ctx, endedBefore); err != nil {
				return fmt.Errorf("pruning ended chains: %w", err)
			}

			pointsBefore := asOf.AddDate(0, 0, -cfg.Maintenance.GKPointsLogRetentionDays)
			if _, err := s.PruneGKPointsLogOlderThan(ctx, pointsBefore); err != nil {
				return fmt.Errorf("pruning gk_points_log: %w", err)
			}

			lastPrunedMonth = currentMonth
		}

		if expireDuringRun {
			return maintenance.ExpireChainsAt(ctx, s, eng.ChainBonus(), cfg.Stats.ChainTimeoutDays, asOf)
		}

		return nil
	})

	// Set progress tracking callbacks on runner
	runner.SetErrorRecorder(tracker)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start progress display
	stopProgress := tracker.Start(ctx)

	// Run replay
	if err := runner.Run(ctx, db, opts); err != nil {
		// Stop progress display and show summary
		stats := stopProgress()
		fmt.Print("\r\033[K")
		fmt.Printf("%s\n", stats)

		// Silently exit on interrupt
		if ctx.Err() != nil {
			os.Exit(130)
		}
		log.Error().Err(err).Msg("replay failed")
		os.Exit(1)
	}

	// Stop progress display and show summary
	stats := stopProgress()
	fmt.Print("\r\033[K")
	fmt.Printf("%s\n", stats)

	// Apply historical chain timeout pass as-of replay date (for chains that would
	// have expired without a subsequent move event on the same GK).
	replayAsOf := time.Now().UTC()
	if year != 0 {
		replayAsOf = time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)
	}

	maintCtx, maintCancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer maintCancel()

	if err := maintenance.ExpireChainsAt(maintCtx, s, eng.ChainBonus(), cfg.Stats.ChainTimeoutDays, replayAsOf); err != nil {
		log.Warn().Err(err).Msg("replay: historical chain expiry pass failed (non-fatal)")
	}

	log.Info().Msg("replay: final materialized view refresh (light + heavy)")
	if err := db.RefreshMaterializedViewsLight(maintCtx); err != nil {
		log.Warn().Err(err).Msg("replay: final light materialized view refresh failed (non-fatal)")
	}
	if err := db.RefreshMaterializedViewsHeavy(maintCtx); err != nil {
		log.Warn().Err(err).Msg("replay: final heavy materialized view refresh failed (non-fatal)")
	}
}

// setupLogging configures zerolog based on the log config.
func setupLogging(cfg config.LogConfig) {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if cfg.Format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
}
