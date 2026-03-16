package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/ws"
	"go.uber.org/zap"
)

type SystemStore interface {
	Ping(ctx context.Context) error
}

type SystemHandler struct {
	store  SystemStore
	hub    *ws.Hub
	logger *zap.Logger
}

func NewSystemHandler(store SystemStore, hub *ws.Hub, logger *zap.Logger) *SystemHandler {
	return &SystemHandler{store: store, hub: hub, logger: logger}
}

func (h *SystemHandler) Healtz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"serverTime": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	if err := h.store.Ping(ctx); err != nil {
		h.logger.Warn("database health check failed", zap.Error(err))
		status = "degraded"
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": status,
			"error":  "database_unavailable",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          status,
		"activeWsClients": h.hub.ActiveConnections(),
		"serverTime":      time.Now().UTC().Format(time.RFC3339),
	})
}
