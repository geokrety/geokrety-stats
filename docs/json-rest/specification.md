---
title: GeoKrety JSON REST API Envelope
version: 2.0
date_created: 2026-03-28
last_updated: 2026-03-29
owner: GeoKrety Development Team
tags: [architecture, api, json-rest, envelope, pagination]
---

# Specification: GeoKrety JSON REST API Envelope

## 1. Introduction

This specification defines the canonical JSON wire format for GeoKrety JSON endpoints. It standardizes the top-level response envelope, the resource-object model used inside `data`, and the rules for links, naming, relationship linkage, side-loaded resources, and GeoKrety-specific identifiers.

Pagination, filters, and sort behavior are documented in [../pagination/specification.md](../pagination/specification.md). This page defines the shared JSON REST envelope used by both single-resource and collection responses.

## 2. Purpose & Scope

### Purpose

- Make every GeoKrety JSON response predictable at the top level
- Separate generic JSON REST rules from GeoKrety-specific API rules
- Give backend, frontend, and documentation authors a single canonical envelope
- Ensure pagination, filtering, sorting, and relationship links all fit the same contract

### In Scope

- Successful JSON response envelopes
- Resource objects in collections and single-resource responses
- Top-level and resource-level links
- Relationship linkage and side-loaded related resources
- Canonical JSON naming conventions
- GeoKrety-specific identifier policy for GeoKret resources

### Out of Scope

- XML payload design
- GraphQL payloads
- Endpoint-specific attribute catalogs
- Full pagination mechanics beyond the shared envelope
- Detailed error taxonomies beyond the shared error shape

## 3. Canonical Conventions

### 3.1 Top-Level Envelope

Every successful JSON response uses these top-level members:

| Member | Type | Required | Meaning |
|--------|------|----------|---------|
| `data` | array, object, or `null` | Yes | The primary resource payload |
| `included` | array | No | Side-loaded related resources referenced by `relationships` |
| `meta` | object | Yes | Response metadata and request context |
| `links` | object | Yes | Canonical URLs and pagination navigation |

All successful responses return top-level `links`.

Universal `meta` rule:

- every successful JSON response includes `meta`
- `meta.execution_time_ms` is the shared baseline field across success responses
- collection endpoints add nested `meta.page`, `meta.query`, and `meta.capabilities`
- endpoint-specific metadata may be added, but it must not contradict the shared envelope or pagination contract

### 3.2 Naming Policy

The canonical JSON REST contract uses `snake_case` for:

- attribute names
- metadata keys
- query parameters documented in examples
- link-preserved filter and sort names

Existing camelCase payloads in legacy handlers or older documentation are non-normative. New documentation and newly specified JSON endpoints must use `snake_case` unless a page explicitly documents a legacy exception.

### 3.3 Identifier Policy

- Every resource object exposes `id` as a JSON string
- Every resource object exposes a stable `type` token such as `user`, `geokrety`, `move`, `picture`, `country`, `waypoint`, `lover`, `watcher`, or `finder`
- For GeoKret resources, `id` must be the public GKID string such as `GK0001`
- Internal database identifiers are not the canonical GeoKret identifier exposed to clients

### 3.4 Link Policy

- `links.self` represents the canonical URL for the current response or resource
- Collection links may include `first`, `prev`, `next`, and `last`
- Relationship objects may expose `links.related` and other relationship-specific URLs
- Clients should follow server-provided links instead of rebuilding pagination URLs when a link is already present

### 3.5 Side-Loading Policy

- `included` contains resource objects referenced from `relationships`
- `included` is deduplicated by `(type, id)` within one response
- related attributes belong in `included`, not inside `relationships.*.data`
- if a response does not need side-loaded related resources, `included` may be omitted

## 4. Resource Object Model

Each item in `data` must use the following structure:

