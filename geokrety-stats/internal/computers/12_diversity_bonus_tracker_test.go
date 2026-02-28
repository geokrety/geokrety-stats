package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestDiversityBonusTracker_Awards5thDropBonus(t *testing.T) {
	s := newSpyStore()
	module := computers.NewDiversityBonusTracker(s, testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.UserState.ActorGKsDroppedThisMonth = 4
	pipeCtx.UserState.ActorGKAlreadyDroppedThisMonth = false

	acc := pipeline.NewAccumulator()
	acc.Add(pipeline.Award{RecipientUserID: 10, Points: 3, Label: "base_move", ModuleSource: "02_base_move_points"})

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Equal(t, 3.0, mustAwardByLabel(acc, "diversity_drops").Points)
	assert.Len(t, s.incrementDropsCalls, 1)
	assert.Len(t, s.setDropsBonusCalls, 1)
}

func TestDiversityBonusTracker_Awards10thOwnerBonus(t *testing.T) {
	s := newSpyStore()
	module := computers.NewDiversityBonusTracker(s, testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeGrab)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.UserState.ActorDistinctOwnersThisMonth = 9
	pipeCtx.UserState.ActorOwnerAlreadyCountedThisMonth = false
	pipeCtx.UserState.OwnersBonusAlreadyAwarded = false

	acc := pipeline.NewAccumulator()
	acc.Add(pipeline.Award{RecipientUserID: 10, Points: 3, Label: "base_move", ModuleSource: "02_base_move_points"})

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Equal(t, 7.0, mustAwardByLabel(acc, "diversity_owners").Points)
	assert.Len(t, s.incrementOwnersCalls, 1)
	assert.Len(t, s.setOwnersBonusCalls, 1)
}

func TestDiversityBonusTracker_AwardsCountryBonusOncePerMonthPerCountry(t *testing.T) {
	s := newSpyStore()
	module := computers.NewDiversityBonusTracker(s, testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.RuntimeFlags.NewCountryVisited = true
	pipeCtx.RuntimeFlags.NewCountryCode = "DE"
	pipeCtx.UserState.ActorCountriesVisitedThisMonth = map[string]bool{}

	acc := pipeline.NewAccumulator()
	acc.Add(pipeline.Award{RecipientUserID: 10, Points: 3, Label: "base_move", ModuleSource: "02_base_move_points"})

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Equal(t, 5.0, mustAwardByLabel(acc, "diversity_country").Points)
	assert.True(t, pipeCtx.UserState.ActorCountriesVisitedThisMonth["DE"])
	assert.Len(t, s.recordDiversityCountryCalls, 1)

	acc2 := pipeline.NewAccumulator()
	acc2.Add(pipeline.Award{RecipientUserID: 10, Points: 3, Label: "base_move", ModuleSource: "02_base_move_points"})
	err = module.Process(context.Background(), pipeCtx, acc2)
	require.NoError(t, err)
	assert.Len(t, acc2.Awards(), 1)
}

func TestDiversityBonusTracker_SkipsWhenNoBaseMoveAward(t *testing.T) {
	module := computers.NewDiversityBonusTracker(newSpyStore(), testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.RuntimeFlags.NewCountryVisited = true
	pipeCtx.RuntimeFlags.NewCountryCode = "DE"

	acc := pipeline.NewAccumulator()
	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Len(t, acc.Awards(), 0)
}
