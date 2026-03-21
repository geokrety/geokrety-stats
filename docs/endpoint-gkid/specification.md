---
title: GeokretId Type Design and Implementation
version: 1.0
date_created: 2026-03-21
last_updated: 2026-03-21
owner: GeoKrety Stats API Team
tags: [design, api, go, type-system, data-contract]
---

# Introduction

The GeoKrety system manages physical trackable items identified by unique IDs in two representations:

1. **Internal ID** (`int64`): Database-internal identifier used exclusively for foreign keys and internal references
2. **Public GKID** (`string`): User-facing identifier following the format `GK` + hexadecimal value (e.g., `GK0001`, `GKB65C`, `GKFFFF`)

Currently, the GeoKrety Stats API represents all GKID fields as `*int64` in API response structs, with conversion logic scattered throughout handler functions. This specification defines a centralized, type-safe `GeokretId` type that encapsulates the bidirectional conversion between internal and public ID representations, automatically handles JSON serialization/deserialization, and improves type safety across the API.

## 1. Purpose & Scope

**Purpose:** Define a robust, encapsulated `GeokretId` type that centralizes GeoKret ID conversion logic, ensures consistent handling across all API boundaries, and reduces error-prone manual conversions.

**Scope:**
- Type definition and constructor methods
- Conversion algorithms (int ↔ string)
- JSON marshaling/unmarshaling behavior
- Nullable handling using Go's pointer semantics
- Integration patterns for handlers, store layer, and API response entities
- Testing strategy and validation
- Migration path for existing code

**Intended Audience:** Backend API developers, systems architects, and code reviewers working with the GeoKrety Stats API codebase.

**Assumptions:**
- GeoKrety IDs are always positive non-zero integers (1–4294967295 range suitable for 32-bit values)
- Public GKID format uses uppercase hexadecimal with leading zeros (4 hex digits minimum)
- The type will be used in all places where GKID values appear in API responses
- Existing database schema remains unchanged (stores integer gkid values)

## 2. Definitions

| Term | Definition |
|------|-----------|
| **Internal ID** | 64-bit signed integer stored in database for foreign key relationships; not directly exposed to users |
| **Public GKID** | User-facing representation of a GeoKret identifier in format `GK` followed by uppercase hexadecimal digits (e.g., `GK0001`, `GKB65C`) |
| **Marshaling** | Process of converting a Go value to JSON format for API responses |
| **Unmarshaling** | Process of parsing JSON input to populate Go values |
| **Nullable/Optional** | A value that may or may not be present; represented as `*GeokretId` (pointer) in Go |
| **Validation** | Ensuring a value conforms to defined constraints (positive, within range, correct format) |
| **Zero Value** | The default value of a type when not explicitly initialized (for pointers, this is `nil`) |

## 3. Requirements, Constraints & Guidelines

### Functional Requirements

- **REQ-001**: The `GeokretId` type shall encapsulate an internal `int64` value and provide safe access through methods
- **REQ-002**: The type shall convert internal `int64` to public GKID string format (`GK` + uppercase hex; minimum 4 hex digits with zero-padding)
- **REQ-003**: The type shall parse public GKID strings (with or without `GK` prefix) to create valid instances
- **REQ-004**: The type shall also accept plain decimal integer input for backward compatibility with legacy systems
- **REQ-005**: JSON marshaling shall serialize `GeokretId` to public GKID string format
- **REQ-006**: JSON unmarshaling shall accept public GKID strings and decimal integers
- **REQ-007**: Nil pointers (`*GeokretId`) shall marshal to JSON `null` without error
- **REQ-008**: The type shall validate that values are positive non-zero integers
- **REQ-009**: Constructor and parsing functions shall return errors for invalid inputs with descriptive messages
- **REQ-010**: Go string conversion via `fmt.Stringer` interface shall return public GKID format
- **REQ-011**: Go string formatting (`%v`, `%s`) shall display public GKID by default

### Non-Functional Requirements

- **NFR-001**: Type operations (conversion, validation) shall complete in O(1) time
- **NFR-002**: No memory allocations beyond the pointer/value itself in happy-path scenarios
- **NFR-003**: Error messages shall be descriptive and aid developers in debugging invalid inputs
- **NFR-004**: Type shall be compatible with pgx and SQL scanning operations (nullable field support)

