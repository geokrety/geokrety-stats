package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func userBaseSelect() string {
	return fmt.Sprintf(`
SELECT
	u.id,
	u.username,
	u.joined_at,
	u.home_country,
	u.avatar_id,
	%s AS avatar_url,
	last_move.last_move_at
`, pictureURLSQL("uap"))
}

func userBaseFrom(alias string) string {
	return fmt.Sprintf(`
FROM %s AS u
LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar_id
LEFT JOIN LATERAL (
	SELECT m.moved_on_datetime AS last_move_at
	FROM geokrety.gk_moves AS m
	WHERE m.author = u.id
	ORDER BY m.moved_on_datetime DESC, m.id DESC
	LIMIT 1
) AS last_move ON TRUE
`, alias)
}

func (s *Store) FetchUserList(ctx context.Context, limit, offset int) ([]UserSearchResult, error) {
	rows := []UserSearchResult{}
	query := userBaseSelect() + userBaseFrom(fmt.Sprintf(`(
	SELECT
		u.id,
		u.username,
		u.joined_on_datetime AS joined_at,
		UPPER(u.home_country) AS home_country,
		u.avatar AS avatar_id
	FROM geokrety.gk_users AS u
	ORDER BY u.joined_on_datetime DESC, u.id DESC
	LIMIT %d OFFSET %d
)`, limit, offset))
	if err := s.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("query user list: %w", err)
	}
	return hydrateUserRows(rows), nil
}

func (s *Store) FetchUserListByIDs(ctx context.Context, userIDs []int64) ([]UserSearchResult, error) {
	if len(userIDs) == 0 {
		return []UserSearchResult{}, nil
	}
	rows := []UserSearchResult{}
	statement := userBaseSelect() + userBaseFrom(`(
	SELECT
		u.id,
		u.username,
		u.joined_on_datetime AS joined_at,
		UPPER(u.home_country) AS home_country,
		u.avatar AS avatar_id
	FROM geokrety.gk_users AS u
	WHERE u.id IN (?)
)`)
	query, args, err := sqlx.In(statement, userIDs)
	if err != nil {
		return nil, fmt.Errorf("build user ids query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query users by ids: %w", err)
	}
	return reorderUsersByID(hydrateUserRows(rows), userIDs), nil
}

func (s *Store) SearchUsers(ctx context.Context, query string, limit, offset int) ([]UserSearchResult, error) {
	rows := []UserSearchResult{}
	statement := userBaseSelect() + userBaseFrom(`(
	SELECT
		u.id,
		u.username,
		u.joined_on_datetime AS joined_at,
		UPPER(u.home_country) AS home_country,
		u.avatar AS avatar_id
	FROM geokrety.gk_users AS u
	WHERE u.username ILIKE '%' || ? || '%'
	ORDER BY u.joined_on_datetime DESC, u.id DESC
	LIMIT ? OFFSET ?
)`)
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(statement), query, limit, offset); err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	return hydrateUserRows(rows), nil
}

func (s *Store) FetchUserDetails(ctx context.Context, userID int64) (UserDetails, error) {
	row := UserDetails{}
	if err := s.db.GetContext(ctx, &row, fmt.Sprintf(`
SELECT
	u.id,
	u.username,
	u.joined_on_datetime AS joined_at,
	UPPER(u.home_country) AS home_country,
	u.avatar AS avatar_id,
	%s AS avatar_url,
	(
		SELECT MAX(m.moved_on_datetime)
		FROM geokrety.gk_moves AS m
		WHERE m.author = u.id
	) AS last_move_at
FROM geokrety.gk_users AS u
LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar
WHERE u.id = $1
`, pictureURLSQL("uap")), userID); err != nil {
		return UserDetails{}, fmt.Errorf("query user details: %w", err)
	}
	if row.HomeCountry != nil {
		country := strings.ToUpper(*row.HomeCountry)
		row.HomeCountry = &country
		row.HomeCountryFlag = countryFlag(country)
	}
	return row, nil
}

