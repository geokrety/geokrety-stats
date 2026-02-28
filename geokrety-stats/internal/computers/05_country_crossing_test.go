package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestCountryCrossing_NonOwnerNewCountryAwardsActorAndOwner(t *testing.T) {
	s := newSpyStore()
	module := computers.NewCountryCrossing(s, testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeDrop)
	evt.Country = "DE"
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.GKState.HomeCountry = "PL"
	pipeCtx.GKHistory.CountriesVisited = map[string]bool{"PL": true}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 2)

	assert.Equal(t, 3.0, mustAwardByLabel(acc, "country_crossing_actor").Points)
	assert.Equal(t, 1.0, mustAwardByLabel(acc, "country_crossing_owner").Points)
	assert.True(t, pipeCtx.RuntimeFlags.NewCountryVisited)
	assert.Equal(t, "DE", pipeCtx.RuntimeFlags.NewCountryCode)
	assert.Len(t, s.addedCountriesVisited, 1)
}

func TestCountryCrossing_OwnerStandardGetsSelfBonusNoOwnerAward(t *testing.T) {
	module := computers.NewCountryCrossing(newSpyStore(), testCfg())
	evt := testEvent(99, 500, pipeline.LogTypeSeen)
	evt.Country = "DE"
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.GKState.GKType = 0
	pipeCtx.GKState.HomeCountry = "PL"
	pipeCtx.GKHistory.CountriesVisited = map[string]bool{"PL": true}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 1)
	assert.Equal(t, 2.0, mustAwardByLabel(acc, "country_crossing_actor").Points)
}

func TestCountryCrossing_OwnerNonTransferableGetsHigherBonus(t *testing.T) {
	module := computers.NewCountryCrossing(newSpyStore(), testCfg())
	evt := testEvent(99, 500, pipeline.LogTypeDip)
	evt.Country = "DE"
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.GKState.GKType = 8
	pipeCtx.GKState.HomeCountry = "PL"
	pipeCtx.GKHistory.CountriesVisited = map[string]bool{"PL": true}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 1)
	assert.Equal(t, 4.0, mustAwardByLabel(acc, "country_crossing_actor").Points)
}

func TestCountryCrossing_SkipsNonEligibleCases(t *testing.T) {
	module := computers.NewCountryCrossing(newSpyStore(), testCfg())

	t.Run("grab not eligible", func(t *testing.T) {
		evt := testEvent(10, 500, pipeline.LogTypeGrab)
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.GKState.HomeCountry = "PL"
		pipeCtx.GKHistory.CountriesVisited = map[string]bool{"PL": true}
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})

	t.Run("home country", func(t *testing.T) {
		evt := testEvent(10, 500, pipeline.LogTypeDrop)
		evt.Country = "PL"
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.GKState.HomeCountry = "PL"
		pipeCtx.GKHistory.CountriesVisited = map[string]bool{"PL": true}
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})

	t.Run("already visited", func(t *testing.T) {
		evt := testEvent(10, 500, pipeline.LogTypeDrop)
		evt.Country = "DE"
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.GKState.HomeCountry = "PL"
		pipeCtx.GKHistory.CountriesVisited = map[string]bool{"PL": true, "DE": true}
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})
}
