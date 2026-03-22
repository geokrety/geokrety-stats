package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/config"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
}

type PictureTypeStats struct {
	GeokretAvatars int64 `json:"geokretAvatars" xml:"geokretAvatars"`
	GeokretMoves   int64 `json:"geokretMoves" xml:"geokretMoves"`
	UserAvatars    int64 `json:"userAvatars" xml:"userAvatars"`
}

type GeokretyTypeStats struct {
	Traditional int64 `json:"traditional" xml:"traditional"`
	Book        int64 `json:"book" xml:"book"`
	Human       int64 `json:"human" xml:"human"`
	Coin        int64 `json:"coin" xml:"coin"`
	Kretypost   int64 `json:"kretypost" xml:"kretypost"`
	Pebble      int64 `json:"pebble" xml:"pebble"`
	Car         int64 `json:"car" xml:"car"`
	Playingcard int64 `json:"playingcard" xml:"playingcard"`
	Dogtag      int64 `json:"dogtag" xml:"dogtag"`
	Jigsaw      int64 `json:"jigsaw" xml:"jigsaw"`
	Easteregg   int64 `json:"easteregg" xml:"easteregg"`
}

type MoveTypeStats struct {
	Dropped   int64 `json:"dropped" xml:"dropped"`
	Grabbed   int64 `json:"grabbed" xml:"grabbed"`
	Commented int64 `json:"commented" xml:"commented"`
	Seen      int64 `json:"seen" xml:"seen"`
	Archived  int64 `json:"archived" xml:"archived"`
	Dipped    int64 `json:"dipped" xml:"dipped"`
}

type GlobalStats struct {
	TotalGeokrety       int64             `json:"totalGeokrety" xml:"totalGeokrety"`
	TotalGeokretyHidden int64             `json:"totalGeokretyHidden" xml:"totalGeokretyHidden"`
	TotalMoves          int64             `json:"totalMoves" xml:"totalMoves"`
	MovesLast30Days     int64             `json:"movesLast30Days" xml:"movesLast30Days"`
	RegisteredUsers     int64             `json:"registeredUsers" xml:"registeredUsers"`
	ActiveUsers         int64             `json:"activeUsers" xml:"activeUsers"`
	ActiveUsersLast30d  int64             `json:"activeUsersLast30d" xml:"activeUsersLast30d"`
	CountriesReached    int64             `json:"countriesReached" xml:"countriesReached"`
	PicturesUploaded    int64             `json:"picturesUploaded" xml:"picturesUploaded"`
	PicturesByType      PictureTypeStats  `json:"picturesByType" xml:"picturesByType"`
	GeokretyByType      GeokretyTypeStats `json:"geokretyByType" xml:"geokretyByType"`
	MovesByType         MoveTypeStats     `json:"movesByType" xml:"movesByType"`
}

type RecentMove struct {
	ID          int64     `db:"id" json:"id" xml:"id"`
	GeokretName string    `db:"geokret_name" json:"geokretName" xml:"geokretName"`
	Type        string    `db:"type" json:"type" xml:"type"`
	Username    string    `db:"username" json:"username" xml:"username"`
	Country     string    `db:"country" json:"country" xml:"country"`
	CountryFlag string    `json:"countryFlag" xml:"countryFlag"`
	Timestamp   time.Time `db:"timestamp" json:"timestamp" xml:"timestamp"`
}

type LeaderboardUser struct {
	Rank        int    `json:"rank" xml:"rank"`
	UserID      int64  `db:"user_id" json:"userId" xml:"userId"`
	Username    string `db:"username" json:"username" xml:"username"`
	Initials    string `json:"initials" xml:"initials"`
	Points      int64  `db:"points" json:"points" xml:"points"`
	MovesCount  int64  `db:"moves_count" json:"movesCount" xml:"movesCount"`
	AvatarColor string `json:"avatarColor" xml:"avatarColor"`
}

