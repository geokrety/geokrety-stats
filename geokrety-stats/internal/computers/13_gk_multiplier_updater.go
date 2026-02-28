package computers

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// GKMultiplierUpdater is computer 13.
// Recalculates and persists the GeoKret's per-event multiplier.
// Order of operations:
//  1. Apply time decay (in-hand: −0.008/day; in-cache: −0.02/week).
//  2. First-move-type bonus: +0.01 if this log_type is new for this user on this GK.
//  3. Country crossing bonus: +0.05 if computer 05 set RuntimeFlags.NewCountryVisited.
//  4. Clamp to [floor, ceiling] (default 1.0 – 2.0).
//  5. Persist new multiplier, log the change, and record user move history.
type GKMultiplierUpdater struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewGKMultiplierUpdater creates a new GKMultiplierUpdater computer.
func NewGKMultiplierUpdater(s store.Store, cfg config.StatsConfig) *GKMultiplierUpdater {
	return &GKMultiplierUpdater{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *GKMultiplierUpdater) Name() string {
	return "13_gk_multiplier_updater"
}

// Process implements the Computer interface.
func (c *GKMultiplierUpdater) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gk := &pipeCtx.GKState

	current := gk.CurrentMultiplier
	prev := current

	// ── Step 1 – Time decay ─────────────────────────────────────────────────
	current = c.applyDecay(current, gk.LastMultiplierAt, event.LoggedAt, gk.IsInCache())

	// ── Step 2 – First-move-type bonus ──────────────────────────────────────
	if c.isFirstMoveType(pipeCtx) {
		current += c.cfg.MultiplierFirstMoveInc
		log.Debug().
			Int64("gk_id", event.GKID).
			Float64("bonus", c.cfg.MultiplierFirstMoveInc).
			Msg("multiplier +bonus (first move type)")
	}

	// ── Step 3 – Country crossing bonus ─────────────────────────────────────
	if pipeCtx.RuntimeFlags.NewCountryVisited {
		current += c.cfg.MultiplierCountryInc
		log.Debug().
			Int64("gk_id", event.GKID).
			Float64("bonus", c.cfg.MultiplierCountryInc).
			Msg("multiplier +bonus (country crossing)")
	}

	// ── Step 4 – Clamp ──────────────────────────────────────────────────────
	current = math.Max(c.cfg.MultiplierMin, math.Min(c.cfg.MultiplierMax, current))

	// ── Step 5 – Persist ────────────────────────────────────────────────────
	// Round to 4dp to keep DB compact.
	current = math.Round(current*10000) / 10000

	if err := c.store.SaveGKMultiplierState(ctx, event.GKID, current, event.LoggedAt, gk.CurrentHolder, nil); err != nil {
		return fmt.Errorf("saving multiplier state: %w", err)
	}

	if err := c.store.LogGKMultiplierChange(ctx, event.GKID, event.LogID, prev, current, "pipeline", c.Name()); err != nil {
		return fmt.Errorf("logging multiplier change: %w", err)
	}

	if err := c.store.AddUserMoveHistory(ctx, event.UserID, event.GKID, event.LogType, event.LoggedAt); err != nil {
		return fmt.Errorf("recording user move history: %w", err)
	}

	// Expose updated multiplier in context so downstream computers can use it.
	gk.CurrentMultiplier = current

	log.Debug().
		Int64("gk_id", event.GKID).
		Float64("prev", prev).
		Float64("new", current).
		Msg("GK multiplier updated")

	return nil
}

// applyDecay reduces the multiplier based on elapsed days since last update.
// In-cache decay:  −0.02 per week (≈ −0.002857/day).
// In-hand decay:   −0.008 per day.
func (c *GKMultiplierUpdater) applyDecay(mult float64, lastUpdated, now time.Time, inCache bool) float64 {
	if lastUpdated.IsZero() {
		return mult
	}

	elapsed := now.Sub(lastUpdated)
	days := elapsed.Hours() / 24

	if days <= 0 {
		return mult
	}

	var decayPerDay float64
	if inCache {
		// −0.02 per full week = −0.02/7 per day:
		decayPerDay = c.cfg.MultiplierInCacheDecayPerWeek / 7.0
	} else {
		decayPerDay = c.cfg.MultiplierInHandDecayPerDay
	}

	return mult - decayPerDay*days
}

// isFirstMoveType returns true if the acting user has never logged this specific
// log_type on this GeoKret in their recorded move history.
func (c *GKMultiplierUpdater) isFirstMoveType(pipeCtx *pipeline.Context) bool {
	_, seen := pipeCtx.UserState.ActorMoveHistoryOnGK[pipeCtx.Event.LogType]
	return !seen
}
