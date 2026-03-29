package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/config"
	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/geokrety/geokrety-stats-api/internal/handlers"
	"github.com/geokrety/geokrety-stats-api/internal/metrics"
	"github.com/geokrety/geokrety-stats-api/internal/ws"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	sharedjsonrest "github.com/geokrety/geokrety-stats/geokrety/jsonrest"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type integrationStore struct {
	handlersTestStore
}

func newIntegrationRouter(t *testing.T) http.Handler {
	t.Helper()
	logger := zap.NewNop()
	reg := prometheus.NewRegistry()
	mc := metrics.New(reg)
	hub := ws.NewHub(logger, mc, 0)
	statsHandler := handlers.NewStatsHandler(&integrationStore{}, logger)
	systemHandler := handlers.NewSystemHandler(&handlersSystemStore{}, hub, logger)
	return NewRouter(config.Config{EnableSwagger: false}, logger, mc, reg, statsHandler, systemHandler, hub)
}

func integrationGKID(value int64) *geokrety.GeokretId {
	parsed, err := geokrety.FromInt(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func int64PtrIntegration(value int64) *int64 { return &value }

func decodeIntegrationMap(t *testing.T, body *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return payload
}

func (s *integrationStore) FetchGeokretyList(ctx context.Context, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1}}, nil
}

func (s *integrationStore) FetchGeokretyByGKIDs(ctx context.Context, gkids []int64) ([]db.GeokretListItem, error) {
	rows := make([]db.GeokretListItem, 0, len(gkids))
	for _, gkid := range gkids {
		rows = append(rows, db.GeokretListItem{ID: gkid, GKID: integrationGKID(gkid), Name: "GK", Type: 1})
	}
	return rows, nil
}

func (s *integrationStore) FetchGeokrety(ctx context.Context, geokretID int64) (db.GeokretDetails, error) {
	return db.GeokretDetails{GeokretListItem: db.GeokretListItem{ID: geokretID, GKID: integrationGKID(geokretID), Name: "GK", Type: 1}}, nil
}

func (s *integrationStore) FetchGeokretyByGKID(ctx context.Context, gkid int64) (db.GeokretDetails, error) {
	return s.FetchGeokrety(ctx, gkid)
}

func (s *integrationStore) ResolveGeokretID(ctx context.Context, gkid int64) (int64, error) {
	return gkid, nil
}

func (s *integrationStore) SearchGeokrety(ctx context.Context, query string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: query, Type: 1}}, nil
}

func (s *integrationStore) FetchGeokretyMoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.MoveRecord, error) {
	return []db.MoveRecord{{ID: 1, GeokretID: geokretID, MovedOn: time.Now(), CreatedOn: time.Now()}}, nil
}

func (s *integrationStore) FetchGeokretyMovesByIDs(ctx context.Context, geokretID int64, moveIDs []int64) ([]db.MoveRecord, error) {
	rows := make([]db.MoveRecord, 0, len(moveIDs))
	for _, moveID := range moveIDs {
		rows = append(rows, db.MoveRecord{ID: moveID, GeokretID: geokretID, MovedOn: time.Now(), CreatedOn: time.Now()})
	}
	return rows, nil
}

func (s *integrationStore) FetchGeokretyMoveDetails(ctx context.Context, geokretID, moveID int64) (db.MoveRecord, error) {
	return db.MoveRecord{ID: moveID, GeokretID: geokretID, MovedOn: time.Now(), CreatedOn: time.Now()}, nil
}

func (s *integrationStore) FetchGeokretyLoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	return []db.SocialUserEntry{{UserID: 1, Username: "lover", At: time.Now()}}, nil
}

func (s *integrationStore) FetchGeokretyWatches(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	return []db.SocialUserEntry{{UserID: 1, Username: "watcher", At: time.Now()}}, nil
}

func (s *integrationStore) FetchGeokretyPictures(ctx context.Context, geokretID int64, limit, offset int) ([]db.PictureInfo, error) {
	return []db.PictureInfo{{ID: 1, GeokretID: int64PtrIntegration(geokretID), GeokretGKID: integrationGKID(geokretID), CreatedOn: time.Now()}}, nil
}

func (s *integrationStore) FetchGeokretyCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretCountryVisit, error) {
	return []db.GeokretCountryVisit{{CountryCode: "PL", FirstVisitedAt: time.Now(), MoveCount: 1}}, nil
}

func (s *integrationStore) FetchGeokretyWaypoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretWaypointVisit, error) {
	now := time.Now()
	return []db.GeokretWaypointVisit{{WaypointCode: "GC123", VisitCount: 1, FirstVisitedAt: now, LastVisitedAt: now}}, nil
}

