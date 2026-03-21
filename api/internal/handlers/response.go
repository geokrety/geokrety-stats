package handlers

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginationMeta struct {
	Limit  int `json:"limit" xml:"limit"`
	Offset int `json:"offset" xml:"offset"`
	Count  int `json:"count" xml:"count"`
}

type ResponseMeta struct {
	RequestedAt string         `json:"requestedAt" xml:"requestedAt"`
	QueryMs     int64          `json:"queryMs" xml:"queryMs"`
	Pagination  PaginationMeta `json:"pagination" xml:"pagination"`
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
			Pagination: PaginationMeta{
				Limit:  limit,
				Offset: offset,
				Count:  count,
			},
		},
	}
	if wantsXML(r) {
		writeXML(w, status, envelope)
		return
	}
	writeJSON(w, status, envelope)
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