type CountryStats struct {
	Code             string  `db:"code" json:"code" xml:"code"`
	Name             string  `db:"name" json:"name" xml:"name"`
	Flag             string  `json:"flag" xml:"flag"`
	MovesCount       int64   `db:"moves_count" json:"movesCount" xml:"movesCount"`
	UsersHome        int64   `db:"users_home" json:"usersHome" xml:"usersHome"`
	ActiveUsers      int64   `db:"active_users" json:"activeUsers" xml:"activeUsers"`
	Dropped          int64   `db:"dropped" json:"dropped" xml:"dropped"`
	Dipped           int64   `db:"dipped" json:"dipped" xml:"dipped"`
	Seen             int64   `db:"seen" json:"seen" xml:"seen"`
	Loves            int64   `db:"loves" json:"loves" xml:"loves"`
	Pictures         int64   `db:"pictures" json:"pictures" xml:"pictures"`
	PointsSum        float64 `db:"points_sum" json:"pointsSum" xml:"pointsSum"`
	PointsSumMoves   float64 `db:"points_sum_moves" json:"pointsSumMoves" xml:"pointsSumMoves"`
	GeokretyInCache  int64   `db:"geokrety_in_cache" json:"geokretyInCache" xml:"geokretyInCache"`
	GeokretyLost     int64   `db:"geokrety_lost" json:"geokretyLost" xml:"geokretyLost"`
	AvgPointsPerMove float64 `db:"avg_points_per_move" json:"avgPointsPerMove" xml:"avgPointsPerMove"`
}

type RecentBorn struct {
	ID        int64              `db:"id" json:"id" xml:"id"`
	GKID      geokrety.GeokretId `db:"gkid" json:"gkid" xml:"gkid"`
	Name      string             `db:"name" json:"name" xml:"name"`
	Type      int16              `db:"type" json:"type" xml:"type"`
	TypeName  string             `json:"typeName" xml:"typeName"`
	BornAt    time.Time          `db:"born_at" json:"bornAt" xml:"bornAt"`
	OwnerID   *int64             `db:"owner_id" json:"ownerId" xml:"ownerId,omitempty"`
	OwnerName string             `db:"owner_name" json:"ownerName" xml:"ownerName"`
}

type RecentLoved struct {
	GeoKretID   int64               `db:"geokret_id" json:"geokretId" xml:"geokretId"`
	GKID        *geokrety.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	GeoKretName string              `db:"geokret_name" json:"geokretName" xml:"geokretName"`
	UserID      int64               `db:"user_id" json:"userId" xml:"userId"`
	Username    string              `db:"username" json:"username" xml:"username"`
	LovedAt     time.Time           `db:"loved_at" json:"lovedAt" xml:"lovedAt"`
}

type RecentWatched struct {
	GeoKretID   int64               `db:"geokret_id" json:"geokretId" xml:"geokretId"`
	GKID        *geokrety.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	GeoKretName string              `db:"geokret_name" json:"geokretName" xml:"geokretName"`
	UserID      int64               `db:"user_id" json:"userId" xml:"userId"`
	Username    string              `db:"username" json:"username" xml:"username"`
	WatchedAt   time.Time           `db:"watched_at" json:"watchedAt" xml:"watchedAt"`
}

type ActiveCountry struct {
	Code           string    `db:"code" json:"code" xml:"code"`
	Moves          int64     `db:"moves" json:"moves" xml:"moves"`
	UniqueUsers    int64     `db:"unique_users" json:"uniqueUsers" xml:"uniqueUsers"`
	LastActivityAt time.Time `db:"last_activity_at" json:"lastActivityAt" xml:"lastActivityAt"`
	Flag           string    `json:"flag" xml:"flag"`
}

