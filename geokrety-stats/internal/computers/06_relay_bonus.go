package computers

import (
	"context"
	"fmt"
	"time"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// RelayBonus is computer 06.
// Awards a time-based circulation bonus when a GeoKret is quickly picked up
// by a new user shortly after being dropped (within 7 days).
type RelayBonus struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewRelayBonus creates a new RelayBonus computer.
func NewRelayBonus(s store.Store, cfg config.StatsConfig) *RelayBonus {
	return &RelayBonus{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *RelayBonus) Name() string {
	return "06_relay_bonus"
}

// Process implements the Computer interface.
func (c *RelayBonus) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState
	hist := &pipeCtx.GKHistory

	// Step 1 – Only GRAB events.
	if event.LogType != pipeline.LogTypeGrab {
		return nil
	}

	// Step 2 – Must have a previous drop.
	if hist.LastDropAt == nil {
		return nil
	}

	// Step 3 – GK must be in cache (not currently held by someone).
	if !gks.IsInCache() {
		return nil
	}

	// Step 4 – The new mover must differ from the previous dropper.
	if event.UserID == hist.LastDropUser {
		return nil
	}

	// Step 5 – Check the 7-day window.
	windowDuration := time.Duration(c.cfg.RelayWindowHours) * time.Hour
	if event.LoggedAt.Sub(*hist.LastDropAt) > windowDuration {
		return nil
	}

	// Step 6 – Award relay bonus to the new grabber (mover).
	acc.Add(pipeline.Award{
		RecipientUserID: event.UserID,
		Points:          c.cfg.RelayMoverBonus,
		Reason: fmt.Sprintf("Relay bonus: GK #%d grabbed within 7 days of last drop (mover)",
			event.GKID),
		Label:         "relay_mover",
		ModuleSource:  c.Name(),
		IsOwnerReward: false,
	})

	// Step 7 – Award relay bonus to the previous dropper.
	if hist.LastDropUser != 0 {
		acc.Add(pipeline.Award{
			RecipientUserID: hist.LastDropUser,
			Points:          c.cfg.RelayDropperBonus,
			Reason: fmt.Sprintf("Relay bonus: GK #%d grabbed within 7 days (previous dropper)",
				event.GKID),
			Label:         "relay_dropper",
			ModuleSource:  c.Name(),
			IsOwnerReward: false,
		})
	}

	return nil
}
