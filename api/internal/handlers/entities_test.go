package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type mockStatsStore struct {
	lastMethod        string
	lastLimit         int
	lastOffset        int
	lastMoveFilters   db.MoveFilters
	lastMoveIDs       []int64
	lastPictureIDs    []int64
	lastRequestedGKID int64
}

func gkidPtr(value int64) *geokrety.GeokretId {
	parsed, err := geokrety.FromInt(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func int64Ptr(value int64) *int64 { return &value }

func testStringPtr(value string) *string { return &value }

func decodePayload(t *testing.T, body *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return payload
}

func withRouteParams(r *http.Request, pairs ...string) *http.Request {
	rctx := chi.NewRouteContext()
	for i := 0; i < len(pairs); i += 2 {
		rctx.URLParams.Add(pairs[i], pairs[i+1])
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func sampleGeokret(gkid int64) db.GeokretListItem {
	ownerID := int64(1)
	holderID := int64(2)
	lastPositionID := int64(77)
	lastLogID := int64(78)
	country := "PL"
	waypoint := "OC146C3"
	avatarURL := "https://cdn.example/avatar.jpg"
	ownerUsername := "owner"
	holderUsername := "holder"
	mission := "Keep moving"
	now := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	lat := 52.23
	lon := 21.01
	return db.GeokretListItem{
		ID:             gkid,
		GKID:           gkidPtr(gkid),
		Name:           "Traveler",
		AvatarID:       int64Ptr(10),
		AvatarURL:      &avatarURL,
		Type:           1,
		TypeName:       "Traditional",
		Missing:        false,
		OwnerID:        &ownerID,
		OwnerUsername:  &ownerUsername,
		HolderID:       &holderID,
		HolderUsername: &holderUsername,
		Country:        &country,
		Waypoint:       &waypoint,
		Lat:            &lat,
		Lon:            &lon,
		BornAt:         &now,
		LastMoveAt:     &now,
		LastPositionID: &lastPositionID,
		LastLogID:      &lastLogID,
		Mission:        &mission,
		CommentsHidden: true,
		GeoJSON: &db.GeoJSONPt{
			Type:        "Point",
			Coordinates: []float64{21.01, 52.23},
		},
	}
}

func sampleMove(id int64) db.MoveRecord {
	country := "PL"
	waypoint := "OC146C3"
	comment := "Moved"
	now := time.Date(2026, 3, 29, 10, 30, 0, 0, time.UTC)
	return db.MoveRecord{
		ID:                 id,
		GeokretID:          1,
		GeokretGKID:        gkidPtr(1),
		MoveType:           0,
		MoveTypeName:       "Dropped",
		AuthorID:           int64Ptr(1),
		AuthorAvatarID:     int64Ptr(15),
		AuthorAvatarURL:    testStringPtr("https://cdn.example/u.jpg"),
		Username:           "walker",
		Country:            &country,
		Waypoint:           &waypoint,
		MovedOn:            now,
		CreatedOn:          now,
		Comment:            &comment,
		PreviousMoveID:     int64Ptr(7),
		PreviousPositionID: int64Ptr(8),
		GeoJSON: &db.GeoJSONPt{
			Type:        "Point",
			Coordinates: []float64{21.01, 52.23},
		},
	}
}

func sampleUser(id int64) db.UserDetails {
	homeCountry := "PL"
	avatarURL := "https://cdn.example/user.jpg"
	joinedAt := time.Date(2026, 1, 2, 13, 0, 0, 0, time.UTC)
	lastMoveAt := time.Date(2026, 3, 29, 6, 30, 0, 0, time.UTC)
	return db.UserDetails{
		ID:              id,
		Username:        "alice",
		JoinedAt:        joinedAt,
		HomeCountry:     &homeCountry,
		HomeCountryFlag: "🇵🇱",
		AvatarID:        int64Ptr(22),
		AvatarURL:       &avatarURL,
		LastMoveAt:      &lastMoveAt,
	}
}

func (m *mockStatsStore) ResolveGeokretID(ctx context.Context, gkid int64) (int64, error) {
	m.lastRequestedGKID = gkid
	return gkid, nil
}

func (m *mockStatsStore) FetchGeokretyList(ctx context.Context, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod, m.lastLimit, m.lastOffset = "FetchGeokretyList", limit, offset
	return []db.GeokretListItem{sampleGeokret(1)}, nil
}

func (m *mockStatsStore) FetchGeokretyByGKIDs(ctx context.Context, gkids []int64) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchGeokretyByGKIDs"
	rows := make([]db.GeokretListItem, 0, len(gkids))
	for _, gkid := range gkids {
		rows = append(rows, sampleGeokret(gkid))
	}
	return rows, nil
}

func (m *mockStatsStore) FetchGeokretyByGKID(ctx context.Context, gkid int64) (db.GeokretDetails, error) {
	m.lastMethod = "FetchGeokretyByGKID"
	return db.GeokretDetails{GeokretListItem: sampleGeokret(gkid)}, nil
}

func (m *mockStatsStore) SearchGeokrety(ctx context.Context, query string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod, m.lastLimit, m.lastOffset = "SearchGeokrety", limit, offset
	item := sampleGeokret(1)
	item.Name = query
	return []db.GeokretListItem{item}, nil
}

func (m *mockStatsStore) FetchGeokretStats(ctx context.Context, geokretID int64) (db.GeokretStats, error) {
	m.lastMethod = "FetchGeokretStats"
	currentCountry := "PL"
	currentWaypoint := "OC146C3"
	return db.GeokretStats{
		GeokretID:             geokretID,
		GKID:                  gkidPtr(geokretID),
		CachesCount:           4,
		PicturesCount:         3,
		LovesCount:            2,
		MovesCount:            9,
		CountriesVisitedCount: 5,
		WaypointsVisitedCount: 6,
		FindersCount:          7,
		WatchersCount:         8,
		LoversCount:           2,
		CurrentCountryCode:    &currentCountry,
		CurrentWaypointCode:   &currentWaypoint,
	}, nil
}

func (m *mockStatsStore) FetchMoveList(ctx context.Context, filters db.MoveFilters, limit, offset int) ([]db.MoveRecord, error) {
	m.lastMethod, m.lastLimit, m.lastOffset, m.lastMoveFilters = "FetchMoveList", limit, offset, filters
	return []db.MoveRecord{sampleMove(9)}, nil
}

func (m *mockStatsStore) FetchMoveListByIDs(ctx context.Context, filters db.MoveFilters, moveIDs []int64) ([]db.MoveRecord, error) {
	m.lastMethod, m.lastMoveFilters = "FetchMoveListByIDs", filters
	m.lastMoveIDs = append([]int64(nil), moveIDs...)
	rows := make([]db.MoveRecord, 0, len(moveIDs))
	for _, moveID := range moveIDs {
		rows = append(rows, sampleMove(moveID))
	}
	return rows, nil
}

func (m *mockStatsStore) FetchMove(ctx context.Context, moveID int64) (db.MoveRecord, error) {
	m.lastMethod = "FetchMove"
	return sampleMove(moveID), nil
}

func (m *mockStatsStore) FetchGeokretyLoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	m.lastMethod = "FetchGeokretyLoves"
	return []db.SocialUserEntry{{UserID: 1, Username: "lover", At: time.Now().UTC()}}, nil
}

func (m *mockStatsStore) FetchGeokretyWatches(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	m.lastMethod = "FetchGeokretyWatches"
	return []db.SocialUserEntry{{UserID: 2, Username: "watcher", At: time.Now().UTC()}}, nil
}

func (m *mockStatsStore) FetchGeokretyFinders(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	m.lastMethod = "FetchGeokretyFinders"
	return []db.SocialUserEntry{{UserID: 3, Username: "finder", At: time.Now().UTC()}}, nil
}

func (m *mockStatsStore) FetchGeokretyCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretCountryVisit, error) {
	m.lastMethod = "FetchGeokretyCountries"
	return []db.GeokretCountryVisit{{CountryCode: "PL", FirstVisitedAt: time.Now().UTC(), MoveCount: 1, Flag: "🇵🇱"}}, nil
}

func (m *mockStatsStore) FetchGeokretyWaypoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretWaypointVisit, error) {
	m.lastMethod = "FetchGeokretyWaypoints"
	return []db.GeokretWaypointVisit{{WaypointCode: "OC146C3", VisitCount: 1, FirstVisitedAt: time.Now().UTC(), LastVisitedAt: time.Now().UTC()}}, nil
}

