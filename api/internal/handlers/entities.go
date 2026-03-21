package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/geokrety/geokrety-stats-api/internal/gkid"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var waypointCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{2,32}$`)
var decimalIdentifierPattern = regexp.MustCompile(`^[0-9]+$`)

type geoJSONFeature struct {
	Type       string                   `json:"type" xml:"type"`
	Geometry   interface{}              `json:"geometry" xml:"geometry"`
	Properties geoJSONFeatureProperties `json:"properties" xml:"properties"`
}

type geoJSONFeatureCollection struct {
	Type     string           `json:"type" xml:"type"`
	Features []geoJSONFeature `json:"features" xml:"features>feature"`
}

type geoJSONFeatureProperties struct {
	MoveID       int64     `json:"moveId" xml:"moveId"`
	MovedOn      time.Time `json:"movedOn" xml:"movedOn"`
	MoveType     int16     `json:"moveType" xml:"moveType"`
	MoveTypeName string    `json:"moveTypeName" xml:"moveTypeName"`
	Country      *string   `json:"country" xml:"country,omitempty"`
	Waypoint     *string   `json:"waypoint" xml:"waypoint,omitempty"`
}

func (h *StatsHandler) GetGeokretyDetailsById(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	geokretID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.store.FetchGeokrety(r.Context(), geokretID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch geokret")
		return
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) GetGeokrety(w http.ResponseWriter, r *http.Request) {
	h.GetGeokretyDetailsById(w, r)
}

func (h *StatsHandler) GetGeokretyList(w http.ResponseWriter, r *http.Request) {
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyList(r.Context(), limit, offset)
	}, "failed to fetch geokrety")
}

func (h *StatsHandler) GetGeokretyDetailsByGkId(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	gkid, ok := parsePublicGKIDParam(w, r, "gkid", "id")
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

func (h *StatsHandler) GetGeokretyMoves(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyMoves(r.Context(), geokretID, limit, offset)
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
	row, err := h.store.FetchGeokretyMoveDetails(r.Context(), geokretID, moveID)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch geokret move")
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

func (h *StatsHandler) GetGeokretyLoves(w http.ResponseWriter, r *http.Request) {
	h.GetGeokretyLovedBy(w, r)
}

func (h *StatsHandler) GetGeokretyWatches(w http.ResponseWriter, r *http.Request) {
	h.GetGeokretyWatchedBy(w, r)
}

func (h *StatsHandler) GetGeokretyPictures(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyPictures(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret pictures")
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

func (h *StatsHandler) GetGeokretyStatsMapCountries(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyStatsMapCountries(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret map countries")
}

func (h *StatsHandler) GetGeokretyWorldChoropleth(w http.ResponseWriter, r *http.Request) {
	h.GetGeokretyStatsMapCountries(w, r)
}

func (h *StatsHandler) GetGeokretyStatsElevation(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyStatsElevation(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret elevation")
}

func (h *StatsHandler) GetGeokretyStatsHeatmapDays(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretyStatsHeatmapDays(r.Context(), geokretID, limit, offset)
	}, "failed to fetch geokret heatmap days")
}

func (h *StatsHandler) GetStatsDormancy(w http.ResponseWriter, r *http.Request) {
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchStatsDormancy(r.Context(), limit, offset)
	}, "failed to fetch dormancy records")
}

func (h *StatsHandler) GetStatsMultiplierVelocity(w http.ResponseWriter, r *http.Request) {
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchStatsMultiplierVelocity(r.Context(), limit, offset)
	}, "failed to fetch multiplier velocity records")
}

