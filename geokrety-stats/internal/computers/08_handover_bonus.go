package computers

import (
	"context"
	"fmt"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// HandoverBonus is computer 08.
// Awards the GK owner a small bonus every time their standard-type GeoKret
// changes hands between non-owner users.
type HandoverBonus struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewHandoverBonus creates a new HandoverBonus computer.
func NewHandoverBonus(s store.Store, cfg config.StatsConfig) *HandoverBonus {
	return &HandoverBonus{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *HandoverBonus) Name() string {
	return "08_handover_bonus"
}

// Process implements the Computer interface.
func (c *HandoverBonus) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState

	// Step 1 – Only GRAB events.
	if event.LogType != pipeline.LogTypeGrab {
		return nil
	}

	// Step 2 – Only standard GK types (non-transferable GKs stay with owner).
	if c.cfg.IsNonTransferable(gks.GKType) {
		return nil
	}

	// Step 3 – The new grabber must not be the owner.
	if event.UserID == gks.OwnerID {
		return nil
	}

	// Step 4 – The GK must currently be held by someone (not in cache).
	// A handover: taking from another user's hands (not from cache).
	// Note: for cache→user grab, this module does not fire (no handover in that case).
	// The "current_holder" is who had it BEFORE this grab.
	if gks.IsInCache() {
		return nil // Taking from cache is not a handover between users.
	}

	// Step 5 – The previous holder must also not be the owner.
	// Handover is specifically: non-owner A to non-owner B.
	if gks.CurrentHolder == gks.OwnerID {
		return nil
	}

	// Step 6 – Award handover bonus to the owner.
	if gks.OwnerID != 0 {
		acc.Add(pipeline.Award{
			RecipientUserID: gks.OwnerID,
			Points:          c.cfg.HandoverOwnerBonus,
			Reason: fmt.Sprintf("Handover bonus: GK #%d changed hands from user #%d to user #%d (owner)",
				event.GKID, gks.CurrentHolder, event.UserID),
			Label:         "handover_owner",
			ModuleSource:  c.Name(),
			IsOwnerReward: true,
		})
	}

	return nil
}
