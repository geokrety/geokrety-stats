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
	filters, ok := parseGeokretListFilters(w, r)
	if !ok {
		return
	}
	gkids, ok := parsePublicGKIDListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(gkids) > 0 {
		rows, err := h.geokrety.FetchGeokretyByGKIDs(r.Context(), gkids)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch geokrety by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.geokrety.FetchGeokretyList(r.Context(), filters, limit, offset)
	}, "failed to fetch geokrety")
}

func (h *StatsHandler) GetGeokretyDetailsByGkId(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	gkid, ok := parsePublicGKIDParam(w, r, "gkid")
	if !ok {
		return
	}
	row, err := h.geokrety.FetchGeokretyByGKID(r.Context(), gkid)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch geokret")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetGeokretStats(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	row, err := h.geokrety.FetchGeokretStats(r.Context(), geokretID)
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
		rows, err := h.moves.FetchMoveListByIDs(r.Context(), filters, moveIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch moves by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.moves.FetchMoveList(r.Context(), filters, limit, offset)
	}, "failed to fetch moves")
}

func (h *StatsHandler) GetMoveDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	moveID, ok := parseInt64PathParam(w, r, "id")
	if !ok {
		return
	}
	row, err := h.moves.FetchMove(r.Context(), moveID)
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
		rows, err := h.moves.FetchMoveListByIDs(r.Context(), filters, moveIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch geokret moves by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.moves.FetchMoveList(r.Context(), filters, limit, offset)
	}, "failed to fetch geokret moves")
}

func (h *StatsHandler) GetGeokretyLovedBy(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	sort, ok := parseSocialSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.geokrety.FetchGeokretyLoves(r.Context(), geokretID, sort, limit, offset)
	}, "failed to fetch geokret lovers")
}

func (h *StatsHandler) GetGeokretyWatchedBy(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	sort, ok := parseSocialSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.geokrety.FetchGeokretyWatches(r.Context(), geokretID, sort, limit, offset)
	}, "failed to fetch geokret watchers")
}

func (h *StatsHandler) GetGeokretyFinders(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	sort, ok := parseSocialSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.geokrety.FetchGeokretyFinders(r.Context(), geokretID, sort, limit, offset)
	}, "failed to fetch geokret finders")
}

func (h *StatsHandler) GetGeokretyPictures(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	filters, ok := h.parsePictureFilters(w, r, &geokretID, nil, nil)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.pictures.FetchPictureList(r.Context(), filters, limit, offset)
	}, "failed to fetch geokret pictures")
}

func (h *StatsHandler) GetGeokretyCountries(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	sort, ok := parseGeokretCountryVisitSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.geokrety.FetchGeokretyCountries(r.Context(), geokretID, sort, limit, offset)
	}, "failed to fetch geokret countries")
}

func (h *StatsHandler) GetGeokretyWaypoints(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	sort, ok := parseWaypointVisitSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.geokrety.FetchGeokretyWaypoints(r.Context(), geokretID, sort, limit, offset)
	}, "failed to fetch geokret waypoints")
}

func (h *StatsHandler) GetCountryDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	code, ok := parseCountryCodeParam(w, r, "code")
	if !ok {
		return
	}
	row, err := h.countries.FetchCountryDetails(r.Context(), code)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch country details")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetCountryList(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	filters, ok := parseCountryListFilters(w, r)
	if !ok {
		return
	}
	codes, ok := parseCountryCodeListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(codes) > 0 {
		rows, err := h.countries.FetchCountryListByCodes(r.Context(), codes)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch countries by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.countries.FetchCountryList(r.Context(), filters, limit, offset)
	}, "failed to fetch countries")
}

func (h *StatsHandler) GetCountryGeokrety(w http.ResponseWriter, r *http.Request) {
	code, ok := parseCountryCodeParam(w, r, "code")
	if !ok {
		return
	}
	sort, ok := parseGeokretSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.countries.FetchCountryGeokrety(r.Context(), code, sort, limit, offset)
	}, "failed to fetch country geokrety")
}

