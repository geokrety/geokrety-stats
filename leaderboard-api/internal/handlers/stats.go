package handlers

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// GlobalStats handles GET /api/v1/stats
func (h *Handler) GlobalStats(c *gin.Context) {
	const q = `
		SELECT total_users, total_gks, total_moves, scored_users,
		       total_points_awarded, countries_reached, total_km, computed_at
		FROM geokrety_stats.mv_global_stats`

	var s models.GlobalStats
	err := h.DB.QueryRow(c.Request.Context(), q).Scan(
		&s.TotalUsers, &s.TotalGKs, &s.TotalMoves, &s.ScoredUsers,
		&s.TotalPointsAwarded, &s.CountriesReached, &s.TotalKm, &s.ComputedAt,
	)
	if err != nil {
		// fallback to direct count
		s = h.computeGlobalStatsFallback(c.Request.Context())
	} else {
		// Always compute images and loves (not in view)
		_ = h.DB.QueryRow(c.Request.Context(),
			`SELECT COALESCE(COUNT(*), 0) FROM geokrety.gk_images`).Scan(&s.TotalImages)
		_ = h.DB.QueryRow(c.Request.Context(),
			`SELECT COALESCE(SUM(loves_count), 0) FROM geokrety.gk_geokrety`).Scan(&s.TotalLoves)
	}
	ok(c, s, models.Meta{}, nil)
}

// computeGlobalStatsFallback computes stats directly (slower).
func (h *Handler) computeGlobalStatsFallback(ctx context.Context) models.GlobalStats {
	var s models.GlobalStats
	_ = h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM geokrety.gk_users`).Scan(&s.TotalUsers)
	_ = h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM geokrety.gk_geokrety`).Scan(&s.TotalGKs)
	_ = h.DB.QueryRow(ctx, `SELECT COUNT(*) FROM geokrety.gk_moves`).Scan(&s.TotalMoves)
	_ = h.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM geokrety_stats.user_points_totals WHERE total_points > 0`).Scan(&s.ScoredUsers)
	_ = h.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_points),0) FROM geokrety_stats.user_points_totals`).Scan(&s.TotalPointsAwarded)
	_ = h.DB.QueryRow(ctx,
		`SELECT COUNT(DISTINCT country) FROM geokrety.gk_moves WHERE country IS NOT NULL`).Scan(&s.CountriesReached)
	_ = h.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(distance),0) FROM geokrety.gk_geokrety`).Scan(&s.TotalKm)
	_ = h.DB.QueryRow(ctx,
		`SELECT COALESCE(COUNT(*), 0) FROM geokrety.gk_images`).Scan(&s.TotalImages)
	_ = h.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(loves_count), 0) FROM geokrety.gk_geokrety`).Scan(&s.TotalLoves)
	return s
}

