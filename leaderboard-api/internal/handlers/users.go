package handlers

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// GetUser handles GET /api/v1/users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}

	u, err := h.fetchUser(c.Request.Context(), id)
	if err != nil {
		errInternal(c, err)
		return
	}
	if u == nil {
		errNotFound(c, "user not found")
		return
	}
	ok(c, u, models.Meta{}, nil)
}

// ListUsers handles GET /api/v1/users
// Query: page, per_page, sort (points|moves|rank)
func (h *Handler) ListUsers(c *gin.Context) {
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_user_stats`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ).Scan(&total)

	const q = `
		SELECT user_id, username, home_country, total_points, rank_all_time,
		       total_moves, distinct_gks, countries_count, active_days
		FROM geokrety_stats.mv_user_stats
		ORDER BY total_points DESC
		LIMIT $1 OFFSET $2`

	rows, err := h.DB.Query(c.Request.Context(), q, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		UserID       int64   `json:"user_id"`
		Username     string  `json:"username"`
		HomeCountry  *string `json:"home_country,omitempty"`
		TotalPoints  float64 `json:"total_points"`
		Rank         int64   `json:"rank"`
		TotalMoves   int64   `json:"total_moves"`
		DistinctGKs  int64   `json:"distinct_gks"`
		CountriesCount int64 `json:"countries_count"`
		ActiveDays   int64   `json:"active_days"`
	}

	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.UserID, &r.Username, &r.HomeCountry, &r.TotalPoints,
			&r.Rank, &r.TotalMoves, &r.DistinctGKs, &r.CountriesCount, &r.ActiveDays); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		errInternal(c, rows.Err())
		return
	}

	hasNext := int64(offset+perPage) < total
	ok(c, out, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
	}, buildLinks(c, page, perPage, hasNext))
}

// UserPointsTimeline handles GET /api/v1/users/:id/points/timeline
func (h *Handler) UserPointsTimeline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_user_points_daily WHERE user_id = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT day::text, points, moves_count
		FROM geokrety_stats.mv_user_points_daily
		WHERE user_id = $1
		ORDER BY day DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	var out []models.UserDailyPoints
	for rows.Next() {
		var r models.UserDailyPoints
		if err := rows.Scan(&r.Day, &r.Points, &r.MovesCount); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	hasNext := int64(offset+perPage) < total
	ok(c, out, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
	}, buildLinks(c, page, perPage, hasNext))
}

// UserCountries handles GET /api/v1/users/:id/countries
func (h *Handler) UserCountries(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}

	const q = `
		SELECT country, move_count, first_visit, last_visit
		FROM geokrety_stats.mv_user_countries
		WHERE user_id = $1
		ORDER BY move_count DESC`

	rows, err := h.DB.Query(c.Request.Context(), q, id)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	var out []models.UserCountry
	for rows.Next() {
		var r models.UserCountry
		if err := rows.Scan(&r.Country, &r.MoveCount, &r.FirstVisit, &r.LastVisit); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// UserMoves handles GET /api/v1/users/:id/moves
func (h *Handler) UserMoves(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety.gk_moves WHERE author = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT m.id, m.geokret, g.name, m.move_type,
		       m.country, m.waypoint, m.distance, m.moved_on_datetime,
		       COALESCE(p.points, 0)::float8 as points,
		       p.chain_id
		FROM geokrety.gk_moves m
		LEFT JOIN geokrety.gk_geokrety g ON g.id = m.geokret
		LEFT JOIN (
			SELECT move_id, SUM(points) as points, MAX(chain_id) as chain_id
			FROM geokrety_stats.user_points_log
			GROUP BY move_id
		) p ON p.move_id = m.id
		WHERE m.author = $1
		ORDER BY m.moved_on_datetime DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.UserMove, 0)
	for rows.Next() {
		var r models.UserMove
		var gkName *string
		var points float64
		var chainID *int64
		if err := rows.Scan(&r.MoveID, &r.GkID, &gkName, &r.MoveType,
			&r.Country, &r.Waypoint, &r.Distance, &r.MovedOn, &points, &chainID); err != nil {
			errInternal(c, err)
			return
		}
		if gkName != nil {
			r.GkName = *gkName
		}
		if points > 0 {
			r.Points = &points
		}
		r.ChainID = chainID
		r.TypeName = moveTypeName(r.MoveType)
		out = append(out, r)
	}

	hasNext := int64(offset+perPage) < total
	ok(c, out, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
	}, buildLinks(c, page, perPage, hasNext))
}

// UserGeokrety handles GET /api/v1/users/:id/geokrety
func (h *Handler) UserGeokrety(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(DISTINCT geokret) FROM geokrety.gk_moves WHERE author = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT g.id, g.name, g.type, g.missing, g.distance,
		       gs.total_points_generated, gs.current_multiplier,
		       MAX(m.moved_on_datetime) AS last_interaction
		FROM (SELECT DISTINCT geokret FROM geokrety.gk_moves WHERE author = $1) dm
		JOIN geokrety.gk_geokrety g ON g.id = dm.geokret
		LEFT JOIN geokrety_stats.mv_gk_stats gs ON gs.gk_id = g.id
		LEFT JOIN geokrety.gk_moves m ON m.geokret = g.id AND m.author = $1
		GROUP BY g.id, g.name, g.type, g.missing, g.distance,
		         gs.total_points_generated, gs.current_multiplier
		ORDER BY last_interaction DESC NULLS LAST
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		GkID                int64    `json:"gk_id"`
		Name                string   `json:"gk_name"`
		GkType              int      `json:"gk_type"`
		Missing             bool     `json:"missing"`
		Distance            int64    `json:"distance_km"`
		TotalPointsGenerated *float64 `json:"total_points_generated,omitempty"`
		CurrentMultiplier   *float64 `json:"current_multiplier,omitempty"`
		LastInteraction     interface{} `json:"last_interaction,omitempty"`
	}

	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.GkID, &r.Name, &r.GkType, &r.Missing,
			&r.Distance, &r.TotalPointsGenerated, &r.CurrentMultiplier, &r.LastInteraction); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	hasNext := int64(offset+perPage) < total
	ok(c, out, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
	}, buildLinks(c, page, perPage, hasNext))
}

// UserPointsBreakdown handles GET /api/v1/users/:id/points/breakdown
func (h *Handler) UserPointsBreakdown(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}

	const q = `
		SELECT label, SUM(points) AS pts, COUNT(*) AS cnt
		FROM geokrety_stats.user_points_log
		WHERE user_id = $1
		GROUP BY label
		ORDER BY pts DESC`

	rows, err := h.DB.Query(c.Request.Context(), q, id)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Source string  `json:"source"`
		Points float64 `json:"points"`
		Count  int64   `json:"count"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Source, &r.Points, &r.Count); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// UserPointsAwards handles GET /api/v1/users/:id/points/awards
