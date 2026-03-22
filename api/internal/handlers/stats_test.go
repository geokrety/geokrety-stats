package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type mockStatsStore struct {
	failMethod   string
	noRowsMethod string
	lastMethod   string
	lastLimit    int
	lastOffset   int
}

func (m *mockStatsStore) maybeFail(method string) error {
	m.lastMethod = method
	if m.noRowsMethod == method {
		return sql.ErrNoRows
	}
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
	return []db.RecentBorn{{ID: 1, GKID: *mustGeokretIdPtr(1), Name: "GK", BornAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentLoved(ctx context.Context, limit, offset int) ([]db.RecentLoved, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentLoved"); err != nil {
		return nil, err
	}
	return []db.RecentLoved{{GeoKretID: 1, GKID: mustGeokretIdPtr(255), Username: "u", LovedAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchRecentWatched(ctx context.Context, limit, offset int) ([]db.RecentWatched, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchRecentWatched"); err != nil {
		return nil, err
	}
	return []db.RecentWatched{{GeoKretID: 1, GKID: mustGeokretIdPtr(255), Username: "u", WatchedAt: time.Now()}}, nil
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
	return []db.DistanceRecord{{GeoKretID: 1, GKID: mustGeokretIdPtr(255), GeoKretName: "GK", KMTotal: 100, Rank: 1}}, nil
}

func (m *mockStatsStore) FetchStatsDormancy(ctx context.Context, limit, offset int) ([]db.DormancyRecord, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchStatsDormancy"); err != nil {
		return nil, err
	}
	return []db.DormancyRecord{{GeokretID: 1, GKID: mustGeokretIdPtr(255), GeokretName: "GK", DormancySeconds: 86400, DormancyDays: 1}}, nil
}

func (m *mockStatsStore) FetchStatsMultiplierVelocity(ctx context.Context, limit, offset int) ([]db.MultiplierVelocityRecord, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchStatsMultiplierVelocity"); err != nil {
		return nil, err
	}
	return []db.MultiplierVelocityRecord{{GeokretID: 1, GKID: mustGeokretIdPtr(255), GeokretName: "GK", AvgDelta: 0.25}}, nil
}

func (m *mockStatsStore) FetchCountryList(ctx context.Context, limit, offset int) ([]db.CountryDetails, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchCountryList"); err != nil {
		return nil, err
	}
	continentName := "Europe"
	return []db.CountryDetails{{Code: "PL", ContinentName: &continentName, CurrentGeokrety: 1, Flag: "🇵🇱"}}, nil
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
	return db.GeokretCirculation{GeoKretID: geokretID, GKID: mustGeokretIdPtr(255), GeoKretName: "GK", Users: 2, Interactions: 6, AvgPerUser: 3}, nil
}

func (m *mockStatsStore) FetchGeokrety(ctx context.Context, geokretID int64) (db.GeokretDetails, error) {
	if err := m.maybeFail("FetchGeokrety"); err != nil {
		return db.GeokretDetails{}, err
	}
	return db.GeokretDetails{GeokretListItem: db.GeokretListItem{ID: geokretID, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchGeokretyList(ctx context.Context, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyList"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchGeokretyListTotal(ctx context.Context) (int64, error) {
	if m.noRowsMethod == "FetchGeokretyListTotal" {
		m.lastMethod = "FetchGeokretyListTotal"
		return 0, sql.ErrNoRows
	}
	if m.failMethod == "FetchGeokretyListTotal" {
		m.lastMethod = "FetchGeokretyListTotal"
		return 0, errors.New("boom")
	}
	return 42, nil
}

func (m *mockStatsStore) FetchGeokretyByGKID(ctx context.Context, gkid int64) (db.GeokretDetails, error) {
	if err := m.maybeFail("FetchGeokretyByGKID"); err != nil {
		return db.GeokretDetails{}, err
	}
	return db.GeokretDetails{GeokretListItem: db.GeokretListItem{ID: gkid, GKID: mustGeokretIdPtr(gkid), Name: "GK"}}, nil
}

func mustGeokretId(t *testing.T, value int64) *geokrety.GeokretId {
	t.Helper()
	parsed, err := geokrety.FromInt(value)
	if err != nil {
		t.Fatalf("FromInt(%d) returned error: %v", value, err)
	}
	return parsed
}

func mustGeokretIdPtr(value int64) *geokrety.GeokretId {
	parsed, err := geokrety.FromInt(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func (m *mockStatsStore) ResolveGeokretID(ctx context.Context, gkid int64) (int64, error) {
	if err := m.maybeFail("ResolveGeokretID"); err != nil {
		return 0, err
	}
	return gkid, nil
}

func (m *mockStatsStore) FetchGeokretyMoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.MoveRecord, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyMoves"); err != nil {
		return nil, err
	}
	return []db.MoveRecord{{ID: 1, GeokretID: geokretID, MovedOn: time.Now(), CreatedOn: time.Now()}}, nil
}

func (m *mockStatsStore) FetchGeokretyMoveDetails(ctx context.Context, geokretID, moveID int64) (db.MoveRecord, error) {
	if err := m.maybeFail("FetchGeokretyMoveDetails"); err != nil {
		return db.MoveRecord{}, err
	}
	return db.MoveRecord{ID: moveID, GeokretID: geokretID, MovedOn: time.Now(), CreatedOn: time.Now()}, nil
}

func (m *mockStatsStore) FetchGeokretyLoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyLoves"); err != nil {
		return nil, err
	}
	return []db.SocialUserEntry{{UserID: 1, Username: "u", At: time.Now()}}, nil
}

func (m *mockStatsStore) FetchGeokretyWatches(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyWatches"); err != nil {
		return nil, err
	}
	return []db.SocialUserEntry{{UserID: 1, Username: "u", At: time.Now()}}, nil
}

func (m *mockStatsStore) FetchGeokretyPictures(ctx context.Context, geokretID int64, limit, offset int) ([]db.PictureInfo, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyPictures"); err != nil {
		return nil, err
	}
	return []db.PictureInfo{{ID: 1, GeokretID: &geokretID, CreatedOn: time.Now()}}, nil
}

