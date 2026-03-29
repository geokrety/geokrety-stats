package handlers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	"github.com/go-chi/chi/v5"
)

var waypointCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{2,32}$`)

func (h *StatsHandler) GetGeokretyList(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	gkids, ok := parsePublicGKIDListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(gkids) > 0 {
		rows, err := h.store.FetchGeokretyByGKIDs(r.Context(), gkids)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch geokrety by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyList(r.Context(), limit, offset)
	}, "failed to fetch geokrety")
}

func (h *StatsHandler) GetGeokretyDetailsByGkId(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	gkid, ok := parsePublicGKIDParam(w, r, "gkid")
	if !ok {
		return
	}
	row, err := h.store.FetchGeokretyByGKID(r.Context(), gkid)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch geokret")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) SearchGeokrety(w http.ResponseWriter, r *http.Request) {
	query, ok := parseSearchQuery(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.SearchGeokrety(r.Context(), query, limit, offset)
	}, "failed to search geokrety")
}

func (h *StatsHandler) GetGeokretStats(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	row, err := h.store.FetchGeokretStats(r.Context(), geokretID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch geokret stats")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetMoveList(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	filters, ok := h.parseMoveFilters(w, r, nil)
	if !ok {
		return
	}
	moveIDs, ok := parseInt64ListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(moveIDs) > 0 {
		rows, err := h.store.FetchMoveListByIDs(r.Context(), filters, moveIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch moves by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchMoveList(r.Context(), filters, limit, offset)
	}, "failed to fetch moves")
}

func (h *StatsHandler) GetMoveDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	moveID, ok := parseInt64PathParam(w, r, "id")
	if !ok {
		return
	}
	row, err := h.store.FetchMove(r.Context(), moveID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch move")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetGeokretyMoves(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	filters, ok := h.parseMoveFilters(w, r, &geokretID)
	if !ok {
		return
	}
	moveIDs, ok := parseInt64ListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(moveIDs) > 0 {
		rows, err := h.store.FetchMoveListByIDs(r.Context(), filters, moveIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch geokret moves by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchMoveList(r.Context(), filters, limit, offset)
	}, "failed to fetch geokret moves")
}

func (h *StatsHandler) GetGeokretyMoveDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	moveID, ok := parseInt64PathParam(w, r, "moveId")
	if !ok {
		return
	}
	row, err := h.store.FetchMove(r.Context(), moveID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch geokret move")
		return
	}
	if row.GeokretID != geokretID {
		writeErrorForRequest(w, r, http.StatusNotFound, "resource not found")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetGeokretyLovedBy(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyLoves(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret lovers")
}

func (h *StatsHandler) GetGeokretyWatchedBy(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyWatches(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret watchers")
}

func (h *StatsHandler) GetGeokretyFinders(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyFinders(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret finders")
}

func (h *StatsHandler) GetGeokretyPictures(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	filters := db.PictureFilters{GeokretID: &geokretID}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchPictureList(r.Context(), filters, limit, offset)
	}, "failed to fetch geokret pictures")
}

func (h *StatsHandler) GetGeokretyCountries(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyCountries(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret countries")
}

func (h *StatsHandler) GetGeokretyWaypoints(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyWaypoints(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret waypoints")
}

func (h *StatsHandler) GetCountryDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	code, ok := parseCountryCodeParam(w, r, "code")
	if !ok {
		return
	}
	row, err := h.store.FetchCountryDetails(r.Context(), code)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch country details")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetCountryList(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	codes, ok := parseCountryCodeListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(codes) > 0 {
		rows, err := h.store.FetchCountryListByCodes(r.Context(), codes)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch countries by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchCountryList(r.Context(), limit, offset)
	}, "failed to fetch countries")
}

func (h *StatsHandler) GetCountryGeokrety(w http.ResponseWriter, r *http.Request) {
	code, ok := parseCountryCodeParam(w, r, "code")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchCountryGeokrety(r.Context(), code, limit, offset)
	}, "failed to fetch country geokrety")
}

func (h *StatsHandler) GetWaypoint(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	code, ok := parseWaypointCodeParam(w, r, "code")
	if !ok {
		return
	}
	row, err := h.store.FetchWaypoint(r.Context(), code)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch waypoint")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetWaypointCurrentGeokrety(w http.ResponseWriter, r *http.Request) {
	code, ok := parseWaypointCodeParam(w, r, "code")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchWaypointCurrentGeokrety(r.Context(), code, limit, offset)
	}, "failed to fetch current waypoint geokrety")
}

func (h *StatsHandler) GetWaypointPastGeokrety(w http.ResponseWriter, r *http.Request) {
	code, ok := parseWaypointCodeParam(w, r, "code")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchWaypointPastGeokrety(r.Context(), code, limit, offset)
	}, "failed to fetch past waypoint geokrety")
}

func (h *StatsHandler) SearchWaypoints(w http.ResponseWriter, r *http.Request) {
	query, ok := parseSearchQuery(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.SearchWaypoints(r.Context(), query, limit, offset)
	}, "failed to search waypoints")
}

func (h *StatsHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.store.FetchUserDetails(r.Context(), userID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch user details")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetUserList(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	userIDs, ok := parseInt64ListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(userIDs) > 0 {
		rows, err := h.store.FetchUserListByIDs(r.Context(), userIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch users by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserList(r.Context(), limit, offset)
	}, "failed to fetch users")
}

func (h *StatsHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query, ok := parseSearchQuery(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.SearchUsers(r.Context(), query, limit, offset)
	}, "failed to search users")
}

func (h *StatsHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.store.FetchUserStats(r.Context(), userID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch user stats")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetUserOwnedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userListHandler(w, r, h.store.FetchUserOwnedGeokrety, "failed to fetch owned geokrety")
}

func (h *StatsHandler) GetUserFoundGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userListHandler(w, r, h.store.FetchUserFoundGeokrety, "failed to fetch found geokrety")
}

func (h *StatsHandler) GetUserLovedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userListHandler(w, r, h.store.FetchUserLovedGeokrety, "failed to fetch loved geokrety")
}

func (h *StatsHandler) GetUserWatchedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userListHandler(w, r, h.store.FetchUserWatchedGeokrety, "failed to fetch watched geokrety")
}

func (h *StatsHandler) GetUserPictures(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	filters := db.PictureFilters{UserID: &userID}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchPictureList(r.Context(), filters, limit, offset)
	}, "failed to fetch user pictures")
}

func (h *StatsHandler) GetUserCountries(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserCountries(r.Context(), userID, limit, offset)
	}, "failed to fetch user countries")
}

func (h *StatsHandler) GetUserWaypoints(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserWaypoints(r.Context(), userID, limit, offset)
	}, "failed to fetch user waypoints")
}

func (h *StatsHandler) GetPictureDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	pictureID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.store.FetchPicture(r.Context(), pictureID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch picture")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetPictureList(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	filters, ok := h.parsePictureFilters(w, r, nil, nil, nil)
	if !ok {
		return
	}
	pictureIDs, ok := parseInt64ListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(pictureIDs) > 0 {
		rows, err := h.store.FetchPictureListByIDs(r.Context(), pictureIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch pictures by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchPictureList(r.Context(), filters, limit, offset)
	}, "failed to fetch pictures")
}

func (h *StatsHandler) parseGeokretRouteID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	gkid, ok := parsePublicGKIDParam(w, r, "gkid")
	if !ok {
		return 0, false
	}
	geokretID, err := h.store.ResolveGeokretID(r.Context(), gkid)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to resolve geokret")
		return 0, false
	}
	return geokretID, true
}

func (h *StatsHandler) userListHandler(w http.ResponseWriter, r *http.Request, fetch func(context.Context, int64, int, int) ([]db.GeokretListItem, error), errMsg string) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return fetch(r.Context(), userID, limit, offset)
	}, errMsg)
}

func parsePublicGKIDParam(w http.ResponseWriter, r *http.Request, keys ...string) (int64, bool) {
	value := ""
	for _, key := range keys {
		value = strings.TrimSpace(chi.URLParam(r, key))
		if value != "" {
			break
		}
	}
	if value == "" {
		writeErrorForRequest(w, r, http.StatusBadRequest, "missing geokret identifier")
		return 0, false
	}
	parsed, err := geokrety.New(value)
	if err != nil {
		writeErrorForRequest(w, r, http.StatusBadRequest, err.Error())
		return 0, false
	}
	return parsed.Int(), true
}

func parseOptionalPublicGKIDQuery(w http.ResponseWriter, r *http.Request, key string) (int64, bool, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, false, true
	}
	parsed, err := geokrety.New(raw)
	if err != nil {
		writeErrorForRequest(w, r, http.StatusBadRequest, err.Error())
		return 0, true, false
	}
	return parsed.Int(), true, true
}

func parseCountryCodeParam(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	value := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, key)))
	if len(value) != 2 {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid path parameter %q", key))
		return "", false
	}
	for _, ch := range value {
		if ch < 'A' || ch > 'Z' {
			writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid path parameter %q", key))
			return "", false
		}
	}
	return value, true
}

func parseOptionalCountryCodeQuery(w http.ResponseWriter, r *http.Request, key string) (*string, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, true
	}
	value := strings.ToUpper(raw)
	if len(value) != 2 {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
		return nil, false
	}
	for _, ch := range value {
		if ch < 'A' || ch > 'Z' {
			writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
			return nil, false
		}
	}
	return &value, true
}

func parseWaypointCodeParam(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	value := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, key)))
	if !waypointCodePattern.MatchString(value) {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid path parameter %q", key))
		return "", false
	}
	return value, true
}

func parseOptionalWaypointCodeQuery(w http.ResponseWriter, r *http.Request, key string) (*string, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, true
	}
	value := strings.ToUpper(raw)
	if !waypointCodePattern.MatchString(value) {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
		return nil, false
	}
	return &value, true
}

func parseSearchQuery(w http.ResponseWriter, r *http.Request) (string, bool) {
	value := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(value) < 2 {
		writeErrorForRequest(w, r, http.StatusBadRequest, "query parameter \"q\" must be at least 2 characters")
		return "", false
	}
	return value, true
}

func parseOptionalInt64Query(w http.ResponseWriter, r *http.Request, key string) (*int64, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, true
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
		return nil, false
	}
	return &parsed, true
}

func parseOptionalDateQuery(w http.ResponseWriter, r *http.Request, key string) (*time.Time, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, true
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
		return nil, false
	}
	return &parsed, true
}

func (h *StatsHandler) parseMoveFilters(w http.ResponseWriter, r *http.Request, fixedGeokretID *int64) (db.MoveFilters, bool) {
	filters := db.MoveFilters{GeokretID: fixedGeokretID}
	if filters.GeokretID == nil {
		gkid, found, ok := parseOptionalPublicGKIDQuery(w, r, "geokret")
		if !ok {
			return db.MoveFilters{}, false
		}
		if found {
			geokretID, err := h.store.ResolveGeokretID(r.Context(), gkid)
			if err != nil {
				h.writeStoreError(w, r, err, "failed to resolve geokret")
				return db.MoveFilters{}, false
			}
			filters.GeokretID = &geokretID
		}
	}
	userID, ok := parseOptionalInt64Query(w, r, "user")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.UserID = userID
	country, ok := parseOptionalCountryCodeQuery(w, r, "country")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.Country = country
	waypoint, ok := parseOptionalWaypointCodeQuery(w, r, "waypoint")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.Waypoint = waypoint
	dateFrom, ok := parseOptionalDateQuery(w, r, "date_from")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.DateFrom = dateFrom
	dateTo, ok := parseOptionalDateQuery(w, r, "date_to")
	if !ok {
		return db.MoveFilters{}, false
	}
	if dateTo != nil {
		exclusive := dateTo.AddDate(0, 0, 1)
		filters.DateTo = &exclusive
	}
	return filters, true
}

func (h *StatsHandler) parsePictureFilters(w http.ResponseWriter, r *http.Request, fixedGeokretID, fixedMoveID, fixedUserID *int64) (db.PictureFilters, bool) {
	filters := db.PictureFilters{GeokretID: fixedGeokretID, MoveID: fixedMoveID, UserID: fixedUserID}
	if filters.GeokretID == nil {
		gkid, found, ok := parseOptionalPublicGKIDQuery(w, r, "geokret")
		if !ok {
			return db.PictureFilters{}, false
		}
		if found {
			geokretID, err := h.store.ResolveGeokretID(r.Context(), gkid)
			if err != nil {
				h.writeStoreError(w, r, err, "failed to resolve geokret")
				return db.PictureFilters{}, false
			}
			filters.GeokretID = &geokretID
		}
	}
	if filters.MoveID == nil {
		moveID, ok := parseOptionalInt64Query(w, r, "move")
		if !ok {
			return db.PictureFilters{}, false
		}
		filters.MoveID = moveID
	}
	if filters.UserID == nil {
		userID, ok := parseOptionalInt64Query(w, r, "user")
		if !ok {
			return db.PictureFilters{}, false
		}
		filters.UserID = userID
	}
	return filters, true
}

func parseInt64ListQuery(w http.ResponseWriter, r *http.Request, key string) ([]int64, bool) {
	parts := splitCSVQuery(r, key)
	if len(parts) == 0 {
		return nil, true
	}
	values := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		parsed, err := strconv.ParseInt(part, 10, 64)
		if err != nil || parsed <= 0 {
			writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
			return nil, false
		}
		if _, ok := seen[parsed]; ok {
			continue
		}
		seen[parsed] = struct{}{}
		values = append(values, parsed)
	}
	return values, true
}

func parseCountryCodeListQuery(w http.ResponseWriter, r *http.Request, key string) ([]string, bool) {
	parts := splitCSVQuery(r, key)
	if len(parts) == 0 {
		return nil, true
	}
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := strings.ToUpper(part)
		if len(value) != 2 {
			writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
			return nil, false
		}
		for _, ch := range value {
			if ch < 'A' || ch > 'Z' {
				writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", key))
				return nil, false
			}
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values, true
}

func parsePublicGKIDListQuery(w http.ResponseWriter, r *http.Request, key string) ([]int64, bool) {
	parts := splitCSVQuery(r, key)
	if len(parts) == 0 {
		return nil, true
	}
	values := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		parsed, err := geokrety.New(part)
		if err != nil {
			writeErrorForRequest(w, r, http.StatusBadRequest, err.Error())
			return nil, false
		}
		value := parsed.Int()
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values, true
}

func splitCSVQuery(r *http.Request, key string) []string {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil
	}
	segments := strings.Split(raw, ",")
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		trimmed := strings.TrimSpace(segment)
		if trimmed == "" {
			continue
		}
		parts = append(parts, trimmed)
	}
	return parts
}
