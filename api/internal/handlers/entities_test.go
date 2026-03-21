package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func withRouteParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withRouteParams(r *http.Request, pairs ...string) *http.Request {
	rctx := chi.NewRouteContext()
	for i := 0; i < len(pairs); i += 2 {
		rctx.URLParams.Add(pairs[i], pairs[i+1])
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func decodeMap(t *testing.T, body *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return payload
}

func TestEntityHandlerSuccessEndpoints(t *testing.T) {
	store := &mockStatsStore{}
	h := NewStatsHandler(store, zap.NewNop())

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		target     string
		params     []string
		expMethod  string
		expStatus  int
		checkLimit bool
		expLimit   int
		expOffset  int
	}{
		{"geokret-list", h.GetGeokretyList, "/api/v3/geokrety/?limit=4&offset=1", nil, "FetchGeokretyList", http.StatusOK, true, 4, 1},
		{"geokret-details", h.GetGeokrety, "/api/v3/geokrety/1", []string{"id", "1"}, "FetchGeokrety", http.StatusOK, false, 0, 0},
		{"geokret-details-by-numeric-gkid", h.GetGeokretyDetailsByGkId, "/api/v3/geokrety/1", []string{"gkid", "1"}, "FetchGeokretyByGKID", http.StatusOK, false, 0, 0},
		{"geokret-details-by-gkid", h.GetGeokretyDetailsByGkId, "/api/v3/geokrety/GK0001", []string{"gkid", "GK0001"}, "FetchGeokretyByGKID", http.StatusOK, false, 0, 0},
		{"geokret-moves", h.GetGeokretyMoves, "/api/v3/geokrety/1/moves?limit=7&offset=2", []string{"id", "1"}, "FetchGeokretyMoves", http.StatusOK, true, 7, 2},
		{"geokret-move-details", h.GetGeokretyMoveDetails, "/api/v3/geokrety/1/moves/9", []string{"id", "1", "moveId", "9"}, "FetchGeokretyMoveDetails", http.StatusOK, false, 0, 0},
		{"geokret-loved-by", h.GetGeokretyLovedBy, "/api/v3/geokrety/1/loved-by?limit=8&offset=1", []string{"id", "1"}, "FetchGeokretyLoves", http.StatusOK, true, 8, 1},
		{"geokret-watched-by", h.GetGeokretyWatchedBy, "/api/v3/geokrety/1/watched-by?limit=8&offset=1", []string{"id", "1"}, "FetchGeokretyWatches", http.StatusOK, true, 8, 1},
		{"geokret-loves-alias", h.GetGeokretyLoves, "/api/v3/geokrety/1/loves", []string{"id", "1"}, "FetchGeokretyLoves", http.StatusOK, true, 20, 0},
		{"geokret-watches-alias", h.GetGeokretyWatches, "/api/v3/geokrety/1/watches", []string{"id", "1"}, "FetchGeokretyWatches", http.StatusOK, true, 20, 0},
		{"geokret-pictures", h.GetGeokretyPictures, "/api/v3/geokrety/1/pictures?limit=3&offset=4", []string{"id", "1"}, "FetchGeokretyPictures", http.StatusOK, true, 3, 4},
		{"geokret-search", h.SearchGeokrety, "/api/v3/geokrety/search?q=gk&limit=9&offset=2", nil, "SearchGeokrety", http.StatusOK, true, 9, 2},
		{"geokret-countries", h.GetGeokretyCountries, "/api/v3/geokrety/1/countries", []string{"id", "1"}, "FetchGeokretyCountries", http.StatusOK, true, 20, 0},
		{"geokret-waypoints", h.GetGeokretyWaypoints, "/api/v3/geokrety/1/waypoints", []string{"id", "1"}, "FetchGeokretyWaypoints", http.StatusOK, true, 20, 0},
		{"geokret-map-countries", h.GetGeokretyStatsMapCountries, "/api/v3/geokrety/1/stats/map/countries", []string{"id", "1"}, "FetchGeokretyStatsMapCountries", http.StatusOK, true, 20, 0},
		{"geokret-choropleth-alias", h.GetGeokretyWorldChoropleth, "/api/v3/geokrety/1/world-choropleth", []string{"id", "1"}, "FetchGeokretyStatsMapCountries", http.StatusOK, true, 20, 0},
		{"geokret-elevation", h.GetGeokretyStatsElevation, "/api/v3/geokrety/1/stats/elevation", []string{"id", "1"}, "FetchGeokretyStatsElevation", http.StatusOK, true, 20, 0},
		{"geokret-heatmap-days", h.GetGeokretyStatsHeatmapDays, "/api/v3/geokrety/1/stats/heatmap/days", []string{"id", "1"}, "FetchGeokretyStatsHeatmapDays", http.StatusOK, true, 20, 0},
		{"geokret-geojson-trip", h.GetGeokretyGeoJSONTrip, "/api/v3/geokrety/1/geojson/trip?limit=6&offset=1", []string{"id", "1"}, "FetchGeokretyTripPoints", http.StatusOK, true, 6, 1},
		{"country-list", h.GetCountryList, "/api/v3/countries/?limit=4&offset=1", nil, "FetchCountryList", http.StatusOK, true, 4, 1},
		{"country-details", h.GetCountryDetails, "/api/v3/countries/PL", []string{"code", "PL"}, "FetchCountryDetails", http.StatusOK, false, 0, 0},
		{"country-geokrety", h.GetCountryGeokrety, "/api/v3/countries/PL/geokrety?limit=4&offset=1", []string{"code", "PL"}, "FetchCountryGeokrety", http.StatusOK, true, 4, 1},
		{"country-geokrety-alias", h.GetCountrySpottedGeokrety, "/api/v3/countries/PL/spotted-geokrety", []string{"code", "PL"}, "FetchCountryGeokrety", http.StatusOK, true, 20, 0},
		{"waypoint-details", h.GetWaypoint, "/api/v3/waypoints/GC123", []string{"code", "GC123"}, "FetchWaypoint", http.StatusOK, false, 0, 0},
		{"waypoint-current", h.GetWaypointCurrentGeokrety, "/api/v3/waypoints/GC123/geokrety-current?limit=5&offset=2", []string{"code", "GC123"}, "FetchWaypointCurrentGeokrety", http.StatusOK, true, 5, 2},
		{"waypoint-current-alias", h.GetWaypointSpottedGeokrety, "/api/v3/waypoints/GC123/spotted-geokrety", []string{"code", "GC123"}, "FetchWaypointCurrentGeokrety", http.StatusOK, true, 20, 0},
		{"waypoint-past", h.GetWaypointPastGeokrety, "/api/v3/waypoints/GC123/geokrety-past?limit=5&offset=2", []string{"code", "GC123"}, "FetchWaypointPastGeokrety", http.StatusOK, true, 5, 2},
		{"waypoint-search", h.SearchWaypoints, "/api/v3/waypoints/search?q=gc&limit=5&offset=2", nil, "SearchWaypoints", http.StatusOK, true, 5, 2},
		{"user-list", h.GetUserList, "/api/v3/users/?limit=4&offset=1", nil, "FetchUserList", http.StatusOK, true, 4, 1},
		{"user-details", h.GetUserDetails, "/api/v3/users/1", []string{"id", "1"}, "FetchUserDetails", http.StatusOK, false, 0, 0},
		{"user-owned", h.GetUserOwnedGeokrety, "/api/v3/users/1/geokrety-owned?limit=5&offset=1", []string{"id", "1"}, "FetchUserOwnedGeokrety", http.StatusOK, true, 5, 1},
		{"user-found", h.GetUserFoundGeokrety, "/api/v3/users/1/geokrety-found", []string{"id", "1"}, "FetchUserFoundGeokrety", http.StatusOK, true, 20, 0},
		{"user-loved", h.GetUserLovedGeokrety, "/api/v3/users/1/geokrety-loved", []string{"id", "1"}, "FetchUserLovedGeokrety", http.StatusOK, true, 20, 0},
		{"user-watched", h.GetUserWatchedGeokrety, "/api/v3/users/1/geokrety-watched", []string{"id", "1"}, "FetchUserWatchedGeokrety", http.StatusOK, true, 20, 0},
		{"user-pictures", h.GetUserPictures, "/api/v3/users/1/pictures?limit=4&offset=1", []string{"id", "1"}, "FetchUserPictures", http.StatusOK, true, 4, 1},
		{"user-countries", h.GetUserCountries, "/api/v3/users/1/countries", []string{"id", "1"}, "FetchUserCountries", http.StatusOK, true, 20, 0},
		{"user-waypoints", h.GetUserWaypoints, "/api/v3/users/1/waypoints", []string{"id", "1"}, "FetchUserWaypoints", http.StatusOK, true, 20, 0},
		{"user-search", h.SearchUsers, "/api/v3/users/search?q=us&limit=2&offset=3", nil, "SearchUsers", http.StatusOK, true, 2, 3},
		{"user-continent-coverage", h.GetUserStatsContinentCoverage, "/api/v3/users/1/stats/continent-coverage?limit=2&offset=1", []string{"id", "1"}, "FetchUserStatsContinentCoverage", http.StatusOK, true, 2, 1},
		{"user-heatmap-days", h.GetUserStatsHeatmapDays, "/api/v3/users/1/stats/heatmap/days", []string{"id", "1"}, "FetchUserStatsHeatmapDays", http.StatusOK, true, 20, 0},
		{"user-heatmap-hours", h.GetUserStatsHeatmapHours, "/api/v3/users/1/stats/heatmap/hours", []string{"id", "1"}, "FetchUserStatsHeatmapHours", http.StatusOK, true, 20, 0},
		{"user-map-countries", h.GetUserStatsMapCountries, "/api/v3/users/1/stats/map/countries", []string{"id", "1"}, "FetchUserStatsMapCountries", http.StatusOK, true, 20, 0},
		{"picture-list", h.GetPictureList, "/api/v3/pictures/?limit=4&offset=1", nil, "FetchPictureList", http.StatusOK, true, 4, 1},
		{"picture-details", h.GetPictureDetails, "/api/v3/pictures/1", []string{"id", "1"}, "FetchPicture", http.StatusOK, false, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			if len(tc.params) > 0 {
				r = withRouteParams(r, tc.params...)
			}
			w := httptest.NewRecorder()
			tc.handler(w, r)
			if w.Code != tc.expStatus {
				t.Fatalf("expected status %d, got %d", tc.expStatus, w.Code)
			}
			if store.lastMethod != tc.expMethod {
				t.Fatalf("expected method %s, got %s", tc.expMethod, store.lastMethod)
			}
			if tc.checkLimit && (store.lastLimit != tc.expLimit || store.lastOffset != tc.expOffset) {
				t.Fatalf("expected limit/offset %d/%d, got %d/%d", tc.expLimit, tc.expOffset, store.lastLimit, store.lastOffset)
			}
			payload := decodeMap(t, w)
			if payload["data"] == nil {
				t.Fatalf("data field missing")
			}
			if payload["meta"] == nil {
				t.Fatalf("meta field missing")
			}
			if tc.name == "geokret-details-by-gkid" || tc.name == "geokret-details-by-numeric-gkid" {
				data := payload["data"].(map[string]any)
				if got := data["gkid"]; got != "GK0001" {
					t.Fatalf("data.gkid = %#v, want GK0001", got)
				}
			}
		})
	}
}