func (m *mockStatsStore) SearchGeokrety(ctx context.Context, query string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("SearchGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: query}}, nil
}

func (m *mockStatsStore) FetchGeokretyCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretCountryVisit, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyCountries"); err != nil {
		return nil, err
	}
	return []db.GeokretCountryVisit{{CountryCode: "PL", FirstVisitedAt: time.Now(), MoveCount: 1}}, nil
}

func (m *mockStatsStore) FetchGeokretyWaypoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretWaypointVisit, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyWaypoints"); err != nil {
		return nil, err
	}
	return []db.GeokretWaypointVisit{{WaypointCode: "GC123", VisitCount: 1, FirstVisitedAt: time.Now(), LastVisitedAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchGeokretyStatsMapCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.CountryCount, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyStatsMapCountries"); err != nil {
		return nil, err
	}
	return []db.CountryCount{{CountryCode: "PL", Value: 1}}, nil
}

func (m *mockStatsStore) FetchGeokretyStatsElevation(ctx context.Context, geokretID int64, limit, offset int) ([]db.ElevationPoint, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyStatsElevation"); err != nil {
		return nil, err
	}
	return []db.ElevationPoint{{MoveID: 1, MovedOn: time.Now(), Elevation: 100}}, nil
}

func (m *mockStatsStore) FetchGeokretyStatsHeatmapDays(ctx context.Context, geokretID int64, limit, offset int) ([]db.DayHeatmapCell, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyStatsHeatmapDays"); err != nil {
		return nil, err
	}
	return []db.DayHeatmapCell{{Day: time.Now(), MoveCount: 1}}, nil
}

func (m *mockStatsStore) FetchGeokretyTripPoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.TripPoint, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchGeokretyTripPoints"); err != nil {
		return nil, err
	}
	return []db.TripPoint{{MoveID: 1, MovedOn: time.Now(), Lat: 1, Lon: 2}}, nil
}

func (m *mockStatsStore) FetchCountryDetails(ctx context.Context, countryCode string) (db.CountryDetails, error) {
	if err := m.maybeFail("FetchCountryDetails"); err != nil {
		return db.CountryDetails{}, err
	}
	continentName := "Europe"
	return db.CountryDetails{Code: countryCode, ContinentName: &continentName}, nil
}

func (m *mockStatsStore) FetchCountryGeokrety(ctx context.Context, countryCode string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchCountryGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchWaypoint(ctx context.Context, waypointCode string) (db.WaypointDetails, error) {
	if err := m.maybeFail("FetchWaypoint"); err != nil {
		return db.WaypointDetails{}, err
	}
	return db.WaypointDetails{WaypointSummary: db.WaypointSummary{WaypointCode: waypointCode, Source: "gc"}}, nil
}

func (m *mockStatsStore) FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchWaypointCurrentGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchWaypointPastGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]db.WaypointSummary, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("SearchWaypoints"); err != nil {
		return nil, err
	}
	return []db.WaypointSummary{{WaypointCode: query, Source: "gc"}}, nil
}

