// Package store provides a mock Store implementation for use in tests.
package store

import (
	"context"
	"time"

	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

// MockStore is a configurable stub implementation of Store for tests.
// All methods return zero values unless overridden via the hook fields.
type MockStore struct {
	// Configurable return values (populated per test)
	IsEventProcessedFn     func(ctx context.Context, moveID int64) (bool, error)
	GetMoveFn              func(ctx context.Context, moveID int64) (*GKMoveRow, error)
	GetGKFn                func(ctx context.Context, gkID int64) (*GKRow, error)
	GetGKMultiplierStateFn func(ctx context.Context, gkID int64) (float64, time.Time, int64, *time.Time, error)
	GetActiveChainFn       func(ctx context.Context, gkID int64) (*ChainRow, error)
	GetChainMembersFn      func(ctx context.Context, chainID int64) ([]int64, error)
	GetChainLastBonusFn    func(ctx context.Context, userID, gkID int64) (time.Time, error)
	GetGKDistinctUsers6MFn func(ctx context.Context, gkID int64, before time.Time) ([]int64, error)
	SaveAwardsFn           func(ctx context.Context, awards []pipeline.FinalAward) error
	MarkEventProcessedFn   func(ctx context.Context, moveID int64, result string) error

	// Recorders for assertions
	RecordedAwards       []pipeline.FinalAward
	SaveAwardsCalled     bool
	MarkProcessedCalled  bool
	MarkProcessedMoveID  int64
}

// NewMockStore creates a MockStore with sensible defaults.
// All queries return empty/zero results by default.
func NewMockStore() *MockStore {
	return &MockStore{}
}

// ---- Event idempotency ----

func (m *MockStore) IsEventProcessed(ctx context.Context, moveID int64) (bool, error) {
	if m.IsEventProcessedFn != nil {
		return m.IsEventProcessedFn(ctx, moveID)
	}
	return false, nil
}

func (m *MockStore) MarkEventProcessed(ctx context.Context, moveID int64, result string) error {
	m.MarkProcessedCalled = true
	m.MarkProcessedMoveID = moveID
	if m.MarkEventProcessedFn != nil {
		return m.MarkEventProcessedFn(ctx, moveID, result)
	}
	return nil
}

// ---- GeoKret read ----

func (m *MockStore) GetMove(ctx context.Context, moveID int64) (*GKMoveRow, error) {
	if m.GetMoveFn != nil {
		return m.GetMoveFn(ctx, moveID)
	}
	return nil, nil
}

func (m *MockStore) GetGK(ctx context.Context, gkID int64) (*GKRow, error) {
	if m.GetGKFn != nil {
		return m.GetGKFn(ctx, gkID)
	}
	return nil, nil
}

func (m *MockStore) GetGKPreviousHolder(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

func (m *MockStore) GetGKHomeCountry(_ context.Context, _ int64) (string, error) {
	return "", nil
}

func (m *MockStore) GetGKLastDrop(_ context.Context, _ int64, _ time.Time) (*time.Time, int64, error) {
	return nil, 0, nil
}

func (m *MockStore) GetGKLastCacheEntry(_ context.Context, _ int64, _ time.Time) (*time.Time, error) {
	return nil, nil
}

func (m *MockStore) GetGKDistinctUsers6M(ctx context.Context, gkID int64, before time.Time) ([]int64, error) {
	if m.GetGKDistinctUsers6MFn != nil {
		return m.GetGKDistinctUsers6MFn(ctx, gkID, before)
	}
	return nil, nil
}

// ---- Multiplier state ----

func (m *MockStore) GetGKMultiplierState(ctx context.Context, gkID int64) (float64, time.Time, int64, *time.Time, error) {
	if m.GetGKMultiplierStateFn != nil {
		return m.GetGKMultiplierStateFn(ctx, gkID)
	}
	return 1.0, time.Time{}, 0, nil, nil
}

func (m *MockStore) SaveGKMultiplierState(_ context.Context, _ int64, _ float64, _ time.Time, _ int64, _ *time.Time) error {
	return nil
}

func (m *MockStore) LogGKMultiplierChange(_ context.Context, _, _ int64, _, _ float64, _, _ string) error {
	return nil
}

// ---- Countries visited ----

func (m *MockStore) GetGKCountriesVisited(_ context.Context, _ int64) (map[string]bool, error) {
	return nil, nil
}

func (m *MockStore) AddGKCountryVisited(_ context.Context, _ int64, _ string, _ time.Time, _ int64) error {
	return nil
}

// ---- User move history ----

func (m *MockStore) GetUserMoveHistoryOnGK(_ context.Context, _, _ int64) (map[pipeline.LogType]bool, error) {
	return nil, nil
}

func (m *MockStore) AddUserMoveHistory(_ context.Context, _, _ int64, _ pipeline.LogType, _ time.Time) error {
	return nil
}

// ---- Owner GK limit ----

func (m *MockStore) GetActorGKsPerOwnerCount(_ context.Context, _, _ int64) (int, error) {
	return 0, nil
}

func (m *MockStore) IsGKAlreadyCountedForOwner(_ context.Context, _, _, _ int64) (bool, error) {
	return false, nil
}

func (m *MockStore) RecordGKCountedForOwner(_ context.Context, _, _, _ int64, _ time.Time) error {
	return nil
}

// ---- Waypoint penalty ----

func (m *MockStore) GetActorGKsAtLocationThisMonth(_ context.Context, _ int64, _ string, _ string) (int, error) {
	return 0, nil
}

func (m *MockStore) RecordActorGKAtLocation(_ context.Context, _ int64, _ string, _ string, _ int64, _ time.Time) error {
	return nil
}

// ---- Monthly diversity ----

func (m *MockStore) GetActorMonthlyDiversity(_ context.Context, _ int64, _ string) (int, bool, int, bool, error) {
	return 0, false, 0, false, nil
}

func (m *MockStore) GetActorMonthlyDiversityCountries(_ context.Context, _ int64, _ string) (map[string]bool, error) {
	return nil, nil
}

func (m *MockStore) GetActorMonthlyDiversityOwners(_ context.Context, _ int64, _ string) (map[int64]bool, error) {
	return nil, nil
}

func (m *MockStore) GetActorMonthlyDiversityDrops(_ context.Context, _ int64, _ string) (map[int64]bool, error) {
	return nil, nil
}

func (m *MockStore) IncrementActorMonthlyDrops(_ context.Context, _ int64, _ string, _ int64, _ time.Time) error {
	return nil
}

func (m *MockStore) SetDropsBonusAwarded(_ context.Context, _ int64, _ string) error {
	return nil
}

func (m *MockStore) IncrementActorMonthlyOwners(_ context.Context, _ int64, _ string, _ int64, _ time.Time) error {
	return nil
}

func (m *MockStore) SetOwnersBonusAwarded(_ context.Context, _ int64, _ string) error {
	return nil
}

func (m *MockStore) RecordActorDiversityCountry(_ context.Context, _ int64, _, _ string, _ time.Time) error {
	return nil
}

// ---- Chain state ----

func (m *MockStore) GetActiveChain(ctx context.Context, gkID int64) (*ChainRow, error) {
	if m.GetActiveChainFn != nil {
		return m.GetActiveChainFn(ctx, gkID)
	}
	return nil, nil
}

func (m *MockStore) GetChainMembers(ctx context.Context, chainID int64) ([]int64, error) {
	if m.GetChainMembersFn != nil {
		return m.GetChainMembersFn(ctx, chainID)
	}
	return nil, nil
}

func (m *MockStore) CreateChain(_ context.Context, _ int64, _ time.Time) (int64, error) {
	return 1, nil
}

func (m *MockStore) AddChainMember(_ context.Context, _ int64, _ int64, _ time.Time) error {
	return nil
}

func (m *MockStore) UpdateChainLastActive(_ context.Context, _ int64, _ time.Time, _ *time.Time) error {
	return nil
}

func (m *MockStore) EndChain(_ context.Context, _ int64, _ time.Time, _ string) error {
	return nil
}

func (m *MockStore) GetChainLastBonus(ctx context.Context, userID, gkID int64) (time.Time, error) {
	if m.GetChainLastBonusFn != nil {
		return m.GetChainLastBonusFn(ctx, userID, gkID)
	}
	return time.Time{}, nil
}

func (m *MockStore) RecordChainCompletion(_ context.Context, _, _, _ int64, _ time.Time) error {
	return nil
}

func (m *MockStore) GetExpiredChains(_ context.Context, _ int, _ time.Time) ([]*ChainRow, error) {
	return nil, nil
}

// ---- Point recording ----

func (m *MockStore) SaveAwards(ctx context.Context, awards []pipeline.FinalAward) error {
	m.SaveAwardsCalled = true
	m.RecordedAwards = append(m.RecordedAwards, awards...)
	if m.SaveAwardsFn != nil {
		return m.SaveAwardsFn(ctx, awards)
	}
	return nil
}

// ---- Historical replay ----

func (m *MockStore) GetMoveIDsPage(_ context.Context, _, _ int64, _, _ *time.Time, _ int) ([]int64, error) {
	return nil, nil
}
