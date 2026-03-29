package db

import (
	"context"
	"fmt"
)

func (s *Store) FetchWaypoint(ctx context.Context, waypointCode string) (WaypointDetails, error) {
	row := WaypointDetails{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	w.id,
	w.waypoint_code,
	w.source,
	UPPER(w.country) AS country,
	w.lat,
	w.lon
FROM stats.waypoints AS w
WHERE UPPER(w.waypoint_code) = UPPER($1)
`, waypointCode); err != nil {
		return WaypointDetails{}, fmt.Errorf("query waypoint details: %w", err)
	}
	row.WaypointSummary = hydrateWaypointSummary(row.WaypointSummary)
	return row, nil
}

func (s *Store) FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := geokretSelectColumns() + geokretBaseFromClause() + `
WHERE UPPER(g.waypoint) = UPPER($1)
ORDER BY g.moved_on_datetime DESC, g.id DESC
LIMIT $2 OFFSET $3
`
	if err := s.db.SelectContext(ctx, &rows, query, waypointCode, limit, offset); err != nil {
		return nil, fmt.Errorf("query waypoint current geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := fmt.Sprintf(`
SELECT
	g.id,
	g.gkid,
	g.name,
	g.avatar AS avatar_id,
	%s AS avatar_url,
	g.type,
	g.missing,
	NULL::timestamp with time zone AS missing_at,
	g.owner AS owner_id,
	ou.username AS owner_username,
	g.holder AS holder_id,
	hu.username AS holder_username,
	NULL::varchar AS country,
	$1::varchar AS waypoint,
	NULL::double precision AS lat,
	NULL::double precision AS lon,
	g.born_on_datetime AS born_at,
	g.updated_on_datetime AS last_move_at,
	NULL::smallint AS last_move_type,
	NULL::bigint AS last_position_id,
	NULL::bigint AS last_log_id,
	g.mission,
	g.non_collectible AS non_collectible_at,
	g.parked AS parked_at,
	g.comments_hidden
FROM stats.gk_cache_visits AS gcv
INNER JOIN stats.waypoints AS w ON w.id = gcv.waypoint_id
INNER JOIN geokrety.gk_geokrety AS g ON g.id = gcv.gk_id
LEFT JOIN geokrety.gk_pictures AS ap ON ap.id = g.avatar
LEFT JOIN geokrety.gk_users AS ou ON ou.id = g.owner
LEFT JOIN geokrety.gk_users AS hu ON hu.id = g.holder
WHERE UPPER(w.waypoint_code) = UPPER($1)
ORDER BY gcv.last_visited_at DESC, g.id DESC
LIMIT $2 OFFSET $3
`, pictureURLSQL("ap"))
	if err := s.db.SelectContext(ctx, &rows, query, waypointCode, limit, offset); err != nil {
		return nil, fmt.Errorf("query waypoint past geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]WaypointSummary, error) {
	rows := []WaypointSummary{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	id,
	waypoint_code,
	source,
	UPPER(country) AS country,
	lat,
	lon
FROM stats.waypoints
WHERE waypoint_code ILIKE '%' || $1 || '%'
ORDER BY waypoint_code ASC
LIMIT $2 OFFSET $3
`, query, limit, offset); err != nil {
		return nil, fmt.Errorf("search waypoints: %w", err)
	}
	for i := range rows {
		rows[i] = hydrateWaypointSummary(rows[i])
	}
	return rows, nil
}
