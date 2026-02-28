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

func TestRescuerBonus_AwardsAfterSixMonthsDormancy(t *testing.T) {
	module := computers.NewRescuerBonus(newSpyStore(), testCfg())
	cacheAt := time.Date(2011, 1, 10, 10, 0, 0, 0, time.UTC)
	evt := testEvent(1330, 4126, pipeline.LogTypeGrab)
	evt.LoggedAt = cacheAt.AddDate(0, 6, 0)

	pipeCtx := testCtx(evt, 405, 1.0)
	pipeCtx.GKState.CurrentHolder = 0
	pipeCtx.GKHistory.LastCacheEntryAt = &cacheAt
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 2)
	assert.Equal(t, 2.0, mustAwardByLabel(acc, "rescuer_grabber").Points)
	assert.Equal(t, 1.0, mustAwardByLabel(acc, "rescuer_owner").Points)
}

func TestRescuerBonus_SkipsBeforeThresholdAndOwnerOrNotCache(t *testing.T) {
	module := computers.NewRescuerBonus(newSpyStore(), testCfg())
	cacheAt := time.Date(2011, 1, 10, 10, 0, 0, 0, time.UTC)

	t.Run("5 months 29 days", func(t *testing.T) {
		evt := testEvent(1330, 4126, pipeline.LogTypeGrab)
		evt.LoggedAt = cacheAt.AddDate(0, 6, -1)
		pipeCtx := testCtx(evt, 405, 1.0)
		pipeCtx.GKState.CurrentHolder = 0
		pipeCtx.GKHistory.LastCacheEntryAt = &cacheAt
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})

	t.Run("owner grab", func(t *testing.T) {
		evt := testEvent(405, 4126, pipeline.LogTypeGrab)
		evt.LoggedAt = cacheAt.AddDate(0, 7, 0)
		pipeCtx := testCtx(evt, 405, 1.0)
		pipeCtx.GKState.CurrentHolder = 0
		pipeCtx.GKHistory.LastCacheEntryAt = &cacheAt
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})

	t.Run("not in cache", func(t *testing.T) {
		evt := testEvent(1330, 4126, pipeline.LogTypeGrab)
		evt.LoggedAt = cacheAt.AddDate(0, 7, 0)
		pipeCtx := testCtx(evt, 405, 1.0)
		pipeCtx.GKState.CurrentHolder = 999
		pipeCtx.GKHistory.LastCacheEntryAt = &cacheAt
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})
}