func (h *StatsHandler) GetGeokretyGeoJSONTrip(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	geokretID, ok := h.parseGeokretRouteID(w, r)
	if !ok {
		return
	}
	limit := queryInt(r, "limit", 500, 1, 5000)
	offset := queryInt(r, "offset", 0, 0, 1_000_000)
	rows, err := h.store.FetchGeokretyTripPoints(r.Context(), geokretID, limit, offset)
	if err != nil {
		h.writeStoreError(w, r, err, "failed to fetch geokret trip")
		return
	}
	features := make([]geoJSONFeature, 0, len(rows))
	for _, row := range rows {
		features = append(features, geoJSONFeature{
			Type:     "Feature",
			Geometry: row.GeoJSON,
			Properties: geoJSONFeatureProperties{
				MoveID:       row.MoveID,
				MovedOn:      row.MovedOn,
				MoveType:     row.MoveType,
				MoveTypeName: row.MoveTypeName,
				Country:      row.Country,
				Waypoint:     row.Waypoint,
			},
		})
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, geoJSONFeatureCollection{Type: "FeatureCollection", Features: features}, started, limit, offset, len(features))
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

func (h *StatsHandler) GetCountrySpottedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.GetCountryGeokrety(w, r)
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

func (h *StatsHandler) GetWaypointSpottedGeokrety(w http.ResponseWriter, r *http.Request) {
	h.GetWaypointCurrentGeokrety(w, r)
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
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserList(r.Context(), limit, offset)
	}, "failed to fetch users")
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
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserPictures(r.Context(), userID, limit, offset)
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

func (h *StatsHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query, ok := parseSearchQuery(w, r)
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.SearchUsers(r.Context(), query, limit, offset)
	}, "failed to search users")
}

func (h *StatsHandler) GetUserStatsHeatmapDays(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserStatsHeatmapDays(r.Context(), userID, limit, offset)
	}, "failed to fetch user day heatmap")
}

func (h *StatsHandler) GetUserStatsContinentCoverage(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserStatsContinentCoverage(r.Context(), userID, limit, offset)
	}, "failed to fetch user continent coverage")
}

func (h *StatsHandler) GetUserStatsHeatmapHours(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserStatsHeatmapHours(r.Context(), userID, limit, offset)
	}, "failed to fetch user hour heatmap")
}

func (h *StatsHandler) GetUserStatsMapCountries(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchUserStatsMapCountries(r.Context(), userID, limit, offset)
	}, "failed to fetch user map countries")
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

func (h *StatsHandler) GetPicture(w http.ResponseWriter, r *http.Request) {
	h.GetPictureDetails(w, r)
}

func (h *StatsHandler) GetPictureList(w http.ResponseWriter, r *http.Request) {
	h.getEntityList(w, r, func(limit, offset int) (interface{}, error) {
		return h.store.FetchPictureList(r.Context(), limit, offset)
	}, "failed to fetch pictures")
}

func (h *StatsHandler) parseGeokretRouteID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	gkid, ok := parsePublicGKIDParam(w, r, "gkid", "id")
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

func (h *StatsHandler) getEntityList(w http.ResponseWriter, r *http.Request, fetch func(limit, offset int) (interface{}, error), errMsg string) {
	started := time.Now()
	limit := queryInt(r, "limit", 20, 1, 1000)
	offset := queryInt(r, "offset", 0, 0, 1_000_000)
	rows, err := fetch(limit, offset)
	if err != nil {
		h.writeStoreError(w, r, err, errMsg)
		return
	}
	count := 0
	v := reflect.ValueOf(rows)
	if v.IsValid() && v.Kind() == reflect.Slice {
		count = v.Len()
	}
	writeEnvelopeForRequest(w, r, http.StatusOK, rows, started, limit, offset, count)
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

func (h *StatsHandler) writeStoreError(w http.ResponseWriter, r *http.Request, err error, message string) {
	if errors.Is(err, sql.ErrNoRows) {
		writeErrorForRequest(w, r, http.StatusNotFound, "resource not found")
		return
	}
	h.logger.Error(message, zap.Error(err))
	writeErrorForRequest(w, r, http.StatusInternalServerError, message)
}

func parseInt64PathParam(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value := chi.URLParam(r, key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid path parameter %q", key))
		return 0, false
	}
	return parsed, true
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
	parsed, err := gkid.New(value)
	if err != nil {
		writeErrorForRequest(w, r, http.StatusBadRequest, err.Error())
		return 0, false
	}
	return parsed.Int(), true
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

func parseWaypointCodeParam(w http.ResponseWriter, r *http.Request, key string) (string, bool) {
	value := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, key)))
	if !waypointCodePattern.MatchString(value) {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid path parameter %q", key))
		return "", false
	}
	return value, true
}

func parseSearchQuery(w http.ResponseWriter, r *http.Request) (string, bool) {
	value := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(value) < 2 {
		writeErrorForRequest(w, r, http.StatusBadRequest, "query parameter \"q\" must be at least 2 characters")
		return "", false
	}
	return value, true
}
