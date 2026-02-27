// Package replay processes historical gk_moves rows in chronological order.
// It is used to bootstrap the stats schema from existing move data or to
// re-process a specific year/period when scoring rules change.
package replay

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// Options control which moves are replayed.
type Options struct {
	// StartID processes only moves with id >= StartID (0 = beginning of time).
	StartID int64
	// EndID processes only moves with id <= EndID (0 = no upper limit).
	EndID int64
	// StartDate processes only moves with moved_on_datetime >= StartDate.
	StartDate *time.Time
	// EndDate processes only moves with moved_on_datetime <= EndDate.
	EndDate *time.Time
	// Year is a convenience shorthand: sets StartDate=Jan 1 and EndDate=Dec 31 for the given year.
	// Only used when Year != 0 and StartDate/EndDate are nil.
	Year int
	// TruncateFirst, if true, wipes the geokrety_stats schema before replaying.
	TruncateFirst bool
}

// Handler is called for each move_id.
type Handler func(ctx context.Context, moveID int64) error

// Runner replays historical gk_moves rows and scores them via handler.
type Runner struct {
	store   store.Store
	cfg     config.ReplayConfig
	handler Handler
}

// New creates a new Replay runner.
func New(s store.Store, cfg config.ReplayConfig, handler Handler) *Runner {
	return &Runner{store: s, cfg: cfg, handler: handler}
}

// Run replays all moves matching opts.
// It processes moves in ascending ID order (oldest first).
func (r *Runner) Run(ctx context.Context, db dbTruncator, opts Options) error {
	// Expand Year shorthand.
	if opts.Year != 0 && opts.StartDate == nil && opts.EndDate == nil {
		start := time.Date(opts.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(opts.Year, 12, 31, 23, 59, 59, 999999999, time.UTC)
		opts.StartDate = &start
		opts.EndDate = &end
	}

	// Optionally wipe stats schema first.
	if opts.TruncateFirst {
		log.Warn().Msg("replay: truncating geokrety_stats schema before replay")
		if err := db.TruncateStatsSchema(ctx); err != nil {
			return fmt.Errorf("truncate stats schema: %w", err)
		}
	}

	batchSize := r.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	var afterID int64 = opts.StartID
	if afterID > 0 {
		afterID-- // GetMoveIDsPage uses >, so adjust
	}

	processed := 0
	errors := 0

	for {
		ids, err := r.store.GetMoveIDsPage(ctx, afterID, opts.EndID, opts.StartDate, opts.EndDate, batchSize)
		if err != nil {
			return fmt.Errorf("fetching move IDs (after %d): %w", afterID, err)
		}
		if len(ids) == 0 {
			break
		}

		for _, id := range ids {
			if err := r.handler(ctx, id); err != nil {
				log.Error().Int64("move_id", id).Err(err).Msg("replay: move processing error")
				errors++
			} else {
				processed++
			}
			afterID = id // next page starts after this id
		}

		log.Info().
			Int64("last_id", afterID).
			Int("processed", processed).
			Int("errors", errors).
			Msg("replay: batch done")

		// Yield between batches according to config.
		if r.cfg.BatchDelay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.cfg.BatchDelay):
			}
		} else {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		if len(ids) < batchSize {
			break // last page
		}
	}

	log.Info().
		Int("processed", processed).
		Int("errors", errors).
		Msg("replay: complete")

	return nil
}

// dbTruncator is a minimal interface used only by replay to wipe state.
type dbTruncator interface {
	TruncateStatsSchema(ctx context.Context) error
}
