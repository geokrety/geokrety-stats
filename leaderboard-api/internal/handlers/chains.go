package handlers

import (
	"strings"
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

	q := `
		SELECT
			gc.id,
			gc.gk_id,
			COALESCE(g.gkid, gc.gk_id) AS public_gk_id,
			COALESCE(g.name, '') AS gk_name,
			pic.key AS gk_avatar_key,
			g.type AS gk_type,
			gc.status,
			gc.started_at,
			gc.ended_at,
			gc.chain_last_active,
			gc.end_reason,
			COALESCE(cm.member_count, 0) AS member_count,
			COALESCE(cp.chain_points, 0)::float8 AS chain_points
		FROM geokrety_stats.gk_chains gc
		LEFT JOIN geokrety.gk_geokrety g ON g.id = gc.gk_id
		LEFT JOIN geokrety.gk_pictures pic ON pic.id = g.avatar
		LEFT JOIN (
			SELECT chain_id, COUNT(*)::bigint AS member_count
			FROM geokrety_stats.gk_chain_members
			GROUP BY chain_id
		) cm ON cm.chain_id = gc.id
		LEFT JOIN (
			SELECT t.chain_id, SUM(t.points)::float8 AS chain_points
			FROM (
				SELECT upl.chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				WHERE upl.chain_id IS NOT NULL
				UNION ALL
				SELECT gc2.id AS chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				JOIN geokrety_stats.gk_chains gc2
				  ON upl.chain_id IS NULL
				 AND upl.label IN ('chain_bonus', 'chain_bonus_owner')
				 AND upl.gk_id = gc2.gk_id
				 AND upl.awarded_at >= gc2.started_at
				 AND upl.awarded_at <= COALESCE(gc2.ended_at, gc2.chain_last_active)
			) t
			GROUP BY t.chain_id
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
		&out.GkAvatarKey,
		&out.GkType,
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
	sort := c.DefaultQuery("sort", "position")
	order := strings.ToLower(c.DefaultQuery("order", "asc"))
	if order != "asc" && order != "desc" {
		order = "asc"
	}
	dir := "ASC"
	if order == "desc" {
		dir = "DESC"
	}

	orderBy := "cm.position " + dir
	switch sort {
	case "user":
		orderBy = "username " + dir + ", cm.position ASC"
	case "joined":
		orderBy = "cm.joined_at " + dir + ", cm.position ASC"
	}

	const countQ = `SELECT COUNT(*) FROM geokrety_stats.gk_chain_members WHERE chain_id = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, id).Scan(&total)

	q := `
		SELECT cm.chain_id, cm.user_id, COALESCE(u.username, CONCAT('User #', cm.user_id::text)) AS username,
		       pic.key AS user_avatar_key,
		       cm.position, cm.joined_at
		FROM geokrety_stats.gk_chain_members cm
		LEFT JOIN geokrety.gk_users u ON u.id = cm.user_id
		LEFT JOIN geokrety.gk_pictures pic ON pic.id = u.avatar
		WHERE cm.chain_id = $1
		ORDER BY ` + orderBy + `
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
		if err := rows.Scan(&r.ChainID, &r.UserID, &r.Username, &r.UserAvatarKey, &r.Position, &r.JoinedAt); err != nil {
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
	sort := c.DefaultQuery("sort", "date")
	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}

	orderBy := "m.moved_on_datetime " + dir + ", m.id " + dir
	switch sort {
	case "user":
		orderBy = "u.username " + dir + ", m.moved_on_datetime DESC"
	case "type":
		orderBy = "m.move_type " + dir + ", m.moved_on_datetime DESC"
	case "chain_points":
		orderBy = "chain_points " + dir + ", m.moved_on_datetime DESC"
	}

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

	q := `
		SELECT
			m.id,
			m.author,
			u.username,
			pic.key AS author_avatar_key,
			m.move_type,
			m.country,
			m.waypoint,
			m.moved_on_datetime,
			COALESCE(p.chain_points, 0)::float8 AS chain_points
		FROM geokrety.gk_moves m
		JOIN geokrety_stats.gk_chains gc ON gc.gk_id = m.geokret
		LEFT JOIN geokrety.gk_users u ON u.id = m.author
		LEFT JOIN geokrety.gk_pictures pic ON pic.id = u.avatar
		LEFT JOIN LATERAL (
			SELECT SUM(upl.points)::float8 AS chain_points
			FROM geokrety_stats.user_points_log upl
			WHERE upl.move_id = m.id
			  AND (
				upl.chain_id = gc.id
				OR (
					upl.chain_id IS NULL
					AND upl.label IN ('chain_bonus', 'chain_bonus_owner')
					AND upl.gk_id = gc.gk_id
					AND upl.awarded_at >= gc.started_at
					AND upl.awarded_at <= COALESCE(gc.ended_at, gc.chain_last_active)
				)
			  )
		) p ON TRUE
		WHERE gc.id = $1
		  AND m.moved_on_datetime >= gc.started_at
		  AND m.moved_on_datetime <= COALESCE(gc.ended_at, gc.chain_last_active)
		ORDER BY ` + orderBy + `
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
		if err := rows.Scan(&r.MoveID, &r.AuthorID, &r.AuthorUsername, &r.AuthorAvatarKey, &r.MoveType, &r.Country, &r.Waypoint, &r.MovedOn, &r.ChainPoints); err != nil {
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
	sort := c.DefaultQuery("sort", "started")
	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}

	orderBy := "gc.started_at " + dir + ", gc.id " + dir
	switch sort {
	case "chain":
		orderBy = "gc.id " + dir
	case "gk":
		orderBy = "COALESCE(g.name, '') " + dir + ", gc.id DESC"
	case "status":
		orderBy = "gc.status " + dir + ", gc.started_at DESC"
	case "last_active":
		orderBy = "gc.chain_last_active " + dir + ", gc.started_at DESC"
	case "members":
		orderBy = "member_count " + dir + ", gc.started_at DESC"
	case "points":
		orderBy = "chain_points " + dir + ", gc.started_at DESC"
	case "completed":
		orderBy = "has_user_completion " + dir + ", gc.started_at DESC"
	}

	const countQ = `
		SELECT COUNT(*)
		FROM geokrety_stats.gk_chain_members cm
		WHERE cm.user_id = $1
	`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, userID).Scan(&total)

	q := `
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
			SELECT t.chain_id, SUM(t.points)::float8 AS chain_points
			FROM (
				SELECT upl.chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				WHERE upl.chain_id IS NOT NULL
				UNION ALL
				SELECT gc3.id AS chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				JOIN geokrety_stats.gk_chains gc3
				  ON upl.chain_id IS NULL
				 AND upl.label IN ('chain_bonus', 'chain_bonus_owner')
				 AND upl.gk_id = gc3.gk_id
				 AND upl.awarded_at >= gc3.started_at
				 AND upl.awarded_at <= COALESCE(gc3.ended_at, gc3.chain_last_active)
			) t
			GROUP BY t.chain_id
		) cp ON cp.chain_id = gc.id
		WHERE cm.user_id = $1
		ORDER BY ` + orderBy + `
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
	sort := c.DefaultQuery("sort", "started")
	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}

	orderBy := "gc.started_at " + dir + ", gc.id " + dir
	switch sort {
	case "chain":
		orderBy = "gc.id " + dir
	case "status":
		orderBy = "gc.status " + dir + ", gc.started_at DESC"
	case "ended":
		orderBy = "gc.ended_at " + dir + " NULLS LAST, gc.started_at DESC"
	case "last_active":
		orderBy = "gc.chain_last_active " + dir + ", gc.started_at DESC"
	case "members":
		orderBy = "member_count " + dir + ", gc.started_at DESC"
	case "points":
		orderBy = "chain_points " + dir + ", gc.started_at DESC"
	}

	const countQ = `SELECT COUNT(*) FROM geokrety_stats.gk_chains WHERE gk_id = $1`
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, gkID).Scan(&total)

	q := `
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
			SELECT t.chain_id, SUM(t.points)::float8 AS chain_points
			FROM (
				SELECT upl.chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				WHERE upl.chain_id IS NOT NULL
				UNION ALL
				SELECT gc4.id AS chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				JOIN geokrety_stats.gk_chains gc4
				  ON upl.chain_id IS NULL
				 AND upl.label IN ('chain_bonus', 'chain_bonus_owner')
				 AND upl.gk_id = gc4.gk_id
				 AND upl.awarded_at >= gc4.started_at
				 AND upl.awarded_at <= COALESCE(gc4.ended_at, gc4.chain_last_active)
			) t
			GROUP BY t.chain_id
		) cp ON cp.chain_id = gc.id
		WHERE gc.gk_id = $1
		ORDER BY ` + orderBy + `
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
	sort := c.DefaultQuery("sort", "started")
	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}

	orderBy := "gc.started_at " + dir + ", gc.id " + dir
	switch sort {
	case "chain":
		orderBy = "gc.id " + dir
	case "gk":
		orderBy = "COALESCE(g.name, '') " + dir + ", gc.id DESC"
	case "status":
		orderBy = "gc.status " + dir + ", gc.started_at DESC"
	case "move_chain_points":
		orderBy = "move_chain_points " + dir + ", gc.started_at DESC"
	case "points":
		orderBy = "chain_points " + dir + ", gc.started_at DESC"
	case "members":
		orderBy = "member_count " + dir + ", gc.started_at DESC"
	}

	q := `
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
			SELECT
				COALESCE(upl.chain_id, gc_match.id) AS chain_id,
				upl.move_id,
				SUM(upl.points)::float8 AS move_chain_points
			FROM geokrety_stats.user_points_log upl
			LEFT JOIN geokrety_stats.gk_chains gc_match
			  ON upl.chain_id IS NULL
			 AND upl.label IN ('chain_bonus', 'chain_bonus_owner')
			 AND upl.gk_id = gc_match.gk_id
			 AND upl.awarded_at >= gc_match.started_at
			 AND upl.awarded_at <= COALESCE(gc_match.ended_at, gc_match.chain_last_active)
			WHERE upl.move_id = $1
			  AND (
				upl.chain_id IS NOT NULL
				OR (
					upl.chain_id IS NULL
					AND upl.label IN ('chain_bonus', 'chain_bonus_owner')
				)
			  )
			GROUP BY COALESCE(upl.chain_id, gc_match.id), upl.move_id
		) upl
		JOIN geokrety_stats.gk_chains gc ON gc.id = upl.chain_id
		LEFT JOIN geokrety.gk_geokrety g ON g.id = gc.gk_id
		LEFT JOIN (
			SELECT chain_id, COUNT(*)::bigint AS member_count
			FROM geokrety_stats.gk_chain_members
			GROUP BY chain_id
		) cm ON cm.chain_id = gc.id
		LEFT JOIN (
			SELECT t.chain_id, SUM(t.points)::float8 AS chain_points
			FROM (
				SELECT upl.chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				WHERE upl.chain_id IS NOT NULL
				UNION ALL
				SELECT gc5.id AS chain_id, upl.points
				FROM geokrety_stats.user_points_log upl
				JOIN geokrety_stats.gk_chains gc5
				  ON upl.chain_id IS NULL
				 AND upl.label IN ('chain_bonus', 'chain_bonus_owner')
				 AND upl.gk_id = gc5.gk_id
				 AND upl.awarded_at >= gc5.started_at
				 AND upl.awarded_at <= COALESCE(gc5.ended_at, gc5.chain_last_active)
			) t
			GROUP BY t.chain_id
		) cp ON cp.chain_id = gc.id
		ORDER BY ` + orderBy + `
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
