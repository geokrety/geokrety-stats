---
title: GeoKrety Pagination and Sorting for JSON REST Collections
version: 2.0
date_created: 2026-03-22
last_updated: 2026-03-28
owner: GeoKrety Development Team
tags: [architecture, infrastructure, design, pagination, api, frontend, json-rest]
---

# Specification: GeoKrety Pagination and Sorting for JSON REST Collections

## 1. Introduction

This specification defines how GeoKrety paginates JSON collections inside the shared JSON REST envelope documented in [../json-rest/specification.md](../json-rest/specification.md). It covers page-based pagination, cursor-based pagination, sorting, filtering, infinite-scroll client behavior, and the minimum validation rules needed for consistent implementation.

This page is normative for new and explicitly migrated JSON endpoints. Existing endpoints that still emit legacy payloads must document that exception in OpenAPI and endpoint-specific docs until they are migrated.

## 2. Purpose & Scope

### Purpose

- Make paginated GeoKrety JSON collections consistent across endpoints
- Standardize page-based navigation for catalog-style lists
- Standardize cursor-based navigation for large collections and infinite scroll
- Keep sorting and filtering deterministic so clients can page safely
- Define link-driven client behavior instead of cursor-field-driven behavior

### In Scope

- Page-based pagination with `page` and `per_page`
- Cursor-based pagination with `limit` and opaque `cursor`
- Top-level pagination links in JSON REST collection responses
- Sorting and filtering rules that affect page identity
- Infinite-scroll frontend behavior
- Backend and frontend implementation guidance
- Validation and testing expectations

### Out of Scope

- Redefining the generic JSON REST resource-object model
- XML pagination behavior
- GraphQL pagination
- Bidirectional cursor navigation

## 3. Shared Collection Contract

Every paginated JSON collection response uses the shared top-level envelope:

```json
{
  "data": [],
  "meta": {},
  "links": {
    "self": "/api/v3/example"
  }
}
```

Rules:

- `data` contains resource objects as defined in [../json-rest/specification.md](../json-rest/specification.md)
- `meta` describes the current page state and applied query context
- `links` carries navigation URLs
- collection navigation is link-driven, not field-driven
- canonical pagination examples must not use nested `meta.pagination`

## 4. Endpoint Mode Selection and Migration Boundary

Each endpoint must document exactly one public pagination mode:

- `page`
- `cursor`

Rules:

- page-based endpoints accept `page` and `per_page`, and reject `limit` and `cursor`
- cursor-based endpoints accept `limit` and optional `cursor`, and reject `page` and `per_page`
- requests mixing pagination modes return `INVALID_PAGINATION_MODE` with HTTP 400
- OpenAPI must declare the supported mode for each endpoint explicitly
- this specification is mandatory for new endpoints and for endpoints that have been explicitly migrated to the JSON REST contract
- older endpoints may remain legacy temporarily, but they must document the exception instead of silently presenting legacy payloads as canonical

## 5. Page-Based Pagination

### 5.1 When to Use It

Use page-based pagination when:

- users expect numbered pages
- the result set is stable enough for page navigation
- exact total counts are meaningful for the UI
- direct jumps to page N are more important than append-only scrolling

### 5.2 Request Contract

Example request:

```text
GET /api/v1/users?page=2&per_page=2&status=active&sort=-created_at
```

Parameters:

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `page` | integer | No | 1-based page number, defaults to `1` |
| `per_page` | integer | No | Defaults to `20`, maximum `100` |
| `sort` | string | No | Comma-separated fields, `-field` means descending |
| filters | varies | No | Endpoint-specific filter parameters |

### 5.3 Response Contract

```json
{
  "data": [
    {
      "id": "101",
      "type": "user",
      "attributes": {
        "username": "alice",
        "status": "active",
        "created_at": "2026-03-01T12:00:00Z"
      },
      "relationships": {
        "profile": {
          "data": {
            "type": "profile",
            "id": "501"
          },
          "links": {
            "related": "/api/v1/profiles/501"
          }
        }
      },
      "links": {
        "self": "/api/v1/users/101"
      }
    },
    {
      "id": "102",
      "type": "user",
      "attributes": {
        "username": "bob",
        "status": "active",
        "created_at": "2026-03-02T08:30:00Z"
      },
      "links": {
        "self": "/api/v1/users/102"
      }
    }
  ],
  "meta": {
    "page": 2,
    "per_page": 2,
    "total_items": 50,
    "total_pages": 25,
    "execution_time_ms": 14,
    "filters": {
      "status": "active"
    },
    "sort": "-created_at,id"
  },
  "links": {
    "self": "/api/v1/users?page=2&per_page=2&status=active&sort=-created_at,id",
    "first": "/api/v1/users?page=1&per_page=2&status=active&sort=-created_at,id",
    "prev": "/api/v1/users?page=1&per_page=2&status=active&sort=-created_at,id",
    "next": "/api/v1/users?page=3&per_page=2&status=active&sort=-created_at,id",
    "last": "/api/v1/users?page=25&per_page=2&status=active&sort=-created_at,id"
  }
}
```

