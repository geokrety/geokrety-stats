package pagination

import (
	"net/http/httptest"
	"testing"
)

func TestCursorRoundTrip(t *testing.T) {
	cursor := EncodeCursor(CurrentCursorVersion, 40)
	version, offset, err := cursor.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if version != CurrentCursorVersion {
		t.Fatalf("version = %d, want %d", version, CurrentCursorVersion)
	}
	if offset != 40 {
		t.Fatalf("offset = %d, want 40", offset)
	}
}

func TestCursorDecodeErrors(t *testing.T) {
	tests := []struct {
		name   string
		cursor Cursor
		want   error
	}{
		{name: "empty", cursor: "", want: ErrInvalidCursor},
		{name: "garbage", cursor: "%%%", want: ErrInvalidCursor},
		{name: "unsupported-version", cursor: EncodeCursor(CurrentCursorVersion+1, 10), want: ErrUnsupportedCursorVersion},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := tc.cursor.Decode()
			if err != tc.want {
				t.Fatalf("Decode() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestParseOffsetRequestUsesCursor(t *testing.T) {
	cursor := EncodeCursor(CurrentCursorVersion, 25)
	r := httptest.NewRequest("GET", "/x?limit=7&cursor="+cursor.String(), nil)
	req, err := ParseOffsetRequest(r, RequestConfig{
		LimitParam:    "limit",
		OffsetParam:   "offset",
		CursorParam:   "cursor",
		DefaultLimit:  20,
		MinLimit:      1,
		MaxLimit:      100,
		DefaultOffset: 0,
		MinOffset:     0,
		MaxOffset:     1000,
	})
	if err != nil {
		t.Fatalf("ParseOffsetRequest() error = %v", err)
	}
	if !req.UsedCursor {
		t.Fatalf("UsedCursor = false, want true")
	}
	if req.Limit != 7 {
		t.Fatalf("Limit = %d, want 7", req.Limit)
	}
	if req.Offset != 25 {
		t.Fatalf("Offset = %d, want 25", req.Offset)
	}
	if req.Cursor != cursor {
		t.Fatalf("Cursor = %q, want %q", req.Cursor, cursor)
	}
}

func TestParseOffsetRequestRejectsOffsetAndCursorTogether(t *testing.T) {
	cursor := EncodeCursor(CurrentCursorVersion, 25)
	r := httptest.NewRequest("GET", "/x?limit=7&offset=3&cursor="+cursor.String(), nil)
	_, err := ParseOffsetRequest(r, RequestConfig{
		LimitParam:    "limit",
		OffsetParam:   "offset",
		CursorParam:   "cursor",
		DefaultLimit:  20,
		MinLimit:      1,
		MaxLimit:      100,
		DefaultOffset: 0,
		MinOffset:     0,
		MaxOffset:     1000,
	})
	if err != ErrAmbiguousPagination {
		t.Fatalf("ParseOffsetRequest() error = %v, want %v", err, ErrAmbiguousPagination)
	}
}

func TestNewOffsetInfo(t *testing.T) {
	total := 42
	info := NewOffsetInfo(4, 4, 4, &total, nil)
	if !info.HasMore {
		t.Fatalf("HasMore = false, want true")
	}
	if info.TotalItems == nil || *info.TotalItems != 42 {
		t.Fatalf("TotalItems = %#v, want 42", info.TotalItems)
	}
	if info.TotalPages == nil || *info.TotalPages != 11 {
		t.Fatalf("TotalPages = %#v, want 11", info.TotalPages)
	}
	if info.NextCursor == nil {
		t.Fatalf("NextCursor = nil, want value")
	}
	if info.Returned != 4 {
		t.Fatalf("Returned = %d, want 4", info.Returned)
	}

	lastPage := NewOffsetInfo(40, 10, 2, nil, nil)
	if lastPage.HasMore {
		t.Fatalf("HasMore = true, want false when returned < limit and total unknown")
	}

	emptyTotal := 0
	emptyPage := NewOffsetInfo(0, 10, 0, &emptyTotal, nil)
	if emptyPage.TotalPages == nil || *emptyPage.TotalPages != 0 {
		t.Fatalf("TotalPages = %#v, want 0", emptyPage.TotalPages)
	}
}

func TestParseOffsetRequestRejectsInvalidValues(t *testing.T) {
	r := httptest.NewRequest("GET", "/x?limit=abc", nil)
	_, err := ParseOffsetRequest(r, RequestConfig{
		LimitParam:    "limit",
		OffsetParam:   "offset",
		CursorParam:   "cursor",
		DefaultLimit:  20,
		MinLimit:      1,
		MaxLimit:      100,
		DefaultOffset: 0,
		MinOffset:     0,
		MaxOffset:     1000,
	})
	if err != ErrInvalidLimit {
		t.Fatalf("ParseOffsetRequest() error = %v, want %v", err, ErrInvalidLimit)
	}

	r = httptest.NewRequest("GET", "/x?offset=-1", nil)
	_, err = ParseOffsetRequest(r, RequestConfig{
		LimitParam:    "limit",
		OffsetParam:   "offset",
		CursorParam:   "cursor",
		DefaultLimit:  20,
		MinLimit:      1,
		MaxLimit:      100,
		DefaultOffset: 0,
		MinOffset:     0,
		MaxOffset:     1000,
	})
	if err != ErrInvalidOffset {
		t.Fatalf("ParseOffsetRequest() error = %v, want %v", err, ErrInvalidOffset)
	}
}
