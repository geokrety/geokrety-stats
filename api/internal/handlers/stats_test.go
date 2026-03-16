package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type mockStatsStore struct {
	failMethod string
	lastMethod string
	lastLimit  int
	lastOffset int
}

func (m *mockStatsStore) maybeFail(method string) error {
	m.lastMethod = method
	if m.failMethod == method {
		return errors.New("boom")
	}
	return nil
}

func (m *mockStatsStore) FetchGlobalStats(ctx context.Context) (db.GlobalStats, error) {
	if err := m.maybeFail("FetchGlobalStats"); err != nil {
		return db.GlobalStats{}, err
	}
	return db.GlobalStats{TotalGeokrety: 10, TotalMoves: 20}, nil
}

func (m *mockStatsStore) FetchCountries(ctx context.Context, limit, offset int) ([]db.CountryStats, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchCountries"); err != nil {
		return nil, err
	}
	return []db.CountryStats{{Code: "PL", Name: "Poland"}}, nil
}

func (m *mockStatsStore) FetchLeaderboard(ctx context.Context, limit, offset int) ([]db.LeaderboardUser, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchLeaderboard"); err != nil {
		return nil, err
	}
	return []db.LeaderboardUser{{UserID: 1, Username: "user", MovesCount: 1, Points: 1}}, nil
}

func (m *mockStatsStore) FetchRecentMoves(ctx context.Context, limit, offset int) ([]db.RecentMove, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentMoves"); err != nil {
		return nil, err
	}
	return []db.RecentMove{{ID: 1, GeokretName: "GK"}}, nil
}

