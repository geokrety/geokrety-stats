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

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	if err := h.store.Ping(ctx); err != nil {
		status = "degraded"
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": status,
			"error":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          status,
		"activeWsClients": h.hub.ActiveConnections(),
		"serverTime":      time.Now().UTC().Format(time.RFC3339),
	})
}
