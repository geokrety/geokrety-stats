package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func (s *Store) FetchPictureListByIDs(ctx context.Context, pictureIDs []int64) ([]PictureInfo, error) {
	if len(pictureIDs) == 0 {
		return []PictureInfo{}, nil
	}
	rows := []PictureInfo{}
	query, args, err := sqlx.In(`
SELECT
	p.id,
	p.type,
	p.filename,
	p.caption,
	p.key,
	p.geokret AS geokret_id,
	gg.gkid AS geokret_gkid,
	p.move AS move_id,
	p.user AS user_id,
	p.author AS author_id,
	u.username AS author_username,
	p.uploaded_on_datetime,
	p.created_on_datetime
FROM geokrety.gk_pictures AS p
LEFT JOIN geokrety.gk_users AS u ON u.id = p.author
LEFT JOIN geokrety.gk_geokrety AS gg ON gg.id = p.geokret
WHERE p.id IN (?)
`, pictureIDs)
	if err != nil {
		return nil, fmt.Errorf("build picture ids query: %w", err)
	}
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query pictures by ids: %w", err)
	}
	return reorderPicturesByID(rows, pictureIDs), nil
}

func reorderPicturesByID(rows []PictureInfo, pictureIDs []int64) []PictureInfo {
	byID := make(map[int64]PictureInfo, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	ordered := make([]PictureInfo, 0, len(rows))
	for _, pictureID := range pictureIDs {
		if row, ok := byID[pictureID]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}