```json
{
  "id": "841337",
  "type": "move",
  "attributes": {
    "date": "2026-03-28T10:00:00Z",
    "operation": "grab",
    "comment": "Taken during an event."
  },
  "relationships": {
    "geokret": {
      "data": {
        "type": "geokrety",
        "id": "GK00FF"
      },
      "links": {
        "related": "/api/v3/geokrety/GK00FF"
      }
    },
    "user": {
      "data": {
        "type": "user",
        "id": "42"
      },
      "links": {
        "related": "/api/v3/users/42"
      }
    }
  },
  "links": {
    "self": "/api/v3/moves/841337"
  }
}
```

### Rules

- `attributes` contains fields that are not identifiers, relationship linkage, or links
- `relationships` is optional and used only when the response needs to expose related-resource linkage or related URLs
- `links` is optional on a resource object, but `links.self` is recommended whenever a stable canonical resource URL exists
- Relationship linkage objects contain only `type` and `id`
- Relationship attributes do not appear inline inside `relationships.*.data`
- Collection-style related data belongs in top-level collections or dedicated endpoints, not in relationship arrays embedded inside another resource

## 5. Response Families

### 5.1 Single-Resource Response

```json
{
  "data": {
    "id": "GK00FF",
    "type": "geokrety",
    "attributes": {
      "name": "Hidden Treasure",
      "type_label": "traditional",
      "missing": false,
      "last_move_at": "2026-03-28T10:00:00Z"
    },
    "relationships": {
      "owner": {
        "data": {
          "type": "user",
          "id": "42"
        },
        "links": {
          "related": "/api/v3/users/42"
        }
      }
    },
    "links": {
      "self": "/api/v3/geokrety/GK00FF"
    }
  },
  "included": [
    {
      "id": "42",
      "type": "user",
      "attributes": {
        "username": "alice"
      },
      "links": {
        "self": "/api/v3/users/42"
      }
    }
  ],
  "meta": {
    "execution_time_ms": 7
  },
  "links": {
    "self": "/api/v3/geokrety/GK00FF"
  }
}
```

If the resource does not exist, the endpoint should return an error response instead of a successful envelope with fake data.

### 5.2 Collection Response

The envelope stays the same for paginated and non-paginated collections. Only the contents of `meta` and `links` change.

```json
{
  "data": [
    {
      "id": "841337",
      "type": "move",
      "attributes": {
        "date": "2026-03-28T10:00:00Z",
        "operation": "grab"
      },
      "relationships": {
        "geokret": {
          "data": {
            "type": "geokrety",
            "id": "GK00FF"
          },
          "links": {
            "related": "/api/v3/geokrety/GK00FF"
          }
        },
        "user": {
          "data": {
            "type": "user",
            "id": "42"
          },
          "links": {
            "related": "/api/v3/users/42"
          }
        }
      },
      "links": {
        "self": "/api/v3/moves/841337"
      }
    }
  ],
  "included": [
    {
      "id": "GK00FF",
      "type": "geokrety",
      "attributes": {
        "name": "Hidden Treasure"
      },
      "links": {
        "self": "/api/v3/geokrety/GK00FF"
      }
    },
    {
      "id": "42",
      "type": "user",
      "attributes": {
        "username": "alice"
      },
      "links": {
        "self": "/api/v3/users/42"
      }
    }
  ],
  "meta": {
    "execution_time_ms": 11,
    "page": {
      "type": "cursor",
      "limit": 20,
      "has_more": true
    },
    "query": {
      "filters": {
        "geokret": "GK00FF"
      },
      "sort": "-date"
    },
    "capabilities": {
      "filters": {
        "geokret": {
          "type": "string"
        },
        "user": {
          "type": "integer"
        }
      },
      "sorts": ["date", "-date", "id", "-id"]
    }
  },
  "links": {
    "self": "/api/v3/moves?limit=20&filter[geokret]=GK00FF&sort=-date",
    "next": "/api/v3/moves?limit=20&cursor=eyJkYXRlIjoiMjAyNi0wMy0yOFQwOTowMDowMFoiLCJpZCI6ODQxMzM2fQ==&filter[geokret]=GK00FF&sort=-date"
  }
}
```

