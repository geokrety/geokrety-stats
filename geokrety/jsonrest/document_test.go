package jsonrest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMetaAndDocument(t *testing.T) {
	started := time.Now().Add(-15 * time.Millisecond)
	meta := NewMeta(started).WithCursor(CursorRequest{Limit: 20}, true).WithSort("-created_at,id")
	doc := NewDocument([]Resource{{ID: "1", Type: "user"}}, meta, Links{"self": "/api/v3/users?limit=20"})

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	metaPayload := payload["meta"].(map[string]any)
	if metaPayload["execution_time_ms"] == nil {
		t.Fatalf("execution_time_ms missing from meta")
	}
	if got := metaPayload["has_more"]; got != true {
		t.Fatalf("has_more = %#v, want true", got)
	}
	if got := payload["links"].(map[string]any)["self"]; got != "/api/v3/users?limit=20" {
		t.Fatalf("links.self = %#v, want /api/v3/users?limit=20", got)
	}
}

func TestLinksBuilders(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v3/users?status=active", nil)
	pageLinks := PageLinks(r, PageRequest{Page: 2, PerPage: 10}, 5, "page", "per_page")
	if got := pageLinks["self"]; got != "/api/v3/users?page=2&per_page=10&status=active" && got != "/api/v3/users?per_page=10&page=2&status=active" {
		t.Fatalf("page self = %q", got)
	}
	if got := pageLinks["next"]; got == "" {
		t.Fatalf("expected next page link")
	}

	cursor := EncodeCursor(CurrentCursorVersion, 30)
	cursorLinks := CursorLinks(r, CursorRequest{Limit: 20}, &cursor, "limit", "cursor")
	if got := cursorLinks["self"]; got != "/api/v3/users?limit=20&status=active" && got != "/api/v3/users?status=active&limit=20" {
		t.Fatalf("cursor self = %q", got)
	}
	if got := cursorLinks["next"]; got == "" {
		t.Fatalf("expected next cursor link")
	}
}

func TestNewErrorDocument(t *testing.T) {
	doc := NewErrorDocument("INVALID_CURSOR", "invalid cursor", time.Date(2026, 3, 29, 1, 2, 3, 0, time.UTC))
	if doc.Error.Code != "INVALID_CURSOR" {
		t.Fatalf("Code = %q, want INVALID_CURSOR", doc.Error.Code)
	}
	if doc.Timestamp != "2026-03-29T01:02:03Z" {
		t.Fatalf("Timestamp = %q, want 2026-03-29T01:02:03Z", doc.Timestamp)
	}
}
