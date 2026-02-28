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

func TestChainStateManager_CreatesChainAndAddsMemberOnFirstDrop(t *testing.T) {
	s := newSpyStore()
	s.nextChainID = 12345
	module := computers.NewChainStateManager(s, testCfg())

	evt := testEvent(101, 500, pipeline.LogTypeDrop)
	pipeCtx := testCtx(evt, 999, 1.0)
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	assert.Equal(t, int64(12345), pipeCtx.ChainState.ActiveChainID)
	assert.Equal(t, []int64{101}, pipeCtx.ChainState.ChainMembers)
	require.Len(t, s.createdChains, 1)
	require.Len(t, s.addedMembers, 1)
	assert.Equal(t, int64(101), s.addedMembers[0].userID)
}

func TestChainStateManager_DIPExtendsTimerWithoutAddingMember(t *testing.T) {
	s := newSpyStore()
	module := computers.NewChainStateManager(s, testCfg())
	now := time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC)

	evt := testEvent(101, 500, pipeline.LogTypeDip)
	evt.LoggedAt = now
	pipeCtx := testCtx(evt, 999, 1.0)
	pipeCtx.ChainState.ActiveChainID = 77
	pipeCtx.ChainState.ChainMembers = []int64{101, 202}
	pipeCtx.ChainState.ChainLastActive = now.Add(-24 * time.Hour)
	pipeCtx.ChainState.HolderAcquiredAt = now.Add(-5 * 24 * time.Hour)
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.Equal(t, []int64{101, 202}, pipeCtx.ChainState.ChainMembers)
	assert.Len(t, s.addedMembers, 0)
	assert.Len(t, s.updatedChainLastActive, 1)
}

func TestChainStateManager_ArchiveEndsActiveChain(t *testing.T) {
	s := newSpyStore()
	s.GetChainMembersFn = func(context.Context, int64) ([]int64, error) { return []int64{1, 2, 3}, nil }
	module := computers.NewChainStateManager(s, testCfg())

	evt := testEvent(101, 500, pipeline.LogTypeArchived)
	pipeCtx := testCtx(evt, 999, 1.0)
	pipeCtx.ChainState.ActiveChainID = 77
	pipeCtx.ChainState.ChainLastActive = evt.LoggedAt
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	assert.True(t, pipeCtx.ChainState.ChainEnded)
	assert.Equal(t, int64(77), pipeCtx.ChainState.EndedChainID)
	assert.Equal(t, []int64{1, 2, 3}, pipeCtx.ChainState.EndedChainMembers)
	require.Len(t, s.endedChains, 1)
	assert.Equal(t, "archived", s.endedChains[0].reason)
}

func TestChainStateManager_ExpiredChainFinalizesThenStartsNewChain(t *testing.T) {
	s := newSpyStore()
	s.nextChainID = 900
	s.GetChainMembersFn = func(context.Context, int64) ([]int64, error) { return []int64{10, 20, 30}, nil }
	module := computers.NewChainStateManager(s, testCfg())
	now := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)

	evt := testEvent(404, 500, pipeline.LogTypeGrab)
	evt.LoggedAt = now
	pipeCtx := testCtx(evt, 999, 1.0)
	pipeCtx.ChainState.ActiveChainID = 88
	pipeCtx.ChainState.ChainMembers = []int64{10, 20}
	pipeCtx.ChainState.ChainLastActive = now.Add(-15 * 24 * time.Hour)
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	assert.True(t, pipeCtx.ChainState.ChainEnded)
	assert.Equal(t, int64(88), pipeCtx.ChainState.EndedChainID)
	assert.Equal(t, []int64{10, 20, 30}, pipeCtx.ChainState.EndedChainMembers)
	assert.Equal(t, int64(900), pipeCtx.ChainState.ActiveChainID)
	assert.Contains(t, pipeCtx.ChainState.ChainMembers, int64(404))
	require.Len(t, s.endedChains, 1)
	assert.Equal(t, "timeout", s.endedChains[0].reason)
}
