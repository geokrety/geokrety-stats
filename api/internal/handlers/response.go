package handlers

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	sharedjsonrest "github.com/geokrety/geokrety-stats/geokrety/jsonrest"
)

type ResponseMeta struct {
	RequestedAt string `json:"requestedAt" xml:"requestedAt"`
	QueryMs     int64  `json:"queryMs" xml:"queryMs"`
}

type Envelope struct {
	XMLName xml.Name     `json:"-" xml:"response"`
	Data    interface{}  `json:"data" xml:"data"`
	Meta    ResponseMeta `json:"meta" xml:"meta"`
}

type ErrorResponse struct {
	XMLName   xml.Name `json:"-" xml:"errorResponse"`
	Error     string   `json:"error" xml:"error"`
	Timestamp string   `json:"timestamp" xml:"timestamp"`
}

type presentedPayload struct {
	Data     any
	Included []sharedjsonrest.Resource
}

func queryInt(r *http.Request, key string, fallback, minValue, maxValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	if parsed < minValue {
		return minValue
	}
	if parsed > maxValue {
		return maxValue
	}
	return parsed
}

func writeEnvelope(w http.ResponseWriter, status int, payload interface{}, started time.Time, limit, offset, count int) {
	writeEnvelopeForRequest(w, nil, status, payload, started, limit, offset, count)
}

func writeEnvelopeForRequest(w http.ResponseWriter, r *http.Request, status int, payload interface{}, started time.Time, limit, offset, count int) {
	if !wantsXML(r) {
		request := r
		if isCollectionPayload(payload) {
			request = normalizedCollectionRequest(r)
		}
		presentation := presentationFromPayload(r, payload)
		links := sharedjsonrest.Links{}
		links.Set("self", sharedjsonrest.SelfLink(request))
		meta := sharedjsonrest.NewMeta(started)
		if isCollectionPayload(payload) {
			filters, sort := currentQueryContext(request)
			meta = meta.WithFilters(filters).WithSort(sort)
			availableFilters, availableSorts := collectionCapabilities(request)
			meta = meta.WithCapabilities(availableFilters, availableSorts)
		}
		writeJSON(w, status, sharedjsonrest.NewDocument(presentation.Data, presentation.Included, meta, links))
		return
	}
	envelope := Envelope{
		Data: payload,
		Meta: ResponseMeta{
			RequestedAt: time.Now().UTC().Format(time.RFC3339),
			QueryMs:     time.Since(started).Milliseconds(),
		},
	}
	if wantsXML(r) {
		writeXML(w, status, envelope)
		return
	}
	writeJSON(w, status, envelope)
}

func writeEnvelopeForOffsetRequest(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	payload interface{},
	started time.Time,
	req sharedjsonrest.CursorRequest,
	totalItems *int,
	hasMore *bool,
	overrideReturned *int,
) {
	returned, _ := payloadCount(payload)
	if overrideReturned != nil {
		returned = *overrideReturned
	}
	if !wantsXML(r) {
		normalizedRequest := normalizedCollectionRequest(r)
		hasMoreValue := hasMore != nil && *hasMore
		var nextCursor *sharedjsonrest.Cursor
		fingerprint := sharedjsonrest.CursorFingerprint(normalizedRequest, "cursor")
		if hasMoreValue {
			cursor := sharedjsonrest.EncodeCursor(sharedjsonrest.CurrentCursorVersion, int64(req.Offset+returned), fingerprint)
			nextCursor = &cursor
		}
		meta := sharedjsonrest.NewMeta(started).WithCursor(req, hasMoreValue)
		filters, sort := currentQueryContext(normalizedRequest)
		meta = meta.WithFilters(filters).WithSort(sort)
		availableFilters, availableSorts := collectionCapabilities(normalizedRequest)
		meta = meta.WithCapabilities(availableFilters, availableSorts)
		links := sharedjsonrest.CursorLinks(normalizedRequest, req, nextCursor, "limit", "cursor")
		presentation := presentationFromPayload(r, payload)
		writeJSON(w, status, sharedjsonrest.NewDocument(presentation.Data, presentation.Included, meta, links))
		return
	}
	envelope := Envelope{
		Data: payload,
		Meta: ResponseMeta{
			RequestedAt: time.Now().UTC().Format(time.RFC3339),
			QueryMs:     time.Since(started).Milliseconds(),
		},
	}
	writeXML(w, status, envelope)
}

