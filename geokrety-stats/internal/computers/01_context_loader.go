package computers

import (
	"context"
	"fmt"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// ContextLoader is computer 01.
// Loads all persistent state required by downstream computers in one place.
// No other computer should independently query the data store.
type ContextLoader struct {
	store store.Store
	cfg   config.StatsConfig
}

// NewContextLoader creates a new ContextLoader computer.
func NewContextLoader(s store.Store, cfg config.StatsConfig) *ContextLoader {
	return &ContextLoader{store: s, cfg: cfg}
}

// Name returns the computer's name.
func (c *ContextLoader) Name() string {
	return "01_context_loader"
}

// Process implements the Computer interface.
// Populates pipeCtx with all state needed by downstream computers.
func (c *ContextLoader) Process(ctx context.Context, pipeCtx *pipeline.Context, acc *pipeline.Accumulator) error {
	event := &pipeCtx.Event

	// 1 – Load GK core state
	gk, err := c.store.GetGK(ctx, event.GKID)
	if err != nil {
		return fmt.Errorf("loading GK %d: %w", event.GKID, err)
	}
	if gk == nil {
		return fmt.Errorf("GK %d not found", event.GKID)
	}

	ownerID := int64(0)
	if gk.Owner != nil {
		ownerID = *gk.Owner
	}

	// 2 – GK holder state
	currentHolder := int64(0)
	if gk.Holder != nil {
		currentHolder = *gk.Holder
	}
	// TODO – load previous holder efficiently (see store method comment)
	previousHolder := int64(0)
	// previousHolder, err := c.store.GetGKPreviousHolder(ctx, event.GKID)
	// if err != nil {
	// 	return fmt.Errorf("loading previous holder for GK %d: %w", event.GKID, err)
	// }

	// 3 – GK multiplier state
	multiplier, lastMultAt, holderID, holderAcquiredAt, err := c.store.GetGKMultiplierState(ctx, event.GKID)
	if err != nil {
		return fmt.Errorf("loading multiplier state for GK %d: %w", event.GKID, err)
	}
	_ = holderID // loaded for maintenance but stored separately

	// 4 – GK geographic history: home country
	homeCountry, err := c.store.GetGKHomeCountry(ctx, event.GKID)
	if err != nil {
		return fmt.Errorf("loading home country for GK %d: %w", event.GKID, err)
	}

	// Load countries visited (from stats schema, which we maintain)
	countriesVisited, err := c.store.GetGKCountriesVisited(ctx, event.GKID)
	if err != nil {
		return fmt.Errorf("loading countries visited for GK %d: %w", event.GKID, err)
	}

	// 5 – GK last-activity timestamps
	lastDropAt, lastDropUser, err := c.store.GetGKLastDrop(ctx, event.GKID, event.LoggedAt)
	if err != nil {
		return fmt.Errorf("loading last drop for GK %d: %w", event.GKID, err)
	}

	lastCacheEntryAt, err := c.store.GetGKLastCacheEntry(ctx, event.GKID, event.LoggedAt)
	if err != nil {
		return fmt.Errorf("loading last cache entry for GK %d: %w", event.GKID, err)
	}

	// 6 – Distinct users in 6-month window (before this event)
	distinctUsers6M, err := c.store.GetGKDistinctUsers6M(ctx, event.GKID, event.LoggedAt)
	if err != nil {
		return fmt.Errorf("loading distinct users 6m for GK %d: %w", event.GKID, err)
	}

	// 7 – Actor move history on this GK
	actorMoveHistory, err := c.store.GetUserMoveHistoryOnGK(ctx, event.UserID, event.GKID)
	if err != nil {
		return fmt.Errorf("loading move history for user %d on GK %d: %w", event.UserID, event.GKID, err)
	}

	// 8 – Actor owner-GK interaction count
	actorGKsPerOwnerCount := 0
	actorGKAlreadyCounted := false
	if ownerID != 0 {
		actorGKsPerOwnerCount, err = c.store.GetActorGKsPerOwnerCount(ctx, event.UserID, ownerID)
		if err != nil {
			return fmt.Errorf("loading owner GK count for user %d: %w", event.UserID, err)
		}
		actorGKAlreadyCounted, err = c.store.IsGKAlreadyCountedForOwner(ctx, event.UserID, ownerID, event.GKID)
		if err != nil {
			return fmt.Errorf("checking GK counted status: %w", err)
		}
	}

	// 9 – Actor waypoint activity this month
	actorGKsAtLocation := 0
	locationID := event.LocationID()
	yearMonth := event.LoggedAt.UTC().Format("2006-01")

	if locationID != "" {
		actorGKsAtLocation, err = c.store.GetActorGKsAtLocationThisMonth(ctx, event.UserID, locationID, yearMonth)
		if err != nil {
			return fmt.Errorf("loading waypoint counts for user %d: %w", event.UserID, err)
		}
	}

	// 10 – Monthly diversity state
	dropsCount, dropsBonusAwarded, ownersCount, ownersBonusAwarded, err :=
		c.store.GetActorMonthlyDiversity(ctx, event.UserID, yearMonth)
	if err != nil {
		return fmt.Errorf("loading monthly diversity for user %d: %w", event.UserID, err)
	}

	actorCountriesThisMonth, err := c.store.GetActorMonthlyDiversityCountries(ctx, event.UserID, yearMonth)
	if err != nil {
		return fmt.Errorf("loading diversity countries for user %d: %w", event.UserID, err)
	}

	actorOwnersThisMonth, err := c.store.GetActorMonthlyDiversityOwners(ctx, event.UserID, yearMonth)
	if err != nil {
		return fmt.Errorf("loading diversity owners for user %d: %w", event.UserID, err)
	}

	actorDropsThisMonth, err := c.store.GetActorMonthlyDiversityDrops(ctx, event.UserID, yearMonth)
	if err != nil {
		return fmt.Errorf("loading diversity drops for user %d: %w", event.UserID, err)
	}

	gkAlreadyDroppedThisMonth := actorDropsThisMonth[event.GKID]
	ownerAlreadyCountedThisMonth := actorOwnersThisMonth[ownerID]

	// 11 – Active chain state for this GK
	chainState := pipeline.ChainState{}
	activeChain, err := c.store.GetActiveChain(ctx, event.GKID)
	if err != nil {
		return fmt.Errorf("loading active chain for GK %d: %w", event.GKID, err)
	}
	if activeChain != nil {
		members, err := c.store.GetChainMembers(ctx, activeChain.ID)
		if err != nil {
			return fmt.Errorf("loading chain members for chain %d: %w", activeChain.ID, err)
		}
		chainState.ActiveChainID = activeChain.ID
		chainState.ChainMembers = members
		chainState.ChainLastActive = activeChain.ChainLastActive
		if activeChain.HolderAcquiredAt != nil {
			chainState.HolderAcquiredAt = *activeChain.HolderAcquiredAt
		}
	}

	// Populate pipeCtx
	pipeCtx.GKState = pipeline.GKState{
		GKID:              gk.ID,
		GKType:            gk.Type,
		OwnerID:           ownerID,
		CreatedAt:         gk.CreatedAt,
		CurrentMultiplier: multiplier,
		LastMultiplierAt:  lastMultAt,
		CurrentHolder:     currentHolder,
		PreviousHolder:    previousHolder,
		HomeCountry:       homeCountry,
	}

	pipeCtx.GKHistory = pipeline.GKHistory{
		LastDropAt:       lastDropAt,
		LastDropUser:     lastDropUser,
		LastCacheEntryAt: lastCacheEntryAt,
		DistinctUsers6M:  distinctUsers6M,
		CountriesVisited: countriesVisited,
	}

	pipeCtx.UserState = pipeline.UserState{
		ActorMoveHistoryOnGK:             actorMoveHistory,
		ActorGKsPerOwnerCount:            actorGKsPerOwnerCount,
		ActorGKAlreadyCountedForOwner:    actorGKAlreadyCounted,
		ActorGKsAtLocationThisMonth:      actorGKsAtLocation,
		ActorCountriesVisitedThisMonth:   actorCountriesThisMonth,
		ActorGKsDroppedThisMonth:         dropsCount,
		ActorGKAlreadyDroppedThisMonth:   gkAlreadyDroppedThisMonth,
		ActorDistinctOwnersThisMonth:     ownersCount,
		ActorOwnerAlreadyCountedThisMonth: ownerAlreadyCountedThisMonth,
		DropsBonusAlreadyAwarded:         dropsBonusAwarded,
		OwnersBonusAlreadyAwarded:        ownersBonusAwarded,
	}

	pipeCtx.ChainState = chainState
	_ = holderAcquiredAt // already loaded into chain state

	return nil
}
