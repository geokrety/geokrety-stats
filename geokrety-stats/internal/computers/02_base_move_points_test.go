package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestBaseMovePoints_FirstNonOwnerDropGetsBaseTimesMultiplier(t *testing.T) {
	module := computers.NewBaseMovePoints(newSpyStore(), testCfg())
	evt := testEvent(10, 777, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 99, 1.25)
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	require.Len(t, acc.Awards(), 1)
	assert.Equal(t, 3.75, acc.Awards()[0].Points)
	assert.Equal(t, "base_move", acc.Awards()[0].Label)
	assert.True(t, pipeCtx.RuntimeFlags.ActorScoredThisGK)
	assert.Equal(t, 3.75, pipeCtx.RuntimeFlags.BasePointsAwarded)
}

func TestBaseMovePoints_ZeroForOwnerStandardGK(t *testing.T) {
	module := computers.NewBaseMovePoints(newSpyStore(), testCfg())
	evt := testEvent(42, 777, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 42, 1.8)
	pipeCtx.GKState.GKType = 0
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Len(t, acc.Awards(), 0)
	assert.False(t, pipeCtx.RuntimeFlags.ActorScoredThisGK)
}

func TestBaseMovePoints_ZeroForRepeatedMoveTypeOnSameGK(t *testing.T) {
	module := computers.NewBaseMovePoints(newSpyStore(), testCfg())
	evt := testEvent(10, 777, pipeline.LogTypeGrab)
	pipeCtx := testCtx(evt, 99, 1.2)
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{pipeline.LogTypeGrab: true}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Len(t, acc.Awards(), 0)
}

func TestBaseMovePoints_ZeroForSelfGrab(t *testing.T) {
	module := computers.NewBaseMovePoints(newSpyStore(), testCfg())
	evt := testEvent(10, 777, pipeline.LogTypeGrab)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.GKState.CurrentHolder = 10
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Len(t, acc.Awards(), 0)
}

func TestBaseMovePoints_ZeroForWaypointRequiredWithoutWaypoint(t *testing.T) {
	module := computers.NewBaseMovePoints(newSpyStore(), testCfg())
	evt := testEvent(10, 777, pipeline.LogTypeSeen)
	evt.Waypoint = ""
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Len(t, acc.Awards(), 0)
}

func TestBaseMovePoints_NonTransferableOwnerDipScoresWhenMonthlyLimitNotHit(t *testing.T) {
	s := newSpyStore()
	s.getActorGKsAtLocationThisMonthFn = func(context.Context, int64, string, string) (int, error) { return 0, nil }
	module := computers.NewBaseMovePoints(s, testCfg())
	evt := testEvent(42, 777, pipeline.LogTypeDip)
	pipeCtx := testCtx(evt, 42, 1.1)
	pipeCtx.GKState.GKType = 8
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 1)
	assert.InDelta(t, 3.3, acc.Awards()[0].Points, 0.0001)
}

func TestBaseMovePoints_NonTransferableOwnerMonthlyLimitZerosBase(t *testing.T) {
	s := newSpyStore()
	s.getActorGKsAtLocationThisMonthFn = func(context.Context, int64, string, string) (int, error) { return 1, nil }
	module := computers.NewBaseMovePoints(s, testCfg())
	evt := testEvent(42, 777, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 42, 1.1)
	pipeCtx.GKState.GKType = 8
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Len(t, acc.Awards(), 0)
	assert.Equal(t, "zeroed_non_transferable_limit", pipeCtx.RuntimeFlags.BasePointsLabel)
}
