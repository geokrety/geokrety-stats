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
	GeokretAvatars int64 `json:"geokretAvatars"` // type 0: PICTURE_GEOKRET_AVATAR
	GeokretMoves   int64 `json:"geokretMoves"`   // type 1: PICTURE_GEOKRET_MOVE
	UserAvatars    int64 `json:"userAvatars"`    // type 2: PICTURE_USER_AVATAR
}

type GeokretyTypeStats struct {
	Traditional int64 `json:"traditional"` // type 0
	Book        int64 `json:"book"`        // type 1
	Human       int64 `json:"human"`       // type 2
	Coin        int64 `json:"coin"`        // type 3
	Kretypost   int64 `json:"kretypost"`   // type 4
	Pebble      int64 `json:"pebble"`      // type 5
	Car         int64 `json:"car"`         // type 6
	Playingcard int64 `json:"playingcard"` // type 7
	Dogtag      int64 `json:"dogtag"`      // type 8
	Jigsaw      int64 `json:"jigsaw"`      // type 9
	Easteregg   int64 `json:"easteregg"`   // type 10
}

type MoveTypeStats struct {
	Dropped   int64 `json:"dropped"`   // type 0
	Grabbed   int64 `json:"grabbed"`   // type 1
	Commented int64 `json:"commented"` // type 2
	Seen      int64 `json:"seen"`      // type 3
	Archived  int64 `json:"archived"`  // type 4
	Dipped    int64 `json:"dipped"`    // type 5
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

	// TODO consider creating a view that aggregates these counts for even faster retrieval
	// TODO can we use a single query with conditional aggregation instead of multiple queries?

	// Use sharded counters for fast, exact counts on high-volume tables
	// This avoids expensive table scans and returns results in ~O(16) time

	// Consolidated single-query aggregation over sharded counters for many entities.
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

	// Map aggregated results into existing stats structure
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
		SELECT COALESCE(active_users, 0)
		FROM stats.daily_activity
		ORDER BY activity_date DESC
		LIMIT 1
	`); err != nil {
		return GlobalStats{}, fmt.Errorf("query active users: %w", err)
	}

	// Moves in last 30 days (direct query with indexed moved_on_datetime)
	if err := s.db.GetContext(ctx, &stats.MovesLast30Days, `
		SELECT COALESCE(COUNT(*), 0)
		FROM geokrety.gk_moves
		WHERE moved_on_datetime >= NOW() - INTERVAL '30 days'
	`); err != nil {
		return GlobalStats{}, fmt.Errorf("query moves last 30 days: %w", err)
	}

	// Active users in last 30 days (users who created moves)
	if err := s.db.GetContext(ctx, &stats.ActiveUsersLast30d, `
		SELECT COALESCE(COUNT(DISTINCT author), 0)
		FROM geokrety.gk_moves
		WHERE moved_on_datetime >= NOW() - INTERVAL '30 days'
	`); err != nil {
		return GlobalStats{}, fmt.Errorf("query active users last 30d: %w", err)
	}

	if err := s.db.GetContext(ctx, &stats.CountriesReached, `
		SELECT COALESCE(COUNT(*), 0)
		FROM stats.country_stats
		WHERE total_moves > 0
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

func (s *Store) FetchRecentMoves(ctx context.Context, limit int) ([]RecentMove, error) {
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
				ELSE 'commented'
			END AS type,
			COALESCE(u.username, m.username, 'unknown') AS username,
			COALESCE(UPPER(m.country), '') AS country,
			m.moved_on_datetime AS timestamp
		FROM geokrety.gk_moves AS m
		LEFT JOIN geokrety.gk_geokrety AS g ON g.id = m.geokret
		LEFT JOIN geokrety.gk_users AS u ON u.id = m.author
		ORDER BY m.moved_on_datetime DESC
		LIMIT $1
	`, limit); err != nil {
		return nil, fmt.Errorf("query recent activity: %w", err)
	}

	for i := range rows {
		rows[i].CountryFlag = countryFlag(rows[i].Country)
	}

	return rows, nil
}

func (s *Store) FetchLeaderboard(ctx context.Context, limit int) ([]LeaderboardUser, error) {
	rows := []LeaderboardUser{}
	if err := s.db.SelectContext(ctx, &rows, `
		SELECT
			COALESCE(u.username, 'unknown') AS username,
			COALESCE(SUM(upd.points), 0)::bigint AS points,
			COALESCE(SUM(upd.moves_count), 0)::bigint AS moves_count
		FROM stats.user_points_daily AS upd
		INNER JOIN geokrety.gk_users AS u ON u.id = upd.user_id
		GROUP BY u.username
		ORDER BY points DESC, moves_count DESC
		LIMIT $1
	`, limit); err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}

	colors := []string{
		"bg-emerald-500",
		"bg-cyan-500",
		"bg-violet-500",
		"bg-sky-500",
		"bg-amber-500",
		"bg-rose-500",
		"bg-teal-500",
		"bg-orange-500",
		"bg-indigo-500",
		"bg-lime-500",
	}
	for i := range rows {
		rows[i].Rank = i + 1
		rows[i].Initials = usernameInitials(rows[i].Username)
		rows[i].AvatarColor = colors[i%len(colors)]
	}

	return rows, nil
}

func (s *Store) FetchCountries(ctx context.Context, limit int) ([]CountryStats, error) {
	rows := []CountryStats{}
	if err := s.db.SelectContext(ctx, &rows, `
		WITH country_names AS (
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
				COUNT(*) AS users_home
			FROM geokrety.gk_users
			WHERE home_country IS NOT NULL
			GROUP BY UPPER(home_country)
		), country_points_users AS (
			SELECT
				UPPER(u.home_country) AS code,
				COALESCE(SUM(upd.points), 0) AS points_sum
			FROM geokrety.gk_users AS u
			INNER JOIN stats.user_points_daily AS upd ON upd.user_id = u.id
			WHERE u.home_country IS NOT NULL
			GROUP BY UPPER(u.home_country)
		), country_pictures AS (
			SELECT
				UPPER(country) AS code,
				COALESCE(SUM(pictures_count), 0) AS pictures
			FROM geokrety.gk_moves
			WHERE country IS NOT NULL
			GROUP BY UPPER(country)
		), country_loves AS (
			SELECT
				cc.country_code AS code,
				COALESCE(SUM(g.loves_count), 0) AS loves,
				COUNT(*) FILTER (WHERE g.missing = FALSE) AS geokrety_in_cache,
				COUNT(*) FILTER (WHERE g.missing = TRUE) AS geokrety_lost
			FROM stats.gk_current_country AS cc
			INNER JOIN geokrety.gk_geokrety AS g ON g.id = cc.geokrety_id
			GROUP BY cc.country_code
		)
		SELECT
			cs.country_code AS code,
			COALESCE(cn.name, cs.country_code) AS name,
			cs.total_moves AS moves_count,
			COALESCE(uh.users_home, 0) AS users_home,
			cs.unique_users AS active_users,
			cs.drops AS dropped,
			cs.dips AS dipped,
			cs.sees AS seen,
			COALESCE(cl.loves, 0) AS loves,
			COALESCE(cp.pictures, 0) AS pictures,
			COALESCE(cpu.points_sum, 0) AS points_sum,
			cs.total_points_awarded AS points_sum_moves,
			COALESCE(cl.geokrety_in_cache, 0) AS geokrety_in_cache,
			COALESCE(cl.geokrety_lost, 0) AS geokrety_lost,
			CASE
				WHEN cs.total_moves > 0 THEN ROUND((cs.total_points_awarded / cs.total_moves::double precision)::numeric, 2)::double precision
				ELSE 0
			END AS avg_points_per_move
		FROM stats.country_stats AS cs
		LEFT JOIN country_names AS cn ON cn.code = cs.country_code
		LEFT JOIN users_home AS uh ON uh.code = cs.country_code
		LEFT JOIN country_points_users AS cpu ON cpu.code = cs.country_code
		LEFT JOIN country_pictures AS cp ON cp.code = cs.country_code
		LEFT JOIN country_loves AS cl ON cl.code = cs.country_code
		ORDER BY cs.total_points_awarded DESC, cs.total_moves DESC
		LIMIT $1
	`, limit); err != nil {
		return nil, fmt.Errorf("query countries: %w", err)
	}

	for i := range rows {
		rows[i].Code = strings.ToUpper(rows[i].Code)
		rows[i].Flag = countryFlag(rows[i].Code)
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
