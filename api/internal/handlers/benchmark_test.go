package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func BenchmarkGetKPIs(b *testing.B) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/api/v3/stats/kpis", nil)
		w := httptest.NewRecorder()
		h.GetKPIs(w, r)
	}
}

func BenchmarkGetRecentMoves(b *testing.B) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/recent-moves?limit=20&offset=0", nil)
		w := httptest.NewRecorder()
		h.GetRecentMoves(w, r)
	}
}

func BenchmarkGetHourlyHeatmap(b *testing.B) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/api/v3/stats/hourly-heatmap?limit=20&offset=0", nil)
		w := httptest.NewRecorder()
		h.GetHourlyHeatmap(w, r)
	}
}

func BenchmarkGetGeokrety(b *testing.B) {
	h := NewStatsHandler(&mockStatsStore{}, zap.NewNop())
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/1", nil)
		r = withRouteParam(r, "id", "1")
		w := httptest.NewRecorder()
		h.GetGeokrety(w, r)
	}
}
