package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// ChainDetail handles GET /api/v1/chains/:id
func (h *Handler) ChainDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid chain id")
		return
	}

	const q = `
		SELECT
			gc.id,
			gc.gk_id,
			COALESCE(g.gkid, gc.gk_id) AS public_gk_id,
			COALESCE(g.name, '') AS gk_name,
			gc.status,
			gc.started_at,
			gc.ended_at,
			gc.chain_last_active,
			gc.end_reason,
			COALESCE(cm.member_count, 0) AS member_count,
			COALESCE(cp.chain_points, 0)::float8 AS chain_points
		FROM geokrety_stats.gk_chains gc
		LEFT JOIN geokrety.gk_geokrety g ON g.id = gc.gk_id
		LEFT JOIN (
			SELECT chain_id, COUNT(*)::bigint AS member_count
			FROM geokrety_stats.gk_chain_members
			GROUP BY chain_id
		) cm ON cm.chain_id = gc.id
		LEFT JOIN (
			SELECT chain_id, SUM(points)::float8 AS chain_points
			FROM geokrety_stats.user_points_log
			WHERE chain_id IS NOT NULL
			GROUP BY chain_id
		) cp ON cp.chain_id = gc.id
		WHERE gc.id = $1
	`

	var out models.ChainSummary
	var publicGkID int64
	if err := h.DB.QueryRow(c.Request.Context(), q, id).Scan(
		&out.ChainID,
		&out.GkID,
		&publicGkID,
		&out.GkName,
		&out.Status,
		&out.StartedAt,
		&out.EndedAt,
		&out.ChainLastActive,
		&out.EndReason,
		&out.MemberCount,
		&out.ChainPoints,
	); err != nil {
		errNotFound(c, "chain not found")
		return
	}
	out.GkHexID = gkHexID(publicGkID)

	ok(c, out, models.Meta{}, nil)
}