func (m *mockStatsStore) FetchCountryList(ctx context.Context, limit, offset int) ([]db.CountryDetails, error) {
	m.lastMethod = "FetchCountryList"
	return []db.CountryDetails{{Code: "PL", Name: "Poland", Flag: "🇵🇱"}}, nil
}

func (m *mockStatsStore) FetchCountryListByCodes(ctx context.Context, codes []string) ([]db.CountryDetails, error) {
	m.lastMethod = "FetchCountryListByCodes"
	rows := make([]db.CountryDetails, 0, len(codes))
	for _, code := range codes {
		rows = append(rows, db.CountryDetails{Code: code, Name: code, Flag: countryFlagFromCode(code)})
	}
	return rows, nil
}

func (m *mockStatsStore) FetchCountryDetails(ctx context.Context, countryCode string) (db.CountryDetails, error) {
	m.lastMethod = "FetchCountryDetails"
	return db.CountryDetails{Code: countryCode, Name: "Poland", Flag: countryFlagFromCode(countryCode)}, nil
}

func (m *mockStatsStore) FetchCountryGeokrety(ctx context.Context, countryCode string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchCountryGeokrety"
	item := sampleGeokret(1)
	item.Country = &countryCode
	return []db.GeokretListItem{item}, nil
}

func (m *mockStatsStore) FetchWaypoint(ctx context.Context, waypointCode string) (db.WaypointDetails, error) {
	m.lastMethod = "FetchWaypoint"
	return db.WaypointDetails{WaypointSummary: db.WaypointSummary{WaypointCode: waypointCode, Source: "opencaching"}}, nil
}