### Constraints

- **CON-001**: The internal integer value must remain invariant after creation; immutability is strongly recommended
- **CON-002**: Only non-zero positive integers shall be accepted (1 to 2^63-1 technically, but practically 1 to 2^32-1 per GeoKrety convention)
- **CON-003**: Public GKID format is fixed: `GK` + exactly 4 uppercase hex digits minimum, no alternatives
- **CON-004**: The zero value (uninitialized `GeokretId{}`) is invalid; deliberate construction via `New()` or `FromInt()` is required
- **CON-005**: Pointer receivers (`*GeokretId`) must handle `nil` gracefully in all exported methods
- **CON-006**: No constructor should accept or validate values above typical GeoKrety ranges; range checking is the responsibility of business logic

### Guidelines

- **GUD-001**: Always use `New()` or `FromInt()` constructors rather than struct literal initialization
- **GUD-002**: Prefer `*GeokretId` in API response structs for optional fields; use non-pointer `GeokretId` only when the field is guaranteed to exist
- **GUD-003**: Use the `String()` method or `%v` formatting for logging and debugging; do not manually call `ToGKID()` for display
- **GUD-004**: Validate GeoKrety ID values immediately upon input (in handlers) before passing to store or business logic layers
- **GUD-005**: Store methods should continue to accept `int64` parameters; conversion happens at API boundary (handler layer)
- **GUD-006**: Document which fields in structs use `GeokretId` vs. raw `int64` to clarify when automatic conversion applies

## 4. Interfaces & Data Contracts

### Type Definition

```go
// GeokretId represents a GeoKret identifier with internal and public representations.
// The zero value is not valid; use New() or FromInt() constructors.
type GeokretId struct {
value int64 // unexported; immutable after creation
}
```

### Constructor Methods

```go
// New creates a GeokretId from a public GKID string (e.g., "GK0001", "GKB65C").
// Accepts formats: "GK00FF" (standard), "00FF" (without prefix), "255" (decimal).
// Returns an error if the input is invalid or the resulting integer is non-positive.
func New(gkid string) (*GeokretId, error)

// FromInt creates a GeokretId from an internal integer value.
// Returns an error if the value is non-positive or invalid.
func FromInt(v int64) (*GeokretId, error)

// NewNullable creates a nullable GeokretId from a public GKID string.
// Returns nil if the input is empty or whitespace; otherwise behaves like New().
func NewNullable(gkid string) (*GeokretId, error)
```

### Accessor and Conversion Methods

```go
// Int returns the internal integer value.
// Panics if receiver is nil.
func (g *GeokretId) Int() int64

// ToGKID returns the public GKID string representation (e.g., "GK0001").
// Panics if receiver is nil.
func (g *GeokretId) ToGKID() string

// String implements the fmt.Stringer interface, returning the public GKID format.
// Returns "nil" for nil receiver (safe for use in logging).
func (g *GeokretId) String() string

// IntOrZero returns the internal integer value, or 0 if receiver is nil.
func (g *GeokretId) IntOrZero() int64

// ToGKIDOrEmpty returns the public GKID, or empty string if receiver is nil.
func (g *GeokretId) ToGKIDOrEmpty() string
```

### JSON Marshaling Interface

```go
// MarshalJSON encodes the GeokretId as a public GKID string in JSON.
// Nil receiver marshals as JSON null.
// Example output: {"gkid": "GK0001"}
func (g *GeokretId) MarshalJSON() ([]byte, error)

// UnmarshalJSON decodes a GeokretId from JSON (string or number).
// Accepts: "GK0001", "0001" (hex without prefix), 1 (decimal integer).
// Sets receiver to nil if input is JSON null.
func (g *GeokretId) UnmarshalJSON(data []byte) error
```

### Entity Integration

All existing entity structs containing GKID fields shall migrate from `*int64` to `*GeokretId`:

**Before:**
```go
type GeokretListItem struct {
    ID       int64   `db:"id" json:"id"`
    GKID     *int64  `db:"gkid" json:"gkid"`
    Name     string  `db:"name" json:"name"`
    // ...
}
```

**After:**
```go
type GeokretListItem struct {
    ID       int64        `db:"id" json:"id"`
    GKID     *GeokretId   `db:"gkid" json:"gkid"`
    Name     string       `db:"name" json:"name"`
    // ...
}
```

