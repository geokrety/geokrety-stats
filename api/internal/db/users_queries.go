package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func (s *Store) FetchUserListByIDs(ctx context.Context, userIDs []int64) ([]UserSearchResult, error) {
	if len(userIDs) == 0 {
		return []UserSearchResult{}, nil
	}
	rows := []UserSearchResult{}
	query, args, err := sqlx.In(`
WITH selected_users AS MATERIALIZED (
	SELECT
		u.id,
		u.username,
		u.joined_on_datetime AS joined_at,
		UPPER(u.home_country) AS home_country,
		u.avatar AS avatar_id
	FROM geokrety.gk_users AS u
	WHERE u.id IN (?)
)
SELECT
	u.id,
	u.username,
	u.joined_at,
	u.home_country,
	u.avatar_id,
	CASE
		WHEN uap.bucket IS NOT NULL AND uap.key IS NOT NULL THEN 'https://minio.geokrety.org/' || uap.bucket || '/' || uap.key
		WHEN uap.filename IS NOT NULL THEN 'https://cdn.geokrety.org/images/obrazki/' || uap.filename
		ELSE NULL
	END AS avatar_url,
	last_move.last_move_at
FROM selected_users AS u
LEFT JOIN geokrety.gk_pictures AS uap ON uap.id = u.avatar_id
LEFT JOIN LATERAL (
	SELECT
		m.moved_on_datetime AS last_move_at
	FROM geokrety.gk_moves AS m
	WHERE m.author = u.id
	ORDER BY m.moved_on_datetime DESC
	LIMIT 1
) AS last_move ON TRUE
`, userIDs)
	if err != nil {
		return nil, fmt.Errorf("build user ids query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query users by ids: %w", err)
	}
	return reorderUsersByID(rows, userIDs), nil
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
