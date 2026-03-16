package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/metrics"
	"github.com/geokrety/geokrety-stats-api/internal/ws"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type mockSystemStore struct {
	err error
}

func (m *mockSystemStore) Ping(ctx context.Context) error {
	return m.err
}

func newTestHub() *ws.Hub {
	reg := prometheus.NewRegistry()
	mc := metrics.New(reg)
	return ws.NewHub(zap.NewNop(), mc, time.Second)
}

func TestHealthOK(t *testing.T) {
	h := NewSystemHandler(&mockSystemStore{}, newTestHub(), zap.NewNop())
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthDegraded(t *testing.T) {
	h := NewSystemHandler(&mockSystemStore{err: errors.New("db down")}, newTestHub(), zap.NewNop())
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHealtzOK(t *testing.T) {
	h := NewSystemHandler(&mockSystemStore{}, newTestHub(), zap.NewNop())
	r := httptest.NewRequest(http.MethodGet, "/healtz", nil)
	w := httptest.NewRecorder()

	h.Healtz(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
