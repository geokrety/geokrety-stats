package jsonrest

import (
	"net/http/httptest"
	"testing"
)

func TestParseCursorRequestDefaultsLimitAndCursorOffset(t *testing.T) {
	cursor := EncodeCursor(CurrentCursorVersion, 25)
	r := httptest.NewRequest("GET", "/x?limit=7&cursor="+cursor.String(), nil)
	req, err := ParseCursorRequest(r, CursorConfig{
		LimitParam:   "limit",
		CursorParam:  "cursor",
		DefaultLimit: 20,
		MinLimit:     1,
		MaxLimit:     100,
		Forbidden:    []string{"page", "per_page"},
		Fingerprint:  CursorFingerprint(r, "cursor"),
	})
	if err != nil {
		t.Fatalf("ParseCursorRequest() error = %v", err)
	}
	if req.Limit != 7 {
		t.Fatalf("Limit = %d, want 7", req.Limit)
	}
	if !req.UsedCursor {
		t.Fatalf("UsedCursor = false, want true")
	}
	if req.Offset != 25 {
		t.Fatalf("Offset = %d, want 25", req.Offset)
	}
}

func TestParseCursorRequestRejectsFingerprintMismatch(t *testing.T) {
	cursor := EncodeCursor(CurrentCursorVersion, 10, "/x?q=one")
	r := httptest.NewRequest("GET", "/x?q=two&cursor="+cursor.String(), nil)
	_, err := ParseCursorRequest(r, CursorConfig{
		LimitParam:   "limit",
		CursorParam:  "cursor",
		DefaultLimit: 20,
		MinLimit:     1,
		MaxLimit:     100,
		Fingerprint:  CursorFingerprint(r, "cursor"),
	})
	if !IsRequestError(err, ErrInvalidCursor) {
		t.Fatalf("ParseCursorRequest() error = %v, want %v", err, ErrInvalidCursor)
	}
}

func TestParseCursorRequestRejectsMixedMode(t *testing.T) {
	r := httptest.NewRequest("GET", "/x?page=2", nil)
	_, err := ParseCursorRequest(r, CursorConfig{
		LimitParam:   "limit",
		CursorParam:  "cursor",
		DefaultLimit: 20,
		MinLimit:     1,
		MaxLimit:     100,
		Forbidden:    []string{"page", "per_page"},
	})
	if !IsRequestError(err, ErrInvalidPaginationMode) {
		t.Fatalf("ParseCursorRequest() error = %v, want %v", err, ErrInvalidPaginationMode)
	}
}

func TestParseCursorRequestRejectsLimitExceeded(t *testing.T) {
	r := httptest.NewRequest("GET", "/x?limit=101", nil)
	_, err := ParseCursorRequest(r, CursorConfig{
		LimitParam:   "limit",
		CursorParam:  "cursor",
		DefaultLimit: 20,
		MinLimit:     1,
		MaxLimit:     100,
	})
	if !IsRequestError(err, ErrLimitExceeded) {
		t.Fatalf("ParseCursorRequest() error = %v, want %v", err, ErrLimitExceeded)
	}
}

func TestParsePageRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/x?page=3&per_page=15", nil)
	req, err := ParsePageRequest(r, PageConfig{
		PageParam:      "page",
		PerPageParam:   "per_page",
		DefaultPage:    1,
		DefaultPerPage: 20,
		MinPerPage:     1,
		MaxPerPage:     100,
		Forbidden:      []string{"limit", "cursor"},
	})
	if err != nil {
		t.Fatalf("ParsePageRequest() error = %v", err)
	}
	if req.Page != 3 {
		t.Fatalf("Page = %d, want 3", req.Page)
	}
	if req.PerPage != 15 {
		t.Fatalf("PerPage = %d, want 15", req.PerPage)
	}
	if req.Offset() != 30 {
		t.Fatalf("Offset() = %d, want 30", req.Offset())
	}
}

func TestParsePageRequestRejectsInvalidPageAndMixedMode(t *testing.T) {
	r := httptest.NewRequest("GET", "/x?page=0", nil)
	_, err := ParsePageRequest(r, PageConfig{
		PageParam:      "page",
		PerPageParam:   "per_page",
		DefaultPage:    1,
		DefaultPerPage: 20,
		MinPerPage:     1,
		MaxPerPage:     100,
		Forbidden:      []string{"limit", "cursor"},
	})
	if !IsRequestError(err, ErrInvalidPage) {
		t.Fatalf("ParsePageRequest() error = %v, want %v", err, ErrInvalidPage)
	}

	r = httptest.NewRequest("GET", "/x?cursor=abc", nil)
	_, err = ParsePageRequest(r, PageConfig{
		PageParam:      "page",
		PerPageParam:   "per_page",
		DefaultPage:    1,
		DefaultPerPage: 20,
		MinPerPage:     1,
		MaxPerPage:     100,
		Forbidden:      []string{"limit", "cursor"},
	})
	if !IsRequestError(err, ErrInvalidPaginationMode) {
		t.Fatalf("ParsePageRequest() error = %v, want %v", err, ErrInvalidPaginationMode)
	}
}

func TestParsePageRequestDefaultsPerPageAndRejectsLimitExceeded(t *testing.T) {
	r := httptest.NewRequest("GET", "/x?per_page=0", nil)
	req, err := ParsePageRequest(r, PageConfig{
		PageParam:      "page",
		PerPageParam:   "per_page",
		DefaultPage:    1,
		DefaultPerPage: 20,
		MinPerPage:     1,
		MaxPerPage:     100,
	})
	if err != nil {
		t.Fatalf("ParsePageRequest() error = %v", err)
	}
	if req.PerPage != 20 {
		t.Fatalf("PerPage = %d, want 20", req.PerPage)
	}

	r = httptest.NewRequest("GET", "/x?per_page=101", nil)
	_, err = ParsePageRequest(r, PageConfig{
		PageParam:      "page",
		PerPageParam:   "per_page",
		DefaultPage:    1,
		DefaultPerPage: 20,
		MinPerPage:     1,
		MaxPerPage:     100,
	})
	if !IsRequestError(err, ErrLimitExceeded) {
		t.Fatalf("ParsePageRequest() error = %v, want %v", err, ErrLimitExceeded)
	}
}
