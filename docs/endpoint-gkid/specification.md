---
title: GeokretId Type Implementation Specification
version: 1.0
date_created: 2026-03-21
last_updated: 2026-03-21
owner: GeoKrety Stats API Team
tags: [design, api, types, gkid, database]
---

# Specification: GeokretId Type for GKID Representation

## 1. Introduction

The GeoKrety system uses two representations for GeoKret identifiers:
- **Internal ID** (`int64`): Database-internal foreign key identifier
- **Public GKID** (`string`): User-facing identifier in format `GK` + uppercase hexadecimal (e.g., `GK0001`, `GKB65C`)

Currently, GKID fields are represented as `*int64` in API response structs, causing a critical issue: **the API exposes internal database integers instead of the public GKID format** that users expect.

This specification defines a `GeokretId` type that:
- **Encapsulates** integer ↔ string conversions
- **Automatically serializes** as public GKID format in JSON responses
- **Accepts flexible input** (decimal, "GKXXXX", or "XXXX" formats)
- **Improves type safety** and reduces conversion errors throughout the codebase
- **Maintains backward compatibility** with existing database and handler logic

## 2. Purpose & Scope

**Purpose:** Define a robust, type-safe Go type for handling GKID conversions throughout the GeoKrety Stats API, ensuring all API responses present GKID in the public string format while accepting flexible input formats from users.

**Scope:**
- Database entity layer (`internal/db/`)
- API response structs and serialization
- Handler parameter parsing (`internal/handlers/`)
- Store/repository methods

**Audience:** Backend developers, API consumers, database team

**Assumptions:**
- Internal GKID values are always non-negative integers (0 to 2^63-1)
- Public GKID format follows the rule: `GK` + 4-16 uppercase hex digits
- JSON serialization must output public GKID format, not integers
- Existing database schema remains unchanged (stores `gkid` as bigint)

---

## 3. Definitions

- **Internal ID**: Unsigned 64-bit integer stored in database (`geokrety.gk_geokrety.gkid`)
- **Public GKID**: User-facing string format: `GK` + uppercase hex (e.g., `GK0001`, `GK3D45F`)
- **Marshaling**: Converting a `GeokretId` to JSON (output as public string)
- **Unmarshaling**: Parsing input (string or int) into a `GeokretId`
- **Nullable GeokretId**: `*GeokretId` where `nil` represents missing/unknown value
- **GeokretIdError**: Custom error type for validation and conversion failures

---

## 4. Requirements

### Functional Requirements

- **REQ-001**: `GeokretId` type shall store a single `int64` value internally (unexported)
- **REQ-002**: Type shall provide constructor `New(int64) GeokretId` to create from integer
- **REQ-003**: Type shall provide `FromString(string) (GeokretId, error)` to parse public GKID or decimal strings
- **REQ-004**: Type shall provide `Int() int64` accessor to get internal integer value
- **REQ-005**: Type shall provide `ToGKID() string` to convert to public GKID format (e.g., `GK0001`)
- **REQ-006**: Type shall implement `String() string` returning public GKID format
- **REQ-007**: Type shall implement `MarshalJSON() ([]byte, error)` to serialize as quoted public GKID string (e.g., `"GK0001"`)
- **REQ-008**: Type shall implement `UnmarshalJSON([]byte) error` to accept input as stringified integer or GKID format
- **REQ-009**: Nullable `*GeokretId` shall support `nil` value with safe accessors (`IntOrZero()`, `ToGKIDOrEmpty()`)
- **REQ-010**: Type shall support scanner interface for database row scanning (`sql.Scanner`)
- **REQ-011**: Input formats shall be flexible: accept "GK00FF", "00FF", and decimal "255" representations

### Non-Functional Requirements

- **NFR-001**: Conversion operations shall complete in O(1) time
- **NFR-002**: No external dependencies (stdlib only)
- **NFR-003**: Type shall be immutable after creation
- **NFR-004**: Error messages shall be developer-friendly and actionable

### Constraints