func (h *StatsHandler) GetWaypoint(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	code, ok := parseWaypointCodeParam(w, r, "code")
	if !ok {
		return
	}
	row, err := h.waypoints.FetchWaypoint(r.Context(), code)
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
	sort, ok := parseGeokretSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.waypoints.FetchWaypointCurrentGeokrety(r.Context(), code, sort, limit, offset)
	}, "failed to fetch current waypoint geokrety")
}

func (h *StatsHandler) GetWaypointPastGeokrety(w http.ResponseWriter, r *http.Request) {
	code, ok := parseWaypointCodeParam(w, r, "code")
	if !ok {
		return
	}
	sort, ok := parseGeokretSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.waypoints.FetchWaypointPastGeokrety(r.Context(), code, sort, limit, offset)
	}, "failed to fetch past waypoint geokrety")
}

func (h *StatsHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.users.FetchUserDetails(r.Context(), userID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch user details")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetUserList(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	filters, ok := parseUserListFilters(w, r)
	if !ok {
		return
	}
	userIDs, ok := parseInt64ListQuery(w, r, "ids")
	if !ok {
		return
	}
	if len(userIDs) > 0 {
		rows, err := h.users.FetchUserListByIDs(r.Context(), userIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch users by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.users.FetchUserList(r.Context(), filters, limit, offset)
	}, "failed to fetch users")
}

func (h *StatsHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.users.FetchUserStats(r.Context(), userID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch user stats")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetUserOwnedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userGeokretyListHandler(w, r, h.users.FetchUserOwnedGeokrety, "failed to fetch owned geokrety")
}

func (h *StatsHandler) GetUserFoundGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userGeokretyListHandler(w, r, h.users.FetchUserFoundGeokrety, "failed to fetch found geokrety")
}

func (h *StatsHandler) GetUserLovedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userGeokretyListHandler(w, r, h.users.FetchUserLovedGeokrety, "failed to fetch loved geokrety")
}

func (h *StatsHandler) GetUserWatchedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.userGeokretyListHandler(w, r, h.users.FetchUserWatchedGeokrety, "failed to fetch watched geokrety")
}

func (h *StatsHandler) GetUserPictures(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	filters, ok := h.parsePictureFilters(w, r, nil, nil, &userID)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.pictures.FetchPictureList(r.Context(), filters, limit, offset)
	}, "failed to fetch user pictures")
}

func (h *StatsHandler) GetUserCountries(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	sort, ok := parseUserCountryVisitSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.users.FetchUserCountries(r.Context(), userID, sort, limit, offset)
	}, "failed to fetch user countries")
}

func (h *StatsHandler) GetUserWaypoints(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	sort, ok := parseWaypointVisitSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.users.FetchUserWaypoints(r.Context(), userID, sort, limit, offset)
	}, "failed to fetch user waypoints")
}

func (h *StatsHandler) GetPictureDetails(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	pictureID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.pictures.FetchPicture(r.Context(), pictureID)
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
		rows, err := h.pictures.FetchPictureListByIDs(r.Context(), pictureIDs)
		if err != nil {
			h.writeStoreError(w, r, err, "failed to fetch pictures by ids")
			return
		}
		writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, len(rows), 0, len(rows))
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.pictures.FetchPictureList(r.Context(), filters, limit, offset)
	}, "failed to fetch pictures")
}

func (h *StatsHandler) parseGeokretRouteID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	gkid, ok := parsePublicGKIDParam(w, r, "gkid")
	if !ok {
		return 0, false
	}
	geokretID, err := h.geokrety.ResolveGeokretID(r.Context(), gkid)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to resolve geokret")
		return 0, false
	}
	return geokretID, true
}

