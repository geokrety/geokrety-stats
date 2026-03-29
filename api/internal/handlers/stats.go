package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type StatsStore interface {
	ResolveGeokretID(ctx context.Context, gkid int64) (int64, error)
	FetchGeokretyList(ctx context.Context, limit, offset int) ([]db.GeokretListItem, error)
	FetchGeokretyByGKIDs(ctx context.Context, gkids []int64) ([]db.GeokretListItem, error)
	FetchGeokretyByGKID(ctx context.Context, gkid int64) (db.GeokretDetails, error)
	SearchGeokrety(ctx context.Context, query string, limit, offset int) ([]db.GeokretListItem, error)
	FetchGeokretStats(ctx context.Context, geokretID int64) (db.GeokretStats, error)
	FetchMoveList(ctx context.Context, filters db.MoveFilters, limit, offset int) ([]db.MoveRecord, error)
	FetchMoveListByIDs(ctx context.Context, filters db.MoveFilters, moveIDs []int64) ([]db.MoveRecord, error)
	FetchMove(ctx context.Context, moveID int64) (db.MoveRecord, error)
	FetchGeokretyLoves(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error)
	FetchGeokretyWatches(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error)
	FetchGeokretyFinders(ctx context.Context, geokretID int64, limit, offset int) ([]db.SocialUserEntry, error)
	FetchGeokretyCountries(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretCountryVisit, error)
	FetchGeokretyWaypoints(ctx context.Context, geokretID int64, limit, offset int) ([]db.GeokretWaypointVisit, error)
	FetchCountryList(ctx context.Context, limit, offset int) ([]db.CountryDetails, error)
	FetchCountryListByCodes(ctx context.Context, codes []string) ([]db.CountryDetails, error)
	FetchCountryDetails(ctx context.Context, countryCode string) (db.CountryDetails, error)
	FetchCountryGeokrety(ctx context.Context, countryCode string, limit, offset int) ([]db.GeokretListItem, error)
	FetchWaypoint(ctx context.Context, waypointCode string) (db.WaypointDetails, error)
	FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error)
	FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]db.GeokretListItem, error)
	SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]db.WaypointSummary, error)
	FetchUserList(ctx context.Context, limit, offset int) ([]db.UserSearchResult, error)
	FetchUserListByIDs(ctx context.Context, userIDs []int64) ([]db.UserSearchResult, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]db.UserSearchResult, error)
	FetchUserDetails(ctx context.Context, userID int64) (db.UserDetails, error)
	FetchUserStats(ctx context.Context, userID int64) (db.UserStats, error)
	FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]db.GeokretListItem, error)
	FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]db.UserCountryVisit, error)
	FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]db.UserWaypointVisit, error)
	FetchPictureList(ctx context.Context, filters db.PictureFilters, limit, offset int) ([]db.PictureInfo, error)
	FetchPictureListByIDs(ctx context.Context, pictureIDs []int64) ([]db.PictureInfo, error)
	FetchPicture(ctx context.Context, pictureID int64) (db.PictureInfo, error)
}

type StatsHandler struct {
	store  StatsStore
	logger *zap.Logger
}

func NewStatsHandler(store StatsStore, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{store: store, logger: logger}
}

func (h *StatsHandler) getEntityList(w http.ResponseWriter, r *http.Request, fetch func(limit, offset int) (interface{}, error), errMsg string) {
	started := time.Now()
	req, err := queryPagination(r, 20, 1000)
	if err != nil {
		writePaginationErrorForRequest(w, r, http.StatusBadRequest, err)
		return
	}
	rows, err := fetch(req.Limit+1, req.Offset)
	if err != nil {
		h.writeStoreError(w, r, err, errMsg)
		return
	}
	pageRows, returned, hasMore := trimPaginatedPayload(rows, req.Limit)
	writeEnvelopeForOffsetRequest(w, r, http.StatusOK, pageRows, started, req, nil, &hasMore, &returned)
}

func (h *StatsHandler) writeStoreError(w http.ResponseWriter, r *http.Request, err error, message string) {
	if errors.Is(err, sql.ErrNoRows) {
		writeErrorForRequest(w, r, http.StatusNotFound, "resource not found")
		return
	}
	h.logger.Error(message, zap.Error(err))
	writeErrorForRequest(w, r, http.StatusInternalServerError, message)
}

func parseInt64Param(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value := chi.URLParam(r, key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		writeErrorForRequest(w, r, http.StatusBadRequest, "invalid identifier")
		return 0, false
	}
	return parsed, true
}

func parseInt64PathParam(w http.ResponseWriter, r *http.Request, key string) (int64, bool) {
	value := chi.URLParam(r, key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		writeErrorForRequest(w, r, http.StatusBadRequest, fmt.Sprintf("invalid path parameter %q", key))
		return 0, false
	}
	return parsed, true
}
