package computers

import (
	"context"
	"fmt"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// CountryCrossing is computer 05.
// Detects when a GeoKret enters a country it has never visited before (excluding
// its home country) and awards points to both the actor and the GK owner.
// Also sets the new_country_visited runtime flag for module 12.
type CountryCrossing struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewCountryCrossing creates a new CountryCrossing computer.
func NewCountryCrossing(s store.Store, cfg config.StatsConfig) *CountryCrossing {
	return &CountryCrossing{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *CountryCrossing) Name() string {
	return "05_country_crossing"
}

// Process implements the Computer interface.
func (c *CountryCrossing) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event
	gks := &pipeCtx.GKState

	// Step 1 – Eligible log types: DROP, DIP, SEEN.
	if event.LogType != pipeline.LogTypeDrop &&
		event.LogType != pipeline.LogTypeDip &&
		event.LogType != pipeline.LogTypeSeen {
		return nil
	}

	// Step 2 – Country data required.
	if !event.HasCountry() {
		return nil
	}

	// Step 3 – Skip home country.
	if event.Country == gks.HomeCountry {
		return nil
	}

	// Step 4 – Check if country is new for this GK.
	if pipeCtx.GKHistory.CountriesVisited[event.Country] {
		return nil // Already visited; one-time reward only.
	}

	// Step 5 – New country confirmed: set runtime flags.
	pipeCtx.RuntimeFlags.NewCountryVisited = true
	pipeCtx.RuntimeFlags.NewCountryCode = event.Country

	// Persist new country to stats schema.
	if err := c.store.AddGKCountryVisited(ctx, event.GKID, event.Country, event.LoggedAt, event.LogID); err != nil {
		return fmt.Errorf("recording new country visited: %w", err)
	}

	isOwner := event.UserID == gks.OwnerID
	isNonTransferable := c.cfg.IsNonTransferable(gks.GKType)

	// Step 6 – Determine actor award amount.
	var actorAward float64
	if isNonTransferable {
		if isOwner {
			actorAward = c.cfg.CountryCrossingNonTransferableBonus
		} else {
			actorAward = c.cfg.CountryCrossingActorBonus
		}
	} else {
		if isOwner {
			actorAward = c.cfg.CountryCrossingOwnerSelfBonus
		} else {
			actorAward = c.cfg.CountryCrossingActorBonus
		}
	}

	// Emit actor award.
	acc.Add(pipeline.Award{
		RecipientUserID: event.UserID,
		Points:          actorAward,
		Reason: fmt.Sprintf("Country crossing bonus: GK #%d entered new country %s (actor)",
			event.GKID, event.Country),
		Label:         "country_crossing_actor",
		ModuleSource:  c.Name(),
		IsOwnerReward: false,
	})

	// Step 7 – Owner award (only when actor != owner).
	if !isOwner && gks.OwnerID != 0 {
		acc.Add(pipeline.Award{
			RecipientUserID: gks.OwnerID,
			Points:          c.cfg.CountryCrossingOwnerBonus,
			Reason: fmt.Sprintf("Country crossing bonus: GK #%d entered new country %s (owner)",
				event.GKID, event.Country),
			Label:         "country_crossing_owner",
			ModuleSource:  c.Name(),
			IsOwnerReward: true,
		})
	}

	return nil
}