### Error Contract

All parsing and constructor functions shall return a `GeokretIdError` with clear context:

```go
type GeokretIdError struct {
    Input  string // the invalid input value
    Reason string // human-readable explanation
}

// Example errors:
// Input: "GK0000", Reason: "gkid must be greater than zero"
// Input: "XYZ123", Reason: "invalid gkid format; expected GK[0-9A-F]+ or decimal integer"
// Input: "-5", Reason: "gkid must be positive"
```

## 5. Acceptance Criteria

- **AC-001**: Given a valid GKID string `"GK0001"`, When creating `New("GK0001")`, Then a non-nil `*GeokretId` is returned with `ToGKID()` = `"GK0001"`

- **AC-002**: Given a valid integer `1`, When creating `FromInt(1)`, Then a non-nil `*GeokretId` is returned with `Int()` = `1` and `ToGKID()` = `"GK0001"`

- **AC-003**: Given an integer `255`, When creating `FromInt(255)`, Then `ToGKID()` = `"GK00FF"` (zero-padded hex)

- **AC-004**: Given JSON input `{"gkid": "GK0001"}`, When unmarshaling into a struct with `*GeokretId` field, Then the field is populated correctly and `Int()` = `1`

- **AC-005**: Given JSON input `{"gkid": 1}` (plain integer), When unmarshaling, Then it is accepted and converted to `*GeokretId` with value `1`

- **AC-006**: Given a `*GeokretId` with value `255`, When marshaling to JSON, Then the output is `"GK00FF"` (not `255`)

- **AC-007**: Given a nil `*GeokretId` receiver, When calling `String()`, Then it returns `"nil"` without panic

- **AC-008**: Given an invalid input `"GK0000"` (zero value), When calling `New()`, Then an error is returned with reason mentioning "greater than zero"

- **AC-009**: Given an empty or whitespace-only string, When calling `NewNullable()`, Then nil is returned without error

- **AC-010**: Given valid inputs in multiple formats (`"GK0001"`, `"0001"`, `"1"`), When parsing each format, Then all resolve to the same internal value

- **AC-011**: Given a legacy endpoint accepting both GKID and plain integer parameters, When the field is populated with `*GeokretId`, Then both input formats work transparently

- **AC-012**: Given API response struct with `GKID *GeokretId` field, When marshaling to JSON, Then the output shows public GKID string (e.g., `"gkid": "GK0001"`), not the internal integer

## 6. Test Automation Strategy

### Test Levels & Frameworks

- **Unit Tests**: Use Go's standard `testing` package; focus on constructor, conversion, and edge cases
- **Integration Tests**: Test with actual database queries and API routes; use prepared test datasets
- **JSON Serialization Tests**: Use `encoding/json` package to verify marshaling/unmarshaling in both directions
- **Database Scanning Tests**: Verify `sql.Scanner` and `driver.Valuer` implementations work with pgx

### Test Coverage Areas

#### Constructor Tests
```
✓ New() with valid GKID format (e.g., "GK0001")
✓ New() with hex without prefix (e.g., "0001")
✓ New() with decimal integer string (e.g., "1")
✓ New() with zero value → error
✓ New() with negative value → error
✓ New() with invalid format → error
✓ FromInt() with positive integer
✓ FromInt() with zero → error
✓ FromInt() with negative → error
✓ NewNullable() with empty string → nil, no error
✓ NewNullable() with whitespace → nil, no error
✓ NewNullable() with invalid format → error
```

#### Conversion Tests
```
✓ Int() returns correct internal value
✓ ToGKID() formats as GK + zero-padded hex
✓ ToGKID() with small value (e.g., 1 → "GK0001")
✓ ToGKID() with large value (e.g., 65535 → "GKFFFF")
✓ String() returns GKID format
✓ String() on nil receiver returns "nil"
✓ IntOrZero() returns value or 0 for nil
✓ ToGKIDOrEmpty() returns GKID or "" for nil
```

#### JSON Marshaling Tests
```
✓ Marshal *GeokretId → JSON string (e.g., "GK0001")
✓ Marshal nil *GeokretId → JSON null
✓ Unmarshal "GK0001" → *GeokretId with value 1
✓ Unmarshal 1 (integer) → *GeokretId with value 1
✓ Unmarshal null → nil *GeokretId
✓ Unmarshal invalid format → error
✓ Round-trip marshal/unmarshal preserves value
```

