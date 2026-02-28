// Package pipeline defines the core data structures flowing through the scoring pipeline.
package pipeline

import (
	"fmt"
	"time"
)

// LogType represents the type of a GeoKret move log.
type LogType int

const (
	LogTypeDrop     LogType = 0
	LogTypeGrab     LogType = 1
	LogTypeComment  LogType = 2
	LogTypeSeen     LogType = 3
	LogTypeArchived LogType = 4
	LogTypeDip      LogType = 5
)

// String returns a human-readable name for the log type.
func (l LogType) String() string {
	switch l {
	case LogTypeDrop:
		return "DROP"
	case LogTypeGrab:
		return "GRAB"
	case LogTypeComment:
		return "COMMENT"
	case LogTypeSeen:
		return "SEEN"
	case LogTypeArchived:
		return "ARCHIVED"
	case LogTypeDip:
		return "DIP"
	default:
		return "UNKNOWN"
	}
}

// IsScoreable returns true if this log type can produce points.
func (l LogType) IsScoreable() bool {
	switch l {
	case LogTypeDrop, LogTypeGrab, LogTypeSeen, LogTypeArchived, LogTypeDip:
		return true
	default:
		return false
	}
}

// HasBasePoints returns true if this log type can produce base points
// (ARCHIVED and DIP don't produce base points by default).
func (l LogType) HasBasePoints() bool {
	switch l {
	case LogTypeDrop, LogTypeGrab, LogTypeSeen:
		return true
	default:
		return false
	}
}

// RequiresWaypoint returns true if this log type requires a waypoint to earn points.
func (l LogType) RequiresWaypoint() bool {
	switch l {
	case LogTypeDrop, LogTypeSeen, LogTypeDip:
		return true
	default:
		return false
	}
}

// Event is the raw input event entering the pipeline.
// Derived from geokrety.gk_moves.
type Event struct {
	// LogID is the unique ID of the move in geokrety.gk_moves.
	LogID int64
	// UserID is the user who performed this move (0 = anonymous).
	UserID int64
	// GKID is the GeoKret ID.
	GKID int64
	// LogType is the type of move.
	LogType LogType
	// Waypoint is the cache/POI identifier (empty = no waypoint).
	Waypoint string
	// Country is the ISO 3166-1 alpha-2 country code (empty = unknown).
	Country string
	// Lat/Lon are the coordinates (0 = no coords).
	Lat float64
	Lon float64
	// LoggedAt is when this event was recorded (from moved_on_datetime).
	LoggedAt time.Time
}

// HasUser returns true if the event has an authenticated user.
func (e *Event) HasUser() bool {
	return e.UserID != 0
}

// HasWaypoint returns true if the event has a waypoint.
func (e *Event) HasWaypoint() bool {
	return e.Waypoint != ""
}

// HasCountry returns true if the event has a country code.
func (e *Event) HasCountry() bool {
	return e.Country != ""
}

// HasCoordinates returns true if the event has lat/lon coordinates.
func (e *Event) HasCoordinates() bool {
	return e.Lat != 0 || e.Lon != 0
}

// LocationID returns a canonical string for the event's location.
// Prefers waypoint over coordinates.
func (e *Event) LocationID() string {
	if e.Waypoint != "" {
		return e.Waypoint
	}
	if e.HasCoordinates() {
		// Round to ~100m precision to group nearby coords
		return formatCoordinate(e.Lat, e.Lon)
	}
	return ""
}

// GKState holds the GeoKret's state loaded before this event.
type GKState struct {
	GKID              int64
	GKType            int
	OwnerID           int64
	CreatedAt         time.Time
	CurrentMultiplier float64
	LastMultiplierAt  time.Time
	// CurrentHolder is the user_id currently holding the GK (0 = in cache).
	CurrentHolder int64
	// PreviousHolder is the user_id who held it before current (0 = was in cache).
	PreviousHolder int64
	// HomeCountry is the first country the GK was ever seen in.
	HomeCountry string
}

// IsInCache returns true if the GK is currently in a cache (not being held by a user).
func (g *GKState) IsInCache() bool {
	return g.CurrentHolder == 0
}

// GKHistory holds derived historical data for the GeoKret.
type GKHistory struct {
	// LastDropAt is the UTC timestamp of the most recent DROP log.
	LastDropAt *time.Time
	// LastDropUser is the user_id who made that DROP.
	LastDropUser int64
	// LastCacheEntryAt is the UTC timestamp of the last time GK entered cache.
	LastCacheEntryAt *time.Time
	// DistinctUsers6M is the list of distinct user_ids who moved this GK in the last 6 months.
	DistinctUsers6M []int64
	// CountriesVisited is the full set of country codes this GK has visited (from stats schema).
	CountriesVisited map[string]bool
}

