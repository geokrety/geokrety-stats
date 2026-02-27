package computers

import (
	"context"
	"fmt"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// WaypointPenalty is computer 04.
// Applies a scaling penalty to base move points when a user interacts with
// multiple different GeoKrety at the same location within the same calendar month.
type WaypointPenalty struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewWaypointPenalty creates a new WaypointPenalty computer.
func NewWaypointPenalty(s store.Store, cfg config.StatsConfig) *WaypointPenalty {
	return &WaypointPenalty{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *WaypointPenalty) Name() string {
	return "04_waypoint_penalty"
}

// Process implements the Computer interface.
func (c *WaypointPenalty) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	us := &pipeCtx.UserState

	// Step 1 – Determine location identity.
	locationID := event.LocationID()
	if locationID == "" {
		return nil // No location data; penalty not applicable.
	}

	// Step 2 – Check if base points exist.
	if !pipeCtx.RuntimeFlags.ActorScoredThisGK {
		return nil
	}

	// Step 3 – Determine penalty tier.
	count := us.ActorGKsAtLocationThisMonth
	scale := c.getPenaltyScale(count)

	// Step 4 – Apply scale.
	yearMonth := event.LoggedAt.UTC().Format("2006-01")

	if scale == 1.0 {
		// No penalty; record this GK at the location.
		if err := c.store.RecordActorGKAtLocation(ctx, event.UserID, locationID, yearMonth, event.GKID, event.LoggedAt); err != nil {
			return fmt.Errorf("recording waypoint location: %w", err)
		}
		return nil
	}

	if scale == 0.0 {
		acc.ZeroByLabel("base_move")
		pipeCtx.RuntimeFlags.ActorScoredThisGK = false
		pipeCtx.RuntimeFlags.BasePointsAwarded = 0
		// Do NOT increment location counter since 0 points earned.
		return nil
	}

	// Partial penalty: scale down.
	acc.ScaleByLabel("base_move", scale)
	pipeCtx.RuntimeFlags.BasePointsAwarded *= scale

	// Step 5 – Record this GK at location (since scale > 0).
	if err := c.store.RecordActorGKAtLocation(ctx, event.UserID, locationID, yearMonth, event.GKID, event.LoggedAt); err != nil {
		return fmt.Errorf("recording waypoint location: %w", err)
	}

	return nil
}

// getPenaltyScale returns the penalty scale factor for the given prior GK count at the location.
func (c *WaypointPenalty) getPenaltyScale(count int) float64 {
	tiers := c.cfg.WaypointPenaltyTiers
	if count < len(tiers) {
		return tiers[count]
	}
	// Beyond last tier → 0.
	return 0.0
}
