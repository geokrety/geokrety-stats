package computers_test

import (
	"context"
	"time"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// testCfg returns a StatsConfig with testing-friendly defaults.
func testCfg() config.StatsConfig {
	return config.StatsConfig{
		BaseMovePoints:              3.0,
		MaxGKsPerOwner:              10,
		WaypointPenaltyTiers:        []float64{1.0, 0.5, 0.25, 0},
		CountryCrossingActorBonus:   3.0,
		CountryCrossingOwnerBonus:   1.0,
		RelayMoverBonus:             2.0,
		RelayDropperBonus:           1.0,
		RelayWindowHours:            168,
		RescuerGrabberBonus:         2.0,
		RescuerOwnerBonus:           1.0,
		RescuerDormancyMonths:       6,
		HandoverOwnerBonus:          1.0,
		ReachOwnerBonus:             5.0,
		ReachMilestoneUsers:         10,
		ReachWindowMonths:           6,
		ChainTimeoutDays:            14,
		ChainMinLength:              3,
		ChainAntiFarmingMonths:      6,
		ChainOwnerShareFraction:     0.25,
		DiversityDropsBonus:         3.0,
		DiversityDropsMilestone:     5,
		DiversityOwnersBonus:        7.0,
		DiversityOwnersMilestone:    10,
		DiversityCountryBonus:       5.0,
		MultiplierMin:               1.0,
		MultiplierMax:               2.0,
		MultiplierFirstMoveInc:      0.01,
		MultiplierCountryInc:        0.05,
		MultiplierInHandDecayPerDay: 0.008,
		MultiplierInCacheDecayPerWeek: 0.02,
		NonTransferableGKTypes:      []int{4},
		FirstFinderWindowHours:      24,
	}
}

// testEvent creates a base DROP event for testing.
func testEvent(userID, gkID int64, logType pipeline.LogType) pipeline.Event {
	return pipeline.Event{
		LogID:    1000,
		UserID:   userID,
		GKID:     gkID,
		LogType:  logType,
		Waypoint: "GC12345",
		Country:  "PL",
		Lat:      50.0,
		Lon:      20.0,
		LoggedAt: time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC),
	}
}

// testCtx returns a basic pipeline context with sensible defaults.
func testCtx(event pipeline.Event, ownerID int64, multiplier float64) *pipeline.Context {
	return &pipeline.Context{
		Event: event,
		GKState: pipeline.GKState{
			GKID:              event.GKID,
			GKType:            0, // standard transferable
			OwnerID:           ownerID,
			CurrentMultiplier: multiplier,
			LastMultiplierAt:  time.Time{},
			CurrentHolder:     0, // in cache
		},
		GKHistory: pipeline.GKHistory{
			CountriesVisited: map[string]bool{},
		},
		UserState:  pipeline.UserState{},
		ChainState: pipeline.ChainState{},
	}
}

// mockStore returns a default mock store instance.
func mockStore() *store.MockStore {
	return store.NewMockStore()
}

var _ = context.Background // ensure context is imported
