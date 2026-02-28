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

func TestRelayBonus_WithinWindowAwardsMoverAndDropper(t *testing.T) {
	module := computers.NewRelayBonus(newSpyStore(), testCfg())
	dropAt := time.Date(2011, 7, 10, 10, 0, 0, 0, time.UTC)

	evt := testEvent(9865, 4443, pipeline.LogTypeGrab)
	evt.LoggedAt = dropAt.Add(6 * 24 * time.Hour)
	pipeCtx := testCtx(evt, 405, 1.0)
	pipeCtx.GKState.CurrentHolder = 0
	pipeCtx.GKHistory.LastDropAt = &dropAt
	pipeCtx.GKHistory.LastDropUser = 9936
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 2)
	assert.Equal(t, 2.0, mustAwardByLabel(acc, "relay_mover").Points)
	assert.Equal(t, 1.0, mustAwardByLabel(acc, "relay_dropper").Points)
}

func TestRelayBonus_SkipsBoundaryAndInvalidCases(t *testing.T) {
	module := computers.NewRelayBonus(newSpyStore(), testCfg())
	dropAt := time.Date(2011, 7, 10, 10, 0, 0, 0, time.UTC)

	t.Run("after 7 days", func(t *testing.T) {
		evt := testEvent(1, 2, pipeline.LogTypeGrab)
		evt.LoggedAt = dropAt.Add(169 * time.Hour)
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.GKState.CurrentHolder = 0
		pipeCtx.GKHistory.LastDropAt = &dropAt
		pipeCtx.GKHistory.LastDropUser = 55
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})

	t.Run("same as dropper", func(t *testing.T) {
		evt := testEvent(55, 2, pipeline.LogTypeGrab)
		evt.LoggedAt = dropAt.Add(2 * 24 * time.Hour)
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.GKState.CurrentHolder = 0
		pipeCtx.GKHistory.LastDropAt = &dropAt
		pipeCtx.GKHistory.LastDropUser = 55
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})

	t.Run("not in cache", func(t *testing.T) {
		evt := testEvent(1, 2, pipeline.LogTypeGrab)
		evt.LoggedAt = dropAt.Add(2 * 24 * time.Hour)
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.GKState.CurrentHolder = 777
		pipeCtx.GKHistory.LastDropAt = &dropAt
		pipeCtx.GKHistory.LastDropUser = 55
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})
}