func (m *mockStatsStore) FetchUserDetails(ctx context.Context, userID int64) (db.UserDetails, error) {
	if err := m.maybeFail("FetchUserDetails"); err != nil {
		return db.UserDetails{}, err
	}
	return db.UserDetails{ID: userID, Username: "u", JoinedAt: time.Now()}, nil
}

func (m *mockStatsStore) FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserOwnedGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserFoundGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserLovedGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserWatchedGeokrety"); err != nil {
		return nil, err
	}
	return []db.GeokretListItem{{ID: 1, Name: "GK"}}, nil
}

func (m *mockStatsStore) FetchUserPictures(ctx context.Context, userID int64, limit, offset int) ([]db.PictureInfo, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserPictures"); err != nil {
		return nil, err
	}
	return []db.PictureInfo{{ID: 1, UserID: &userID, CreatedOn: time.Now()}}, nil
}

func (m *mockStatsStore) FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]db.UserCountryVisit, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserCountries"); err != nil {
		return nil, err
	}
	return []db.UserCountryVisit{{CountryCode: "PL", MoveCount: 1, FirstVisit: time.Now(), LastVisit: time.Now()}}, nil
}

func (m *mockStatsStore) FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]db.UserWaypointVisit, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserWaypoints"); err != nil {
		return nil, err
	}
	return []db.UserWaypointVisit{{WaypointCode: "GC123", VisitCount: 1, FirstVisitedAt: time.Now(), LastVisitedAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchUserList(ctx context.Context, limit, offset int) ([]db.UserSearchResult, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserList"); err != nil {
		return nil, err
	}
	return []db.UserSearchResult{{ID: 1, Username: "u", JoinedAt: time.Now()}}, nil
}

func (m *mockStatsStore) SearchUsers(ctx context.Context, query string, limit, offset int) ([]db.UserSearchResult, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("SearchUsers"); err != nil {
		return nil, err
	}
	return []db.UserSearchResult{{ID: 1, Username: query, JoinedAt: time.Now()}}, nil
}

func (m *mockStatsStore) FetchUserStatsContinentCoverage(ctx context.Context, userID int64, limit, offset int) ([]db.UserContinentCoverage, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserStatsContinentCoverage"); err != nil {
		return nil, err
	}
	return []db.UserContinentCoverage{{UserID: userID, Username: "u", ContinentCode: "EU", ContinentName: "Europe", Moves: 10, Share: 0.5}}, nil
}

func (m *mockStatsStore) FetchUserStatsHeatmapDays(ctx context.Context, userID int64, limit, offset int) ([]db.DayHeatmapCell, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserStatsHeatmapDays"); err != nil {
		return nil, err
	}
	return []db.DayHeatmapCell{{Day: time.Now(), MoveCount: 1}}, nil
}

func (m *mockStatsStore) FetchUserStatsHeatmapHours(ctx context.Context, userID int64, limit, offset int) ([]db.HourHeatmapCell, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserStatsHeatmapHours"); err != nil {
		return nil, err
	}
	return []db.HourHeatmapCell{{DayOfWeek: 1, HourUTC: 12, MoveCount: 1}}, nil
}

func (m *mockStatsStore) FetchUserStatsMapCountries(ctx context.Context, userID int64, limit, offset int) ([]db.CountryCount, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchUserStatsMapCountries"); err != nil {
		return nil, err
	}
	return []db.CountryCount{{CountryCode: "PL", Value: 1}}, nil
}

func (m *mockStatsStore) FetchPicture(ctx context.Context, pictureID int64) (db.PictureInfo, error) {
	if err := m.maybeFail("FetchPicture"); err != nil {
		return db.PictureInfo{}, err
	}
	return db.PictureInfo{ID: pictureID, CreatedOn: time.Now()}, nil
}

