package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type PaginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type ResponseMeta struct {
	RequestedAt string         `json:"requestedAt"`
	QueryMs     int64          `json:"queryMs"`
	Pagination  PaginationMeta `json:"pagination"`
}

type Envelope struct {
	Data interface{}  `json:"data"`
	Meta ResponseMeta `json:"meta"`
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
	writeJSON(w, status, Envelope{
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
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
