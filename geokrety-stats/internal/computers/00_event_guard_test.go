package computers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

func TestEventGuard_HaltsOnAnonymous(t *testing.T) {
	guard := computers.NewEventGuard(mockStore())
	ctx := context.Background()

	evt := testEvent(0, 123, pipeline.LogTypeDrop) // UserID = 0 = anonymous
	pipeCtx := testCtx(evt, 456, 1.0)
	acc := pipeline.NewAccumulator()

	err := guard.Process(ctx, pipeCtx, acc)
	require.Error(t, err)
	assert.True(t, computers.IsHalt(err))
	assert.Contains(t, err.Error(), "anonymous")
}

func TestEventGuard_HaltsOnNonScoreable(t *testing.T) {
	guard := computers.NewEventGuard(mockStore())
	ctx := context.Background()

	evt := testEvent(1, 123, pipeline.LogTypeComment) // COMMENT is not scoreable
	pipeCtx := testCtx(evt, 456, 1.0)
	acc := pipeline.NewAccumulator()

	err := guard.Process(ctx, pipeCtx, acc)
	require.Error(t, err)
	assert.True(t, computers.IsHalt(err))
	assert.Contains(t, err.Error(), "not scoreable")
}

func TestEventGuard_PassesValidEvent(t *testing.T) {
	guard := computers.NewEventGuard(mockStore())
	ctx := context.Background()

	evt := testEvent(1, 123, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 456, 1.0)
	acc := pipeline.NewAccumulator()

	err := guard.Process(ctx, pipeCtx, acc)
	assert.NoError(t, err)
}

func TestEventGuard_AllowsArchivedAndDipAsScoreable(t *testing.T) {
	guard := computers.NewEventGuard(mockStore())
	ctx := context.Background()

	for _, logType := range []pipeline.LogType{pipeline.LogTypeArchived, pipeline.LogTypeDip} {
		evt := testEvent(1, 123, logType)
		pipeCtx := testCtx(evt, 456, 1.0)
		acc := pipeline.NewAccumulator()

		err := guard.Process(ctx, pipeCtx, acc)
		require.NoError(t, err)
	}
}
