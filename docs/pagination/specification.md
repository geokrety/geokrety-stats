---
title: Generic Pagination Implementation for JSON REST APIs and Infinity Scrolling
version: 1.0
date_created: 2026-03-22
last_updated: 2026-03-22
owner: GeoKrety Development Team
tags: [architecture, infrastructure, design, pagination, api, frontend]
---

# Specification: Generic Pagination Implementation for JSON REST APIs and Infinity Scrolling

## 1. Introduction

This specification defines a robust, generic, and reusable pagination system for GeoKrety that supports both **cursor-based pagination** (for infinite scrolling scenarios) and **offset-based pagination** (for traditional page-by-page navigation). The system is designed to work across different data sources (SQL databases, REST APIs, GraphQL, in-memory), provide strong type safety in both Go (backend) and TypeScript (frontend), and eliminate pagination implementation inconsistencies across the codebase.

The pagination system is intentionally lightweight and library-like, not a framework, to ensure flexibility and minimal coupling with specific business logic. It provides clean abstractions for both API developers (who emit paginated responses) and frontend developers (who consume pagination through Vue 3 composition functions with transparent state management).

---

## 2. Purpose & Scope

### Purpose
To provide a unified, generic pagination abstraction that:

- Eliminates inconsistent pagination implementations across GeoKrety APIs
- Enables infinite scrolling user experiences without complex frontend state management
- Supports high-performance pagination over large datasets (millions of records)
- Maintains type safety throughout the pagination lifecycle (Go + TypeScript)
- Prevents common pagination bugs (SQL injection, incorrect offset handling, state management issues)
- Makes pagination a straightforward, reusable pattern rather than a per-endpoint concern

### Scope
**In Scope:**

- Generic type definitions and implementations (Go backend, TypeScript frontend)
- Cursor-based pagination for infinite scroll use cases
- Offset-based pagination for traditional page-by-page navigation
- API response envelope patterns and examples
- Vue 3 composition functions for state management
- Example endpoint implementations
- Unit and integration test patterns
- Security considerations and attack prevention strategies

**Out of Scope:**

- GraphQL pagination (different conventions; separate concern)
- UI-specific frameworks beyond Vue 3 (patterns are reusable)
- Database-specific optimizations (core logic remains DB-agnostic)
- ORM-specific implementations (examples use raw SQL for clarity)

### Intended Audience
- Backend API developers (Go) extending GeoKrety with new paginated endpoints
- Frontend developers (TypeScript/Vue 3) integrating pagination into UI components
- Database architects optimizing pagination query performance
- Security reviewers auditing pagination implementations

---

## 3. Definitions

- **Cursor**: An opaque, versioned token representing a position in a paginated result set. Clients pass it back to request subsequent pages. The internal structure is not client-visible.

- **Offset**: A numeric position (0-based) in a sorted result set, used alongside a limit to define page bounds.

- **Limit / Page Size**: Maximum number of items to return in a single response. Server enforces reasonable defaults and maximums.

- **Pagination (Generic Type)**: A type parameter `Pagination<T>` representing a typed collection that can be paginated (T = item type).

- **Page<T>**: Response envelope containing paginated items of type T, along with metadata (cursors, totals, etc.).

- **Cursor-Based Pagination**: Pagination strategy using opaque cursors; optimal for infinite scroll, real-time feeds, and append-only data. Stateless from client perspective.

- **Offset-Based Pagination**: Pagination strategy using offset + limit; suitable for traditional page-by-page navigation, filters/sorts that are stable.

- **Infinity Scrolling**: Progressive loading pattern where new items appear as user scrolls to bottom; implemented via cursor-based pagination with automatic `fetchNextPage()` calls.

