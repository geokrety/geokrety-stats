package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func (s *Store) FetchGeokretyByGKIDs(ctx context.Context, gkids []int64) ([]GeokretListItem, error) {
	if len(gkids) == 0 {
		return []GeokretListItem{}, nil
	}
	rows := []GeokretListItem{}
	query, args, err := sqlx.In(`
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.avatar AS avatar_id,
	CASE
		WHEN ap.bucket IS NOT NULL AND ap.key IS NOT NULL THEN 'https://minio.geokrety.org/' || ap.bucket || '/' || ap.key
		WHEN ap.filename IS NOT NULL THEN 'https://cdn.geokrety.org/images/obrazki/' || ap.filename
		ELSE NULL
	END AS avatar_url,
	gg.type,
	'https://cdn.geokrety.org/images/icons/types/' || gg.type || '.svg' AS type_icon_url,
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
	gg.loves_count,
	gg.pictures_count,
	gg.caches_count,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_pictures AS ap ON ap.id = gg.avatar
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
LEFT JOIN LATERAL (
	SELECT mc.created_on_datetime
	FROM geokrety.gk_moves_comments AS mc
	WHERE mc.move = g.last_position
		AND mc.type = 1
	ORDER BY mc.created_on_datetime DESC, mc.id DESC
	LIMIT 1
) AS missing_comment ON TRUE
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

func (s *Store) FetchGeokretyMovesByIDs(ctx context.Context, geokretID int64, moveIDs []int64) ([]MoveRecord, error) {
	if len(moveIDs) == 0 {
		return []MoveRecord{}, nil
	}
	rows := []MoveRecord{}
	query, args, err := sqlx.In(`
SELECT
	m.id,
	m.geokret AS geokret_id,
	m.move_type,
	m.author AS author_id,
	u.avatar AS author_avatar_id,
	CASE
		WHEN uap.bucket IS NOT NULL AND uap.key IS NOT NULL THEN 'https://minio.geokrety.org/' || uap.bucket || '/' || uap.key
		WHEN uap.filename IS NOT NULL THEN 'https://cdn.geokrety.org/images/obrazki/' || uap.filename
		ELSE NULL
	END AS author_avatar_url,
	COALESCE(u.username, m.username) AS username,
	UPPER(m.country) AS country,
	UPPER(m.waypoint) AS waypoint,
	m.lat,
	m.lon,
	NULLIF(m.elevation, -32768)::bigint AS elevation,
	m.km_distance::double precision AS km_distance,
	m.moved_on_datetime,
	m.created_on_datetime,
	m.pictures_count,
	m.comments_count,
	m.comment,
	m.comment_hidden
FROM geokrety.gk_moves AS m
LEFT JOIN geokrety.gk_users AS u ON u.id = m.author
LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar
WHERE m.geokret = ?
	AND m.id IN (?)
`, geokretID, moveIDs)
	if err != nil {
		return nil, fmt.Errorf("build geokret move ids query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query geokret moves by ids: %w", err)
	}
	return reorderMovesByID(hydrateMoveRecords(rows), moveIDs), nil
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

func reorderMovesByID(rows []MoveRecord, moveIDs []int64) []MoveRecord {
	byID := make(map[int64]MoveRecord, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	ordered := make([]MoveRecord, 0, len(rows))
	for _, moveID := range moveIDs {
		if row, ok := byID[moveID]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}
