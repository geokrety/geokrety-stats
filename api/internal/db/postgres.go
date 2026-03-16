package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB
}

type PictureTypeStats struct {
	GeokretAvatars int64 `json:"geokretAvatars"`
	GeokretMoves   int64 `json:"geokretMoves"`
	UserAvatars    int64 `json:"userAvatars"`
}

type GeokretyTypeStats struct {
	Traditional int64 `json:"traditional"`
	Book        int64 `json:"book"`
	Human       int64 `json:"human"`
	Coin        int64 `json:"coin"`
	Kretypost   int64 `json:"kretypost"`
	Pebble      int64 `json:"pebble"`
	Car         int64 `json:"car"`
	Playingcard int64 `json:"playingcard"`
	Dogtag      int64 `json:"dogtag"`
	Jigsaw      int64 `json:"jigsaw"`
	Easteregg   int64 `json:"easteregg"`
}

type MoveTypeStats struct {
	Dropped   int64 `json:"dropped"`
	Grabbed   int64 `json:"grabbed"`
	Commented int64 `json:"commented"`
	Seen      int64 `json:"seen"`
	Archived  int64 `json:"archived"`
	Dipped    int64 `json:"dipped"`
}

type GlobalStats struct {
	TotalGeokrety       int64             `json:"totalGeokrety"`
	TotalGeokretyHidden int64             `json:"totalGeokretyHidden"`
	TotalMoves          int64             `json:"totalMoves"`
	MovesLast30Days     int64             `json:"movesLast30Days"`
	RegisteredUsers     int64             `json:"registeredUsers"`
	ActiveUsers         int64             `json:"activeUsers"`
	ActiveUsersLast30d  int64             `json:"activeUsersLast30d"`
	CountriesReached    int64             `json:"countriesReached"`
	PicturesUploaded    int64             `json:"picturesUploaded"`
	PicturesByType      PictureTypeStats  `json:"picturesByType"`
	GeokretyByType      GeokretyTypeStats `json:"geokretyByType"`
	MovesByType         MoveTypeStats     `json:"movesByType"`
}

type RecentMove struct {
	ID          int64     `db:"id" json:"id"`
	GeokretName string    `db:"geokret_name" json:"geokretName"`
	Type        string    `db:"type" json:"type"`
	Username    string    `db:"username" json:"username"`
	Country     string    `db:"country" json:"country"`
	CountryFlag string    `json:"countryFlag"`
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
}

type LeaderboardUser struct {
	Rank        int    `json:"rank"`
	UserID      int64  `db:"user_id" json:"userId"`
	Username    string `db:"username" json:"username"`
	Initials    string `json:"initials"`
	Points      int64  `db:"points" json:"points"`
	MovesCount  int64  `db:"moves_count" json:"movesCount"`
	AvatarColor string `json:"avatarColor"`
}

type CountryStats struct {
	Code             string  `db:"code" json:"code"`
	Name             string  `db:"name" json:"name"`
	Flag             string  `json:"flag"`
	MovesCount       int64   `db:"moves_count" json:"movesCount"`
	UsersHome        int64   `db:"users_home" json:"usersHome"`
	ActiveUsers      int64   `db:"active_users" json:"activeUsers"`
	Dropped          int64   `db:"dropped" json:"dropped"`
	Dipped           int64   `db:"dipped" json:"dipped"`
	Seen             int64   `db:"seen" json:"seen"`
	Loves            int64   `db:"loves" json:"loves"`
	Pictures         int64   `db:"pictures" json:"pictures"`
	PointsSum        float64 `db:"points_sum" json:"pointsSum"`
	PointsSumMoves   float64 `db:"points_sum_moves" json:"pointsSumMoves"`
	GeokretyInCache  int64   `db:"geokrety_in_cache" json:"geokretyInCache"`
	GeokretyLost     int64   `db:"geokrety_lost" json:"geokretyLost"`
	AvgPointsPerMove float64 `db:"avg_points_per_move" json:"avgPointsPerMove"`
}

type RecentBorn struct {
	ID        int64     `db:"id" json:"id"`
	GKID      int64     `db:"gkid" json:"gkid"`
	Name      string    `db:"name" json:"name"`
	Type      int16     `db:"type" json:"type"`
	BornAt    time.Time `db:"born_at" json:"bornAt"`
	OwnerID   *int64    `db:"owner_id" json:"ownerId"`
	OwnerName string    `db:"owner_name" json:"ownerName"`
}

type RecentLoved struct {
	GeoKretID   int64     `db:"geokret_id" json:"geokretId"`
	GeoKretName string    `db:"geokret_name" json:"geokretName"`
	UserID      int64     `db:"user_id" json:"userId"`
	Username    string    `db:"username" json:"username"`
	LovedAt     time.Time `db:"loved_at" json:"lovedAt"`
}

type RecentWatched struct {
	GeoKretID   int64     `db:"geokret_id" json:"geokretId"`
	GeoKretName string    `db:"geokret_name" json:"geokretName"`
	UserID      int64     `db:"user_id" json:"userId"`
	Username    string    `db:"username" json:"username"`
	WatchedAt   time.Time `db:"watched_at" json:"watchedAt"`
}

type ActiveCountry struct {
	Code           string    `db:"code" json:"code"`
	Moves          int64     `db:"moves" json:"moves"`
	UniqueUsers    int64     `db:"unique_users" json:"uniqueUsers"`
	LastActivityAt time.Time `db:"last_activity_at" json:"lastActivityAt"`
	Flag           string    `json:"flag"`
}

type ActiveWaypoint struct {
	Waypoint       string     `db:"waypoint" json:"waypoint"`
	Country        *string    `db:"country" json:"country"`
	Moves          int64      `db:"moves" json:"moves"`
	LastActivityAt time.Time  `db:"last_activity_at" json:"lastActivityAt"`
	Lat            *float64   `db:"lat" json:"lat"`
	Lon            *float64   `db:"lon" json:"lon"`
	GeoJSON        *GeoJSONPt `json:"geojson"`
}

type GeoJSONPt struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type RecentRegisteredUser struct {
	ID          int64     `db:"id" json:"id"`
	Username    string    `db:"username" json:"username"`
	JoinedAt    time.Time `db:"joined_at" json:"joinedAt"`
	HomeCountry *string   `db:"home_country" json:"homeCountry"`
}

type RecentActiveUser struct {
	UserID         int64     `db:"user_id" json:"userId"`
	Username       string    `db:"username" json:"username"`
	MovesCount     int64     `db:"moves_count" json:"movesCount"`
	CountriesCount int64     `db:"countries_count" json:"countriesCount"`
	LastMoveAt     time.Time `db:"last_move_at" json:"lastMoveAt"`
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
	return rows, nil
}

func (s *Store) FetchRecentLoved(ctx context.Context, limit, offset int) ([]RecentLoved, error) {
	rows := []RecentLoved{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
l.geokret AS geokret_id,
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
