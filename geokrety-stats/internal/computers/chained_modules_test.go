package computers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestChainedModules_BaseLimitPenaltyCountryDiversityStack(t *testing.T) {
	s := newSpyStore()
	cfg := testCfg()
	base := computers.NewBaseMovePoints(s, cfg)
	ownerLimit := computers.NewOwnerGKLimitFilter(s, cfg)
	waypoint := computers.NewWaypointPenalty(s, cfg)
	country := computers.NewCountryCrossing(s, cfg)
	diversity := computers.NewDiversityBonusTracker(s, cfg)

	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	evt.Country = "DE"
	evt.LoggedAt = time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	pipeCtx := testCtx(evt, 99, 1.2)
	pipeCtx.GKState.HomeCountry = "PL"
	pipeCtx.GKHistory.CountriesVisited = map[string]bool{"PL": true}
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	pipeCtx.UserState.ActorGKsPerOwnerCount = 9
	pipeCtx.UserState.ActorGKAlreadyCountedForOwner = false
	pipeCtx.UserState.ActorGKsAtLocationThisMonth = 1
	pipeCtx.UserState.ActorGKsDroppedThisMonth = 4
	pipeCtx.UserState.ActorGKAlreadyDroppedThisMonth = false
	pipeCtx.UserState.ActorDistinctOwnersThisMonth = 9
	pipeCtx.UserState.ActorOwnerAlreadyCountedThisMonth = false
	pipeCtx.UserState.ActorCountriesVisitedThisMonth = map[string]bool{}
	acc := pipeline.NewAccumulator()

	require.NoError(t, base.Process(context.Background(), pipeCtx, acc))
	require.NoError(t, ownerLimit.Process(context.Background(), pipeCtx, acc))
	require.NoError(t, waypoint.Process(context.Background(), pipeCtx, acc))
	require.NoError(t, country.Process(context.Background(), pipeCtx, acc))
	require.NoError(t, diversity.Process(context.Background(), pipeCtx, acc))

	baseAward := mustAwardByLabel(acc, "base_move")
	assert.InDelta(t, 1.8, baseAward.Points, 0.0001) // 3*1.2*50%
	assert.Equal(t, 3.0, mustAwardByLabel(acc, "country_crossing_actor").Points)
	assert.Equal(t, 1.0, mustAwardByLabel(acc, "country_crossing_owner").Points)
	assert.Equal(t, 3.0, mustAwardByLabel(acc, "diversity_drops").Points)
	assert.Equal(t, 7.0, mustAwardByLabel(acc, "diversity_owners").Points)
	assert.Equal(t, 5.0, mustAwardByLabel(acc, "diversity_country").Points)
}

func TestChainedModules_ChainStateThenChainBonus(t *testing.T) {
	s := newSpyStore()
	cfg := testCfg()
	state := computers.NewChainStateManager(s, cfg)
	bonus := computers.NewChainBonus(s, cfg)
	now := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)

	evt := testEvent(10, 500, pipeline.LogTypeArchived)
	evt.LoggedAt = now
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.ChainState.ActiveChainID = 77
	s.GetChainMembersFn = func(context.Context, int64) ([]int64, error) {
		return []int64{1, 2, 3}, nil
	}
	acc := pipeline.NewAccumulator()

	require.NoError(t, state.Process(context.Background(), pipeCtx, acc))
	require.True(t, pipeCtx.ChainState.ChainEnded)
	require.NoError(t, bonus.Process(context.Background(), pipeCtx, acc))

	assert.Equal(t, 9.0, mustAwardByLabel(acc, "chain_bonus").Points)
	assert.Equal(t, 6.75, mustAwardByLabel(acc, "chain_bonus_owner").Points)
}
