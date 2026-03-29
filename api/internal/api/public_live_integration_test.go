//go:build liveintegration

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

const defaultLiveAPIBaseURL = "http://localhost:7415"

type liveFixtures struct {
	primaryGeokret   string
	secondaryGeokret string
	primaryUserID    string
	secondaryUserID  string
	primaryUsername  string
	countryCode      string
	waypointCode     string
	moveID           string
	pictureID        string
}

type liveEndpointCase struct {
	name         string
	path         string
	collection   bool
	resourceType string
}

func TestPublicLiveIntegrationCoversAllPublicEndpoints(t *testing.T) {
	client := &http.Client{Timeout: 20 * time.Second}
	fixtures := discoverLiveFixtures(t, client)

	testCases := []liveEndpointCase{
		{
			name:         "geokrety-list",
			path:         livePath("/api/v3/geokrety", url.Values{"filter[owner_id]": {fixtures.primaryUserID}, "limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "geokret-details",
			path:         "/api/v3/geokrety/" + fixtures.secondaryGeokret,
			collection:   false,
			resourceType: "geokrety",
		},
		{
			name:         "geokret-stats",
			path:         "/api/v3/geokrety/" + fixtures.primaryGeokret + "/stats",
			collection:   false,
			resourceType: "geokrety_stats",
		},
		{
			name:         "geokret-moves",
			path:         livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/moves", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "move",
		},
		{
			name:         "geokret-loved-by",
			path:         livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/loved-by", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "lover",
		},
		{
			name:         "geokret-watched-by",
			path:         livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/watched-by", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "watcher",
		},
		{
			name:         "geokret-finders",
			path:         livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/finders", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "finder",
		},
		{
			name:         "geokret-pictures",
			path:         livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/pictures", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "picture",
		},
		{
			name:         "geokret-countries",
			path:         livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/countries", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "country_visit",
		},
		{
			name:         "geokret-waypoints",
			path:         livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/waypoints", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "waypoint_visit",
		},
		{
			name:         "moves-list",
			path:         livePath("/api/v3/moves", url.Values{"filter[geokret]": {fixtures.primaryGeokret}, "limit": {"1"}}),
			collection:   true,
			resourceType: "move",
		},
		{
			name:         "move-details",
			path:         "/api/v3/moves/" + fixtures.moveID,
			collection:   false,
			resourceType: "move",
		},
		{
			name:         "countries-list",
			path:         livePath("/api/v3/countries", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "country",
		},
		{
			name:         "country-details",
			path:         "/api/v3/countries/" + fixtures.countryCode,
			collection:   false,
			resourceType: "country",
		},
		{
			name:         "country-geokrety",
			path:         livePath("/api/v3/countries/"+fixtures.countryCode+"/geokrety", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "waypoint-details",
			path:         "/api/v3/waypoints/" + fixtures.waypointCode,
			collection:   false,
			resourceType: "waypoint",
		},
		{
			name:         "waypoint-geokrety-current",
			path:         livePath("/api/v3/waypoints/"+fixtures.waypointCode+"/geokrety-current", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "waypoint-geokrety-past",
			path:         livePath("/api/v3/waypoints/"+fixtures.waypointCode+"/geokrety-past", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "users-list",
			path:         livePath("/api/v3/users", url.Values{"filter[username]": {fixtures.primaryUsername}, "limit": {"1"}}),
			collection:   true,
			resourceType: "user",
		},
		{
			name:         "user-details",
			path:         "/api/v3/users/" + fixtures.secondaryUserID,
			collection:   false,
			resourceType: "user",
		},
		{
			name:         "user-stats",
			path:         "/api/v3/users/" + fixtures.primaryUserID + "/stats",
			collection:   false,
			resourceType: "user_stats",
		},
		{
			name:         "user-geokrety-owned",
			path:         livePath("/api/v3/users/"+fixtures.primaryUserID+"/geokrety-owned", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "user-geokrety-found",
			path:         livePath("/api/v3/users/"+fixtures.primaryUserID+"/geokrety-found", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "user-geokrety-loved",
			path:         livePath("/api/v3/users/"+fixtures.primaryUserID+"/geokrety-loved", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "user-geokrety-watched",
			path:         livePath("/api/v3/users/"+fixtures.primaryUserID+"/geokrety-watched", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "geokrety",
		},
		{
			name:         "user-pictures",
			path:         livePath("/api/v3/users/"+fixtures.primaryUserID+"/pictures", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "picture",
		},
		{
			name:         "user-countries",
			path:         livePath("/api/v3/users/"+fixtures.primaryUserID+"/countries", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "country_visit",
		},
		{
			name:         "user-waypoints",
			path:         livePath("/api/v3/users/"+fixtures.primaryUserID+"/waypoints", url.Values{"limit": {"1"}}),
			collection:   true,
			resourceType: "waypoint_visit",
		},
		{
			name:         "pictures-list",
			path:         livePath("/api/v3/pictures", url.Values{"filter[geokret]": {fixtures.primaryGeokret}, "limit": {"1"}}),
			collection:   true,
			resourceType: "picture",
		},
		{
			name:         "picture-details",
			path:         "/api/v3/pictures/" + fixtures.pictureID,
			collection:   false,
			resourceType: "picture",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := liveRequestJSON(t, client, tc.path, http.StatusOK)
			if tc.collection {
				assertLiveCollectionDocument(t, payload, tc.resourceType)
				return
			}
			assertLiveSingleDocument(t, payload, tc.resourceType)
		})
	}
}

func TestPublicLiveIntegrationPreservesQueryStateInPaginationLinks(t *testing.T) {
	client := &http.Client{Timeout: 20 * time.Second}
	fixtures := discoverLiveFixtures(t, client)
	path := livePath("/api/v3/moves", url.Values{"filter[geokret]": {fixtures.primaryGeokret}, "limit": {"1"}})
	payload := liveRequestJSON(t, client, path, http.StatusOK)
	assertLiveCollectionDocument(t, payload, "move")

	links := liveMapField(t, payload, "links")
	selfQuery := liveLinkQueryValues(t, liveStringField(t, links, "self"))
	if got := selfQuery.Get("filter[geokret]"); got != fixtures.primaryGeokret {
		t.Fatalf("links.self filter[geokret] = %q, want %q", got, fixtures.primaryGeokret)
	}
	if got := selfQuery.Get("limit"); got != "1" {
		t.Fatalf("links.self limit = %q, want 1", got)
	}

	page := liveMapField(t, liveMapField(t, payload, "meta"), "page")
	hasMore, ok := page["has_more"].(bool)
	if !ok {
		t.Fatalf("meta.page.has_more missing or invalid: %#v", page["has_more"])
	}
	if !hasMore {
		return
	}
	nextQuery := liveLinkQueryValues(t, liveStringField(t, links, "next"))
	if got := nextQuery.Get("filter[geokret]"); got != fixtures.primaryGeokret {
		t.Fatalf("links.next filter[geokret] = %q, want %q", got, fixtures.primaryGeokret)
	}
	if got := nextQuery.Get("limit"); got != "1" {
		t.Fatalf("links.next limit = %q, want 1", got)
	}
	if got := nextQuery.Get("sort"); strings.TrimSpace(got) == "" {
		t.Fatalf("links.next sort missing: %#v", got)
	}
}

func TestPublicLiveIntegrationReturnsJSONRESTErrors(t *testing.T) {
	client := &http.Client{Timeout: 20 * time.Second}

	badSort := liveRequestJSON(t, client, livePath("/api/v3/geokrety", url.Values{"sort": {"definitely-invalid"}, "limit": {"1"}}), http.StatusBadRequest)
	assertLiveErrorDocument(t, badSort)

	badFilter := liveRequestJSON(t, client, livePath("/api/v3/moves", url.Values{"filter[country]": {"XYZ"}, "limit": {"1"}}), http.StatusBadRequest)
	assertLiveErrorDocument(t, badFilter)
}

func discoverLiveFixtures(t *testing.T, client *http.Client) liveFixtures {
	t.Helper()
	fixtures := liveFixtures{
		primaryGeokret:   "GKB65C",
		secondaryGeokret: "GK10288",
		primaryUserID:    "26422",
		secondaryUserID:  "1",
		primaryUsername:  "kumy",
	}

	primaryGeokret := liveRequestJSON(t, client, "/api/v3/geokrety/"+fixtures.primaryGeokret, http.StatusOK)
	secondaryGeokret := liveRequestJSON(t, client, "/api/v3/geokrety/"+fixtures.secondaryGeokret, http.StatusOK)
	liveRequestJSON(t, client, "/api/v3/users/"+fixtures.primaryUserID, http.StatusOK)
	liveRequestJSON(t, client, "/api/v3/users/"+fixtures.secondaryUserID, http.StatusOK)

	fixtures.countryCode = firstNonEmpty(
		liveRelationshipID(primaryGeokret, "country"),
		liveRelationshipID(secondaryGeokret, "country"),
	)
	fixtures.waypointCode = firstNonEmpty(
		liveRelationshipID(primaryGeokret, "waypoint"),
		liveRelationshipID(secondaryGeokret, "waypoint"),
	)
	fixtures.moveID = firstNonEmpty(
		liveRelationshipID(primaryGeokret, "last_position"),
		liveRelationshipID(secondaryGeokret, "last_position"),
		firstCollectionResourceID(liveRequestJSON(t, client, livePath("/api/v3/moves", url.Values{"filter[geokret]": {fixtures.primaryGeokret}, "limit": {"1"}}), http.StatusOK)),
	)
	fixtures.pictureID = firstNonEmpty(
		firstCollectionResourceID(liveRequestJSON(t, client, livePath("/api/v3/geokrety/"+fixtures.primaryGeokret+"/pictures", url.Values{"limit": {"1"}}), http.StatusOK)),
		firstCollectionResourceID(liveRequestJSON(t, client, livePath("/api/v3/pictures", url.Values{"limit": {"1"}}), http.StatusOK)),
	)

	if fixtures.countryCode == "" {
		t.Fatal("failed to discover a live country code from the required GeoKret fixtures")
	}
	if fixtures.waypointCode == "" {
		t.Fatal("failed to discover a live waypoint code from the required GeoKret fixtures")
	}
	if fixtures.moveID == "" {
		t.Fatal("failed to discover a live move id from the required GeoKret fixtures")
	}
	if fixtures.pictureID == "" {
		t.Fatal("failed to discover a live picture id from the live API")
	}

	return fixtures
}

func liveRequestJSON(t *testing.T, client *http.Client, path string, expectedStatus int) map[string]any {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, liveAbsoluteURL(path), nil)
	if err != nil {
		t.Fatalf("build request for %s: %v", path, err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("request %s failed: %v; start the API locally with 'cd /home/kumy/GIT/geokrety-stats/api && make run' before running make test-integration", path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body for %s: %v", path, err)
	}
	if response.StatusCode != expectedStatus {
		t.Fatalf("GET %s returned %d, want %d; body: %s", path, response.StatusCode, expectedStatus, strings.TrimSpace(string(body)))
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode JSON for %s: %v; body: %s", path, err, strings.TrimSpace(string(body)))
	}
	return payload
}

func liveAbsoluteURL(path string) string {
	trimmed := strings.TrimSpace(path)
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	return strings.TrimRight(liveBaseURL(), "/") + trimmed
}

func liveBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("GEOKRETY_STATS_API_BASE_URL")); value != "" {
		return value
	}
	return defaultLiveAPIBaseURL
}

func livePath(path string, query url.Values) string {
	if len(query) == 0 {
		return path
	}
	return path + "?" + query.Encode()
}

func assertLiveSingleDocument(t *testing.T, payload map[string]any, expectedType string) {
	t.Helper()
	assertLiveSuccessEnvelope(t, payload)
	assertLiveResourceContract(t, liveMapField(t, payload, "data"), expectedType, true)
	assertLiveIncludedResources(t, payload)
}

func assertLiveCollectionDocument(t *testing.T, payload map[string]any, expectedType string) {
	t.Helper()
	assertLiveSuccessEnvelope(t, payload)

	for index, item := range liveSliceField(t, payload, "data") {
		resource, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("data[%d] missing or invalid resource object: %#v", index, item)
		}
		assertLiveResourceContract(t, resource, expectedType, true)
	}

	meta := liveMapField(t, payload, "meta")
	page := liveMapField(t, meta, "page")
	pageType := liveStringField(t, page, "type")
	if pageType != "cursor" && pageType != "page" {
		t.Fatalf("meta.page.type = %q, want cursor or page", pageType)
	}
	if pageType == "cursor" && !liveHasNumber(page["limit"]) {
		t.Fatalf("meta.page.limit missing or invalid: %#v", page["limit"])
	}
	if pageType == "page" && !liveHasNumber(page["number"]) {
		t.Fatalf("meta.page.number missing or invalid: %#v", page["number"])
	}
	hasMore, ok := page["has_more"].(bool)
	if pageType == "cursor" && !ok {
		t.Fatalf("meta.page.has_more missing or invalid: %#v", page["has_more"])
	}

	query := liveMapField(t, meta, "query")
	if strings.TrimSpace(liveStringField(t, query, "sort")) == "" {
		t.Fatal("meta.query.sort missing")
	}

	capabilities := liveMapField(t, meta, "capabilities")
	if len(liveSliceField(t, capabilities, "sorts")) == 0 {
		t.Fatal("meta.capabilities.sorts missing")
	}

	links := liveMapField(t, payload, "links")
	if strings.TrimSpace(liveStringField(t, links, "self")) == "" {
		t.Fatal("links.self missing")
	}
	if ok && hasMore && strings.TrimSpace(liveOptionalStringField(links, "next")) == "" {
		t.Fatal("links.next missing while meta.page.has_more is true")
	}

	assertLiveIncludedResources(t, payload)
}

func assertLiveSuccessEnvelope(t *testing.T, payload map[string]any) {
	t.Helper()
	if _, ok := payload["data"]; !ok {
		t.Fatalf("success payload missing data: %#v", payload)
	}
	meta := liveMapField(t, payload, "meta")
	if !liveHasNumber(meta["execution_time_ms"]) {
		t.Fatalf("meta.execution_time_ms missing or invalid: %#v", meta["execution_time_ms"])
	}
	if _, ok := payload["links"]; !ok {
		t.Fatalf("success payload missing links: %#v", payload)
	}
}

func assertLiveErrorDocument(t *testing.T, payload map[string]any) {
	t.Helper()
	errorObject := liveMapField(t, payload, "error")
	if strings.TrimSpace(liveStringField(t, errorObject, "code")) == "" {
		t.Fatal("error.code missing")
	}
	if strings.TrimSpace(liveStringField(t, errorObject, "message")) == "" {
		t.Fatal("error.message missing")
	}
	if strings.TrimSpace(liveStringField(t, payload, "timestamp")) == "" {
		t.Fatal("timestamp missing")
	}
}

func assertLiveIncludedResources(t *testing.T, payload map[string]any) {
	t.Helper()
	includedRaw, ok := payload["included"]
	if !ok || includedRaw == nil {
		return
	}
	included, ok := includedRaw.([]any)
	if !ok {
		t.Fatalf("included missing or invalid: %#v", includedRaw)
	}
	for index, item := range included {
		resource, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("included[%d] missing or invalid resource object: %#v", index, item)
		}
		assertLiveResourceContract(t, resource, "", false)
	}
}

func assertLiveResourceContract(t *testing.T, resource map[string]any, expectedType string, requireSelfLink bool) {
	t.Helper()
	if strings.TrimSpace(liveStringField(t, resource, "id")) == "" {
		t.Fatal("resource.id missing")
	}
	resourceType := liveStringField(t, resource, "type")
	if expectedType != "" && resourceType != expectedType {
		t.Fatalf("resource.type = %q, want %q", resourceType, expectedType)
	}
	if attributes, ok := resource["attributes"]; ok && attributes != nil {
		if _, ok := attributes.(map[string]any); !ok {
			t.Fatalf("resource.attributes invalid: %#v", attributes)
		}
	}
	if relationships, ok := resource["relationships"]; ok && relationships != nil {
		relationshipMap, ok := relationships.(map[string]any)
		if !ok {
			t.Fatalf("resource.relationships invalid: %#v", relationships)
		}
		for name, raw := range relationshipMap {
			relationship, ok := raw.(map[string]any)
			if !ok {
				t.Fatalf("relationship %q invalid: %#v", name, raw)
			}
			if data, ok := relationship["data"]; ok {
				assertLiveRelationshipData(t, name, data)
			}
			if links, ok := relationship["links"]; ok && links != nil {
				linkMap, ok := links.(map[string]any)
				if !ok {
					t.Fatalf("relationship %q links invalid: %#v", name, links)
				}
				if related, ok := linkMap["related"]; ok && related != nil {
					if strings.TrimSpace(fmt.Sprint(related)) == "" {
						t.Fatalf("relationship %q links.related missing", name)
					}
				}
			}
		}
	}
	if links, ok := resource["links"]; ok && links != nil {
		linkMap, ok := links.(map[string]any)
		if !ok {
			t.Fatalf("resource.links invalid: %#v", links)
		}
		if requireSelfLink || strings.TrimSpace(liveOptionalStringField(linkMap, "self")) != "" {
			if strings.TrimSpace(liveStringField(t, linkMap, "self")) == "" {
				t.Fatal("resource.links.self missing")
			}
		}
	} else if requireSelfLink {
		t.Fatal("resource.links.self missing")
	}
}

func assertLiveRelationshipData(t *testing.T, name string, data any) {
	t.Helper()
	switch typed := data.(type) {
	case nil:
		return
	case map[string]any:
		if len(typed) == 0 {
			return
		}
		if strings.TrimSpace(liveStringField(t, typed, "id")) == "" {
			t.Fatalf("relationship %q data.id missing", name)
		}
		if strings.TrimSpace(liveStringField(t, typed, "type")) == "" {
			t.Fatalf("relationship %q data.type missing", name)
		}
		for key := range typed {
			if key != "id" && key != "type" {
				t.Fatalf("relationship %q data contains forbidden field %q", name, key)
			}
		}
	case []any:
		for _, item := range typed {
			resource, ok := item.(map[string]any)
			if !ok {
				t.Fatalf("relationship %q data item invalid: %#v", name, item)
			}
			assertLiveRelationshipData(t, name, resource)
		}
	default:
		t.Fatalf("relationship %q data invalid: %#v", name, data)
	}
}

func liveMapField(t *testing.T, object map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := object[key]
	if !ok || value == nil {
		t.Fatalf("field %q missing or nil in %#v", key, object)
	}
	mapped, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("field %q invalid: %#v", key, value)
	}
	return mapped
}

func liveSliceField(t *testing.T, object map[string]any, key string) []any {
	t.Helper()
	value, ok := object[key]
	if !ok || value == nil {
		t.Fatalf("field %q missing or nil in %#v", key, object)
	}
	slice, ok := value.([]any)
	if !ok {
		t.Fatalf("field %q invalid: %#v", key, value)
	}
	return slice
}

func liveStringField(t *testing.T, object map[string]any, key string) string {
	t.Helper()
	value, ok := object[key]
	if !ok || value == nil {
		t.Fatalf("field %q missing or nil in %#v", key, object)
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("field %q invalid: %#v", key, value)
	}
	return text
}

func liveOptionalStringField(object map[string]any, key string) string {
	value, ok := object[key]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func liveHasNumber(value any) bool {
	switch value.(type) {
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func liveLinkQueryValues(t *testing.T, link string) url.Values {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse link %q: %v", link, err)
	}
	return parsed.Query()
}

func liveRelationshipID(payload map[string]any, name string) string {
	data, ok := payload["data"].(map[string]any)
	if !ok {
		return ""
	}
	relationships, ok := data["relationships"].(map[string]any)
	if !ok {
		return ""
	}
	relationship, ok := relationships[name].(map[string]any)
	if !ok {
		return ""
	}
	identifier, ok := relationship["data"].(map[string]any)
	if !ok {
		return ""
	}
	id, _ := identifier["id"].(string)
	return strings.TrimSpace(id)
}

func firstCollectionResourceID(payload map[string]any) string {
	data, ok := payload["data"].([]any)
	if !ok || len(data) == 0 {
		return ""
	}
	resource, ok := data[0].(map[string]any)
	if !ok {
		return ""
	}
	id, _ := resource["id"].(string)
	return strings.TrimSpace(id)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
