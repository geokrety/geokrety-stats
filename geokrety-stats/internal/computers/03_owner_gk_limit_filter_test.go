package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestOwnerGKLimitFilter_ZeroesBaseMoveWhenLimitReached(t *testing.T) {
	module := computers.NewOwnerGKLimitFilter(newSpyStore(), testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.RuntimeFlags.ActorScoredThisGK = true
	pipeCtx.RuntimeFlags.BasePointsAwarded = 3
	pipeCtx.UserState.ActorGKsPerOwnerCount = testCfg().MaxGKsPerOwner
	pipeCtx.UserState.ActorGKAlreadyCountedForOwner = false

	acc := pipeline.NewAccumulator()
	acc.Add(pipeline.Award{RecipientUserID: 10, Points: 3, Label: "base_move", ModuleSource: "02_base_move_points"})

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	award := mustAwardByLabel(acc, "base_move")
	assert.Equal(t, 0.0, award.Points)
	assert.False(t, pipeCtx.RuntimeFlags.ActorScoredThisGK)
	assert.Equal(t, 0.0, pipeCtx.RuntimeFlags.BasePointsAwarded)
}

func TestOwnerGKLimitFilter_RecordsGKWhenUnderLimitAndNotCounted(t *testing.T) {
	s := newSpyStore()
	module := computers.NewOwnerGKLimitFilter(s, testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.RuntimeFlags.ActorScoredThisGK = true
	pipeCtx.UserState.ActorGKsPerOwnerCount = 4
	pipeCtx.UserState.ActorGKAlreadyCountedForOwner = false

	acc := pipeline.NewAccumulator()
	acc.Add(pipeline.Award{RecipientUserID: 10, Points: 3, Label: "base_move", ModuleSource: "02_base_move_points"})

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	assert.Len(t, s.recordedGKCountedForOwner, 1)
	assert.Equal(t, int64(10), s.recordedGKCountedForOwner[0].actorID)
	assert.Equal(t, int64(99), s.recordedGKCountedForOwner[0].ownerID)
}

func TestOwnerGKLimitFilter_SkipsOwnerAndNonScoredEvents(t *testing.T) {
	module := computers.NewOwnerGKLimitFilter(newSpyStore(), testCfg())

	t.Run("owner move", func(t *testing.T) {
		evt := testEvent(99, 500, pipeline.LogTypeDrop)
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.RuntimeFlags.ActorScoredThisGK = true
		acc := pipeline.NewAccumulator()
		acc.Add(pipeline.Award{RecipientUserID: 99, Points: 3, Label: "base_move", ModuleSource: "02_base_move_points"})
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Equal(t, 3.0, mustAwardByLabel(acc, "base_move").Points)
	})

	t.Run("no base points", func(t *testing.T) {
		evt := testEvent(10, 500, pipeline.LogTypeDrop)
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.RuntimeFlags.ActorScoredThisGK = false
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})
}
