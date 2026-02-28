package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// GlobalStats handles GET /api/v1/stats
func (h *Handler) GlobalStats(c *gin.Context) {
	const q = `
		SELECT total_users, total_gks, total_moves, scored_users,
		       total_points_awarded, countries_reached, total_km,
		       total_images, total_loves, computed_at
		FROM geokrety_stats.mv_global_stats`

	var s models.GlobalStats
	err := h.DB.QueryRow(c.Request.Context(), q).Scan(
		&s.TotalUsers, &s.TotalGKs, &s.TotalMoves, &s.ScoredUsers,
		&s.TotalPointsAwarded, &s.CountriesReached, &s.TotalKm,
		&s.TotalImages, &s.TotalLoves, &s.ComputedAt,
	)
	if err != nil {
		// fallback to direct count
		s = h.computeGlobalStatsFallback(c.Request.Context())
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
		`SELECT COUNT(*) FROM geokrety.gk_pictures`).Scan(&s.TotalImages)
	_ = h.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(loves_count), 0) FROM geokrety.gk_geokrety`).Scan(&s.TotalLoves)
	now := time.Now()
	s.ComputedAt = &now
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
		       drops, grabs, dips, comments, sees
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
		Sees        int64  `json:"seen"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Day, &r.Moves, &r.ActiveUsers, &r.ActiveGKs,
			&r.Drops, &r.Grabs, &r.Dips, &r.Comments, &r.Sees); err != nil {
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
		       cs.drops, cs.grabs, cs.dips, cs.comments, cs.seen,
		       cs.total_points_awarded,
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
		Drops                int64   `json:"drops"`
		Grabs                int64   `json:"grabs"`
		Dips                 int64   `json:"dips"`
		Comments             int64   `json:"comments"`
		Seen                 int64   `json:"seen"`
		TotalPointsAwarded   float64 `json:"total_points_awarded"`
		AvgPointsPerMove     float64 `json:"avg_points_per_move"`
		TotalLoves           int64   `json:"total_loves"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Country, &r.TotalMoves, &r.UniqueGks, &r.UniqueUsers,
			&r.Drops, &r.Grabs, &r.Dips, &r.Comments, &r.Seen,
			&r.TotalPointsAwarded, &r.AvgPointsPerMove, &r.TotalLoves); err != nil {
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

// UserEvolution handles GET /api/v1/stats/evolution/users
// Returns monthly cumulative user count since 2007
func (h *Handler) UserEvolution(c *gin.Context) {
	const q = `
		WITH monthly_counts AS (
			SELECT
				DATE_TRUNC('month', joined_on_datetime)::DATE AS month,
				COUNT(*) AS created_count
			FROM gk_users
			WHERE joined_on_datetime >= '2007-01-01'
			GROUP BY 1
		)
		SELECT
			month,
			SUM(created_count) OVER (ORDER BY month) AS cumulative_users
		FROM monthly_counts
		ORDER BY month ASC
	`

	rows, err := h.DB.Query(c.Request.Context(), q)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Month           string `json:"month"`
		CumulativeUsers int64  `json:"total_users"`
	}
	var out []row
	for rows.Next() {
		var r row
		var t time.Time
		if err := rows.Scan(&t, &r.CumulativeUsers); err != nil {
			errInternal(c, err)
			return
		}
		r.Month = t.Format("2006-01-02")
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// GeoKretyEvolution handles GET /api/v1/stats/evolution/geokrety
// Returns monthly creation count and in-cache count of geokrety since 2007
func (h *Handler) GeoKretyEvolution(c *gin.Context) {
	const q = `
		WITH monthly_counts AS (
			SELECT
				DATE_TRUNC('month', created_on_datetime)::DATE AS month,
				COUNT(*) AS created_count
			FROM gk_geokrety
			WHERE created_on_datetime >= '2007-01-01'
			GROUP BY 1
		)
		SELECT
			month,
			created_count,
			SUM(created_count) OVER (ORDER BY month) AS cumulative_gks
		FROM monthly_counts
		ORDER BY month ASC
	`

	rows, err := h.DB.Query(c.Request.Context(), q)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Month        string `json:"month"`
		CreatedCount int64  `json:"created_count"`
		CumulativeGk int64  `json:"total_gks"`
	}
	var out []row
	for rows.Next() {
		var r row
		var t time.Time
		if err := rows.Scan(&t, &r.CreatedCount, &r.CumulativeGk); err != nil {
			errInternal(c, err)
			return
		}
		r.Month = t.Format("2006-01-02")
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// CountryMoveTypeEvolution handles GET /api/v1/stats/countries/:country/evolution/move-types
// Returns monthly move types count for a specific country since 2007
func (h *Handler) CountryMoveTypeEvolution(c *gin.Context) {
	country := c.Param("country")

	const q = `
		WITH effective_moves AS (
			SELECT
				DATE_TRUNC('month', moved_on_datetime)::DATE AS month,
				move_type,
				COALESCE(country, LAG(country) OVER (PARTITION BY geokret ORDER BY id)) as effective_country
			FROM geokrety.gk_moves
			WHERE moved_on_datetime >= '2007-01-01'
		)
		SELECT
			month,
			COUNT(*) FILTER (WHERE move_type = 0) AS drops,
			COUNT(*) FILTER (WHERE move_type = 1) AS grabs,
			COUNT(*) FILTER (WHERE move_type = 5) AS dips,
			COUNT(*) FILTER (WHERE move_type = 3) AS seen,
			COUNT(*) AS total_moves
		FROM effective_moves
		WHERE effective_country = $1
		GROUP BY 1
		ORDER BY month ASC
	`

	rows, err := h.DB.Query(c.Request.Context(), q, strings.ToUpper(country))
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Month      string `json:"month"`
		Drops      int64  `json:"drops"`
		Grabs      int64  `json:"grabs"`
		Dips       int64  `json:"dips"`
		Seen       int64  `json:"seen"`
		TotalMoves int64  `json:"total_moves"`
	}
	var out []row
	for rows.Next() {
		var r row
		var t time.Time
		if err := rows.Scan(&t, &r.Drops, &r.Grabs, &r.Dips, &r.Seen, &r.TotalMoves); err != nil {
			errInternal(c, err)
			return
		}
		r.Month = t.Format("2006-01-02")
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
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