func (m *mockStatsStore) FetchRecentBorn(ctx context.Context, limit, offset int) ([]db.RecentBorn, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentBorn"); err != nil {
		return nil, err
	}
	return []db.RecentBorn{{ID: 1, Name: "GK", BornAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentLoved(ctx context.Context, limit, offset int) ([]db.RecentLoved, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentLoved"); err != nil {
		return nil, err
	}
	return []db.RecentLoved{{GeoKretID: 1, Username: "u", LovedAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentWatched(ctx context.Context, limit, offset int) ([]db.RecentWatched, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentWatched"); err != nil {
		return nil, err
	}
	return []db.RecentWatched{{GeoKretID: 1, Username: "u", WatchedAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentActiveCountries(ctx context.Context, limit, offset int) ([]db.ActiveCountry, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentActiveCountries"); err != nil {
		return nil, err
	}
	return []db.ActiveCountry{{Code: "PL", Moves: 10, LastActivityAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentActiveWaypoints(ctx context.Context, limit, offset int) ([]db.ActiveWaypoint, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentActiveWaypoints"); err != nil {
		return nil, err
	}
	return []db.ActiveWaypoint{{Waypoint: "GC123", Moves: 10, LastActivityAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentRegisteredUsers(ctx context.Context, limit, offset int) ([]db.RecentRegisteredUser, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentRegisteredUsers"); err != nil {
		return nil, err
	}
	return []db.RecentRegisteredUser{{ID: 1, Username: "u", JoinedAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentActiveUsers(ctx context.Context, limit, offset int) ([]db.RecentActiveUser, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentActiveUsers"); err != nil {
		return nil, err
	}
	return []db.RecentActiveUser{{UserID: 1, Username: "u", LastMoveAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchHourlyHeatmap(ctx context.Context, limit, offset int) ([]db.HourlyHeatmapCell, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchHourlyHeatmap"); err != nil {
		return nil, err
	}
	return []db.HourlyHeatmapCell{{ActivityDate: time.Now(), HourUTC: 12, MoveType: 5, MoveCount: 10}}, nil
}

func (m *mockStatsStore) FetchCountryFlows(ctx context.Context, limit, offset int) ([]db.CountryFlow, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchCountryFlows"); err != nil {
		return nil, err
	}
	return []db.CountryFlow{{FromCountry: "PL", ToCountry: "DE", MoveCount: 2}}, nil
}

func (m *mockStatsStore) FetchTopCaches(ctx context.Context, limit, offset int) ([]db.TopCache, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchTopCaches"); err != nil {
		return nil, err
	}
	return []db.TopCache{{WaypointCode: "GC123", TotalGKVisits: 3, DistinctGKCount: 2}}, nil
}

func (m *mockStatsStore) FetchFirstFinderLeaderboard(ctx context.Context, limit, offset int) ([]db.FirstFinderLeaderboardEntry, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchFirstFinderLeaderboard"); err != nil {
		return nil, err
	}
	return []db.FirstFinderLeaderboardEntry{{UserID: 1, Username: "u", FirstFinds: 4, Rank: 1}}, nil
}

func (m *mockStatsStore) FetchDistanceRecords(ctx context.Context, limit, offset int) ([]db.DistanceRecord, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchDistanceRecords"); err != nil {
		return nil, err
	}
	return []db.DistanceRecord{{GeoKretID: 1, GeoKretName: "GK", KMTotal: 100, Rank: 1}}, nil
}

func (m *mockStatsStore) FetchUserNetwork(ctx context.Context, userID int64, limit, offset int) ([]db.UserNetworkEdge, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserNetwork"); err != nil {
		return nil, err
	}
	return []db.UserNetworkEdge{{UserID: userID, RelatedUserID: 2, RelatedUsername: "peer", SharedGeoKretyCount: 3}}, nil
}

func (m *mockStatsStore) FetchGeokretTimeline(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretTimelineEvent, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretTimeline"); err != nil {
		return nil, err
	}
	return []db.GeokretTimelineEvent{{GeoKretID: geokretID, EventType: "km_100", OccurredAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchGeokretCirculation(ctx context.Context, geokretID int64) (db.GeokretCirculation, error) {
	if err := m.maybeFail("FetchGeokretCirculation"); err != nil {
		return db.GeokretCirculation{}, err
	}
	return db.GeokretCirculation{GeoKretID: geokretID, GeoKretName: "GK", Users: 2, Interactions: 6, AvgPerUser: 3}, nil
}

func TestStatsHandlerSuccessEndpoints(t *testing.T) {
	store := &mockStatsStore{}
	h := NewStatsHandler(store, zap.NewNop())

	tests := []struct {
		name       string
		target     string
		handler    http.HandlerFunc
		method     string
		expStatus  int
		expMethod  string
		expLimit   int
		expOffset  int
		checkLimit bool
	}{
		{"kpis", "/api/v3/stats/kpis", h.GetKPIs, "FetchGlobalStats", http.StatusOK, "FetchGlobalStats", 0, 0, false},
		{"countries", "/api/v3/stats/countries?limit=33&offset=2", h.GetCountries, "FetchCountries", http.StatusOK, "FetchCountries", 33, 2, true},
		{"leaderboard", "/api/v3/stats/leaderboard?limit=15&offset=3", h.GetLeaderboard, "FetchLeaderboard", http.StatusOK, "FetchLeaderboard", 15, 3, true},
		{"recent-moves", "/api/v3/geokrety/recent-moves?limit=11&offset=1", h.GetRecentMoves, "FetchRecentMoves", http.StatusOK, "FetchRecentMoves", 11, 1, true},
		{"recent-born", "/api/v3/geokrety/recent-born?limit=11&offset=1", h.GetRecentBorn, "FetchRecentBorn", http.StatusOK, "FetchRecentBorn", 11, 1, true},
		{"recent-loved", "/api/v3/geokrety/recent-loved?limit=11&offset=1", h.GetRecentLoved, "FetchRecentLoved", http.StatusOK, "FetchRecentLoved", 11, 1, true},
		{"recent-watched", "/api/v3/geokrety/recent-watched?limit=11&offset=1", h.GetRecentWatched, "FetchRecentWatched", http.StatusOK, "FetchRecentWatched", 11, 1, true},
		{"countries-active", "/api/v3/countries/recent-active?limit=11&offset=1", h.GetRecentActiveCountries, "FetchRecentActiveCountries", http.StatusOK, "FetchRecentActiveCountries", 11, 1, true},
		{"waypoints-active", "/api/v3/waypoints/recent-active?limit=11&offset=1", h.GetRecentActiveWaypoints, "FetchRecentActiveWaypoints", http.StatusOK, "FetchRecentActiveWaypoints", 11, 1, true},
		{"users-registered", "/api/v3/users/recent-registered?limit=11&offset=1", h.GetRecentRegisteredUsers, "FetchRecentRegisteredUsers", http.StatusOK, "FetchRecentRegisteredUsers", 11, 1, true},
		{"users-active", "/api/v3/users/recent-active?limit=11&offset=1", h.GetRecentActiveUsers, "FetchRecentActiveUsers", http.StatusOK, "FetchRecentActiveUsers", 11, 1, true},
		{"hourly-heatmap", "/api/v3/stats/hourly-heatmap?limit=11&offset=1", h.GetHourlyHeatmap, "FetchHourlyHeatmap", http.StatusOK, "FetchHourlyHeatmap", 11, 1, true},
		{"country-flows", "/api/v3/stats/country-flows?limit=11&offset=1", h.GetCountryFlows, "FetchCountryFlows", http.StatusOK, "FetchCountryFlows", 11, 1, true},
		{"top-caches", "/api/v3/stats/top-caches?limit=11&offset=1", h.GetTopCaches, "FetchTopCaches", http.StatusOK, "FetchTopCaches", 11, 1, true},
		{"first-finder", "/api/v3/stats/first-finder-leaderboard?limit=11&offset=1", h.GetFirstFinderLeaderboard, "FetchFirstFinderLeaderboard", http.StatusOK, "FetchFirstFinderLeaderboard", 11, 1, true},
		{"distance-records", "/api/v3/stats/distance-records?limit=11&offset=1", h.GetDistanceRecords, "FetchDistanceRecords", http.StatusOK, "FetchDistanceRecords", 11, 1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			w := httptest.NewRecorder()
			tc.handler(w, r)

			if w.Code != tc.expStatus {
				t.Fatalf("expected status %d, got %d", tc.expStatus, w.Code)
			}
			if store.lastMethod != tc.expMethod {
				t.Fatalf("expected method %s, got %s", tc.expMethod, store.lastMethod)
			}
			if tc.checkLimit {
				if store.lastLimit != tc.expLimit || store.lastOffset != tc.expOffset {
					t.Fatalf("expected limit/offset %d/%d, got %d/%d", tc.expLimit, tc.expOffset, store.lastLimit, store.lastOffset)
				}
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if payload["data"] == nil {
				t.Fatalf("data field missing")
			}
			if payload["meta"] == nil {
				t.Fatalf("meta field missing")
			}
		})
	}
}

func TestStatsHandlerErrorEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		failMethod string
		target     string
		handler    http.HandlerFunc
	}{
		{"kpis", "FetchGlobalStats", "/api/v3/stats/kpis", nil},
		{"countries", "FetchCountries", "/api/v3/stats/countries", nil},
		{"leaderboard", "FetchLeaderboard", "/api/v3/stats/leaderboard", nil},
		{"recent-moves", "FetchRecentMoves", "/api/v3/geokrety/recent-moves", nil},
		{"heatmap", "FetchHourlyHeatmap", "/api/v3/stats/hourly-heatmap", nil},
		{"country-flows", "FetchCountryFlows", "/api/v3/stats/country-flows", nil},
		{"top-caches", "FetchTopCaches", "/api/v3/stats/top-caches", nil},
		{"first-finder", "FetchFirstFinderLeaderboard", "/api/v3/stats/first-finder-leaderboard", nil},
		{"distance-records", "FetchDistanceRecords", "/api/v3/stats/distance-records", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStatsStore{failMethod: tc.failMethod}
			h := NewStatsHandler(store, zap.NewNop())

			handler := map[string]http.HandlerFunc{
				"FetchGlobalStats":            h.GetKPIs,
				"FetchCountries":              h.GetCountries,
				"FetchLeaderboard":            h.GetLeaderboard,
				"FetchRecentMoves":            h.GetRecentMoves,
				"FetchHourlyHeatmap":          h.GetHourlyHeatmap,
				"FetchCountryFlows":           h.GetCountryFlows,
				"FetchTopCaches":              h.GetTopCaches,
				"FetchFirstFinderLeaderboard": h.GetFirstFinderLeaderboard,
				"FetchDistanceRecords":        h.GetDistanceRecords,
			}[tc.failMethod]

			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			w := httptest.NewRecorder()
			handler(w, r)

			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected status 500, got %d", w.Code)
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if payload["error"] == nil {
				t.Fatalf("error field missing")
			}
		})
	}
}

func TestStatsHandlerParamEndpoints(t *testing.T) {
	store := &mockStatsStore{}
	h := NewStatsHandler(store, zap.NewNop())

	tests := []struct {
		name      string
		handler   http.HandlerFunc
		path      string
		key       string
		value     string
		expMethod string
	}{
		{"user-network", h.GetUserNetwork, "/api/v3/users/123/network?limit=9&offset=2", "id", "123", "FetchUserNetwork"},
		{"gk-timeline", h.GetGeokretTimeline, "/api/v3/geokrety/123/timeline?limit=9&offset=2", "id", "123", "FetchGeokretTimeline"},
		{"gk-circulation", h.GetGeokretCirculation, "/api/v3/geokrety/123/circulation", "id", "123", "FetchGeokretCirculation"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add(tc.key, tc.value)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			tc.handler(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}
			if store.lastMethod != tc.expMethod {
				t.Fatalf("expected method %s, got %s", tc.expMethod, store.lastMethod)
			}
		})
	}
}

func TestStatsHandlerInvalidIdentifier(t *testing.T) {
	store := &mockStatsStore{}
	h := NewStatsHandler(store, zap.NewNop())

	tests := []http.HandlerFunc{h.GetUserNetwork, h.GetGeokretTimeline, h.GetGeokretCirculation}
	for _, handler := range tests {
		r := httptest.NewRequest(http.MethodGet, "/bad", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "abc")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		handler(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	}
}

func TestStatsHandlerGeokretCirculationError(t *testing.T) {
	store := &mockStatsStore{failMethod: "FetchGeokretCirculation"}
	h := NewStatsHandler(store, zap.NewNop())
	r := httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/123/circulation", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetGeokretCirculation(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestQueryIntBounds(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/x?limit=-1&offset=abc", nil)
	if got := queryInt(r, "limit", 20, 1, 1000); got != 1 {
		t.Fatalf("expected lower bound 1, got %d", got)
	}
	if got := queryInt(r, "offset", 0, 0, 1000); got != 0 {
		t.Fatalf("expected fallback 0, got %d", got)
	}

	r2 := httptest.NewRequest(http.MethodGet, "/x?limit=99999", nil)
	if got := queryInt(r2, "limit", 20, 1, 1000); got != 1000 {
		t.Fatalf("expected upper bound 1000, got %d", got)
	}
}
