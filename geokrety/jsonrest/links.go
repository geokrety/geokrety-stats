package jsonrest

import (
	"net/http"
	"net/url"
	"strconv"
)

func PageLinks(r *http.Request, req PageRequest, totalPages int, pageParam, perPageParam string) Links {
	links := Links{}
	links.Set("self", normalizedLink(r, func(values url.Values) {
		values.Set(pageParam, strconv.Itoa(req.Page))
		values.Set(perPageParam, strconv.Itoa(req.PerPage))
	}))
	links.Set("first", normalizedLink(r, func(values url.Values) {
		values.Set(pageParam, "1")
		values.Set(perPageParam, strconv.Itoa(req.PerPage))
	}))
	if req.Page > 1 {
		links.Set("prev", normalizedLink(r, func(values url.Values) {
			values.Set(pageParam, strconv.Itoa(req.Page-1))
			values.Set(perPageParam, strconv.Itoa(req.PerPage))
		}))
	}
	if totalPages > 0 {
		links.Set("last", normalizedLink(r, func(values url.Values) {
			values.Set(pageParam, strconv.Itoa(totalPages))
			values.Set(perPageParam, strconv.Itoa(req.PerPage))
		}))
	}
	if totalPages > 0 && req.Page < totalPages {
		links.Set("next", normalizedLink(r, func(values url.Values) {
			values.Set(pageParam, strconv.Itoa(req.Page+1))
			values.Set(perPageParam, strconv.Itoa(req.PerPage))
		}))
	}
	return links
}

func CursorLinks(r *http.Request, req CursorRequest, nextCursor *Cursor, limitParam, cursorParam string) Links {
	links := Links{}
	links.Set("self", normalizedLink(r, func(values url.Values) {
		values.Set(limitParam, strconv.Itoa(req.Limit))
		if req.UsedCursor && !req.Cursor.IsEmpty() {
			values.Set(cursorParam, req.Cursor.String())
		} else {
			values.Del(cursorParam)
		}
	}))
	if nextCursor != nil && !nextCursor.IsEmpty() {
		links.Set("next", normalizedLink(r, func(values url.Values) {
			values.Set(limitParam, strconv.Itoa(req.Limit))
			values.Set(cursorParam, nextCursor.String())
		}))
	}
	return links
}

func SelfLink(r *http.Request) string {
	return normalizedLink(r, nil)
}

func (l Links) Set(name, href string) Links {
	if l == nil {
		l = Links{}
	}
	if name == "" || href == "" {
		return l
	}
	l[name] = href
	return l
}

func normalizedLink(r *http.Request, mutate func(url.Values)) string {
	if r == nil || r.URL == nil {
		return ""
	}
	clone := *r.URL
	values := clone.Query()
	if mutate != nil {
		mutate(values)
	}
	clone.RawQuery = values.Encode()
	if clone.RawQuery == "" {
		return clone.Path
	}
	return clone.Path + "?" + clone.RawQuery
}