type ActiveWaypoint struct {
	Waypoint       string     `db:"waypoint" json:"waypoint" xml:"waypoint"`
	Country        *string    `db:"country" json:"country" xml:"country,omitempty"`
	Moves          int64      `db:"moves" json:"moves" xml:"moves"`
	LastActivityAt time.Time  `db:"last_activity_at" json:"lastActivityAt" xml:"lastActivityAt"`
	Lat            *float64   `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon            *float64   `db:"lon" json:"lon" xml:"lon,omitempty"`
	GeoJSON        *GeoJSONPt `json:"geojson" xml:"geojson,omitempty"`
}

type GeoJSONPt struct {
	Type        string    `json:"type" xml:"type"`
	Coordinates []float64 `json:"coordinates" xml:"coordinates>coordinate"`
}

type RecentRegisteredUser struct {
	ID          int64     `db:"id" json:"id" xml:"id"`
	Username    string    `db:"username" json:"username" xml:"username"`
	JoinedAt    time.Time `db:"joined_at" json:"joinedAt" xml:"joinedAt"`
	HomeCountry *string   `db:"home_country" json:"homeCountry" xml:"homeCountry,omitempty"`
}

type RecentActiveUser struct {
	UserID         int64     `db:"user_id" json:"userId" xml:"userId"`
	Username       string    `db:"username" json:"username" xml:"username"`
	MovesCount     int64     `db:"moves_count" json:"movesCount" xml:"movesCount"`
	CountriesCount int64     `db:"countries_count" json:"countriesCount" xml:"countriesCount"`
	LastMoveAt     time.Time `db:"last_move_at" json:"lastMoveAt" xml:"lastMoveAt"`
}

type HourlyHeatmapCell struct {
	ActivityDate time.Time `db:"activity_date" json:"activityDate" xml:"activityDate"`
	HourUTC      int       `db:"hour_utc" json:"hourUtc" xml:"hourUtc"`
	MoveType     int       `db:"move_type" json:"moveType" xml:"moveType"`
	MoveCount    int64     `db:"move_count" json:"moveCount" xml:"moveCount"`
}

type CountryFlow struct {
	YearMonth     time.Time `db:"year_month" json:"yearMonth" xml:"yearMonth"`
	FromCountry   string    `db:"from_country" json:"fromCountry" xml:"fromCountry"`
	ToCountry     string    `db:"to_country" json:"toCountry" xml:"toCountry"`
	MoveCount     int64     `db:"move_count" json:"moveCount" xml:"moveCount"`
	UniqueGKCount int64     `db:"unique_gk_count" json:"uniqueGkCount" xml:"uniqueGkCount"`
	FromFlag      string    `json:"fromFlag" xml:"fromFlag"`
	ToFlag        string    `json:"toFlag" xml:"toFlag"`
}

type TopCache struct {
	WaypointCode    string `db:"waypoint_code" json:"waypointCode" xml:"waypointCode"`
	TotalGKVisits   int64  `db:"total_gk_visits" json:"totalGkVisits" xml:"totalGkVisits"`
	DistinctGKCount int64  `db:"distinct_gks" json:"distinctGks" xml:"distinctGks"`
}

type FirstFinderLeaderboardEntry struct {
	UserID     int64  `db:"user_id" json:"userId" xml:"userId"`
	Username   string `db:"username" json:"username" xml:"username"`
	FirstFinds int64  `db:"first_finds" json:"firstFinds" xml:"firstFinds"`
	Rank       int    `json:"rank" xml:"rank"`
}

type DistanceRecord struct {
	GeoKretID   int64               `db:"gk_id" json:"geokretId" xml:"geokretId"`
	GKID        *geokrety.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	GeoKretName string              `db:"name" json:"geokretName" xml:"geokretName"`
	KMTotal     float64             `db:"km_total" json:"kmTotal" xml:"kmTotal"`
	Rank        int                 `json:"rank" xml:"rank"`
}

type UserNetworkEdge struct {
	UserID              int64     `db:"user_id" json:"userId" xml:"userId"`
	RelatedUserID       int64     `db:"related_user_id" json:"relatedUserId" xml:"relatedUserId"`
	RelatedUsername     string    `db:"related_username" json:"relatedUsername" xml:"relatedUsername"`
	SharedGeoKretyCount int64     `db:"shared_geokrety_count" json:"sharedGeokretyCount" xml:"sharedGeokretyCount"`
	FirstSeenAt         time.Time `db:"first_seen_at" json:"firstSeenAt" xml:"firstSeenAt"`
	LastSeenAt          time.Time `db:"last_seen_at" json:"lastSeenAt" xml:"lastSeenAt"`
}

type GeokretTimelineEvent struct {
	GeoKretID     int64     `db:"gk_id" json:"geokretId" xml:"geokretId"`
	EventType     string    `db:"event_type" json:"eventType" xml:"eventType"`
	OccurredAt    time.Time `db:"occurred_at" json:"occurredAt" xml:"occurredAt"`
	ActorUserID   *int64    `db:"actor_user_id" json:"actorUserId" xml:"actorUserId,omitempty"`
	ActorUsername *string   `db:"actor_username" json:"actorUsername" xml:"actorUsername,omitempty"`
}

type GeokretCirculation struct {
	GeoKretID    int64               `db:"geokrety_id" json:"geokretId" xml:"geokretId"`
	GKID         *geokrety.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	GeoKretName  string              `db:"name" json:"geokretName" xml:"geokretName"`
	Users        int64               `db:"users" json:"users" xml:"users"`
	Interactions int64               `db:"interactions" json:"interactions" xml:"interactions"`
	AvgPerUser   float64             `json:"avgInteractionsPerUser" xml:"avgInteractionsPerUser"`
}

func Open(cfg config.Config) (*Store, error) {
	database, err := sqlx.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}
	database.SetMaxOpenConns(cfg.DBMaxOpenConns)
	database.SetMaxIdleConns(cfg.DBMaxIdleConns)
	database.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Store{db: database}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) FetchGlobalStats(ctx context.Context) (GlobalStats, error) {
	stats := GlobalStats{}

	type countersAgg struct {
		TotalMoves       int64 `db:"total_moves"`
		RegisteredUsers  int64 `db:"registered_users"`
		TotalGeokrety    int64 `db:"total_geokrety"`
		PicturesUploaded int64 `db:"pictures_uploaded"`
		PicturesType0    int64 `db:"pictures_type_0"`
		PicturesType1    int64 `db:"pictures_type_1"`
		PicturesType2    int64 `db:"pictures_type_2"`
		GeokretyType0    int64 `db:"geokrety_type_0"`
		GeokretyType1    int64 `db:"geokrety_type_1"`
		GeokretyType2    int64 `db:"geokrety_type_2"`
		GeokretyType3    int64 `db:"geokrety_type_3"`
		GeokretyType4    int64 `db:"geokrety_type_4"`
		GeokretyType5    int64 `db:"geokrety_type_5"`
		GeokretyType6    int64 `db:"geokrety_type_6"`
		GeokretyType7    int64 `db:"geokrety_type_7"`
		GeokretyType8    int64 `db:"geokrety_type_8"`
		GeokretyType9    int64 `db:"geokrety_type_9"`
		GeokretyType10   int64 `db:"geokrety_type_10"`
		MovesType0       int64 `db:"moves_type_0"`
		MovesType1       int64 `db:"moves_type_1"`
		MovesType2       int64 `db:"moves_type_2"`
		MovesType3       int64 `db:"moves_type_3"`
		MovesType4       int64 `db:"moves_type_4"`
		MovesType5       int64 `db:"moves_type_5"`
	}

	agg := countersAgg{}
	if err := s.db.GetContext(ctx, &agg, `
SELECT
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_moves'), 0) AS total_moves,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_users'), 0) AS registered_users,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety'), 0) AS total_geokrety,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_pictures'), 0) AS pictures_uploaded,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_pictures_type_0'), 0) AS pictures_type_0,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_pictures_type_1'), 0) AS pictures_type_1,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_pictures_type_2'), 0) AS pictures_type_2,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_0'), 0) AS geokrety_type_0,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_1'), 0) AS geokrety_type_1,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_2'), 0) AS geokrety_type_2,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_3'), 0) AS geokrety_type_3,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_4'), 0) AS geokrety_type_4,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_5'), 0) AS geokrety_type_5,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_6'), 0) AS geokrety_type_6,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_7'), 0) AS geokrety_type_7,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_8'), 0) AS geokrety_type_8,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_9'), 0) AS geokrety_type_9,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_geokrety_type_10'), 0) AS geokrety_type_10,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_moves_type_0'), 0) AS moves_type_0,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_moves_type_1'), 0) AS moves_type_1,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_moves_type_2'), 0) AS moves_type_2,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_moves_type_3'), 0) AS moves_type_3,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_moves_type_4'), 0) AS moves_type_4,
COALESCE(SUM(cnt) FILTER (WHERE entity = 'gk_moves_type_5'), 0) AS moves_type_5
FROM stats.entity_counters_shard
`); err != nil {
		return GlobalStats{}, fmt.Errorf("query aggregated counters: %w", err)
	}

	stats.TotalMoves = agg.TotalMoves
	stats.RegisteredUsers = agg.RegisteredUsers
	stats.TotalGeokrety = agg.TotalGeokrety
	stats.PicturesUploaded = agg.PicturesUploaded
	stats.PicturesByType.GeokretAvatars = agg.PicturesType0
	stats.PicturesByType.GeokretMoves = agg.PicturesType1
	stats.PicturesByType.UserAvatars = agg.PicturesType2
	stats.GeokretyByType.Traditional = agg.GeokretyType0
	stats.GeokretyByType.Book = agg.GeokretyType1
	stats.GeokretyByType.Human = agg.GeokretyType2
	stats.GeokretyByType.Coin = agg.GeokretyType3
	stats.GeokretyByType.Kretypost = agg.GeokretyType4
	stats.GeokretyByType.Pebble = agg.GeokretyType5
	stats.GeokretyByType.Car = agg.GeokretyType6
	stats.GeokretyByType.Playingcard = agg.GeokretyType7
	stats.GeokretyByType.Dogtag = agg.GeokretyType8
	stats.GeokretyByType.Jigsaw = agg.GeokretyType9
	stats.GeokretyByType.Easteregg = agg.GeokretyType10
	stats.MovesByType.Dropped = agg.MovesType0
	stats.MovesByType.Grabbed = agg.MovesType1
	stats.MovesByType.Commented = agg.MovesType2
	stats.MovesByType.Seen = agg.MovesType3
	stats.MovesByType.Archived = agg.MovesType4
	stats.MovesByType.Dipped = agg.MovesType5

	if err := s.db.GetContext(ctx, &stats.ActiveUsers, `
