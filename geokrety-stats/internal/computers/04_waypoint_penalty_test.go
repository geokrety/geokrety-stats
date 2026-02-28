package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestWaypointPenalty_TiersApplied10050_25_0(t *testing.T) {
	module := computers.NewWaypointPenalty(newSpyStore(), testCfg())
	tiers := []struct {
		name       string
		priorCount int
		expected   float64
	}{
		{name: "first", priorCount: 0, expected: 3.0},
		{name: "second", priorCount: 1, expected: 1.5},
		{name: "third", priorCount: 2, expected: 0.75},
		{name: "fourth", priorCount: 3, expected: 0.0},
	}

	for _, tc := range tiers {
		t.Run(tc.name, func(t *testing.T) {
			s := newSpyStore()
			module := computers.NewWaypointPenalty(s, testCfg())
			evt := testEvent(10, 500+int64(tc.priorCount), pipeline.LogTypeDrop)
			pipeCtx := testCtx(evt, 99, 1.0)
			pipeCtx.RuntimeFlags.ActorScoredThisGK = true
			pipeCtx.RuntimeFlags.BasePointsAwarded = 3.0
			pipeCtx.UserState.ActorGKsAtLocationThisMonth = tc.priorCount

			acc := pipeline.NewAccumulator()
			acc.Add(pipeline.Award{RecipientUserID: 10, Points: 3.0, Label: "base_move", ModuleSource: "02_base_move_points"})

			err := module.Process(context.Background(), pipeCtx, acc)
			require.NoError(t, err)
			assert.InDelta(t, tc.expected, mustAwardByLabel(acc, "base_move").Points, 0.0001)

			if tc.expected > 0 {
				assert.Len(t, s.recordedActorGKAtLocation, 1)
			} else {
				assert.Len(t, s.recordedActorGKAtLocation, 0)
				assert.False(t, pipeCtx.RuntimeFlags.ActorScoredThisGK)
			}
		})
	}
	_ = module
}

func TestWaypointPenalty_SkipsWhenNoLocationOrNoBasePoints(t *testing.T) {
	module := computers.NewWaypointPenalty(newSpyStore(), testCfg())

	t.Run("no location", func(t *testing.T) {
		evt := testEvent(10, 500, pipeline.LogTypeGrab)
		evt.Waypoint = ""
		evt.Lat = 0
		evt.Lon = 0
		pipeCtx := testCtx(evt, 99, 1.0)
		pipeCtx.RuntimeFlags.ActorScoredThisGK = true
		acc := pipeline.NewAccumulator()
		acc.Add(pipeline.Award{RecipientUserID: 10, Points: 3.0, Label: "base_move", ModuleSource: "02_base_move_points"})
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