func (m *mockStatsStore) FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchWaypointCurrentGeokrety"
	item := sampleGeokret(1)
	item.Waypoint = &waypointCode
	return []db.GeokretListItem{item}, nil
}

func (m *mockStatsStore) FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchWaypointPastGeokrety"
	item := sampleGeokret(2)
	item.Waypoint = &waypointCode
	return []db.GeokretListItem{item}, nil
}

func (m *mockStatsStore) SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]db.WaypointSummary, error) {
	m.lastMethod = "SearchWaypoints"
	return []db.WaypointSummary{{WaypointCode: "OC146C3", Source: "opencaching"}}, nil
}

func (m *mockStatsStore) FetchUserList(ctx context.Context, limit, offset int) ([]db.UserSearchResult, error) {
	m.lastMethod = "FetchUserList"
	user := sampleUser(1)
	return []db.UserSearchResult{{ID: user.ID, Username: user.Username, JoinedAt: user.JoinedAt, HomeCountry: user.HomeCountry, HomeCountryFlag: user.HomeCountryFlag, AvatarID: user.AvatarID, AvatarURL: user.AvatarURL, LastMoveAt: user.LastMoveAt}}, nil
}

func (m *mockStatsStore) FetchUserListByIDs(ctx context.Context, userIDs []int64) ([]db.UserSearchResult, error) {
	m.lastMethod = "FetchUserListByIDs"
	rows := make([]db.UserSearchResult, 0, len(userIDs))
	for _, userID := range userIDs {
		user := sampleUser(userID)
		rows = append(rows, db.UserSearchResult{ID: user.ID, Username: user.Username, JoinedAt: user.JoinedAt, HomeCountry: user.HomeCountry, HomeCountryFlag: user.HomeCountryFlag, AvatarID: user.AvatarID, AvatarURL: user.AvatarURL, LastMoveAt: user.LastMoveAt})
	}
	return rows, nil
}

