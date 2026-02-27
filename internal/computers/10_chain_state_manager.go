package computers

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// ChainStateManager is computer 10.
// Maintains the state of the movement chain for this GeoKret.
// Does NOT award points — that is the responsibility of computer 11.
type ChainStateManager struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewChainStateManager creates a new ChainStateManager computer.
func NewChainStateManager(s store.Store, cfg config.StatsConfig) *ChainStateManager {
	return &ChainStateManager{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *ChainStateManager) Name() string {
	return "10_chain_state_manager"
}

// Process implements the Computer interface.
func (c *ChainStateManager) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	cs := &pipeCtx.ChainState

	// Step 1 – Check for expired chain before processing new event.
	if cs.ActiveChainID != 0 {
		if c.isChainExpired(cs.ChainLastActive, event.LoggedAt) {
			log.Debug().
				Int64("chain_id", cs.ActiveChainID).
				Int64("gk_id", event.GKID).
				Msg("chain expired before new event; finalizing")

			// Signal chain ended (to be processed by computer 11).
			members, err := c.store.GetChainMembers(ctx, cs.ActiveChainID)
			if err != nil {
				return fmt.Errorf("loading chain members on expiry: %w", err)
			}

			cs.ChainEnded = true
			cs.EndedChainID = cs.ActiveChainID
			cs.EndedChainMembers = members

			// End the chain in the store.
			if err := c.store.EndChain(ctx, cs.ActiveChainID, event.LoggedAt, "timeout"); err != nil {
				return fmt.Errorf("ending expired chain: %w", err)
			}

			// Reset: start fresh below.
			cs.ActiveChainID = 0
			cs.ChainMembers = nil
			cs.ChainLastActive = time.Time{}
			cs.HolderAcquiredAt = time.Time{}
		}
	}

	// Step 2 – Handle ARCHIVED event.
	if event.LogType == pipeline.LogTypeArchived {
		if cs.ActiveChainID != 0 {
			members, err := c.store.GetChainMembers(ctx, cs.ActiveChainID)
			if err != nil {
				return fmt.Errorf("loading chain members on archive: %w", err)
			}

			cs.ChainEnded = true
			cs.EndedChainID = cs.ActiveChainID
			cs.EndedChainMembers = members

			if err := c.store.EndChain(ctx, cs.ActiveChainID, event.LoggedAt, "archived"); err != nil {
				return fmt.Errorf("ending chain on archive: %w", err)
			}
			cs.ActiveChainID = 0
		}
		return nil // No new chain for ARCHIVED events.
	}

	// Step 3 – Handle COMMENT event (should not reach here due to event guard, but defensive).
	if event.LogType == pipeline.LogTypeComment {
		return nil
	}

	// Step 4 – Handle DIP event (timer extension only).
	if event.LogType == pipeline.LogTypeDip {
		if cs.ActiveChainID == 0 {
			return nil // No active chain to extend.
		}

		timeoutDur := time.Duration(c.cfg.ChainTimeoutDays) * 24 * time.Hour

		currentDeadline := cs.ChainLastActive.Add(timeoutDur)
		maxDeadline := cs.HolderAcquiredAt.Add(timeoutDur)

		// Extend by 1 day, capped at maxDeadline.
		newDeadline := currentDeadline.Add(24 * time.Hour)
		if newDeadline.After(maxDeadline) {
			newDeadline = maxDeadline
		}

		// Set chain_last_active such that the 14-day window ends at newDeadline.
		newLastActive := newDeadline.Add(-timeoutDur)
		cs.ChainLastActive = newLastActive

		if err := c.store.UpdateChainLastActive(ctx, cs.ActiveChainID, newLastActive, nil); err != nil {
			return fmt.Errorf("updating chain last active on DIP: %w", err)
		}

		return nil // DIP does not add members.
	}

	// Step 5 – Handle GRAB, DROP, SEEN: full timer reset.
	// For GRAB, DROP, SEEN: reset the 14-day countdown.

	// Step 7 – Ensure chain exists.
	if cs.ActiveChainID == 0 {
		chainID, err := c.store.CreateChain(ctx, event.GKID, event.LoggedAt)
		if err != nil {
			return fmt.Errorf("creating new chain: %w", err)
		}
		cs.ActiveChainID = chainID
		cs.ChainMembers = nil
		cs.ChainLastActive = event.LoggedAt
		cs.HolderAcquiredAt = event.LoggedAt
	} else {
		// Reset timer for existing chain.
		cs.ChainLastActive = event.LoggedAt

		// Update holder_acquired_at on GRAB (new holder takes possession).
		var holderAcquiredAt *time.Time
		if event.LogType == pipeline.LogTypeGrab {
			t := event.LoggedAt
			holderAcquiredAt = &t
			cs.HolderAcquiredAt = t
		}

		if err := c.store.UpdateChainLastActive(ctx, cs.ActiveChainID, event.LoggedAt, holderAcquiredAt); err != nil {
			return fmt.Errorf("updating chain last active: %w", err)
		}
	}

	// Step 6 – Update chain members (DROP and GRAB only).
	if event.LogType == pipeline.LogTypeDrop || event.LogType == pipeline.LogTypeGrab {
		// Self-grab check: don't add member for self-grab (grab where user == current holder).
		isSelfGrab := event.LogType == pipeline.LogTypeGrab &&
			event.UserID == pipeCtx.GKState.CurrentHolder

		if !isSelfGrab {
			alreadyMember := false
			for _, uid := range cs.ChainMembers {
				if uid == event.UserID {
					alreadyMember = true
					break
				}
			}
			if !alreadyMember {
				if err := c.store.AddChainMember(ctx, cs.ActiveChainID, event.UserID, event.LoggedAt); err != nil {
					return fmt.Errorf("adding chain member: %w", err)
				}
				cs.ChainMembers = append(cs.ChainMembers, event.UserID)
			}
		}
	}

	return nil
}

// isChainExpired returns true if the chain has been inactive for >= ChainTimeoutDays.
func (c *ChainStateManager) isChainExpired(lastActive, now time.Time) bool {
	if lastActive.IsZero() {
		return true
	}
	elapsed := now.Sub(lastActive)
	return elapsed >= time.Duration(c.cfg.ChainTimeoutDays)*24*time.Hour
}
