package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	"go.uber.org/zap"
)

type StatsHandler struct {
	store  *db.Store
	logger *zap.Logger
}

func NewStatsHandler(store *db.Store, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{store: store, logger: logger}
}

func (h *StatsHandler) GetGlobal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.store.FetchGlobalStats(ctx)
	if err != nil {
		h.logger.Error("failed to fetch global stats", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch global stats")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *StatsHandler) GetRecent(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	ctx := r.Context()
	rows, err := h.store.FetchRecentMoves(ctx, limit)
	if err != nil {
		h.logger.Error("failed to fetch recent activity", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch recent activity")
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *StatsHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 10)
	ctx := r.Context()
	rows, err := h.store.FetchLeaderboard(ctx, limit)
	if err != nil {
		h.logger.Error("failed to fetch leaderboard", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch leaderboard")
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *StatsHandler) GetCountries(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 250)
	ctx := r.Context()
	rows, err := h.store.FetchCountries(ctx, limit)
	if err != nil {
		h.logger.Error("failed to fetch countries", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to fetch countries")
		return
	}

	response := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		response = append(response, map[string]interface{}{
			"code":        row.Code,
			"name":        row.Name,
			"flag":        row.Flag,
			"movesCount":  row.MovesCount,
			"usersHome":   row.UsersHome,
			"activeUsers": row.ActiveUsers,
			"movesByType": map[string]int64{
				"dropped": row.Dropped,
				"dipped":  row.Dipped,
				"seen":    row.Seen,
			},
			"loves":            row.Loves,
			"pictures":         row.Pictures,
			"pointsSum":        row.PointsSum,
			"pointsSumMoves":   row.PointsSumMoves,
			"geokretyInCache":  row.GeokretyInCache,
			"geokretyLost":     row.GeokretyLost,
			"avgPointsPerMove": row.AvgPointsPerMove,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func queryInt(r *http.Request, key string, fallback int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return fallback
	}
	if parsed > 1000 {
		return 1000
	}
	return parsed
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