// Returns paginated list of individual point award log entries.
func (h *Handler) UserPointsAwards(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}
	page, perPage, offset := parsePagination(c)
	label := c.Query("label") // optional filter by label
	sort := c.DefaultQuery("sort", "date")
	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}

	var orderBy string
	switch sort {
	case "points":
		orderBy = "points " + dir + ", awarded_at DESC"
	case "label":
		orderBy = "label " + dir + ", awarded_at DESC"
	default:
		orderBy = "awarded_at " + dir
	}

	type awardRow struct {
		ID        int64   `json:"id"`
		Label     string  `json:"label"`
		Reason    string  `json:"reason"`
		Points    float64 `json:"points"`
		MoveID    *int64  `json:"move_id,omitempty"`
		GkID      *int64  `json:"gk_id,omitempty"`
		AwardedAt string  `json:"awarded_at"`
	}

	var total int64
	out := make([]awardRow, 0)

	if label != "" {
		_ = h.DB.QueryRow(c.Request.Context(),
			`SELECT COUNT(*) FROM geokrety_stats.user_points_log WHERE user_id = $1 AND label = $2`,
			id, label).Scan(&total)

		qrows, qerr := h.DB.Query(c.Request.Context(), `
			SELECT id, label, reason, points, move_id, gk_id, awarded_at::text
			FROM geokrety_stats.user_points_log
			WHERE user_id = $1 AND label = $2
			ORDER BY `+orderBy+`
			LIMIT $3 OFFSET $4`, id, label, perPage, offset)
		if qerr != nil {
			errInternal(c, qerr)
			return
		}
		defer qrows.Close()
		for qrows.Next() {
			var r awardRow
			if err2 := qrows.Scan(&r.ID, &r.Label, &r.Reason, &r.Points, &r.MoveID, &r.GkID, &r.AwardedAt); err2 != nil {
				errInternal(c, err2)
				return
			}
			out = append(out, r)
		}
	} else {
		_ = h.DB.QueryRow(c.Request.Context(),
			`SELECT COUNT(*) FROM geokrety_stats.user_points_log WHERE user_id = $1`, id).Scan(&total)

		qrows, qerr := h.DB.Query(c.Request.Context(), `
			SELECT id, label, reason, points, move_id, gk_id, awarded_at::text
			FROM geokrety_stats.user_points_log
			WHERE user_id = $1
			ORDER BY `+orderBy+`
			LIMIT $2 OFFSET $3`, id, perPage, offset)
		if qerr != nil {
			errInternal(c, qerr)
			return
		}
		defer qrows.Close()
		for qrows.Next() {
			var r awardRow
			if err2 := qrows.Scan(&r.ID, &r.Label, &r.Reason, &r.Points, &r.MoveID, &r.GkID, &r.AwardedAt); err2 != nil {
				errInternal(c, err2)
				return
			}
			out = append(out, r)
		}
	}

	hasNext := int64(offset+perPage) < total
	ok(c, out, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
	}, buildLinks(c, page, perPage, hasNext))
}