func (h *StatsHandler) userGeokretyListHandler(w http.ResponseWriter, r *http.Request, fetch func(context.Context, int64, db.Sort, int, int) ([]db.GeokretListItem, error), errMsg string) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	sort, ok := parseGeokretSort(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return fetch(r.Context(), userID, sort, limit, offset)
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

func queryParamValue(r *http.Request, keys ...string) string {
	if r == nil || r.URL == nil {
		return ""
	}
	for _, key := range keys {
		raw := strings.TrimSpace(r.URL.Query().Get(key))
		if raw != "" {
			return raw
		}
	}
	return ""
}

func parseOptionalPublicGKIDQuery(w http.ResponseWriter, r *http.Request, keys ...string) (int64, bool, bool) {
	raw := queryParamValue(r, keys...)
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

func parseOptionalCountryCodeQuery(w http.ResponseWriter, r *http.Request, keys ...string) (*string, bool) {
	raw := queryParamValue(r, keys...)
	if raw == "" {
		return nil, true
	}
	value := strings.ToUpper(raw)
	if len(value) != 2 {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", keys[0]))
		return nil, false
	}
	for _, ch := range value {
		if ch < 'A' || ch > 'Z' {
			writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", keys[0]))
			return nil, false
		}
	}
	return &value, true
}

func parseWaypointCodeParam(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	value := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, key)))
	if value == "SEARCH" {
		writeErrorForRequest(w, r, http.StatusNotFound, "resource not found")
		return "", false
	}
	if !waypointCodePattern.MatchString(value) {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid path parameter %q", key))
		return "", false
	}
	return value, true
}

func parseOptionalWaypointCodeQuery(w http.ResponseWriter, r *http.Request, keys ...string) (*string, bool) {
	raw := queryParamValue(r, keys...)
	if raw == "" {
		return nil, true
	}
	value := strings.ToUpper(raw)
	if !waypointCodePattern.MatchString(value) {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", keys[0]))
		return nil, false
	}
	return &value, true
}

func parseOptionalStringQuery(w http.ResponseWriter, r *http.Request, minLength int, keys ...string) (*string, bool) {
	raw := queryParamValue(r, keys...)
	if raw == "" {
		return nil, true
	}
	if len(raw) < minLength {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("query parameter %q must be at least %d characters", keys[0], minLength))
		return nil, false
	}
	return &raw, true
}

func parseOptionalInt64Query(w http.ResponseWriter, r *http.Request, keys ...string) (*int64, bool) {
	raw := queryParamValue(r, keys...)
	if raw == "" {
		return nil, true
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", keys[0]))
		return nil, false
	}
	return &parsed, true
}

func parseOptionalDateQuery(w http.ResponseWriter, r *http.Request, keys ...string) (*time.Time, bool) {
	raw := queryParamValue(r, keys...)
	if raw == "" {
		return nil, true
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", keys[0]))
		return nil, false
	}
	return &parsed, true
}

func parseSortQuery(w http.ResponseWriter, r *http.Request, defaultSort db.Sort, allowed ...db.Sort) (db.Sort, bool) {
	value := queryParamValue(r, "sort")
	if value == "" {
		return defaultSort, true
	}
	for _, candidate := range allowed {
		if value == candidate.String() {
			return candidate, true
		}
	}
	writeErrorForRequest(w, r, http.StatusBadRequest, "invalid query parameter \"sort\"")
	return db.Sort{}, false
}

func parseGeokretSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.DescSort("last_move_at"), db.AscSort("name"), db.DescSort("name"), db.AscSort("born_at"), db.DescSort("born_at"), db.AscSort("last_move_at"), db.DescSort("last_move_at"))
}

func parseMoveSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.DescSort("date"), db.AscSort("date"), db.DescSort("date"), db.AscSort("id"), db.DescSort("id"))
}

func parseCountryListSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.AscSort("code"), db.AscSort("code"), db.DescSort("code"), db.AscSort("name"), db.DescSort("name"))
}

func parseUserSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.DescSort("joined_at"), db.AscSort("username"), db.DescSort("username"), db.AscSort("joined_at"), db.DescSort("joined_at"), db.AscSort("last_move_at"), db.DescSort("last_move_at"))
}

func parsePictureSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.DescSort("created_on"), db.AscSort("created_on"), db.DescSort("created_on"), db.AscSort("id"), db.DescSort("id"))
}

func parseSocialSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	defaultSort := db.DescSort("at")
	attribute := "at"
	if r != nil && r.URL != nil {
		switch {
		case strings.HasSuffix(r.URL.Path, "/loved-by"):
			defaultSort = db.DescSort("loved_on_date")
			attribute = "loved_on_date"
		case strings.HasSuffix(r.URL.Path, "/watched-by"):
			defaultSort = db.DescSort("watched_on_date")
			attribute = "watched_on_date"
		case strings.HasSuffix(r.URL.Path, "/finders"):
			defaultSort = db.DescSort("found_on_date")
			attribute = "found_on_date"
		}
	}
	return parseSortQuery(w, r, defaultSort, db.AscSort(attribute), db.DescSort(attribute))
}

func parseGeokretCountryVisitSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.DescSort("first_visited_at"), db.AscSort("country_code"), db.DescSort("country_code"), db.AscSort("first_visited_at"), db.DescSort("first_visited_at"), db.AscSort("move_count"), db.DescSort("move_count"))
}

func parseUserCountryVisitSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.DescSort("last_visit"), db.AscSort("country_code"), db.DescSort("country_code"), db.AscSort("first_visit"), db.DescSort("first_visit"), db.AscSort("last_visit"), db.DescSort("last_visit"), db.AscSort("move_count"), db.DescSort("move_count"))
}

func parseWaypointVisitSort(w http.ResponseWriter, r *http.Request) (db.Sort, bool) {
	return parseSortQuery(w, r, db.DescSort("last_visited_at"), db.AscSort("waypoint_code"), db.DescSort("waypoint_code"), db.AscSort("first_visited_at"), db.DescSort("first_visited_at"), db.AscSort("last_visited_at"), db.DescSort("last_visited_at"), db.AscSort("visit_count"), db.DescSort("visit_count"))
}

func parseGeokretListFilters(w http.ResponseWriter, r *http.Request) (db.GeokretListFilters, bool) {
	name, ok := parseOptionalStringQuery(w, r, 2, "name", "filter[name]")
	if !ok {
		return db.GeokretListFilters{}, false
	}
	ownerID, ok := parseOptionalInt64Query(w, r, "owner", "owner_id", "filter[owner]", "filter[owner_id]")
	if !ok {
		return db.GeokretListFilters{}, false
	}
	countries, ok := parseCountryCodeListQueryKeys(w, r, "country", "filter[country]")
	if !ok {
		return db.GeokretListFilters{}, false
	}
	sort, ok := parseGeokretSort(w, r)
	if !ok {
		return db.GeokretListFilters{}, false
	}
	return db.GeokretListFilters{Name: name, OwnerID: ownerID, Countries: countries, Sort: sort}, true
}

func parseUserListFilters(w http.ResponseWriter, r *http.Request) (db.UserListFilters, bool) {
	username, ok := parseOptionalStringQuery(w, r, 2, "username", "filter[username]")
	if !ok {
		return db.UserListFilters{}, false
	}
	countries, ok := parseCountryCodeListQueryKeys(w, r, "country", "filter[country]")
	if !ok {
		return db.UserListFilters{}, false
	}
	sort, ok := parseUserSort(w, r)
	if !ok {
		return db.UserListFilters{}, false
	}
	return db.UserListFilters{Username: username, Countries: countries, Sort: sort}, true
}

func parseCountryListFilters(w http.ResponseWriter, r *http.Request) (db.CountryListFilters, bool) {
	sort, ok := parseCountryListSort(w, r)
	if !ok {
		return db.CountryListFilters{}, false
	}
	return db.CountryListFilters{Sort: sort}, true
}

