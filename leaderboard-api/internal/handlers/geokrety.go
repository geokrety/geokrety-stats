package handlers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/geokrety/leaderboard-api/internal/models"
)

// gkHexID converts a numeric GeoKrety public ID (gkid column) to its GKXXXX format.
// The gkid column is the actual public ID (different from internal id).
// Example: 15598 → "GK3CEE"
func gkHexID(publicID int64) string {
	return fmt.Sprintf("GK%04X", publicID)
}

// ListGeokrety handles GET /api/v1/geokrety
// Query: page, per_page, sort (points|moves|users|countries|loves|multiplier|distance), search
func (h *Handler) ListGeokrety(c *gin.Context) {
	page, perPage, offset := parsePagination(c)
	search := c.Query("search")

	// Map frontend sort key → SQL ORDER BY expression
	sortMap := map[string]string{
		"points":     "s.total_points_generated DESC",
		"moves":      "s.total_moves DESC",
		"users":      "s.distinct_users DESC",
		"countries":  "s.countries_count DESC",
		"loves":      "s.loves_count DESC",
		"multiplier": "s.current_multiplier DESC",
		"distance":   "s.distance DESC",
		"avg_points": "(s.total_points_generated / NULLIF(s.total_moves, 0)) DESC",
	}
	sortParam := c.DefaultQuery("sort", "points")
	orderBy, ok2 := sortMap[sortParam]
	if !ok2 {
		orderBy = "s.total_points_generated DESC"
	}

	baseSelect := `
		SELECT s.gk_id, COALESCE(g.gkid, s.gk_id) AS public_gk_id, s.name, s.gk_type, s.missing, s.distance,
		       s.owner_id, s.owner_username, s.holder_id, s.holder_username,
		       s.in_cache, s.is_non_collectible, s.is_parked, s.loves_count,
		       s.total_moves, s.distinct_users, s.countries_count,
		       s.total_points_generated, s.current_multiplier, s.last_move_at`

	var (
		countQ, listQ string
		countArgs     []interface{}
		listArgs      []interface{}
	)

	fromClause := `
		FROM geokrety_stats.mv_gk_stats s
		LEFT JOIN geokrety.gk_geokrety g ON g.id = s.gk_id`

	if search != "" {
		countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_gk_stats WHERE name ILIKE $1`
		listQ = baseSelect + fromClause + `
			WHERE s.name ILIKE $1
			ORDER BY ` + orderBy + `
			LIMIT $2 OFFSET $3`
		countArgs = []interface{}{"%" + search + "%"}
		listArgs = []interface{}{"%" + search + "%", perPage, offset}
	} else {
		countQ = `SELECT COUNT(*) FROM geokrety_stats.mv_gk_stats`
		listQ = baseSelect + fromClause + `
			ORDER BY ` + orderBy + `
			LIMIT $1 OFFSET $2`
		countArgs = nil
		listArgs = []interface{}{perPage, offset}
	}

	var total int64
	if countArgs != nil {
		_ = h.DB.QueryRow(c.Request.Context(), countQ, countArgs...).Scan(&total)
	} else {
		_ = h.DB.QueryRow(c.Request.Context(), countQ).Scan(&total)
	}

	dbRows, err := h.DB.Query(c.Request.Context(), listQ, listArgs...)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer dbRows.Close()

	type row struct {
		GkID                int64    `json:"gk_id"`
		GkHexID             string   `json:"gk_hex_id"`
		Name                string   `json:"gk_name"`
		GkType              int      `json:"gk_type"`
		GkTypeName          string   `json:"gk_type_name"`
		Missing             bool     `json:"missing"`
		InCache             bool     `json:"in_cache"`
		IsNonCollectible    bool     `json:"is_non_collectible"`
		IsParked            bool     `json:"is_parked"`
		LovesCount          int      `json:"loves_count"`
		Distance            int64    `json:"distance_km"`
		OwnerID             *int64   `json:"owner_id,omitempty"`
		OwnerUsername       *string  `json:"owner_username,omitempty"`
		HolderID            *int64   `json:"holder_id,omitempty"`
		HolderUsername      *string  `json:"holder_username,omitempty"`
		TotalMoves          int64    `json:"total_moves"`
		DistinctUsers       int64    `json:"distinct_users"`
		CountriesCount      int64    `json:"countries_count"`
		TotalPointsGenerated float64 `json:"total_points_generated"`
		CurrentMultiplier   float64    `json:"current_multiplier"`
		LastMoveAt          *time.Time `json:"last_move_at,omitempty"`
	}

	var out []row
	for dbRows.Next() {
		var r row
		var publicGkID int64
		if err := dbRows.Scan(&r.GkID, &publicGkID, &r.Name, &r.GkType, &r.Missing, &r.Distance,
			&r.OwnerID, &r.OwnerUsername, &r.HolderID, &r.HolderUsername,
			&r.InCache, &r.IsNonCollectible, &r.IsParked, &r.LovesCount,
			&r.TotalMoves, &r.DistinctUsers, &r.CountriesCount,
			&r.TotalPointsGenerated, &r.CurrentMultiplier, &r.LastMoveAt); err != nil {
			errInternal(c, err)
			return
		}
		r.GkTypeName = gkTypeName(r.GkType)
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

// GetGeoKret handles GET /api/v1/geokrety/:id
func (h *Handler) GetGeoKret(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
		return
	}

	const q = `
		SELECT s.gk_id, COALESCE(g.gkid, s.gk_id) AS public_gk_id, s.name, s.avatar_bucket, s.avatar_key, s.gk_type, s.missing, s.distance, s.caches_count,
		       s.created_on_datetime, s.born_on_datetime,
		       s.owner_id, s.owner_username, s.holder_id, s.holder_username,
		       COALESCE(o.home_country, NULL) AS owner_home_country,
		       COALESCE(h.home_country, NULL) AS holder_home_country,
		       COALESCE(m.country, NULL) AS cache_country,
		       s.in_cache, s.is_non_collectible, s.is_parked, s.loves_count,
		       s.total_moves, s.total_drops, s.total_grabs, s.total_comments, s.total_seen, s.total_dips,
		       s.distinct_users, s.countries_count, s.caches_count_distinct,
		       s.total_points_generated, s.users_awarded, s.current_multiplier,
		       s.first_move_at, s.last_move_at
		FROM geokrety_stats.mv_gk_stats s
		LEFT JOIN geokrety.gk_geokrety g ON g.id = s.gk_id
		LEFT JOIN geokrety.gk_users o ON o.id = s.owner_id
		LEFT JOIN geokrety.gk_users h ON h.id = s.holder_id
		LEFT JOIN LATERAL (
			SELECT country FROM geokrety.gk_moves
			WHERE geokret = $1 AND country IS NOT NULL
			ORDER BY moved_on_datetime DESC LIMIT 1
		) m ON TRUE
		WHERE s.gk_id = $1`

	var g models.GeoKret
	var publicGkID int64
	var avatarBucket sql.NullString
	var avatarKey sql.NullString
	err = h.DB.QueryRow(c.Request.Context(), q, id).Scan(
		&g.GkID, &publicGkID, &g.Name, &avatarBucket, &avatarKey, &g.GkType, &g.Missing, &g.Distance, &g.CachesCount,
		&g.CreatedAt, &g.BornAt,
		&g.OwnerID, &g.OwnerUsername, &g.HolderID, &g.HolderUsername,
		&g.OwnerHomeCountry, &g.HolderHomeCountry, &g.CacheCountry,
		&g.InCache, &g.IsNonCollectible, &g.IsParked, &g.LovesCount,
		&g.TotalMoves, &g.TotalDrops, &g.TotalGrabs, &g.TotalComments, &g.TotalSeen, &g.TotalDips,
		&g.DistinctUsers, &g.CountriesCount, &g.DistinctCaches,
		&g.TotalPointsGenerated, &g.UsersAwarded, &g.CurrentMultiplier,
		&g.FirstMoveAt, &g.LastMoveAt,
	)
	if err != nil {
		// fallback to raw table (no stats)
		const fallback = `
			SELECT g.id, g.gkid, g.name, pic.bucket, pic.key, g.type, g.missing, g.distance, g.caches_count,
			       g.created_on_datetime, g.born_on_datetime
			FROM geokrety.gk_geokrety g
			LEFT JOIN geokrety.gk_pictures pic ON pic.id = g.avatar
			WHERE g.id = $1`
		err2 := h.DB.QueryRow(c.Request.Context(), fallback, id).Scan(
			&g.GkID, &publicGkID, &g.Name, &avatarBucket, &avatarKey, &g.GkType, &g.Missing, &g.Distance, &g.CachesCount,
			&g.CreatedAt, &g.BornAt,
		)
		if err2 != nil {
			errNotFound(c, "geokret not found")
			return
		}
		g.CurrentMultiplier = 1.0
	}

		g.GkHexID = gkHexID(publicGkID)
		g.GkTypeName = gkTypeName(g.GkType)
		g.Avatar = avatarRef(avatarBucket, avatarKey)
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
	sort := c.DefaultQuery("sort", "date")
	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	dir := "DESC"
	if order == "asc" {
		dir = "ASC"
	}
	awardingOnly := c.DefaultQuery("awarding_only", "false") == "true"

	var moveTypes []int32
	if typesParam := c.Query("types"); typesParam != "" {
		for _, raw := range strings.Split(typesParam, ",") {
			trimmed := strings.TrimSpace(raw)
			if trimmed == "" {
				continue
			}
			t, convErr := strconv.Atoi(trimmed)
			if convErr != nil || t < 0 || t > 5 {
				c.JSON(400, gin.H{"errors": []gin.H{{"status": "400", "title": "invalid move type filter"}}})
				return
			}
			moveTypes = append(moveTypes, int32(t))
		}
	}

	var orderBy string
	switch sort {
	case "author":
		orderBy = "LOWER(COALESCE(u.username, '')) " + dir + ", m.moved_on_datetime DESC"
	case "waypoint":
		orderBy = "m.waypoint " + dir + " NULLS LAST, m.moved_on_datetime DESC"
	case "type":
		orderBy = "m.move_type " + dir + ", m.moved_on_datetime DESC"
	case "country":
		orderBy = "m.country " + dir + " NULLS LAST, m.moved_on_datetime DESC"
	case "points":
		orderBy = "COALESCE(p.points, 0) " + dir + ", m.moved_on_datetime DESC"
	default:
		orderBy = "m.moved_on_datetime " + dir
	}

	whereParts := []string{"m.geokret = $1"}
	args := []interface{}{id}
	argPos := 2
	if len(moveTypes) > 0 {
		whereParts = append(whereParts, "m.move_type = ANY($"+strconv.Itoa(argPos)+"::int[])")
		args = append(args, moveTypes)
		argPos++
	}
	if awardingOnly {
		whereParts = append(whereParts, `EXISTS (
			SELECT 1 FROM geokrety_stats.user_points_log upl
			WHERE upl.move_id = m.id AND upl.points > 0
		)`)
	}
	whereSQL := strings.Join(whereParts, " AND ")

	countQ := `SELECT COUNT(*) FROM geokrety.gk_moves m WHERE ` + whereSQL
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, args...).Scan(&total)

	q := `
		SELECT m.id, m.author, u.username, pic.bucket, pic.key, m.move_type,
		       m.country, m.waypoint, m.distance, m.moved_on_datetime,
		       COALESCE(p.points, 0)::float8 as points,
		       p.chain_id
		FROM geokrety.gk_moves m
		LEFT JOIN geokrety.gk_users u ON u.id = m.author
		LEFT JOIN geokrety.gk_pictures pic ON pic.id = u.avatar
		LEFT JOIN (
			SELECT move_id, SUM(points) as points, MAX(chain_id) as chain_id
			FROM geokrety_stats.user_points_log
			GROUP BY move_id
		) p ON p.move_id = m.id
		WHERE ` + whereSQL + `
		ORDER BY ` + orderBy + `
		LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)

	queryArgs := append(args, perPage, offset)
	rows, err := h.DB.Query(c.Request.Context(), q, queryArgs...)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	out := make([]models.GkMove, 0)
	for rows.Next() {
		var r models.GkMove
		var uname *string
		var avatarBucket sql.NullString
		var avatarKey sql.NullString
		var points float64
		var chainID *int64
		if err := rows.Scan(&r.MoveID, &r.AuthorID, &uname, &avatarBucket, &avatarKey, &r.MoveType,
			&r.Country, &r.Waypoint, &r.Distance, &r.MovedOn, &points, &chainID); err != nil {
			errInternal(c, err)
			return
		}
		if uname != nil {
			r.AuthorUsername = *uname
		}
		r.AuthorAvatar = avatarRef(avatarBucket, avatarKey)
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

// GeoKretPointsLog handles GET /api/v1/geokrety/:id/points/log
// Returns detailed point award entries for a given GeoKret.
func (h *Handler) GeoKretPointsLog(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		errNotFound(c, "invalid gk id")
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
	awardingOnly := c.DefaultQuery("awarding_only", "false") == "true"

	var moveTypes []int32
	if typesParam := c.Query("types"); typesParam != "" {
		for _, raw := range strings.Split(typesParam, ",") {
			trimmed := strings.TrimSpace(raw)
			if trimmed == "" {
				continue
			}
			t, convErr := strconv.Atoi(trimmed)
			if convErr != nil || t < 0 || t > 5 {
				c.JSON(400, gin.H{"errors": []gin.H{{"status": "400", "title": "invalid move type filter"}}})
				return
			}
			moveTypes = append(moveTypes, int32(t))
		}
	}

	var orderBy string
	switch sort {
	case "points":
		orderBy = "pl.points " + dir + ", pl.awarded_at DESC"
	case "user":
		orderBy = "LOWER(COALESCE(u.username, '')) " + dir + ", pl.awarded_at DESC"
	case "label":
		orderBy = "pl.label " + dir + ", pl.awarded_at DESC"
	case "type":
		orderBy = "m.move_type " + dir + ", pl.awarded_at DESC"
	case "country":
		orderBy = "m.country " + dir + " NULLS LAST, pl.awarded_at DESC"
	case "waypoint":
		orderBy = "m.waypoint " + dir + " NULLS LAST, pl.awarded_at DESC"
	default:
		orderBy = "pl.awarded_at " + dir
	}

	whereParts := []string{"pl.gk_id = $1"}
	args := []interface{}{id}
	argPos := 2
	if awardingOnly {
		whereParts = append(whereParts, "pl.points > 0")
	}
	if len(moveTypes) > 0 {
		whereParts = append(whereParts, "m.move_type = ANY($"+strconv.Itoa(argPos)+"::int[])")
		args = append(args, moveTypes)
		argPos++
	}
	whereSQL := strings.Join(whereParts, " AND ")

	countQ := `SELECT COUNT(*) FROM geokrety_stats.user_points_log pl LEFT JOIN geokrety.gk_moves m ON m.id = pl.move_id WHERE ` + whereSQL
	var total int64
	_ = h.DB.QueryRow(c.Request.Context(), countQ, args...).Scan(&total)

	q := `
		SELECT pl.id, pl.user_id, u.username, pic.bucket, pic.key,
		       pl.move_id, m.move_type, m.waypoint, m.country,
		       pl.label, pl.points, pl.awarded_at
		FROM geokrety_stats.user_points_log pl
		LEFT JOIN geokrety.gk_moves m ON m.id = pl.move_id
		LEFT JOIN geokrety.gk_users u ON u.id = pl.user_id
		LEFT JOIN geokrety.gk_pictures pic ON pic.id = u.avatar
		WHERE ` + whereSQL + `
		ORDER BY ` + orderBy + `
		LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)

	queryArgs := append(args, perPage, offset)
	rows, err := h.DB.Query(c.Request.Context(), q, queryArgs...)
	if err != nil {
		errInternal(c, err)
		return
	}
	defer rows.Close()

	type row struct {
		ID          int64      `json:"id"`
		UserID      int64      `json:"user_id"`
		Username    *string    `json:"username,omitempty"`
		MoveID      *int64     `json:"move_id,omitempty"`
		MoveType    *int       `json:"move_type,omitempty"`
		TypeName    string     `json:"type_name"`
		Waypoint    *string    `json:"waypoint,omitempty"`
		Country     *string    `json:"country,omitempty"`
		Label       string     `json:"label"`
		Points      float64    `json:"points"`
		AwardedAt   *time.Time `json:"awarded_at,omitempty"`
		AuthorAvatar *string   `json:"author_avatar,omitempty"`
	}

	out2 := make([]row, 0)
	for rows.Next() {
		var r row
		var avatarBucket sql.NullString
		var avatarKey sql.NullString
		if err := rows.Scan(&r.ID, &r.UserID, &r.Username, &avatarBucket, &avatarKey,
			&r.MoveID, &r.MoveType, &r.Waypoint, &r.Country,
			&r.Label, &r.Points, &r.AwardedAt); err != nil {
			errInternal(c, err)
			return
		}
		r.AuthorAvatar = avatarRef(avatarBucket, avatarKey)
		if r.MoveType != nil {
			r.TypeName = moveTypeName(*r.MoveType)
		} else {
			r.TypeName = moveTypeName(-1)
		}
		out2 = append(out2, r)
	}
	if rows.Err() != nil {
		errInternal(c, rows.Err())
		return
	}

	hasNext := int64(offset+perPage) < total
	ok(c, out2, models.Meta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasNext: hasNext,
		HasPrev: page > 1,
	}, buildLinks(c, page, perPage, hasNext))
}

