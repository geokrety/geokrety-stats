---
title: TrackingCode Value Type - GeoKrety Shared Module
version: 1.0
date_created: 2026-03-22
last_updated: 2026-03-22
owner: GeoKrety Stats Development Team
tags: [design, architecture, go, security, tracking-code]
---

# Introduction

This specification defines a new Go value type for managing GeoKrety tracking codes in a safe, explicit, and reusable way. The immediate trigger is `U03` from the live task inputs: create a struct for managing a GeoKret tracking code such that its textual string representation is obfuscated, for example `ABCDEF -> A****F`.

The design target is the shared `geokrety` Go module already used by the API through a local module replace. The new type should follow the same broad style as `GeokretId`: private storage, explicit constructors, receiver methods, predictable formatting behavior, and strong default safety. Unlike `GeokretId`, the main concern here is not public identifier formatting but secret redaction.

## 1. Purpose & Scope

### Purpose

Introduce a dedicated `TrackingCode` value type that:

- prevents accidental leakage of full tracking codes in logs, API responses, and debug output
- provides a single masking rule for all textual representations
- keeps the raw value available only for internal application logic and database persistence
- aligns with the existing `GeokretId` pattern of encapsulating domain-specific identifiers in dedicated Go types

### Scope

This specification covers:

1. A new `TrackingCode` struct in the shared `geokrety` module
2. Constructor functions and helper methods
3. Masked textual formatting rules
4. JSON/XML/text/YAML/CSV default serialization behavior
5. Optional database scan/value support for raw persistence
6. Unit tests and acceptance criteria

### Out of Scope

This specification does not define:

1. the authoritative validation regex for tracking codes if that logic already belongs to the website/database layer
2. database schema changes
3. endpoint changes to expose tracking codes publicly
4. permission models for privileged administrative access to raw tracking codes

## 2. Definitions

| Term | Definition |
|------|------------|
| Tracking Code | A secret code assigned to a GeoKret for operations that must never be exposed publicly in full. |
| Masked Value | A redacted string form that reveals only the first character and replaces the remainder with `*`. |
| Raw Value | The original tracking code string kept internally by the type for database or privileged internal use. |
| Obfuscated String Representation | The result returned by `String()` and equivalent text-oriented formatters or marshalers. |

## 2.5 Why Masking Matters (Security Context)

Tracking codes function as authentication credentials for GeoKrety operations. Full exposure of a tracking code through logs, error messages, network captures, or debug output enables attackers to:

- Impersonate users and falsify tracking history
- Manipulate GeoKrety movement data
- Harvest credentials from unencrypted logs or backups

This specification mitigates exposure risk by defaulting to masked representation (e.g., `ABCDEF` → `A*****`). This approach:

- **Reduces accidental exposure**: Public APIs and formatters never expose the full code without explicit developer action
- **Maintains usability**: Masked codes are still readable in logs (visible prefix aids debugging)
- **Keeps raw access explicit**: Developers using `RawForInternalUseOnly()` must acknowledge the sensitivity

The trade-off is reduced auditability for increased security-by-default. Privileged audit logging can still access raw values when needed.

## 3. Requirements, Constraints & Guidelines

### Functional Requirements

- **REQ-001**: A new value type named `TrackingCode` must be introduced in the shared `geokrety` module.
- **REQ-002**: The raw tracking code must be stored in a private field.
- **REQ-003**: `TrackingCode.String()` must return a masked value, not the raw value.
- **REQ-004**: For `ABCDEF`, the masked string must be exactly `A*****`.
- **REQ-005**: JSON, XML, text, YAML, and CSV textual serialization must use the masked value by default.
- **REQ-006a**: Tracking codes must have a minimum length of 4 characters after normalization; values shorter than 4 characters must be rejected.
- **REQ-006b**: Tracking codes must contain only uppercase letters (A-Z) and digits (0-9) after normalization; any non-alphanumeric characters, special characters, or lowercase letters must be rejected.
- **REQ-006c**: Tracking codes must be normalized during construction; the normalization rule is:
  1. Trim surrounding whitespace
  2. Convert to uppercase
  3. Validate that only uppercase letters (A-Z) and digits (0-9) remain
