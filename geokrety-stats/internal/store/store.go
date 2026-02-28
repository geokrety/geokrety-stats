// Package store defines the data access interface for the scoring pipeline.
// All database queries needed by the computers are declared here.
// This interface enables easy mocking in unit tests.
package store

import (
	"context"
	"time"

	"github.com/geokrety/geokrety-points-system/internal/pipeline"
)

// GKMoveRow represents a row from geokrety.gk_moves for context loading.
type GKMoveRow struct {
	ID            int64
	GeoKretID     int64
	Author        *int64
	MoveType      int
	Waypoint      *string
	Country       *string
	Lat           *float64
	Lon           *float64
	MovedOnDatetime time.Time
}

// GKRow represents a row from geokrety.gk_geokrety.
type GKRow struct {
	ID        int64
	Type      int
	Owner     *int64
	CreatedAt time.Time
	Holder    *int64 // Current holder user_id (nil = in cache)
}

// ChainRow represents an active chain from geokrety_stats.gk_chains.
type ChainRow struct {
	ID               int64
	GKID             int64
	Status           string
	StartedAt        time.Time
	EndedAt          *time.Time
	ChainLastActive  time.Time
	HolderAcquiredAt *time.Time
}

// Store is the interface for all database operations used by pipeline computers.
// The postgres implementation queries geokrety (read-only) and geokrety_stats (read-write).
type Store interface {
	// ---- Event idempotency ----

	// IsEventProcessed returns true if this move_id has already been handled.
	IsEventProcessed(ctx context.Context, moveID int64) (bool, error)

	// MarkEventProcessed records a move_id as processed with the given result.
	MarkEventProcessed(ctx context.Context, moveID int64, result string) error

	// ---- GeoKret read (geokrety schema, read-only) ----

	// GetMove loads a gk_moves row by ID.
	GetMove(ctx context.Context, moveID int64) (*GKMoveRow, error)

	// GetGK loads a gk_geokrety row by ID.
	GetGK(ctx context.Context, gkID int64) (*GKRow, error)

	// GetGKPreviousHolder returns the user_id who held the GK before the current holder.
	// Returns 0 if the previous holder cannot be determined.
	GetGKPreviousHolder(ctx context.Context, gkID int64) (int64, error)

	// GetGKHomeCountry returns the first country ever recorded for this GK.
	GetGKHomeCountry(ctx context.Context, gkID int64) (string, error)

	// GetGKLastDrop loads the timestamp and user_id of the most recent DROP log.
	GetGKLastDrop(ctx context.Context, gkID int64, before time.Time) (dropAt *time.Time, dropUser int64, err error)

	// GetGKLastCacheEntry loads the timestamp when the GK last entered a cache.
	GetGKLastCacheEntry(ctx context.Context, gkID int64, before time.Time) (*time.Time, error)

	// GetGKDistinctUsers6M returns distinct user_ids who moved this GK in the last 6 months.
	GetGKDistinctUsers6M(ctx context.Context, gkID int64, before time.Time) ([]int64, error)

	// ---- Multiplier state (geokrety_stats schema) ----

	// GetGKMultiplierState loads current multiplier and last update time.
	// Returns a default state (multiplier=1.0) if none exists.
	GetGKMultiplierState(ctx context.Context, gkID int64) (multiplier float64, lastUpdatedAt time.Time, holderID int64, holderAcquiredAt *time.Time, err error)

	// SaveGKMultiplierState persists updated multiplier state.
	SaveGKMultiplierState(ctx context.Context, gkID int64, multiplier float64, at time.Time, holderID int64, holderAcquiredAt *time.Time) error

	// LogGKMultiplierChange appends a row to gk_points_log.
	LogGKMultiplierChange(ctx context.Context, gkID int64, moveID int64, oldMult, newMult float64, reason, source string) error

	// ---- Countries visited (geokrety_stats schema) ----

	// GetGKCountriesVisited returns all country codes visited by this GK (from stats schema).
	GetGKCountriesVisited(ctx context.Context, gkID int64) (map[string]bool, error)

	// AddGKCountryVisited records that this GK visited a new country.
	AddGKCountryVisited(ctx context.Context, gkID int64, country string, at time.Time, moveID int64) error

	// ---- User move history (geokrety_stats schema) ----

	// GetUserMoveHistoryOnGK returns the set of log_types the user has previously logged on this GK.
	GetUserMoveHistoryOnGK(ctx context.Context, userID, gkID int64) (map[pipeline.LogType]bool, error)

	// AddUserMoveHistory records a (user, gk, log_type) tuple.
	AddUserMoveHistory(ctx context.Context, userID, gkID int64, logType pipeline.LogType, at time.Time) error

	// ---- Owner GK limit (geokrety_stats schema) ----

	// GetActorGKsPerOwnerCount returns how many distinct GKs from ownerID the actor has already earned base points from.
	GetActorGKsPerOwnerCount(ctx context.Context, actorID, ownerID int64) (int, error)

	// IsGKAlreadyCountedForOwner returns true if the actor already earned base points on this gk from this owner.
	IsGKAlreadyCountedForOwner(ctx context.Context, actorID, ownerID, gkID int64) (bool, error)

	// RecordGKCountedForOwner increments the actor-owner GK count.
	RecordGKCountedForOwner(ctx context.Context, actorID, ownerID, gkID int64, at time.Time) error

	// ---- Waypoint penalty (geokrety_stats schema) ----

	// GetActorGKsAtLocationThisMonth returns count of distinct GKs actor has scored at this location this month.
	GetActorGKsAtLocationThisMonth(ctx context.Context, actorID int64, locationID string, yearMonth string) (int, error)

	// RecordActorGKAtLocation records that actor scored a GK at this location this month.
	RecordActorGKAtLocation(ctx context.Context, actorID int64, locationID string, yearMonth string, gkID int64, at time.Time) error

	// ---- Monthly diversity (geokrety_stats schema) ----

	// GetActorMonthlyDiversity loads the actor's monthly diversity state.
	GetActorMonthlyDiversity(ctx context.Context, actorID int64, yearMonth string) (
		dropsCount int, dropsBonusAwarded bool,
		ownersCount int, ownersBonusAwarded bool,
		err error,
	)

	// GetActorMonthlyDiversityCountries returns countries for which actor already got diversity bonus this month.
	GetActorMonthlyDiversityCountries(ctx context.Context, actorID int64, yearMonth string) (map[string]bool, error)

	// GetActorMonthlyDiversityOwners returns owner_ids already counted for actor this month.
	GetActorMonthlyDiversityOwners(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error)

	// GetActorMonthlyDiversityDrops returns gk_ids already counted in actor's monthly drops.
	GetActorMonthlyDiversityDrops(ctx context.Context, actorID int64, yearMonth string) (map[int64]bool, error)

	// IncrementActorMonthlyDrops adds a GK to actor's monthly drop count.
	IncrementActorMonthlyDrops(ctx context.Context, actorID int64, yearMonth string, gkID int64, at time.Time) error

	// SetDropsBonusAwarded marks the 5-drops bonus as awarded this month.
	SetDropsBonusAwarded(ctx context.Context, actorID int64, yearMonth string) error

	// IncrementActorMonthlyOwners adds an owner to actor's monthly distinct owners count.
	IncrementActorMonthlyOwners(ctx context.Context, actorID int64, yearMonth string, ownerID int64, at time.Time) error

	// SetOwnersBonusAwarded marks the 10-owners bonus as awarded this month.
	SetOwnersBonusAwarded(ctx context.Context, actorID int64, yearMonth string) error

	// RecordActorDiversityCountry records that actor received diversity country bonus for this country this month.
	RecordActorDiversityCountry(ctx context.Context, actorID int64, yearMonth, country string, at time.Time) error

	// ---- Chain state (geokrety_stats schema) ----

	// GetActiveChain loads the currently active chain for the GK. Returns nil if none.
	GetActiveChain(ctx context.Context, gkID int64) (*ChainRow, error)

	// GetChainMembers returns the ordered list of user_ids in a chain.
	GetChainMembers(ctx context.Context, chainID int64) ([]int64, error)

	// CreateChain creates a new chain record and returns the new chain ID.
	CreateChain(ctx context.Context, gkID int64, at time.Time) (int64, error)

	// AddChainMember adds a user to a chain if not already present.
	AddChainMember(ctx context.Context, chainID int64, userID int64, at time.Time) error

	// UpdateChainLastActive updates the chain_last_active and optionally holder_acquired_at.
	UpdateChainLastActive(ctx context.Context, chainID int64, at time.Time, holderAcquiredAt *time.Time) error

	// EndChain marks a chain as ended with the given reason.
	EndChain(ctx context.Context, chainID int64, at time.Time, reason string) error

	// GetChainLastBonus returns the last time a user received a chain bonus for a specific GK.
	// Returns zero time if never.
	GetChainLastBonus(ctx context.Context, userID, gkID int64) (time.Time, error)

	// RecordChainCompletion records that a user earned a chain bonus for this chain.
	RecordChainCompletion(ctx context.Context, userID, gkID, chainID int64, at time.Time) error

	// GetExpiredChains returns all active chains that have been inactive for >= timeoutDays days.
	GetExpiredChains(ctx context.Context, timeoutDays int, asOf time.Time) ([]*ChainRow, error)

	// ---- Point recording (geokrety_stats schema) ----

	// SaveAwards persists the final award list and updates user_points_totals.
	SaveAwards(ctx context.Context, awards []pipeline.FinalAward) error

	// ---- Historical replay (geokrety schema, read-only) ----

	// GetMoveIDsPage returns a page of move IDs in ascending order.
	// Only IDs with id > afterID, id <= maxID (0 = no limit),
	// and moved_on_datetime within the optional date range are returned.
	GetMoveIDsPage(ctx context.Context, afterID, maxID int64, startDate, endDate *time.Time, limit int) ([]int64, error)

	// RotateWaypointMonthlyPartitions ensures near-future partitions exist and drops obsolete month partitions.
	RotateWaypointMonthlyPartitions(ctx context.Context, asOf time.Time, retainMonths int, futureMonths int) error

	// PruneEndedChainsOlderThan deletes ended chain rows older than the provided timestamp.
	PruneEndedChainsOlderThan(ctx context.Context, before time.Time) (int64, error)

	// PruneGKPointsLogOlderThan deletes multiplier audit rows older than the provided timestamp.
	PruneGKPointsLogOlderThan(ctx context.Context, before time.Time) (int64, error)
}
