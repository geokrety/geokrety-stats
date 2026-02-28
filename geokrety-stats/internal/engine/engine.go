// Package engine wires all scoring computers together and exposes a
// single ProcessMove method used by both the AMQP consumer and the
// historical replay runner.
package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// Engine runs the scoring pipeline for individual GeoKret moves.
type Engine struct {
	runner              *pipeline.Runner
	store               store.Store
	cfg                 config.StatsConfig
	suppressVerboseLog  bool // suppress individual move scored logs (useful during replay)
	progressCallback    func(*MoveResult) // optional callback for progress tracking
	skipCallback        func(int) // optional callback when move is skipped (halt), receives logType
}

// MoveResult holds result data from processing a move.
type MoveResult struct {
	MoveID   int64
	GKID     int64
	Awards   int
	LoggedAt time.Time
	LogType  int
}

// New builds an Engine with all 15 scoring computers wired up.
func New(s store.Store, cfg config.StatsConfig) *Engine {
	runner := pipeline.NewRunner(
		computers.NewEventGuard(s),
		computers.NewContextLoader(s, cfg),
		computers.NewBaseMovePoints(s, cfg),
		computers.NewOwnerGKLimitFilter(s, cfg),
		computers.NewWaypointPenalty(s, cfg),
		computers.NewCountryCrossing(s, cfg),
		computers.NewRelayBonus(s, cfg),
		computers.NewRescuerBonus(s, cfg),
		computers.NewHandoverBonus(s, cfg),
		computers.NewReachBonus(s, cfg),
		computers.NewChainStateManager(s, cfg),
		computers.NewChainBonus(s, cfg),
		computers.NewDiversityBonusTracker(s, cfg),
		computers.NewGKMultiplierUpdater(s, cfg),
		computers.NewPointsAggregator(s, cfg),
	)
	return &Engine{runner: runner, store: s, cfg: cfg}
}

// SetSuppressVerboseLog controls whether individual move logs are printed.
func (e *Engine) SetSuppressVerboseLog(suppress bool) {
	e.suppressVerboseLog = suppress
}

// SetProgressCallback sets an optional callback for move result tracking.
func (e *Engine) SetProgressCallback(cb func(*MoveResult)) {
	e.progressCallback = cb
}

// SetSkipCallback sets an optional callback for skipped moves (halts).
func (e *Engine) SetSkipCallback(cb func(int)) {
	e.skipCallback = cb
}

// ProcessMove loads a gk_moves row and runs the full scoring pipeline.
// It is idempotent: already-processed moves are silently skipped.
func (e *Engine) ProcessMove(ctx context.Context, moveID int64) error {
	start := time.Now()

	row, err := e.store.GetMove(ctx, moveID)
	if err != nil {
		return fmt.Errorf("loading move %d: %w", moveID, err)
	}
	if row == nil {
		return fmt.Errorf("move %d not found", moveID)
	}

	event := pipeline.Event{
		LogID:    row.ID,
		GKID:     row.GeoKretID,
		LogType:  pipeline.LogType(row.MoveType),
		LoggedAt: row.MovedOnDatetime,
	}
	if row.Author != nil {
		event.UserID = *row.Author
	}
	if row.Waypoint != nil {
		event.Waypoint = *row.Waypoint
	}
	if row.Country != nil {
		event.Country = *row.Country
	}
	if row.Lat != nil {
		event.Lat = *row.Lat
	}
	if row.Lon != nil {
		event.Lon = *row.Lon
	}

	result, err := e.runner.Run(ctx, event)
	if err != nil {
		return fmt.Errorf("pipeline error for move %d: %w", moveID, err)
	}

	elapsed := time.Since(start)

	if result.Halted {
		log.Debug().
			Int64("move_id", moveID).
			Str("reason", result.HaltReason).
			Dur("duration", elapsed).
			Msg("pipeline halted (move not scored)")

		// Call skip callback if set
		if e.skipCallback != nil {
			e.skipCallback(int(event.LogType))
		}
		return nil
	}

	// Call progress callback if set.
	if e.progressCallback != nil {
		e.progressCallback(&MoveResult{
			MoveID:   moveID,
			GKID:     event.GKID,
			Awards:   len(result.FinalAwards),
			LoggedAt: event.LoggedAt,
			LogType:  int(event.LogType),
		})
	}

	// Log move result if verbose logging is not suppressed.
	if !e.suppressVerboseLog {
		log.Info().
			Int64("move_id", moveID).
			Int64("gk_id", event.GKID).
			Int("awards", len(result.FinalAwards)).
			Dur("duration", elapsed).
			Msg("move scored")
	}

	return nil
}

// ChainBonus returns the ChainBonus computer, exposing its AwardTimeoutBonus
// method for use by the maintenance scheduler.
func (e *Engine) ChainBonus() *computers.ChainBonus {
	// The runner stores computers as pipeline.Computer (interface).
	// We need the concrete type for the timeout bonus method.
	// Reconstruct a standalone instance — safe since ChainBonus is stateless.
	return computers.NewChainBonus(e.store, e.cfg)
}
