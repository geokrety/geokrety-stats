package computers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/geokrety/geokrety-points-system/internal/computers"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

func TestContextLoader_LoadsAllMajorState(t *testing.T) {
	ms := newSpyStore()
	ownerID := int64(999)
	holderID := int64(333)
	createdAt := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	lastMultAt := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	holderAcquiredAt := time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC)
	lastDropAt := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	lastCacheAt := time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC)

	ms.GetGKFn = func(ctx context.Context, gkID int64) (*store.GKRow, error) {
		return &store.GKRow{ID: gkID, Type: 0, Owner: &ownerID, Holder: &holderID, CreatedAt: createdAt}, nil
	}
	ms.GetGKMultiplierStateFn = func(ctx context.Context, gkID int64) (float64, time.Time, int64, *time.Time, error) {
		return 1.37, lastMultAt, holderID, &holderAcquiredAt, nil
	}
	ms.GetGKDistinctUsers6MFn = func(ctx context.Context, gkID int64, before time.Time) ([]int64, error) {
		return []int64{11, 22, 33}, nil
	}
	ms.GetActiveChainFn = func(ctx context.Context, gkID int64) (*store.ChainRow, error) {
		return &store.ChainRow{ID: 888, GKID: gkID, ChainLastActive: lastDropAt, HolderAcquiredAt: &holderAcquiredAt}, nil
	}
	ms.GetChainMembersFn = func(ctx context.Context, chainID int64) ([]int64, error) {
		return []int64{11, 22}, nil
	}

	ms.getGKHomeCountryFn = func(context.Context, int64) (string, error) { return "PL", nil }
	ms.getGKCountriesVisitedFn = func(context.Context, int64) (map[string]bool, error) {
		return map[string]bool{"PL": true, "DE": true}, nil
	}
	ms.getGKLastDropFn = func(context.Context, int64, time.Time) (*time.Time, int64, error) {
		return &lastDropAt, 123, nil
	}
	ms.getGKLastCacheEntryFn = func(context.Context, int64, time.Time) (*time.Time, error) {
		return &lastCacheAt, nil
	}
	ms.getUserMoveHistoryOnGKFn = func(context.Context, int64, int64) (map[pipeline.LogType]bool, error) {
		return map[pipeline.LogType]bool{pipeline.LogTypeDrop: true}, nil
	}
	ms.getActorGKsPerOwnerCountFn = func(context.Context, int64, int64) (int, error) { return 7, nil }
	ms.isGKAlreadyCountedForOwnerFn = func(context.Context, int64, int64, int64) (bool, error) { return true, nil }
	ms.getActorGKsAtLocationThisMonthFn = func(context.Context, int64, string, string) (int, error) { return 2, nil }
	ms.getActorMonthlyDiversityFn = func(context.Context, int64, string) (int, bool, int, bool, error) {
		return 4, false, 9, false, nil
	}
	ms.getActorMonthlyDiversityCountriesFn = func(context.Context, int64, string) (map[string]bool, error) {
		return map[string]bool{"DE": true}, nil
	}
	ms.getActorMonthlyDiversityOwnersFn = func(context.Context, int64, string) (map[int64]bool, error) {
		return map[int64]bool{ownerID: true}, nil
	}
	ms.getActorMonthlyDiversityDropsFn = func(context.Context, int64, string) (map[int64]bool, error) {
		return map[int64]bool{1234: true}, nil
	}

	loader := computers.NewContextLoader(ms, testCfg())
	evt := testEvent(44, 1234, pipeline.LogTypeDrop)
	pipeCtx := &pipeline.Context{Event: evt}

	err := loader.Process(context.Background(), pipeCtx, pipeline.NewAccumulator())
	require.NoError(t, err)

	assert.Equal(t, int64(1234), pipeCtx.GKState.GKID)
	assert.Equal(t, ownerID, pipeCtx.GKState.OwnerID)
	assert.Equal(t, 1.37, pipeCtx.GKState.CurrentMultiplier)
	assert.Equal(t, "PL", pipeCtx.GKState.HomeCountry)
	assert.Equal(t, int64(123), pipeCtx.GKHistory.LastDropUser)
	assert.Equal(t, []int64{11, 22, 33}, pipeCtx.GKHistory.DistinctUsers6M)
	assert.True(t, pipeCtx.GKHistory.CountriesVisited["DE"])

	assert.Equal(t, 7, pipeCtx.UserState.ActorGKsPerOwnerCount)
	assert.True(t, pipeCtx.UserState.ActorGKAlreadyCountedForOwner)
	assert.Equal(t, 2, pipeCtx.UserState.ActorGKsAtLocationThisMonth)
	assert.Equal(t, 4, pipeCtx.UserState.ActorGKsDroppedThisMonth)
	assert.Equal(t, 9, pipeCtx.UserState.ActorDistinctOwnersThisMonth)

	assert.Equal(t, int64(888), pipeCtx.ChainState.ActiveChainID)
	assert.Equal(t, []int64{11, 22}, pipeCtx.ChainState.ChainMembers)
}

func TestContextLoader_ReturnsErrorWhenGKMissing(t *testing.T) {
	ms := newSpyStore()
	ms.GetGKFn = func(context.Context, int64) (*store.GKRow, error) { return nil, nil }

	loader := computers.NewContextLoader(ms, testCfg())
	pipeCtx := &pipeline.Context{Event: testEvent(1, 555, pipeline.LogTypeGrab)}
	err := loader.Process(context.Background(), pipeCtx, pipeline.NewAccumulator())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GK 555 not found")
}

func TestContextLoader_BubblesStoreErrors(t *testing.T) {
	ms := newSpyStore()
	ms.GetGKFn = func(context.Context, int64) (*store.GKRow, error) {
		return nil, errors.New("db down")
	}

	loader := computers.NewContextLoader(ms, testCfg())
	pipeCtx := &pipeline.Context{Event: testEvent(1, 42, pipeline.LogTypeDrop)}
	err := loader.Process(context.Background(), pipeCtx, pipeline.NewAccumulator())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading GK 42")
}
