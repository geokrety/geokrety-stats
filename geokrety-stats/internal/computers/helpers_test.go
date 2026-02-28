package computers_test

import (
	"context"
	"fmt"
	"time"

	"github.com/geokrety/geokrety-points-system/internal/config"
	"github.com/geokrety/geokrety-points-system/internal/pipeline"
	"github.com/geokrety/geokrety-points-system/internal/store"
)

// testCfg returns a StatsConfig with testing-friendly defaults.
func testCfg() config.StatsConfig {
	return config.StatsConfig{
		BaseMovePoints:              3.0,
		MaxGKsPerOwner:              10,
		WaypointPenaltyTiers:        []float64{1.0, 0.5, 0.25, 0},
		CountryCrossingActorBonus:   3.0,
		CountryCrossingOwnerBonus:   1.0,
		RelayMoverBonus:             2.0,
		RelayDropperBonus:           1.0,
		RelayWindowHours:            168,
		RescuerGrabberBonus:         2.0,
		RescuerOwnerBonus:           1.0,
		RescuerDormancyMonths:       6,
		HandoverOwnerBonus:          1.0,
		ReachOwnerBonus:             5.0,
		ReachMilestoneUsers:         10,
		ReachWindowMonths:           6,
		ChainTimeoutDays:            14,
		ChainMinLength:              3,
		ChainAntiFarmingMonths:      6,
		ChainOwnerShareFraction:     0.25,
		DiversityDropsBonus:         3.0,
		DiversityDropsMilestone:     5,
		DiversityOwnersBonus:        7.0,
		DiversityOwnersMilestone:    10,
		DiversityCountryBonus:       5.0,
		MultiplierMin:               1.0,
		MultiplierMax:               2.0,
		MultiplierFirstMoveInc:      0.01,
		MultiplierCountryInc:        0.05,
		MultiplierInHandDecayPerDay: 0.008,
		MultiplierInCacheDecayPerWeek: 0.02,
		CountryCrossingOwnerSelfBonus: 2.0,
		CountryCrossingNonTransferableBonus: 4.0,
		NonTransferableGKTypes:      []int{8, 9, 10},
		FirstFinderWindowHours:      168,
	}
}

// testEvent creates a base DROP event for testing.
func testEvent(userID, gkID int64, logType pipeline.LogType) pipeline.Event {
	return pipeline.Event{
		LogID:    1000,
		UserID:   userID,
		GKID:     gkID,
		LogType:  logType,
		Waypoint: "GC12345",
		Country:  "PL",
		Lat:      50.0,
		Lon:      20.0,
		LoggedAt: time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC),
	}
}

// testCtx returns a basic pipeline context with sensible defaults.
func testCtx(event pipeline.Event, ownerID int64, multiplier float64) *pipeline.Context {
	return &pipeline.Context{
		Event: event,
		GKState: pipeline.GKState{
			GKID:              event.GKID,
			GKType:            0, // standard transferable
			OwnerID:           ownerID,
			CurrentMultiplier: multiplier,
			LastMultiplierAt:  time.Time{},
			CurrentHolder:     0, // in cache
		},
		GKHistory: pipeline.GKHistory{
			CountriesVisited: map[string]bool{},
		},
		UserState:  pipeline.UserState{},
		ChainState: pipeline.ChainState{},
	}
}

// mockStore returns a default mock store instance.
func mockStore() *store.MockStore {
	return store.NewMockStore()
}

func awardByLabel(acc *pipeline.Accumulator, label string) (pipeline.Award, bool) {
	for _, award := range acc.Awards() {
		if award.Label == label {
			return award, true
		}
	}
	return pipeline.Award{}, false
}

func awardsForRecipient(acc *pipeline.Accumulator, userID int64) []pipeline.Award {
	result := make([]pipeline.Award, 0)
	for _, award := range acc.Awards() {
		if award.RecipientUserID == userID {
			result = append(result, award)
		}
	}
	return result
}

type chainEndCall struct {
	chainID int64
	at      time.Time
	reason  string
}

type multiplierSaveCall struct {
	gkID             int64
	multiplier       float64
	at               time.Time
	holderID         int64
	holderAcquiredAt *time.Time
}

