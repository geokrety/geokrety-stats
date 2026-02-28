package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// Leaderboard handles GET /api/v1/leaderboard
// Query params:
//   - period: all|today|week|month|3months|year|YYYY-MM (default: all)
//   - page, per_page
//   - sort: rank|points|moves (default: rank)
func (h *Handler) Leaderboard(c *gin.Context) {
	period := c.DefaultQuery("period", "all")
	page, perPage, offset := parsePagination(c)

	var (
		entries []models.LeaderboardEntry
		total   int64
		err     error
	)

	switch {
	case period == "all":
		entries, total, err = h.leaderboardAllTime(c.Request.Context(), offset, perPage)
	case period == "today":
		entries, total, err = h.leaderboardDay(c.Request.Context(), "today", offset, perPage)
	case period == "week":
		entries, total, err = h.leaderboardPeriodDays(c.Request.Context(), 7, offset, perPage)
	case period == "month":
		entries, total, err = h.leaderboardPeriodDays(c.Request.Context(), 30, offset, perPage)
	case period == "3months":
		entries, total, err = h.leaderboardPeriodDays(c.Request.Context(), 90, offset, perPage)
	case period == "year":
		entries, total, err = h.leaderboardPeriodDays(c.Request.Context(), 365, offset, perPage)
	case len(period) == 7 && strings.ContainsRune(period, '-'):
		// YYYY-MM – monthly board
		entries, total, err = h.leaderboardMonth(c.Request.Context(), period, offset, perPage)
	case len(period) == 4:
		// YYYY – yearly board
		if yr, e := strconv.Atoi(period); e == nil {
			entries, total, err = h.leaderboardYear(c.Request.Context(), yr, offset, perPage)
		}
	default:
		entries, total, err = h.leaderboardAllTime(c.Request.Context(), offset, perPage)
	}

	if err != nil {
		errInternal(c, err)
		return
	}

	hasNext := int64(offset+perPage) < total
	ok(c, entries, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
		Period:  period,
	}, buildLinks(c, page, perPage, hasNext))
}

// leaderboardAllTime returns entries from the all-time materialized view.
func (h *Handler) leaderboardAllTime(ctx context.Context, offset, limit int) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_all_time`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ).Scan(&total)

	const q = `
		SELECT user_id, username, home_country, total_points, rank,
		       distinct_gks, total_moves, last_active
		FROM geokrety_stats.mv_leaderboard_all_time
		ORDER BY rank ASC
		LIMIT $1 OFFSET $2`

	rows, err := h.DB.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.HomeCountry,
			&e.TotalPoints, &e.Rank, &e.DistinctGKs, &e.TotalMoves, &e.LastActive); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardDay returns today's leaderboard.
func (h *Handler) leaderboardDay(ctx context.Context, _ string, offset, limit int) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_daily WHERE day = CURRENT_DATE`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ).Scan(&total)

	const q = `
		SELECT user_id, username, points_day, rank
		FROM geokrety_stats.mv_leaderboard_daily
		WHERE day = CURRENT_DATE
		ORDER BY rank ASC
		LIMIT $1 OFFSET $2`

	rows, err := h.DB.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardPeriodDays aggregates points over the last N days directly.
func (h *Handler) leaderboardPeriodDays(ctx context.Context, days int, offset, limit int) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `
		SELECT COUNT(DISTINCT user_id)
		FROM geokrety_stats.user_points_log
		WHERE awarded_at >= NOW() - (interval '1 day' * $1)`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ, days).Scan(&total)

	const q = `
		SELECT u.id, u.username, SUM(pl.points) AS pts,
		       RANK() OVER (ORDER BY SUM(pl.points) DESC) AS rank
		FROM geokrety.gk_users u
		JOIN geokrety_stats.user_points_log pl ON pl.user_id = u.id
		WHERE pl.awarded_at >= NOW() - (interval '1 day' * $1)
		GROUP BY u.id, u.username
		ORDER BY pts DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(ctx, q, days, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardMonth returns the monthly leaderboard for a given YYYY-MM.
func (h *Handler) leaderboardMonth(ctx context.Context, yearMonth string, offset, limit int) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_monthly WHERE year_month = $1`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ, yearMonth).Scan(&total)

	const q = `
		SELECT user_id, username, points_month, rank
		FROM geokrety_stats.mv_leaderboard_monthly
		WHERE year_month = $1
		ORDER BY rank ASC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(ctx, q, yearMonth, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardYear returns the yearly leaderboard for a given year.
func (h *Handler) leaderboardYear(ctx context.Context, year int, offset, limit int) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_yearly WHERE year = $1`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ, year).Scan(&total)

	const q = `
		SELECT user_id, username, points_year, rank
		FROM geokrety_stats.mv_leaderboard_yearly
		WHERE year = $1
		ORDER BY rank ASC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(ctx, q, year, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// TopLeaderboard returns only the top N for WebSocket broadcasts.
func (h *Handler) TopLeaderboard(ctx context.Context, n int) ([]models.LeaderboardEntry, error) {
	entries, _, err := h.leaderboardAllTime(ctx, 0, n)
	return entries, err
}
