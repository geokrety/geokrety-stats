package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// Leaderboard handles GET /api/v1/leaderboard
// Query params:
//   - period: all|today|week|month|3months|year|YYYY-MM (default: all)
//   - page, per_page
//   - sort: points|moves|gks|countries (default: points)
func (h *Handler) Leaderboard(c *gin.Context) {
	period := c.DefaultQuery("period", "all")
	sort := c.DefaultQuery("sort", "points")
	page, perPage, offset := parsePagination(c)

	var (
		entries []models.LeaderboardEntry
		total   int64
		err     error
	)

	switch {
	case period == "all":
		entries, total, err = h.leaderboardAllTime(c.Request.Context(), offset, perPage, sort)
	case period == "today":
		entries, total, err = h.leaderboardDay(c.Request.Context(), "today", offset, perPage, sort)
	case period == "week":
		entries, total, err = h.leaderboardPeriodDays(c.Request.Context(), 7, offset, perPage, sort)
	case period == "month":
		entries, total, err = h.leaderboardPeriodDays(c.Request.Context(), 30, offset, perPage, sort)
	case period == "3months":
		entries, total, err = h.leaderboardPeriodDays(c.Request.Context(), 90, offset, perPage, sort)
	case period == "year":
		// Use actual current year from yearly materialized view
		currentYear := time.Now().Year()
		entries, total, err = h.leaderboardYear(c.Request.Context(), currentYear, offset, perPage, sort)
	case len(period) == 7 && strings.ContainsRune(period, '-'):
		// YYYY-MM – monthly board
		entries, total, err = h.leaderboardMonth(c.Request.Context(), period, offset, perPage, sort)
	case len(period) == 4:
		// YYYY – yearly board
		if yr, e := strconv.Atoi(period); e == nil {
			entries, total, err = h.leaderboardYear(c.Request.Context(), yr, offset, perPage, sort)
		}
	default:
		entries, total, err = h.leaderboardAllTime(c.Request.Context(), offset, perPage, sort)
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
func (h *Handler) leaderboardAllTime(ctx context.Context, offset, limit int, sort string) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_all_time`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ).Scan(&total)

	var orderBy string
	switch sort {
	case "moves":
		orderBy = "l.total_moves DESC, l.rank ASC"
	case "gks":
		orderBy = "l.distinct_gks DESC, l.rank ASC"
	case "countries":
		orderBy = "countries_count DESC, l.rank ASC"
	case "avg_points":
		orderBy = "(l.total_points / NULLIF(l.total_moves, 0)) DESC, l.rank ASC"
	default:
		orderBy = "l.rank ASC"
	}

	q := `
		SELECT l.user_id, l.username, l.home_country, l.total_points, l.rank,
		       l.distinct_gks, l.total_moves, l.last_active,
		       COALESCE(s.countries_visited_count, 0) AS countries_count
		FROM geokrety_stats.mv_leaderboard_all_time l
		LEFT JOIN geokrety_stats.mv_user_stats s ON s.user_id = l.user_id
		ORDER BY ` + orderBy + `
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
			&e.TotalPoints, &e.Rank, &e.GkCount, &e.MoveCount, &e.LastActive, &e.CountriesCount); err != nil {
			return nil, 0, err
		}
		if e.MoveCount > 0 {
			e.AvgPointsPerMove = e.TotalPoints / float64(e.MoveCount)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardDay returns today's leaderboard.
func (h *Handler) leaderboardDay(ctx context.Context, _ string, offset, limit int, sort string) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_daily WHERE day = CURRENT_DATE`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ).Scan(&total)

	var orderBy string
	switch sort {
	case "moves":
		orderBy = "moves DESC, l.rank ASC"
	case "gks":
		orderBy = "gks DESC, l.rank ASC"
	case "countries":
		orderBy = "countries DESC, l.rank ASC"
	case "avg_points":
		orderBy = "(l.points_day / NULLIF(COALESCE(s.total_moves, 0), 0)) DESC, l.rank ASC"
	default:
		orderBy = "l.rank ASC"
	}

	q := `
		SELECT l.user_id, l.username, l.points_day, l.rank,
		       COALESCE(s.total_moves, 0) AS moves,
		       COALESCE(s.distinct_gks_interacted, 0) AS gks,
		       COALESCE(s.countries_visited_count, 0) AS countries
		FROM geokrety_stats.mv_leaderboard_daily l
		LEFT JOIN geokrety_stats.mv_user_stats s ON s.user_id = l.user_id
		WHERE l.day = CURRENT_DATE
		ORDER BY ` + orderBy + `
		LIMIT $1 OFFSET $2`

	rows, err := h.DB.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank, &e.MoveCount, &e.GkCount, &e.CountriesCount); err != nil {
			return nil, 0, err
		}
		if e.MoveCount > 0 {
			e.AvgPointsPerMove = e.PointsPeriod / float64(e.MoveCount)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardPeriodDays aggregates points over the last N days directly.
func (h *Handler) leaderboardPeriodDays(ctx context.Context, days int, offset, limit int, sort string) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `
		SELECT COUNT(DISTINCT user_id)
		FROM geokrety_stats.user_points_log
		WHERE awarded_at >= NOW() - (interval '1 day' * $1)`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ, days).Scan(&total)

	var orderBy string
	switch sort {
	case "moves":
		orderBy = "moves DESC, pts DESC"
	case "gks":
		orderBy = "gks DESC, pts DESC"
	case "countries":
		orderBy = "countries DESC, pts DESC"
	case "avg_points":
		orderBy = "(SUM(pl.points) / NULLIF(COALESCE(s.total_moves, 0), 0)) DESC, pts DESC"
	default:
		orderBy = "pts DESC"
	}

	q := `
		SELECT u.id, u.username, SUM(pl.points) AS pts,
		       RANK() OVER (ORDER BY SUM(pl.points) DESC) AS rank,
		       COALESCE(s.total_moves, 0) AS moves,
		       COALESCE(s.distinct_gks_interacted, 0) AS gks,
		       COALESCE(s.countries_visited_count, 0) AS countries
		FROM geokrety.gk_users u
		JOIN geokrety_stats.user_points_log pl ON pl.user_id = u.id
		LEFT JOIN geokrety_stats.mv_user_stats s ON s.user_id = u.id
		WHERE pl.awarded_at >= NOW() - (interval '1 day' * $1)
		GROUP BY u.id, u.username, s.total_moves, s.distinct_gks_interacted, s.countries_visited_count
		ORDER BY ` + orderBy + `
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(ctx, q, days, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank, &e.MoveCount, &e.GkCount, &e.CountriesCount); err != nil {
			return nil, 0, err
		}
		if e.MoveCount > 0 {
			e.AvgPointsPerMove = e.PointsPeriod / float64(e.MoveCount)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardMonth returns the monthly leaderboard for a given YYYY-MM.
func (h *Handler) leaderboardMonth(ctx context.Context, yearMonth string, offset, limit int, sort string) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_monthly WHERE year_month = $1`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ, yearMonth).Scan(&total)

	var orderBy string
	switch sort {
	case "moves":
		orderBy = "moves DESC, l.rank ASC"
	case "gks":
		orderBy = "gks DESC, l.rank ASC"
	case "countries":
		orderBy = "countries DESC, l.rank ASC"
	case "avg_points":
		orderBy = "(l.points_month / NULLIF(COALESCE(s.total_moves, 0), 0)) DESC, l.rank ASC"
	default:
		orderBy = "l.rank ASC"
	}

	q := `
		SELECT l.user_id, l.username, l.points_month, l.rank,
		       COALESCE(s.total_moves, 0) AS moves,
		       COALESCE(s.distinct_gks_interacted, 0) AS gks,
		       COALESCE(s.countries_visited_count, 0) AS countries
		FROM geokrety_stats.mv_leaderboard_monthly l
		LEFT JOIN geokrety_stats.mv_user_stats s ON s.user_id = l.user_id
		WHERE l.year_month = $1
		ORDER BY ` + orderBy + `
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(ctx, q, yearMonth, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank, &e.MoveCount, &e.GkCount, &e.CountriesCount); err != nil {
			return nil, 0, err
		}
		if e.MoveCount > 0 {
			e.AvgPointsPerMove = e.PointsPeriod / float64(e.MoveCount)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// leaderboardYear returns the yearly leaderboard for a given year.
func (h *Handler) leaderboardYear(ctx context.Context, year int, offset, limit int, sort string) ([]models.LeaderboardEntry, int64, error) {
	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_leaderboard_yearly WHERE year = $1`
	var total int64
	_ = h.DB.QueryRow(ctx, countQ, year).Scan(&total)

	var orderBy string
	switch sort {
	case "moves":
		orderBy = "moves DESC, l.rank ASC"
	case "gks":
		orderBy = "gks DESC, l.rank ASC"
	case "countries":
		orderBy = "countries DESC, l.rank ASC"
	case "avg_points":
		orderBy = "(l.points_year / NULLIF(COALESCE(s.total_moves, 0), 0)) DESC, l.rank ASC"
	default:
		orderBy = "l.rank ASC"
	}

	q := `
		SELECT l.user_id, l.username, l.points_year, l.rank,
		       COALESCE(s.total_moves, 0) AS moves,
		       COALESCE(s.distinct_gks_interacted, 0) AS gks,
		       COALESCE(s.countries_visited_count, 0) AS countries
		FROM geokrety_stats.mv_leaderboard_yearly l
		LEFT JOIN geokrety_stats.mv_user_stats s ON s.user_id = l.user_id
		WHERE l.year = $1
		ORDER BY ` + orderBy + `
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(ctx, q, year, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.LeaderboardEntry
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.PointsPeriod, &e.Rank, &e.MoveCount, &e.GkCount, &e.CountriesCount); err != nil {
			return nil, 0, err
		}
		if e.MoveCount > 0 {
			e.AvgPointsPerMove = e.PointsPeriod / float64(e.MoveCount)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// TopLeaderboard returns only the top N for WebSocket broadcasts.
func (h *Handler) TopLeaderboard(ctx context.Context, n int) ([]models.LeaderboardEntry, error) {
	entries, _, err := h.leaderboardAllTime(ctx, 0, n, "points")
	return entries, err
}
