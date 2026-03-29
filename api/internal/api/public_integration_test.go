package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func decodeAPIJSON(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return payload
}

func TestPublicIntegrationMoveContract(t *testing.T) {
	r := newRouterForTests(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v3/moves/9", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	payload := decodeAPIJSON(t, w)
	data := payload["data"].(map[string]any)
	if got := data["type"]; got != "move" {
		t.Fatalf("data.type = %#v, want move", got)
	}
	if got := data["links"].(map[string]any)["self"]; got != "/api/v3/moves/9" {
		t.Fatalf("data.links.self = %#v, want /api/v3/moves/9", got)
	}
	relationships := data["relationships"].(map[string]any)
	if got := relationships["pictures"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/pictures?move=9" {
		t.Fatalf("pictures related link = %#v, want /api/v3/pictures?move=9", got)
	}
}

func TestPublicIntegrationGeokretContract(t *testing.T) {
	r := newRouterForTests(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v3/geokrety/GK0001", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	payload := decodeAPIJSON(t, w)
	data := payload["data"].(map[string]any)
	if got := data["type"]; got != "geokrety" {
		t.Fatalf("data.type = %#v, want geokrety", got)
	}
	relationships := data["relationships"].(map[string]any)
	owner := relationships["owner"].(map[string]any)
	if got := owner["links"].(map[string]any)["related"]; got != "/api/v3/users/1" {
		t.Fatalf("owner related link = %#v, want /api/v3/users/1", got)
	}
}

func TestPublicIntegrationUserContract(t *testing.T) {
	r := newRouterForTests(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v3/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	payload := decodeAPIJSON(t, w)
	data := payload["data"].(map[string]any)
	attributes := data["attributes"].(map[string]any)
	joinedAt := attributes["joined_at"].(string)
	if strings.Contains(joinedAt, "T") {
		t.Fatalf("joined_at = %q, want date-only format", joinedAt)
	}
	relationships := data["relationships"].(map[string]any)
	if got := relationships["home_country"].(map[string]any)["links"].(map[string]any)["related"]; got != "/api/v3/countries/PL" {
		t.Fatalf("home_country related link = %#v, want /api/v3/countries/PL", got)
	}
}