func TestEntityHandlerAcceptsBareHexGKID(t *testing.T) {
	store := &mockStatsStore{}
	h := NewStatsHandler(store, zap.NewNop())
	r := withRouteParams(httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/00FF", nil), "gkid", "00FF")
	w := httptest.NewRecorder()

	h.GetGeokretyDetailsByGkId(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	payload := decodeMap(t, w)
	data := payload["data"].(map[string]any)
	if got := data["gkid"]; got != "GK00FF" {
		t.Fatalf("data.gkid = %#v, want GK00FF", got)
	}
}

func TestEntityHandlerXMLResponseUsesCanonicalGKID(t *testing.T) {
	store := &mockStatsStore{}
	h := NewStatsHandler(store, zap.NewNop())
	r := withRouteParams(httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/GK0001", nil), "gkid", "GK0001")
	r.Header.Set("Accept", "application/xml")
	w := httptest.NewRecorder()

	h.GetGeokretyDetailsByGkId(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "application/xml" {
		t.Fatalf("content-type = %q, want application/xml", got)
	}
	if body := w.Body.String(); !strings.Contains(body, "<gkid>GK0001</gkid>") {
		t.Fatalf("expected XML body to contain canonical gkid, got %s", body)
	}
}

func TestEntityHandlerInvalidRequests(t *testing.T) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	tests := []struct {
		name    string
		handler http.HandlerFunc
		target  string
		params  []string
	}{
		{"bad-id", h.GetGeokrety, "/api/v3/geokrety/abc", []string{"id", "abc"}},
		{"bad-gkid", h.GetGeokretyDetailsByGkId, "/api/v3/geokrety/GKZZ", []string{"gkid", "GKZZ"}},
		{"bad-id-moves", h.GetGeokretyMoves, "/api/v3/geokrety/GKZZ/moves", []string{"id", "GKZZ"}},
		{"bad-id-move-details", h.GetGeokretyMoveDetails, "/api/v3/geokrety/GKZZ/moves/1", []string{"id", "GKZZ", "moveId", "1"}},
		{"bad-move-id", h.GetGeokretyMoveDetails, "/api/v3/geokrety/1/moves/x", []string{"id", "1", "moveId", "x"}},
		{"bad-id-loved-by", h.GetGeokretyLovedBy, "/api/v3/geokrety/GKZZ/loved-by", []string{"id", "GKZZ"}},
		{"bad-id-watched-by", h.GetGeokretyWatchedBy, "/api/v3/geokrety/GKZZ/watched-by", []string{"id", "GKZZ"}},
		{"bad-id-pictures", h.GetGeokretyPictures, "/api/v3/geokrety/GKZZ/pictures", []string{"id", "GKZZ"}},
		{"bad-id-countries", h.GetGeokretyCountries, "/api/v3/geokrety/GKZZ/countries", []string{"id", "GKZZ"}},
		{"bad-id-waypoints", h.GetGeokretyWaypoints, "/api/v3/geokrety/GKZZ/waypoints", []string{"id", "GKZZ"}},
		{"bad-id-map-countries", h.GetGeokretyStatsMapCountries, "/api/v3/geokrety/GKZZ/stats/map/countries", []string{"id", "GKZZ"}},
		{"bad-id-elevation", h.GetGeokretyStatsElevation, "/api/v3/geokrety/GKZZ/stats/elevation", []string{"id", "GKZZ"}},
		{"bad-id-heatmap-days", h.GetGeokretyStatsHeatmapDays, "/api/v3/geokrety/GKZZ/stats/heatmap/days", []string{"id", "GKZZ"}},
		{"bad-id-geojson-trip", h.GetGeokretyGeoJSONTrip, "/api/v3/geokrety/GKZZ/geojson/trip", []string{"id", "GKZZ"}},
		{"bad-country", h.GetCountryDetails, "/api/v3/countries/POL", []string{"code", "POL"}},
		{"bad-country-symbols", h.GetCountryGeokrety, "/api/v3/countries/1!/geokrety", []string{"code", "1!"}},
		{"bad-waypoint", h.GetWaypoint, "/api/v3/waypoints/*", []string{"code", "*"}},
		{"bad-waypoint-current", h.GetWaypointCurrentGeokrety, "/api/v3/waypoints/*/geokrety-current", []string{"code", "*"}},
		{"bad-waypoint-past", h.GetWaypointPastGeokrety, "/api/v3/waypoints/*/geokrety-past", []string{"code", "*"}},
		{"bad-id-user-details", h.GetUserDetails, "/api/v3/users/abc", []string{"id", "abc"}},
		{"bad-id-user-owned", h.GetUserOwnedGeokrety, "/api/v3/users/abc/geokrety-owned", []string{"id", "abc"}},
		{"bad-id-user-pictures", h.GetUserPictures, "/api/v3/users/abc/pictures", []string{"id", "abc"}},
		{"bad-id-user-countries", h.GetUserCountries, "/api/v3/users/abc/countries", []string{"id", "abc"}},
		{"bad-id-user-waypoints", h.GetUserWaypoints, "/api/v3/users/abc/waypoints", []string{"id", "abc"}},
		{"bad-id-user-heatmap-days", h.GetUserStatsHeatmapDays, "/api/v3/users/abc/stats/heatmap/days", []string{"id", "abc"}},
		{"bad-id-user-heatmap-hours", h.GetUserStatsHeatmapHours, "/api/v3/users/abc/stats/heatmap/hours", []string{"id", "abc"}},
		{"bad-id-user-map-countries", h.GetUserStatsMapCountries, "/api/v3/users/abc/stats/map/countries", []string{"id", "abc"}},
		{"bad-id-user-continent-coverage", h.GetUserStatsContinentCoverage, "/api/v3/users/abc/stats/continent-coverage", []string{"id", "abc"}},
		{"bad-id-picture", h.GetPictureDetails, "/api/v3/pictures/abc", []string{"id", "abc"}},
		{"bad-search-users", h.SearchUsers, "/api/v3/users/search?q=x", nil},
		{"bad-search-geokrety", h.SearchGeokrety, "/api/v3/geokrety/search?q=x", nil},
		{"bad-search-waypoints", h.SearchWaypoints, "/api/v3/waypoints/search?q=x", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			if len(tc.params) > 0 {
				r = withRouteParams(r, tc.params...)
			}
			w := httptest.NewRecorder()
			tc.handler(w, r)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestEntityHandlerStoreErrors(t *testing.T) {
	tests := []struct {
		name       string
		failMethod string
		noRows     string
		handler    http.HandlerFunc
		target     string
		params     []string
		expStatus  int
	}{
		{"details-500", "FetchGeokrety", "", NewStatsHandler(&mockStatsStore{failMethod: "FetchGeokrety"}, zap.NewNop()).GetGeokrety, "/api/v3/geokrety/1", []string{"id", "1"}, http.StatusInternalServerError},
		{"details-404", "", "FetchGeokrety", NewStatsHandler(&mockStatsStore{noRowsMethod: "FetchGeokrety"}, zap.NewNop()).GetGeokrety, "/api/v3/geokrety/1", []string{"id", "1"}, http.StatusNotFound},
		{"details-by-gkid-404", "", "FetchGeokretyByGKID", NewStatsHandler(&mockStatsStore{noRowsMethod: "FetchGeokretyByGKID"}, zap.NewNop()).GetGeokretyDetailsByGkId, "/api/v3/geokrety/GK0001", []string{"gkid", "GK0001"}, http.StatusNotFound},
		{"move-details-404", "", "FetchGeokretyMoveDetails", NewStatsHandler(&mockStatsStore{noRowsMethod: "FetchGeokretyMoveDetails"}, zap.NewNop()).GetGeokretyMoveDetails, "/api/v3/geokrety/1/moves/2", []string{"id", "1", "moveId", "2"}, http.StatusNotFound},
		{"geojson-500", "FetchGeokretyTripPoints", "", NewStatsHandler(&mockStatsStore{failMethod: "FetchGeokretyTripPoints"}, zap.NewNop()).GetGeokretyGeoJSONTrip, "/api/v3/geokrety/1/geojson/trip", []string{"id", "1"}, http.StatusInternalServerError},
		{"country-404", "", "FetchCountryDetails", NewStatsHandler(&mockStatsStore{noRowsMethod: "FetchCountryDetails"}, zap.NewNop()).GetCountryDetails, "/api/v3/countries/PL", []string{"code", "PL"}, http.StatusNotFound},
		{"waypoint-404", "", "FetchWaypoint", NewStatsHandler(&mockStatsStore{noRowsMethod: "FetchWaypoint"}, zap.NewNop()).GetWaypoint, "/api/v3/waypoints/GC123", []string{"code", "GC123"}, http.StatusNotFound},
		{"user-404", "", "FetchUserDetails", NewStatsHandler(&mockStatsStore{noRowsMethod: "FetchUserDetails"}, zap.NewNop()).GetUserDetails, "/api/v3/users/1", []string{"id", "1"}, http.StatusNotFound},
		{"picture-404", "", "FetchPicture", NewStatsHandler(&mockStatsStore{noRowsMethod: "FetchPicture"}, zap.NewNop()).GetPictureDetails, "/api/v3/pictures/1", []string{"id", "1"}, http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			r = withRouteParams(r, tc.params...)
			w := httptest.NewRecorder()
			tc.handler(w, r)
			if w.Code != tc.expStatus {
				t.Fatalf("expected status %d, got %d", tc.expStatus, w.Code)
			}
		})
	}
}

func TestEntityListHandlersInternalServerError(t *testing.T) {
	tests := []struct {
		name    string
		store   *mockStatsStore
		handler http.HandlerFunc
		target  string
		params  []string
	}{
		{"geokret-moves", &mockStatsStore{failMethod: "FetchGeokretyMoves"}, nil, "/api/v3/geokrety/1/moves", []string{"id", "1"}},
		{"geokret-loved-by", &mockStatsStore{failMethod: "FetchGeokretyLoves"}, nil, "/api/v3/geokrety/1/loved-by", []string{"id", "1"}},
		{"country-geokrety", &mockStatsStore{failMethod: "FetchCountryGeokrety"}, nil, "/api/v3/countries/PL/geokrety", []string{"code", "PL"}},
		{"waypoint-current", &mockStatsStore{failMethod: "FetchWaypointCurrentGeokrety"}, nil, "/api/v3/waypoints/GC123/geokrety-current", []string{"code", "GC123"}},
		{"user-owned", &mockStatsStore{failMethod: "FetchUserOwnedGeokrety"}, nil, "/api/v3/users/1/geokrety-owned", []string{"id", "1"}},
		{"user-pictures", &mockStatsStore{failMethod: "FetchUserPictures"}, nil, "/api/v3/users/1/pictures", []string{"id", "1"}},
		{"user-countries", &mockStatsStore{failMethod: "FetchUserCountries"}, nil, "/api/v3/users/1/countries", []string{"id", "1"}},
		{"user-waypoints", &mockStatsStore{failMethod: "FetchUserWaypoints"}, nil, "/api/v3/users/1/waypoints", []string{"id", "1"}},
		{"user-heatmap-days", &mockStatsStore{failMethod: "FetchUserStatsHeatmapDays"}, nil, "/api/v3/users/1/stats/heatmap/days", []string{"id", "1"}},
		{"user-heatmap-hours", &mockStatsStore{failMethod: "FetchUserStatsHeatmapHours"}, nil, "/api/v3/users/1/stats/heatmap/hours", []string{"id", "1"}},
		{"user-map-countries", &mockStatsStore{failMethod: "FetchUserStatsMapCountries"}, nil, "/api/v3/users/1/stats/map/countries", []string{"id", "1"}},
		{"user-continent-coverage", &mockStatsStore{failMethod: "FetchUserStatsContinentCoverage"}, nil, "/api/v3/users/1/stats/continent-coverage", []string{"id", "1"}},
	}

	for i := range tests {
		tests[i].handler = map[string]http.HandlerFunc{
			"geokret-moves":           NewStatsHandler(tests[i].store, zap.NewNop()).GetGeokretyMoves,
			"geokret-loved-by":        NewStatsHandler(tests[i].store, zap.NewNop()).GetGeokretyLovedBy,
			"country-geokrety":        NewStatsHandler(tests[i].store, zap.NewNop()).GetCountryGeokrety,
			"waypoint-current":        NewStatsHandler(tests[i].store, zap.NewNop()).GetWaypointCurrentGeokrety,
			"user-owned":              NewStatsHandler(tests[i].store, zap.NewNop()).GetUserOwnedGeokrety,
			"user-pictures":           NewStatsHandler(tests[i].store, zap.NewNop()).GetUserPictures,
			"user-countries":          NewStatsHandler(tests[i].store, zap.NewNop()).GetUserCountries,
			"user-waypoints":          NewStatsHandler(tests[i].store, zap.NewNop()).GetUserWaypoints,
			"user-heatmap-days":       NewStatsHandler(tests[i].store, zap.NewNop()).GetUserStatsHeatmapDays,
			"user-heatmap-hours":      NewStatsHandler(tests[i].store, zap.NewNop()).GetUserStatsHeatmapHours,
			"user-map-countries":      NewStatsHandler(tests[i].store, zap.NewNop()).GetUserStatsMapCountries,
			"user-continent-coverage": NewStatsHandler(tests[i].store, zap.NewNop()).GetUserStatsContinentCoverage,
		}[tests[i].name]
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			r = withRouteParams(r, tc.params...)
			w := httptest.NewRecorder()
			tc.handler(w, r)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500, got %d", w.Code)
			}
		})
	}
}
