package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	"github.com/geokrety/geokrety-stats-api/internal/ws"
	"go.uber.org/zap"
)

type SystemHandler struct {
	store  *db.Store
	hub    *ws.Hub
	logger *zap.Logger
}

func NewSystemHandler(store *db.Store, hub *ws.Hub, logger *zap.Logger) *SystemHandler {
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

type PublishRequest struct {
	Type string          `json:"type"`
	Path string          `json:"path"`
	Data json.RawMessage `json:"data"`
}

func (h *SystemHandler) PublishWS(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if req.Type == "" {
		req.Type = "stats_update"
	}
	if req.Path == "" {
		req.Path = "global"
	}

	h.hub.PublishRaw(r.Context(), req.Type, req.Path, req.Data)
	h.logger.Info("published websocket message", zap.String("type", req.Type), zap.String("path", req.Path))

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"status": "queued",
	})
}
