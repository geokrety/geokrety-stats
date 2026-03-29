package jsonrest

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type PageConfig struct {
	PageParam      string
	PerPageParam   string
	DefaultPage    int
	DefaultPerPage int
	MinPerPage     int
	MaxPerPage     int
	Forbidden      []string
}

type CursorConfig struct {
	LimitParam   string
	CursorParam  string
	DefaultLimit int
	MinLimit     int
	MaxLimit     int
	Forbidden    []string
	Fingerprint  string
}

type PageRequest struct {
	Page    int
	PerPage int
}

type CursorRequest struct {
	Limit      int
	Cursor     Cursor
	UsedCursor bool
	Offset     int
}

func (r PageRequest) Offset() int {
	return (r.Page - 1) * r.PerPage
}

func ParsePageRequest(r *http.Request, cfg PageConfig) (PageRequest, error) {
	if usesAnyParam(r, cfg.Forbidden...) {
		return PageRequest{}, ErrInvalidPaginationMode
	}
	pageValue := strings.TrimSpace(r.URL.Query().Get(cfg.PageParam))
	page := cfg.DefaultPage
	if pageValue != "" {
		parsed, err := strconv.Atoi(pageValue)
		if err != nil || parsed < 1 {
			return PageRequest{}, ErrInvalidPage
		}
		page = parsed
	}
	perPage := parseLimitLikeValue(r, cfg.PerPageParam, cfg.DefaultPerPage, cfg.MinPerPage, cfg.MaxPerPage)
	if perPage == 0 {
		return PageRequest{}, ErrLimitExceeded
	}
	return PageRequest{Page: page, PerPage: perPage}, nil
}

func ParseCursorRequest(r *http.Request, cfg CursorConfig) (CursorRequest, error) {
	if usesAnyParam(r, cfg.Forbidden...) {
		return CursorRequest{}, ErrInvalidPaginationMode
	}
	limit := parseLimitLikeValue(r, cfg.LimitParam, cfg.DefaultLimit, cfg.MinLimit, cfg.MaxLimit)
	if limit == 0 {
		return CursorRequest{}, ErrLimitExceeded
	}
	request := CursorRequest{Limit: limit}
	cursorValue := strings.TrimSpace(r.URL.Query().Get(cfg.CursorParam))
	if cursorValue == "" {
		return request, nil
	}
	payload, err := Cursor(cursorValue).Decode()
	if err != nil {
		return CursorRequest{}, err
	}
	if payload.Fingerprint != "" && cfg.Fingerprint != "" && payload.Fingerprint != cfg.Fingerprint {
		return CursorRequest{}, ErrInvalidCursor
	}
	request.Cursor = Cursor(cursorValue)
	request.UsedCursor = true
	request.Offset = int(payload.Offset)
	return request, nil
}

func CursorFingerprint(r *http.Request, ignored ...string) string {
	if r == nil || r.URL == nil {
		return ""
	}
	ignoredSet := make(map[string]struct{}, len(ignored))
	for _, key := range ignored {
		ignoredSet[key] = struct{}{}
	}
	values := url.Values{}
	for key, entries := range r.URL.Query() {
		if _, skip := ignoredSet[key]; skip {
			continue
		}
		for _, entry := range entries {
			values.Add(key, entry)
		}
	}
	if encoded := values.Encode(); encoded != "" {
		return r.URL.Path + "?" + encoded
	}
	return r.URL.Path
}

func parseLimitLikeValue(r *http.Request, key string, fallback, minValue, maxValue int) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minValue {
		return fallback
	}
	if parsed > maxValue {
		return 0
	}
	return parsed
}

func usesAnyParam(r *http.Request, keys ...string) bool {
	for _, key := range keys {
		if key == "" {
			continue
		}
		if _, ok := r.URL.Query()[key]; ok {
			return true
		}
	}
	return false
}