- **REQ-007**: The type must expose a deliberate internal-use method for retrieving the raw value when necessary.
- **REQ-008**: Database persistence helpers, if implemented, must store and retrieve the raw value, not the masked value.
- **REQ-009**: Nil or zero-value instances must never accidentally render the raw code.
- **REQ-010**: The implementation must include unit tests for masking, constructors, formatting, marshaling, and database helpers if present.

### Security Requirements

- **SEC-001**: Public-facing or textual representations of the type must be masked by default.
- **SEC-002**: The raw tracking code must not be returned by `String()`, `fmt` formatting, or default JSON/XML/YAML/text serialization.
- **SEC-003**: The type must make accidental disclosure harder than using a plain `string`.
- **SEC-004**: The implementation must remain consistent with the API rule that `geokrety.gk_geokrety.tracking_code` must never be exposed in API responses.

### Design Requirements

- **DES-001**: Follow the style of `geokrety/geokrety/gkid.go`: a private field, explicit constructors, and receiver methods.
- **DES-002**: Prefer method names that make safe behavior obvious, such as `String()` for masked output and `Raw()` for explicit internal access.
- **DES-003**: If `fmt.Formatter` is implemented, `%s`, `%v`, and `%q` must use the masked form.
- **DES-004**: If `database/sql.Scanner` and `driver.Valuer` are implemented, `Scan()` must accept the raw DB value and `Value()` must return the raw DB value.
- **DES-005**: Default marshaling behavior should be safe-by-default and masked.

### Assumptions

1. Tracking-code format validation is enforced at construction time: minimum 4 characters, only [A-Z] letters allowed, lowercase converted to uppercase.
2. Normalization rule: trim whitespace → uppercase → validate only [A-Z].
3. The masked representation preserves the original string length whenever possible.

## 4. Proposed Go API

### File Placement

```
geokrety/
└── geokrety/
    ├── tracking_code.go
    └── tracking_code_test.go
```

### Struct Definition

```go
type TrackingCode struct {
	value string
}
```

### Constructors

```go
func NewTrackingCode(raw string) (*TrackingCode, error)
func NewNullableTrackingCode(raw string) (*TrackingCode, error)
```

Constructor behavior:

1. Trim surrounding whitespace
2. Reject empty results
3. Normalize the stored raw value to uppercase

### Core Methods

```go
func (t *TrackingCode) Raw() string
func (t *TrackingCode) Masked() string
func (t *TrackingCode) String() string
func (t *TrackingCode) Valid() bool
```

### Optional Formatting & Persistence Helpers

```go
func (t TrackingCode) Format(state fmt.State, verb rune)
func (t TrackingCode) MarshalJSON() ([]byte, error)
func (t *TrackingCode) UnmarshalJSON(data []byte) error
func (t TrackingCode) MarshalText() ([]byte, error)
func (t *TrackingCode) UnmarshalText(text []byte) error
func (t TrackingCode) MarshalXML(enc *xml.Encoder, start xml.StartElement) error
func (t *TrackingCode) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error
func (t TrackingCode) MarshalXMLAttr(name xml.Name) (xml.Attr, error)
func (t *TrackingCode) UnmarshalXMLAttr(attr xml.Attr) error
func (t TrackingCode) MarshalYAML() (any, error)
func (t *TrackingCode) UnmarshalYAML(node *yaml.Node) error
func (t TrackingCode) MarshalCSV() (string, error)
func (t *TrackingCode) UnmarshalCSV(value string) error
func (t *TrackingCode) Scan(src any) error
func (t TrackingCode) Value() (driver.Value, error)
```

### Minimum Length Requirement

Tracking codes must be at least 4 characters long after trimming and normalization. Shorter codes must be rejected during construction.

### Masking Rules

### Primary Rule

For a tracking code with length `n >= 4`:

```
masked = first_character + strings.Repeat("*", n-1)
```

Examples:

| Raw | Masked |
|-----|--------|
| `ABCD` | `A***` |
| `ABCDEF` | `A*****` |
| `GK1234` | `G*****` |
| `SECRET42` | `S*******` |