// DailyActivity handles GET /api/v1/stats/activity/daily
// Returns global move counts per day for the last N days (default 90)
func (h *Handler) DailyActivity(c *gin.Context) {
	days := 90
	if d := c.Query("days"); d != "" {
		if v := parseInt(d, 90); v > 0 && v <= 365 {
			days = v
		}
	}

	const q = `
		SELECT activity_date::text AS day,
		       total_moves AS moves,
		       active_users,
		       active_gks,
		       drops, grabs, dips, comments, seen
		FROM geokrety_stats.mv_daily_activity
		WHERE activity_date >= CURRENT_DATE - (interval '1 day' * $1)
		ORDER BY activity_date DESC`

	rows, err := h.DB.Query(c.Request.Context(), q, days)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Day         string `json:"day"`
		Moves       int64  `json:"total_moves"`
		ActiveUsers int64  `json:"active_users"`
		ActiveGKs   int64  `json:"active_gks"`
		Drops       int64  `json:"drops"`
		Grabs       int64  `json:"grabs"`
		Dips        int64  `json:"dips"`
		Comments    int64  `json:"comments"`
		Seen        int64  `json:"seen"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Day, &r.Moves, &r.ActiveUsers, &r.ActiveGKs,
			&r.Drops, &r.Grabs, &r.Dips, &r.Comments, &r.Seen); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// TopCountries handles GET /api/v1/stats/countries
func (h *Handler) TopCountries(c *gin.Context) {
	const q = `
		SELECT cs.country, cs.total_moves, cs.unique_gks, cs.unique_users,
		       cs.total_points_awarded, cs.drops, cs.grabs, cs.dips, cs.comments, cs.seen,
		       CASE WHEN cs.total_moves > 0 THEN ROUND(cs.total_points_awarded::NUMERIC / cs.total_moves, 2) ELSE 0 END AS avg_points_per_move,
		       COALESCE(lv.total_loves, 0) AS total_loves
		FROM geokrety_stats.mv_country_summary cs
		LEFT JOIN (
		    SELECT dm.country, SUM(g.loves_count) AS total_loves
		    FROM (SELECT DISTINCT country, geokret FROM geokrety.gk_moves WHERE country IS NOT NULL) dm
		    JOIN geokrety.gk_geokrety g ON g.id = dm.geokret
		    GROUP BY dm.country
		) lv ON lv.country = cs.country
		ORDER BY cs.total_points_awarded DESC
		LIMIT 50`

	rows, err := h.DB.Query(c.Request.Context(), q)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Country              string  `json:"country"`
		TotalMoves           int64   `json:"total_moves"`
		UniqueGks            int64   `json:"unique_gks"`
		UniqueUsers          int64   `json:"unique_users"`
		TotalPointsAwarded   float64 `json:"total_points_awarded"`
		Drops                int64   `json:"drops"`
		Grabs                int64   `json:"grabs"`
		Dips                 int64   `json:"dips"`
		Comments             int64   `json:"comments"`
		Seen                 int64   `json:"seen"`
		AvgPointsPerMove     float64 `json:"avg_points_per_move"`
		TotalLoves           int64   `json:"total_loves"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Country, &r.TotalMoves, &r.UniqueGks, &r.UniqueUsers,
			&r.TotalPointsAwarded, &r.Drops, &r.Grabs, &r.Dips, &r.Comments, &r.Seen,
			&r.AvgPointsPerMove, &r.TotalLoves); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// PointsBreakdownGlobal handles GET /api/v1/stats/points/breakdown
func (h *Handler) PointsBreakdownGlobal(c *gin.Context) {
	const q = `
		SELECT label, SUM(points) AS pts, COUNT(*) AS cnt, COUNT(DISTINCT user_id) AS users
		FROM geokrety_stats.user_points_log
		GROUP BY label
		ORDER BY pts DESC`

	rows, err := h.DB.Query(c.Request.Context(), q)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Label  string  `json:"label"`
		Points float64 `json:"points"`
		Count  int64   `json:"count"`
		Users  int64   `json:"users"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Label, &r.Points, &r.Count, &r.Users); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// AvailablePeriods handles GET /api/v1/stats/periods
// Returns available months/years that have data in the leaderboard.
func (h *Handler) AvailablePeriods(c *gin.Context) {
	const monthQ = `
		SELECT DISTINCT year_month
		FROM geokrety_stats.mv_leaderboard_monthly
		ORDER BY year_month DESC
		LIMIT 60`

	rows, err := h.DB.Query(c.Request.Context(), monthQ)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	months := []string{}
	for rows.Next() {
		var m string
		_ = rows.Scan(&m)
		months = append(months, m)
	}

	const yearQ = `
		SELECT DISTINCT year
		FROM geokrety_stats.mv_leaderboard_yearly
		ORDER BY year DESC`

	rowsY, err := h.DB.Query(c.Request.Context(), yearQ)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rowsY.Close()

	years := []int{}
	for rowsY.Next() {
		var y int
		_ = rowsY.Scan(&y)
		years = append(years, y)
	}

	ok(c, gin.H{"months": months, "years": years}, models.Meta{}, nil)
}

// parseInt parses an int with a default.
func parseInt(s string, def int) int {
	v := def
	if s != "" {
		if n, err := parseInt64(s); err == nil {
			v = int(n)
		}
	}
	return v
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, errBadInt
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

var errBadInt = errorf("not an integer")

func errorf(s string) error {
	return &strError{s}
}

type strError struct{ s string }

func (e *strError) Error() string { return e.s }
