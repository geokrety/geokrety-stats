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

func TestGKMultiplierUpdater_FirstMoveAndCountryIncrease(t *testing.T) {
	s := newSpyStore()
	module := computers.NewGKMultiplierUpdater(s, testCfg())
	now := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	evt.LoggedAt = now

	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.GKState.LastMultiplierAt = now
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	pipeCtx.RuntimeFlags.NewCountryVisited = true
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	assert.Equal(t, 1.06, pipeCtx.GKState.CurrentMultiplier)
	require.Len(t, s.saveMultiplierCalls, 1)
	assert.Equal(t, 1.06, s.saveMultiplierCalls[0].multiplier)
	assert.Equal(t, 1, s.logMultiplierChangeCalls)
	assert.Len(t, s.addUserMoveHistoryCalls, 1)
}

func TestGKMultiplierUpdater_AppliesInHandDecayAndFloor(t *testing.T) {
	s := newSpyStore()
	module := computers.NewGKMultiplierUpdater(s, testCfg())
	last := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	now := last.Add(62 * 24 * time.Hour)
	evt := testEvent(10, 500, pipeline.LogTypeGrab)
	evt.LoggedAt = now

	pipeCtx := testCtx(evt, 99, 1.2)
	pipeCtx.GKState.CurrentHolder = 10
	pipeCtx.GKState.LastMultiplierAt = last
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{pipeline.LogTypeGrab: true}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Equal(t, 1.0, pipeCtx.GKState.CurrentMultiplier)
}

func TestGKMultiplierUpdater_AppliesInCacheDecayAndCeiling(t *testing.T) {
	s := newSpyStore()
	module := computers.NewGKMultiplierUpdater(s, testCfg())
	last := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	now := last.Add(14 * 24 * time.Hour)
	evt := testEvent(10, 500, pipeline.LogTypeSeen)
	evt.LoggedAt = now

	pipeCtx := testCtx(evt, 99, 1.99)
	pipeCtx.GKState.CurrentHolder = 0
	pipeCtx.GKState.LastMultiplierAt = last
	pipeCtx.UserState.ActorMoveHistoryOnGK = map[pipeline.LogType]bool{}
	pipeCtx.RuntimeFlags.NewCountryVisited = true
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	assert.InDelta(t, 2.0, pipeCtx.GKState.CurrentMultiplier, 0.0001)
}