- **Cursor Enumeration Attack**: Security threat where attacker potentially guesses cursor values. **NOTE: Pagination data is public; threat model assumes adversary uses different cursors via URL params.** Real mitigation is authorization (don't return unauthorized data) + SQL safety (no injection), NOT cursor obfuscation.

- **Stable Sort Order**: A sort order that remains consistent across queries (e.g., `ORDER BY created_at DESC, id DESC`). Required for reliable cursor-based pagination.

- **Versioning**: Pagination cursors include a version identifier, allowing cursor format evolution without breaking existing clients.

---

## 4. Requirements, Constraints & Guidelines

### Functional Requirements

- **FR-001**: System must support cursor-based pagination for infinite scroll use cases with opaque, versioned cursors.

- **FR-002**: System must support offset-based pagination for traditional page-by-page navigation, with optional total counts when the endpoint chooses to expose page-number UX.

- **FR-003**: Cursors must be versioned to support backward compatibility when cursor format changes.

- **FR-004**: System must work with any data type `T` (generic implementation); not coupled to GeoKrety-specific types.

- **FR-005**: Cursor format must be opaque and versioned; invalid, malformed, or context-mismatched cursors must be rejected with clear error messages.

- **FR-006**: Response envelopes must be consistent across all paginated endpoints (cursor-based and offset-based).

- **FR-007**: System must gracefully handle empty result sets, returning empty `data[]` with appropriate metadata.

- **FR-008**: System must support filtering and sorting alongside pagination; any change to filter or sort context must invalidate existing cursors and require a reset to the first page.

- **FR-009**: Frontend Vue 3 composables must provide transparent state management for infinity scroll without requiring manual cursor handling.

- **FR-010**: System must enforce server-side limits (max page size, reasonable defaults) to prevent resource exhaustion.

### Non-Functional Requirements

- **NFR-001** (Performance): Pagination logic must not degrade with large datasets (millions of records). Cursor encoding/decoding < 1ms. Database queries with pagination must use indexed sorts.

- **NFR-002** (Type Safety): TypeScript implementation must be fully typed; no `any` types or implicit `unknown`. Go implementation must use generics where applicable.

- **NFR-003** (Security): Authorization checks must prevent unauthorized data access across paginated endpoints. SQL queries must use parameterized statements to prevent injection. Cursor formats may remain simple base64-encoded payloads as long as clients treat them as opaque.

- **NFR-004** (Maintainability): All functions, types, and methods must have JSDoc (TS) / GoDoc (Go) comments explaining purpose and usage. Examples in comments encouraged.

- **NFR-005** (Testing): Unit tests must achieve >90% code coverage for pagination logic. Integration tests must validate cursor handling with real database backends.

### Constraints

- **CON-001**: No external pagination library dependencies. Implement from first principles to demonstrate understanding of pagination mechanics.

- **CON-002**: Cursor encodings must support versioning; breaking changes to cursor format must increment version identifier.

- **CON-003**: Total item counts (for offset pagination) are optional; expensive for large datasets and should default to `undefined` unless explicitly enabled.

- **CON-004**: Pagination is unidirectional (forward-only via `nextCursor`); bidirectional pagination adds complexity and is deferred.

- **CON-005**: Sort order must be stable (deterministic) to ensure reliable cursor-based pagination. Avoid non-deterministic sorts like database-assigned UUIDs without secondary sort keys.

### Guidelines

- **GUD-001**: Prefer cursor-based pagination for append-only, real-time feeds (user activity, notifications, comments). Prefer offset-based for catalog-like results with stable filters/sorts.

- **GUD-002**: Cursor opacity is non-negotiable; clients must never parse cursor structure. Server reserves right to change cursor format in next version.

- **GUD-003**: API endpoints should default to cursor-based pagination for new features; migrate existing offset-based endpoints only if necessary.

- **GUD-004**: Document the stable sort order explicitly in API reference. Example: `ORDER BY created_at DESC, id DESC`.

- **GUD-005**: Frontend components should use `usePagination()` composition for infinite scroll patterns to minimize boilerplate.

---

## 5. Pagination Patterns

### Pagination Patterns

#### 5.1 Cursor-Based Pagination

**What it is:** Pagination using an opaque token (cursor) representing the last visible row in a deterministic sort order. Clients request the first page, receive items plus a `nextCursor`, and use that cursor to seek the next window without relying on SQL OFFSET.

**When to use:**

- Infinite scroll feeds (user activity, notifications)
- Real-time data that changes during pagination (new items appear)
- Append-only datasets (logs, transactions, audit trails)
- Performance-critical scenarios (avoids expensive OFFSET queries on large tables)

**Why:**

- Stateless from client perspective (cursor encapsulates position)
- Efficient for large datasets when backed by matching keyset predicates and indexes
- More resilient to concurrent inserts than OFFSET-based pagination because the next page is anchored on the last visible row
- Natural fit for infinite scroll UX patterns

**Trade-offs:**

- ❌ Cannot jump to arbitrary page (must follow cursor chain)
- ❌ Cannot navigate backward (cursor points forward only)
- ✅ Better performance than offset pagination on large tables
- ✅ More stable under concurrent inserts when using deterministic keyset ordering

#### 5.2 Offset-Based Pagination

**What it is:** Pagination using numeric offset (position) and limit (page size). Clients specify which range of items they want (`offset=20&limit=10`).

**When to use:**

- Catalog-like features (product listings, search results, filterable lists)
- UI with "page 1, 2, 3" navigation
- Stable result sets (filters/sort order don't change mid-pagination)
- Small-to-medium datasets where OFFSET cost is acceptable

**Why:**

- Intuitive for users ("Page 5 of 20")
- Easy to jump to arbitrary pages
- Supports backward pagination (previous page)
- Familiar RESTful pattern

**Trade-offs:**

- ❌ OFFSET becomes expensive on large tables (O(n) cost to scan to position)
- ❌ Sensitive to concurrent data changes (new items shift offsets)
- ✅ Jump-to-page UX is simpler
- ✅ Total count is meaningful

#### 5.3 Infinity Scrolling State Model

Infinity scrolling is implemented via cursor-based pagination + frontend state management:

```
State: {
  items: T[]              // Accumulated items from all fetched pages
  currentCursor: string | null
  nextCursor: string | null
  hasMore: boolean        // true if nextCursor is not null
  isLoading: boolean      // true during fetch
  error: Error | null
}

User scrolls to bottom:
  - Call fetchNextPage()
  - Set isLoading = true, issueFetch(nextCursor)
  - On response: append items, update nextCursor, setisLoading = false
  - If nextCursor is null, set hasMore = false (end reached)
```

#### 5.4 Comparison Matrix

| Criterion | Cursor-Based | Offset-Based |
|-----------|-------------|------------|
| **Jump to Page** | ❌ No | ✅ Yes |
| **Backward Pagination** | ❌ No | ✅ Yes |
| **Large Datasets** | ✅ Fast | ❌ Slow (OFFSET) |
| **Real-time Updates** | ✅ Resilient | ❌ Fragile |
| **Infinity Scroll** | ✅ Natural | ⚠️ Possible but awkward |
| **Implementation Complexity** | ⚠️ Medium | ✅ Simple |
| **User Pagination Display** | ❌ Can't show "2/47" | ✅ Can show "Page 2 of 47" |

---

### 5.5 Sorting Specification

### API Contract (Request)

```
GET /api/v3/users/{id}/geokrety-found?sort=last_move_at:desc,id:asc&limit=20
GET /api/v3/geokrety/search?q=geokrety&sort=relevance:desc,last_move_at:desc&limit=20&offset=0
```

**Format:** `field1:order1,field2:order2`
**Orders:** `asc` (ascending), `desc` (descending)

### Sorting Rules

- **Stable Sort**: Must include a unique tiebreaker, usually `id`, as the final sort column to ensure pagination stability
- **Default Sort**: Each endpoint documents its default sort order
- **Allowed Fields**: API response should include `sortableFields: ["last_move_at", "relevance", "id"]`
- **Multi-Column**: Supports up to 3 sort columns (prevent complexity)

### Response Indication

Include current sort in response metadata:

```json
{
  "data": [...],
  "meta": {
    "sort": ["last_move_at:desc", "id:asc"],
    "pagination": {...}
  }
}
```

### Sorting + Cursor Interaction

- **Cursor-based pagination**: When user changes sort mid-pagination, reset to first page (new sort may reorder results)
- **Offset-based pagination**: When user changes sort mid-pagination, reset offset to 0
- Document this in error message when old cursor used with new sort

### Sorting + Filtering Interaction

- Filters can change which rows match; sorting still applies consistently
- If filtering changes mid-pagination, reset to first page
- Example: Search results sorted by relevance; user adds new filter term; start from page 0

### Examples

- **Recent GeoKrety feed** (cursor-based): `?sort=last_move_at:desc,id:desc&cursor=...`
- **Search results** (offset-based): `?sort=relevance:desc,last_move_at:desc&offset=0&limit=20`
- **Comment thread** (cursor-based): `?sort=created_at:asc,id:asc&cursor=...` (oldest first)

---

## 6. API Contract & Data Models

### 6.1 Request Envelope: Cursor-Based Pagination

**Query Parameters:**
```
GET /api/v3/users/{id}/geokrety-found?limit=20&cursor=eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAzLTIyVDEwOjMwOjAwWiIsImlkIjo0Mn0sIkMiOiJ1c2VyOjQyfHNvcnQ6bGFzdF9tb3ZlX2F0OmRlc2MsaWQ6ZGVzY3xmaWx0ZXI6YWxsIn0=
```

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `limit` | int | No | Page size (items per request). Default: 20. Max: 100. |
| `cursor` | string | No | Opaque pagination token. Omit for first page. |
| `sort` | string | No | Sort expression using `field:order[,field:order]`. Defaults are endpoint-specific and must be documented. |

**Default Behavior:**

- Omitting `cursor` requests the first page
- Omitting `limit` uses server default (20)
- Server enforces `limit ≤ maxLimit` (100)

### 6.2 Request Envelope: Offset-Based Pagination

**Query Parameters:**
```
GET /api/v3/geokrety/search?q=geocaching&limit=20&offset=40
```

| Parameter | Type | Required | Notes |
|-----------|------|----------|-------|
| `q` | string | No | Search query (example). |
| `limit` | int | No | Page size. Default: 20. Max: 100. |
| `offset` | int | No | Numeric position (0-based). Default: 0. |
| `sort` | string | No | Sort expression using `field:order[,field:order]`. Defaults are endpoint-specific and must be documented. |

### 6.3 Response Envelope: Cursor-Based Pagination

```json
{
  "data": [
    {
      "gkid": "GK0001",
      "type": "geocached",
      "lastMoveAt": "2026-03-22T10:30:00Z",
      "location": "Eiffel Tower, Paris"
    },
    ...
  ],
  "meta": {
    "requestedAt": "2026-03-22T15:00:01Z",
    "queryMs": 12,
    "pagination": {
      "cursor": null,
      "nextCursor": "eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAzLTIyVDEwOjEwOjAwWiIsImlkIjoyMn0sIkMiOiJ1c2VyOjQyfHNvcnQ6bGFzdF9tb3ZlX2F0OmRlc2MsaWQ6ZGVzY3xmaWx0ZXI6YWxsIn0=",
      "hasMore": true,
      "count": 20
    }
  }
}
```

**Fields:**

- `data`: Array of items (type `T[]`)
- `meta.requestedAt`, `meta.queryMs`: Reuse the repository's shared response metadata so paginated endpoints stay compatible with the existing response envelope helpers
- `meta.pagination.cursor`: Cursor used to fetch this page. `null` on the first page because the request omitted `cursor`
- `meta.pagination.nextCursor`: Cursor for next page. `null` if no more pages.
- `meta.pagination.hasMore`: Boolean convenience flag (`nextCursor !== null`)
- `meta.pagination.count`: Items in this response (0 to limit)

**Repository alignment:** In GeoKrety Stats, cursor pagination extends the existing shared `meta` envelope rather than replacing it. `requestedAt` and `queryMs` remain present, and `meta.pagination` is expanded from the current offset-only helper to also carry cursor fields.

### 6.4 Response Envelope: Offset-Based Pagination

```json
{
  "data": [
    {
      "gkid": "GK0A1F",
      "name": "Hidden Treasure",
      "lastMoveAt": "2025-10-15T14:22:00Z"
    },
    ...
  ],
  "meta": {
    "requestedAt": "2025-10-15T14:22:03Z",
    "queryMs": 18,
    "pagination": {
      "offset": 40,
      "limit": 20,
      "totalItems": 847,
      "totalPages": 43,
      "hasMore": true,
      "count": 20
    }
  }
}
```

**Fields:**

- `data`: Array of items
- `meta.pagination.offset`: Position of first item in this response
- `meta.pagination.limit`: Requested page size
- `meta.pagination.totalItems`: Total items matching query (may be expensive; consider optional)
- `meta.pagination.totalPages`: `ceil(totalItems / limit)`
- `meta.pagination.hasMore`: `offset + limit < totalItems`
- `meta.pagination.count`: Items in this response

**Note on totalItems:** Including `totalItems` requires a full table scan (expensive). Consider:

- Returning it only on first request
- Returning `null` after first page (client caches total)
- Using a fast approximation (last known count) instead of exact count

### 6.5 Error Responses

#### Invalid Cursor

```json
{
  "error": {
    "code": "INVALID_CURSOR",
        "message": "Cursor is malformed, expired, or does not match the current request context."
    },
    "timestamp": "2026-03-22T15:02:11Z"
}
```

#### Cursor Version Mismatch

```json
{
  "error": {
    "code": "CURSOR_VERSION_MISMATCH",
        "message": "Cursor version 2 is not supported by this server version. Please upgrade."
    },
    "timestamp": "2026-03-22T15:02:11Z"
}
```

#### Out of Bounds (Offset)

```json
{
  "error": {
    "code": "OUT_OF_BOUNDS",
        "message": "Offset 500 exceeds total items (247)."
    },
    "timestamp": "2026-03-22T15:02:11Z"
}
```

#### Max Limit Exceeded

```json
{
  "error": {
    "code": "LIMIT_EXCEEDED",
        "message": "Requested limit 500 exceeds maximum allowed limit 100."
    },
    "timestamp": "2026-03-22T15:02:11Z"
}
```

#### Invalid Sort Field

```json
{
  "error": {
    "code": "INVALID_SORT_FIELD",
        "message": "Sort field 'invalid_field' is not sortable. Allowed fields: last_move_at, relevance, id."
    },
    "timestamp": "2026-03-22T15:02:11Z"
}
```

#### Invalid Sort Order

```json
{
  "error": {
    "code": "INVALID_SORT_ORDER",
        "message": "Sort order must be 'asc' or 'desc', got 'invalid'."
    },
    "timestamp": "2026-03-22T15:02:11Z"
}
```

#### Sort Complexity Exceeded

```json
{
  "error": {
    "code": "SORT_COMPLEXITY_EXCEEDED",
        "message": "Maximum 3 sort columns allowed, got 4."
    },
    "timestamp": "2026-03-22T15:02:11Z"
}
```

**Repository alignment:** The current shared error helper already emits `error` plus `timestamp`. To implement this specification without diverging from repository conventions, extend that helper so `error` becomes an object with at least `code` and `message`, while keeping `timestamp` unchanged.

### 6.6 Example: User Activity Feed (Cursor-Based Infinity Scroll)

**First Request:**
```
GET /api/v3/users/42/geokrety-found?sort=last_move_at:desc,id:desc&limit=20

Response 200:
{
  "data": [
    { "id": 1, "type": "geocached", "gkid": "GK0001", "when": "2026-03-22T15:00:00Z" },
    { "id": 2, "type": "spotted", "gkid": "GK0002", "when": "2026-03-22T14:30:00Z" },
    ...
  ],
  "meta": {
    "requestedAt": "2026-03-22T15:00:01Z",
    "queryMs": 12,
    "pagination": {
      "cursor": null,
      "nextCursor": "eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAzLTIyVDE0OjEwOjAwWiIsImlkIjoyMH0sIkMiOiJ1c2VyOjQyfHNvcnQ6bGFzdF9tb3ZlX2F0OmRlc2MsaWQ6ZGVzY3xmaWx0ZXI6YWxsIn0=",
      "hasMore": true,
      "count": 20
    },
    "sort": ["last_move_at:desc", "id:desc"],
    "sortableFields": ["last_move_at", "type", "id"]
  }
}
```

**Second Request (user scrolls):**
```
GET /api/v3/users/42/geokrety-found?sort=last_move_at:desc,id:desc&limit=20&cursor=eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAzLTIyVDE0OjEwOjAwWiIsImlkIjoyMH0sIkMiOiJ1c2VyOjQyfHNvcnQ6bGFzdF9tb3ZlX2F0OmRlc2MsaWQ6ZGVzY3xmaWx0ZXI6YWxsIn0=

Response 200:
{
  "data": [
    { "id": 21, "type": "geocached", "gkid": "GK0AA1", "when": "2026-03-22T14:00:00Z" },
    ...
  ],
  "meta": {
        "requestedAt": "2026-03-22T15:00:04Z",
        "queryMs": 11,
    "pagination": {
            "cursor": "eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAzLTIyVDE0OjEwOjAwWiIsImlkIjoyMH0sIkMiOiJ1c2VyOjQyfHNvcnQ6bGFzdF9tb3ZlX2F0OmRlc2MsaWQ6ZGVzY3xmaWx0ZXI6YWxsIn0=",
            "nextCursor": "eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAzLTIyVDEzOjIwOjAwWiIsImlkIjo0MH0sIkMiOiJ1c2VyOjQyfHNvcnQ6bGFzdF9tb3ZlX2F0OmRlc2MsaWQ6ZGVzY3xmaWx0ZXI6YWxsIn0=",
      "hasMore": true,
      "count": 20
    },
        "sort": ["last_move_at:desc", "id:desc"],
        "sortableFields": ["last_move_at", "type", "id"]
  }
}
```

**Last Request (no more data):**
```
GET /api/v3/users/42/geokrety-found?sort=last_move_at:desc,id:desc&limit=20&cursor=eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAxLTAxVDAwOjA1OjAwWiIsImlkIjo0MDB9LCJDIjoidXNlcjo0Mnxzb3J0Omxhc3RfbW92ZV9hdDpkZXNjLGlkOmRlc2N8ZmlsdGVyOmFsbCJ9

Response 200:
{
  "data": [
    { "id": 401, "type": "spotted", "gkid": "GKFFFF", "when": "2026-01-01T00:00:00Z" }
  ],
  "meta": {
    "requestedAt": "2026-03-22T15:00:09Z",
    "queryMs": 9,
    "pagination": {
      "cursor": "eyJWIjoxLCJLIjp7Imxhc3RNb3ZlQXQiOiIyMDI2LTAxLTAxVDAwOjA1OjAwWiIsImlkIjo0MDB9LCJDIjoidXNlcjo0Mnxzb3J0Omxhc3RfbW92ZV9hdDpkZXNjLGlkOmRlc2N8ZmlsdGVyOmFsbCJ9",
      "nextCursor": null,
      "hasMore": false,
      "count": 1
    },
    "sort": ["last_move_at:desc", "id:desc"],
    "sortableFields": ["last_move_at", "type", "id"]
  }
}
```

### 6.7 Example: Search Results (Offset-Based Pagination)

**First Request:**
```
GET /api/v3/geokrety/search?q=hiking+cache&sort=relevance:desc,last_move_at:desc&limit=20&offset=0

Response 200:
{
  "data": [
    { "gkid": "GK0A1F", "name": "Forest Trail Cache", "type": "traditional" },
    ...
  ],
  "meta": {
        "requestedAt": "2026-03-22T15:00:12Z",
        "queryMs": 27,
    "pagination": {
      "offset": 0,
      "limit": 20,
      "totalItems": 847,
      "totalPages": 43,
      "hasMore": true,
      "count": 20
    },
        "sort": ["relevance:desc", "last_move_at:desc"],
        "sortableFields": ["relevance", "last_move_at", "name"]
  }
}
```

**Jump to Page 5 (offset=80):**
```
GET /api/v3/geokrety/search?q=hiking+cache&sort=relevance:desc,last_move_at:desc&limit=20&offset=80

Response 200:
{
  "data": [ ... (items 81-100) ... ],
  "meta": {
    "pagination": {
      "offset": 80,
      "limit": 20,
      "totalItems": 847,
      "totalPages": 43,
      "hasMore": true,
      "count": 20
    },
        "sort": ["relevance:desc", "last_move_at:desc"],
        "sortableFields": ["relevance", "last_move_at", "name"]
  }
}
```

---

## 7. Generic Type Definitions

### 7.1 Cursor Type (Go)

```go
// Cursor is an opaque, versioned pagination token.
// Clients must treat it as opaque and never parse or modify it.
type Cursor string

// CursorKey stores the last visible row for keyset pagination.
type CursorKey struct {
    LastMoveAt time.Time `json:"lastMoveAt"`
    ID         int64     `json:"id"`
}

// CursorPayload is the internal cursor structure.
type CursorPayload struct {
    V int       `json:"V"`
    K CursorKey `json:"K"`
    C string    `json:"C"` // Context fingerprint: user/sort/filter scope
}

// EncodeCursor creates a cursor from keyset position and request context.
func EncodeCursor(version int, key CursorKey, context string) Cursor {
    payload := CursorPayload{V: version, K: key, C: context}
    encodedBytes, _ := json.Marshal(payload)
    return Cursor(base64.StdEncoding.EncodeToString(encodedBytes))
}

// DecodeCursor parses a cursor into its payload.
func (c Cursor) Decode() (CursorPayload, error) {
    decoded, err := base64.StdEncoding.DecodeString(string(c))
    if err != nil {
        return CursorPayload{}, ErrInvalidCursor
    }

    var payload CursorPayload
    if err := json.Unmarshal(decoded, &payload); err != nil {
        return CursorPayload{}, ErrInvalidCursor
    }
    if payload.V != 1 {
        return CursorPayload{}, ErrCursorVersionMismatch
    }

    return payload, nil
}
```

### 7.2 Page Type (Go)

```go
// Meta contains metadata for a paginated response.
type Meta struct {
    RequestedAt    time.Time      `json:"requestedAt"`
    QueryMs        int64          `json:"queryMs"`
    Pagination     PaginationInfo `json:"pagination"`
    Sort           []string       `json:"sort,omitempty"`           // Current sort order
    SortableFields []string       `json:"sortableFields,omitempty"` // Allowed sort fields
}

// MetaOffset contains metadata for an offset-based paginated response.
type MetaOffset struct {
    RequestedAt    time.Time            `json:"requestedAt"`
    QueryMs        int64                `json:"queryMs"`
    Pagination     PaginationInfoOffset `json:"pagination"`
    Sort           []string             `json:"sort,omitempty"`           // Current sort order
    SortableFields []string             `json:"sortableFields,omitempty"` // Allowed sort fields
}

// Page is a paginated response containing items of type T.
type Page[T any] struct {
    Data []T   `json:"data"`
    Meta Meta  `json:"meta"`
}

// PaginationInfo contains pagination metadata for cursor-based pagination.
type PaginationInfo struct {
    Cursor     *Cursor `json:"cursor"`      // Cursor used to fetch this page; nil on first page
    NextCursor *Cursor `json:"nextCursor"`  // Cursor for next page (null if last page)
    HasMore    bool    `json:"hasMore"`     // Convenience flag
    Count      int     `json:"count"`       // Items in this response
}

// PaginationInfoOffset contains pagination metadata for offset-based pagination.
type PaginationInfoOffset struct {
    Offset     int    `json:"offset"`      // Position of first item
    Limit      int    `json:"limit"`       // Requested page size
    TotalItems *int   `json:"totalItems"`  // Total items (may be null, expensive to compute)
    TotalPages *int   `json:"totalPages"`  // Total pages (computed from totalItems)
    HasMore    bool   `json:"hasMore"`     // More pages available
    Count      int    `json:"count"`       // Items in this response
}

// PageOffset is a paginated response for offset-based pagination.
type PageOffset[T any] struct {
    Data []T        `json:"data"`
    Meta MetaOffset `json:"meta"`
}
```

### 7.3 Repository Interface (Go)

```go
// PaginationQuery encapsulates pagination parameters for a query.
type PaginationQuery struct {
    Cursor Cursor
    Limit  int
    Sort   []string
    Filter string
}

// PaginationQueryOffset encapsulates offset-based pagination parameters.
type PaginationQueryOffset struct {
    Offset      int
    Limit       int
    IncludeTotal bool  // If true, compute totalItems (expensive)
}

// Repository pattern for paginated data access.
type ActivityRepository interface {
    // GetActivityPaginated returns paginated user GeoKrety results.
    // sort: "last_move_at DESC" or similar (must be stable/deterministic)
    GetActivityPaginated(ctx context.Context, userID int, query PaginationQuery) (*Page[Activity], error)
}

// Implementation example using SQL.
func (r *sqlActivityRepo) GetActivityPaginated(
    ctx context.Context,
    userID int,
    query PaginationQuery,
) (*Page[Activity], error) {
    started := time.Now()
    const maxLimit = 100
    var after *CursorKey
    contextKey := fmt.Sprintf("user:%d|sort:%s|filter:%s", userID, strings.Join(query.Sort, ","), query.Filter)

    // If cursor provided, decode and validate it against the current request context.
    if query.Cursor != "" {
        payload, err := query.Cursor.Decode()
        if err != nil {
            return nil, fmt.Errorf("invalid cursor: %w", err)
        }
        if payload.C != contextKey {
            return nil, fmt.Errorf("invalid cursor context")
        }
        after = &payload.K
    }

    // Enforce limit bounds.
    limit := query.Limit
    if limit < 1 {
        limit = 20
    }
    if limit > maxLimit {
        return nil, ErrLimitExceeded
    }

    // Fetch limit+1 items using keyset pagination to detect whether more rows exist.
    var rows *sql.Rows
    var err error
    if after == nil {
        rows, err = r.db.QueryContext(ctx,
            `SELECT g.id, g.type, g.gkid, g.moved_on_datetime
             FROM geokrety.gk_geokrety_with_details AS g
             INNER JOIN geokrety.gk_moves AS m ON m.geokret = g.id
             WHERE m.author = $1
             ORDER BY g.moved_on_datetime DESC, g.id DESC
             LIMIT $2`,
            userID, limit+1,
        )
    } else {
        rows, err = r.db.QueryContext(ctx,
                        `SELECT g.id, g.type, g.gkid, g.moved_on_datetime
                         FROM geokrety.gk_geokrety_with_details AS g
                         INNER JOIN geokrety.gk_moves AS m ON m.geokret = g.id
                         WHERE m.author = $1
                             AND (g.moved_on_datetime, g.id) < ($2, $3)
                         ORDER BY g.moved_on_datetime DESC, g.id DESC
             LIMIT $4`,
                        userID, after.LastMoveAt, after.ID, limit+1,
        )
    }
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []Activity
    for rows.Next() {
        // Scan into Activity struct
        var a Activity
        if err := rows.Scan(&a.ID, &a.Type, &a.GKID, &a.LastMoveAt); err != nil {
            return nil, err
        }
        items = append(items, a)
    }

    // Detect if more pages exist
    hasMore := len(items) > limit
    if hasMore {
        items = items[:limit]  // Trim the +1 probe row
    }

    queryMs := time.Since(started).Milliseconds()

    // Build response
    var currentCursor *Cursor
    if after != nil {
        cc := EncodeCursor(1, *after, contextKey)
        currentCursor = &cc
    }
    var nextCursor *Cursor
    if hasMore {
        last := items[len(items)-1]
        nc := EncodeCursor(1, CursorKey{LastMoveAt: last.LastMoveAt, ID: int64(last.ID)}, contextKey)
        nextCursor = &nc
    }

    return &Page[Activity]{
        Data: items,
        Meta: Meta{
            Pagination: PaginationInfo{
                Cursor:     currentCursor,
                NextCursor: nextCursor,
                HasMore:    hasMore,
                Count:      len(items),
            },
            RequestedAt:    time.Now().UTC(),
            QueryMs:        queryMs,
            Sort:           []string{"last_move_at:desc", "id:desc"},
            SortableFields: []string{"last_move_at", "type", "id"},
        },
    }, nil
}
```

### 7.4 TypeScript Type Definitions

```typescript
/**
 * Opaque pagination cursor. Clients must treat as opaque.
 */
export type Cursor = string & { readonly __brand: 'Cursor' };

/**
 * Create a branded Cursor type (for type safety).
 */
export function cursor(c: string): Cursor {
    return c as Cursor;
}

/**
 * Pagination metadata for cursor-based pagination.
 */
export interface PaginationInfo {
    cursor: Cursor | null;
    nextCursor: Cursor | null;
    hasMore: boolean;
    count: number;
}

/**
 * Pagination metadata for offset-based pagination.
 */
export interface PaginationInfoOffset {
    offset: number;
    limit: number;
    totalItems?: number;  // Optional (expensive)
    totalPages?: number;
    hasMore: boolean;
    count: number;
}

/**
 * Metadata for a paginated response.
 */
export interface Meta {
    requestedAt: string;
    queryMs: number;
    pagination: PaginationInfo;
    sort?: string[];      // Current sort order (e.g., ["last_move_at:desc", "id:asc"])
    sortableFields?: string[];  // Allowed sort fields
}

/**
 * Metadata for an offset-based paginated response.
 */
export interface MetaOffset {
    requestedAt: string;
    queryMs: number;
    pagination: PaginationInfoOffset;
    sort?: string[];      // Current sort order
    sortableFields?: string[];  // Allowed sort fields
}

/**
 * Paginated response for cursor-based pagination.
 */
export interface Page<T> {
    data: T[];
    meta: Meta;
}

/**
 * Paginated response for offset-based pagination.
 */
export interface PageOffset<T> {
    data: T[];
    meta: MetaOffset;
}

/**
 * Query parameters for cursor-based pagination.
 */
export interface PaginationQuery {
    cursor?: Cursor;
    limit?: number;
}

/**
 * Query parameters for offset-based pagination.
 */
export interface PaginationQueryOffset {
    offset?: number;
    limit?: number;
    includeTotal?: boolean;
}

/**
 * Opaque cursor encoding/decoding (internal).
 */
export interface CursorPayload {
    V: number;  // Version
    K: {
        lastMoveAt: string;
        id: number;
    };
    C: string;  // Context fingerprint (user/sort/filter scope)
}

export function encodeCursor(version: number, lastMoveAt: string, id: number, context: string): Cursor {
    const payload: CursorPayload = {
        V: version,
        K: { lastMoveAt, id },
        C: context,
    };
    const encoded = btoa(JSON.stringify(payload));
    return cursor(encoded);
}

export function decodeCursor(c: Cursor): CursorPayload {
    try {
        const decoded = atob(c);
        return JSON.parse(decoded);
    } catch (e) {
        throw new Error('Invalid cursor');
    }
}
```

---

## 8. Implementation Patterns

### 8.1 Backend Pattern: Go

**File Structure:**
```
geokrety/
  pagination/
    cursor.go          # Cursor encoding/decoding
    page.go            # Page[T] and response types
    repository.go      # PaginationRepository interface
    errors.go          # Pagination-specific errors
    pagination_test.go # Unit tests
  handlers/
    geokrety_found.go  # Example: GET /api/v3/users/{id}/geokrety-found
    search.go          # Example: GET /api/v3/geokrety/search
```

**Integration in Handler:**
```go
// handlers/geokrety_found.go
func (h *Handler) GetUserFoundGeokrety(w http.ResponseWriter, r *http.Request) {
    started := time.Now()
    userID, err := strconv.Atoi(chi.URLParam(r, "id"))
    if err != nil {
        writePaginationErrorForRequest(w, r, http.StatusBadRequest, "INVALID_USER_ID", "user id must be numeric")
        return
    }

    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    cursorStr := r.URL.Query().Get("cursor")

    query := pagination.PaginationQuery{
        Cursor: pagination.Cursor(cursorStr),
        Limit:  limit,
    }

    page, err := h.activityRepo.GetActivityPaginated(r.Context(), userID, query)
    if err != nil {
        code := "INVALID_CURSOR"
        switch {
        case errors.Is(err, pagination.ErrCursorVersionMismatch):
            code = "CURSOR_VERSION_MISMATCH"
        case errors.Is(err, pagination.ErrLimitExceeded):
            code = "LIMIT_EXCEEDED"
        }

        writePaginationErrorForRequest(w, r, http.StatusBadRequest, code, err.Error())
        return
    }

    writePaginatedEnvelopeForRequest(w, r, http.StatusOK, page.Data, page.Meta, started)
}
```

**Repository alignment:** The live repository already centralizes HTTP responses in `api/internal/handlers/response.go` and routing in `api/internal/api/router.go`. Implementation should extend those shared helpers with `writePaginatedEnvelopeForRequest(...)` and `writePaginationErrorForRequest(...)` so cursor fields and structured errors reuse the same envelope conventions instead of introducing parallel response writers.

### 8.2 Frontend Pattern: Vue 3 Composition API

**Composable: `usePagination.ts`**
```typescript
import { ref, computed, Ref } from 'vue';
import type { Page, Cursor } from '@/types/pagination';

export interface UsePaginationOptions<T> {
    fetchFn: (cursor: Cursor | undefined) => Promise<Page<T>>;
    pageSize?: number;
}

export function usePagination<T>(options: UsePaginationOptions<T>) {
    const items: Ref<T[]> = ref([]);
    const currentCursor: Ref<Cursor | undefined> = ref(undefined);
    const nextCursor: Ref<Cursor | null | undefined> = ref(undefined);
    const isLoading = ref(false);
    const error: Ref<Error | null> = ref(null);

    const hasMore = computed(() => nextCursor.value !== undefined && nextCursor.value !== null);

    async function fetchNextPage() {
        if (isLoading.value || nextCursor.value === null) return;
        isLoading.value = true;
        error.value = null;

        try {
            const requestCursor = nextCursor.value === undefined ? undefined : nextCursor.value;
            const page = await options.fetchFn(requestCursor);
            items.value.push(...page.data);
            currentCursor.value = page.meta.pagination.cursor;
            nextCursor.value = page.meta.pagination.nextCursor;
        } catch (err) {
            error.value = err as Error;
        } finally {
            isLoading.value = false;
        }
    }

    function reset() {
        items.value = [];
        currentCursor.value = undefined;
        nextCursor.value = undefined;
        error.value = null;
    }

    async function retry() {
        reset();
        await fetchNextPage();
    }

    return {
        items: computed(() => items.value),
        hasMore,
        isLoading: computed(() => isLoading.value),
        error: computed(() => error.value),
        fetchNextPage,
        reset,
        retry,
    };
}
```

**Component: `InfiniteScrollFeed.vue`**
```vue
<script setup lang="ts">
import { onMounted } from 'vue';
import { usePagination } from '@/composables/usePagination';
import { useIntersectionObserver } from '@/composables/useIntersectionObserver';
import type { Activity } from '@/types';

const userId = 42;

const {
    items,
    hasMore,
    isLoading,
    error,
    fetchNextPage,
} = usePagination<Activity>({
    fetchFn: async (cursor) => {
        const params = new URLSearchParams({ limit: '20' });
        if (cursor) {
            params.set('cursor', cursor);
        }

        const response = await fetch(`/api/v3/users/${userId}/geokrety-found?${params.toString()}`, {
            headers: { Accept: 'application/json' },
            credentials: 'include',
        });
        if (!response.ok) {
            throw new Error(`Request failed with status ${response.status}`);
        }
        return await response.json() as Page<Activity>;
    },
});

// Intersection Observer: auto-fetch when sentinel reaches viewport
const { observerElement } = useIntersectionObserver(
    async () => {
        if (hasMore.value && !isLoading.value) {
            await fetchNextPage();
        }
    },
    { threshold: 0.5 }
);

onMounted(async () => {
    await fetchNextPage();
});
</script>

<template>
    <div class="feed">
        <div v-for="item in items" :key="item.id" class="activity-item">
            <span class="time">{{ new Date(item.lastMoveAt).toLocaleString() }}</span>
            <span class="event">{{ item.type }}: {{ item.gkid }}</span>
        </div>

        <div v-if="error" class="error">{{ error.message }}</div>
        <div v-if="isLoading" class="loading">Loading...</div>
        <div v-if="!hasMore && items.length > 0" class="end">No more items</div>

        <!-- Sentinel for intersection observer -->
        <div v-if="hasMore && items.length > 0" ref="observerElement" class="sentinel"></div>
    </div>
</template>

<style scoped>
.feed {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.activity-item {
    border: 1px solid #ddd;
    padding: 1rem;
    border-radius: 4px;
}

.time { font-size: 0.875rem; color: #666; }
.event { margin-left: 0.5rem; font-weight: 500; }

.error { color: red; padding: 1rem; background: #ffe0e0; border-radius: 4px; }
.loading { text-align: center; color: #666; padding: 1rem; }
.end { text-align: center; color: #999; padding: 1rem; }

.sentinel { height: 200px; }  /* Trigger fetch when visible */
</style>
```

### 8.3 Offset-Based Pattern

**Composable: `useOffsetPagination.ts`**
```typescript
export interface UseOffsetPaginationOptions<T> {
    fetchFn: (offset: number, limit: number) => Promise<PageOffset<T>>;
    pageSize?: number;
}

export function useOffsetPagination<T>(options: UseOffsetPaginationOptions<T>) {
    const items: Ref<T[]> = ref([]);
    const offset = ref(0);
    const limit = ref(options.pageSize ?? 20);
    const totalItems: Ref<number | undefined> = ref(undefined);
    const isLoading = ref(false);
    const error: Ref<Error | null> = ref(null);

    const totalPages = computed(() =>
        totalItems.value ? Math.ceil(totalItems.value / limit.value) : undefined
    );

    const currentPage = computed(() => Math.floor(offset.value / limit.value) + 1);

    async function fetchPage(pageOffset: number = 0) {
        if (isLoading.value) return;
        isLoading.value = true;
        error.value = null;

        try {
            const newOffset = pageOffset * limit.value;
            const page = await options.fetchFn(newOffset, limit.value);
            items.value = page.data;
            offset.value = page.meta.pagination.offset;
            totalItems.value = page.meta.pagination.totalItems;
        } catch (err) {
            error.value = err as Error;
        } finally {
            isLoading.value = false;
        }
    }

    async function nextPage() {
        if (!hasMore.value) return;
        await fetchPage(currentPage.value);
    }

    async function previousPage() {
        if (currentPage.value <= 1) return;
        await fetchPage(currentPage.value - 2);
    }

    const hasMore = computed(() => {
        if (totalItems.value === undefined) return true;  // Unknown
        return offset.value + limit.value < totalItems.value;
    });

    return {
        items: computed(() => items.value),
        currentPage,
        totalPages,
        totalItems: computed(() => totalItems.value),
        isLoading: computed(() => isLoading.value),
        error: computed(() => error.value),
        hasMore,
        fetchPage,
        nextPage,
        previousPage,
    };
}
```

---

## 9. Design Decisions & Rationale

### 9.1 Cursor Format: Base64-Encoded Versioned JSON

**Decision:** Cursors are base64-encoded JSON objects containing a version, the last visible stable-sort key, and a request-context fingerprint.

**Format Example:**
```
{"V":1,"K":{"lastMoveAt":"2026-03-22T10:30:00Z","id":42},"C":"user:42|sort:last_move_at:desc,id:desc|filter:all"}
    →  base64 encode  →  "..."
```

**Rationale:**

- ✅ **Versioning support**: Version field allows cursor format to evolve
- ✅ **Simplicity**: Easy to implement, debug, and extend
- ✅ **Language-agnostic**: JSON is universal
- ✅ **Deterministic**: Same inputs always produce the same cursor for the same keyset position and request context
- ✅ **Safe to log**: Base64 encoding is not encryption (not secret)

**Alternatives Considered:**

- Opaque UUIDs (no versioning, harder to debug, requires server-side storage)
- Encrypted blobs (overly complex, adds crypto overhead, no benefit for non-secret data)
- Offset-only tokens (simple but do not provide the large-dataset or mutation-resilience properties required for cursor-based pagination)
- Direct JSON (not base64, allows client parsing, breaks opacity contract)

**Anti-Pattern:** Clients MUST NOT parse cursor structure. Server reserves right to change format in next version.

### 9.2 Cursor Versioning Strategy

**Decision:** Cursors include a `V` (version) field and a `C` (context) field; server checks both during decode.

**Strategy:**
1. Current version is 1
2. When cursor format changes incompatibly (for example, keyset shape or context rules change), bump version to 2
3. Server currently accepts only v1 and rejects unsupported versions with `CURSOR_VERSION_MISMATCH`
4. When a future v2 is introduced, the server may temporarily support both versions during migration
5. Until a migration plan exists, this specification standardizes only v1 behavior

**Example Version Migration:**
```go
func (c Cursor) Decode() (CursorPayload, error) {
    decoded, _ := base64.StdEncoding.DecodeString(string(c))
    var payload CursorPayload
    json.Unmarshal(decoded, &payload)

    if payload.V != 1 {
        return CursorPayload{}, ErrCursorVersionMismatch
    }
    return payload, nil
}
```

### 9.3 Total Count Inclusion Decision

**Decision:** `totalItems` is optional in offset-based responses and is omitted by default unless explicitly requested or cheaply available.

**Rationale:**

- ❌ Computing exact count on large tables is expensive (full table scan)
- ✅ Clients don't always need "Page X of Y" (many modern UIs just have "Load More")
- ✅ Reduces database load

**Enablement Strategy:**
```go
type PaginationQueryOffset struct {
    Offset       int
    Limit        int
    IncludeTotal bool  // Client request: "I want totalItems"
}

// Server defaults: totalItems omitted unless includeTotal=true
// or the endpoint can provide it cheaply enough for the requested query.
```

### 9.4 Forward-Only vs. Bidirectional Pagination

**Decision:** Cursor-based pagination is forward-only (next cursor only). Backward navigation is not supported.

**Rationale:**

- ✅ Simpler state model (no previous cursor)
- ✅ Natural for infinite scroll (users scroll down, not up)
- ✅ Reflects append-only data semantics (new entries always at top)
- ✅ Prevents cursor-based attacks (can't reverse-enumerate)

**Offset-based pagination** supports backward navigation naturally (`previousPage()`, `nextPage()`).

**Anti-Pattern:** Do NOT implement backward cursor pagination. Clients needing to re-visit earlier items should reset state and restart from first page.

### 9.5 Interaction Between Sorting and Pagination

**Requirement:** Sort order MUST be stable and deterministic.

**Definition of Stable Sort:**
```sql
-- ✅ GOOD: Stable sort
ORDER BY created_at DESC, id DESC

-- ❌ BAD: Non-deterministic sort (can vary between queries)
ORDER BY RANDOM()
ORDER BY created_at DESC  -- If multiple rows have same created_at, order is undefined
ORDER BY updated_at DESC  -- If updated_at changes during pagination, cursor breaks
```

**Guideline:** Always include a unique tiebreaker (usually `id`) as the final sort key to ensure deterministic ordering when primary sort has ties.

**API Contract:** Document the sort order explicitly for every paginated endpoint.

```
GET /api/v3/users/{id}/geokrety-found
- Sorted by: last_move_at DESC, id DESC (reverse chronological by last move time)
- Cursor represents position in this sorted order
- If sort changes, the client must reset pagination and request the first page for the new sort
```

### 9.6 Filter Change Handling During Pagination

**Scenario:** User is paginating through activity feed, then adds a filter (e.g., "show only geocaches"). What happens to the cursor?

**Decision:** When filters change, client must reset pagination state.

**Implementation Pattern:**

```typescript
// ❌ WRONG: Carrying old cursor to new filter
const cursor = // ... from previous filter
fetch(`/api/v3/users/${userId}/geokrety-found?${new URLSearchParams({ cursor, filter: newFilter })}`)

// ✅ RIGHT: Reset cursor when filter changes
const filter = ref('all');

watch(() => filter.value, () => {
    // Filter changed; reset pagination
    pagination.reset();
});

// Client explicitly fetches first page with new filter
fetch(`/api/v3/users/${userId}/geokrety-found?${new URLSearchParams({ filter: filter.value })}`)
```

**Why:** Cursors encode a position in a specific result set. Changing filters creates a different result set, invalidating the cursor. The server rejects old cursors when the context fingerprint embedded in the cursor no longer matches the requested user, sort, or filter context.

**Server Detection:** Server includes a context fingerprint in the cursor and compares it to the current request context before executing the next-page query.

---

## 10. Security Considerations

### 10.1 Pagination Security Model

**Core Principle:** Pagination data (cursors, offsets, page counts, etc.) **is inherently public information**. Anyone can make multiple requests with different cursors/offsets to traverse paginated data. This is by design and not a security threat—pagination enables public APIs.

**Real Security Concerns:**
1. **Authorization**: Only return data the authenticated user is authorized to see
   - Enforce at the repository layer: `db.query("... WHERE user_id = $1", currentUserId)`
    - URL structure should reflect the data owner: `/api/v3/users/{id}/geokrety-found`
   - Don't rely on "hidden cursors" to prevent unauthorized access

2. **SQL Injection**: Cursor-derived values entered into SQL queries can expose data
   - Always use parameterized queries (prepared statements)
- Never concatenate decoded cursor values into SQL strings
   - Validate cursor format before decoding (see below)

3. **Rate Limiting**: Prevent abuse by limiting pagination requests per second/user
   - On public endpoints, rate-limit aggressively
   - On paginated list endpoints, cap max page size (e.g., max 100 items/page)

4. **Data Leakage Through Metadata**:
   - Total counts can reveal dataset sizes (acceptable tradeoff for UX)
   - If counts are sensitive, optionally cap them (`total > 1000 = "1000+"`)
   - Sort order stability: Don't change sort between requests (breaks cursor logic)

**Anti-Pattern:** Encrypting or signing cursors to "hide" pagination structure adds complexity without benefit. Example:
```go
// ❌ Unnecessary: Encrypted cursor
cursor = aes.Encrypt(json.Marshal(Offset: 40))

// ✅ Sufficient: Simple opaque base64 + authorization checks
cursor = base64.Encode(json.Marshal(Offset: 40))
// Real security: Check user owns the data via URL path / auth token
```

**Summary:** Pagination security is authorization + SQL safety, not cursor obfuscation.

### 10.2 SQL Injection Prevention

**Pattern:** Use parameterized queries (prepared statements); never concatenate cursor-derived values into SQL.

```go
// ✅ SAFE: Parameterized query
db.QueryContext(ctx,
    "SELECT * FROM geokrety.gk_moves WHERE author = $1 ORDER BY ... LIMIT $2 OFFSET $3",
    userID,    // Parameterized
    limit,     // Parameterized
    offset,    // From decoded cursor, but still parameterized
)

// ❌ UNSAFE: String concatenation
query := fmt.Sprintf("SELECT * FROM activity LIMIT %d OFFSET %d", limit, offset)
// If offset comes from untrusted cursor, risk of injection
```

### 10.3 Server-Side Limits

**Requirement:** Server enforces reasonable limits to prevent resource exhaustion.

```go
const (
    DefaultLimit = 20
    MaxLimit     = 100
    MaxOffset    = 100_000  // Endpoint-specific cap for offset-based pagination
)

func (h *Handler) getActivity(w http.ResponseWriter, r *http.Request) {
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

    // Enforce bounds
    if limit < 1 { limit = DefaultLimit }
    if limit > MaxLimit {
        http.Error(w, "LIMIT_EXCEEDED", http.StatusBadRequest)
        return
    }
}
```

**Decision:** Default missing or too-small limits to `DefaultLimit`, but reject limits above `MaxLimit` with `LIMIT_EXCEEDED`.

---

## 11. Testing Strategy

### 11.1 Unit Tests (Go)

**Test Cursor Encoding/Decoding:**
```go
func TestEncodeCursor(t *testing.T) {
    c := EncodeCursor(1, CursorKey{LastMoveAt: ts, ID: 40}, "user:1|sort:last_move_at:desc,id:desc|filter:all")
    payload, err := c.Decode()
    assert.NoError(t, err)
    assert.Equal(t, 1, payload.V)
    assert.Equal(t, int64(40), payload.K.ID)
}

func TestDecodeCursor_InvalidFormat(t *testing.T) {
    _, err := Cursor("not_base64").Decode()
    assert.Equal(t, ErrInvalidCursor, err)
}

func TestDecodeCursor_MalformedJSON(t *testing.T) {
    malformed := base64.StdEncoding.EncodeToString([]byte("{invalid"))
    _, err := Cursor(malformed).Decode()
    assert.Equal(t, ErrInvalidCursor, err)
}

func TestDecodeCursor_UnsupportedVersion(t *testing.T) {
    // Cursor with version 99
    payload := `{"V":99,"K":{"lastMoveAt":"2026-03-22T10:30:00Z","id":40},"C":"user:1|sort:last_move_at:desc,id:desc|filter:all"}`
    c := Cursor(base64.StdEncoding.EncodeToString([]byte(payload)))
    _, err := c.Decode()
    assert.Equal(t, ErrCursorVersionMismatch, err)
}

func TestEncodeCursor_EdgeCases(t *testing.T) {
    tests := []struct {
        name    string
        version int
        offset  int64
    }{
        {"current version", 1, 40},
        {"future version", 99, 100},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            c := EncodeCursor(tc.version, CursorKey{LastMoveAt: ts, ID: tc.offset}, "user:1|sort:last_move_at:desc,id:desc|filter:all")
            payload, err := c.Decode()
            if tc.version == 1 {
                assert.NoError(t, err)
                assert.Equal(t, tc.version, payload.V)
                assert.Equal(t, tc.offset, payload.K.ID)
            } else {
                assert.Equal(t, ErrCursorVersionMismatch, err)
            }
        })
    }
}
```

**Test Pagination Logic:**
```go
func TestActivityRepo_GetActivityPaginated(t *testing.T) {
    db := setupTestDB(t)
    repo := NewActivityRepository(db)

    // Insert test data: 50 activity items
    for i := 0; i < 50; i++ {
        insertActivity(t, db, Activity{ID: i, Type: "geocached"})
    }

    t.Run("first page", func(t *testing.T) {
        page, err := repo.GetActivityPaginated(context.Background(), 1, PaginationQuery{
            Limit: 20,
        })
        assert.NoError(t, err)
        assert.Equal(t, 20, len(page.Data))
        assert.True(t, page.Meta.Pagination.HasMore)
        assert.NotNil(t, page.Meta.Pagination.NextCursor)
    })

    t.Run("middle page", func(t *testing.T) {
        // Fetch first page
        page1, _ := repo.GetActivityPaginated(context.Background(), 1, PaginationQuery{Limit: 20})

        // Fetch second page using nextCursor
        page2, err := repo.GetActivityPaginated(context.Background(), 1, PaginationQuery{
            Cursor: *page1.Meta.Pagination.NextCursor,
            Limit:  20,
        })
        assert.NoError(t, err)
        assert.Equal(t, 20, len(page2.Data))
        assert.True(t, page2.Meta.Pagination.HasMore)
    })

    t.Run("last page", func(t *testing.T) {
        // Fetch pages until no more
        var lastPage *Page[Activity]
        var nextCursor Cursor
        for {
            page, _ := repo.GetActivityPaginated(context.Background(), 1, PaginationQuery{
                Cursor: nextCursor,
                Limit:  20,
            })
            lastPage = page
            if !page.Meta.Pagination.HasMore {
                break
            }
            nextCursor = *page.Meta.Pagination.NextCursor
        }

        assert.Equal(t, 10, len(lastPage.Data))  // 50 % 20 = 10
        assert.Nil(t, lastPage.Meta.Pagination.NextCursor)
        assert.False(t, lastPage.Meta.Pagination.HasMore)
    })

    t.Run("empty result set", func(t *testing.T) {
        page, err := repo.GetActivityPaginated(context.Background(), 99, PaginationQuery{Limit: 20})
        assert.NoError(t, err)
        assert.Equal(t, 0, len(page.Data))
        assert.Nil(t, page.Meta.Pagination.NextCursor)
        assert.False(t, page.Meta.Pagination.HasMore)
    })
}
```

### 11.2 Integration Tests (TypeScript/Vue 3)

```typescript
import { describe, it, expect, vi } from 'vitest';
import { usePagination } from '@/composables/usePagination';

describe('usePagination', () => {
    it('fetches and accumulates items', async () => {
        // Mock API
        const mockAPI = {
            callCount: 0,
            async fetch(cursor?: string) {
                const startIndex = cursor
                    ? mockData.findIndex((item) => item.id === decodeCursor(cursor).K.id) + 1
                    : 0;
                const items = mockData.slice(startIndex, startIndex + 3);  // 3 items per page
                const lastItem = items[items.length - 1];
                return {
                    data: items,
                    meta: {
                        pagination: {
                            cursor: cursor ?? null,
                            nextCursor: startIndex + 3 < mockData.length && lastItem
                                ? encodeCursor(1, lastItem.lastMoveAt, lastItem.id, 'user:1|sort:last_move_at:desc,id:desc|filter:all')
                                : null,
                            hasMore: startIndex + 3 < mockData.length,
                            count: items.length,
                        },
                        requestedAt: '2026-03-22T15:00:01Z',
                        queryMs: 4,
                        sort: ['last_move_at:desc', 'id:desc'],
                        sortableFields: ['last_move_at', 'id'],
                    },
                };
            },
        };

        const { items, hasMore, fetchNextPage } = usePagination({
            fetchFn: (cursor) => mockAPI.fetch(cursor),
        });

        // Initial state
        expect(items.value).toEqual([]);
        expect(hasMore.value).toBe(false);

        // Fetch first page
        await fetchNextPage();
        expect(items.value).toHaveLength(3);
        expect(hasMore.value).toBe(true);

        // Fetch second page
        await fetchNextPage();
        expect(items.value).toHaveLength(6);
        expect(hasMore.value).toBe(mockData.length > 6);
    });

    it('handles errors gracefully', async () => {
        const mockAPI = {
            async fetch() {
                throw new Error('Network error');
            },
        };

        const { error, fetchNextPage } = usePagination({
            fetchFn: (cursor) => mockAPI.fetch(cursor),
        });

        await fetchNextPage();
        expect(error.value?.message).toBe('Network error');
    });
});
```

### 11.3 Example Scenarios

**Scenario 1: User Activity Feed (10K records)**

- User loads page: expect 20 items fetched
- User scrolls 5 times: expect total 100 items in state
- Performance: cursor encoding/decoding < 1ms

**Scenario 2: Search Results (847 items)**

- User searches for "hiking cache", gets 847 results
- Pagination shows "Item 1-20 of 847"
- User can jump to page 5 (offset 80-99)
- User can jump backward to page 3 (offset 40-59)

**Scenario 3: Comment Thread (reverse chronological)**

- Comments are ordered `created_at DESC, id DESC`
- User loads first page (most recent comments)
- User scrolls for older comments
- Cursor handles concurrent new comments (they appear in previous pages, not current)

---

## 12. Implementation Readiness & Operational Guidance

### 12.1 Feasibility Verdict

This specification is **fully feasible with mitigations** and is considered ready for implementation once the operational prerequisites below are accepted by backend, frontend, and database owners.

**Implementation Summary:**

- Go implementation is idiomatic with Go 1.18+ generics and standard `encoding/json` / `encoding/base64`
- TypeScript and Vue 3 implementation is idiomatic with generic interfaces and composition functions
- Performance targets are realistic if stable-sort indexes are created before rollout
- No schema redesign is required; only index additions and endpoint-level integration work are expected

### 12.2 Critical Success Factors

**Must do before rollout:**

- Create database indexes for the stable sort orders used by high-volume paginated endpoints
- Implement cursor versioning from day 1, even if only v1 is initially emitted
- Test malformed base64, malformed JSON payloads, unsupported versions, and out-of-range offsets

**Should do during implementation:**

- Cap offset-based pagination to prevent deep pagination abuse on large tables
- Explicitly reset pagination state whenever sort or filter changes
- Add guidance for virtualized rendering if a single infinite-scroll view can accumulate more than 10,000 items

**Nice to have after first rollout:**

- Additional query-plan tuning for endpoints with specialized filters or custom ranking functions
- Analytics on pagination depth and user behavior
- Cursor expiration only if a later security or product requirement demands it

### 12.3 Database Prerequisites

At minimum, the following indexes should exist before rollout to large or high-traffic endpoints:

```sql
CREATE INDEX IF NOT EXISTS idx_gk_moves_author_moved_on_id
    ON geokrety.gk_moves(author, moved_on_datetime DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_gk_moves_moved_on_id
    ON geokrety.gk_moves(moved_on_datetime DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_gk_geokrety_with_details_moved_on_id
    ON geokrety.gk_geokrety_with_details(moved_on_datetime DESC, id DESC);
```

The database team should validate query plans with `EXPLAIN ANALYZE` before production deployment and confirm that stable-sort endpoints use index scans rather than sequential scans. On large production tables, index creation should use production-safe rollout procedures such as concurrent index builds where supported by the migration workflow.

For every offset-based endpoint, the implementation must document which allowed sort combinations are backed by indexes and which, if any, are intentionally limited to small result sets or approximate totals.

### 12.4 Endpoint Selection Guidance

**Use cursor-based pagination for:**

- `/api/v3/users/{id}/geokrety-found`
- `/api/v3/geokrety/{gkid}/moves`
- Notification feeds, comment threads, or any append-only event stream

**Use offset-based pagination for:**

- `/api/v3/geokrety/search`
- Catalog-style lists where users expect page numbers or direct page jumps
- Small and stable result sets where `totalItems` is meaningful and inexpensive enough to compute

### 12.5 Definition of Done

**Backend done when:**

- `pagination/` package exists with documented exported types and helpers
- All examples and real handlers return `meta.pagination`, not top-level pagination fields
- Unit and integration tests cover invalid cursor, empty page, last page, and sort/filter reset scenarios
- OpenAPI docs include sort metadata and concrete paginated examples

**Frontend done when:**

- Cursor-based and offset-based composables are fully typed with zero `any`
- State reset on sort/filter changes is implemented and tested
- Infinite scroll prevents overlapping fetches and surfaces recoverable errors cleanly
- Long-list behavior is profiled and a virtualization strategy is documented if needed

**Database and operations done when:**

- Required indexes exist
- Slow query monitoring is in place for paginated endpoints
- Baseline query timings are recorded and reviewed against the performance goals in this document

### 12.7 Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Missing stable-sort indexes | Medium | High | Create indexes before rollout; verify with `EXPLAIN ANALYZE` |
| Slow deep OFFSET queries | Medium | Medium | Enforce `MaxOffset`, prefer cursor-based pagination for large datasets |
| Browser memory growth on long infinite scrolls | Low | Medium | Add virtualization guidance and monitor long-list behavior |
| Cursor parsing bugs | Low | Medium | Add exhaustive invalid-input tests and explicit error types |
| Sort/filter changes reusing stale cursors | Medium | Low | Reset pagination state client-side and reject stale cursor contexts server-side |

---

## 13. Implementation Checklist

### Backend Checklist

- [ ] Create `pagination/cursor.go`, `pagination/page.go`, and `pagination/errors.go`
- [ ] Add GoDoc comments to all exported pagination types and functions
- [ ] Ensure all handlers return `meta.pagination`
- [ ] Add OpenAPI examples for at least one cursor-based and one offset-based endpoint
- [ ] Add unit and integration tests covering all documented error cases

### Frontend Checklist

- [ ] Create `types/pagination.ts`
- [ ] Implement `usePagination.ts` and `useOffsetPagination.ts`
- [ ] Add component examples using infinite scroll and classic page navigation
- [ ] Reset pagination state on sort and filter changes
- [ ] Test memory behavior and loading guards on long scroll sessions

### Database & Operations Checklist

- [ ] Create and apply index migration(s)
- [ ] Run `EXPLAIN ANALYZE` on representative paginated queries
- [ ] Enable slow-query monitoring for pagination endpoints
- [ ] Record p95 latency targets and validate them after rollout

---

## 14. Acceptance Criteria

- **AC-001**: Cursor format documents version field; server handles v1 and emits v1 (ready for v2 later without breaking changes)

- **AC-002**: Offset-based pagination defaults missing or too-small limits to 20 and rejects limits above 100 with `LIMIT_EXCEEDED`

- **AC-003**: Invalid cursors (malformed or context-mismatched) produce `INVALID_CURSOR` error with HTTP 400

- **AC-004**: Vue 3 composable `usePagination()` handles infinity scroll: auto-fetches when `fetchNextPage()` is called, accumulates items, manages state transparently

- **AC-005**: This specification contains enough backend, frontend, database, and rollout guidance to implement at least one cursor-based endpoint and one offset-based endpoint without consulting additional project documents

- **AC-006**: All TypeScript code is fully typed; zero `any` types. Go code uses generics appropriately.

- **AC-007**: Every exported function, type, method has JSDoc (TS) or GoDoc (Go) comments explaining purpose, parameters, returns, and usage examples

- **AC-008**: Unit tests cover >90% of pagination logic; integration tests validate cursor handling, edge cases (empty, single page, exact page boundary), and error scenarios

- **AC-009**: A benchmark exists showing cursor encode/decode overhead is negligible relative to query time, with a target of < 1ms per operation on reference development hardware

- **AC-010**: Paginated endpoints reject unauthorized cursor manipulation (user cannot use cursor from one user's activity to access another user's activity)

- **AC-011**: Response envelope wraps pagination in `meta` field (not top-level); all endpoints follow this structure

- **AC-012**: Sorting specification documented with API contract examples; `sort` parameter format and allowed fields validated

- **AC-013**: Sort changes reset pagination cursor/offset to first page; server detects filter/sort changes and rejects stale cursors with clear error messages

- **AC-014**: Response metadata includes current sort order (`sort: ["field:order"]`) and sortable fields list (`sortableFields: [...]`) for client discovery

- **AC-015**: All paginated endpoints document allowed sort fields, default sort order, and sort constraints (max 3 columns) in OpenAPI spec using `x-sortable-fields` and `x-default-sort` extensions

- **AC-016**: Required stable-sort indexes are created and verified with representative query plans before production rollout

- **AC-017**: Offset-based pagination enforces a documented maximum offset to avoid unbounded deep scans on large tables

- **AC-018**: The specification alone is sufficient for backend, frontend, and database teams to start implementation without consulting `QUICK_START.md` or `FEASIBILITY_REVIEW.md`

- **AC-019**: Unsupported cursor versions produce `CURSOR_VERSION_MISMATCH` with HTTP 400

---

## 15. Validation Criteria

- ✅ Specification document is complete, self-contained, and requires no external references
- ✅ All design decisions (cursor format, versioning, totals, forward-only behavior, sorting, and filter resets) are explicitly stated with rationale
- ✅ Type definitions are generic (`Pagination<T>`, `Page<T>`) and work with any type
- ✅ Example endpoints demonstrate both cursor-based and offset-based pagination
- ✅ Example Vue 3 components show realistic usage (infinite scroll, pagination controls)
- ✅ Security considerations address cursor enumeration and SQL injection
- ✅ Testing strategy covers happy path, edge cases, and error scenarios
- ✅ API contract includes examples with exact JSON structures
- ✅ Documentation is clear enough for implementation without further clarification
- ✅ Implementation rollout, risk mitigations, and operational prerequisites are captured in this file

---

## 16. Related Specifications & Further Reading

- [REST API Design Best Practices](https://www.rfc-editor.org/rfc/rfc7231)
- [GraphQL Cursor Pagination](https://relay.dev/docs/guides/graphql-server/) (for reference only; this spec is REST-focused)
- [OWASP: Preventing Enumeration Attacks](https://owasp.org/www-community/attacks/Enumeration)