// UserState holds the actor's state relevant to the current event.
type UserState struct {
	// ActorMoveHistoryOnGK is the set of log_types the actor has previously logged on this GK.
	ActorMoveHistoryOnGK map[LogType]bool
	// ActorGKsPerOwnerCount is how many distinct GKs from this owner the actor has already earned base points from.
	ActorGKsPerOwnerCount int
	// ActorGKAlreadyCountedForOwner is true if this specific GK was already counted in actor's owner-GK limit.
	ActorGKAlreadyCountedForOwner bool
	// ActorGKsAtLocationThisMonth is the count of distinct GKs scored at this location this month.
	ActorGKsAtLocationThisMonth int
	// ActorCountriesVisitedThisMonth is the set of countries for which actor already got diversity country bonus.
	ActorCountriesVisitedThisMonth map[string]bool
	// ActorGKsDroppedThisMonth is the count of distinct GKs dropped this month (scored drops).
	ActorGKsDroppedThisMonth int
	// ActorGKAlreadyDroppedThisMonth is true if this GK was already counted in dropped set this month.
	ActorGKAlreadyDroppedThisMonth bool
	// ActorDistinctOwnersThisMonth is the count of distinct GK owners interacted with this month.
	ActorDistinctOwnersThisMonth int
	// ActorOwnerAlreadyCountedThisMonth is true if this GK's owner was already counted this month.
	ActorOwnerAlreadyCountedThisMonth bool
	// DropsBonusAlreadyAwarded is true if the 5-drops diversity bonus was already awarded this month.
	DropsBonusAlreadyAwarded bool
	// OwnersBonusAlreadyAwarded is true if the 10-owners diversity bonus was already awarded this month.
	OwnersBonusAlreadyAwarded bool
}

// ChainState holds the active chain state for the GeoKret.
type ChainState struct {
	// ActiveChainID is the ID of the current active chain (0 = none).
	ActiveChainID int64
	// ChainMembers is the ordered list of distinct user_ids in the chain.
	ChainMembers []int64
	// ChainLastActive is the UTC timestamp of last chain-extending activity.
	ChainLastActive time.Time
	// HolderAcquiredAt is when the current holder took possession.
	HolderAcquiredAt time.Time

	// These are written by module 10 (chain state manager):
	// ChainEnded is true if a chain was ended during this event processing.
	ChainEnded bool
	// EndedChainID is the ID of the chain that ended.
	EndedChainID int64
	// EndedChainMembers is the list of users in the ended chain.
	EndedChainMembers []int64
}

// RuntimeFlags holds mutable flags written by computers and read by later ones.
type RuntimeFlags struct {
	// NewCountryVisited is true if this move entered a new country for this GK.
	NewCountryVisited bool
	// NewCountryCode is the ISO code of the new country (if any).
	NewCountryCode string
	// BasePointsAwarded is the raw base points awarded to the actor.
	BasePointsAwarded float64
	// BasePointsLabel is the label used for base award entries.
	BasePointsLabel string
	// ActorScoredThisGK is true if actor earned any base points this event.
	ActorScoredThisGK bool
}

// Context is the shared context flowing through the scoring pipeline.
// It is assembled by computer 01 and flows read-mostly through the rest.
type Context struct {
	Event        Event
	GKState      GKState
	GKHistory    GKHistory
	UserState    UserState
	ChainState   ChainState
	RuntimeFlags RuntimeFlags
	// AggregatedAwards is populated by computer 14 (aggregator) with the final
	// validated and consolidated list of awards. The pipeline runner reads this.
	AggregatedAwards []FinalAward
}

// Award is a single point award record in the accumulator.
type Award struct {
	// RecipientUserID is the user who receives the points.
	RecipientUserID int64
	// Points is the amount awarded (always >= 0).
	Points float64
	// Reason is a human-readable explanation.
	Reason string
	// Label is a machine-readable category (e.g., "base_move", "relay_mover").
	Label string
	// ModuleSource is the computer that generated this award.
	ModuleSource string
	// IsOwnerReward indicates whether this is an owner-directed reward.
	IsOwnerReward bool
}

// FinalAward is the per-recipient summary emitted by the aggregator.
type FinalAward struct {
	RecipientUserID int64
	TotalPoints     float64
	EventLogID      int64
	GKID            int64
	Awards          []Award
}

// formatCoordinate returns a compact string representation of coordinates.
// Rounded to 4 decimal places (~11m precision).
func formatCoordinate(lat, lon float64) string {
	// Use integer multiplication to avoid floating-point formatting issues
	iLat := int64(lat * 10000)
	iLon := int64(lon * 10000)
	return fmt.Sprintf("%d,%d", iLat, iLon)
}
