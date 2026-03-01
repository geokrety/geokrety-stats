// Package models defines JSON:API-style response wrappers and domain types.
package models

import "time"

// ── JSON:API envelope ────────────────────────────────────────────────────────

// Response wraps a collection of resources.
type Response struct {
	Data  interface{} `json:"data"`
	Meta  Meta        `json:"meta"`
	Links *Links      `json:"links,omitempty"`
}

// Meta holds pagination and context metadata.
type Meta struct {
	Total     int64  `json:"total,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	HasNext   bool   `json:"has_next,omitempty"`
	HasPrev   bool   `json:"has_prev,omitempty"`
	Period    string `json:"period,omitempty"`
	ComputedAt *time.Time `json:"computed_at,omitempty"`
}

// Links holds pagination links.
type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// ── Leaderboard ──────────────────────────────────────────────────────────────

type LeaderboardEntry struct {
	Rank               int64      `json:"rank"`
	UserID             int64      `json:"user_id"`
	Username           string     `json:"username"`
	HomeCountry        *string    `json:"home_country,omitempty"`
	TotalPoints        float64    `json:"total_points"`
	GkCount            int64      `json:"gk_count,omitempty"`
	MoveCount          int64      `json:"move_count,omitempty"`
	CountriesCount     int64      `json:"countries_count,omitempty"`
	LastActive         *time.Time `json:"last_active,omitempty"`
	AvgPointsPerMove   float64    `json:"avg_points_per_move,omitempty"`
	PointsPeriod       float64    `json:"points_period,omitempty"`
}

// ── User ─────────────────────────────────────────────────────────────────────

type User struct {
	UserID             int64      `json:"user_id"`
	Username           string     `json:"username"`
	Avatar             *string    `json:"avatar,omitempty"`
	HomeCountry        *string    `json:"home_country,omitempty"`
	JoinedAt           *time.Time `json:"joined_at,omitempty"`
	TotalPoints        float64    `json:"total_points"`
	RankAllTime        int64      `json:"rank_all_time,omitempty"`
	TotalMoves         int64      `json:"total_moves"`
	TotalDrops         int64      `json:"total_drops"`
	TotalGrabs         int64      `json:"total_grabs"`
	TotalComments      int64      `json:"total_comments"`
	TotalSeen          int64      `json:"total_seen"`
	TotalDips          int64      `json:"total_dips"`
	TotalArchived      int64      `json:"total_archived,omitempty"`
	DistinctGKs        int64      `json:"distinct_gks"`
	DistinctOwners     int64      `json:"distinct_owners"`
	CountriesCount     int64      `json:"countries_count"`
	CachesCount        int64      `json:"caches_count"`
	KmContributed      float64    `json:"km_contributed"`
	ActiveDays         int64      `json:"active_days"`
	FirstMoveAt        *time.Time `json:"first_move_at,omitempty"`
	LastMoveAt         *time.Time `json:"last_move_at,omitempty"`
	PtsBase            float64    `json:"pts_base"`
	PtsRelay           float64    `json:"pts_relay"`
	PtsRescuer         float64    `json:"pts_rescuer"`
	PtsChain           float64    `json:"pts_chain"`
	PtsCountry         float64    `json:"pts_country"`
	PtsDiversity       float64    `json:"pts_diversity"`
	PtsHandover        float64    `json:"pts_handover"`
	PtsReach           float64    `json:"pts_reach"`
}

// UserDailyPoints is a single day's point total for a user.
type UserDailyPoints struct {
	Day         string  `json:"day"`   // YYYY-MM-DD
	Points      float64 `json:"points"`
	MovesCount  int64   `json:"moves_count"`
}

// UserCountry is a single country entry for a user.
type UserCountry struct {
	Country    string     `json:"country"`
	MoveCount  int64      `json:"move_count"`
	FirstVisit *time.Time `json:"first_visit,omitempty"`
	LastVisit  *time.Time `json:"last_visit,omitempty"`
}

// UserMove is a single move entry.
type UserMove struct {
	MoveID   int64       `json:"move_id"`
	GkID     int64       `json:"gk_id"`
	GkName   string      `json:"gk_name,omitempty"`
	GkAvatar *string     `json:"gk_avatar,omitempty"`
	MoveType int         `json:"move_type"`
	TypeName string      `json:"type_name"`
	Country  *string     `json:"country,omitempty"`
	Waypoint *string     `json:"waypoint,omitempty"`
	Distance *int        `json:"distance,omitempty"`
	MovedOn  *time.Time  `json:"moved_on,omitempty"`
	Points   *float64    `json:"points,omitempty"`
	ChainID  *int64      `json:"chain_id,omitempty"`
}

// ── GeoKret ──────────────────────────────────────────────────────────────────

type GeoKret struct {
	GkID                int64      `json:"gk_id"`
	GkHexID             string     `json:"gk_hex_id"`          // e.g. "GK1234" (GKXXXX public code)
	Name                string     `json:"gk_name"`
	Avatar              *string    `json:"avatar,omitempty"`
	GkType              int        `json:"gk_type"`
	GkTypeName          string     `json:"gk_type_name"`
	Missing             bool       `json:"missing"`
	InCache             bool       `json:"in_cache"`            // true when holder is NULL (GK is in a cache)
	IsNonCollectible    bool       `json:"is_non_collectible"` // non-transferable/sealed GK
	IsParked            bool       `json:"is_parked"`           // currently parked
	LovesCount          int        `json:"loves_count"`
	Distance            int64      `json:"distance_km"`
	CachesCount         int        `json:"caches_count"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	BornAt              *time.Time `json:"born_at,omitempty"`
	OwnerID             *int64     `json:"owner_id,omitempty"`
	OwnerUsername       *string    `json:"owner_username,omitempty"`
	OwnerHomeCountry    *string    `json:"owner_home_country,omitempty"`
	HolderID            *int64     `json:"holder_id,omitempty"`
	HolderUsername      *string    `json:"holder_username,omitempty"`
	HolderHomeCountry   *string    `json:"holder_home_country,omitempty"`
	CacheCountry        *string    `json:"cache_country,omitempty"` // Country where GK is currently cached
	TotalMoves          int64      `json:"total_moves"`
	TotalDrops          int64      `json:"total_drops"`
	TotalGrabs          int64      `json:"total_grabs"`
	TotalComments       int64      `json:"total_comments"`
	TotalSeen           int64      `json:"total_seen"`
	TotalDips           int64      `json:"total_dips"`
	DistinctUsers       int64      `json:"distinct_users"`
	CountriesCount      int64      `json:"countries_count"`
	DistinctCaches      int64      `json:"distinct_caches"`
	TotalPointsGenerated float64   `json:"total_points_generated"`
	RankAllTime         *int64     `json:"rank_all_time,omitempty"`
	UsersAwarded        int64      `json:"users_awarded"`
	CurrentMultiplier   float64    `json:"current_multiplier"`
	FirstMoveAt         *time.Time `json:"first_move_at,omitempty"`
	LastMoveAt          *time.Time `json:"last_move_at,omitempty"`
}

