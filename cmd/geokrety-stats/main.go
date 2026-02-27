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

	s := store.New(db.Pool)

	// ── Engine (pipeline + all computers) ────────────────────────────────────
	eng := engine.New(s, cfg.Stats)

	// ── Replay mode ──────────────────────────────────────────────────────────
	if *replayMode {
		runReplay(*cfg, db, s, eng, *replayYear, *replayStart, *replayEnd, *replayTruncate)
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
	year int, startID, endID int64, truncate bool,
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

	runner := replay.New(s, cfg.Replay, func(ctx context.Context, moveID int64) error {
		return eng.ProcessMove(ctx, moveID)
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	start := time.Now()
	if err := runner.Run(ctx, db, opts); err != nil {
		log.Error().Err(err).Msg("replay failed")
		os.Exit(1)
	}

	log.Info().Dur("elapsed", time.Since(start)).Msg("replay completed successfully")
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
