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

func (h *handlersTestStore) FetchHourlyHeatmap(ctx context.Context, limit, offset int) ([]db.HourlyHeatmapCell, error) {
	return []db.HourlyHeatmapCell{}, nil
}

func (h *handlersTestStore) FetchCountryFlows(ctx context.Context, limit, offset int) ([]db.CountryFlow, error) {
	return []db.CountryFlow{}, nil
}

func (h *handlersTestStore) FetchTopCaches(ctx context.Context, limit, offset int) ([]db.TopCache, error) {
	return []db.TopCache{}, nil
}

func (h *handlersTestStore) FetchFirstFinderLeaderboard(ctx context.Context, limit, offset int) ([]db.FirstFinderLeaderboardEntry, error) {
	return []db.FirstFinderLeaderboardEntry{}, nil
}

func (h *handlersTestStore) FetchDistanceRecords(ctx context.Context, limit, offset int) ([]db.DistanceRecord, error) {
	return []db.DistanceRecord{}, nil
}

func (h *handlersTestStore) FetchStatsDormancy(ctx context.Context, limit, offset int) ([]db.DormancyRecord, error) {
	return []db.DormancyRecord{}, nil
}

func (h *handlersTestStore) FetchStatsMultiplierVelocity(ctx context.Context, limit, offset int) ([]db.MultiplierVelocityRecord, error) {
	return []db.MultiplierVelocityRecord{}, nil
}

func (h *handlersTestStore) FetchCountryList(ctx context.Context, limit, offset int) ([]db.CountryDetails, error) {
	return []db.CountryDetails{}, nil
}

func (h *handlersTestStore) FetchUserNetwork(ctx context.Context, userID int64, limit, offset int) ([]db.UserNetworkEdge, error) {
	return []db.UserNetworkEdge{}, nil
}

func (h *handlersTestStore) FetchGeokretTimeline(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretTimelineEvent, error) {
	return []db.GeokretTimelineEvent{}, nil
}

func (h *handlersTestStore) FetchGeokretCirculation(ctx context.Context, geokretID int64) (db.GeokretCirculation, error) {
	return db.GeokretCirculation{}, nil
}

func (h *handlersTestStore) FetchGeokrety(ctx context.Context, geokretID int64) (db.GeokretDetails, error) {
	return db.GeokretDetails{}, nil
}

func (h *handlersTestStore) FetchGeokretyList(ctx context.Context, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchGeokretyByGKID(ctx context.Context, gkid int64) (db.GeokretDetails, error) {
	return db.GeokretDetails{}, nil
}

func (h *handlersTestStore) ResolveGeokretID(ctx context.Context, gkid int64) (int64, error) {
	return gkid, nil
}

func (h *handlersTestStore) FetchGeokretyMoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.MoveRecord, error) {
	return []db.MoveRecord{}, nil
}

func (h *handlersTestStore) FetchGeokretyMoveDetails(ctx context.Context, geokretID, moveID int64) (db.MoveRecord, error) {
	return db.MoveRecord{}, nil
}

func (h *handlersTestStore) FetchGeokretyLoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	return []db.SocialUserEntry{}, nil
}

func (h *handlersTestStore) FetchGeokretyWatches(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	return []db.SocialUserEntry{}, nil
}

func (h *handlersTestStore) FetchGeokretyPictures(ctx context.Context, geokretID int64, limit, offset int) ([]db.PictureInfo, error) {
	return []db.PictureInfo{}, nil
}

func (h *handlersTestStore) SearchGeokrety(ctx context.Context, query string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchGeokretyCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretCountryVisit, error) {
	return []db.GeokretCountryVisit{}, nil
}

func (h *handlersTestStore) FetchGeokretyWaypoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretWaypointVisit, error) {
	return []db.GeokretWaypointVisit{}, nil
}

func (h *handlersTestStore) FetchGeokretyStatsMapCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.CountryCount, error) {
	return []db.CountryCount{}, nil
}

func (h *handlersTestStore) FetchGeokretyStatsElevation(ctx context.Context, geokretID int64, limit, offset int) ([]db.ElevationPoint, error) {
	return []db.ElevationPoint{}, nil
}

func (h *handlersTestStore) FetchGeokretyStatsHeatmapDays(ctx context.Context, geokretID int64, limit, offset int) ([]db.DayHeatmapCell, error) {
	return []db.DayHeatmapCell{}, nil
}

func (h *handlersTestStore) FetchGeokretyTripPoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.TripPoint, error) {
	return []db.TripPoint{}, nil
}

func (h *handlersTestStore) FetchCountryDetails(ctx context.Context, countryCode string) (db.CountryDetails, error) {
	return db.CountryDetails{}, nil
}

func (h *handlersTestStore) FetchCountryGeokrety(ctx context.Context, countryCode string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchWaypoint(ctx context.Context, waypointCode string) (db.WaypointDetails, error) {
	return db.WaypointDetails{}, nil
}

func (h *handlersTestStore) FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]db.WaypointSummary, error) {
	return []db.WaypointSummary{}, nil
}

func (h *handlersTestStore) FetchUserDetails(ctx context.Context, userID int64) (db.UserDetails, error) {
	return db.UserDetails{}, nil
}

