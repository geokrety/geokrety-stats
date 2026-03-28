---
title: GeoKrety JSON REST API Envelope
version: 1.0
date_created: 2026-03-28
last_updated: 2026-03-28
owner: GeoKrety Development Team
tags: [architecture, api, json-rest, envelope, pagination]
---

# Specification: GeoKrety JSON REST API Envelope

## 1. Introduction

This specification defines the canonical JSON wire format for GeoKrety JSON endpoints. It standardizes the top-level response envelope, the resource-object model used inside `data`, and the rules for links, naming, and GeoKrety-specific identifiers.

Pagination is part of this contract, but its operational details are documented separately in [../pagination/specification.md](../pagination/specification.md). This page defines the shared envelope that all JSON collection responses must use.

This contract is normative for new and explicitly migrated JSON endpoints. Existing endpoints that still emit legacy payloads must document that exception clearly until they are migrated.

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
- Relationship objects
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
| `meta` | object | Yes | Response metadata and request context |
| `links` | object | Required for collections | Canonical URLs and pagination navigation |

Collection endpoints must always return top-level `links`. Single-resource endpoints should return `links` when a canonical `self` URL or related navigation is useful.

Universal `meta` rule:

- every successful JSON response includes `meta`
- `meta.execution_time_ms` is the shared baseline field across success responses
- collection endpoints add pagination-mode-specific fields such as `page`, `per_page`, `total_items`, `total_pages`, `limit`, or `has_more`
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
- Every resource object exposes a stable `type` token such as `user`, `geokret`, `move`, or `picture`
- For GeoKret resources, `id` must be the public GKID string such as `GK0001`
- Internal database identifiers are not the canonical GeoKret identifier exposed to clients

### 3.4 Link Policy

- `links.self` represents the canonical URL for the current response or resource
- Collection links may include `first`, `prev`, `next`, and `last`
- Relationship objects may expose `links.related` and other relationship-specific URLs
- Clients should follow server-provided links instead of rebuilding pagination URLs when a link is already present

## 4. Resource Object Model

Each item in `data` must use the following structure:

```json
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
}
```

### Rules

- `attributes` contains fields that are not identifiers, relationship linkage, or links
- `relationships` is optional and used only when the response needs to expose related-resource linkage or related URLs
- `links` is optional on a resource object, but `links.self` is recommended whenever a stable canonical resource URL exists
- Relationship linkage may be a single object, an array, or `null`, depending on the relationship cardinality

## 5. Response Families

### 5.1 Single-Resource Response

```json
{
  "data": {
    "id": "GK00FF",
    "type": "geokret",
    "attributes": {
      "name": "Hidden Treasure",
      "type_label": "traditional",
      "missing": false,
      "last_move_at": "2026-03-28T10:00:00Z"
    },
    "links": {
      "self": "/api/v3/geokrety/GK00FF"
    }
  },
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
      "id": "101",
      "type": "user",
      "attributes": {
        "username": "alice",
        "status": "active"
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
        "status": "inactive"
      },
      "links": {
        "self": "/api/v1/users/102"
      }
    }
  ],
  "meta": {
    "execution_time_ms": 11
  },
  "links": {
    "self": "/api/v1/users"
  }
}
```

### 5.3 Paginated Collections

Paginated collections still use top-level `data`, `meta`, and `links`. The pagination mode only changes the metadata fields and which navigation links are present.

- Page-based collections use `page`, `per_page`, `total_items`, and `total_pages`
- Cursor-based collections use `limit`, `has_more`, and `links.next`
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

- Successful responses use `data`, `meta`, and optional or required `links`
- Error responses use `error` and `timestamp`
- Documentation must not mix the success and error envelopes in a single example

## 8. Pagination and Sorting Integration

- Collection pagination is defined in [../pagination/specification.md](../pagination/specification.md)
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

- Every new JSON collection example in the docs uses top-level `data`, `meta`, and `links`
- Every canonical resource example uses `id`, `type`, `attributes`, and optional `relationships` or `links`
- The JSON naming policy is explicit and no longer ambiguous between camelCase and snake_case
- GeoKret responses use public GKIDs as canonical client-facing identifiers
- The pagination page can specialize page-based and cursor-based behavior without redefining the shared envelope
- OpenAPI wording can reference this contract without inventing additional response-shape rules
