package handlers

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/pagination"
)

type ResponseMeta struct {
	RequestedAt string                 `json:"requestedAt" xml:"requestedAt"`
	QueryMs     int64                  `json:"queryMs" xml:"queryMs"`
	Pagination  *pagination.OffsetInfo `json:"pagination,omitempty" xml:"pagination,omitempty"`
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
	req pagination.OffsetRequest,
	totalItems *int,
	hasMore *bool,
	overrideReturned *int,
) {
	returned, _ := payloadCount(payload)
	if overrideReturned != nil {
		returned = *overrideReturned
	}
	envelope := Envelope{
		Data: payload,
		Meta: ResponseMeta{
			RequestedAt: time.Now().UTC().Format(time.RFC3339),
			QueryMs:     time.Since(started).Milliseconds(),
			Pagination:  offsetPaginationInfo(req, returned, totalItems, hasMore),
		},
	}
	if wantsXML(r) {
		writeXML(w, status, envelope)
		return
	}
	writeJSON(w, status, envelope)
}

func offsetPaginationInfo(req pagination.OffsetRequest, returned int, totalItems *int, hasMore *bool) *pagination.OffsetInfo {
	info := pagination.NewOffsetInfo(req.Offset, req.Limit, returned, totalItems, hasMore)
	return &info
}

func queryPagination(r *http.Request, defaultLimit, maxLimit int) (pagination.OffsetRequest, error) {
	return pagination.ParseOffsetRequest(r, pagination.RequestConfig{
		LimitParam:    "limit",
		OffsetParam:   "offset",
		CursorParam:   "cursor",
		DefaultLimit:  defaultLimit,
		MinLimit:      1,
		MaxLimit:      maxLimit,
		DefaultOffset: 0,
		MinOffset:     0,
		MaxOffset:     1_000_000,
	})
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

func paginationErrorMessage(err error) string {
	switch err {
	case pagination.ErrInvalidCursor:
		return "invalid cursor"
	case pagination.ErrUnsupportedCursorVersion:
		return "unsupported cursor version"
	case pagination.ErrInvalidLimit:
		return "invalid limit"
	case pagination.ErrInvalidOffset:
		return "invalid offset"
	case pagination.ErrAmbiguousPagination:
		return "cannot combine cursor with offset"
	default:
		return "invalid pagination parameters"
	}
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
	payload := ErrorResponse{
		Error:     message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if wantsXML(r) {
		writeXML(w, status, payload)
		return
	}
	writeJSON(w, status, payload)
}
