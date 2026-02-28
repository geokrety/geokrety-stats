package computers

import (
	"context"
	"fmt"
	"time"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// BaseMovePoints is computer 02.
// Calculates the raw base points earned by the actor for the current event.
// Base points are multiplied by the GK's current multiplier.
type BaseMovePoints struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewBaseMovePoints creates a new BaseMovePoints computer.
func NewBaseMovePoints(s store.Store, cfg config.StatsConfig) *BaseMovePoints {
	return &BaseMovePoints{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *BaseMovePoints) Name() string {
	return "02_base_move_points"
}

// Process implements the Computer interface.
func (c *BaseMovePoints) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState
	us := &pipeCtx.UserState

	isOwner := event.UserID == gks.OwnerID
	isNonTransferable := c.cfg.IsNonTransferable(gks.GKType)

	// Step 1a – ARCHIVED always produces 0 base points.
	if event.LogType == pipeline.LogTypeArchived {
		c.setNoPoints(pipeCtx)
		return nil
	}

	// Step 1b – DIP produces 0 points except for owner of non-transferable GK.
	if event.LogType == pipeline.LogTypeDip {
		if !(isOwner && isNonTransferable) {
			c.setNoPoints(pipeCtx)
			return nil
		}
	}

	// Step 1c – Self-grab: grab where grabber already holds the GK → treated as DIP.
	if event.LogType == pipeline.LogTypeGrab {
		if event.UserID == gks.CurrentHolder {
			c.setNoPoints(pipeCtx)
			return nil
		}
	}

	// Step 2 – Waypoint requirement for location-based moves.
	if event.LogType.RequiresWaypoint() && !event.HasWaypoint() {
		c.setNoPoints(pipeCtx)
		return nil
	}

	// Step 3 – is_owner already set above.

	// Step 4 – Owner of standard GK always earns 0 base points.
	if isOwner && !isNonTransferable {
		c.setNoPoints(pipeCtx)
		return nil
	}

	// Step 5 – First-move check: actor earns base points at most once per log_type per GK.
	isFirstMove := !us.ActorMoveHistoryOnGK[event.LogType]
	if !isFirstMove {
		c.setNoPoints(pipeCtx)
		return nil
	}

	// Step 6 – First Finder window (for new GKs only, < 7 days old).
	// Standard check: if GK is new AND flag applies. For new GKs within window, normal +3 applies.
	// No separate bonus — the +3 × multiplier is the same; this is just a gating note.
	// Nothing to do here; standard flow continues.
	_ = c.isWithinFirstFinderWindow(event.LoggedAt, gks.CreatedAt)

	// Step 7 – Non-transferable GK owner monthly limit.
	// Once per (gk_type, waypoint, calendar month) combination.
	if isOwner && isNonTransferable {
		if err := c.checkNonTransferableOwnerLimit(ctx, pipeCtx, acc); err != nil {
			return err
		}
		// Check if zeroed by limit
		if !pipeCtx.RuntimeFlags.ActorScoredThisGK && pipeCtx.RuntimeFlags.BasePointsLabel == "" {
			// Not set yet, either we pass or zero (handled inside checkNonTransferableOwnerLimit)
		}
		if !pipeCtx.RuntimeFlags.ActorScoredThisGK && pipeCtx.RuntimeFlags.BasePointsAwarded == 0 && pipeCtx.RuntimeFlags.BasePointsLabel == "zeroed_non_transferable_limit" {
			return nil
		}
	}

	// Step 8 – Compute base points.
	basePoints := c.cfg.BaseMovePoints * gks.CurrentMultiplier

	acc.Add(pipeline.Award{
		RecipientUserID: event.UserID,
		Points:          basePoints,
		Reason: fmt.Sprintf("First %s of GK #%d by user #%d (multiplier: %.2fx)",
			event.LogType, event.GKID, event.UserID, gks.CurrentMultiplier),
		Label:          "base_move",
		ModuleSource:   c.Name(),
		IsOwnerReward:  false,
	})

	pipeCtx.RuntimeFlags.BasePointsAwarded = basePoints
	pipeCtx.RuntimeFlags.BasePointsLabel = "base_move"
	pipeCtx.RuntimeFlags.ActorScoredThisGK = true

	return nil
}

func (c *BaseMovePoints) setNoPoints(pipeCtx *pipeline.Context) {
	pipeCtx.RuntimeFlags.BasePointsAwarded = 0
	pipeCtx.RuntimeFlags.ActorScoredThisGK = false
}

// isWithinFirstFinderWindow returns true if the event is within the first-finder window.
func (c *BaseMovePoints) isWithinFirstFinderWindow(loggedAt, createdAt time.Time) bool {
	windowDuration := time.Duration(c.cfg.FirstFinderWindowHours) * time.Hour
	return loggedAt.Before(createdAt.Add(windowDuration))
}

// checkNonTransferableOwnerLimit checks if the owner has already earned base points
// for this (gk_type, waypoint, month) combination. Zeroes out if hit.
func (c *BaseMovePoints) checkNonTransferableOwnerLimit(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState
	yearMonth := event.LoggedAt.UTC().Format("2006-01")

	// Determine the location key for the monthly limit.
	// null waypoint = treat as unique per-event (no deduplication).
	if !event.HasWaypoint() {
		// No waypoint — no deduplication possible, allow scoring.
		return nil
	}

	// Build a key: (gk_type, waypoint, year_month)
	// We check via the waypoint monthly counts table using a special "owner_nontransfer" location prefix.
	locationKey := fmt.Sprintf("owner_nt_%d_%s", gks.GKType, event.Waypoint)

	count, err := c.store.GetActorGKsAtLocationThisMonth(ctx, event.UserID, locationKey, yearMonth)
	if err != nil {
		return fmt.Errorf("checking non-transferable limit: %w", err)
	}

	if count > 0 {
		// Already scored this combo this month — zero out.
		pipeCtx.RuntimeFlags.BasePointsAwarded = 0
		pipeCtx.RuntimeFlags.ActorScoredThisGK = false
		pipeCtx.RuntimeFlags.BasePointsLabel = "zeroed_non_transferable_limit"
		return nil
	}

	return nil
}
