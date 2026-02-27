package computers

import (
	"context"
	"fmt"
	"time"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// RescuerBonus is computer 07.
// Awards a bonus when a non-owner user grabs a GeoKret that has been sitting
// dormant in a cache for 6 or more months.
type RescuerBonus struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewRescuerBonus creates a new RescuerBonus computer.
func NewRescuerBonus(s store.Store, cfg config.StatsConfig) *RescuerBonus {
	return &RescuerBonus{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *RescuerBonus) Name() string {
	return "07_rescuer_bonus"
}

// Process implements the Computer interface.
func (c *RescuerBonus) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState
	hist := &pipeCtx.GKHistory

	// Step 1 – Only GRAB events.
	if event.LogType != pipeline.LogTypeGrab {
		return nil
	}

	// Step 2 – GK must be in cache (not held by anyone).
	if !gks.IsInCache() {
		return nil
	}

	// Step 3 – Grabber must not be the owner.
	if event.UserID == gks.OwnerID {
		return nil
	}

	// Step 4 – Check dormancy: GK must have been in cache for >= RescuerDormancyMonths.
	if hist.LastCacheEntryAt == nil {
		return nil
	}

	// Compute months dormant using a simple approach: subtract months.
	dormancyThreshold := hist.LastCacheEntryAt.AddDate(0, c.cfg.RescuerDormancyMonths, 0)
	if event.LoggedAt.Before(dormancyThreshold) {
		return nil // Not dormant enough.
	}

	monthsDormant := monthsBetween(*hist.LastCacheEntryAt, event.LoggedAt)

	// Step 5 – Award rescuer bonus to the grabber.
	acc.Add(pipeline.Award{
		RecipientUserID: event.UserID,
		Points:          c.cfg.RescuerGrabberBonus,
		Reason: fmt.Sprintf("Rescuer bonus: GK #%d dormant in cache for %d months (grabber)",
			event.GKID, monthsDormant),
		Label:         "rescuer_grabber",
		ModuleSource:  c.Name(),
		IsOwnerReward: false,
	})

	// Step 6 – Award rescuer bonus to the owner.
	if gks.OwnerID != 0 {
		acc.Add(pipeline.Award{
			RecipientUserID: gks.OwnerID,
			Points:          c.cfg.RescuerOwnerBonus,
			Reason: fmt.Sprintf("Rescuer bonus: GK #%d rescued from cache by user #%d after %d months dormancy (owner)",
				event.GKID, event.UserID, monthsDormant),
			Label:         "rescuer_owner",
			ModuleSource:  c.Name(),
			IsOwnerReward: true,
		})
	}

	return nil
}

// monthsBetween returns the approximate number of whole months between two times.
func monthsBetween(from, to time.Time) int {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	return years*12 + months
}
