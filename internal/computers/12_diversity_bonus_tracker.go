package computers

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// DiversityBonusTracker is computer 12.
// Awards three one-time monthly diversity milestones for the acting user:
//
//	Bonus A (+3): reaching the 5th distinct scored DROP this calendar month.
//	Bonus B (+7): reaching the 10th distinct owner interaction this calendar month.
//	Bonus C (+5): first visit to a new country this calendar month (uses
//	              RuntimeFlags.NewCountryVisited set by computer 05).
type DiversityBonusTracker struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewDiversityBonusTracker creates a new DiversityBonusTracker computer.
func NewDiversityBonusTracker(s store.Store, cfg config.StatsConfig) *DiversityBonusTracker {
	return &DiversityBonusTracker{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *DiversityBonusTracker) Name() string {
	return "12_diversity_bonus_tracker"
}

// Process implements the Computer interface.
func (c *DiversityBonusTracker) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	// Only scored moves (non-zero base points) qualify for diversity tracking.
	if !acc.HasLabel("base_move") {
		return nil
	}

	event := &pipeCtx.Event
	us := &pipeCtx.UserState
	yearMonth := event.LoggedAt.Format("2006-01")

	// ── Bonus A: 5th distinct DROP this month ────────────────────────────────
	if event.LogType == pipeline.LogTypeDrop && !us.ActorGKAlreadyDroppedThisMonth {
		newDropCount := us.ActorGKsDroppedThisMonth + 1
		us.ActorGKsDroppedThisMonth = newDropCount

		log.Debug().
			Int64("user_id", event.UserID).
			Int("drop_count", newDropCount).
			Int("threshold", c.cfg.DiversityDropsMilestone).
			Str("year_month", yearMonth).
			Msg("tracking monthly drop count")

		if err := c.store.IncrementActorMonthlyDrops(ctx, event.UserID, yearMonth, event.GKID, event.LoggedAt); err != nil {
			return fmt.Errorf("incrementing monthly drops: %w", err)
		}

		if newDropCount == c.cfg.DiversityDropsMilestone && !us.DropsBonusAlreadyAwarded {
			acc.Add(pipeline.Award{
				RecipientUserID: event.UserID,
				Points:          c.cfg.DiversityDropsBonus,
				Label:           "diversity_drops",
				ModuleSource:    c.Name(),
			})
			us.DropsBonusAlreadyAwarded = true

			if err := c.store.SetDropsBonusAwarded(ctx, event.UserID, yearMonth); err != nil {
				return fmt.Errorf("setting drops bonus awarded: %w", err)
			}

			log.Debug().
				Int64("user_id", event.UserID).
				Str("year_month", yearMonth).
				Msg("diversity bonus A awarded (5th drop)")
		}
	}

	// ── Bonus B: 10th distinct owner interaction this month ──────────────────
	ownerID := pipeCtx.GKState.OwnerID
	if ownerID != 0 && ownerID != event.UserID && !us.ActorOwnerAlreadyCountedThisMonth && !us.OwnersBonusAlreadyAwarded {
		newOwnerCount := us.ActorDistinctOwnersThisMonth + 1
		us.ActorDistinctOwnersThisMonth = newOwnerCount

		log.Debug().
			Int64("user_id", event.UserID).
			Int("owner_count", newOwnerCount).
			Int("threshold", c.cfg.DiversityOwnersMilestone).
			Msg("tracking monthly distinct owner count")

		if err := c.store.IncrementActorMonthlyOwners(ctx, event.UserID, yearMonth, ownerID, event.LoggedAt); err != nil {
			return fmt.Errorf("incrementing monthly owners: %w", err)
		}

		if newOwnerCount >= c.cfg.DiversityOwnersMilestone {
			acc.Add(pipeline.Award{
				RecipientUserID: event.UserID,
				Points:          c.cfg.DiversityOwnersBonus,
				Label:           "diversity_owners",
				ModuleSource:    c.Name(),
			})
			us.OwnersBonusAlreadyAwarded = true

			if err := c.store.SetOwnersBonusAwarded(ctx, event.UserID, yearMonth); err != nil {
				return fmt.Errorf("setting owners bonus awarded: %w", err)
			}

			log.Debug().
				Int64("user_id", event.UserID).
				Str("year_month", yearMonth).
				Msg("diversity bonus B awarded (10th owner)")
		}
	}

	// ── Bonus C: first new country for this GK, not yet visited by actor this month ──
	if pipeCtx.RuntimeFlags.NewCountryVisited {
		country := pipeCtx.RuntimeFlags.NewCountryCode

		alreadyThisMonth := us.ActorCountriesVisitedThisMonth[country]

		if !alreadyThisMonth && country != "" {
			if us.ActorCountriesVisitedThisMonth == nil {
				us.ActorCountriesVisitedThisMonth = make(map[string]bool)
			}
			us.ActorCountriesVisitedThisMonth[country] = true

			acc.Add(pipeline.Award{
				RecipientUserID: event.UserID,
				Points:          c.cfg.DiversityCountryBonus,
				Label:           "diversity_country",
				ModuleSource:    c.Name(),
			})
			log.Debug().
				Int64("user_id", event.UserID).
				Str("country", country).
				Str("year_month", yearMonth).
				Msg("diversity bonus C awarded (new country this month)")

			if err := c.store.RecordActorDiversityCountry(ctx, event.UserID, yearMonth, country, event.LoggedAt); err != nil {
				return fmt.Errorf("recording diversity country: %w", err)
			}
		}
	}

	return nil
}