### 5.3 Paginated Collections

Paginated collections still use top-level `data`, optional `included`, `meta`, and `links`. The pagination mode only changes `meta.page` and which navigation links are present.

- Page-based collections use `meta.page.type = "page"` with `number`, `size`, `total_items`, and `total_pages`
- Cursor-based collections use `meta.page.type = "cursor"` with `limit`, `has_more`, and `links.next`
- Detailed rules live in [../pagination/specification.md](../pagination/specification.md)

## 6. GeoKrety-Specific Overlays

### 6.1 GeoKret Identifier Rules

GeoKrety accepts three request forms for GeoKret identifiers:

- numeric decimal such as `255`
- bare hexadecimal such as `00FF`
- GK-prefixed hexadecimal such as `GK00FF`

These are request-parsing rules. In JSON responses, the canonical GeoKret resource identifier is always the public GKID string.

Disambiguation rule:

- bare digit-only values without a leading zero are parsed as decimal
- zero-padded digit-only values are parsed as hexadecimal
- parsing is case-insensitive for hexadecimal input

### 6.2 JSON and XML Boundary

This specification governs JSON responses only. GeoKrety may continue to expose XML payloads, but XML does not redefine the JSON REST contract described here.

### 6.3 Legacy Payloads

Legacy JSON payloads that flatten domain DTOs directly into `data` or use camelCase keys remain implementation history, not the normative contract. Any page documenting a legacy payload must label it explicitly instead of presenting it as the canonical response model.

Migration rule:

- new endpoints follow this contract by default
- migrated endpoints replace older examples with this contract
- non-migrated legacy endpoints must document the exception in OpenAPI and endpoint notes until the migration is complete

## 7. Error Responses

GeoKrety keeps the existing structured error family separate from the success envelope:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "GeoKret GK00FF was not found."
  },
  "timestamp": "2026-03-28T11:20:00Z"
}
```

Rules:

- Successful responses use `data`, optional `included`, `meta`, and `links`
- Error responses use `error` and `timestamp`
- Documentation must not mix the success and error envelopes in a single example

## 8. Pagination and Sorting Integration

- Collection pagination is defined in [../pagination/specification.md](../pagination/specification.md)
- Effective filters belong in `meta.query.filters`
- Effective sort belongs in `meta.query.sort`
- Advertised filter and sort capabilities belong in `meta.capabilities`
- Sort and filter state that affects page identity must be preserved in `links.self` and any pagination links
- Clients should treat top-level pagination links as the source of truth for navigation
- Relationship links and pagination links serve different purposes and should not be conflated

## 9. Documentation and OpenAPI Rules

- Canonical documentation examples must be valid JSON
- Collection examples must show top-level `links`
- Resource examples must use `id`, `type`, and `attributes`
- OpenAPI descriptions and examples should describe the same top-level envelope used here
- Until the OpenAPI source tree is fully migrated, non-migrated endpoints must keep their legacy schema fragments explicit instead of implying they already match this contract
- Endpoint pages may add domain-specific metadata, but they must not contradict the envelope defined here

## 10. Acceptance Criteria

- Every new JSON collection example in the docs uses top-level `data`, optional `included`, `meta`, and `links`
- Every canonical resource example uses `id`, `type`, `attributes`, and optional `relationships` or `links`
- Relationship linkage examples expose only `id` and `type`
- The JSON naming policy is explicit and no longer ambiguous between camelCase and snake_case
- GeoKret responses use public GKIDs as canonical client-facing identifiers
- The pagination page can specialize page-based and cursor-based behavior without redefining the shared envelope
- OpenAPI wording can reference this contract without inventing additional response-shape rules