func (m *mockStatsStore) SearchUsers(ctx context.Context, query string, limit, offset int) ([]db.UserSearchResult, error) {
	m.lastMethod = "SearchUsers"
	user := sampleUser(1)
	user.Username = query
	return []db.UserSearchResult{{ID: user.ID, Username: user.Username, JoinedAt: user.JoinedAt, HomeCountry: user.HomeCountry, HomeCountryFlag: user.HomeCountryFlag, AvatarID: user.AvatarID, AvatarURL: user.AvatarURL, LastMoveAt: user.LastMoveAt}}, nil
}

func (m *mockStatsStore) FetchUserDetails(ctx context.Context, userID int64) (db.UserDetails, error) {
	m.lastMethod = "FetchUserDetails"
	return sampleUser(userID), nil
}

func (m *mockStatsStore) FetchUserStats(ctx context.Context, userID int64) (db.UserStats, error) {
	m.lastMethod = "FetchUserStats"
	return db.UserStats{UserID: userID, OwnedGeokretyCount: 2, FoundGeokretyCount: 3, LovedGeokretyCount: 4, WatchedGeokretyCount: 5, PicturesCount: 6, CountriesVisitedCount: 7, WaypointsVisitedCount: 8, MovesCount: 9, DistinctGeokretyCount: 10}, nil
}

func (m *mockStatsStore) FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchUserOwnedGeokrety"
	return []db.GeokretListItem{sampleGeokret(1)}, nil
}

func (m *mockStatsStore) FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchUserFoundGeokrety"
	return []db.GeokretListItem{sampleGeokret(1)}, nil
}

func (m *mockStatsStore) FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchUserLovedGeokrety"
	return []db.GeokretListItem{sampleGeokret(1)}, nil
}

func (m *mockStatsStore) FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastMethod = "FetchUserWatchedGeokrety"
	return []db.GeokretListItem{sampleGeokret(1)}, nil
}

func (m *mockStatsStore) FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]db.UserCountryVisit, error) {
	m.lastMethod = "FetchUserCountries"
	return []db.UserCountryVisit{{CountryCode: "PL", MoveCount: 1, FirstVisit: time.Now().UTC(), LastVisit: time.Now().UTC(), Flag: "🇵🇱"}}, nil
}

func (m *mockStatsStore) FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]db.UserWaypointVisit, error) {
	m.lastMethod = "FetchUserWaypoints"
	return []db.UserWaypointVisit{{WaypointCode: "OC146C3", VisitCount: 1, FirstVisitedAt: time.Now().UTC(), LastVisitedAt: time.Now().UTC()}}, nil
}

func (m *mockStatsStore) FetchPictureList(ctx context.Context, filters db.PictureFilters, limit, offset int) ([]db.PictureInfo, error) {
	m.lastMethod = "FetchPictureList"
	return []db.PictureInfo{{ID: 1, GeokretID: int64Ptr(1), GeokretGKID: gkidPtr(1), MoveID: int64Ptr(9), UserID: int64Ptr(1), AuthorID: int64Ptr(1), AuthorUsername: testStringPtr("alice"), CreatedOn: time.Now().UTC()}}, nil
}