Rules:

- `links.self` matches the normalized request URL
- `links.prev` is omitted on the first page
- `links.next` is omitted on the last page
- `total_items` and `total_pages` are part of the external page-based contract
- backends may still compute the page internally with `LIMIT` and `OFFSET`, but the public contract remains page-based

Validation behavior:

- `page < 1` returns `INVALID_PAGE` with HTTP 400
- missing or too-small `per_page` resolves to the server default of `20`
- `per_page > 100` returns `LIMIT_EXCEEDED` with HTTP 400
- a first-page request against an empty collection returns HTTP 200 with `data: []`, `page: 1`, `total_items: 0`, `total_pages: 1`, and `links.self`, `links.first`, and `links.last` all pointing to the normalized first-page URL
- a request for `page > 1` against an empty collection returns `OUT_OF_BOUNDS` with HTTP 400
- a request for `page > total_pages` on a non-empty collection returns `OUT_OF_BOUNDS` with HTTP 400

## 6. Cursor-Based Pagination

### 6.1 When to Use It

Use cursor-based pagination when:

- the collection can grow very large
- the UI is append-only or infinite-scroll oriented
- new rows may appear while the client is paging
- deep page jumps are not required

### 6.2 Request Contract

Example request:

```text
GET /api/v1/moves?limit=20&cursor=eyJpZCI6MjAxfQ==&sort=-created_at
```

Parameters:

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `limit` | integer | No | Defaults to `20`, maximum `100` |
| `cursor` | string | No | Opaque, versioned token for the next window |
| `sort` | string | No | Deterministic sort expression |
| filters | varies | No | Endpoint-specific filter parameters |

### 6.3 Response Contract

```json
{
  "data": [
    {
      "id": "201",
      "type": "move",
      "attributes": {
        "lat": 48.8566,
        "lon": 2.3522,
        "created_at": "2026-03-28T10:00:00Z"
      }
    },
    {
      "id": "200",
      "type": "move",
      "attributes": {
        "lat": 51.5074,
        "lon": -0.1278,
        "created_at": "2026-03-28T09:58:00Z"
      }
    }
  ],
  "meta": {
    "limit": 20,
    "has_more": true,
    "execution_time_ms": 9,
    "sort": "-created_at,id"
  },
  "links": {
    "self": "/api/v1/moves?limit=20&cursor=eyJpZCI6MjAxfQ==&sort=-created_at,id",
    "next": "/api/v1/moves?limit=20&cursor=eyJpZCI6MjAwfQ==&sort=-created_at,id"
  }
}
```

Rules:

- `links.next` is present only when `meta.has_more` is `true`
- the cursor remains opaque to clients even if its implementation is a base64-encoded JSON payload
- the response body does not expose `nextCursor` as a canonical field
- cursor pagination is forward-only in this specification
- missing or too-small `limit` resolves to the server default of `20`
- `limit > 100` returns `LIMIT_EXCEEDED` with HTTP 400

## 7. Infinite Scroll Contract

The frontend contract for infinite-scroll collections is:

1. Request the first page without a cursor.
2. Append `data` to the current list.
3. If `meta.has_more` is true, call `links.next`.
4. Repeat until `links.next` is absent or `meta.has_more` is false.

Minimal pseudocode:

```text
GET /moves?limit=20
append data
if meta.has_more:
    call links.next
repeat
```

## 8. Sorting and Filtering

### 8.1 Sort Syntax

Canonical sort syntax is:

- `sort=-created_at`
- `sort=-last_move_at,id`
- `sort=name`

Rules:

- a leading `-` means descending order
- no prefix means ascending order
- multiple fields are comma-separated
- endpoints should allow at most three sort columns

### 8.2 Deterministic Sort Requirement

Cursor-based pagination requires a stable sort order, for example:

```sql
ORDER BY created_at DESC, id DESC
```

When a client omits the deterministic tie-breaker:

- the server appends the endpoint-defined unique final key automatically
- the server echoes the normalized sort expression in `meta.sort`
- the server uses the normalized sort expression in `links.self` and all pagination links

Reject these cases:

- unknown sort fields with `INVALID_SORT_FIELD`
- more than three sort columns with `SORT_COMPLEXITY_EXCEEDED`

Avoid non-deterministic or mutable primary ordering such as:

- `ORDER BY RANDOM()`
- `ORDER BY created_at DESC` without a unique tie-breaker
- `ORDER BY updated_at DESC` when `updated_at` can change during pagination

### 8.3 Filter and Sort Reset Behavior

- page-based navigation must reset to `page=1` when filters or sort change
- cursor-based navigation must restart from the first request when filters or sort change
- stale cursors must be rejected if they no longer match the active sort or filter context
- servers should preserve active filters and sort expressions in every navigation link they generate

### 8.4 Sort Discovery

OpenAPI is the authoritative place to document allowed sort fields. If an endpoint wants runtime discovery, it may also expose `meta.sortable_fields`, but that is optional and must not replace OpenAPI.

## 9. Security and Error Handling

Pagination security is based on:

- authorization checks on the underlying data
- parameterized SQL queries
- server-enforced page-size limits

Pagination security is not based on hiding cursor structure. Cursor opacity is a contract boundary, not a secrecy boundary.

Common errors:

- `INVALID_PAGINATION_MODE`
- `INVALID_PAGE`
- `OUT_OF_BOUNDS`
- `LIMIT_EXCEEDED`
- `INVALID_CURSOR`
- `CURSOR_VERSION_MISMATCH`
- `INVALID_SORT_FIELD`
- `SORT_COMPLEXITY_EXCEEDED`

Cursor failure mapping:

- malformed cursor: `INVALID_CURSOR`, HTTP 400
- cursor with unsupported version: `CURSOR_VERSION_MISMATCH`, HTTP 400
- cursor whose filter or sort context no longer matches the request: `INVALID_CURSOR`, HTTP 400

## 10. Backend and Frontend Implementation Guidance

### 10.1 Backend Guidance

- generate `links.self` from the normalized request URL
- generate page-based `first`, `prev`, `next`, and `last` links on the server
- generate cursor-based `links.next` only when more rows exist
- preserve all active filters and normalized sort expressions in generated links
- validate `page`, `per_page`, `limit`, `cursor`, and sort expressions before querying data
- page-based handlers may translate `page` and `per_page` into internal `LIMIT` and `OFFSET`

### 10.2 Frontend Guidance

Frontend composables should store the current `links.next` URL rather than parsing cursor internals. A minimal state model is:

```text
items: T[]
next_link: string | null
has_more: boolean
is_loading: boolean
error: Error | null
```

Behavior:

- call the first URL explicitly
- append returned `data`
- set `next_link` from `links.next`
- stop when `next_link` is absent or `meta.has_more` is false
- reset local state when filters or sort change

## 11. Testing and Rollout Guidance

### 11.1 Contract Tests

- validate that page-based responses always include the expected top-level links
- validate that cursor-based responses omit `links.next` on the final page
- validate first-page, middle-page, last-page, and empty-page behavior
- validate stale cursor rejection after filter or sort changes
- validate mixed-parameter rejection for unsupported pagination modes
- validate that generated links preserve filters and normalized sort expressions

### 11.2 Documentation Validation

- keep all canonical examples valid JSON
- ensure the pagination page links to the global JSON REST page
- ensure docs navigation includes the new JSON REST section
- ensure OpenAPI examples and top-level wording do not contradict this page

### 11.3 Operational Guidance

- verify stable-sort indexes before production rollout of cursor endpoints
- enforce `per_page` and `limit` maximums server-side
- prefer cursor-based pagination for large and append-only collections
- use page-based pagination only where exact page navigation is part of the product requirement
- document legacy endpoints explicitly until they are migrated to this contract

## 12. Acceptance Criteria

- canonical paginated examples use top-level `data`, `meta`, and `links`
- page-based examples use `page`, `per_page`, `total_items`, and `total_pages`
- cursor-based examples use `limit`, `has_more`, and `links.next`
- infinite-scroll guidance tells clients to follow `links.next` instead of parsing cursor fields from the response body
- sort syntax is documented with a JSON REST-friendly expression such as `-created_at,id`
- each endpoint documents exactly one public pagination mode and rejects mixed-mode parameters
- stable sort, filter reset, cursor-versioning, and rollout-boundary guidance remain documented
- the page references the shared JSON REST envelope instead of redefining it from scratch

## 13. Implementation Checklist

- [x] Rewrite canonical examples to use resource objects inside `data`
- [x] Replace `meta.pagination` examples with top-level `links`
- [x] Change public page-based examples from offset-oriented to page-oriented
- [x] Keep cursor-based examples link-driven and forward-only
- [x] Define endpoint mode selection and mixed-parameter rejection
- [x] Preserve stable sorting, filter-reset, rollout-boundary, and validation guidance
- [x] Link this page to the global JSON REST API specification
