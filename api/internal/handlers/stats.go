package handlers

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type StatsStore interface {
	FetchGlobalStats(ctx context.Context) (db.GlobalStats, error)
	FetchCountries(ctx context.Context, limit, offset int) ([]db.CountryStats, error)
	FetchLeaderboard(ctx context.Context, limit, offset int) ([]db.LeaderboardUser, error)
	FetchRecentMoves(ctx context.Context, limit, offset int) ([]db.RecentMove, error)
	FetchRecentBorn(ctx context.Context, limit, offset int) ([]db.RecentBorn, error)
	FetchRecentLoved(ctx context.Context, limit, offset int) ([]db.RecentLoved, error)
	FetchRecentWatched(ctx context.Context, limit, offset int) ([]db.RecentWatched, error)
	FetchRecentActiveCountries(ctx context.Context, limit, offset int) ([]db.ActiveCountry, error)
	FetchRecentActiveWaypoints(ctx context.Context, limit, offset int) ([]db.ActiveWaypoint, error)
	FetchRecentRegisteredUsers(ctx context.Context, limit, offset int) ([]db.RecentRegisteredUser, error)
	FetchRecentActiveUsers(ctx context.Context, limit, offset int) ([]db.RecentActiveUser, error)
	FetchHourlyHeatmap(ctx context.Context, limit, offset int) ([]db.HourlyHeatmapCell, error)
	FetchCountryFlows(ctx context.Context, limit, offset int) ([]db.CountryFlow, error)
	FetchTopCaches(ctx context.Context, limit, offset int) ([]db.TopCache, error)
	FetchFirstFinderLeaderboard(ctx context.Context, limit, offset int) ([]db.FirstFinderLeaderboardEntry, error)
	FetchDistanceRecords(ctx context.Context, limit, offset int) ([]db.DistanceRecord, error)
	FetchUserNetwork(ctx context.Context, userID int64, limit, offset int) ([]db.UserNetworkEdge, error)
	FetchGeokretTimeline(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretTimelineEvent, error)
	FetchGeokretCirculation(ctx context.Context, geokretID int64) (db.GeokretCirculation, error)
	FetchGeokrety(ctx context.Context, geokretID int64) (db.GeokretDetails, error)
	FetchGeokretyMoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.MoveRecord, error)
	FetchGeokretyMoveDetails(ctx context.Context, geokretID, moveID int64) (db.MoveRecord, error)
	FetchGeokretyLoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error)
	FetchGeokretyWatches(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error)
	FetchGeokretyPictures(ctx context.Context, geokretID int64, limit, offset int) ([]db.PictureInfo, error)
	SearchGeokrety(ctx context.Context, query string, limit, offset int) ([]db.GeokretListItem, error)
	FetchGeokretyCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretCountryVisit, error)
	FetchGeokretyWaypoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretWaypointVisit, error)
	FetchGeokretyStatsMapCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.CountryCount, error)
	FetchGeokretyStatsElevation(ctx context.Context, geokretID int64, limit, offset int) ([]db.ElevationPoint, error)
	FetchGeokretyStatsHeatmapDays(ctx context.Context, geokretID int64, limit, offset int) ([]db.DayHeatmapCell, error)
	FetchGeokretyTripPoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.TripPoint, error)
	FetchCountryDetails(ctx context.Context, countryCode string) (db.CountryDetails, error)
	FetchCountryGeokrety(ctx context.Context, countryCode string, limit, offset int) ([]db.GeokretListItem, error)
	FetchWaypoint(ctx context.Context, waypointCode string) (db.WaypointDetails, error)
	FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error)
	FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error)
	SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]db.WaypointSummary, error)
	FetchUserDetails(ctx context.Context, userID int64) (db.UserDetails, error)
	FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserPictures(ctx context.Context, userID int64, limit, offset int) ([]db.PictureInfo, error)
	FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]db.UserCountryVisit, error)
	FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]db.UserWaypointVisit, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]db.UserSearchResult, error)
	FetchUserStatsHeatmapDays(ctx context.Context, userID int64, limit, offset int) ([]db.DayHeatmapCell, error)
	FetchUserStatsHeatmapHours(ctx context.Context, userID int64, limit, offset int) ([]db.HourHeatmapCell, error)
	FetchUserStatsMapCountries(ctx context.Context, userID int64, limit, offset int) ([]db.CountryCount, error)
	FetchPicture(ctx context.Context, pictureID int64) (db.PictureInfo, error)
}

type StatsHandler struct {
	store  StatsStore
	logger *zap.Logger
}

func NewStatsHandler(store StatsStore, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{store: store, logger: logger}
}

