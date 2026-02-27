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

// ChainBonus is computer 11.
// Awards bonus points when a chain ends with at least cfg.ChainMinLength members.
// Formula: bonus_per_user = min(N², 8·N) where N = chain length.
// Anti-farming: a user who already received a chain bonus in the last 6 months gets nothing.
// Owner also receives 25 % of the total distributed to non-owner participants.
type ChainBonus struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewChainBonus creates a new ChainBonus computer.
func NewChainBonus(s store.Store, cfg config.StatsConfig) *ChainBonus {
	return &ChainBonus{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *ChainBonus) Name() string {
	return "11_chain_bonus"
}

// Process implements the Computer interface.
func (c *ChainBonus) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	cs := &pipeCtx.ChainState

	// Nothing to do unless computer 10 signalled chain end.
	if !cs.ChainEnded || cs.EndedChainID == 0 {
		return nil
	}

	members := cs.EndedChainMembers
	n := len(members)

	if n < c.cfg.ChainMinLength {
		log.Debug().
			Int("members", n).
			Int("min_length", c.cfg.ChainMinLength).
			Int64("chain_id", cs.EndedChainID).
			Msg("chain too short for bonus")
		return nil
	}

	bonusPerUser := c.bonusPerUser(n)

	// Determine GK owner for the owner-share calculation.
	ownerID := pipeCtx.GKState.OwnerID

	// Award points to each qualifying member.
	var totalNonOwnerPoints float64
	cooldownCutoff := pipeCtx.Event.LoggedAt.AddDate(0, -c.cfg.ChainAntiFarmingMonths, 0)

	for _, userID := range members {
		lastBonus, err := c.store.GetChainLastBonus(ctx, userID, pipeCtx.Event.GKID)
		if err != nil {
			return fmt.Errorf("checking chain last bonus for user %d: %w", userID, err)
		}

		if !lastBonus.IsZero() && lastBonus.After(cooldownCutoff) {
			log.Debug().
				Int64("user_id", userID).
				Time("last_bonus", lastBonus).
				Msg("chain bonus withheld (cooldown active)")
			continue
		}

		if userID == ownerID {
			// Owner share calculated below from total non-owner points.
			continue
		}

		acc.Add(pipeline.Award{
			RecipientUserID: userID,
			Points:          bonusPerUser,
			Label:           "chain_bonus",
			ModuleSource:    c.Name(),
		})
		totalNonOwnerPoints += bonusPerUser
	}

	// Owner receives 25 % of non-owner total.
	if totalNonOwnerPoints > 0 && ownerID != 0 {
		ownerShare := math.Round(totalNonOwnerPoints*c.cfg.ChainOwnerShareFraction*100) / 100
		if ownerShare > 0 {
			acc.Add(pipeline.Award{
				RecipientUserID: ownerID,
				Points:          ownerShare,
				Label:           "chain_bonus_owner",
				ModuleSource:    c.Name(),
				IsOwnerReward:   true,
			})
		}
	}

	// Record completion per qualifying member for cooldown tracking.
	for _, userID := range members {
		if userID == ownerID {
			continue
		}
		if err := c.store.RecordChainCompletion(ctx, userID, pipeCtx.Event.GKID, cs.EndedChainID, pipeCtx.Event.LoggedAt); err != nil {
			return fmt.Errorf("recording chain completion for user %d: %w", userID, err)
		}
	}

	return nil
}

// bonusPerUser computes min(N², 8·N).
func (c *ChainBonus) bonusPerUser(n int) float64 {
	sq := float64(n * n)
	linear := float64(8 * n)
	if sq < linear {
		return sq
	}
	return linear
}

// AwardTimeoutBonus is called by the maintenance scheduler for chains that expire
// outside of regular move processing (no active event in the pipeline).
// It evaluates the stored chain membership and awards points accordingly,
// then marks the chain as ended.
func (c *ChainBonus) AwardTimeoutBonus(ctx context.Context, chainID int64, gkID int64, now time.Time) error {
	members, err := c.store.GetChainMembers(ctx, chainID)
	if err != nil {
		return fmt.Errorf("loading chain members for timeout: %w", err)
	}

	n := len(members)
	if n < c.cfg.ChainMinLength {
		return c.store.EndChain(ctx, chainID, now, "timeout_short")
	}

	bonusPerUser := c.bonusPerUser(n)
	cooldownCutoff := now.AddDate(0, -c.cfg.ChainAntiFarmingMonths, 0)

	var finals []pipeline.FinalAward

	for _, userID := range members {
		lastBonus, err := c.store.GetChainLastBonus(ctx, userID, gkID)
		if err != nil {
			return fmt.Errorf("checking chain last bonus (timeout) for user %d: %w", userID, err)
		}
		if !lastBonus.IsZero() && lastBonus.After(cooldownCutoff) {
			continue
		}
		finals = append(finals, pipeline.FinalAward{
			RecipientUserID: userID,
			TotalPoints:     bonusPerUser,
			GKID:            gkID,
			Awards: []pipeline.Award{{
				RecipientUserID: userID,
				Points:          bonusPerUser,
				Label:           "chain_bonus",
				ModuleSource:    "11_chain_bonus",
			}},
		})
		if err := c.store.RecordChainCompletion(ctx, userID, gkID, chainID, now); err != nil {
			return fmt.Errorf("recording timeout chain completion for user %d: %w", userID, err)
		}
	}

	if len(finals) > 0 {
		if err := c.store.SaveAwards(ctx, finals); err != nil {
			return fmt.Errorf("saving timeout chain awards: %w", err)
		}
	}

	return c.store.EndChain(ctx, chainID, now, "timeout")
}