// UserRankHistory handles GET /api/v1/users/:id/rank/history
func (h *Handler) UserRankHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}

	const q = `
		SELECT year_month, points_month,
		       RANK() OVER (PARTITION BY year_month ORDER BY points_month DESC) AS rank
		FROM geokrety_stats.mv_leaderboard_monthly
		WHERE user_id = $1
		ORDER BY year_month ASC`

	rows, err := h.DB.Query(c.Request.Context(), q, id)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Month  string  `json:"month"`
		Points float64 `json:"points"`
		Rank   int64   `json:"rank"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Month, &r.Points, &r.Rank); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// fetchUser loads a full user stats profile by ID.
// UserRelatedUsers handles GET /api/v1/users/:id/related-users
// Returns other users who have interacted with the same geokrety
func (h *Handler) UserRelatedUsers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_user_related_users WHERE user_id = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT r.related_user_id, u.username, u.home_country,
		       COALESCE(us.total_points, 0) total_points, COALESCE(us.total_moves, 0) total_moves,
		       r.shared_geokrety_count
		FROM geokrety_stats.mv_user_related_users r
		JOIN geokrety.gk_users u ON u.id = r.related_user_id
		LEFT JOIN geokrety_stats.mv_user_stats us ON us.user_id = r.related_user_id
		WHERE r.user_id = $1
		ORDER BY r.shared_geokrety_count DESC, us.total_points DESC NULLS LAST
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		UserID              int64    `json:"user_id"`
		Username            string   `json:"username"`
		HomeCountry         *string  `json:"home_country,omitempty"`
		TotalPoints         float64  `json:"total_points"`
		TotalMoves          int64    `json:"total_moves"`
		SharedGeokretyCount int64    `json:"shared_geokrety_count"`
	}

	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.UserID, &r.Username, &r.HomeCountry, &r.TotalPoints, &r.TotalMoves, &r.SharedGeokretyCount); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		errInternal(c, rows.Err())
		return
	}

	hasNext := int64(offset+perPage) < total
	ok(c, out, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
	}, buildLinks(c, page, perPage, hasNext))
}

func (h *Handler) fetchUser(ctx context.Context, id int64) (*models.User, error) {
	const q = `
		SELECT us.user_id, us.username, us.home_country, us.joined_on_datetime,
		       us.total_points, COALESCE(lb.rank, 0) AS rank, us.total_moves, us.total_drops, us.total_grabs,
		       us.total_comments, us.total_seen, us.total_dips, us.total_archived, us.distinct_gks_interacted,
		       us.distinct_owners, us.countries_visited_count, 0 AS caches_visited,
		       us.km_contributed, us.active_days, us.first_move_at, us.last_move_at,
		       us.pts_base, us.pts_relay, us.pts_rescuer, us.pts_chain, us.pts_country,
		       us.pts_diversity, us.pts_handover, us.pts_reach
		FROM geokrety_stats.mv_user_stats us
		LEFT JOIN geokrety_stats.mv_leaderboard_all_time lb ON lb.user_id = us.user_id
		WHERE us.user_id = $1`

	var u models.User
	err := h.DB.QueryRow(ctx, q, id).Scan(
		&u.UserID, &u.Username, &u.HomeCountry, &u.JoinedAt,
		&u.TotalPoints, &u.RankAllTime, &u.TotalMoves, &u.TotalDrops, &u.TotalGrabs,
		&u.TotalComments, &u.TotalSeen, &u.TotalDips, &u.TotalArchived, &u.DistinctGKs,
		&u.DistinctOwners, &u.CountriesCount, &u.CachesCount,
		&u.KmContributed, &u.ActiveDays, &u.FirstMoveAt, &u.LastMoveAt,
		&u.PtsBase, &u.PtsRelay, &u.PtsRescuer, &u.PtsChain, &u.PtsCountry,
		&u.PtsDiversity, &u.PtsHandover, &u.PtsReach,
	)
	if err != nil {
		// Try fallback: user exists in gk_users but may not be in mv_user_stats
		const fallback = `
			SELECT id, username, home_country, joined_on_datetime
			FROM geokrety.gk_users WHERE id = $1`
		var fu models.User
		if err2 := h.DB.QueryRow(ctx, fallback, id).Scan(
			&fu.UserID, &fu.Username, &fu.HomeCountry, &fu.JoinedAt,
		); err2 != nil {
			return nil, nil // not found
		}
		fu.TotalArchived = 0
		return &fu, nil
	}
	return &u, nil
}