func (h *StatsHandler) GetKPIs(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	stats, err := h.store.FetchGlobalStats(r.Context())
	if err != nil {
		h.logger.Error("failed to fetch global stats", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch global stats")
		return
	}
	writeEnvelope(w, http.StatusOK, stats, started, 1, 0, 1)
}

func (h *StatsHandler) GetCountries(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	limit := queryInt(r, "limit", 50, 1, 1000)
	offset := queryInt(r, "offset", 0, 0, 1_000_000)
	rows, err := h.store.FetchCountries(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to fetch countries", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch countries")
		return
	}
	writeEnvelope(w, http.StatusOK, rows, started, limit, offset, len(rows))
}

func (h *StatsHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	limit := queryInt(r, "limit", 20, 1, 1000)
	offset := queryInt(r, "offset", 0, 0, 1_000_000)
	rows, err := h.store.FetchLeaderboard(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to fetch leaderboard", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}
	writeEnvelope(w, http.StatusOK, rows, started, limit, offset, len(rows))
}

func (h *StatsHandler) GetRecentMoves(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentMoves(ctx, limit, offset)
	}, "failed to fetch recent activity")
}

func (h *StatsHandler) GetRecentBorn(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentBorn(ctx, limit, offset)
	}, "failed to fetch recently born geokrety")
}

func (h *StatsHandler) GetRecentLoved(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentLoved(ctx, limit, offset)
	}, "failed to fetch recently loved geokrety")
}

func (h *StatsHandler) GetRecentWatched(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentWatched(ctx, limit, offset)
	}, "failed to fetch recently watched geokrety")
}

func (h *StatsHandler) GetRecentActiveCountries(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentActiveCountries(ctx, limit, offset)
	}, "failed to fetch recent active countries")
}

func (h *StatsHandler) GetRecentActiveWaypoints(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentActiveWaypoints(ctx, limit, offset)
	}, "failed to fetch recent active waypoints")
}

func (h *StatsHandler) GetRecentRegisteredUsers(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentRegisteredUsers(ctx, limit, offset)
	}, "failed to fetch recent registered users")
}

func (h *StatsHandler) GetRecentActiveUsers(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchRecentActiveUsers(ctx, limit, offset)
	}, "failed to fetch recent active users")
}

func (h *StatsHandler) GetHourlyHeatmap(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchHourlyHeatmap(ctx, limit, offset)
	}, "failed to fetch hourly heatmap")
}

func (h *StatsHandler) GetCountryFlows(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchCountryFlows(ctx, limit, offset)
	}, "failed to fetch country flows")
}

func (h *StatsHandler) GetTopCaches(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchTopCaches(ctx, limit, offset)
	}, "failed to fetch top caches")
}

func (h *StatsHandler) GetFirstFinderLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchFirstFinderLeaderboard(ctx, limit, offset)
	}, "failed to fetch first finder leaderboard")
}

func (h *StatsHandler) GetDistanceRecords(w http.ResponseWriter, r *http.Request) {
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchDistanceRecords(ctx, limit, offset)
	}, "failed to fetch distance records")
}

func (h *StatsHandler) GetUserNetwork(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchUserNetwork(ctx, userID, limit, offset)
	}, "failed to fetch user network")
}

func (h *StatsHandler) GetGeokretTimeline(w http.ResponseWriter, r *http.Request) {
	geokretID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	h.getRecentList(w, r, func(ctx context.Context, limit, offset int) (interface{}, error) {
		return h.store.FetchGeokretTimeline(ctx, geokretID, limit, offset)
	}, "failed to fetch geokret timeline")
}

func (h *StatsHandler) GetGeokretCirculation(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	geokretID, ok := parseInt64Param(w, r, "id")
	if !ok {
		return
	}
	row, err := h.store.FetchGeokretCirculation(r.Context(), geokretID)
	if err != nil {
		h.logger.Error("failed to fetch geokret circulation", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch geokret circulation")
		return
	}
	writeEnvelope(w, http.StatusOK, row, started, 1, 0, 1)
}

func (h *StatsHandler) getRecentList(
	w http.ResponseWriter,
	r *http.Request,
	fetch func(context.Context, int, int) (interface{}, error),
	errMsg string,
) {
	started := time.Now()
	limit := queryInt(r, "limit", 20, 1, 1000)
	offset := queryInt(r, "offset", 0, 0, 1_000_000)
	rows, err := fetch(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error(errMsg, zap.Error(err))
		writeError(w, http.StatusInternalServerError, errMsg)
		return
	}
	count := 0
	v := reflect.ValueOf(rows)
	if v.IsValid() && v.Kind() == reflect.Slice {
		count = v.Len()
	}
	writeEnvelope(w, http.StatusOK, rows, started, limit, offset, count)
}

func parseInt64Param(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value := chi.URLParam(r, key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		writeError(w, http.StatusBadRequest, "invalid identifier")
		return 0, false
	}
	return parsed, true
}