func (h *StatsHandler) parseMoveFilters(w http.ResponseWriter, r *http.Request, fixedGeokretID *int64) (db.MoveFilters, bool) {
	filters := db.MoveFilters{GeokretID: fixedGeokretID}
	if filters.GeokretID == nil {
		gkid, found, ok := parseOptionalPublicGKIDQuery(w, r, "geokret", "filter[geokret]")
		if !ok {
			return db.MoveFilters{}, false
		}
		if found {
			geokretID, err := h.geokrety.ResolveGeokretID(r.Context(), gkid)
			if err != nil {
				h.writeStoreError(w, r, err, "failed to resolve geokret")
				return db.MoveFilters{}, false
			}
			filters.GeokretID = &geokretID
		}
	}
	userID, ok := parseOptionalInt64Query(w, r, "user", "filter[user]")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.UserID = userID
	country, ok := parseOptionalCountryCodeQuery(w, r, "country", "filter[country]")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.Country = country
	waypoint, ok := parseOptionalWaypointCodeQuery(w, r, "waypoint", "filter[waypoint]")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.Waypoint = waypoint
	dateFrom, ok := parseOptionalDateQuery(w, r, "date_from", "filter[date_from]")
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.DateFrom = dateFrom
	dateTo, ok := parseOptionalDateQuery(w, r, "date_to", "filter[date_to]")
	if !ok {
		return db.MoveFilters{}, false
	}
	if dateTo != nil {
		exclusive := dateTo.AddDate(0, 0, 1)
		filters.DateTo = &exclusive
	}
	sort, ok := parseMoveSort(w, r)
	if !ok {
		return db.MoveFilters{}, false
	}
	filters.Sort = sort
	return filters, true
}

func (h *StatsHandler) parsePictureFilters(w http.ResponseWriter, r *http.Request, fixedGeokretID, fixedMoveID, fixedUserID *int64) (db.PictureFilters, bool) {
	filters := db.PictureFilters{GeokretID: fixedGeokretID, MoveID: fixedMoveID, UserID: fixedUserID}
	if filters.GeokretID == nil {
		gkid, found, ok := parseOptionalPublicGKIDQuery(w, r, "geokret", "filter[geokret]")
		if !ok {
			return db.PictureFilters{}, false
		}
		if found {
			geokretID, err := h.geokrety.ResolveGeokretID(r.Context(), gkid)
			if err != nil {
				h.writeStoreError(w, r, err, "failed to resolve geokret")
				return db.PictureFilters{}, false
			}
			filters.GeokretID = &geokretID
		}
	}
	if filters.MoveID == nil {
		moveID, ok := parseOptionalInt64Query(w, r, "move", "filter[move]")
		if !ok {
			return db.PictureFilters{}, false
		}
		filters.MoveID = moveID
	}
	if filters.UserID == nil {
		userID, ok := parseOptionalInt64Query(w, r, "user", "filter[user]")
		if !ok {
			return db.PictureFilters{}, false
		}
		filters.UserID = userID
	}
	sort, ok := parsePictureSort(w, r)
	if !ok {
		return db.PictureFilters{}, false
	}
	filters.Sort = sort
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

func parseCountryCodeListQueryKeys(w http.ResponseWriter, r *http.Request, keys ...string) ([]string, bool) {
	parts := splitCSVQueryKeys(r, keys...)
	if len(parts) == 0 {
		return nil, true
	}
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := strings.ToUpper(part)
		if len(value) != 2 {
			writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", keys[0]))
			return nil, false
		}
		for _, ch := range value {
			if ch < 'A' || ch > 'Z' {
				writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid query parameter %q", keys[0]))
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
	return splitCSVQueryKeys(r, key)
}

func splitCSVQueryKeys(r *http.Request, keys ...string) []string {
	if r == nil || r.URL == nil {
		return nil
	}
	for _, key := range keys {
		raw := strings.TrimSpace(r.URL.Query().Get(key))
		if raw == "" {
			continue
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
		if len(parts) > 0 {
			return parts
		}
	}
	return nil
}