func (m *mockStatsStore) FetchPictureList(ctx context.Context, limit, offset int) ([]db.PictureInfo, error) {
	m.lastLimit, m.lastOffset = limit, offset
	if err := m.maybeFail("FetchPictureList"); err != nil {
		return nil, err
	}
	return []db.PictureInfo{{ID: 1, CreatedOn: time.Now()}}, nil
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
		{"seasonal-heatmap", "/api/v3/stats/seasonal-heatmap?limit=11&offset=1", h.GetHourlyHeatmap, "FetchHourlyHeatmap", http.StatusOK, "FetchHourlyHeatmap", 11, 1, true},
		{"country-flows", "/api/v3/stats/country-flows?limit=11&offset=1", h.GetCountryFlows, "FetchCountryFlows", http.StatusOK, "FetchCountryFlows", 11, 1, true},
		{"top-caches", "/api/v3/stats/top-caches?limit=11&offset=1", h.GetTopCaches, "FetchTopCaches", http.StatusOK, "FetchTopCaches", 11, 1, true},
		{"first-finder", "/api/v3/stats/first-finder-leaderboard?limit=11&offset=1", h.GetFirstFinderLeaderboard, "FetchFirstFinderLeaderboard", http.StatusOK, "FetchFirstFinderLeaderboard", 11, 1, true},
		{"distance-records", "/api/v3/stats/distance-records?limit=11&offset=1", h.GetDistanceRecords, "FetchDistanceRecords", http.StatusOK, "FetchDistanceRecords", 11, 1, true},
		{"dormancy", "/api/v3/stats/dormancy?limit=11&offset=1", h.GetStatsDormancy, "FetchStatsDormancy", http.StatusOK, "FetchStatsDormancy", 11, 1, true},
		{"multiplier-velocity", "/api/v3/stats/multiplier-velocity?limit=11&offset=1", h.GetStatsMultiplierVelocity, "FetchStatsMultiplierVelocity", http.StatusOK, "FetchStatsMultiplierVelocity", 11, 1, true},
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

func TestStatsHandlersSerializeGKIDStrings(t *testing.T) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	tests := []struct {
		name    string
		handler http.HandlerFunc
		target  string
		want    string
		path    []string
	}{
		{name: "recent-born", handler: h.GetRecentBorn, target: "/api/v3/geokrety/recent-born", want: "GK0001"},
		{name: "recent-loved", handler: h.GetRecentLoved, target: "/api/v3/geokrety/recent-loved", want: "GK00FF"},
		{name: "recent-watched", handler: h.GetRecentWatched, target: "/api/v3/geokrety/recent-watched", want: "GK00FF"},
		{name: "distance-records", handler: h.GetDistanceRecords, target: "/api/v3/stats/distance-records", want: "GK00FF"},
		{name: "dormancy", handler: h.GetStatsDormancy, target: "/api/v3/stats/dormancy", want: "GK00FF"},
		{name: "multiplier-velocity", handler: h.GetStatsMultiplierVelocity, target: "/api/v3/stats/multiplier-velocity", want: "GK00FF"},
		{name: "circulation", handler: h.GetGeokretCirculation, target: "/api/v3/geokrety/255/circulation", want: "GK00FF", path: []string{"id", "255"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tc.target, nil)
			if len(tc.path) > 0 {
				r = withRouteParams(r, tc.path...)
			}
			w := httptest.NewRecorder()

			tc.handler(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
			payload := decodeMap(t, w)
			data := payload["data"]
			switch typed := data.(type) {
			case []any:
				row := typed[0].(map[string]any)
				if got := row["gkid"]; got != tc.want {
					t.Fatalf("gkid = %#v, want %s", got, tc.want)
				}
			case map[string]any:
				if got := typed["gkid"]; got != tc.want {
					t.Fatalf("gkid = %#v, want %s", got, tc.want)
				}
			default:
				t.Fatalf("unexpected data payload type %T", data)
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
		{"dormancy", "FetchStatsDormancy", "/api/v3/stats/dormancy", nil},
		{"velocity", "FetchStatsMultiplierVelocity", "/api/v3/stats/multiplier-velocity", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &mockStatsStore{failMethod: tc.failMethod}
			h := NewStatsHandler(store, zap.NewNop())

			handler := map[string]http.HandlerFunc{
				"FetchGlobalStats":             h.GetKPIs,
				"FetchCountries":               h.GetCountries,
				"FetchLeaderboard":             h.GetLeaderboard,
				"FetchRecentMoves":             h.GetRecentMoves,
				"FetchHourlyHeatmap":           h.GetHourlyHeatmap,
				"FetchCountryFlows":            h.GetCountryFlows,
				"FetchTopCaches":               h.GetTopCaches,
				"FetchFirstFinderLeaderboard":  h.GetFirstFinderLeaderboard,
				"FetchDistanceRecords":         h.GetDistanceRecords,
				"FetchStatsDormancy":           h.GetStatsDormancy,
				"FetchStatsMultiplierVelocity": h.GetStatsMultiplierVelocity,
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
		rctx.URLParams.Add("id", "GKZZ")
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