// ChainMembers handles GET /api/v1/chains/:id/members
func (h *Handler) ChainMembers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid chain id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety_stats.gk_chain_members WHERE chain_id = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT cm.chain_id, cm.user_id, COALESCE(u.username, CONCAT('User #', cm.user_id::text)) AS username,
		       cm.position, cm.joined_at
		FROM geokrety_stats.gk_chain_members cm
		LEFT JOIN geokrety.gk_users u ON u.id = cm.user_id
		WHERE cm.chain_id = $1
		ORDER BY cm.position ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.ChainMember, 0)
	for rows.Next() {
		var r models.ChainMember
		if err := rows.Scan(&r.ChainID, &r.UserID, &r.Username, &r.Position, &r.JoinedAt); err != nil {
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

// ChainMoves handles GET /api/v1/chains/:id/moves
func (h *Handler) ChainMoves(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid chain id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `
		SELECT COUNT(*)
		FROM geokrety.gk_moves m
		JOIN geokrety_stats.gk_chains gc ON gc.gk_id = m.geokret
		WHERE gc.id = $1
		  AND m.moved_on_datetime >= gc.started_at
		  AND m.moved_on_datetime <= COALESCE(gc.ended_at, gc.chain_last_active)
	`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	const q = `
		SELECT
			m.id,
			m.author,
			u.username,
			m.move_type,
			m.country,
			m.waypoint,
			m.moved_on_datetime,
			COALESCE(p.chain_points, 0)::float8 AS chain_points
		FROM geokrety.gk_moves m
		JOIN geokrety_stats.gk_chains gc ON gc.gk_id = m.geokret
		LEFT JOIN geokrety.gk_users u ON u.id = m.author
		LEFT JOIN (
			SELECT move_id, SUM(points)::float8 AS chain_points
			FROM geokrety_stats.user_points_log
			WHERE chain_id = $1
			GROUP BY move_id
		) p ON p.move_id = m.id
		WHERE gc.id = $1
		  AND m.moved_on_datetime >= gc.started_at
		  AND m.moved_on_datetime <= COALESCE(gc.ended_at, gc.chain_last_active)
		ORDER BY m.moved_on_datetime DESC, m.id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.DB.Query(c.Request.Context(), q, id, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.ChainMove, 0)
	for rows.Next() {
		var r models.ChainMove
		if err := rows.Scan(&r.MoveID, &r.AuthorID, &r.AuthorUsername, &r.MoveType, &r.Country, &r.Waypoint, &r.MovedOn, &r.ChainPoints); err != nil {
			errInternal(c, err)
			return
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

// UserChains handles GET /api/v1/users/:id/chains
func (h *Handler) UserChains(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid user id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `
		SELECT COUNT(*)
		FROM geokrety_stats.gk_chain_members cm
		WHERE cm.user_id = $1
	`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, userID).Scan(&total)

	const q = `
		SELECT
			gc.id,
			gc.gk_id,
			COALESCE(g.gkid, gc.gk_id) AS public_gk_id,
			COALESCE(g.name, '') AS gk_name,
			gc.status,
			gc.started_at,
			gc.ended_at,
			gc.chain_last_active,
			gc.end_reason,
			COALESCE(cm2.member_count, 0) AS member_count,
			COALESCE(cp.chain_points, 0)::float8 AS chain_points,
			EXISTS (
				SELECT 1 FROM geokrety_stats.gk_chain_completions cc
				WHERE cc.user_id = $1 AND cc.chain_id = gc.id
			) AS has_user_completion
		FROM geokrety_stats.gk_chain_members cm
		JOIN geokrety_stats.gk_chains gc ON gc.id = cm.chain_id
		LEFT JOIN geokrety.gk_geokrety g ON g.id = gc.gk_id
		LEFT JOIN (
			SELECT chain_id, COUNT(*)::bigint AS member_count
			FROM geokrety_stats.gk_chain_members
			GROUP BY chain_id
		) cm2 ON cm2.chain_id = gc.id
		LEFT JOIN (
			SELECT chain_id, SUM(points)::float8 AS chain_points
			FROM geokrety_stats.user_points_log
			WHERE chain_id IS NOT NULL
			GROUP BY chain_id
		) cp ON cp.chain_id = gc.id
		WHERE cm.user_id = $1
		ORDER BY gc.started_at DESC, gc.id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.DB.Query(c.Request.Context(), q, userID, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.ChainSummary, 0)
	for rows.Next() {
		var r models.ChainSummary
		var publicGkID int64
		if err := rows.Scan(
			&r.ChainID,
			&r.GkID,
			&publicGkID,
			&r.GkName,
			&r.Status,
			&r.StartedAt,
			&r.EndedAt,
			&r.ChainLastActive,
			&r.EndReason,
			&r.MemberCount,
			&r.ChainPoints,
			&r.HasUserCompletion,
		); err != nil {
			errInternal(c, err)
			return
		}
		r.GkHexID = gkHexID(publicGkID)
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

// GeoKretChains handles GET /api/v1/geokrety/:id/chains
func (h *Handler) GeoKretChains(c *gin.Context) {
	gkID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid geokret id")
		return
	}
	page, perPage, offset := parsePagination(c)

	const countQ = `SELECT COUNT(*) FROM geokrety_stats.gk_chains WHERE gk_id = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, gkID).Scan(&total)

	const q = `
		SELECT
			gc.id,
			gc.gk_id,
			COALESCE(g.gkid, gc.gk_id) AS public_gk_id,
			COALESCE(g.name, '') AS gk_name,
			gc.status,
			gc.started_at,
			gc.ended_at,
			gc.chain_last_active,
			gc.end_reason,
			COALESCE(cm.member_count, 0) AS member_count,
			COALESCE(cp.chain_points, 0)::float8 AS chain_points
		FROM geokrety_stats.gk_chains gc
		LEFT JOIN geokrety.gk_geokrety g ON g.id = gc.gk_id
		LEFT JOIN (
			SELECT chain_id, COUNT(*)::bigint AS member_count
			FROM geokrety_stats.gk_chain_members
			GROUP BY chain_id
		) cm ON cm.chain_id = gc.id
		LEFT JOIN (
			SELECT chain_id, SUM(points)::float8 AS chain_points
			FROM geokrety_stats.user_points_log
			WHERE chain_id IS NOT NULL
			GROUP BY chain_id
		) cp ON cp.chain_id = gc.id
		WHERE gc.gk_id = $1
		ORDER BY gc.started_at DESC, gc.id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.DB.Query(c.Request.Context(), q, gkID, perPage, offset)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.ChainSummary, 0)
	for rows.Next() {
		var r models.ChainSummary
		var publicGkID int64
		if err := rows.Scan(
			&r.ChainID,
			&r.GkID,
			&publicGkID,
			&r.GkName,
			&r.Status,
			&r.StartedAt,
			&r.EndedAt,
			&r.ChainLastActive,
			&r.EndReason,
			&r.MemberCount,
			&r.ChainPoints,
		); err != nil {
			errInternal(c, err)
			return
		}
		r.GkHexID = gkHexID(publicGkID)
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

// MoveChains handles GET /api/v1/moves/:id/chains
func (h *Handler) MoveChains(c *gin.Context) {
	moveID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid move id")
		return
	}

	const q = `
		SELECT
			gc.id,
			gc.gk_id,
			COALESCE(g.gkid, gc.gk_id) AS public_gk_id,
			COALESCE(g.name, '') AS gk_name,
			gc.status,
			gc.started_at,
			gc.ended_at,
			gc.chain_last_active,
			gc.end_reason,
			COALESCE(cm.member_count, 0) AS member_count,
			COALESCE(cp.chain_points, 0)::float8 AS chain_points,
			upl.move_id,
			COALESCE(upl.move_chain_points, 0)::float8 AS move_chain_points
		FROM (
			SELECT chain_id, move_id, SUM(points)::float8 AS move_chain_points
			FROM geokrety_stats.user_points_log
			WHERE move_id = $1 AND chain_id IS NOT NULL
			GROUP BY chain_id, move_id
		) upl
		JOIN geokrety_stats.gk_chains gc ON gc.id = upl.chain_id
		LEFT JOIN geokrety.gk_geokrety g ON g.id = gc.gk_id
		LEFT JOIN (
			SELECT chain_id, COUNT(*)::bigint AS member_count
			FROM geokrety_stats.gk_chain_members
			GROUP BY chain_id
		) cm ON cm.chain_id = gc.id
		LEFT JOIN (
			SELECT chain_id, SUM(points)::float8 AS chain_points
			FROM geokrety_stats.user_points_log
			WHERE chain_id IS NOT NULL
			GROUP BY chain_id
		) cp ON cp.chain_id = gc.id
		ORDER BY gc.started_at DESC, gc.id DESC
	`

	rows, err := h.DB.Query(c.Request.Context(), q, moveID)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.ChainSummary, 0)
	for rows.Next() {
		var r models.ChainSummary
		var publicGkID int64
		if err := rows.Scan(
			&r.ChainID,
			&r.GkID,
			&publicGkID,
			&r.GkName,
			&r.Status,
			&r.StartedAt,
			&r.EndedAt,
			&r.ChainLastActive,
			&r.EndReason,
			&r.MemberCount,
			&r.ChainPoints,
			&r.MoveID,
			&r.MoveChainPoints,
		); err != nil {
			errInternal(c, err)
			return
		}
		r.GkHexID = gkHexID(publicGkID)
		out = append(out, r)
	}

	ok(c, out, models.Meta{Total: int64(len(out))}, nil)
}