SELECT COALESCE(COUNT(*), 0)
FROM stats.daily_active_users
WHERE activity_date = CURRENT_DATE
`); err != nil {
		return GlobalStats{}, fmt.Errorf("query active users: %w", err)
	}

	if err := s.db.GetContext(ctx, &stats.MovesLast30Days, `
SELECT COALESCE(COUNT(*), 0)
FROM geokrety.gk_moves
WHERE moved_on_datetime >= NOW() - INTERVAL '30 days'
`); err != nil {
		return GlobalStats{}, fmt.Errorf("query moves last 30 days: %w", err)
	}

	if err := s.db.GetContext(ctx, &stats.ActiveUsersLast30d, `
SELECT COALESCE(COUNT(DISTINCT author), 0)
FROM geokrety.gk_moves
WHERE moved_on_datetime >= NOW() - INTERVAL '30 days'
AND author IS NOT NULL
`); err != nil {
		return GlobalStats{}, fmt.Errorf("query active users last 30d: %w", err)
	}

	if err := s.db.GetContext(ctx, &stats.CountriesReached, `
SELECT COALESCE(COUNT(DISTINCT country_code), 0)
FROM stats.country_daily_stats
WHERE moves_count > 0
`); err != nil {
		return GlobalStats{}, fmt.Errorf("query countries reached: %w", err)
	}

	if err := s.db.GetContext(ctx, &stats.TotalGeokretyHidden, `