func (s *Store) FetchUserStats(ctx context.Context, userID int64) (UserStats, error) {
	row := UserStats{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	u.id AS user_id,
	(
		SELECT COUNT(*)::bigint FROM geokrety.gk_geokrety AS g WHERE g.owner = u.id
	) AS owned_geokrety_count,
	(
		SELECT COUNT(DISTINCT m.geokret)::bigint FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS found_geokrety_count,
	(
		SELECT COUNT(*)::bigint FROM geokrety.gk_loves AS l WHERE l.user = u.id
	) AS loved_geokrety_count,
	(
		SELECT COUNT(*)::bigint FROM geokrety.gk_watched AS w WHERE w.user = u.id
	) AS watched_geokrety_count,
	u.pictures_count,
	(
		SELECT COUNT(*)::bigint FROM stats.user_countries AS uc WHERE uc.user_id = u.id
	) AS countries_visited_count,
	(
		SELECT COUNT(*)::bigint FROM stats.user_cache_visits AS uv WHERE uv.user_id = u.id
	) AS waypoints_visited_count,
	(
		SELECT COUNT(*)::bigint FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS moves_count,
	(
		SELECT COUNT(DISTINCT m.geokret)::bigint FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS distinct_geokrety_count
FROM geokrety.gk_users AS u
WHERE u.id = $1
`, userID); err != nil {
		return UserStats{}, fmt.Errorf("query user stats: %w", err)
	}
	return row, nil
}

func (s *Store) FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	return s.fetchUserGeokretyList(ctx, `gg.owner = $1`, userID, limit, offset, "query user owned geokrety")
}

func (s *Store) FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := geokretSelectColumns() + geokretBaseFromClause() + `
INNER JOIN geokrety.gk_moves AS authored_move ON authored_move.geokret = g.id
WHERE authored_move.author = $1
ORDER BY g.id, authored_move.moved_on_datetime DESC, authored_move.id DESC
LIMIT $2 OFFSET $3
`
	wrapped := `SELECT DISTINCT ON (id) * FROM (` + query + `) AS found_geokrety ORDER BY id, last_move_at DESC NULLS LAST`
	if err := s.db.SelectContext(ctx, &rows, wrapped, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user found geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := geokretSelectColumns() + geokretBaseFromClause() + `
INNER JOIN geokrety.gk_loves AS l ON l.geokret = g.id
WHERE l.user = $1
ORDER BY l.created_on_datetime DESC, l.id DESC
LIMIT $2 OFFSET $3
`
	if err := s.db.SelectContext(ctx, &rows, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user loved geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := geokretSelectColumns() + geokretBaseFromClause() + `
INNER JOIN geokrety.gk_watched AS w ON w.geokret = g.id
WHERE w.user = $1
ORDER BY w.created_on_datetime DESC, w.id DESC
LIMIT $2 OFFSET $3
`
	if err := s.db.SelectContext(ctx, &rows, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user watched geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]UserCountryVisit, error) {
	rows := []UserCountryVisit{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	UPPER(country_code) AS country_code,
	move_count,
	first_visit,
	last_visit
FROM stats.user_countries
WHERE user_id = $1
ORDER BY last_visit DESC, country_code ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user countries: %w", err)
	}
	for i := range rows {
		rows[i].Flag = countryFlag(rows[i].CountryCode)
	}
	return rows, nil
}

func (s *Store) FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]UserWaypointVisit, error) {
	rows := []UserWaypointVisit{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	w.waypoint_code,
	v.visit_count,
	v.first_visited_at,
	v.last_visited_at,
	UPPER(w.country) AS country,
	w.lat,
	w.lon
FROM stats.user_cache_visits AS v
INNER JOIN stats.waypoints AS w ON w.id = v.waypoint_id
WHERE v.user_id = $1
ORDER BY v.last_visited_at DESC, w.waypoint_code ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user waypoints: %w", err)
	}
	return hydrateWaypointVisits(rows), nil
}

func (s *Store) fetchUserGeokretyList(ctx context.Context, whereClause string, userID int64, limit, offset int, label string) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := geokretSelectColumns() + geokretBaseFromClause() + fmt.Sprintf(`
WHERE %s
ORDER BY g.moved_on_datetime DESC, g.id DESC
LIMIT $2 OFFSET $3
`, whereClause)
	if err := s.db.SelectContext(ctx, &rows, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	return hydrateGeokretListItems(rows), nil
}

func reorderUsersByID(rows []UserSearchResult, userIDs []int64) []UserSearchResult {
	byID := make(map[int64]UserSearchResult, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	ordered := make([]UserSearchResult, 0, len(rows))
	for _, userID := range userIDs {
		if row, ok := byID[userID]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}
