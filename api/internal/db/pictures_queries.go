package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func pictureSelectColumns() string {
	return `
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
`
}

func pictureBaseFromClause() string {
	return `
FROM geokrety.gk_pictures AS p
LEFT JOIN geokrety.gk_users AS u ON u.id = p.author
LEFT JOIN geokrety.gk_geokrety AS gg ON gg.id = p.geokret
`
}

func (s *Store) FetchPictureList(ctx context.Context, filters PictureFilters, limit, offset int) ([]PictureInfo, error) {
	rows := []PictureInfo{}
	conditions := []string{"TRUE"}
	args := make([]any, 0, 5)
	if filters.GeokretID != nil {
		conditions = append(conditions, "p.geokret = ?")
		args = append(args, *filters.GeokretID)
	}
	if filters.MoveID != nil {
		conditions = append(conditions, "p.move = ?")
		args = append(args, *filters.MoveID)
	}
	if filters.UserID != nil {
		conditions = append(conditions, "(p.user = ? OR p.author = ?)")
		args = append(args, *filters.UserID, *filters.UserID)
	}
	query := pictureSelectColumns() + pictureBaseFromClause() + `
WHERE ` + strings.Join(conditions, ` AND `) + `
ORDER BY ` + pictureOrderBy(filters.Sort) + `
LIMIT ? OFFSET ?
`
	args = append(args, limit, offset)
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("query picture list: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchPictureListByIDs(ctx context.Context, pictureIDs []int64) ([]PictureInfo, error) {
	if len(pictureIDs) == 0 {
		return []PictureInfo{}, nil
	}
	rows := []PictureInfo{}
	query, args, err := sqlx.In(pictureSelectColumns()+pictureBaseFromClause()+`
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

func (s *Store) FetchPicture(ctx context.Context, pictureID int64) (PictureInfo, error) {
	row := PictureInfo{}
	query := pictureSelectColumns() + pictureBaseFromClause() + `
WHERE p.id = $1
`
	if err := s.db.GetContext(ctx, &row, query, pictureID); err != nil {
		return PictureInfo{}, fmt.Errorf("query picture details: %w", err)
	}
	return row, nil
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
