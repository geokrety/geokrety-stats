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

func TestChainBonus_SkipsWhenChainNotEndedOrTooShort(t *testing.T) {
	module := computers.NewChainBonus(newSpyStore(), testCfg())

	t.Run("not ended", func(t *testing.T) {
		pipeCtx := testCtx(testEvent(1, 500, pipeline.LogTypeDrop), 99, 1.0)
		pipeCtx.ChainState.ChainEnded = false
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})

	t.Run("too short", func(t *testing.T) {
		pipeCtx := testCtx(testEvent(1, 500, pipeline.LogTypeDrop), 99, 1.0)
		pipeCtx.ChainState.ChainEnded = true
		pipeCtx.ChainState.EndedChainID = 10
		pipeCtx.ChainState.EndedChainMembers = []int64{1, 2}
		acc := pipeline.NewAccumulator()
		require.NoError(t, module.Process(context.Background(), pipeCtx, acc))
		assert.Len(t, acc.Awards(), 0)
	})
}

func TestChainBonus_AwardsParticipantsAndOwnerShare(t *testing.T) {
	s := newSpyStore()
	module := computers.NewChainBonus(s, testCfg())

	evt := testEvent(1, 500, pipeline.LogTypeDrop)
	evt.LoggedAt = time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.ChainState.ChainEnded = true
	pipeCtx.ChainState.EndedChainID = 10
	pipeCtx.ChainState.EndedChainMembers = []int64{1, 2, 3}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)
	awards := acc.Awards()
	require.Len(t, awards, 4)

	var chainCount int
	var ownerShare float64
	for _, award := range awards {
		if award.Label == "chain_bonus" {
			chainCount++
			assert.Equal(t, 9.0, award.Points)
		}
		if award.Label == "chain_bonus_owner" {
			ownerShare = award.Points
		}
	}
	assert.Equal(t, 3, chainCount)
	assert.Equal(t, 6.75, ownerShare)
	assert.Len(t, s.recordChainCompletionCalls, 3)
}

func TestChainBonus_RespectsCooldownPerUser(t *testing.T) {
	s := newSpyStore()
	recent := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.GetChainLastBonusFn = func(_ context.Context, userID, gkID int64) (time.Time, error) {
		if userID == 2 {
			return recent, nil
		}
		return time.Time{}, nil
	}
	module := computers.NewChainBonus(s, testCfg())
	evt := testEvent(1, 500, pipeline.LogTypeDrop)
	evt.LoggedAt = time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	pipeCtx := testCtx(evt, 99, 1.0)
	pipeCtx.ChainState.ChainEnded = true
	pipeCtx.ChainState.EndedChainID = 10
	pipeCtx.ChainState.EndedChainMembers = []int64{1, 2, 3}
	acc := pipeline.NewAccumulator()

	err := module.Process(context.Background(), pipeCtx, acc)
	require.NoError(t, err)

	var participantAwards int
	var ownerShare float64
	for _, award := range acc.Awards() {
		if award.Label == "chain_bonus" {
			participantAwards++
		}
		if award.Label == "chain_bonus_owner" {
			ownerShare = award.Points
		}
	}
	assert.Equal(t, 2, participantAwards)
	assert.Equal(t, 4.5, ownerShare)
}