type spyStore struct {
	*store.MockStore

	getGKHomeCountryFn                func(ctx context.Context, gkID int64) (string, error)
	getGKCountriesVisitedFn           func(ctx context.Context, gkID int64) (map[string]bool, error)
	getGKLastDropFn                   func(ctx context.Context, gkID int64, before time.Time) (*time.Time, int64, error)
	getGKLastCacheEntryFn             func(ctx context.Context, gkID int64, before time.Time) (*time.Time, error)
	getUserMoveHistoryOnGKFn          func(ctx context.Context, userID, gkID int64) (map[pipeline.LogType]bool, error)
	getActorGKsPerOwnerCountFn        func(ctx context.Context, actorID, ownerID int64) (int, error)
	isGKAlreadyCountedForOwnerFn      func(ctx context.Context, actorID, ownerID, gkID int64) (bool, error)
	getActorGKsAtLocationThisMonthFn  func(ctx context.Context, actorID int64, locationID string, yearMonth string) (int, error)
	getActorMonthlyDiversityFn        func(ctx context.Context, actorID int64, yearMonth string) (int, bool, int, bool, error)
	getActorMonthlyDiversityCountriesFn func(ctx context.Context, actorID int64, yearMonth string) (map[string]bool, error)
	getActorMonthlyDiversityOwnersFn  func(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error)
	getActorMonthlyDiversityDropsFn   func(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error)

	nextChainID int64

	createdChains []int64
	addedMembers  []struct {
		chainID int64
		userID  int64
	}
	updatedChainLastActive []struct {
		chainID          int64
		at               time.Time
		holderAcquiredAt *time.Time
	}
	endedChains []chainEndCall

	recordedGKCountedForOwner []struct {
		actorID int64
		ownerID int64
		gkID    int64
		at      time.Time
	}
	recordedActorGKAtLocation []struct {
		actorID    int64
		locationID string
		yearMonth  string
		gkID       int64
		at         time.Time
	}
	addedCountriesVisited []struct {
		gkID    int64
		country string
		at      time.Time
		moveID  int64
	}

	incrementDropsCalls []struct {
		actorID   int64
		yearMonth string
		gkID      int64
		at        time.Time
	}
	setDropsBonusCalls []struct {
		actorID   int64
		yearMonth string
	}
	incrementOwnersCalls []struct {
		actorID   int64
		yearMonth string
		ownerID   int64
		at        time.Time
	}
	setOwnersBonusCalls []struct {
		actorID   int64
		yearMonth string
	}
	recordDiversityCountryCalls []struct {
		actorID   int64
		yearMonth string
		country   string
		at        time.Time
	}

	saveMultiplierCalls      []multiplierSaveCall
	logMultiplierChangeCalls int
	addUserMoveHistoryCalls  []struct {
		userID  int64
		gkID    int64
		logType pipeline.LogType
		at      time.Time
	}

	recordChainCompletionCalls []struct {
		userID  int64
		gkID    int64
		chainID int64
		at      time.Time
	}
}

func newSpyStore() *spyStore {
	return &spyStore{
		MockStore:   store.NewMockStore(),
		nextChainID: 777,
	}
}

func (s *spyStore) CreateChain(_ context.Context, _ int64, _ time.Time) (int64, error) {
	chainID := s.nextChainID
	s.createdChains = append(s.createdChains, chainID)
	if s.nextChainID == 0 {
		chainID = 1
	}
	return chainID, nil
}

func (s *spyStore) AddChainMember(_ context.Context, chainID int64, userID int64, _ time.Time) error {
	s.addedMembers = append(s.addedMembers, struct {
		chainID int64
		userID  int64
	}{chainID: chainID, userID: userID})
	return nil
}

func (s *spyStore) UpdateChainLastActive(_ context.Context, chainID int64, at time.Time, holderAcquiredAt *time.Time) error {
	s.updatedChainLastActive = append(s.updatedChainLastActive, struct {
		chainID          int64
		at               time.Time
		holderAcquiredAt *time.Time
	}{chainID: chainID, at: at, holderAcquiredAt: holderAcquiredAt})
	return nil
}

func (s *spyStore) EndChain(_ context.Context, chainID int64, at time.Time, reason string) error {
	s.endedChains = append(s.endedChains, chainEndCall{chainID: chainID, at: at, reason: reason})
	return nil
}

func (s *spyStore) RecordGKCountedForOwner(_ context.Context, actorID, ownerID, gkID int64, at time.Time) error {
	s.recordedGKCountedForOwner = append(s.recordedGKCountedForOwner, struct {
		actorID int64
		ownerID int64
		gkID    int64
		at      time.Time
	}{actorID: actorID, ownerID: ownerID, gkID: gkID, at: at})
	return nil
}

