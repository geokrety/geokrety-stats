package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geokrety/geokrety-stats-api/internal/config"
	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/geokrety/geokrety-stats-api/internal/handlers"
	"github.com/geokrety/geokrety-stats-api/internal/metrics"
	"github.com/geokrety/geokrety-stats-api/internal/ws"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func testRouter(t *testing.T) http.Handler {
	t.Helper()

	store := &handlersTestStore{}
	systemStore := &handlersSystemStore{}
	logger := zap.NewNop()
	reg := prometheus.NewRegistry()
	mc := metrics.New(reg)
	hub := ws.NewHub(logger, mc, 0)
	statsHandler := handlers.NewStatsHandler(store, logger)
	systemHandler := handlers.NewSystemHandler(systemStore, hub, logger)

	return NewRouter(config.Config{EnableSwagger: false}, logger, mc, reg, statsHandler, systemHandler, hub)
}

type handlersTestStore struct{}

type handlersSystemStore struct{}

func (h *handlersSystemStore) Ping(ctx context.Context) error { return nil }

func (h *handlersTestStore) FetchGlobalStats(ctx context.Context) (db.GlobalStats, error) {
	return db.GlobalStats{}, nil
}

func (h *handlersTestStore) FetchCountries(ctx context.Context, limit, offset int) ([]db.CountryStats, error) {
	return []db.CountryStats{}, nil
}

func (h *handlersTestStore) FetchLeaderboard(ctx context.Context, limit, offset int) ([]db.LeaderboardUser, error) {
	return []db.LeaderboardUser{}, nil
}

func (h *handlersTestStore) FetchRecentMoves(ctx context.Context, limit, offset int) ([]db.RecentMove, error) {
	return []db.RecentMove{}, nil
}

func (h *handlersTestStore) FetchRecentBorn(ctx context.Context, limit, offset int) ([]db.RecentBorn, error) {
	return []db.RecentBorn{}, nil
}

func (h *handlersTestStore) FetchRecentLoved(ctx context.Context, limit, offset int) ([]db.RecentLoved, error) {
	return []db.RecentLoved{}, nil
}

func (h *handlersTestStore) FetchRecentWatched(ctx context.Context, limit, offset int) ([]db.RecentWatched, error) {
	return []db.RecentWatched{}, nil
}

func (h *handlersTestStore) FetchRecentActiveCountries(ctx context.Context, limit, offset int) ([]db.ActiveCountry, error) {
	return []db.ActiveCountry{}, nil
}

func (h *handlersTestStore) FetchRecentActiveWaypoints(ctx context.Context, limit, offset int) ([]db.ActiveWaypoint, error) {
	return []db.ActiveWaypoint{}, nil
}

func (h *handlersTestStore) FetchRecentRegisteredUsers(ctx context.Context, limit, offset int) ([]db.RecentRegisteredUser, error) {
	return []db.RecentRegisteredUser{}, nil
}

func (h *handlersTestStore) FetchRecentActiveUsers(ctx context.Context, limit, offset int) ([]db.RecentActiveUser, error) {
	return []db.RecentActiveUser{}, nil
}

func TestV3RoutesReachable(t *testing.T) {
	r := testRouter(t)
	paths := []string{
		"/health",
		"/metrics",
		"/api/v3/stats/kpis",
		"/api/v3/stats/countries",
		"/api/v3/stats/leaderboard",
		"/api/v3/geokrety/recent-moves",
		"/api/v3/geokrety/recent-born",
		"/api/v3/geokrety/recent-loved",
		"/api/v3/geokrety/recent-watched",
		"/api/v3/countries/recent-active",
		"/api/v3/waypoints/recent-active",
		"/api/v3/users/recent-registered",
		"/api/v3/users/recent-active",
	}

	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Fatalf("route not found: %s", p)
		}
	}
}

func TestCORSOptions(t *testing.T) {
	r := testRouter(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/v3/stats/kpis", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
