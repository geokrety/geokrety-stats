package pagination

import (
	"net/http"
	"strconv"
)

type RequestConfig struct {
	LimitParam    string
	OffsetParam   string
	CursorParam   string
	DefaultLimit  int
	MinLimit      int
	MaxLimit      int
	DefaultOffset int
	MinOffset     int
	MaxOffset     int
}

type OffsetRequest struct {
	Limit      int
	Offset     int
	Cursor     Cursor
	UsedCursor bool
}

type CursorInfo struct {
	Type       string  `json:"type,omitempty" xml:"type,omitempty"`
	Cursor     Cursor  `json:"cursor,omitempty" xml:"cursor,omitempty"`
	NextCursor *Cursor `json:"nextCursor,omitempty" xml:"nextCursor,omitempty"`
	HasMore    bool    `json:"hasMore" xml:"hasMore"`
	Count      int     `json:"count" xml:"count"`
}

type OffsetInfo struct {
	Type       string  `json:"type,omitempty" xml:"type,omitempty"`
	Limit      int     `json:"limit" xml:"limit"`
	Offset     int     `json:"offset" xml:"offset"`
	Cursor     Cursor  `json:"cursor,omitempty" xml:"cursor,omitempty"`
	NextCursor *Cursor `json:"nextCursor,omitempty" xml:"nextCursor,omitempty"`
	HasMore    bool    `json:"hasMore" xml:"hasMore"`
	Count      int     `json:"count" xml:"count"`
	Returned   int     `json:"returned,omitempty" xml:"returned,omitempty"`
	TotalItems *int    `json:"totalItems,omitempty" xml:"totalItems,omitempty"`
	TotalPages *int    `json:"totalPages,omitempty" xml:"totalPages,omitempty"`
}

type Meta struct {
	Pagination CursorInfo `json:"pagination" xml:"pagination"`
	Sort       []string   `json:"sort,omitempty" xml:"sort>item,omitempty"`
}

type MetaOffset struct {
	Pagination OffsetInfo `json:"pagination" xml:"pagination"`
	Sort       []string   `json:"sort,omitempty" xml:"sort>item,omitempty"`
}

type Page[T any] struct {
	Data []T  `json:"data" xml:"data>item"`
	Meta Meta `json:"meta" xml:"meta"`
}

type PageOffset[T any] struct {
	Data []T        `json:"data" xml:"data>item"`
	Meta MetaOffset `json:"meta" xml:"meta"`
}

func ParseOffsetRequest(r *http.Request, cfg RequestConfig) (OffsetRequest, error) {
	limit, err := parseOptionalInt(r, cfg.LimitParam, cfg.DefaultLimit, cfg.MinLimit, cfg.MaxLimit, ErrInvalidLimit)
	if err != nil {
		return OffsetRequest{}, err
	}
	request := OffsetRequest{Limit: limit, Offset: cfg.DefaultOffset}
	cursorValue := r.URL.Query().Get(cfg.CursorParam)
	if cursorValue != "" {
		if _, hasOffset := r.URL.Query()[cfg.OffsetParam]; hasOffset {
			return OffsetRequest{}, ErrAmbiguousPagination
		}
		cursor := Cursor(cursorValue)
		_, offset, err := cursor.Decode()
		if err != nil {
			return OffsetRequest{}, err
		}
		if offset < int64(cfg.MinOffset) || offset > int64(cfg.MaxOffset) {
			return OffsetRequest{}, ErrInvalidCursor
		}
		request.Offset = int(offset)
		request.Cursor = cursor
		request.UsedCursor = true
		return request, nil
	}
	offset, err := parseOptionalInt(r, cfg.OffsetParam, cfg.DefaultOffset, cfg.MinOffset, cfg.MaxOffset, ErrInvalidOffset)
	if err != nil {
		return OffsetRequest{}, err
	}
	request.Offset = offset
	return request, nil
}

func NewCursorInfo(cursor Cursor, nextCursor *Cursor, count int) CursorInfo {
	return CursorInfo{
		Type:       "cursor",
		Cursor:     cursor,
		NextCursor: nextCursor,
		HasMore:    nextCursor != nil,
		Count:      count,
	}
}

func NewOffsetInfo(offset, limit, returned int, totalItems *int, hasMore *bool) OffsetInfo {
	info := OffsetInfo{
		Type:     "offset",
		Limit:    limit,
		Offset:   offset,
		HasMore:  false,
		Count:    returned,
		Returned: returned,
	}
	if offset > 0 {
		info.Cursor = EncodeCursor(CurrentCursorVersion, int64(offset))
	}
	if hasMore != nil {
		info.HasMore = *hasMore
	}
	if totalItems != nil {
		info.TotalItems = totalItems
		if limit > 0 {
			pages := 0
			if *totalItems > 0 {
				pages = (*totalItems + limit - 1) / limit
			}
			info.TotalPages = &pages
		}
		info.HasMore = offset+returned < *totalItems
	}
	if info.HasMore {
		next := EncodeCursor(CurrentCursorVersion, int64(offset+returned))
		info.NextCursor = &next
	}
	return info
}

func NewPage[T any](items []T, cursor Cursor, nextCursor *Cursor, sort []string) *Page[T] {
	return &Page[T]{
		Data: items,
		Meta: Meta{
			Pagination: NewCursorInfo(cursor, nextCursor, len(items)),
			Sort:       sort,
		},
	}
}

func NewPageOffset[T any](items []T, offset, limit int, totalItems *int, sort []string) *PageOffset[T] {
	return &PageOffset[T]{
		Data: items,
		Meta: MetaOffset{
			Pagination: NewOffsetInfo(offset, limit, len(items), totalItems, nil),
			Sort:       sort,
		},
	}
}

func queryInt(r *http.Request, key string, fallback, minValue, maxValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return clampInt(parsed, minValue, maxValue)
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func parseOptionalInt(r *http.Request, key string, fallback, minValue, maxValue int, invalidErr error) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, invalidErr
	}
	if parsed < minValue || parsed > maxValue {
		return 0, invalidErr
	}
	return parsed, nil
}