func queryPagination(r *http.Request, defaultLimit, maxLimit int) (sharedjsonrest.CursorRequest, error) {
	normalizedRequest := normalizedCollectionRequest(r)
	return sharedjsonrest.ParseCursorRequest(normalizedRequest, sharedjsonrest.CursorConfig{
		LimitParam:   "limit",
		CursorParam:  "cursor",
		DefaultLimit: defaultLimit,
		MinLimit:     1,
		MaxLimit:     maxLimit,
		Forbidden:    []string{"offset", "page", "per_page"},
		Fingerprint:  sharedjsonrest.CursorFingerprint(normalizedRequest, "cursor"),
	})
}

func writeRawEnvelopeForOffsetRequest(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	payload interface{},
	started time.Time,
	req sharedjsonrest.CursorRequest,
	hasMore bool,
) {
	if wantsXML(r) {
		envelope := Envelope{
			Data: payload,
			Meta: ResponseMeta{
				RequestedAt: time.Now().UTC().Format(time.RFC3339),
				QueryMs:     time.Since(started).Milliseconds(),
			},
		}
		writeXML(w, status, envelope)
		return
	}
	normalizedRequest := normalizedCollectionRequest(r)
	var nextCursor *sharedjsonrest.Cursor
	fingerprint := sharedjsonrest.CursorFingerprint(normalizedRequest, "cursor")
	if hasMore {
		cursor := sharedjsonrest.EncodeCursor(sharedjsonrest.CurrentCursorVersion, int64(req.Offset+req.Limit), fingerprint)
		nextCursor = &cursor
	}
	meta := sharedjsonrest.NewMeta(started).WithCursor(req, hasMore)
	filters, sort := currentQueryContext(normalizedRequest)
	meta = meta.WithFilters(filters).WithSort(sort)
	availableFilters, availableSorts := collectionCapabilities(normalizedRequest)
	meta = meta.WithCapabilities(availableFilters, availableSorts)
	links := sharedjsonrest.CursorLinks(normalizedRequest, req, nextCursor, "limit", "cursor")
	presentation := presentationFromPayload(r, payload)
	writeJSON(w, status, sharedjsonrest.NewDocument(presentation.Data, presentation.Included, meta, links))
}

func presentationFromPayload(r *http.Request, payload interface{}) presentedPayload {
	presented := resourceDataFromPayload(r, payload)
	if normalized, ok := presented.(presentedPayload); ok {
		return normalized
	}
	return presentedPayload{Data: presented}
}

func currentQueryContext(r *http.Request) (map[string]any, string) {
	if r == nil || r.URL == nil {
		return nil, ""
	}
	filters := map[string]any{}
	sort := ""
	for key, values := range r.URL.Query() {
		if key == "limit" || key == "cursor" || key == "ids" || key == "page[number]" || key == "page[size]" {
			continue
		}
		if key == "sort" {
			if len(values) > 0 {
				sort = strings.TrimSpace(values[len(values)-1])
			}
			continue
		}
		normalizedKey := key
		if strings.HasPrefix(key, "filter[") && strings.HasSuffix(key, "]") {
			normalizedKey = strings.TrimSuffix(strings.TrimPrefix(key, "filter["), "]")
		}
		if normalizedKey == "" {
			continue
		}
		filters[normalizedKey] = normalizeQueryValues(values)
	}
	if len(filters) == 0 {
		filters = nil
	}
	if sort == "" {
		sort = defaultCollectionSort(r)
	}
	return filters, sort
}

