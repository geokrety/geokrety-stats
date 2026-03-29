package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func moveSelectColumns() string {
	return fmt.Sprintf(`
SELECT
	m.id,
	m.geokret AS geokret_id,
	g.gkid AS geokret_gkid,
	m.move_type,
	m.author AS author_id,
	u.avatar AS author_avatar_id,
	%s AS author_avatar_url,
	COALESCE(u.username, NULLIF(m.username, ''), 'unknown') AS username,
	UPPER(m.country) AS country,
	UPPER(m.waypoint) AS waypoint,
	m.lat,
	m.lon,
	NULLIF(m.elevation, -32768)::bigint AS elevation,
	m.km_distance::double precision AS km_distance,
	m.moved_on_datetime,
	m.created_on_datetime,
	m.comment,
	m.comment_hidden,
	m.previous_move_id,
	m.previous_position_id
`, pictureURLSQL("uap"))
}

func moveBaseFromClause() string {
	return `
FROM geokrety.gk_moves AS m
INNER JOIN geokrety.gk_geokrety AS g ON g.id = m.geokret
LEFT JOIN geokrety.gk_users AS u ON u.id = m.author
LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar
`
}

func (s *Store) FetchMoveList(ctx context.Context, filters MoveFilters, limit, offset int) ([]MoveRecord, error) {
	rows := []MoveRecord{}
	conditions := []string{"TRUE"}
	args := make([]any, 0, 8)
	if filters.GeokretID != nil {
		conditions = append(conditions, "m.geokret = ?")
		args = append(args, *filters.GeokretID)
	}
	if filters.UserID != nil {
		conditions = append(conditions, "m.author = ?")
		args = append(args, *filters.UserID)
	}
	if filters.Country != nil {
		conditions = append(conditions, "UPPER(m.country) = UPPER(?)")
		args = append(args, *filters.Country)
	}
	if filters.Waypoint != nil {
		conditions = append(conditions, "UPPER(m.waypoint) = UPPER(?)")
		args = append(args, *filters.Waypoint)
	}
	if filters.DateFrom != nil {
		conditions = append(conditions, "m.moved_on_datetime >= ?")
		args = append(args, filters.DateFrom.UTC())
	}
	if filters.DateTo != nil {
		conditions = append(conditions, "m.moved_on_datetime < ?")
		args = append(args, filters.DateTo.UTC())
	}
	query := moveSelectColumns() + moveBaseFromClause() + `
WHERE ` + strings.Join(conditions, ` AND `) + `
ORDER BY ` + moveOrderBy(filters.Sort) + `
LIMIT ? OFFSET ?
`
	args = append(args, limit, offset)
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query moves: %w", err)
	}
	return hydrateMoveRecords(rows), nil
}

func (s *Store) FetchMoveListByIDs(ctx context.Context, filters MoveFilters, moveIDs []int64) ([]MoveRecord, error) {
	if len(moveIDs) == 0 {
		return []MoveRecord{}, nil
	}
	rows := []MoveRecord{}
	conditions := []string{"m.id IN (?)"}
	args := []any{moveIDs}
	if filters.GeokretID != nil {
		conditions = append(conditions, "m.geokret = ?")
		args = append(args, *filters.GeokretID)
	}
	if filters.UserID != nil {
		conditions = append(conditions, "m.author = ?")
		args = append(args, *filters.UserID)
	}
	if filters.Country != nil {
		conditions = append(conditions, "UPPER(m.country) = UPPER(?)")
		args = append(args, *filters.Country)
	}
	if filters.Waypoint != nil {
		conditions = append(conditions, "UPPER(m.waypoint) = UPPER(?)")
		args = append(args, *filters.Waypoint)
	}
	if filters.DateFrom != nil {
		conditions = append(conditions, "m.moved_on_datetime >= ?")
		args = append(args, filters.DateFrom.UTC())
	}
	if filters.DateTo != nil {
		conditions = append(conditions, "m.moved_on_datetime < ?")
		args = append(args, filters.DateTo.UTC())
	}
	query, expandedArgs, err := sqlx.In(moveSelectColumns()+moveBaseFromClause()+`
WHERE `+strings.Join(conditions, ` AND `)+`
`, args...)
	if err != nil {
		return nil, fmt.Errorf("build move ids query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), expandedArgs...); err != nil {
		return nil, fmt.Errorf("query moves by ids: %w", err)
	}
	return reorderMovesByID(hydrateMoveRecords(rows), moveIDs), nil
}

func (s *Store) FetchMove(ctx context.Context, moveID int64) (MoveRecord, error) {
	row := MoveRecord{}
	query := moveSelectColumns() + moveBaseFromClause() + `
WHERE m.id = $1
`
	if err := s.db.GetContext(ctx, &row, query, moveID); err != nil {
		return MoveRecord{}, fmt.Errorf("query move details: %w", err)
	}
	return hydrateMoveRecords([]MoveRecord{row})[0], nil
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