func (s *spyStore) GetGKHomeCountry(ctx context.Context, gkID int64) (string, error) {
	if s.getGKHomeCountryFn != nil {
		return s.getGKHomeCountryFn(ctx, gkID)
	}
	return s.MockStore.GetGKHomeCountry(ctx, gkID)
}

func (s *spyStore) GetGKCountriesVisited(ctx context.Context, gkID int64) (map[string]bool, error) {
	if s.getGKCountriesVisitedFn != nil {
		return s.getGKCountriesVisitedFn(ctx, gkID)
	}
	return s.MockStore.GetGKCountriesVisited(ctx, gkID)
}

func (s *spyStore) GetGKLastDrop(ctx context.Context, gkID int64, before time.Time) (*time.Time, int64, error) {
	if s.getGKLastDropFn != nil {
		return s.getGKLastDropFn(ctx, gkID, before)
	}
	return s.MockStore.GetGKLastDrop(ctx, gkID, before)
}

func (s *spyStore) GetGKLastCacheEntry(ctx context.Context, gkID int64, before time.Time) (*time.Time, error) {
	if s.getGKLastCacheEntryFn != nil {
		return s.getGKLastCacheEntryFn(ctx, gkID, before)
	}
	return s.MockStore.GetGKLastCacheEntry(ctx, gkID, before)
}

func (s *spyStore) GetUserMoveHistoryOnGK(ctx context.Context, userID, gkID int64) (map[pipeline.LogType]bool, error) {
	if s.getUserMoveHistoryOnGKFn != nil {
		return s.getUserMoveHistoryOnGKFn(ctx, userID, gkID)
	}
	return s.MockStore.GetUserMoveHistoryOnGK(ctx, userID, gkID)
}

func (s *spyStore) GetActorGKsPerOwnerCount(ctx context.Context, actorID, ownerID int64) (int, error) {
	if s.getActorGKsPerOwnerCountFn != nil {
		return s.getActorGKsPerOwnerCountFn(ctx, actorID, ownerID)
	}
	return s.MockStore.GetActorGKsPerOwnerCount(ctx, actorID, ownerID)
}

func (s *spyStore) IsGKAlreadyCountedForOwner(ctx context.Context, actorID, ownerID, gkID int64) (bool, error) {
	if s.isGKAlreadyCountedForOwnerFn != nil {
		return s.isGKAlreadyCountedForOwnerFn(ctx, actorID, ownerID, gkID)
	}
	return s.MockStore.IsGKAlreadyCountedForOwner(ctx, actorID, ownerID, gkID)
}

func (s *spyStore) GetActorGKsAtLocationThisMonth(ctx context.Context, actorID int64, locationID string, yearMonth string) (int, error) {
	if s.getActorGKsAtLocationThisMonthFn != nil {
		return s.getActorGKsAtLocationThisMonthFn(ctx, actorID, locationID, yearMonth)
	}
	return s.MockStore.GetActorGKsAtLocationThisMonth(ctx, actorID, locationID, yearMonth)
}

func (s *spyStore) GetActorMonthlyDiversity(ctx context.Context, actorID int64, yearMonth string) (int, bool, int, bool, error) {
	if s.getActorMonthlyDiversityFn != nil {
		return s.getActorMonthlyDiversityFn(ctx, actorID, yearMonth)
	}
	return s.MockStore.GetActorMonthlyDiversity(ctx, actorID, yearMonth)
}

func (s *spyStore) GetActorMonthlyDiversityCountries(ctx context.Context, actorID int64, yearMonth string) (map[string]bool, error) {
	if s.getActorMonthlyDiversityCountriesFn != nil {
		return s.getActorMonthlyDiversityCountriesFn(ctx, actorID, yearMonth)
	}
	return s.MockStore.GetActorMonthlyDiversityCountries(ctx, actorID, yearMonth)
}

func (s *spyStore) GetActorMonthlyDiversityOwners(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error) {
	if s.getActorMonthlyDiversityOwnersFn != nil {
		return s.getActorMonthlyDiversityOwnersFn(ctx, actorID, yearMonth)
	}
	return s.MockStore.GetActorMonthlyDiversityOwners(ctx, actorID, yearMonth)
}

func (s *spyStore) GetActorMonthlyDiversityDrops(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error) {
	if s.getActorMonthlyDiversityDropsFn != nil {
		return s.getActorMonthlyDiversityDropsFn(ctx, actorID, yearMonth)
	}
	return s.MockStore.GetActorMonthlyDiversityDrops(ctx, actorID, yearMonth)
}