### Short-Value Rule

Because all valid tracking codes must be at least 4 characters, items with fewer than 4 characters are invalid and should not be constructed. The `NewTrackingCode` constructor must reject inputs shorter than 4 characters.

This preserves the principle that `String()` must never reveal the full raw code.

## 6. Serialization Contract

### JSON

- Marshal as the masked string
- Unmarshal from a raw string input
- `null` may reset the value to invalid/nil if that matches the surrounding style used by `GeokretId`

Example:

```json
"A****F"
```

### XML

- Element text and attribute values must use the masked string when marshaled
- Unmarshal should accept raw input strings

### Text / fmt

- `String()` returns masked text
- If `Format` is implemented, `%s`, `%v`, and `%q` must be safe and masked

### YAML / CSV

- Default textual forms must be masked
- If unmarshaling is supported, accept raw input for internal workflows

## 7. Validation Strategy

### Minimum Validation

The first implementation should reject:

1. empty strings
2. strings that become empty after trimming whitespace

The first implementation should also normalize accepted values by converting them to uppercase before storage.

### Deferred Validation

If an authoritative tracking-code format is later confirmed from the website/database layer, the constructors may be tightened in a follow-up change.

## 8. Database Behavior

If database helpers are implemented:

- `Scan()` must load the raw value from DB results
- `Value()` must return the raw value for inserts/updates
- textual methods must remain masked even after `Scan()`

This gives safe defaults in logs and API serialization without breaking persistence.

## 9. Acceptance Criteria

- **AC-001**: Given `ABCD`, when `String()` is called, then it returns `A***`.
- **AC-002**: Given `ABCDEF`, when `Masked()` is called, then it returns `A*****`.
- **AC-003**: Given a string shorter than 4 characters after trimming, when `NewTrackingCode(input)` is called, then it returns an error.
- **AC-003a**: Given `abCDef`, when `NewTrackingCode("abCDef")` is called, then the stored raw value becomes `ABCDEF` and `String()` returns `A*****`.
- **AC-004**: Given a valid tracking code, when `MarshalJSON()` is called, then it serializes the masked value.
- **AC-005**: Given a valid tracking code loaded with `Scan()`, when `Value()` is called, then it returns the raw value.
- **AC-006**: Given a valid tracking code, when `fmt.Sprintf("%v", code)` is called, then the raw value is not present.
- **AC-007**: All unit tests for the new type pass.

## 10. Test Cases

Suggested unit-test coverage:

1. constructor rejection of strings shorter than 4 characters
2. constructor success for a 4+ character raw value
3. constructor rejection of empty and whitespace-only values
4. constructor normalization to uppercase
5. `Masked()` for 4-character and longer values
6. `String()` returns masked form
7. `Raw()` returns the normalized internal value
8. JSON marshal/unmarshal behavior
9. XML marshal/unmarshal behavior
10. YAML marshal/unmarshal behavior if implemented
11. CSV marshal/unmarshal behavior if implemented
12. `fmt.Sprintf("%s")`, `%v`, and `%q` stay masked if `Format` is implemented
13. `Scan()` and `Value()` preserve normalized raw DB persistence while textual output remains masked

## 11. Rationale

Using a dedicated `TrackingCode` type is safer than passing raw `string` values through the codebase because:

1. the masking rule becomes centralized
2. accidental exposure through logs and serialization becomes much less likely
3. future validation rules can be added without changing call sites
4. the type fits the same domain-value-object pattern already used by `GeokretId`

## 12. Implementation Notes

- Prefer the package path `geokrety/geokrety` to keep the type close to `GeokretId`.
- If a raw accessor is added, document clearly that it is for internal/trusted use only.
- If there is any uncertainty about format validation, keep version 1 focused on masking and safe output rather than guessing a regex.

## 13. Next Step

This specification is intended to drive the next implementation task for `U03`: add `tracking_code.go` and `tracking_code_test.go` to the shared `geokrety` module and update any internal API/domain usage that should rely on the new safe type.