SELECT COALESCE(COUNT(*), 0)
FROM geokrety.gk_geokrety_in_caches
`); err != nil {
		return GlobalStats{}, fmt.Errorf("query hidden geokrety: %w", err)
	}

	return stats, nil
}

func (s *Store) FetchRecentMoves(ctx context.Context, limit, offset int) ([]RecentMove, error) {
	rows := []RecentMove{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
m.id,
COALESCE(g.name, 'Unknown GeoKret') AS geokret_name,
CASE m.move_type
WHEN 0 THEN 'dropped'
WHEN 1 THEN 'grabbed'
WHEN 2 THEN 'commented'
WHEN 3 THEN 'seen'
WHEN 4 THEN 'archived'
WHEN 5 THEN 'dipped'
ELSE 'unknown'
END AS type,
COALESCE(u.username, m.username, 'unknown') AS username,
COALESCE(UPPER(m.country), '') AS country,
m.moved_on_datetime AS timestamp
FROM geokrety.gk_moves AS m
LEFT JOIN geokrety.gk_geokrety AS g ON g.id = m.geokret
LEFT JOIN geokrety.gk_users AS u ON u.id = m.author
ORDER BY m.moved_on_datetime DESC, m.id DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent activity: %w", err)
	}

	for i := range rows {
		rows[i].CountryFlag = countryFlag(rows[i].Country)
	}

	return rows, nil
}

func (s *Store) FetchLeaderboard(ctx context.Context, limit, offset int) ([]LeaderboardUser, error) {
	rows := []LeaderboardUser{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
u.id AS user_id,
u.username,
COUNT(*)::bigint AS points,
COUNT(*)::bigint AS moves_count
FROM geokrety.gk_moves AS m
INNER JOIN geokrety.gk_users AS u ON u.id = m.author
GROUP BY u.id, u.username
ORDER BY points DESC, u.username ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}

	colors := []string{
		"bg-emerald-500",
		"bg-cyan-500",
		"bg-sky-500",
		"bg-amber-500",
		"bg-rose-500",
		"bg-teal-500",
		"bg-orange-500",
		"bg-indigo-500",
		"bg-lime-500",
	}
	for i := range rows {
		rows[i].Rank = offset + i + 1
		rows[i].Initials = usernameInitials(rows[i].Username)
		rows[i].AvatarColor = colors[i%len(colors)]
	}

	return rows, nil
}

func (s *Store) FetchCountries(ctx context.Context, limit, offset int) ([]CountryStats, error) {
	rows := []CountryStats{}
	if err := s.db.SelectContext(ctx, &rows, `
