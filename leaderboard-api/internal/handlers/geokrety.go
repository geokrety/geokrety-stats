package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// ListGeokrety handles GET /api/v1/geokrety
// Query: page, per_page, sort (points|distance|moves), search
func (h *Handler) ListGeokrety(c *gin.Context) {
	page, perPage, offset := parsePagination(c)
	search := c.Query("search")

	var (
		countQ, listQ string
		args          []interface{}
	)

	if search != "" {
		countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_gk_stats WHERE name ILIKE $1`
		listQ = `
			SELECT gk_id, name, tracking_code, gk_type, missing, distance,
			       owner_id, owner_username, holder_id, holder_username,
			       total_moves, distinct_users, countries_count,
			       total_points_generated, current_multiplier, last_move_at
			FROM geokrety_stats.mv_gk_stats
			WHERE name ILIKE $1
			ORDER BY total_points_generated DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{"%" + search + "%", perPage, offset}
	} else {
		countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_gk_stats`
		listQ = `
			SELECT gk_id, name, tracking_code, gk_type, missing, distance,
			       owner_id, owner_username, holder_id, holder_username,
			       total_moves, distinct_users, countries_count,
			       total_points_generated, current_multiplier, last_move_at
			FROM geokrety_stats.mv_gk_stats
			ORDER BY total_points_generated DESC
			LIMIT $1 OFFSET $2`
		args = []interface{}{perPage, offset}
	}

	var total int64
	if search != "" {
		_ = h.DB.QueryRow(c.Request.Context(), countQ, "%"+search+"%").Scan(&total)
	} else {
		_ = h.DB.QueryRow(c.Request.Context(), countQ).Scan(&total)
	}

	rows, err := h.DB.Query(c.Request.Context(), listQ, args...)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		GkID                int64    `json:"gk_id"`
		Name                string   `json:"gk_name"`
		TrackingCode        *string  `json:"tracking_code,omitempty"`
		GkType              int      `json:"gk_type"`
		GkTypeName          string   `json:"gk_type_name"`
		Missing             bool     `json:"missing"`
		Distance            int64    `json:"distance_km"`
		OwnerID             *int64   `json:"owner_id,omitempty"`
		OwnerUsername       *string  `json:"owner_username,omitempty"`
		HolderID            *int64   `json:"holder_id,omitempty"`
		HolderUsername      *string  `json:"holder_username,omitempty"`
		TotalMoves          int64    `json:"total_moves"`
		DistinctUsers       int64    `json:"distinct_users"`
		CountriesCount      int64    `json:"countries_count"`
		TotalPointsGenerated float64 `json:"total_points_generated"`
		CurrentMultiplier   float64  `json:"current_multiplier"`
		LastMoveAt          interface{} `json:"last_move_at,omitempty"`
	}

	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.GkID, &r.Name, &r.TrackingCode, &r.GkType, &r.Missing,
			&r.Distance, &r.OwnerID, &r.OwnerUsername, &r.HolderID, &r.HolderUsername,
			&r.TotalMoves, &r.DistinctUsers, &r.CountriesCount,
			&r.TotalPointsGenerated, &r.CurrentMultiplier, &r.LastMoveAt); err != nil {
			errInternal(c, err)
			return
		}
		r.GkTypeName = gkTypeName(r.GkType)
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

// GetGeoKret handles GET /api/v1/geokrety/:id
func (h *Handler) GetGeoKret(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
		return
	}

	const q = `
		SELECT gk_id, name, tracking_code, gk_type, missing, distance, caches_count,
		       created_at, born_at, owner_id, owner_username, holder_id, holder_username,
		       total_moves, total_drops, total_grabs, total_seen, total_dips,
		       distinct_users, countries_count, caches_count_distinct,
		       total_points_generated, users_awarded, current_multiplier,
		       first_move_at, last_move_at
		FROM geokrety_stats.mv_gk_stats
		WHERE gk_id = $1`

	var g models.GeoKret
	err = h.DB.QueryRow(c.Request.Context(), q, id).Scan(
		&g.GkID, &g.Name, &g.TrackingCode, &g.GkType, &g.Missing, &g.Distance, &g.CachesCount,
		&g.CreatedAt, &g.BornAt, &g.OwnerID, &g.OwnerUsername, &g.HolderID, &g.HolderUsername,
		&g.TotalMoves, &g.TotalDrops, &g.TotalGrabs, &g.TotalSeen, &g.TotalDips,
		&g.DistinctUsers, &g.CountriesCount, &g.DistinctCaches,
		&g.TotalPointsGenerated, &g.UsersAwarded, &g.CurrentMultiplier,
		&g.FirstMoveAt, &g.LastMoveAt,
	)
	if err != nil {
		// fallback to raw table
		const fallback = `
			SELECT id, name, tracking_code, type, missing, distance, caches_count,
			       created_on_datetime, born_on_datetime
			FROM geokrety.gk_geokrety WHERE id = $1`
		err2 := h.DB.QueryRow(c.Request.Context(), fallback, id).Scan(
			&g.GkID, &g.Name, &g.TrackingCode, &g.GkType, &g.Missing, &g.Distance, &g.CachesCount,
			&g.CreatedAt, &g.BornAt,
		)
		if err2 != nil {
			errNotFound(c, "geokret not found")
			return
		}
		g.CurrentMultiplier = 1.0
	}

	g.GkTypeName = gkTypeName(g.GkType)
	ok(c, g, models.Meta{}, nil)
}

// GeoKretCountries handles GET /api/v1/geokrety/:id/countries
func (h *Handler) GeoKretCountries(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
		return
	}

	const q = `
		SELECT country, COUNT(*) as move_count
		FROM geokrety.gk_moves
		WHERE geokret = $1 AND country IS NOT NULL
		GROUP BY country
		ORDER BY move_count DESC`

	rows, err := h.DB.Query(c.Request.Context(), q, id)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.GkCountry, 0)
	for rows.Next() {
		var r models.GkCountry
		if err := rows.Scan(&r.Country, &r.MoveCount); err != nil {
			errInternal(c, err)
			return
		}
		out = append(out, r)
	}
	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}

// GeoKretMoves handles GET /api/v1/geokrety/:id/moves
func (h *Handler) GeoKretMoves(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety.gk_moves WHERE geokret = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT m.id, m.author, u.username, m.move_type,
		       m.country, m.waypoint, m.distance, m.moved_on_datetime,
		       COALESCE(p.points, 0)::float8 as points
		FROM geokrety.gk_moves m
		LEFT JOIN geokrety.gk_users u ON u.id = m.author
		LEFT JOIN (
			SELECT move_id, SUM(points) as points
			FROM geokrety_stats.user_points_log
			GROUP BY move_id
		) p ON p.move_id = m.id
		WHERE m.geokret = $1
		ORDER BY m.moved_on_datetime DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.GkMove, 0)
	for rows.Next() {
		var r models.GkMove
		var uname *string
		var points float64
		if err := rows.Scan(&r.MoveID, &r.AuthorID, &uname, &r.MoveType,
			&r.Country, &r.Waypoint, &r.Distance, &r.MovedOn, &points); err != nil {
			errInternal(c, err)
			return
		}
		if uname != nil {
			r.AuthorUsername = *uname
		}
		if points > 0 {
			r.Points = &points
		}
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

// GeoKretHolderHistory handles GET /api/v1/geokrety/:id/holders
func (h *Handler) GeoKretHolderHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety.gk_moves WHERE geokret = $1 AND move_type IN (0,1)`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT m.id, m.author, u.username, m.move_type, m.moved_on_datetime,
		       p_next.points AS points_earned
		FROM geokrety.gk_moves m
		LEFT JOIN geokrety.gk_users u ON u.id = m.author
		LEFT JOIN (
			SELECT move_id, SUM(points) AS points
			FROM geokrety_stats.user_points_log
			GROUP BY move_id
		) p_next ON p_next.move_id = m.id
		WHERE m.geokret = $1 AND m.move_type IN (0, 1)
		ORDER BY m.moved_on_datetime DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		MoveID       int64      `json:"move_id"`
		UserID       *int64     `json:"user_id,omitempty"`
		Username     *string    `json:"username,omitempty"`
		MoveType     int        `json:"move_type"`
		MoveTypeName string     `json:"move_type_name"`
		MovedAt      interface{} `json:"moved_at,omitempty"`
		PointsEarned *float64   `json:"points_earned,omitempty"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.MoveID, &r.UserID, &r.Username, &r.MoveType,
			&r.MovedAt, &r.PointsEarned); err != nil {
			errInternal(c, err)
			return
		}
		r.MoveTypeName = moveTypeName(r.MoveType)
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

// GeoKretPointsTimeline handles GET /api/v1/geokrety/:id/points/timeline
func (h *Handler) GeoKretPointsTimeline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `
		SELECT COUNT(DISTINCT DATE(awarded_at))
		FROM geokrety_stats.user_points_log WHERE gk_id = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT DATE(awarded_at AT TIME ZONE 'UTC')::text AS day,
		       SUM(points) AS pts,
		       COUNT(DISTINCT user_id) AS users
		FROM geokrety_stats.user_points_log
		WHERE gk_id = $1
		GROUP BY DATE(awarded_at AT TIME ZONE 'UTC')
		ORDER BY day DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		Day    string  `json:"day"`
		Points float64 `json:"points"`
		Users  int64   `json:"users"`
	}
	out := make([]row, 0)
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.Day, &r.Points, &r.Users); err != nil {
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

// GeoKretRelatedUsers handles GET /api/v1/geokrety/:id/related-users
// Returns users who have interacted with this geokrety
func (h *Handler) GeoKretRelatedUsers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `
		SELECT COUNT(*) FROM geokrety_stats.mv_geokrety_related_users WHERE geokret = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT r.user_id, u.username, u.home_country,
		       COALESCE(us.total_points, 0) total_points, COALESCE(us.total_moves, 0) total_moves,
		       r.interaction_count,
		       r.last_interaction
		FROM geokrety_stats.mv_geokrety_related_users r
		JOIN geokrety.gk_users u ON u.id = r.user_id
		LEFT JOIN geokrety_stats.mv_user_stats us ON us.user_id = r.user_id
		WHERE r.geokret = $1
		ORDER BY r.interaction_count DESC, r.last_interaction DESC NULLS LAST
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		UserID             int64       `json:"user_id"`
		Username           string      `json:"username"`
		HomeCountry        *string     `json:"home_country,omitempty"`
		TotalPoints        float64     `json:"total_points"`
		TotalMoves         int64       `json:"total_moves"`
		InteractionCount   int64       `json:"interaction_count"`
		LastInteraction    interface{} `json:"last_interaction,omitempty"`
	}

	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.UserID, &r.Username, &r.HomeCountry, &r.TotalPoints,
			&r.TotalMoves, &r.InteractionCount, &r.LastInteraction); err != nil {
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

