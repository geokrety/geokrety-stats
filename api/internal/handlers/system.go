package handlers

import (
	"context"
	"encoding/xml"
	"net/http"
	"strings"
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

type livenessStatus struct {
	XMLName    xml.Name `json:"-" xml:"liveness"`
	Status     string   `json:"status" xml:"status"`
	ServerTime string   `json:"serverTime" xml:"serverTime"`
}

type healthStatus struct {
	XMLName         xml.Name `json:"-" xml:"health"`
	Status          string   `json:"status" xml:"status"`
	ActiveWsClients int64    `json:"activeWsClients,omitempty" xml:"activeWsClients,omitempty"`
	ServerTime      string   `json:"serverTime,omitempty" xml:"serverTime,omitempty"`
	Error           string   `json:"error,omitempty" xml:"error,omitempty"`
}

func NewSystemHandler(store SystemStore, hub *ws.Hub, logger *zap.Logger) *SystemHandler {
	return &SystemHandler{store: store, hub: hub, logger: logger}
}

func (h *SystemHandler) Healtz(w http.ResponseWriter, r *http.Request) {
	payload := livenessStatus{
		Status:     "ok",
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}
	if acceptsXML(r) {
		writeXML(w, http.StatusOK, payload)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := "ok"
	if err := h.store.Ping(ctx); err != nil {
		h.logger.Warn("database health check failed", zap.Error(err))
		status = "degraded"
		payload := healthStatus{Status: status, Error: "database_unavailable"}
		if acceptsXML(r) {
			writeXML(w, http.StatusServiceUnavailable, payload)
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, payload)
		return
	}

	payload := healthStatus{
		Status:          status,
		ActiveWsClients: h.hub.ActiveConnections(),
		ServerTime:      time.Now().UTC().Format(time.RFC3339),
	}
	if acceptsXML(r) {
		writeXML(w, http.StatusOK, payload)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func acceptsXML(r *http.Request) bool {
	if r == nil {
		return false
	}
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml")
}