- **CON-001**: Valid GKID values must be >= 1 (zero is invalid)
- **CON-002**: GKID values must fit within int64 range (0 to 9,223,372,036,854,775,807)
- **CON-003**: Public GKID format uses exactly 2 uppercase hex digits minimum (not "GK" alone)
- **CON-004**: Nil is the only valid "empty" state for `*GeokretId`

### Guidelines

- **GUD-001**: Always use `GeokretId` in API response structs, never `int64` or raw `string`
- **GUD-002**: Use `*GeokretId` for nullable fields (e.g., optional GKID references)
- **GUD-003**: Validate and convert input at handler boundary, not in store/repository
- **GUD-004**: Log GKID as public format in error/debug messages (e.g., `"GK0001"`, not `1`)

---

## 5. Design

### 5.1 Type Definition

```go
package db

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "strconv"
    "strings"
)

// GeokretId represents a GeoKret identifier.
// Internally stores int64; externally presented as public GKID format.
type GeokretId struct {
    value int64 // unexported: prevents accidental mutation
}

// GeokretIdError represents a validation or conversion error.
type GeokretIdError struct {
    Input   string
    Message string
}

func (e GeokretIdError) Error() string {
    return fmt.Sprintf("invalid geokret id %q: %s", e.Input, e.Message)
}
```

### 5.2 Constructor & Factory Methods

```go
// New creates a GeokretId from an integer value.
func New(value int64) GeokretId {
    return GeokretId{value: value}
}

// FromString parses a string into GeokretId.
// Accepts formats: "GK00FF", "00FF", "255" (decimal).
func FromString(s string) (GeokretId, error) {
    // See Implementation Details section below
}

// NewNullable creates a pointer to GeokretId (for nullable fields).
func NewNullable(value int64) *GeokretId {
    id := New(value)
    return &id
}

// NullableFromString parses a string into *GeokretId.
func NullableFromString(s string) (*GeokretId, error) {
    id, err := FromString(s)
    if err != nil {
        return nil, err
    }
    return &id, nil
}
```

### 5.3 Accessor Methods

```go
// Int returns the internal integer value.
func (g GeokretId) Int() int64 {
    return g.value
}

// ToGKID returns the public GKID format (e.g., "GK0001").
func (g GeokretId) ToGKID() string {
    return "GK" + strings.ToUpper(strconv.FormatInt(g.value, 16))
}

// String returns the public GKID format (satisfies fmt.Stringer).
func (g GeokretId) String() string {
    return g.ToGKID()
}

// IntOrZero returns the integer value, or 0 if receiver is nil (for *GeokretId).
func (g *GeokretId) IntOrZero() int64 {
    if g == nil {
        return 0
    }
    return g.value
}

// ToGKIDOrEmpty returns the public GKID or empty string if nil.
func (g *GeokretId) ToGKIDOrEmpty() string {
    if g == nil {
        return ""
    }
    return g.ToGKID()
}
```

### 5.4 JSON Marshaling

```go
// MarshalJSON serializes GeokretId as a quoted GKID string.
func (g GeokretId) MarshalJSON() ([]byte, error) {
    return json.Marshal(g.ToGKID())
}

// UnmarshalJSON parses JSON input (int or string) into GeokretId.
func (g *GeokretId) UnmarshalJSON(data []byte) error {
    var raw interface{}
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }

    var s string
    switch v := raw.(type) {
    case string:
        s = v
    case float64:
        s = strconv.FormatInt(int64(v), 10)
    default:
        return GeokretIdError{Input: string(data), Message: "expected string or number"}
    }

    id, err := FromString(s)
    if err != nil {
        return err
    }
    *g = id
    return nil
}
```

### 5.5 Database Scanning

```go
// Scan implements sql.Scanner for database scanning.
func (g *GeokretId) Scan(value interface{}) error {
    if value == nil {
        *g = GeokretId{value: 0}
        return nil
    }

    switch v := value.(type) {
    case int64:
        *g = New(v)
    case int:
        *g = New(int64(v))
    default:
        return GeokretIdError{Input: fmt.Sprintf("%T", value), Message: "cannot scan type"}
    }
    return nil
}

// Value implements driver.Valuer for database writing.
func (g GeokretId) Value() (driver.Value, error) {
    return g.value, nil
}
```

