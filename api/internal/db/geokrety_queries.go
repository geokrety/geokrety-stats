package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func geokretSelectColumns() string {
	return fmt.Sprintf(`
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.avatar AS avatar_id,
	%s AS avatar_url,
	gg.type,
	gg.missing,
	CASE WHEN gg.missing THEN missing_comment.created_on_datetime END AS missing_at,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at,
	lm.move_type AS last_move_type,
	g.last_position AS last_position_id,
	g.last_log AS last_log_id,
	gg.mission,
	g.non_collectible AS non_collectible_at,
	g.parked AS parked_at,
	gg.comments_hidden
`, pictureURLSQL("ap"))
}

func geokretBaseFromClause() string {
	return `
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_pictures AS ap ON ap.id = gg.avatar
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
LEFT JOIN geokrety.gk_moves AS lm ON lm.id = g.last_position
LEFT JOIN LATERAL (
	SELECT mc.created_on_datetime
	FROM geokrety.gk_moves_comments AS mc
	WHERE mc.move = g.last_position
		AND mc.type = 1
	ORDER BY mc.created_on_datetime DESC, mc.id DESC
	LIMIT 1
) AS missing_comment ON TRUE
`
}

func (s *Store) FetchGeokretyList(ctx context.Context, filters GeokretListFilters, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	conditions := []string{"TRUE"}
	args := make([]any, 0, 6)
	if filters.Name != nil {
		conditions = append(conditions, "g.name ILIKE ?")
		args = append(args, "%"+strings.TrimSpace(*filters.Name)+"%")
	}
	if filters.OwnerID != nil {
		conditions = append(conditions, "gg.owner = ?")
		args = append(args, *filters.OwnerID)
	}
	if len(filters.Countries) > 0 {
		conditions = append(conditions, "UPPER(g.country) IN (?)")
		args = append(args, filters.Countries)
	}
	query := geokretSelectColumns() + geokretBaseFromClause() + `
WHERE ` + strings.Join(conditions, ` AND `) + `
ORDER BY ` + geokretOrderBy(filters.Sort) + `
LIMIT ? OFFSET ?
`
	args = append(args, limit, offset)
	query, expandedArgs, err := sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("build geokret list query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), expandedArgs...); err != nil {
		return nil, fmt.Errorf("query geokret list: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchGeokretyByGKIDs(ctx context.Context, gkids []int64) ([]GeokretListItem, error) {
	if len(gkids) == 0 {
		return []GeokretListItem{}, nil
	}
	rows := []GeokretListItem{}
	query, args, err := sqlx.In(geokretSelectColumns()+geokretBaseFromClause()+`
WHERE g.gkid IN (?)
`, gkids)
	if err != nil {
		return nil, fmt.Errorf("build geokret ids query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query geokrety by ids: %w", err)
	}
	return reorderGeokretyByGKID(hydrateGeokretListItems(rows), gkids), nil
}

func (s *Store) FetchGeokretyByGKID(ctx context.Context, gkid int64) (GeokretDetails, error) {
	return s.fetchGeokretDetails(ctx, "g.gkid = $1", gkid, "query geokret details by gkid")
}

func (s *Store) ResolveGeokretID(ctx context.Context, gkid int64) (int64, error) {
	var geokretID int64
	if err := s.db.GetContext(ctx, &geokretID, `
SELECT id
FROM geokrety.gk_geokrety
WHERE gkid = $1
`, gkid); err != nil {
		return 0, fmt.Errorf("resolve geokret id: %w", err)
	}
	return geokretID, nil
}

func (s *Store) FetchGeokretStats(ctx context.Context, geokretID int64) (GeokretStats, error) {
	row := GeokretStats{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	g.id AS geokret_id,
	g.gkid,
	g.caches_count,
	g.pictures_count,
	(
		SELECT COUNT(*)::bigint
		FROM geokrety.gk_moves AS m
		WHERE m.geokret = g.id
	) AS moves_count,
	(
		SELECT COUNT(*)::bigint
		FROM stats.gk_countries_visited AS v
		WHERE v.geokrety_id = g.id
	) AS countries_visited_count,
	(
		SELECT COUNT(*)::bigint
		FROM stats.gk_cache_visits AS v
		WHERE v.gk_id = g.id
	) AS waypoints_visited_count,
	(
		SELECT COUNT(DISTINCT m.author)::bigint
		FROM geokrety.gk_moves AS m
		WHERE m.geokret = g.id
			AND m.author IS NOT NULL
			AND NOT(m.move_type = 4
				OR (m.move_type = 3 AND m.position IS NOT NULL)
			)
	) AS finders_count,
	(
		SELECT COUNT(*)::bigint
		FROM geokrety.gk_watched AS w
		WHERE w.geokret = g.id
	) AS watchers_count,
	(
		SELECT COUNT(*)::bigint
		FROM geokrety.gk_loves AS l
		WHERE l.geokret = g.id
	) AS lovers_count,
	(
		SELECT CASE
			WHEN m.country IS NULL OR BTRIM(m.country) = '' THEN NULL
			ELSE UPPER(BTRIM(m.country))
		END
		FROM geokrety.gk_moves AS m
		WHERE m.id = g.last_position
	) AS current_country_code,
	(
		SELECT CASE
			WHEN m.waypoint IS NULL OR BTRIM(m.waypoint) = '' THEN NULL
			ELSE UPPER(BTRIM(m.waypoint))
		END
		FROM geokrety.gk_moves AS m
		WHERE m.id = g.last_position
	) AS current_waypoint_code
FROM geokrety.gk_geokrety AS g
WHERE g.id = $1
`, geokretID); err != nil {
		return GeokretStats{}, fmt.Errorf("query geokret stats: %w", err)
	}
	if row.CurrentCountryCode != nil {
		country := strings.ToUpper(*row.CurrentCountryCode)
		row.CurrentCountryCode = &country
	}
	if row.CurrentWaypointCode != nil {
		waypoint := strings.ToUpper(*row.CurrentWaypointCode)
		row.CurrentWaypointCode = &waypoint
	}
	return row, nil
}

func (s *Store) FetchGeokretyLoves(ctx context.Context, geokretID int64, sort Sort, limit, offset int) ([]SocialUserEntry, error) {
	rows := []SocialUserEntry{}
	if err := s.db.SelectContext(ctx, &rows, fmt.Sprintf(`
SELECT
	l.user AS user_id,
	COALESCE(u.username, 'unknown') AS username,
	u.avatar AS avatar_id,
	%s AS avatar_url,
	l.created_on_datetime AS at
FROM geokrety.gk_loves AS l
LEFT JOIN geokrety.gk_users AS u ON u.id = l.user
LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar
WHERE l.geokret = $1
ORDER BY %s
LIMIT $2 OFFSET $3
`, pictureURLSQL("uap"), socialOrderBy(sort)), geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret loves: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyWatches(ctx context.Context, geokretID int64, sort Sort, limit, offset int) ([]SocialUserEntry, error) {
	rows := []SocialUserEntry{}
	if err := s.db.SelectContext(ctx, &rows, fmt.Sprintf(`
SELECT
	w.user AS user_id,
	COALESCE(u.username, 'unknown') AS username,
	u.avatar AS avatar_id,
	%s AS avatar_url,
	w.created_on_datetime AS at
FROM geokrety.gk_watched AS w
LEFT JOIN geokrety.gk_users AS u ON u.id = w.user
LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar
WHERE w.geokret = $1
ORDER BY %s
LIMIT $2 OFFSET $3
`, pictureURLSQL("uap"), socialOrderBy(sort)), geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret watches: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyFinders(ctx context.Context, geokretID int64, sort Sort, limit, offset int) ([]SocialUserEntry, error) {
	rows := []SocialUserEntry{}
	if err := s.db.SelectContext(ctx, &rows, fmt.Sprintf(`
WITH latest_finders AS (
	SELECT DISTINCT ON (m.author)
		m.author AS user_id,
		COALESCE(u.username, 'unknown') AS username,
		u.avatar AS avatar_id,
		%s AS avatar_url,
		m.moved_on_datetime AS at
	FROM geokrety.gk_moves AS m
	LEFT JOIN geokrety.gk_users AS u ON u.id = m.author
	LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar
	WHERE m.geokret = $1
		AND m.author IS NOT NULL
	ORDER BY m.author, m.moved_on_datetime DESC, m.id DESC
)
SELECT user_id, username, avatar_id, avatar_url, at
FROM latest_finders
ORDER BY %s
LIMIT $2 OFFSET $3
`, pictureURLSQL("uap"), socialOrderBy(sort)), geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret finders: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyCountries(ctx context.Context, geokretID int64, sort Sort, limit, offset int) ([]GeokretCountryVisit, error) {
	rows := []GeokretCountryVisit{}
	query := `
SELECT
	UPPER(country_code) AS country_code,
	first_visited_at,
	move_count::bigint AS move_count
FROM stats.gk_countries_visited
WHERE geokrety_id = $1
ORDER BY ` + geokretCountryVisitOrderBy(sort) + `
LIMIT $2 OFFSET $3
`
	if err := s.db.SelectContext(ctx, &rows, query, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret countries: %w", err)
	}
	for i := range rows {
		rows[i].Flag = countryFlag(rows[i].CountryCode)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyWaypoints(ctx context.Context, geokretID int64, sort Sort, limit, offset int) ([]GeokretWaypointVisit, error) {
	rows := []GeokretWaypointVisit{}
	query := `
SELECT
	w.waypoint_code,
	UPPER(w.country) AS country,
	v.visit_count,
	v.first_visited_at,
	v.last_visited_at,
	w.lat,
	w.lon
FROM stats.gk_cache_visits AS v
INNER JOIN stats.waypoints AS w ON w.id = v.waypoint_id
WHERE v.gk_id = $1
ORDER BY ` + waypointVisitOrderBy(sort) + `
LIMIT $2 OFFSET $3
`
	if err := s.db.SelectContext(ctx, &rows, query, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret waypoints: %w", err)
	}
	return hydrateGKWaypointVisits(rows), nil
}

func (s *Store) fetchGeokretDetails(ctx context.Context, predicate string, value any, label string) (GeokretDetails, error) {
	row := GeokretDetails{}
	query := geokretSelectColumns() + geokretBaseFromClause() + fmt.Sprintf(`
WHERE %s
`, predicate)
	if err := s.db.GetContext(ctx, &row, query, value); err != nil {
		return GeokretDetails{}, fmt.Errorf("%s: %w", label, err)
	}
	row.GeokretListItem = hydrateGeokretListItems([]GeokretListItem{row.GeokretListItem})[0]
	return row, nil
}

func reorderGeokretyByGKID(rows []GeokretListItem, gkids []int64) []GeokretListItem {
	byID := make(map[int64]GeokretListItem, len(rows))
	for _, row := range rows {
		if row.GKID == nil {
			continue
		}
		byID[row.GKID.Int()] = row
	}
	ordered := make([]GeokretListItem, 0, len(rows))
	for _, gkid := range gkids {
		if row, ok := byID[gkid]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}