WITH country_agg AS (
SELECT
UPPER(country_code) AS code,
SUM(moves_count)::bigint AS moves_count,
SUM(unique_users)::bigint AS active_users,
SUM(drops)::bigint AS dropped,
SUM(dips)::bigint AS dipped,
SUM(sees)::bigint AS seen,
SUM(points_contributed)::double precision AS points_sum_moves
FROM stats.country_daily_stats
GROUP BY UPPER(country_code)
), country_names AS (
SELECT
UPPER(original) AS code,
MIN(country) AS name
FROM geokrety.gk_waypoints_country
WHERE LENGTH(original) = 2
AND country IS NOT NULL
GROUP BY UPPER(original)
), users_home AS (
SELECT
UPPER(home_country) AS code,
COUNT(*)::bigint AS users_home
FROM geokrety.gk_users
WHERE home_country IS NOT NULL
GROUP BY UPPER(home_country)
), country_loves AS (
SELECT
				UPPER(d.country) AS code,
				COALESCE(SUM(g.loves_count), 0)::bigint AS loves
			FROM geokrety.gk_geokrety_with_details AS d
			INNER JOIN geokrety.gk_geokrety AS g ON g.id = d.id
WHERE country IS NOT NULL
			GROUP BY UPPER(d.country)
), country_pictures AS (
SELECT
UPPER(country) AS code,
COALESCE(SUM(pictures_count), 0)::bigint AS pictures
FROM geokrety.gk_moves
WHERE country IS NOT NULL
GROUP BY UPPER(country)
), country_presence AS (
SELECT
UPPER(country) AS code,
COUNT(*) FILTER (WHERE missing = FALSE)::bigint AS geokrety_in_cache,
COUNT(*) FILTER (WHERE missing = TRUE)::bigint AS geokrety_lost
FROM geokrety.gk_geokrety_in_caches
WHERE country IS NOT NULL
GROUP BY UPPER(country)
)
SELECT
ca.code,
COALESCE(cn.name, ca.code) AS name,
ca.moves_count,
COALESCE(uh.users_home, 0) AS users_home,
ca.active_users,
ca.dropped,
ca.dipped,
ca.seen,
COALESCE(cl.loves, 0) AS loves,
COALESCE(cp.pictures, 0) AS pictures,
ca.points_sum_moves AS points_sum,
ca.points_sum_moves,
COALESCE(pr.geokrety_in_cache, 0) AS geokrety_in_cache,
COALESCE(pr.geokrety_lost, 0) AS geokrety_lost,
CASE
WHEN ca.moves_count > 0 THEN ROUND((ca.points_sum_moves / ca.moves_count)::numeric, 2)::double precision
ELSE 0
END AS avg_points_per_move
FROM country_agg AS ca
LEFT JOIN country_names AS cn ON cn.code = ca.code
LEFT JOIN users_home AS uh ON uh.code = ca.code
LEFT JOIN country_loves AS cl ON cl.code = ca.code
LEFT JOIN country_pictures AS cp ON cp.code = ca.code
LEFT JOIN country_presence AS pr ON pr.code = ca.code
ORDER BY ca.moves_count DESC, ca.code ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query countries: %w", err)
	}

	for i := range rows {
		rows[i].Code = strings.ToUpper(rows[i].Code)
		rows[i].Flag = countryFlag(rows[i].Code)
	}

	return rows, nil
}

func (s *Store) FetchRecentBorn(ctx context.Context, limit, offset int) ([]RecentBorn, error) {
	rows := []RecentBorn{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
g.id,
g.gkid,
g.name,
g.type,
g.born_on_datetime AS born_at,
g.owner AS owner_id,
COALESCE(u.username, 'Abandoned') AS owner_name
FROM geokrety.gk_geokrety AS g
LEFT JOIN geokrety.gk_users AS u ON u.id = g.owner
ORDER BY g.born_on_datetime DESC, g.id DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent born: %w", err)
	}
	for i := range rows {
		rows[i].TypeName = geokrety.TypeName(rows[i].Type)
	}
	return rows, nil
}

