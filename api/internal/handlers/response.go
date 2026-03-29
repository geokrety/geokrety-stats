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
		data := resourceDataFromPayload(payload)
		links := sharedjsonrest.Links{}
		links.Set("self", sharedjsonrest.SelfLink(r))
		writeJSON(w, status, sharedjsonrest.NewDocument(data, sharedjsonrest.NewMeta(started), links))
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
		hasMoreValue := hasMore != nil && *hasMore
		var nextCursor *sharedjsonrest.Cursor
		fingerprint := sharedjsonrest.CursorFingerprint(r, "cursor")
		if hasMoreValue {
			cursor := sharedjsonrest.EncodeCursor(sharedjsonrest.CurrentCursorVersion, int64(req.Offset+returned), fingerprint)
			nextCursor = &cursor
		}
		meta := sharedjsonrest.NewMeta(started).WithCursor(req, hasMoreValue)
		links := sharedjsonrest.CursorLinks(r, req, nextCursor, "limit", "cursor")
		writeJSON(w, status, sharedjsonrest.NewDocument(resourceDataFromPayload(payload), meta, links))
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
	return sharedjsonrest.ParseCursorRequest(r, sharedjsonrest.CursorConfig{
		LimitParam:   "limit",
		CursorParam:  "cursor",
		DefaultLimit: defaultLimit,
		MinLimit:     1,
		MaxLimit:     maxLimit,
		Forbidden:    []string{"offset", "page", "per_page"},
		Fingerprint:  sharedjsonrest.CursorFingerprint(r, "cursor"),
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
	var nextCursor *sharedjsonrest.Cursor
	fingerprint := sharedjsonrest.CursorFingerprint(r, "cursor")
	if hasMore {
		cursor := sharedjsonrest.EncodeCursor(sharedjsonrest.CurrentCursorVersion, int64(req.Offset+req.Limit), fingerprint)
		nextCursor = &cursor
	}
	meta := sharedjsonrest.NewMeta(started).WithCursor(req, hasMore)
	links := sharedjsonrest.CursorLinks(r, req, nextCursor, "limit", "cursor")
	writeJSON(w, status, sharedjsonrest.NewDocument(payload, meta, links))
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