---

## 6. Implementation Details

### 6.1 String Parsing Algorithm

```go
func FromString(s string) (GeokretId, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return GeokretId{}, GeokretIdError{Input: s, Message: "empty string"}
    }

    normalized := strings.ToUpper(s)

    // Try parsing as decimal integer first
    if isDecimal(normalized) {
        val, err := strconv.ParseInt(normalized, 10, 64)
        if err != nil || val <= 0 {
            return GeokretId{}, GeokretIdError{Input: s, Message: "invalid decimal or zero value"}
        }
        return New(val), nil
    }

    // Try parsing as GKID (with or without "GK" prefix)
    hexPart := normalized
    if strings.HasPrefix(hexPart, "GK") {
        hexPart = strings.TrimPrefix(hexPart, "GK")
    }

    if hexPart == "" || !isValidHex(hexPart) {
        return GeokretId{}, GeokretIdError{Input: s, Message: "invalid hex format"}
    }

    val, err := strconv.ParseInt(hexPart, 16, 64)
    if err != nil || val <= 0 {
        return GeokretId{}, GeokretIdError{Input: s, Message: "invalid hex value or zero"}
    }

    return New(val), nil
}

func isDecimal(s string) bool {
    for _, c := range s {
        if c < '0' || c > '9' {
            return false
        }
    }
    return true
}

func isValidHex(s string) bool {
    for _, c := range s {
        if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
            return false
        }
    }
    return true
}
```

### 6.2 Entity Integration

**Before:**
```go
type GeokretListItem struct {
    ID       int64
    GKID     *int64      // ← Problem: exposes internal ID as JSON
}
```

**After:**
```go
type GeokretListItem struct {
    ID       int64
    GKID     *GeokretId  // ← Solution: automatically serializes as "GK0001"
}
```

### 6.3 Handler Integration

**Before:**
```go
func (h *StatsHandler) GetGeokretyDetailsByGkId(w http.ResponseWriter, r *http.Request) {
    gkid, ok := parsePublicGKIDParam(w, r, "gkid")
    // gkid is int64; later converted for response
}
```

**After:**
```go
func (h *StatsHandler) GetGeokretyDetailsByGkId(w http.ResponseWriter, r *http.Request) {
    gkid, ok := parsePublicGKIDParam(w, r, "gkid")
    // gkid is GeokretId; automatically serializes correctly
}
```

---

## 7. Testing Strategy

### Unit Tests
- Constructor validation (valid/invalid values, zero handling, range limits)
- String conversion (ToGKID, String method)
- Parsing function (decimal, GKID with/without prefix, invalid formats)
- Nullable handling (nil checks, IntOrZero, ToGKIDOrEmpty)
- JSON marshaling (output format validation)
- JSON unmarshaling (decimal and string inputs)
- Database scanning (int64, int, nil values)
- Error cases (invalid formats, zero values, out-of-range)

### Integration Tests
- Entity hydration from database rows
- Handler parameter parsing and response serialization
- Full request/response cycle

### JSON Serialization Tests
- Plain GeokretId serializes to quoted GKID: `"GK0001"`
- Nullable *GeokretId: nil serializes to JSON `null`, non-nil to `"GKXXXX"`
- Structs containing GeokretId fields

### Coverage Target
- ≥95% line coverage on `geokret_id.go`

---

## 8. Acceptance Criteria

- **AC-001**: `GeokretId{value: 1}.String()` returns `"GK0001"`
- **AC-002**: `FromString("GK00FF")` returns `GeokretId{value: 255}` without error
- **AC-003**: `FromString("255")` returns `GeokretId{value: 255}` without error
- **AC-004**: `FromString("0")` returns `GeokretIdError` (zero is invalid)
- **AC-005**: `FromString("INVALID")` returns `GeokretIdError`
- **AC-006**: JSON output for `GeokretId` is quoted string: `"GK0001"`, not `1`
- **AC-007**: JSON unmarshaling `"GK00FF"` produces `GeokretId{value: 255}`
- **AC-008**: JSON unmarshaling `255` produces `GeokretId{value: 255}`
- **AC-009**: Nullable `*GeokretId` with nil value marshals to `null`
- **AC-010**: Nullable `*GeokretId` non-nil marshals to public GKID: `"GKXXXX"`
- **AC-011**: Invalid JSON input type (bool, array) returns UnmarshalTypeError
- **AC-012**: Database Scan handles int64, int, and nil without panic