func (s *integrationStore) FetchCountryList(ctx context.Context, limit, offset int) ([]db.CountryDetails, error) {
	return []db.CountryDetails{{Code: "PL", Flag: "PL"}}, nil
}

func (s *integrationStore) FetchCountryListByCodes(ctx context.Context, codes []string) ([]db.CountryDetails, error) {
	rows := make([]db.CountryDetails, 0, len(codes))
	for _, code := range codes {
		rows = append(rows, db.CountryDetails{Code: code, Flag: code})
	}
	return rows, nil
}

func (s *integrationStore) FetchCountryDetails(ctx context.Context, countryCode string) (db.CountryDetails, error) {
	return db.CountryDetails{Code: countryCode, Flag: countryCode}, nil
}

func (s *integrationStore) FetchCountryGeokrety(ctx context.Context, countryCode string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1, Country: &countryCode}}, nil
}

func (s *integrationStore) SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]db.WaypointSummary, error) {
	return []db.WaypointSummary{{WaypointCode: "GC123", Source: "gc"}}, nil
}

func (s *integrationStore) FetchWaypoint(ctx context.Context, waypointCode string) (db.WaypointDetails, error) {
	return db.WaypointDetails{WaypointSummary: db.WaypointSummary{WaypointCode: waypointCode, Source: "gc"}, CurrentGeokrety: 1}, nil
}

func (s *integrationStore) FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1, Waypoint: &waypointCode}}, nil
}

func (s *integrationStore) FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1, Waypoint: &waypointCode}}, nil
}

func (s *integrationStore) FetchUserList(ctx context.Context, limit, offset int) ([]db.UserSearchResult, error) {
	return []db.UserSearchResult{{ID: 1, Username: "user", JoinedAt: time.Now()}}, nil
}

func (s *integrationStore) FetchUserListByIDs(ctx context.Context, userIDs []int64) ([]db.UserSearchResult, error) {
	rows := make([]db.UserSearchResult, 0, len(userIDs))
	for _, userID := range userIDs {
		rows = append(rows, db.UserSearchResult{ID: userID, Username: "user", JoinedAt: time.Now()})
	}
	return rows, nil
}

func (s *integrationStore) SearchUsers(ctx context.Context, query string, limit, offset int) ([]db.UserSearchResult, error) {
	return []db.UserSearchResult{{ID: 1, Username: query, JoinedAt: time.Now()}}, nil
}

func (s *integrationStore) FetchUserDetails(ctx context.Context, userID int64) (db.UserDetails, error) {
	return db.UserDetails{ID: userID, Username: "user", JoinedAt: time.Now()}, nil
}

func (s *integrationStore) FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1}}, nil
}

func (s *integrationStore) FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1}}, nil
}

func (s *integrationStore) FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1}}, nil
}

func (s *integrationStore) FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	return []db.GeokretListItem{{ID: 1, GKID: integrationGKID(1), Name: "GK", Type: 1}}, nil
}

func (s *integrationStore) FetchUserPictures(ctx context.Context, userID int64, limit, offset int) ([]db.PictureInfo, error) {
	return []db.PictureInfo{{ID: 1, GeokretID: int64PtrIntegration(1), GeokretGKID: integrationGKID(1), UserID: int64PtrIntegration(userID), CreatedOn: time.Now()}}, nil
}

func (s *integrationStore) FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]db.UserCountryVisit, error) {
	now := time.Now()
	return []db.UserCountryVisit{{CountryCode: "PL", MoveCount: 1, FirstVisit: now, LastVisit: now}}, nil
}

func (s *integrationStore) FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]db.UserWaypointVisit, error) {
	now := time.Now()
	return []db.UserWaypointVisit{{WaypointCode: "GC123", VisitCount: 1, FirstVisitedAt: now, LastVisitedAt: now}}, nil
}

func (s *integrationStore) FetchPictureList(ctx context.Context, limit, offset int) ([]db.PictureInfo, error) {
	return []db.PictureInfo{{ID: 1, GeokretID: int64PtrIntegration(1), GeokretGKID: integrationGKID(1), CreatedOn: time.Now()}}, nil
}

func (s *integrationStore) FetchPictureListByIDs(ctx context.Context, pictureIDs []int64) ([]db.PictureInfo, error) {
	rows := make([]db.PictureInfo, 0, len(pictureIDs))
	for _, pictureID := range pictureIDs {
		rows = append(rows, db.PictureInfo{ID: pictureID, GeokretID: int64PtrIntegration(1), GeokretGKID: integrationGKID(1), CreatedOn: time.Now()})
	}
	return rows, nil
}

