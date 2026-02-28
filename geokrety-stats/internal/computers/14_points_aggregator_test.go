package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestPointsAggregator_FiltersZeroes(t *testing.T) {
	agg := computers.NewPointsAggregator(mockStore(), testCfg())
	ctx := context.Background()

	evt := testEvent(1, 123, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 456, 1.0)
	acc := pipeline.NewAccumulator()

	// Add a mix of zero and non-zero awards
	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 0, Label: "zero", ModuleSource: "test"})
	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 3.5, Label: "base", ModuleSource: "test"})
	acc.Add(pipeline.Award{RecipientUserID: 2, Points: 0, Label: "zero", ModuleSource: "test"})

	err := agg.Process(ctx, pipeCtx, acc)
	require.NoError(t, err)

	// Should have exactly one aggregated award
	require.Len(t, pipeCtx.AggregatedAwards, 1)
	assert.Equal(t, int64(1), pipeCtx.AggregatedAwards[0].RecipientUserID)
	assert.Equal(t, 3.5, pipeCtx.AggregatedAwards[0].TotalPoints)
}

func TestPointsAggregator_MergesSameRecipient(t *testing.T) {
	agg := computers.NewPointsAggregator(mockStore(), testCfg())
	ctx := context.Background()

	evt := testEvent(1, 123, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 456, 1.0)
	acc := pipeline.NewAccumulator()

	// Add multiple awards from same recipient with different labels
	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 3.0, Label: "base", ModuleSource: "computer_02"})
	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 2.0, Label: "relay", ModuleSource: "computer_06"})
	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 1.0, Label: "relay", ModuleSource: "computer_06"}) // duplicate label

	err := agg.Process(ctx, pipeCtx, acc)
	require.NoError(t, err)

	// Should produce 2 FinalAwards (one per label)
	require.Len(t, pipeCtx.AggregatedAwards, 2)

	// Find base and relay awards
	var baseFinal, relayFinal pipeline.FinalAward
	for _, fa := range pipeCtx.AggregatedAwards {
		if len(fa.Awards) > 0 {
			if fa.Awards[0].Label == "base" {
				baseFinal = fa
			} else if fa.Awards[0].Label == "relay" {
				relayFinal = fa
			}
		}
	}

	// Base: single award of 3.0
	assert.Equal(t, 3.0, baseFinal.TotalPoints)
	require.Len(t, baseFinal.Awards, 1)
	assert.Equal(t, "base", baseFinal.Awards[0].Label)

	// Relay: two awards merged (2.0 + 1.0 = 3.0)
	assert.Equal(t, 3.0, relayFinal.TotalPoints)
	require.Len(t, relayFinal.Awards, 2) // Both relay awards included
	assert.Equal(t, "relay", relayFinal.Awards[0].Label)
	assert.Equal(t, 2.0, relayFinal.Awards[0].Points)
	assert.Equal(t, "relay", relayFinal.Awards[1].Label)
	assert.Equal(t, 1.0, relayFinal.Awards[1].Points)
}

func TestPointsAggregator_RoundsCorrectly(t *testing.T) {
	agg := computers.NewPointsAggregator(mockStore(), testCfg())
	ctx := context.Background()

	evt := testEvent(1, 123, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 456, 1.0)
	acc := pipeline.NewAccumulator()

	// Add awards that require rounding
	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 2.225, Label: "test", ModuleSource: "test"}) // rounds to 2.23 (half-up)
	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 2.224, Label: "test", ModuleSource: "test"}) // rounds to 2.22 (down)

	err := agg.Process(ctx, pipeCtx, acc)
	require.NoError(t, err)

	require.Len(t, pipeCtx.AggregatedAwards, 1)
	// 2.225 + 2.224 = 4.449 → rounded to 4.45 (half-up)
	assert.Equal(t, 4.45, pipeCtx.AggregatedAwards[0].TotalPoints)
}

func TestPointsAggregator_MarksProcessed(t *testing.T) {
	ms := mockStore()
	agg := computers.NewPointsAggregator(ms, testCfg())
	ctx := context.Background()

	evt := testEvent(1, 123, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 456, 1.0)
	acc := pipeline.NewAccumulator()

	acc.Add(pipeline.Award{RecipientUserID: 1, Points: 3.0, Label: "base", ModuleSource: "test"})

	err := agg.Process(ctx, pipeCtx, acc)
	require.NoError(t, err)

	// Should have marked event as processed
	assert.True(t, ms.MarkProcessedCalled)
	assert.Equal(t, int64(1000), ms.MarkProcessedMoveID) // LogID from testEvent
}