#### Entity Integration Tests
```
✓ Scan from database integer row → *GeokretId
✓ Scan null value → nil *GeokretId
✓ HTTP response JSON with GKID field → public format
✓ Entity with multiple *GeokretId fields → all converted
```

#### Backward Compatibility Tests
```
✓ parsePublicGKIDParam() handler function works with GeokretId
✓ Store methods receiving int64 still work (no changes)
✓ Legacy numeric API responses behave correctly
```

### Test Data

Use representative samples:
- **Boundary values**: 1, 255, 65535, 2^32-1
- **Special values**: high values for stress testing
- **Invalid formats**: empty string, null, non-hex characters, zero, negative numbers
- **Case variations**: "GK0001", "gk0001", "Gk0001"

### CI/CD Integration

- Run unit tests on every PR
- Run integration tests in staging environment with test database
- Verify JSON serialization tests pass before API deployment
- Generate coverage report; maintain >95% coverage for this type

### Coverage Threshold

Minimum 95% code coverage including:
- All constructor paths
- All error branches
- Nil receiver handling
- Edge cases (max value, zero, negative)

## 7. Rationale & Context

### Why a Type Wrapper?

The current approach of using `int64` fields directly scatters conversion logic throughout handler code, leading to:
- Repeated validation logic
- Inconsistent error handling
- Difficulty ensuring JSON serialization formats correctly
- Lack of type safety (easy to confuse internal ID with public GKID)

A dedicated `GeokretId` type provides:
- **Single source of truth** for conversion logic
- **Type safety** at compile time; Go compiler prevents mixing GKID with plain integers
- **Automatic serialization** to public format in API responses
- **Centralized validation** and error messaging
- **Enhanced readability** of code (clear intent that a field uses GKID)

### Design Decisions

**Pointer-based nullability**: Go uses `*T` to represent optional values. Alternatives (custom `Option<T>` enum, three-state booleans) are less idiomatic.

**Unexported field**: The internal `value` field is unexported to enforce immutability and prevent direct manipulation.

**Separate constructors for different input types**: `New()` and `FromInt()` clearly distinguish source context (user input vs. internal system value).

**`String()` method returns GKID, not formatted with type name**: Logging `g.String()` yields `"GK0001"` (clean), not `"GeokretId(GK0001)"` (overly verbose for logs).

**Panic on nil receiver in `Int()` and `ToGKID()`**: These are primary accessors; if a developer calls them on nil, it's a logic error worth surfacing. Convenience methods `IntOrZero()` and `ToGKIDOrEmpty()` provide nil-safe alternatives.

**JSON unmarshaling accepts both string and integer**: Maintains compatibility with legacy clients or systems that may send plain integers instead of GKID strings.

## 8. Dependencies & External Integrations

### Data Dependencies
- **DAT-001**: PostgreSQL database column `geokrety.gk_geokrety.gkid` (integer type) — The type reads from and writes to this column via pgx driver
- **DAT-002**: API client payloads and responses — External systems expecting GKID strings in API JSON responses

### Technology Platform Dependencies
- **PLT-001**: Go 1.21+ — Standard library packages used (`encoding/json`, `strconv`, `fmt`); no compatibility guarantees for earlier versions
- **PLT-002**: pgx PostgreSQL driver — Database scanning and type conversion integration

### Infrastructure Dependencies
- **INF-001**: HTTP API router (chi/other) — Handler layer that parses incoming GKID parameters and uses this type
- **INF-002**: SQL database layer — Store queries that work with int64 values; type converts at handler boundary

### Third-Party Services
- None directly; this type is self-contained

### Compliance Dependencies
- **COM-001**: Data validation — Must reject zero and negative values per GeoKrety system rules

## 9. Examples & Edge Cases

### Basic Usage