func (h *handlersTestStore) FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{}, nil
}

func (h *handlersTestStore) FetchUserPictures(ctx context.Context, userID int64, limit, offset int) ([]db.PictureInfo, error) {
	return []db.PictureInfo{}, nil
}

func (h *handlersTestStore) FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]db.UserCountryVisit, error) {
	return []db.UserCountryVisit{}, nil
}

func (h *handlersTestStore) FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]db.UserWaypointVisit, error) {
	return []db.UserWaypointVisit{}, nil
}

func (h *handlersTestStore) FetchUserList(ctx context.Context, limit, offset int) ([]db.UserSearchResult, error) {
	return []db.UserSearchResult{}, nil
}

func (h *handlersTestStore) SearchUsers(ctx context.Context, query string, limit, offset int) ([]db.UserSearchResult, error) {
	return []db.UserSearchResult{}, nil
}

func (h *handlersTestStore) FetchUserStatsContinentCoverage(ctx context.Context, userID int64, limit, offset int) ([]db.UserContinentCoverage, error) {
	return []db.UserContinentCoverage{}, nil
}

func (h *handlersTestStore) FetchUserStatsHeatmapDays(ctx context.Context, userID int64, limit, offset int) ([]db.DayHeatmapCell, error) {
	return []db.DayHeatmapCell{}, nil
}

func (h *handlersTestStore) FetchUserStatsHeatmapHours(ctx context.Context, userID int64, limit, offset int) ([]db.HourHeatmapCell, error) {
	return []db.HourHeatmapCell{}, nil
}

func (h *handlersTestStore) FetchUserStatsMapCountries(ctx context.Context, userID int64, limit, offset int) ([]db.CountryCount, error) {
	return []db.CountryCount{}, nil
}

func (h *handlersTestStore) FetchPicture(ctx context.Context, pictureID int64) (db.PictureInfo, error) {
	return db.PictureInfo{}, nil
}

func (h *handlersTestStore) FetchPictureList(ctx context.Context, limit, offset int) ([]db.PictureInfo, error) {
	return []db.PictureInfo{}, nil
}

func TestV3RoutesReachable(t *testing.T) {
	r := testRouter(t)
	paths := []string{
		"/health",
		"/metrics",
		"/api/v3/stats/kpis",
		"/api/v3/stats/countries",
		"/api/v3/stats/leaderboard",
		"/api/v3/stats/hourly-heatmap",
		"/api/v3/stats/seasonal-heatmap",
		"/api/v3/stats/country-flows",
		"/api/v3/stats/top-caches",
		"/api/v3/stats/first-finder-leaderboard",
		"/api/v3/stats/distance-records",
		"/api/v3/stats/dormancy",
		"/api/v3/stats/multiplier-velocity",
		"/api/v3/geokrety/recent-moves",
		"/api/v3/geokrety/recent-born",
		"/api/v3/geokrety/recent-loved",
		"/api/v3/geokrety/recent-watched",
		"/api/v3/geokrety/GK0001",
		"/api/v3/geokrety/1",
		"/api/v3/geokrety/GK0001/moves",
		"/api/v3/geokrety/1/moves",
		"/api/v3/geokrety/1/moves/2",
		"/api/v3/geokrety/1/loved-by",
		"/api/v3/geokrety/1/watched-by",
		"/api/v3/geokrety/1/pictures",
		"/api/v3/geokrety/search?q=gk",
		"/api/v3/geokrety/1/timeline",
		"/api/v3/geokrety/1/circulation",
		"/api/v3/geokrety/1/countries",
		"/api/v3/geokrety/1/waypoints",
		"/api/v3/geokrety/1/stats/map/countries",
		"/api/v3/geokrety/1/stats/elevation",
		"/api/v3/geokrety/1/stats/heatmap/days",
		"/api/v3/geokrety/1/geojson/trip",
		"/api/v3/countries/",
		"/api/v3/countries/PL",
		"/api/v3/countries/recent-active",
		"/api/v3/countries/PL/geokrety",
		"/api/v3/waypoints/recent-active",
		"/api/v3/waypoints/GC123",
		"/api/v3/waypoints/GC123/geokrety-current",
		"/api/v3/waypoints/GC123/geokrety-past",
		"/api/v3/waypoints/search?q=gc",
		"/api/v3/users/recent-registered",
		"/api/v3/users/recent-active",
		"/api/v3/users/",
		"/api/v3/users/1",
		"/api/v3/users/1/geokrety-owned",
		"/api/v3/users/1/geokrety-found",
		"/api/v3/users/1/geokrety-loved",
		"/api/v3/users/1/geokrety-watched",
		"/api/v3/users/1/pictures",
		"/api/v3/users/1/countries",
		"/api/v3/users/1/waypoints",
		"/api/v3/users/1/network",
		"/api/v3/users/search?q=us",
		"/api/v3/users/1/stats/continent-coverage",
		"/api/v3/users/1/stats/heatmap/days",
		"/api/v3/users/1/stats/heatmap/hours",
		"/api/v3/users/1/stats/map/countries",
		"/api/v3/pictures/",
		"/api/v3/pictures/1",
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