func defaultCollectionSort(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	path := r.URL.Path
	switch {
	case path == "/api/v3/geokrety":
		return "-last_move_at"
	case path == "/api/v3/users":
		return "-joined_at"
	case path == "/api/v3/countries":
		return "code"
	case path == "/api/v3/moves" || strings.HasSuffix(path, "/moves"):
		return "-date"
	case path == "/api/v3/pictures" || strings.HasSuffix(path, "/pictures"):
		return "-created_on"
	case strings.HasSuffix(path, "/loved-by"):
		return "-loved_on_date"
	case strings.HasSuffix(path, "/watched-by"):
		return "-watched_on_date"
	case strings.HasSuffix(path, "/finders"):
		return "-found_on_date"
	case strings.HasSuffix(path, "/countries") && strings.Contains(path, "/geokrety/"):
		return "-first_visited_at"
	case strings.HasSuffix(path, "/countries") && strings.Contains(path, "/users/"):
		return "-last_visit"
	case strings.HasSuffix(path, "/waypoints"):
		return "-last_visited_at"
	case strings.HasSuffix(path, "/geokrety") || strings.HasSuffix(path, "/geokrety-owned") || strings.HasSuffix(path, "/geokrety-found") || strings.HasSuffix(path, "/geokrety-loved") || strings.HasSuffix(path, "/geokrety-watched") || strings.HasSuffix(path, "/geokrety-current") || strings.HasSuffix(path, "/geokrety-past"):
		return "-last_move_at"
	default:
		return ""
	}
}

func normalizeQueryValues(values []string) any {
	if len(values) == 0 {
		return ""
	}
	if len(values) > 1 {
		items := make([]string, 0, len(values))
		for _, value := range values {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			items = append(items, trimmed)
		}
		return items
	}
	value := strings.TrimSpace(values[0])
	if strings.Contains(value, ",") {
		parts := strings.Split(value, ",")
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			items = append(items, trimmed)
		}
		if len(items) > 1 {
			return items
		}
	}
	return value
}

func collectionCapabilities(r *http.Request) (map[string]any, []string) {
	if r == nil || r.URL == nil {
		return nil, nil
	}
	path := r.URL.Path
	switch {
	case path == "/api/v3/geokrety":
		return map[string]any{
			"name":     map[string]any{"type": "string", "operators": []string{"contains"}},
			"owner_id": map[string]any{"type": "integer", "operators": []string{"eq"}},
			"country":  map[string]any{"type": "string", "multi": true, "operators": []string{"eq", "in"}},
		}, []string{"name", "-name", "born_at", "-born_at", "last_move_at", "-last_move_at"}
	case strings.HasSuffix(path, "/geokrety") || strings.HasSuffix(path, "/geokrety-owned") || strings.HasSuffix(path, "/geokrety-found") || strings.HasSuffix(path, "/geokrety-loved") || strings.HasSuffix(path, "/geokrety-watched") || strings.HasSuffix(path, "/geokrety-current") || strings.HasSuffix(path, "/geokrety-past"):
		return nil, []string{"name", "-name", "born_at", "-born_at", "last_move_at", "-last_move_at"}
	case path == "/api/v3/users":
		return map[string]any{
			"username": map[string]any{"type": "string", "operators": []string{"contains"}},
			"country":  map[string]any{"type": "string", "multi": true, "operators": []string{"eq", "in"}},
		}, []string{"username", "-username", "joined_at", "-joined_at", "last_move_at", "-last_move_at"}
	case path == "/api/v3/moves":
		return map[string]any{
			"geokret":   map[string]any{"type": "string", "operators": []string{"eq", "in"}},
			"user":      map[string]any{"type": "integer", "operators": []string{"eq", "in"}},
			"country":   map[string]any{"type": "string", "operators": []string{"eq"}},
			"waypoint":  map[string]any{"type": "string", "operators": []string{"eq", "in"}},
			"date_from": map[string]any{"type": "date", "operators": []string{"gte"}},
			"date_to":   map[string]any{"type": "date", "operators": []string{"lte"}},
		}, []string{"date", "-date", "id", "-id"}
	case strings.HasSuffix(path, "/moves"):
		return map[string]any{
			"user":      map[string]any{"type": "integer", "operators": []string{"eq", "in"}},
			"country":   map[string]any{"type": "string", "operators": []string{"eq"}},
			"waypoint":  map[string]any{"type": "string", "operators": []string{"eq", "in"}},
			"date_from": map[string]any{"type": "date", "operators": []string{"gte"}},
			"date_to":   map[string]any{"type": "date", "operators": []string{"lte"}},
		}, []string{"date", "-date", "id", "-id"}
	case path == "/api/v3/pictures":
		return map[string]any{
			"geokret": map[string]any{"type": "string", "operators": []string{"eq", "in"}},
			"move":    map[string]any{"type": "integer", "operators": []string{"eq", "in"}},
			"user":    map[string]any{"type": "integer", "operators": []string{"eq", "in"}},
		}, []string{"created_on", "-created_on", "id", "-id"}
	case strings.HasSuffix(path, "/pictures"):
		return map[string]any{
			"move": map[string]any{"type": "integer", "operators": []string{"eq", "in"}},
			"user": map[string]any{"type": "integer", "operators": []string{"eq", "in"}},
		}, []string{"created_on", "-created_on", "id", "-id"}
	case path == "/api/v3/countries":
		return nil, []string{"code", "-code", "name", "-name"}
	case strings.HasSuffix(path, "/loved-by"):
		return nil, []string{"loved_on_date", "-loved_on_date"}
	case strings.HasSuffix(path, "/watched-by"):
		return nil, []string{"watched_on_date", "-watched_on_date"}
	case strings.HasSuffix(path, "/finders"):
		return nil, []string{"found_on_date", "-found_on_date"}
	case strings.HasSuffix(path, "/countries") && strings.Contains(path, "/geokrety/"):
		return nil, []string{"country_code", "-country_code", "first_visited_at", "-first_visited_at", "move_count", "-move_count"}
	case strings.HasSuffix(path, "/countries") && strings.Contains(path, "/users/"):
		return nil, []string{"country_code", "-country_code", "first_visit", "-first_visit", "last_visit", "-last_visit", "move_count", "-move_count"}
	case strings.HasSuffix(path, "/waypoints"):
		return nil, []string{"waypoint_code", "-waypoint_code", "first_visited_at", "-first_visited_at", "last_visited_at", "-last_visited_at", "visit_count", "-visit_count"}
	default:
		return nil, nil
	}
}

