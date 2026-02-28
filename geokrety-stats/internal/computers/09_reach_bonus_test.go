package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestReachBonus_AwardsOwnerOn10thDistinctUser(t *testing.T) {
	module := computers.NewReachBonus(newSpyStore(), testCfg())
	evt := testEvent(10, 500, pipeline.LogTypeGrab)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.RuntimeFlags.ActorScoredThisGK = true
	pipeCtx.GKHistory.DistinctUsers6M = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	require.Len(t, acc.Awards(), 1)
	assert.Equal(t, 5.0, mustAwardByLabel(acc, "reach_owner").Points)
}

func TestReachBonus_SkipsIfAlreadyCountedOrNotScoredOrOwner(t *testing.T) {
	module := computers.NewReachBonus(newSpyStore(), testCfg())
	cases := []struct {
		name      string
		setupFunc func(*pipeline.Context)
	}{
		{name: "actor not scored", setupFunc: func(p *pipeline.Context) { p.RuntimeFlags.ActorScoredThisGK = false }},
		{name: "actor is owner", setupFunc: func(p *pipeline.Context) { p.RuntimeFlags.ActorScoredThisGK = true; p.Event.UserID = p.GKState.OwnerID }},
		{name: "actor already in list", setupFunc: func(p *pipeline.Context) { p.RuntimeFlags.ActorScoredThisGK = true; p.GKHistory.DistinctUsers6M = []int64{1, 2, p.Event.UserID} }},
		{name: "milestone not reached", setupFunc: func(p *pipeline.Context) { p.RuntimeFlags.ActorScoredThisGK = true; p.GKHistory.DistinctUsers6M = []int64{1, 2, 3} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evt := testEvent(10, 500, pipeline.LogTypeDrop)
			pipeCtx := testCtx(evt, 99, 1.0)
			tc.setupFunc(pipeCtx)
			acc := pipeline.NewAccumulator()

			require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
			assert.Len(t, acc.Awards(), 0)
		})
	}
}
