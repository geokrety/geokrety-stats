package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sharedjsonrest "github.com/geokrety/geokrety-stats/geokrety/jsonrest"
)

func TestWriteEnvelopeForOffsetRequestIncludesPaginationMetadata(t *testing.T) {
	started := time.Now().Add(-10 * time.Millisecond)
	cursor := sharedjsonrest.EncodeCursor(sharedjsonrest.CurrentCursorVersion, 4)
	request := sharedjsonrest.CursorRequest{Limit: 4, Offset: 4, UsedCursor: true, Cursor: cursor}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/x?limit=4&cursor="+cursor.String(), nil)

	hasMore := true
	writeEnvelopeForOffsetRequest(w, r, http.StatusOK, []int{1, 2, 3, 4}, started, request, nil, &hasMore, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	meta := payload["meta"].(map[string]any)
	if got := meta["limit"]; got != float64(4) {
		t.Fatalf("meta.limit = %#v, want 4", got)
	}
	if got := meta["has_more"]; got != true {
		t.Fatalf("meta.has_more = %#v, want true", got)
	}
	links := payload["links"].(map[string]any)
	if got := links["self"]; got == nil || got == "" {
		t.Fatalf("links.self = %#v, want non-empty string", got)
	}
	if got := links["next"]; got == nil || got == "" {
		t.Fatalf("links.next = %#v, want non-empty string", got)
	}
}

func TestWriteEnvelopeForOffsetRequestOmitsNextLinkOnTerminalPage(t *testing.T) {
	started := time.Now().Add(-10 * time.Millisecond)
	request := sharedjsonrest.CursorRequest{Limit: 4, Offset: 8}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/x?limit=4", nil)

	hasMore := false
	writeEnvelopeForOffsetRequest(w, r, http.StatusOK, []int{1, 2}, started, request, nil, &hasMore, nil)

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	meta := payload["meta"].(map[string]any)
	if got := meta["has_more"]; got != false {
		t.Fatalf("meta.has_more = %#v, want false", got)
	}
	links := payload["links"].(map[string]any)
	if _, ok := links["next"]; ok {
		t.Fatalf("links.next = %#v, want omitted on terminal page", links["next"])
	}
}
