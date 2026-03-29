package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/config"
	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/geokrety/geokrety-stats-api/internal/handlers"
	"github.com/geokrety/geokrety-stats-api/internal/metrics"
	"github.com/geokrety/geokrety-stats-api/internal/ws"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type apiTestStore struct{}

type apiSystemStore struct{}

func (s *apiSystemStore) Ping(ctx context.Context) error { return nil }

func apiGKID(value int64) *geokrety.GeokretId {
	parsed, err := geokrety.FromInt(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func apiInt64Ptr(value int64) *int64 { return &value }

func apiStringPtr(value string) *string { return &value }

func apiCountryFlag(code string) string {
	if len(code) != 2 {
		return ""
	}
	runes := []rune(strings.ToUpper(code))
	if runes[0] < 'A' || runes[0] > 'Z' || runes[1] < 'A' || runes[1] > 'Z' {
		return ""
	}
	return string([]rune{runes[0] - 'A' + 0x1F1E6, runes[1] - 'A' + 0x1F1E6})
}

func newRouterForTests(t *testing.T) http.Handler {
	t.Helper()
	logger := zap.NewNop()
	reg := prometheus.NewRegistry()
	mc := metrics.New(reg)
	hub := ws.NewHub(logger, mc, 0)
	store := &apiTestStore{}
	statsHandler := handlers.NewStatsHandler(handlers.StatsHandlerStores{
		Geokrety:  store,
		Moves:     store,
		Countries: store,
		Waypoints: store,
		Users:     store,
		Pictures:  store,
	}, logger)
	systemHandler := handlers.NewSystemHandler(&apiSystemStore{}, hub, logger)
	return NewRouter(config.Config{EnableSwagger: false}, logger, mc, reg, statsHandler, systemHandler, hub)
}

func sampleAPIMove(id int64) db.MoveRecord {
	country := "PL"
	waypoint := "OC146C3"
	return db.MoveRecord{
		ID:          id,
		GeokretID:   1,
		GeokretGKID: apiGKID(1),
		MoveType:    0,
		Username:    "walker",
		Country:     &country,
		Waypoint:    &waypoint,
		MovedOn:     time.Date(2026, 3, 29, 10, 30, 0, 0, time.UTC),
		CreatedOn:   time.Date(2026, 3, 29, 10, 30, 0, 0, time.UTC),
	}
}

func (s *apiTestStore) ResolveGeokretID(ctx context.Context, gkid int64) (int64, error) {
	return gkid, nil
}

func (s *apiTestStore) FetchGeokretyList(ctx context.Context, filters db.GeokretListFilters, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: apiGKID(1), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchGeokretyByGKIDs(ctx context.Context, gkids []int64) ([]db.GeokretListItem, error) {
	rows := make([]db.GeokretListItem, 0, len(gkids))
	for _, gkid := range gkids {
		rows = append(rows, db.GeokretListItem{ID: gkid, GKID: apiGKID(gkid), Name: "Traveler", Type: 1})
	}
	return rows, nil
}

func (s *apiTestStore) FetchGeokretyByGKID(ctx context.Context, gkid int64) (db.GeokretDetails, error) {
	ownerID := int64(1)
	ownerUsername := "owner"
	return db.GeokretDetails{GeokretListItem: db.GeokretListItem{ID: gkid, GKID: apiGKID(gkid), Name: "Traveler", Type: 1, OwnerID: &ownerID, OwnerUsername: &ownerUsername}}, nil
}

func (s *apiTestStore) FetchGeokretStats(ctx context.Context, geokretID int64) (db.GeokretStats, error) {
	return db.GeokretStats{GeokretID: geokretID, GKID: apiGKID(geokretID), MovesCount: 9}, nil
}

func (s *apiTestStore) FetchMoveList(ctx context.Context, filters db.MoveFilters, limit, offset int) ([]db.MoveRecord, error) {
	return []db.MoveRecord{sampleAPIMove(9)}, nil
}

func (s *apiTestStore) FetchMoveListByIDs(ctx context.Context, filters db.MoveFilters, moveIDs []int64) ([]db.MoveRecord, error) {
	rows := make([]db.MoveRecord, 0, len(moveIDs))
	for _, moveID := range moveIDs {
		rows = append(rows, sampleAPIMove(moveID))
	}
	return rows, nil
}

func (s *apiTestStore) FetchMove(ctx context.Context, moveID int64) (db.MoveRecord, error) {
	return sampleAPIMove(moveID), nil
}

func (s *apiTestStore) FetchGeokretyLoves(ctx context.Context, geokretID int64, sort db.Sort, limit, offset int) ([]db.SocialUserEntry, error) {
	return []db.SocialUserEntry{{UserID: 1, Username: "lover", At: time.Now().UTC()}}, nil
}

func (s *apiTestStore) FetchGeokretyWatches(ctx context.Context, geokretID int64, sort db.Sort, limit, offset int) ([]db.SocialUserEntry, error) {
	return []db.SocialUserEntry{{UserID: 2, Username: "watcher", At: time.Now().UTC()}}, nil
}

func (s *apiTestStore) FetchGeokretyFinders(ctx context.Context, geokretID int64, sort db.Sort, limit, offset int) ([]db.SocialUserEntry, error) {
	return []db.SocialUserEntry{{UserID: 3, Username: "finder", At: time.Now().UTC()}}, nil
}

func (s *apiTestStore) FetchGeokretyCountries(ctx context.Context, geokretID int64, sort db.Sort, limit, offset int) ([]db.GeokretCountryVisit, error) {
	return []db.GeokretCountryVisit{{CountryCode: "PL", FirstVisitedAt: time.Now().UTC(), MoveCount: 1, Flag: "🇵🇱"}}, nil
}

func (s *apiTestStore) FetchGeokretyWaypoints(ctx context.Context, geokretID int64, sort db.Sort, limit, offset int) ([]db.GeokretWaypointVisit, error) {
	return []db.GeokretWaypointVisit{{WaypointCode: "OC146C3", VisitCount: 1, FirstVisitedAt: time.Now().UTC(), LastVisitedAt: time.Now().UTC()}}, nil
}

func (s *apiTestStore) FetchCountryList(ctx context.Context, filters db.CountryListFilters, limit, offset int) ([]db.CountryDetails, error) {
	return []db.CountryDetails{{Code: "PL", Name: "Poland", Flag: "🇵🇱"}}, nil
}

func (s *apiTestStore) FetchCountryListByCodes(ctx context.Context, codes []string) ([]db.CountryDetails, error) {
	rows := make([]db.CountryDetails, 0, len(codes))
	for _, code := range codes {
		rows = append(rows, db.CountryDetails{Code: code, Name: code, Flag: apiCountryFlag(code)})
	}
	return rows, nil
}

func (s *apiTestStore) FetchCountryDetails(ctx context.Context, countryCode string) (db.CountryDetails, error) {
	return db.CountryDetails{Code: countryCode, Name: "Poland", Flag: "🇵🇱"}, nil
}

func (s *apiTestStore) FetchCountryGeokrety(ctx context.Context, countryCode string, sort db.Sort, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: apiGKID(1), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchWaypoint(ctx context.Context, waypointCode string) (db.WaypointDetails, error) {
	return db.WaypointDetails{WaypointSummary: db.WaypointSummary{WaypointCode: waypointCode, Source: "opencaching"}}, nil
}

func (s *apiTestStore) FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, sort db.Sort, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: apiGKID(1), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, sort db.Sort, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 2, GKID: apiGKID(2), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchUserList(ctx context.Context, filters db.UserListFilters, limit, offset int) ([]db.UserSearchResult, error) {
	homeCountry := "PL"
	return []db.UserSearchResult{{ID: 1, Username: "alice", JoinedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), HomeCountry: &homeCountry, HomeCountryFlag: "🇵🇱"}}, nil
}

func (s *apiTestStore) FetchUserListByIDs(ctx context.Context, userIDs []int64) ([]db.UserSearchResult, error) {
	rows := make([]db.UserSearchResult, 0, len(userIDs))
	for _, userID := range userIDs {
		rows = append(rows, db.UserSearchResult{ID: userID, Username: "alice", JoinedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)})
	}
	return rows, nil
}

func (s *apiTestStore) FetchUserDetails(ctx context.Context, userID int64) (db.UserDetails, error) {
	homeCountry := "PL"
	return db.UserDetails{ID: userID, Username: "alice", JoinedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), HomeCountry: &homeCountry, HomeCountryFlag: "🇵🇱"}, nil
}

func (s *apiTestStore) FetchUserStats(ctx context.Context, userID int64) (db.UserStats, error) {
	return db.UserStats{UserID: userID, MovesCount: 9}, nil
}

func (s *apiTestStore) FetchUserOwnedGeokrety(ctx context.Context, userID int64, sort db.Sort, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: apiGKID(1), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchUserFoundGeokrety(ctx context.Context, userID int64, sort db.Sort, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: apiGKID(1), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchUserLovedGeokrety(ctx context.Context, userID int64, sort db.Sort, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: apiGKID(1), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchUserWatchedGeokrety(ctx context.Context, userID int64, sort db.Sort, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: apiGKID(1), Name: "Traveler", Type: 1}}, nil
}

func (s *apiTestStore) FetchUserCountries(ctx context.Context, userID int64, sort db.Sort, limit, offset int) ([]db.UserCountryVisit, error) {
	return []db.UserCountryVisit{{CountryCode: "PL", MoveCount: 1, FirstVisit: time.Now().UTC(), LastVisit: time.Now().UTC(), Flag: "🇵🇱"}}, nil
}

func (s *apiTestStore) FetchUserWaypoints(ctx context.Context, userID int64, sort db.Sort, limit, offset int) ([]db.UserWaypointVisit, error) {
	return []db.UserWaypointVisit{{WaypointCode: "OC146C3", VisitCount: 1, FirstVisitedAt: time.Now().UTC(), LastVisitedAt: time.Now().UTC()}}, nil
}

func (s *apiTestStore) FetchPictureList(ctx context.Context, filters db.PictureFilters, limit, offset int) ([]db.PictureInfo, error) {
	return []db.PictureInfo{{ID: 1, GeokretID: apiInt64Ptr(1), GeokretGKID: apiGKID(1), MoveID: apiInt64Ptr(9), UserID: apiInt64Ptr(1), AuthorID: apiInt64Ptr(1), AuthorUsername: apiStringPtr("alice"), CreatedOn: time.Now().UTC()}}, nil
}

func (s *apiTestStore) FetchPictureListByIDs(ctx context.Context, pictureIDs []int64) ([]db.PictureInfo, error) {
	rows := make([]db.PictureInfo, 0, len(pictureIDs))
	for _, pictureID := range pictureIDs {
		rows = append(rows, db.PictureInfo{ID: pictureID, GeokretID: apiInt64Ptr(1), GeokretGKID: apiGKID(1), MoveID: apiInt64Ptr(9), UserID: apiInt64Ptr(1), AuthorID: apiInt64Ptr(1), AuthorUsername: apiStringPtr("alice"), CreatedOn: time.Now().UTC()})
	}
	return rows, nil
}

func (s *apiTestStore) FetchPicture(ctx context.Context, pictureID int64) (db.PictureInfo, error) {
	return db.PictureInfo{ID: pictureID, GeokretID: apiInt64Ptr(1), GeokretGKID: apiGKID(1), MoveID: apiInt64Ptr(9), UserID: apiInt64Ptr(1), AuthorID: apiInt64Ptr(1), AuthorUsername: apiStringPtr("alice"), CreatedOn: time.Now().UTC()}, nil
}

func TestRouterExposesNewPublicSurface(t *testing.T) {
	r := newRouterForTests(t)
	paths := []string{
		"/api/v3/moves",
		"/api/v3/moves/9",
		"/api/v3/geokrety/GK0001/stats",
		"/api/v3/geokrety/GK0001/finders",
		"/api/v3/users/1/stats",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
		})
	}
}

func TestRouterRemovesRedundantAliases(t *testing.T) {
	r := newRouterForTests(t)
	paths := []string{
		"/api/v3/geokrety/search",
		"/api/v3/geokrety/GK0001/moves/9",
		"/api/v3/users/search",
		"/api/v3/waypoints/search",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d", w.Code)
			}
		})
	}
}