```go
// Constructor from GKID string
gid, err := geo.New("GK0001")
if err != nil {
    log.Fatal(err)
}
fmt.Println(gid.Int())      // Output: 1
fmt.Println(gid.ToGKID())   // Output: GK0001
fmt.Println(gid)            // Output: GK0001 (via String())

// Constructor from integer
gid2, err := geo.FromInt(255)
if err != nil {
    log.Fatal(err)
}
fmt.Println(gid2.ToGKID())  // Output: GK00FF

// In a struct
type GeokretListItem struct {
    ID      int64       `json:"id"`
    GKID    *GeokretId  `json:"gkid"`    // Auto-marshals to public format
    Name    string      `json:"name"`
}

item := GeokretListItem{
    ID:   123,
    GKID: gid,
    Name: "Test Geokrety",
}

data, _ := json.Marshal(item)
fmt.Println(string(data))
// Output: {"id":123,"gkid":"GK0001","name":"Test Geokrety"}
```

### Nullable Handling

```go
// Nil value
var nilGid *GeokretId
fmt.Println(nilGid)         // Output: nil (safe; no panic)
fmt.Println(nilGid.IntOrZero())      // Output: 0
fmt.Println(nilGid.ToGKIDOrEmpty())  // Output: ""

// In JSON
type Response struct {
    GKID *GeokretId `json:"gkid,omitempty"`
}

resp := Response{GKID: nil}
data, _ := json.Marshal(resp)
fmt.Println(string(data))   // Output: {} (field omitted due to nil)
```

### Error Handling

```go
// Zero value error
gid, err := geo.New("GK0000")
fmt.Println(err)  // Output: gkid must be greater than zero

// Negative value
gid, err := geo.FromInt(-5)
fmt.Println(err)  // Output: gkid must be positive

// Invalid format
gid, err := geo.New("INVALID")
fmt.Println(err)  // Output: invalid gkid format; expected GK[0-9A-F]+ or decimal integer

// JSON with null
type Response struct {
    GKID *GeokretId `json:"gkid"`
}
var resp Response
json.Unmarshal([]byte(`{"gkid": null}`), &resp)
fmt.Println(resp.GKID)  // Output: <nil>
```

### Input Format Flexibility

```go
// All these inputs represent the same GeoKret (ID = 255)
formats := []string{"GK00FF", "gk00ff", "00FF", "FF", "255"}

for _, f := range formats {
    gid, _ := geo.New(f)
    fmt.Println(gid.ToGKID())  // All output: GK00FF
}
```

### Handler Integration

```go
// Old code (before)
func GetGeokretyByGkId(w http.ResponseWriter, r *http.Request) {
    gkidStr := chi.URLParam(r, "gkid")
    gkid, ok := parsePublicGKIDParam(w, r, "gkid")  // Custom parsing
    if !ok { return }

    // ... fetch from store with int64 gkid
}

// New code (after)
func GetGeokretyByGkId(w http.ResponseWriter, r *http.Request) {
    gkidStr := chi.URLParam(r, "gkid")
    gkid, err := geo.New(gkidStr)  // Centralized, type-safe
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    // ... fetch from store with gkid.Int()
    // API response struct automatically serializes as "GK00FF"
}
```

### Database Scanning

```go
// Store layer (no changes needed)
func (s *Store) FetchGeokretyByGKID(ctx context.Context, gkid int64) (GeokretDetails, error) {
    // ... existing query
}

// Handler layer (bridges to new type)
func (h *StatsHandler) GetGeokretyDetailsByGkId(w http.ResponseWriter, r *http.Request) {
    gkidVal, err := geo.New(chi.URLParam(r, "gkid"))
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }

    details, err := h.store.FetchGeokretyByGKID(r.Context(), gkidVal.Int())
    // ...response includes GeokretDetails.GKID (*GeokretId) → auto-serialized
}
```

## 10. Validation Criteria

✓ Type compiles without errors in Go 1.21+

✓ All constructor functions work with representative test inputs

✓ JSON marshaling produces correctly formatted strings ("GK" + hex)

✓ JSON unmarshaling accepts multiple input formats and converts correctly

✓ Nil receiver methods (`String()`, `IntOrZero()`, etc.) don't panic

✓ Panic-prone methods (`Int()`, `ToGKID()`) clearly document invariant (non-nil receiver)

✓ Error messages are descriptive and aid debugging

✓ Type integrates seamlessly with existing HTTP handler code

✓ Type integrates seamlessly with existing entity structs

✓ Database transaction tests show correct round-trip (DB int → GeokretId → JSON → client)

✓ Backward compatibility maintained: existing store methods unchanged

✓ New API responses show public GKID format in JSON

✓ Code coverage ≥95% for the type implementation
