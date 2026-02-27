package computers

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// OwnerGKLimitFilter is computer 03.
// Enforces the anti-farming rule limiting how many different GKs from a single
// owner a user can earn base points from (maximum MaxGKsPerOwner, all time).
type OwnerGKLimitFilter struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewOwnerGKLimitFilter creates a new OwnerGKLimitFilter computer.
func NewOwnerGKLimitFilter(s store.Store, cfg config.StatsConfig) *OwnerGKLimitFilter {
	return &OwnerGKLimitFilter{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *OwnerGKLimitFilter) Name() string {
	return "03_owner_gk_limit_filter"
}

// Process implements the Computer interface.
func (c *OwnerGKLimitFilter) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState
	us := &pipeCtx.UserState

	// Step 1 – Only applies to non-owner moves.
	if event.UserID == gks.OwnerID {
		return nil
	}

	// Step 2 – Only applies if base points were awarded.
	if !pipeCtx.RuntimeFlags.ActorScoredThisGK {
		return nil
	}

	// Step 3 – If this GK was already counted, no change needed.
	if us.ActorGKAlreadyCountedForOwner {
		return nil
	}

	// Step 4 – Check if the actor has already hit the limit for this owner.
	if us.ActorGKsPerOwnerCount >= c.cfg.MaxGKsPerOwner {
		// Zero out base_move awards.
		acc.ZeroByLabel("base_move")
		pipeCtx.RuntimeFlags.ActorScoredThisGK = false
		pipeCtx.RuntimeFlags.BasePointsAwarded = 0
		log.Debug().
			Int64("user_id", event.UserID).
			Int("count", us.ActorGKsPerOwnerCount).
			Int64("owner_id", gks.OwnerID).
			Msg("owner GK limit reached; base points zeroed")
		return nil
	}

	// Step 4b – Under the limit: record this GK as counted.
	if err := c.store.RecordGKCountedForOwner(ctx, event.UserID, gks.OwnerID, event.GKID, event.LoggedAt); err != nil {
		return fmt.Errorf("recording GK count for owner: %w", err)
	}

	return nil
}
