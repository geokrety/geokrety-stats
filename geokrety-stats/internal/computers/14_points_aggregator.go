package computers

import (
	"context"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// PointsAggregator is computer 14 – the final stage of the pipeline.
// It:
//  1. Discards zero-point awards.
//  2. Validates each remaining award (non-zero recipient, source set).
//  3. Merges duplicate (recipient, label) pairs.
//  4. Rounds each total to 2 decimal places (round-half-up).
//  5. Writes the FinalAward slice into pipeCtx.AggregatedAwards.
//  6. Calls store.SaveAwards() and store.MarkEventProcessed().
type PointsAggregator struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewPointsAggregator creates a new PointsAggregator computer.
func NewPointsAggregator(s store.Store, cfg config.StatsConfig) *PointsAggregator {
	return &PointsAggregator{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *PointsAggregator) Name() string {
	return "14_points_aggregator"
}

// Process implements the Computer interface.
func (c *PointsAggregator) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	raw := acc.Awards()

	// ── 1. Filter zero-point awards ─────────────────────────────────────────
	var nonZero []pipeline.Award
	for _, a := range raw {
		if a.Points == 0 {
			continue
		}
		nonZero = append(nonZero, a)
	}

	// ── 2 & 3. Validate + merge by (recipient, label) ───────────────────────
	type key struct {
		recipientUserID int64
		label           string
	}
	merged := make(map[key]float64)
	subAwards := make(map[key][]pipeline.Award)

	for _, a := range nonZero {
		if a.RecipientUserID == 0 {
			log.Warn().
				Str("label", a.Label).
				Str("module", a.ModuleSource).
				Msg("award has no recipient – skipping")
			continue
		}
		if a.ModuleSource == "" {
			return fmt.Errorf("award label=%q has no module_source set", a.Label)
		}
		k := key{a.RecipientUserID, a.Label}
		merged[k] += a.Points
		subAwards[k] = append(subAwards[k], a)
	}

	// ── 4. Round and build FinalAward slice ──────────────────────────────────
	var finals []pipeline.FinalAward
	for k, total := range merged {
		rounded := roundHalfUpAgg(total, 2)
		if rounded == 0 {
			continue
		}
		finals = append(finals, pipeline.FinalAward{
			RecipientUserID: k.recipientUserID,
			TotalPoints:     rounded,
			EventLogID:      pipeCtx.Event.LogID,
			GKID:            pipeCtx.Event.GKID,
			AwardedAt:       pipeCtx.Event.LoggedAt,
			Awards:          subAwards[k],
		})
	}

	// ── 5. Write into context ────────────────────────────────────────────────
	pipeCtx.AggregatedAwards = finals

	log.Debug().
		Int64("log_id", pipeCtx.Event.LogID).
		Int("awards", len(finals)).
		Msg("aggregated awards ready")

	// ── 6. Persist ───────────────────────────────────────────────────────────
	if err := c.store.SaveAwards(ctx, finals); err != nil {
		return fmt.Errorf("saving awards: %w", err)
	}

	if err := c.store.MarkEventProcessed(ctx, pipeCtx.Event.LogID, "ok"); err != nil {
		return fmt.Errorf("marking event processed: %w", err)
	}

	return nil
}

// roundHalfUpAgg rounds x to dp decimal places using round-half-up semantics.
func roundHalfUpAgg(x float64, dp int) float64 {
	pow := math.Pow(10, float64(dp))
	return math.Floor(x*pow+0.5) / pow
}
