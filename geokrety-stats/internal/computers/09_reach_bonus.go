package computers

import (
	"context"
	"fmt"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// ReachBonus is computer 09.
// Awards the GK owner a one-time bonus each time their GeoKret reaches the
// milestone of ReachMilestoneUsers distinct users within a rolling 6-month window.
type ReachBonus struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewReachBonus creates a new ReachBonus computer.
func NewReachBonus(s store.Store, cfg config.StatsConfig) *ReachBonus {
	return &ReachBonus{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *ReachBonus) Name() string {
	return "09_reach_bonus"
}

// Process implements the Computer interface.
func (c *ReachBonus) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState
	hist := &pipeCtx.GKHistory

	// Step 1 – Only applies if actor earned base points this event.
	if !pipeCtx.RuntimeFlags.ActorScoredThisGK {
		return nil
	}

	// Step 2 – Actor must not be the owner.
	if event.UserID == gks.OwnerID {
		return nil
	}

	// Step 3 – Check if actor is already in the 6-month window.
	for _, uid := range hist.DistinctUsers6M {
		if uid == event.UserID {
			return nil // Already counted; this move doesn't add a new person.
		}
	}

	// Step 4 – Compute new distinct count.
	previousCount := len(hist.DistinctUsers6M)
	newCount := previousCount + 1

	// Step 5 – Check milestone: bonus fires when count crosses ReachMilestoneUsers.
	milestone := c.cfg.ReachMilestoneUsers
	if !(previousCount < milestone && newCount >= milestone) {
		return nil // Milestone not reached this event.
	}

	// Step 6 – Award reach bonus to the owner.
	if gks.OwnerID != 0 {
		acc.Add(pipeline.Award{
			RecipientUserID: gks.OwnerID,
			Points:          c.cfg.ReachOwnerBonus,
			Reason: fmt.Sprintf("Reach bonus: GK #%d reached %d distinct users in the last 6 months (owner)",
				event.GKID, milestone),
			Label:         "reach_owner",
			ModuleSource:  c.Name(),
			IsOwnerReward: true,
		})
	}

	return nil
}