func (s *spyStore) RecordActorGKAtLocation(_ context.Context, actorID int64, locationID string, yearMonth string, gkID int64, at time.Time) error {
	s.recordedActorGKAtLocation = append(s.recordedActorGKAtLocation, struct {
		actorID    int64
		locationID string
		yearMonth  string
		gkID       int64
		at         time.Time
	}{actorID: actorID, locationID: locationID, yearMonth: yearMonth, gkID: gkID, at: at})
	return nil
}

func (s *spyStore) AddGKCountryVisited(_ context.Context, gkID int64, country string, at time.Time, moveID int64) error {
	s.addedCountriesVisited = append(s.addedCountriesVisited, struct {
		gkID    int64
		country string
		at      time.Time
		moveID  int64
	}{gkID: gkID, country: country, at: at, moveID: moveID})
	return nil
}

func (s *spyStore) IncrementActorMonthlyDrops(_ context.Context, actorID int64, yearMonth string, gkID int64, at time.Time) error {
	s.incrementDropsCalls = append(s.incrementDropsCalls, struct {
		actorID   int64
		yearMonth string
		gkID      int64
		at        time.Time
	}{actorID: actorID, yearMonth: yearMonth, gkID: gkID, at: at})
	return nil
}

func (s *spyStore) SetDropsBonusAwarded(_ context.Context, actorID int64, yearMonth string) error {
	s.setDropsBonusCalls = append(s.setDropsBonusCalls, struct {
		actorID   int64
		yearMonth string
	}{actorID: actorID, yearMonth: yearMonth})
	return nil
}

func (s *spyStore) IncrementActorMonthlyOwners(_ context.Context, actorID int64, yearMonth string, ownerID int64, at time.Time) error {
	s.incrementOwnersCalls = append(s.incrementOwnersCalls, struct {
		actorID   int64
		yearMonth string
		ownerID   int64
		at        time.Time
	}{actorID: actorID, yearMonth: yearMonth, ownerID: ownerID, at: at})
	return nil
}

func (s *spyStore) SetOwnersBonusAwarded(_ context.Context, actorID int64, yearMonth string) error {
	s.setOwnersBonusCalls = append(s.setOwnersBonusCalls, struct {
		actorID   int64
		yearMonth string
	}{actorID: actorID, yearMonth: yearMonth})
	return nil
}

func (s *spyStore) RecordActorDiversityCountry(_ context.Context, actorID int64, yearMonth, country string, at time.Time) error {
	s.recordDiversityCountryCalls = append(s.recordDiversityCountryCalls, struct {
		actorID   int64
		yearMonth string
		country   string
		at        time.Time
	}{actorID: actorID, yearMonth: yearMonth, country: country, at: at})
	return nil
}

func (s *spyStore) SaveGKMultiplierState(_ context.Context, gkID int64, multiplier float64, at time.Time, holderID int64, holderAcquiredAt *time.Time) error {
	s.saveMultiplierCalls = append(s.saveMultiplierCalls, multiplierSaveCall{
		gkID: gkID, multiplier: multiplier, at: at, holderID: holderID, holderAcquiredAt: holderAcquiredAt,
	})
	return nil
}

func (s *spyStore) LogGKMultiplierChange(_ context.Context, _, _ int64, _, _ float64, _, _ string) error {
	s.logMultiplierChangeCalls++
	return nil
}

func (s *spyStore) AddUserMoveHistory(_ context.Context, userID, gkID int64, logType pipeline.LogType, at time.Time) error {
	s.addUserMoveHistoryCalls = append(s.addUserMoveHistoryCalls, struct {
		userID  int64
		gkID    int64
		logType pipeline.LogType
		at      time.Time
	}{userID: userID, gkID: gkID, logType: logType, at: at})
	return nil
}

func (s *spyStore) RecordChainCompletion(_ context.Context, userID, gkID, chainID int64, at time.Time) error {
	s.recordChainCompletionCalls = append(s.recordChainCompletionCalls, struct {
		userID  int64
		gkID    int64
		chainID int64
		at      time.Time
	}{userID: userID, gkID: gkID, chainID: chainID, at: at})
	return nil
}

func mustAwardByLabel(acc *pipeline.Accumulator, label string) pipeline.Award {
	award, ok := awardByLabel(acc, label)
	if !ok {
		panic(fmt.Sprintf("award with label %s not found", label))
	}
	return award
}

var _ = context.Background // ensure context is imported
