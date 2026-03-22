package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/pagination"
)

func TestWriteEnvelopeForOffsetRequestIncludesPaginationMetadata(t *testing.T) {
	started := time.Now().Add(-10 * time.Millisecond)
	request := pagination.OffsetRequest{Limit: 4, Offset: 4}
	totalItems := 42
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/x", nil)

	hasMore := true
	writeEnvelopeForOffsetRequest(w, r, http.StatusOK, []int{1, 2, 3, 4}, started, request, &totalItems, &hasMore, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	meta := payload["meta"].(map[string]any)
	page := meta["pagination"].(map[string]any)
	if got := page["type"]; got != "offset" {
		t.Fatalf("pagination.type = %#v, want offset", got)
	}
	if got := page["totalItems"]; got != float64(42) {
		t.Fatalf("pagination.totalItems = %#v, want 42", got)
	}
	if got := page["count"]; got != float64(4) {
		t.Fatalf("pagination.count = %#v, want 4", got)
	}
	if got := page["returned"]; got != float64(4) {
		t.Fatalf("pagination.returned = %#v, want 4", got)
	}
	if got := page["hasMore"]; got != true {
		t.Fatalf("pagination.hasMore = %#v, want true", got)
	}
	if got := page["nextCursor"]; got == nil || got == "" {
		t.Fatalf("pagination.nextCursor = %#v, want non-empty string", got)
	}
}