func isCollectionPayload(payload interface{}) bool {
	if payload == nil {
		return false
	}
	value := reflect.ValueOf(payload)
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return false
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return false
	}
	return value.Kind() == reflect.Slice || value.Kind() == reflect.Array
}

func normalizedCollectionRequest(r *http.Request) *http.Request {
	if r == nil || r.URL == nil {
		return r
	}
	if strings.TrimSpace(r.URL.Query().Get("sort")) != "" {
		return r
	}
	sort := defaultCollectionSort(r)
	if sort == "" {
		return r
	}
	clone := r.Clone(r.Context())
	clonedURL := *r.URL
	clone.URL = &clonedURL
	values := clone.URL.Query()
	values.Set("sort", sort)
	clone.URL.RawQuery = values.Encode()
	return clone
}

func payloadCount(payload interface{}) (int, bool) {
	if payload == nil {
		return 0, false
	}
	value := reflect.ValueOf(payload)
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return 0, false
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return 0, false
	}
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		return value.Len(), true
	}
	return 1, false
}

func trimPaginatedPayload(payload interface{}, limit int) (interface{}, int, bool) {
	value := reflect.ValueOf(payload)
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return payload, 0, false
		}
		value = value.Elem()
	}
	if !value.IsValid() || (value.Kind() != reflect.Slice && value.Kind() != reflect.Array) {
		count, _ := payloadCount(payload)
		return payload, count, false
	}
	returned := value.Len()
	if returned <= limit {
		return payload, returned, false
	}
	return value.Slice(0, limit).Interface(), limit, true
}

func writePaginationErrorForRequest(w http.ResponseWriter, r *http.Request, status int, err error) {
	requestErr, ok := err.(*sharedjsonrest.RequestError)
	if !ok {
		requestErr = &sharedjsonrest.RequestError{Code: "INVALID_PAGINATION_MODE", Message: "invalid pagination parameters"}
	}
	if wantsXML(r) {
		writeErrorForRequest(w, r, status, requestErr.Message)
		return
	}
	writeJSON(w, status, sharedjsonrest.NewErrorDocument(requestErr.Code, requestErr.Message, time.Now()))
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeXML(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(payload)
}

func wantsXML(r *http.Request) bool {
	if r == nil {
		return false
	}
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeErrorForRequest(w, nil, status, message)
}

func writeErrorForRequest(w http.ResponseWriter, r *http.Request, status int, message string) {
	if wantsXML(r) {
		payload := ErrorResponse{
			Error:     message,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		writeXML(w, status, payload)
		return
	}
	writeJSON(w, status, sharedjsonrest.NewErrorDocument(errorCodeForStatus(status), message, time.Now()))
}

func errorCodeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusNotFound:
		return "NOT_FOUND"
	default:
		return "INTERNAL_ERROR"
	}
}