func (m *mockStatsStore) FetchPictureListByIDs(ctx context.Context, pictureIDs []int64) ([]db.PictureInfo, error) {
	m.lastMethod = "FetchPictureListByIDs"
	m.lastPictureIDs = append([]int64(nil), pictureIDs...)
	rows := make([]db.PictureInfo, 0, len(pictureIDs))
	for _, pictureID := range pictureIDs {
		rows = append(rows, db.PictureInfo{ID: pictureID, GeokretID: int64Ptr(1), GeokretGKID: gkidPtr(1), MoveID: int64Ptr(9), UserID: int64Ptr(1), AuthorID: int64Ptr(1), AuthorUsername: testStringPtr("alice"), CreatedOn: time.Now().UTC()})
	}
	return rows, nil
}

func (m *mockStatsStore) FetchPicture(ctx context.Context, pictureID int64) (db.PictureInfo, error) {
	m.lastMethod = "FetchPicture"
	return db.PictureInfo{ID: pictureID, GeokretID: int64Ptr(1), GeokretGKID: gkidPtr(1), MoveID: int64Ptr(9), UserID: int64Ptr(1), AuthorID: int64Ptr(1), AuthorUsername: testStringPtr("alice"), CreatedOn: time.Now().UTC()}, nil
}

func TestGetGeokretyDetailsIncludesRelationshipLinks(t *testing.T) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	r := withRouteParams(httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/GK0001", nil), "gkid", "GK0001")
	w := httptest.NewRecorder()

	h.GetGeokretyDetailsByGkId(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	payload := decodePayload(t, w)
	data := payload["data"].(map[string]any)
	if got := data["type"]; got != "geokrety" {
		t.Fatalf("data.type = %#v, want geokrety", got)
	}
	attributes := data["attributes"].(map[string]any)
	if _, ok := attributes["type_icon_url"]; ok {
		t.Fatal("type_icon_url should not be present")
	}
	if got := attributes["collectible"]; got != true {
		t.Fatalf("attributes.collectible = %#v, want true", got)
	}
	relationships := data["relationships"].(map[string]any)
	owner := relationships["owner"].(map[string]any)
	ownerLinks := owner["links"].(map[string]any)
	if got := ownerLinks["related"]; got != "/api/v3/users/1" {
		t.Fatalf("owner related link = %#v, want /api/v3/users/1", got)
	}
	ownerData := owner["data"].(map[string]any)
	ownerAttributes := ownerData["attributes"].(map[string]any)
	if got := ownerAttributes["username"]; got != "owner" {
		t.Fatalf("owner username = %#v, want owner", got)
	}
	if got := relationships["last_position"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/moves/77" {
		t.Fatalf("last_position related link = %#v, want /api/v3/moves/77", got)
	}
	if got := relationships["stats"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/geokrety/GK0001/stats" {
		t.Fatalf("stats related link = %#v, want /api/v3/geokrety/GK0001/stats", got)
	}
	if got := relationships["finders"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/geokrety/GK0001/finders" {
		t.Fatalf("finders related link = %#v, want /api/v3/geokrety/GK0001/finders", got)
	}
}

func TestGetUserDetailsUsesDateOnlyAndHomeCountryRelationship(t *testing.T) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	r := withRouteParams(httptest.NewRequest(http.MethodGet, "/api/v3/users/1", nil), "id", "1")
	w := httptest.NewRecorder()

	h.GetUserDetails(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	payload := decodePayload(t, w)
	data := payload["data"].(map[string]any)
	attributes := data["attributes"].(map[string]any)
	if joinedAt, ok := attributes["joined_at"].(string); !ok || strings.Contains(joinedAt, "T") {
		t.Fatalf("joined_at = %#v, want date-only string", attributes["joined_at"])
	}
	if _, ok := attributes["home_country"]; ok {
		t.Fatal("home_country should be exposed through relationship data")
	}
	homeCountry := data["relationships"].(map[string]any)["home_country"].(map[string]any)
	if got := homeCountry["links"].(map[string]any)["related"]; got != "/api/v3/countries/PL" {
		t.Fatalf("home_country related link = %#v, want /api/v3/countries/PL", got)
	}
	homeCountryData := homeCountry["data"].(map[string]any)
	if got := homeCountryData["attributes"].(map[string]any)["flag"]; got != "🇵🇱" {
		t.Fatalf("home_country flag = %#v, want 🇵🇱", got)
	}
	if got := data["relationships"].(map[string]any)["stats"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/users/1/stats" {
		t.Fatalf("stats related link = %#v, want /api/v3/users/1/stats", got)
	}
}

func TestGetMoveDetailsUsesRootSelfLinkAndPicturesRelationship(t *testing.T) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	r := withRouteParams(httptest.NewRequest(http.MethodGet, "/api/v3/moves/9", nil), "id", "9")
	w := httptest.NewRecorder()

	h.GetMoveDetails(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	payload := decodePayload(t, w)
	data := payload["data"].(map[string]any)
	if got := data["links"].(map[string]any)["self"]; got != "/api/v3/moves/9" {
		t.Fatalf("move self link = %#v, want /api/v3/moves/9", got)
	}
	attributes := data["attributes"].(map[string]any)
	if got := attributes["type"]; got != "drop" {
		t.Fatalf("attributes.type = %#v, want drop", got)
	}
	relationships := data["relationships"].(map[string]any)
	if got := relationships["pictures"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/pictures?move=9" {
		t.Fatalf("pictures related link = %#v, want /api/v3/pictures?move=9", got)
	}
	if got := relationships["geokret"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/geokrety/GK0001" {
		t.Fatalf("geokret related link = %#v, want /api/v3/geokrety/GK0001", got)
	}
}

func TestGetMoveListParsesFiltersAndBatchIDs(t *testing.T) {
	store := &mockStatsStore{}
	h := NewStatsHandler(store, zap.NewNop())
	r := httptest.NewRequest(http.MethodGet, "/api/v3/moves?geokret=GK0001&user=1&country=pl&waypoint=oc146c3&date_from=2026-03-01&date_to=2026-03-29&limit=2", nil)
	w := httptest.NewRecorder()

	h.GetMoveList(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if store.lastMethod != "FetchMoveList" {
		t.Fatalf("store method = %s, want FetchMoveList", store.lastMethod)
	}
	if store.lastMoveFilters.GeokretID == nil || *store.lastMoveFilters.GeokretID != 1 {
		t.Fatalf("geokret filter = %#v, want 1", store.lastMoveFilters.GeokretID)
	}
	if store.lastMoveFilters.Country == nil || *store.lastMoveFilters.Country != "PL" {
		t.Fatalf("country filter = %#v, want PL", store.lastMoveFilters.Country)
	}
	if store.lastMoveFilters.Waypoint == nil || *store.lastMoveFilters.Waypoint != "OC146C3" {
		t.Fatalf("waypoint filter = %#v, want OC146C3", store.lastMoveFilters.Waypoint)
	}
	if store.lastMoveFilters.DateTo == nil || store.lastMoveFilters.DateTo.UTC().Format("2006-01-02") != "2026-03-30" {
		t.Fatalf("date_to filter = %#v, want exclusive 2026-03-30", store.lastMoveFilters.DateTo)
	}

	batchReq := httptest.NewRequest(http.MethodGet, "/api/v3/moves?ids=7,9", nil)
	batchResp := httptest.NewRecorder()
	h.GetMoveList(batchResp, batchReq)
	if batchResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for batch ids, got %d", batchResp.Code)
	}
	if store.lastMethod != "FetchMoveListByIDs" {
		t.Fatalf("store method = %s, want FetchMoveListByIDs", store.lastMethod)
	}
	if len(store.lastMoveIDs) != 2 || store.lastMoveIDs[0] != 7 || store.lastMoveIDs[1] != 9 {
		t.Fatalf("move ids = %#v, want [7 9]", store.lastMoveIDs)
	}
	batchPayload := decodePayload(t, batchResp)
	if got := len(batchPayload["data"].([]any)); got != 2 {
		t.Fatalf("batch data length = %d, want 2", got)
	}
}