---

## 9. Diagrams & Examples

### GKID Format Conversion

```
Integer → GKID Format
   1    → "GK0001"
  255   → "GK00FF"
 61765  → "GKF1A5"

GKID Format → Integer
"GK0001"   →    1
"GK00FF"   →  255
"GKF1A5"   → 61765
```

### JSON Serialization Example

```go
type GeokretListItem struct {
    ID   int64      `json:"id"`
    GKID *GeokretId `json:"gkid"`
    Name string     `json:"name"`
}

item := GeokretListItem{
    ID:   123,
    GKID: NewNullable(255),
    Name: "TestGK",
}

// JSON Output:
// {
//   "id": 123,
//   "gkid": "GK00FF",    ← Public GKID, not integer!
//   "name": "TestGK"
// }
```

### Handler Integration Example

```go
func (h *StatsHandler) GetGeokretyDetailsByGkId(w http.ResponseWriter, r *http.Request) {
    gkid, ok := parseGeokretIdParam(w, r, "gkid", "id")
    if !ok {
        return  // Error already written
    }

    row, err := h.store.FetchGeokretyByGKID(r.Context(), gkid.Int())
    if err != nil {
        h.writeStoreError(w, err, "failed to fetch geokret")
        return
    }

    writeEnvelope(w, http.StatusOK, row, ...)
    // row.GKID is *GeokretId, automatically serializes as "GK..."
}
```

---

## 10. Deliverables

1. **Type Definition** (`api/internal/db/geokret_id.go`)
   - `GeokretId` struct
   - `GeokretIdError` type
   - Constructor and factory methods
   - Accessor methods
   - JSON marshaling/unmarshaling
   - Database scanning

2. **Unit Tests** (`api/internal/db/geokret_id_test.go`)
   - Comprehensive test suite (≥95% coverage)
   - Edge cases and error scenarios

3. **Database Entity Updates** (multiple files)
   - Change `GKID` fields from `*int64` to `*GeokretId`
   - Update affected types: `GeokretListItem`, `GeokretDetails`, `DormancyRecord`, `MultiplierVelocityRecord`, etc.
   - Update store method signatures if needed

4. **Handler Updates** (`api/internal/handlers/entities.go`)
   - Create/update `parseGeokretIdParam()` to use new type
   - Verify JSON serialization in responses
   - Update any custom marshaling logic

5. **Documentation**
   - Code comments on public methods
   - API documentation updated to reflect GKID format in responses
   - Migration guide for API consumers

6. **Backward Compatibility Check**
   - Verify all existing endpoint tests pass
   - Spot-check API response format (should now show `"GK..."` instead of integers)

---

## 11. Validation Criteria

- [ ] `GeokretId` type defined and compiles
- [ ] All constructor/factory methods implemented
- [ ] All accessor methods functional
- [ ] JSON marshaling produces quoted GKID strings
- [ ] JSON unmarshaling accepts flexible input
- [ ] Database Scan implementation handles nulls
- [ ] Unit tests written (≥95% coverage)
- [ ] All unit tests pass
- [ ] Entity structs updated to use `GeokretId`
- [ ] Handlers updated to parse into `GeokretId`
- [ ] Integration tests pass (entity → JSON)
- [ ] API response verification (actual "GK..." format, not integer)
- [ ] Error messages are actionable

---

## 12. Related Specifications / Further Reading

- [Schema: GeoKret Stats](../schema/specs.md) — Database schema reference
- [GeoKrety System Expert Skill](../../.github/skills/geokrety-system-expert/SKILL.md) — Domain knowledge
- [API Instructions](../../.github/instructions/api.instructions.md) — API design guidelines
- [Conventional Commits](vscode-userdata:/home/kumy/.config/Code/User/prompts/conventional-commit.instructions.md) — Commit message format