// GkCountry is a country entry for a geokret.
type GkCountry struct {
	Country    string `json:"country"`
	MoveCount  int64  `json:"move_count"`
}

// GkMove is a single move for a geokret.
type GkMove struct {
	MoveID        int64       `json:"move_id"`
	AuthorID      *int64      `json:"author_id,omitempty"`
	AuthorUsername string      `json:"author_username,omitempty"`
	AuthorAvatar  *string     `json:"author_avatar,omitempty"`
	MoveType      int         `json:"move_type"`
	TypeName      string      `json:"type_name"`
	Country       *string     `json:"country,omitempty"`
	Waypoint      *string     `json:"waypoint,omitempty"`
	Distance      *int        `json:"distance,omitempty"`
	MovedOn       *time.Time  `json:"moved_on,omitempty"`
	Points        *float64    `json:"points,omitempty"`
	ChainID       *int64      `json:"chain_id,omitempty"`
}

type ChainSummary struct {
	ChainID             int64      `json:"chain_id"`
	GkID                int64      `json:"gk_id"`
	GkHexID             string     `json:"gk_hex_id"`
	GkName              string     `json:"gk_name"`
	GkAvatarKey         *string    `json:"gk_avatar_key,omitempty"`
	GkType              int        `json:"gk_type"`
	Status              string     `json:"status"`
	StartedAt           time.Time  `json:"started_at"`
	EndedAt             *time.Time `json:"ended_at,omitempty"`
	ChainLastActive     time.Time  `json:"chain_last_active"`
	EndReason           *string    `json:"end_reason,omitempty"`
	MemberCount         int64      `json:"member_count"`
	ChainPoints         float64    `json:"chain_points"`
	HasUserCompletion   bool       `json:"has_user_completion,omitempty"`
	MoveID              *int64     `json:"move_id,omitempty"`
	MoveChainPoints     *float64   `json:"move_chain_points,omitempty"`
}

type ChainMember struct {
	ChainID        int64      `json:"chain_id"`
	UserID         int64      `json:"user_id"`
	Username       string     `json:"username"`
	UserAvatarKey  *string    `json:"user_avatar_key,omitempty"`
	Position       int        `json:"position"`
	JoinedAt       *time.Time `json:"joined_at,omitempty"`
}

type ChainMove struct {
	MoveID         int64      `json:"move_id"`
	AuthorID       *int64     `json:"author_id,omitempty"`
	AuthorUsername *string    `json:"author_username,omitempty"`
	AuthorAvatarKey *string    `json:"author_avatar_key,omitempty"`
	MoveType       int        `json:"move_type"`
	TypeName       string     `json:"type_name"`
	Country        *string    `json:"country,omitempty"`
	Waypoint       *string    `json:"waypoint,omitempty"`
	MovedOn        *time.Time `json:"moved_on,omitempty"`
	ChainPoints    float64    `json:"chain_points"`
}

// ── Global Stats ─────────────────────────────────────────────────────────────

type GlobalStats struct {
	TotalUsers         int64      `json:"total_users"`
	TotalGKs           int64      `json:"total_gks"`
	TotalMoves         int64      `json:"total_moves"`
	ScoredUsers        int64      `json:"scored_users"`
	TotalPointsAwarded float64    `json:"total_points_awarded"`
	CountriesReached   int64      `json:"countries_reached"`
	TotalKm            int64      `json:"total_km"`
	TotalImages        int64      `json:"total_images"`
	TotalLoves         int64      `json:"total_loves"`
	ComputedAt         *time.Time `json:"computed_at,omitempty"`
}

// ── WebSocket ────────────────────────────────────────────────────────────────

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
