package jsonrest

import "time"

type Links map[string]string

type Meta map[string]any

type Document struct {
	Data     any        `json:"data"`
	Included []Resource `json:"included,omitempty"`
	Meta     Meta       `json:"meta"`
	Links    Links      `json:"links,omitempty"`
}

type Resource struct {
	ID            string                  `json:"id"`
	Type          string                  `json:"type"`
	Attributes    map[string]any          `json:"attributes,omitempty"`
	Relationships map[string]Relationship `json:"relationships,omitempty"`
	Links         Links                   `json:"links,omitempty"`
}

type Relationship struct {
	Data  any   `json:"data,omitempty"`
	Links Links `json:"links,omitempty"`
}

type Identifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ErrorDocument struct {
	Error     ErrorPayload `json:"error"`
	Timestamp string       `json:"timestamp"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewMeta(started time.Time) Meta {
	meta := Meta{}
	meta.Set("execution_time_ms", time.Since(started).Milliseconds())
	return meta
}

func (m Meta) Set(key string, value any) Meta {
	if m == nil {
		m = Meta{}
	}
	if key == "" || value == nil {
		return m
	}
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return m
		}
	case Links:
		if len(typed) == 0 {
			return m
		}
	case map[string]any:
		if len(typed) == 0 {
			return m
		}
	}
	m[key] = value
	return m
}

func (m Meta) WithPage(req PageRequest, totalItems, totalPages int) Meta {
	m.Set("page", map[string]any{
		"type":        "page",
		"number":      req.Page,
		"size":        req.PerPage,
		"total_items": totalItems,
		"total_pages": totalPages,
	})
	return m
}

func (m Meta) WithCursor(req CursorRequest, hasMore bool) Meta {
	m.Set("page", map[string]any{
		"type":     "cursor",
		"limit":    req.Limit,
		"has_more": hasMore,
	})
	return m
}

func (m Meta) WithSort(sort string) Meta {
	if sort == "" {
		return m
	}
	query := metaMapValue(m, "query")
	query["sort"] = sort
	m["query"] = query
	return m
}

func (m Meta) WithFilters(filters map[string]any) Meta {
	if len(filters) == 0 {
		return m
	}
	query := metaMapValue(m, "query")
	query["filters"] = filters
	m["query"] = query
	return m
}

func (m Meta) WithCapabilities(filters map[string]any, sorts []string) Meta {
	capabilities := map[string]any{}
	if len(filters) > 0 {
		capabilities["filters"] = filters
	}
	if len(sorts) > 0 {
		capabilities["sorts"] = sorts
	}
	if len(capabilities) == 0 {
		return m
	}
	m["capabilities"] = capabilities
	return m
}

func NewDocument(data any, included []Resource, meta Meta, links Links) Document {
	return Document{Data: data, Included: included, Meta: meta, Links: links}
}

func NewErrorDocument(code, message string, now time.Time) ErrorDocument {
	return ErrorDocument{
		Error: ErrorPayload{
			Code:    code,
			Message: message,
		},
		Timestamp: now.UTC().Format(time.RFC3339),
	}
}

func metaMapValue(meta Meta, key string) map[string]any {
	if meta == nil {
		meta = Meta{}
	}
	if existing, ok := meta[key].(map[string]any); ok {
		return existing
	}
	return map[string]any{}
}