func (s *Store) FetchRecentLoved(ctx context.Context, limit, offset int) ([]RecentLoved, error) {
	rows := []RecentLoved{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
l.geokret AS geokret_id,
g.gkid AS gkid,
g.name AS geokret_name,
l.user AS user_id,
COALESCE(u.username, 'unknown') AS username,
l.created_on_datetime AS loved_at
FROM geokrety.gk_loves AS l
INNER JOIN geokrety.gk_geokrety AS g ON g.id = l.geokret
LEFT JOIN geokrety.gk_users AS u ON u.id = l.user
ORDER BY l.created_on_datetime DESC, l.id DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent loved: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchRecentWatched(ctx context.Context, limit, offset int) ([]RecentWatched, error) {
	rows := []RecentWatched{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
w.geokret AS geokret_id,
g.gkid AS gkid,
g.name AS geokret_name,
w.user AS user_id,
COALESCE(u.username, 'unknown') AS username,
w.created_on_datetime AS watched_at
FROM geokrety.gk_watched AS w
INNER JOIN geokrety.gk_geokrety AS g ON g.id = w.geokret
LEFT JOIN geokrety.gk_users AS u ON u.id = w.user
ORDER BY w.created_on_datetime DESC, w.id DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent watched: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchRecentActiveCountries(ctx context.Context, limit, offset int) ([]ActiveCountry, error) {
	rows := []ActiveCountry{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
UPPER(country) AS code,
COUNT(*)::bigint AS moves,
COUNT(DISTINCT author)::bigint AS unique_users,
MAX(moved_on_datetime) AS last_activity_at
FROM geokrety.gk_moves
WHERE country IS NOT NULL
AND moved_on_datetime >= NOW() - INTERVAL '30 days'
GROUP BY UPPER(country)
ORDER BY last_activity_at DESC, moves DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent active countries: %w", err)
	}

	for i := range rows {
		rows[i].Flag = countryFlag(rows[i].Code)
	}

	return rows, nil
}

func (s *Store) FetchRecentActiveWaypoints(ctx context.Context, limit, offset int) ([]ActiveWaypoint, error) {
	rows := []ActiveWaypoint{}
	if err := s.db.SelectContext(ctx, &rows, `
		WITH grouped AS (
			SELECT
				UPPER(m.waypoint) AS waypoint,
				COUNT(*)::bigint AS moves,
				MAX(m.moved_on_datetime) AS last_activity_at
			FROM geokrety.gk_moves AS m
			WHERE m.waypoint IS NOT NULL
				AND BTRIM(m.waypoint) <> ''
				AND m.moved_on_datetime >= NOW() - INTERVAL '30 days'
			GROUP BY UPPER(m.waypoint)
		), latest_move AS (
			SELECT DISTINCT ON (UPPER(m.waypoint))
				UPPER(m.waypoint) AS waypoint,
				UPPER(m.country) AS country,
				m.lat,
				m.lon,
				m.moved_on_datetime
			FROM geokrety.gk_moves AS m
			WHERE m.waypoint IS NOT NULL
				AND BTRIM(m.waypoint) <> ''
				AND m.moved_on_datetime >= NOW() - INTERVAL '30 days'
			ORDER BY UPPER(m.waypoint), m.moved_on_datetime DESC, m.id DESC
		)
SELECT
			g.waypoint,
			lm.country,
			g.moves,
			g.last_activity_at,
			lm.lat,
			lm.lon
		FROM grouped AS g
		LEFT JOIN latest_move AS lm ON lm.waypoint = g.waypoint
		ORDER BY g.last_activity_at DESC, g.moves DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent active waypoints: %w", err)
	}

	for i := range rows {
		if rows[i].Lat != nil && rows[i].Lon != nil {
			rows[i].GeoJSON = &GeoJSONPt{
				Type:        "Point",
				Coordinates: []float64{*rows[i].Lon, *rows[i].Lat},
			}
		}
	}

	return rows, nil
}

func (s *Store) FetchRecentRegisteredUsers(ctx context.Context, limit, offset int) ([]RecentRegisteredUser, error) {
	rows := []RecentRegisteredUser{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
id,
username,
joined_on_datetime AS joined_at,
UPPER(home_country) AS home_country
FROM geokrety.gk_users
ORDER BY joined_on_datetime DESC, id DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent registered users: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchRecentActiveUsers(ctx context.Context, limit, offset int) ([]RecentActiveUser, error) {
	rows := []RecentActiveUser{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
u.id AS user_id,
u.username,
COUNT(*)::bigint AS moves_count,
COUNT(DISTINCT UPPER(m.country))::bigint AS countries_count,
MAX(m.moved_on_datetime) AS last_move_at
FROM geokrety.gk_moves AS m
INNER JOIN geokrety.gk_users AS u ON u.id = m.author
WHERE m.moved_on_datetime >= NOW() - INTERVAL '30 days'
GROUP BY u.id, u.username
ORDER BY last_move_at DESC, moves_count DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query recent active users: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchHourlyHeatmap(ctx context.Context, limit, offset int) ([]HourlyHeatmapCell, error) {
	rows := []HourlyHeatmapCell{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	activity_date,
	hour_utc,
	move_type,
	move_count
FROM stats.v_uc8_seasonal_heatmap
ORDER BY activity_date DESC, hour_utc DESC, move_type ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query hourly heatmap: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchCountryFlows(ctx context.Context, limit, offset int) ([]CountryFlow, error) {
	rows := []CountryFlow{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	year_month,
	from_country,
	to_country,
	move_count,
	unique_gk_count
FROM stats.v_uc7_country_flow
ORDER BY year_month DESC, move_count DESC, from_country ASC, to_country ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query country flows: %w", err)
	}
	for i := range rows {
		rows[i].FromCountry = strings.ToUpper(rows[i].FromCountry)
		rows[i].ToCountry = strings.ToUpper(rows[i].ToCountry)
		rows[i].FromFlag = countryFlag(rows[i].FromCountry)
		rows[i].ToFlag = countryFlag(rows[i].ToCountry)
	}
	return rows, nil
}

func (s *Store) FetchTopCaches(ctx context.Context, limit, offset int) ([]TopCache, error) {
	rows := []TopCache{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	waypoint_code,
	total_gk_visits,
	distinct_gks
FROM stats.v_uc10_cache_popularity
ORDER BY total_gk_visits DESC, distinct_gks DESC, waypoint_code ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query top caches: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchFirstFinderLeaderboard(ctx context.Context, limit, offset int) ([]FirstFinderLeaderboardEntry, error) {
	rows := []FirstFinderLeaderboardEntry{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	v.finder_user_id AS user_id,
	COALESCE(u.username, 'unknown') AS username,
	v.first_finds
FROM stats.v_uc14_first_finder_hof AS v
LEFT JOIN geokrety.gk_users AS u ON u.id = v.finder_user_id
ORDER BY v.first_finds DESC, v.finder_user_id ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query first finder leaderboard: %w", err)
	}
	for i := range rows {
		rows[i].Rank = offset + i + 1
	}
	return rows, nil
}

func (s *Store) FetchDistanceRecords(ctx context.Context, limit, offset int) ([]DistanceRecord, error) {
	rows := []DistanceRecord{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	v.gk_id,
	g.gkid,
	COALESCE(g.name, 'Unknown GeoKret') AS name,
	v.km_total
FROM stats.v_uc15_distance_records AS v
LEFT JOIN geokrety.gk_geokrety AS g ON g.id = v.gk_id
ORDER BY v.km_total DESC NULLS LAST, v.gk_id ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query distance records: %w", err)
	}
	for i := range rows {
		rows[i].Rank = offset + i + 1
	}
	return rows, nil
}

func (s *Store) FetchUserNetwork(ctx context.Context, userID int64, limit, offset int) ([]UserNetworkEdge, error) {
	rows := []UserNetworkEdge{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	v.user_id,
	v.related_user_id,
	COALESCE(u.username, 'unknown') AS related_username,
	v.shared_geokrety_count,
	v.first_seen_at,
	v.last_seen_at
FROM stats.v_uc2_user_network AS v
LEFT JOIN geokrety.gk_users AS u ON u.id = v.related_user_id
WHERE v.user_id = $1
ORDER BY v.shared_geokrety_count DESC, v.related_user_id ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user network: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretTimeline(ctx context.Context, geokretID int64, limit, offset int) ([]GeokretTimelineEvent, error) {
	rows := []GeokretTimelineEvent{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	v.gk_id,
	v.event_type,
	v.occurred_at,
	v.actor_user_id,
	u.username AS actor_username
FROM stats.v_uc13_gk_timeline AS v
LEFT JOIN geokrety.gk_users AS u ON u.id = v.actor_user_id
WHERE v.gk_id = $1
ORDER BY v.occurred_at DESC, v.event_type ASC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret timeline: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretCirculation(ctx context.Context, geokretID int64) (GeokretCirculation, error) {
	row := GeokretCirculation{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	v.geokrety_id,
	g.gkid,
	COALESCE(g.name, 'Unknown GeoKret') AS name,
	v.users,
	v.interactions
FROM stats.v_uc3_gk_circulation AS v
LEFT JOIN geokrety.gk_geokrety AS g ON g.id = v.geokrety_id
WHERE v.geokrety_id = $1
`, geokretID); err != nil {
		return GeokretCirculation{}, fmt.Errorf("query geokret circulation: %w", err)
	}
	if row.Users > 0 {
		row.AvgPerUser = float64(row.Interactions) / float64(row.Users)
	}
	return row, nil
}

func countryFlag(code string) string {
	if len(code) != 2 {
		return ""
	}
	code = strings.ToUpper(code)
	return string([]rune{rune(code[0]) - 'A' + 0x1F1E6, rune(code[1]) - 'A' + 0x1F1E6})
}

func usernameInitials(username string) string {
	parts := strings.FieldsFunc(username, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	if len(parts) == 0 {
		if username == "" {
			return "?"
		}
		if len(username) == 1 {
			return strings.ToUpper(username)
		}
		return strings.ToUpper(username[:2])
	}

	if len(parts) == 1 {
		p := []rune(parts[0])
		if len(p) == 1 {
			return strings.ToUpper(parts[0])
		}
		return strings.ToUpper(string(p[0:2]))
	}

	return strings.ToUpper(string([]rune(parts[0])[0]) + string([]rune(parts[1])[0]))
}
