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