func (s *integrationStore) FetchPicture(ctx context.Context, pictureID int64) (db.PictureInfo, error) {
	return db.PictureInfo{ID: pictureID, GeokretID: int64PtrIntegration(1), GeokretGKID: integrationGKID(1), CreatedOn: time.Now()}, nil
}

func TestPublicAPIIntegrationCollectionContracts(t *testing.T) {
	r := newIntegrationRouter(t)
	tests := []struct {
		path     string
		wantType string
	}{
		{"/api/v3/geokrety?limit=1", "geokret"},
		{"/api/v3/geokrety?ids=GK0001,GK0002", "geokret"},
		{"/api/v3/geokrety/search?q=gk", "geokret"},
		{"/api/v3/geokrety/GK0001/moves?limit=1", "move"},
		{"/api/v3/geokrety/GK0001/moves?ids=7,9", "move"},
		{"/api/v3/geokrety/GK0001/loved-by", "user"},
		{"/api/v3/geokrety/GK0001/watched-by", "user"},
		{"/api/v3/geokrety/GK0001/pictures", "picture"},
		{"/api/v3/geokrety/GK0001/countries", "country_visit"},
		{"/api/v3/geokrety/GK0001/waypoints", "waypoint_visit"},
		{"/api/v3/countries?limit=1", "country"},
		{"/api/v3/countries?ids=PL,DE", "country"},
		{"/api/v3/countries/PL/geokrety", "geokret"},
		{"/api/v3/waypoints/search?q=gc", "waypoint"},
		{"/api/v3/waypoints/GC123/geokrety-current", "geokret"},
		{"/api/v3/waypoints/GC123/geokrety-past", "geokret"},
		{"/api/v3/users?limit=1", "user"},
		{"/api/v3/users?ids=1,2", "user"},
		{"/api/v3/users/search?q=us", "user"},
		{"/api/v3/users/1/geokrety-owned", "geokret"},
		{"/api/v3/users/1/geokrety-found", "geokret"},
		{"/api/v3/users/1/geokrety-loved", "geokret"},
		{"/api/v3/users/1/geokrety-watched", "geokret"},
		{"/api/v3/users/1/pictures", "picture"},
		{"/api/v3/users/1/countries", "country_visit"},
		{"/api/v3/users/1/waypoints", "waypoint_visit"},
		{"/api/v3/pictures?limit=1", "picture"},
		{"/api/v3/pictures?ids=1,2", "picture"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
			payload := decodeIntegrationMap(t, w)
			if payload["data"] == nil || payload["meta"] == nil || payload["links"] == nil {
				t.Fatalf("json rest envelope missing fields: %#v", payload)
			}
			data := payload["data"].([]any)
			if len(data) == 0 {
				t.Fatalf("expected non-empty data array")
			}
			first := data[0].(map[string]any)
			if got := first["type"]; got != tc.wantType {
				t.Fatalf("resource type = %#v, want %s", got, tc.wantType)
			}
		})
	}
}

func TestPublicAPIIntegrationDetailContracts(t *testing.T) {
	r := newIntegrationRouter(t)
	tests := []struct {
		path     string
		wantType string
	}{
		{"/api/v3/geokrety/GK0001", "geokret"},
		{"/api/v3/geokrety/GK0001/moves/7", "move"},
		{"/api/v3/countries/PL", "country"},
		{"/api/v3/waypoints/GC123", "waypoint"},
		{"/api/v3/users/1", "user"},
		{"/api/v3/pictures/1", "picture"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
			payload := decodeIntegrationMap(t, w)
			if payload["data"] == nil || payload["meta"] == nil || payload["links"] == nil {
				t.Fatalf("json rest envelope missing fields: %#v", payload)
			}
			data := payload["data"].(map[string]any)
			if got := data["type"]; got != tc.wantType {
				t.Fatalf("resource type = %#v, want %s", got, tc.wantType)
			}
			if data["links"] == nil {
				t.Fatalf("resource links missing")
			}
		})
	}
}

func TestPublicAPIIntegrationErrorsAndPagination(t *testing.T) {
	r := newIntegrationRouter(t)
	cursor := sharedjsonrest.EncodeCursor(sharedjsonrest.CurrentCursorVersion, 1)
	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "invalid user batch ids", path: "/api/v3/users?ids=1,nope", want: http.StatusBadRequest},
		{name: "invalid geokret ids", path: "/api/v3/geokrety?ids=bad!", want: http.StatusBadRequest},
		{name: "cursor pagination", path: "/api/v3/geokrety?limit=1&cursor=" + cursor.String(), want: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, w.Code)
			}
		})
	}
}
