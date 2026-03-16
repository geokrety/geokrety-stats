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
	store := &mockStatsStore{failMethod: "FetchRecentMoves"}
	h := NewStatsHandler(store, zap.NewNop())

	r := httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/recent-moves", nil)
	w := httptest.NewRecorder()
	h.GetRecentMoves(w, r)

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
