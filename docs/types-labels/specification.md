---
title: Type-Label Helpers Refactoring - GeoKrety Stats Go API
version: 1.0
date_created: 2026-03-22
last_updated: 2026-03-22
owner: GeoKrety Stats Development Team
tags: [design, architecture, go, refactoring, idioms]
---

# Introduction

This specification defines the requirements, design patterns, and implementation strategy for refactoring type-label helper functions in the GeoKrety Stats Go API. The refactoring moves away from switch-statement-based implementations toward idiomatic Go patterns using structs and methods, aligning with existing best practices established in the codebase (e.g., the `GeokretId` type in `geokrety/gkid.go`).

The refactoring scope includes:

- `geokrety/geokrety/types.go`: `TypeName()` function for GeoKret types
- `geokrety/move/types.go`: `TypeName()` function for move types
- All usages and tests throughout the codebase
- Supporting infrastructure (lookup maps, error handling, json/xml/csv/yaml serialization)

## 1. Purpose & Scope

### Purpose

To improve code maintainability, extensibility, and adherence to Go idioms by refactoring type-label resolution functions from imperative switch statements to declarative, struct-driven implementations. This enables:

- **Extensibility**: Adding new types without modifying switch statements or core functions
- **Testability**: Easier unit testing with dependency injection support
- **Type Safety**: Compile-time detection of undefined types through constants
- **Consistency**: Alignment with established patterns in the codebase (e.g., `GeokretId` struct)
- **Separation of Concerns**: Clear separation between type definitions and label resolution logic
- **Data Format Support**: JSON, XML, CSV, and YAML marshaling/unmarshaling for API responses and bulk operations

### Scope

This specification covers:

1. **Refactoring targets:**
   - `geokrety/geokrety/types.go`: GeoKret type labels (Traditional, Book/CD/DVD, etc.)
   - `geokrety/move/types.go`: Move type labels (Dropped, Grabbed, Commented, etc.)

2. **Refactoring impact:**
   - All usages of `TypeName()` functions (currently limited to unit tests)
   - Export of refactored functions as part of the public API if needed
   - Documentation updates (godoc comments, OpenAPI specs)
   - JSON/XML/CSV/YAML serialization support for type labels in API responses

3. **Out of scope:**
   - Changes to database schema or constants
   - API endpoint signature changes (if `TypeName` is exported and used by clients)
   - Other utility functions or type helpers beyond the specified scope

### Intended Audience

- Go developers implementing and maintaining the GeoKrety Stats API
- Developers adding new GeoKret or move types in the future
- Code reviewers evaluating refactoring adherence to Go idioms
- AI/LLM-based code generators implementing the specification

### Assumptions

1. The `int16` type is appropriate for type IDs across both GeoKret and move types (no overflow issues expected with current or foreseeable type counts).
2. Unknown type IDs should return "Unknown" for backward compatibility and graceful degradation.
3. The refactoring should be backward-compatible with existing callers.
4. Go 1.18+ is available (supports generics, if needed for advanced patterns).
5. External packages for YAML support (`gopkg.in/yaml.v3`) and CSV handling are available or will be added as dependencies.

## 2. Definitions

| Term | Definition |
|------|-----------|
| **GeoKret Type** | A categorization of physical GeoKrety items (e.g., Traditional, Coin, Car). Type IDs are integers in the range 0-10 (currently). |
| **Move Type** | A categorization of actions performed on GeoKrety in the field (e.g., Dropped, Grabbed, Dipped). Type IDs are integers in the range 0-5 (currently). |
| **TypeName** | A function that converts an integer type ID to its human-readable string label (e.g., `TypeName(0)` → "Traditional"). |
| **Type Registry** | A data structure (struct with a map or other lookup mechanism) that encapsulates type ID-to-label mapping. |
| **Idiomatic Go** | Coding patterns and conventions that follow Go's design principles, as outlined in "Effective Go" and the Go community standards (e.g., explicit error handling, composition over inheritance, interface-driven design). |
| **Switch Statement** | An imperative control flow structure used in the current implementation; should be replaced with a declarative approach. |
| **Unknown Type** | A type ID not present in the registry; should return "Unknown" as a default value for graceful degradation. |
| **Marshaling** | Serialization of type data to different formats (JSON, XML, CSV, YAML) for API responses and data exchange. |
| **Unmarshaling** | Deserialization of type data from various formats back to type ID values for request parsing. |

## 3. Requirements, Constraints & Guidelines

### Functional Requirements

- **REQ-001**: The refactored implementation must support all existing GeoKret type IDs (0-10) with their current labels.
- **REQ-002**: The refactored implementation must support all existing move type IDs (0-5) with their current labels.
- **REQ-003**: Unknown or invalid type IDs must return the string "Unknown" without raising errors.
- **REQ-004**: The refactored functions must maintain backward compatibility with all existing callers (same signature, same return values).
- **REQ-005**: The implementation must provide a mechanism to query available types (for testing, logging, or documentation purposes).
- **REQ-006**: The implementation must support future addition of new types without code modifications to switch statements or core functions.
- **REQ-007**: Type labels must be retrievable efficiently (O(1) lookups via maps or constants).
- **REQ-008**: Type IDs and labels must support marshaling/unmarshaling in JSON format, following the same pattern as `GeokretId` in `gkid.go`.
- **REQ-009**: Type IDs and labels must support marshaling/unmarshaling in XML format, following the same pattern as `GeokretId`.
- **REQ-010**: Type IDs must support marshaling/unmarshaling in CSV format for bulk exports/imports.
- **REQ-011**: Type IDs of registries must support marshaling/unmarshaling in YAML format for configuration and documentation purposes.

### Design & Architecture Requirements

- **DES-001**: Use a struct-based type registry (e.g., `TypeRegistry`, `TypeMap`) to encapsulate type ID-to-label mapping, following the pattern established in `geokrety/gkid.go`.
- **DES-002**: Implement a receiver method on the struct (e.g., `(*TypeRegistry).Name(id int16) string`) for label lookup, replacing the package-level function.
- **DES-003**: Use a map (`map[int16]string`) or similar data structure for efficient label resolution.
- **DES-004**: Support initialization of the type registry via a constructor function (e.g., `NewGeokretTypeRegistry()`, `NewMoveTypeRegistry()`).
- **DES-005**: Make the type registry instance either a package-level variable (singleton) or injectable via function parameters, depending on testability needs.
- **DES-006**: Provide constants for all valid type IDs to enable compile-time type checking and avoid magic numbers.
- **DES-007**: Include comprehensive godoc comments on the struct and methods, documenting the purpose, usage, and any edge cases.
- **DES-008**: Implement `MarshalJSON()` and `UnmarshalJSON()` methods for JSON serialization/deserialization, mirroring the behavior of `GeokretId`.
- **DES-009**: Implement `MarshalXML()`/`UnmarshalXML()` and `MarshalXMLAttr()`/`UnmarshalXMLAttr()` methods for XML support.
- **DES-010**: Implement custom CSV marshaling handlers (functions or methods) to convert type IDs and labels to/from CSV format (delimited strings or numeric values).
- **DES-011**: Implement custom YAML marshaling handlers to support YAML serialization/deserialization of type registries and individual type IDs.

### Code Quality & Testing Requirements

- **QUA-001**: Refactored code must follow Go naming conventions and idioms (e.g., no leading `I` for interfaces that aren't needed, clear error handling).
- **QUA-002**: All refactored functions must have unit tests with minimum 100% code coverage.
- **QUA-003**: Unit tests must cover all valid type IDs and at least one invalid/unknown type ID.
- **QUA-004**: Tests should be named descriptively and use table-driven test patterns where appropriate.
- **QUA-005**: All existing tests must pass with the refactored implementation without modification to test logic.

### Constraints

- **CON-001**: The refactored implementation must not introduce breaking changes to the public API (function signatures must remain identical or be deprecation-wrapped).
- **CON-002**: Performance must not degrade; map lookups must be O(1) or better.
- **CON-003**: The implementation must be compatible with the existing Go version used in the project (1.18+).
- **CON-004**: Type labels must exactly match the strings currently returned by the switch-based implementation to ensure backward compatibility.

### Guidelines & Best Practices

- **GUD-001**: Follow the struct-with-receiver-method pattern established in `geokrety/gkid.go` for consistency across the codebase.
- **GUD-002**: Use constants or enums (const blocks) for type IDs to avoid magic numbers in the codebase.
- **GUD-003**: Document why a particular type ID exists and what it represents, especially for less obvious types (e.g., "Hidden GeoKret").
- **GUD-004**: Consider exporting the type registry if it's useful for external callers (e.g., API responses listing valid types); document the export carefully.
- **GUD-005**: Use table-driven tests for comprehensive coverage and maintainability.
- **GUD-006**: Include examples in godoc comments showing how to use the new struct and methods.

## 4. Interfaces & Data Contracts

### GeoKret Type Registry

#### Struct Definition

```go
// GeokretTypeRegistry encapsulates the mapping of GeoKret type IDs to human-readable labels.
// It provides type-safe access to type information and supports future extensibility.
type GeokretTypeRegistry struct {
	types map[int16]string
}

// Type ID Constants
const (
	GeokretTypeTraditional    int16 = 0
	GeokretTypeBook           int16 = 1
	GeokretTypeHumanPet       int16 = 2
	GeokretTypeCoin           int16 = 3
	GeokretTypeKretyPost      int16 = 4
	GeokretTypePebble         int16 = 5
	GeokretTypeCar            int16 = 6
	GeokretTypePlayingCard    int16 = 7
	GeokretTypeDogTagPet      int16 = 8
	GeokretTypeJigsawPart     int16 = 9
	GeokretTypeHidden         int16 = 10
)
```

#### Constructor Function

```go
// NewGeokretTypeRegistry creates a new instance of GeokretTypeRegistry with all defined types.
// Returns a pointer to the registry instance.
func NewGeokretTypeRegistry() *GeokretTypeRegistry {
	return &GeokretTypeRegistry{
		types: map[int16]string{
			GeokretTypeTraditional:    "Traditional",
			GeokretTypeBook:           "Book/CD/DVD...",
			GeokretTypeHumanPet:       "Human/Pet",
			GeokretTypeCoin:           "Coin",
			GeokretTypeKretyPost:      "KretyPost",
			GeokretTypePebble:         "Pebble",
			GeokretTypeCar:            "Car",
			GeokretTypePlayingCard:    "Playing card",
			GeokretTypeDogTagPet:      "Dog tag/pet",
			GeokretTypeJigsawPart:     "Jigsaw part",
			GeokretTypeHidden:         "Hidden GeoKret",
		},
	}
}
```

#### Receiver Methods

```go
// Name returns the human-readable label for a GeoKret type ID.
// If the type ID is unknown, it returns "Unknown".
//
// Example:
//   registry := NewGeokretTypeRegistry()
//   fmt.Println(registry.Name(0))  // Output: "Traditional"
func (r *GeokretTypeRegistry) Name(typeID int16) string {
	if name, ok := r.types[typeID]; ok {
		return name
	}
	return "Unknown"
}

// IsValid checks whether a type ID is known to the registry.
func (r *GeokretTypeRegistry) IsValid(typeID int16) bool {
	_, ok := r.types[typeID]
	return ok
}

// All returns a copy of the map of all registered types for iteration or inspection.
// Returns a defensive copy to prevent external mutations of the internal mapping.
func (r *GeokretTypeRegistry) All() map[int16]string {
	result := make(map[int16]string, len(r.types))
	for k, v := range r.types {
		result[k] = v
	}
	return result
}

// MarshalJSON serializes a type ID to its JSON representation (the type label string).
// Returns an error if the type ID is unknown or invalid.
func (r *GeokretTypeRegistry) MarshalJSON(typeID int16) ([]byte, error) {
	if !r.IsValid(typeID) {
		return nil, fmt.Errorf("invalid geokrety type ID: %d", typeID)
	}
	return json.Marshal(r.Name(typeID))
}

// UnmarshalJSON deserializes a JSON value (string or integer) into a type ID.
// Accepts both type label strings (e.g., "Traditional", "Coin") and integer IDs.
func (r *GeokretTypeRegistry) UnmarshalJSON(data []byte) (int16, error) {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, err
	}
	switch v := raw.(type) {
	case string:
		// Look up type ID by label
		for id, label := range r.types {
			if label == v {
				return id, nil
			}
		}
		return 0, fmt.Errorf("unknown geokrety type: %q", v)
	case float64:
		typeID := int16(v)
		if !r.IsValid(typeID) {
			return 0, fmt.Errorf("unknown geokrety type ID: %d", typeID)
		}
		return typeID, nil
	default:
		return 0, fmt.Errorf("invalid type for geokrety type: %T", raw)
	}
}

// MarshalXML serializes a type ID to its XML representation (the type label string).
func (r *GeokretTypeRegistry) MarshalXML(typeID int16, enc *xml.Encoder, start xml.StartElement) error {
	if !r.IsValid(typeID) {
		return fmt.Errorf("invalid geokrety type ID: %d", typeID)
	}
	label := r.Name(typeID)
	return enc.EncodeElement(label, start)
}

// UnmarshalXML deserializes an XML value (element or attribute) into a type ID.
func (r *GeokretTypeRegistry) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) (int16, error) {
	var raw string
	if err := dec.DecodeElement(&raw, &start); err != nil {
		return 0, err
	}
	for id, label := range r.types {
		if label == raw {
			return id, nil
		}
	}
	return 0, fmt.Errorf("unknown geokrety type: %q", raw)
}

// UnmarshalXMLAttr deserializes an XML attribute value into a type ID.
func (r *GeokretTypeRegistry) UnmarshalXMLAttr(attr xml.Attr) (int16, error) {
	for id, label := range r.types {
		if label == attr.Value {
			return id, nil
		}
	}
	return 0, fmt.Errorf("unknown geokrety type: %q", attr.Value)
}
```

### Move Type Registry

#### Struct Definition

```go
// MoveTypeRegistry encapsulates the mapping of move type IDs to human-readable labels.
// It provides type-safe access to move type information and supports future extensibility.
type MoveTypeRegistry struct {
	types map[int16]string
}

// Type ID Constants
const (
	MoveTypeDropped    int16 = 0
	MoveTypeGrabbed    int16 = 1
	MoveTypeCommented  int16 = 2
	MoveTypeSeen       int16 = 3
	MoveTypeArchived   int16 = 4
	MoveTypeDipped     int16 = 5
)
```

#### Constructor Function

```go
// NewMoveTypeRegistry creates a new instance of MoveTypeRegistry with all defined types.
// Returns a pointer to the registry instance.
func NewMoveTypeRegistry() *MoveTypeRegistry {
	return &MoveTypeRegistry{
		types: map[int16]string{
			MoveTypeDropped:    "Dropped",
			MoveTypeGrabbed:    "Grabbed",
			MoveTypeCommented:  "Commented",
			MoveTypeSeen:       "Seen",
			MoveTypeArchived:   "Archived",
			MoveTypeDipped:     "Dipped",
		},
	}
}
```

#### Receiver Methods

```go
// Name returns the human-readable label for a move type ID.
// If the type ID is unknown, it returns "Unknown".
//
// Example:
//   registry := NewMoveTypeRegistry()
//   fmt.Println(registry.Name(0))  // Output: "Dropped"
func (r *MoveTypeRegistry) Name(typeID int16) string {
	if name, ok := r.types[typeID]; ok {
		return name
	}
	return "Unknown"
}

// IsValid checks whether a type ID is known to the registry.
func (r *MoveTypeRegistry) IsValid(typeID int16) bool {
	_, ok := r.types[typeID]
	return ok
}

// All returns a copy of the map of all registered types for iteration or inspection.
// Returns a defensive copy to prevent external mutations of the internal mapping.
func (r *MoveTypeRegistry) All() map[int16]string {
	result := make(map[int16]string, len(r.types))
	for k, v := range r.types {
		result[k] = v
	}
	return result
}

// MarshalJSON serializes a type ID to its JSON representation (the type label string).
// Returns an error if the type ID is unknown or invalid.
func (r *MoveTypeRegistry) MarshalJSON(typeID int16) ([]byte, error) {
	if !r.IsValid(typeID) {
		return nil, fmt.Errorf("invalid move type ID: %d", typeID)
	}
	return json.Marshal(r.Name(typeID))
}

// UnmarshalJSON deserializes a JSON value (string or integer) into a type ID.
// Accepts both type label strings (e.g., "Dropped", "Grabbed") and integer IDs.
func (r *MoveTypeRegistry) UnmarshalJSON(data []byte) (int16, error) {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, err
	}
	switch v := raw.(type) {
	case string:
		// Look up type ID by label
		for id, label := range r.types {
			if label == v {
				return id, nil
			}
		}
		return 0, fmt.Errorf("unknown move type: %q", v)
	case float64:
		typeID := int16(v)
		if !r.IsValid(typeID) {
			return 0, fmt.Errorf("unknown move type ID: %d", typeID)
		}
		return typeID, nil
	default:
		return 0, fmt.Errorf("invalid type for move type: %T", raw)
	}
}

// MarshalXML serializes a type ID to its XML representation (the type label string).
func (r *MoveTypeRegistry) MarshalXML(typeID int16, enc *xml.Encoder, start xml.StartElement) error {
	if !r.IsValid(typeID) {
		return fmt.Errorf("invalid move type ID: %d", typeID)
	}
	label := r.Name(typeID)
	return enc.EncodeElement(label, start)
}

// UnmarshalXML deserializes an XML value (element or attribute) into a type ID.
func (r *MoveTypeRegistry) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) (int16, error) {
	var raw string
	if err := dec.DecodeElement(&raw, &start); err != nil {
		return 0, err
	}
	for id, label := range r.types {
		if label == raw {
			return id, nil
		}
	}
	return 0, fmt.Errorf("unknown move type: %q", raw)
}

// UnmarshalXMLAttr deserializes an XML attribute value into a type ID.
func (r *MoveTypeRegistry) UnmarshalXMLAttr(attr xml.Attr) (int16, error) {
	for id, label := range r.types {
		if label == attr.Value {
			return id, nil
		}
	}
	return 0, fmt.Errorf("unknown move type: %q", attr.Value)
}
```

### Package-Level Singleton Instances (Optional)

For backward compatibility and convenience, package-level singleton registries may be instantiated:

```go
var (
	// DefaultGeokretTypeRegistry is the default singleton instance of GeokretTypeRegistry.
	DefaultGeokretTypeRegistry = NewGeokretTypeRegistry()

	// DefaultMoveTypeRegistry is the default singleton instance of MoveTypeRegistry.
	DefaultMoveTypeRegistry = NewMoveTypeRegistry()
)
```

A backward-compatible wrapper function can be provided:

```go
// TypeName returns the human-readable label for a GeoKret type ID (for backward compatibility).
// Deprecated: Use DefaultGeokretTypeRegistry.Name(typeID) instead.
func TypeName(typeID int16) string {
	return DefaultGeokretTypeRegistry.Name(typeID)
}
```

### File Structure & Placement

The refactored structs and functions should be organized as follows:

```
geokrety/
├── geokrety/
│   ├── types.go           (contains GeokretTypeRegistry, constants, constructor, and methods)
│   └── types_test.go      (contains unit tests for GeokretTypeRegistry)
└── move/
    ├── types.go           (contains MoveTypeRegistry, constants, constructor, and methods)
    └── types_test.go      (contains unit tests for MoveTypeRegistry)
```

**Placement guidelines:**
- **GeokretTypeRegistry**: All struct definition, constants, and methods in `geokrety/geokrety/types.go`
- **MoveTypeRegistry**: All struct definition, constants, and methods in `geokrety/move/types.go`
- **Singleton instances**: Include package-level singletons in the same file as their struct definitions (optional)
- **Backward-compatible wrapper**: Can be placed in the same file or in a separate utilities file, depending on codebase conventions (optional)
- **Tests**: Follow Go naming convention (`types_test.go`); place in the same directory as the source file

### CSV and YAML Marshaling Support

For CSV and YAML formats, implement helper functions or packages at module level:

```go
// CSV Marshaling Support
// MarshalCSV converts a type ID to its CSV representation (comma or pipe-delimited format)
// Example: "0,Traditional" for GeoKret types
func (r *GeokretTypeRegistry) MarshalCSV(typeID int16) (string, error) {
	if !r.IsValid(typeID) {
		return "", fmt.Errorf("invalid geokrety type ID: %d", typeID)
	}
	return fmt.Sprintf("%d,%s", typeID, r.Name(typeID)), nil
}

// UnmarshalCSV parses a CSV line into a type ID
// Accepts both ID and label (ID,Label format or just ID or just Label)
func (r *GeokretTypeRegistry) UnmarshalCSV(csvLine string) (int16, error) {
	parts := strings.Split(csvLine, ",")
	if len(parts) == 1 {
		// Try as integer ID first
		if id, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 16); err == nil {
			typeID := int16(id)
			if r.IsValid(typeID) {
				return typeID, nil
			}
		}
		// Try as label
		trimmed := strings.TrimSpace(parts[0])
		for id, label := range r.types {
			if label == trimmed {
				return id, nil
			}
		}
		return 0, fmt.Errorf("unknown geokrety type: %q", parts[0])
	}
	if len(parts) >= 2 {
		if id, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 16); err == nil {
			return int16(id), nil
		}
	}
	return 0, fmt.Errorf("invalid CSV format for geokrety type: %q", csvLine)
}

// YAML Marshaling Support (using gopkg.in/yaml.v3 or similar)
// MarshalYAML converts a type ID to its YAML representation
func (r *GeokretTypeRegistry) MarshalYAML(typeID int16) (interface{}, error) {
	if !r.IsValid(typeID) {
		return nil, fmt.Errorf("invalid geokrety type ID: %d", typeID)
	}
	return map[string]interface{}{
		"id":    typeID,
		"label": r.Name(typeID),
	}, nil
}

// UnmarshalYAML parses a YAML node into a type ID
// Accepts both integer ID and string label formats
func (r *GeokretTypeRegistry) UnmarshalYAML(data []byte) (int16, error) {
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return 0, err
	}
	switch v := raw.(type) {
	case int:
		typeID := int16(v)
		if !r.IsValid(typeID) {
			return 0, fmt.Errorf("unknown geokrety type ID: %d", typeID)
		}
		return typeID, nil
	case string:
		// Look up type ID by label
		for id, label := range r.types {
			if label == v {
				return id, nil
			}
		}
		return 0, fmt.Errorf("unknown geokrety type: %q", v)
	case map[string]interface{}:
		// Support structured YAML: {id: 0, label: "Traditional"}
		if idVal, ok := v["id"]; ok {
			if id, ok := idVal.(int); ok {
				typeID := int16(id)
				if r.IsValid(typeID) {
					return typeID, nil
				}
			}
		}
		return 0, fmt.Errorf("invalid YAML structure for geokrety type")
	default:
		return 0, fmt.Errorf("invalid type for geokrety type: %T", raw)
	}
}
```

### Singleton Pattern Guidance

The singleton pattern is **optional** and should be used based on the following criteria:

1. **Use singletons** (`DefaultGeokretTypeRegistry`, `DefaultMoveTypeRegistry`) if:
   - Most callers need a registry instance and would benefit from a convenient, pre-initialized default
   - The overhead of creating a new registry instance for every call is undesirable
   - Backward compatibility with existing `TypeName()` function is required

2. **Avoid singletons** if:
   - Dependency injection patterns are preferred in the codebase
   - Tests require registry instances to be injected (testability is paramount)
   - Different registry configurations may be needed in the future

3. **Recommended approach** for this codebase:
   - Provide both singleton instances (for convenience) and support for constructor injection
   - Document both usage patterns in godoc comments
   - In tests, explicitly instantiate registries (don't rely on singletons) for testability

## 5. Acceptance Criteria

- **AC-001**: Given a GeoKret type ID 0, When `GeokretTypeRegistry.Name(0)` is called, Then it returns "Traditional".
- **AC-002**: Given a GeoKret type ID 10, When `GeokretTypeRegistry.Name(10)` is called, Then it returns "Hidden GeoKret".
- **AC-003**: Given an unknown GeoKret type ID (e.g., 99), When `GeokretTypeRegistry.Name(99)` is called, Then it returns "Unknown".
- **AC-004**: Given a move type ID 5, When `MoveTypeRegistry.Name(5)` is called, Then it returns "Dipped".
- **AC-005**: Given a move type ID 0, When `MoveTypeRegistry.Name(0)` is called, Then it returns "Dropped".
- **AC-006**: Given an unknown move type ID (e.g., 99), When `MoveTypeRegistry.Name(99)` is called, Then it returns "Unknown".
- **AC-007**: Given a `GeokretTypeRegistry` instance, When `IsValid(0)` is called, Then it returns `true`; When `IsValid(99)` is called, Then it returns `false`.
- **AC-008**: Given a `MoveTypeRegistry` instance, When `IsValid(5)` is called, Then it returns `true`; When `IsValid(99)` is called, Then it returns `false`.
- **AC-009**: All existing unit tests must pass without modification to test logic or structure.
- **AC-010**: The refactored code must achieve 100% code coverage for the new registry struct and methods.
- **AC-011**: Type ID constants (e.g., `GeokretTypeTraditional`, `MoveTypeDropped`) must be compile-time accessible and correctly valued.
- **AC-012**: JSON marshaling of type labels must match the format used in `GeokretId` marshaling.
- **AC-013**: XML marshaling must support both element and attribute serialization.
- **AC-014**: CSV and YAML marshaling must handle both ID-based and label-based input.

## 6. Test Automation Strategy

### Test Levels

| Level | Scope | Tools | Coverage |
|-------|-------|-------|----------|
| **Unit** | Individual methods of registry structs and constants | Go `testing` package, `testing/quick` (if applicable) | 100% mandatory |
| **Integration** | Interaction with other packages (if applicable); verification of constants in production code | Go `testing` package | Implicit through unit tests |
| **End-to-End** | API endpoints that use type labels (if exposed); validation through API responses | Go `testing` package; potentially API integration tests | Out of scope for this spec, but should be verified if labels are part of API responses |

### Test Frameworks & Tools

- **Testing Framework**: Go's built-in `testing` package (standard library)
- **Assertions**: Optional use of `github.com/stretchr/testify/assert` for cleaner assertions (if already in project dependencies)
- **Table-Driven Tests**: Recommended pattern for comprehensive test coverage
- **Test File Naming**: Follow Go conventions (`*_test.go`)

### Test Data Management

- **Valid Type IDs**: Use constants defined in the codebase (e.g., `GeokretTypeTraditional`, `MoveTypeDropped`)
- **Invalid Type IDs**: Use values outside the defined range (e.g., -1, 99, 1000, max `int16`)
- **Edge Cases**: Test boundary values (e.g., 0, maximum defined ID, maximum `int16`)

### Test Cases for GeoKretTypeRegistry

```
Table-Driven Tests for Name(), IsValid(), All() method, and JSON/XML/CSV/YAML marshaling:
1. "Traditional" for ID 0
2. "Book/CD/DVD..." for ID 1
3. "Human/Pet" for ID 2
4. "Coin" for ID 3
5. "KretyPost" for ID 4
6. "Pebble" for ID 5
7. "Car" for ID 6
8. "Playing card" for ID 7
9. "Dog tag/pet" for ID 8
10. "Jigsaw part" for ID 9
11. "Hidden GeoKret" for ID 10
12. "Unknown" for ID 99 (invalid)
13. "Unknown" for ID -1 (negative)
14. "Unknown" for ID 32767 (max int16)
15. IsValid returns true for each valid ID
16. IsValid returns false for each invalid ID
17. All() method returns map with all types
18. JSON marshaling produces correct string labels
19. JSON unmarshaling accepts both strings and integers
20. XML marshaling and unmarshaling work correctly (element and attribute)
21. CSV marshaling produces "ID,Label" format
22. CSV unmarshaling accepts ID only, Label only, or "ID,Label" format
23. YAML marshaling produces structured format
24. YAML unmarshaling accepts ID, Label, or structured format
```

### Test Cases for MoveTypeRegistry

```
Table-Driven Tests:
1. "Dropped" for ID 0
2. "Grabbed" for ID 1
3. "Commented" for ID 2
4. "Seen" for ID 3
5. "Archived" for ID 4
6. "Dipped" for ID 5
7. "Unknown" for ID 99 (invalid)
8. "Unknown" for ID -1 (negative)
9. "Unknown" for ID 32767 (max int16)
10. IsValid returns true for each valid ID
11. IsValid returns false for each invalid ID
12. All() method returns map with all types
13-24. JSON, XML, CSV, YAML marshaling tests (same as GeoKret types)
```

### Test Execution & CI/CD Integration

- **Local Execution**: `go test ./geokrety/geokrety` and `go test ./geokrety/move` during development
- **CI/CD Pipeline**: Tests must be integrated into GitHub Actions or equivalent CI/CD system
- **Coverage Reporting**: Collect coverage metrics using `go test -coverprofile=coverage.out` and report to CI/CD dashboard
- **Coverage Thresholds**: Minimum 100% coverage for refactored code; overall project coverage should not decrease

### Performance Testing

- **Benchmark Tests**: Optional but recommended for `Name()` method to ensure O(1) lookup performance
- **Benchmark Example**:
  ```
  BenchmarkGeokretTypeRegistry_Name/valid_id-8
  BenchmarkGeokretTypeRegistry_Name/invalid_id-8
  ```

## 7. Rationale & Context

### Why Refactor from Switch Statements?

1. **Maintainability**: Switch statements are imperative and require reading all branches to understand the complete set of types. A map-based approach is declarative and more readable.
2. **Extensibility**: Adding a new type to a switch statement requires modifying the function; a map-based approach allows initialization-time configuration without function changes.
3. **Consistency**: The codebase already uses struct-based patterns for similar concerns (see `GeokretId` in `gkid.go`), making this refactoring align with established idioms.
4. **Testability**: Struct-based approaches enable dependency injection and easier mocking if needed in the future.

### Why Use a Struct vs. Package-Level Map?

1. **Type Safety**: A struct ensures encapsulation and prevents accidental mutations of the type mapping.
2. **Cohesion**: The struct groups related data and behavior (types and their labels) together, improving code organization.
3. **Flexibility**: A struct can be extended with additional methods (e.g., validation, filtering) without polluting the package namespace.
4. **Testing**: Struct receivers can be more easily mocked or injected in tests.

### Why Use Constants for Type IDs?

1. **Compile-Time Checking**: Constants prevent typos and ensure type IDs are defined correctly at compile time.
2. **Documentation**: Named constants document the meaning and purpose of each type ID.
3. **Refactoring Safety**: If a type ID value changes, the constant can be updated in one place, and the compiler ensures consistency.

### Reference Pattern: GeokretId

The `GeokretId` struct in `geokrety/gkid.go` demonstrates idiomatic Go patterns used as inspiration:
- Uses a struct with a private value field
- Provides constructor functions (`New`, `NewNullable`, `FromInt`)
- Implements receiver methods for behavior
- Includes custom error types for detailed error information
- Provides validation and type-safe conversions
- Implements JSON, XML, and TEXT marshaling for consistent data format support

The refactoring applies similar principles to type-label resolution, maintaining consistency across the codebase.

## 8. Dependencies & External Integrations

### External Systems
- None required; this is a self-contained refactoring of internal utility functions.

### Third-Party Services
- None required for JSON and XML support (Go standard library)
- **YAML Support**: `gopkg.in/yaml.v3` package (may need to be added as a dependency)

### Infrastructure Dependencies
- **Go Runtime**: Version 1.18 or later (already required by the project)

### Data Dependencies
- None; type mappings are hardcoded in the registry and do not depend on external data sources.

### Technology Platform Dependencies
- **Go Standard Library**: Basic types (`map`, `int16`, `string`), JSON/XML encoding, testing (`testing` package)
- **Optional**: YAML library (`gopkg.in/yaml.v3`) for YAML marshaling support

### Compliance Dependencies
- None; this is a code refactoring with no compliance or regulatory impact.

## 9. Examples & Edge Cases

### Example 1: Basic Usage

```go
package main

import (
	"fmt"
	"github.com/geokrety/geokrety-stats/geokrety/geokrety"
	"github.com/geokrety/geokrety-stats/geokrety/move"
)

func main() {
	// Create or use singleton registries
	geokretRegistry := geokrety.NewGeokretTypeRegistry()
	moveRegistry := move.NewMoveTypeRegistry()

	// Look up type labels
	fmt.Println(geokretRegistry.Name(0))   // Output: "Traditional"
	fmt.Println(geokretRegistry.Name(4))   // Output: "KretyPost"
	fmt.Println(geokretRegistry.Name(99))  // Output: "Unknown"

	fmt.Println(moveRegistry.Name(0))     // Output: "Dropped"
	fmt.Println(moveRegistry.Name(5))     // Output: "Dipped"
	fmt.Println(moveRegistry.Name(99))    // Output: "Unknown"
}
```

### Example 2: Validation Before Lookup

```go
func processGeokretType(typeID int16, registry *geokrety.GeokretTypeRegistry) error {
	if !registry.IsValid(typeID) {
		return fmt.Errorf("unknown geokrety type: %d", typeID)
	}
	label := registry.Name(typeID)
	// Use label...
	return nil
}
```

### Example 3: Iterating All Types

```go
func listAllGeokretTypes(registry *geokrety.GeokretTypeRegistry) {
	for typeID, label := range registry.All() {
		fmt.Printf("ID: %d, Label: %s\n", typeID, label)
	}
}
```

### Example 4: JSON Marshaling

```go
func marshalTypeForJSON(typeID int16, registry *geokrety.GeokretTypeRegistry) ([]byte, error) {
	return registry.MarshalJSON(typeID)
}
```

### Example 5: Backward Compatibility Wrapper

```go
// Old code using package-level function
func TestBackwardCompatibility(t *testing.T) {
	// This should still work via the deprecated wrapper
	label := geokrety.TypeName(0)
	if label != "Traditional" {
		t.Fatalf("TypeName(0) = %q, want Traditional", label)
	}
}
```

### Edge Cases & Handling

| Edge Case | Input | Expected Output | Rationale |
|-----------|-------|-----------------|-----------|
| Unknown type ID | 99 | "Unknown" | Graceful degradation; no error thrown for forward compatibility |
| Negative type ID | -1 | "Unknown" | Treat as invalid |
| Maximum int16 | 32767 | "Unknown" | Out of defined range |
| Minimum int16 | -32768 | "Unknown" | Negative and out of range |
| Valid boundary (min) | 0 | "Traditional" | First valid type ID |
| Valid boundary (max GeoKret) | 10 | "Hidden GeoKret" | Last defined GeoKret type |
| Valid boundary (max Move) | 5 | "Dipped" | Last defined move type |
| Concurrent access | Multiple goroutines calling `Name()` | Correct label per ID | Maps are read-only; no synchronization required |
| JSON string input | "Traditional" | Unmarshals to ID 0 | Label-to-ID reverse lookup |
| JSON integer input | 0 | Unmarshals to ID 0 | Direct ID parsing |

## 10. Validation Criteria

To confirm compliance with this specification, the following validation criteria must be checked during and after implementation:

### Code Structure Validation

- [ ] `GeokretTypeRegistry` struct defined in `geokrety/geokrety/types.go` with `types map[int16]string` private field
- [ ] `MoveTypeRegistry` struct defined in `geokrety/move/types.go` with `types map[int16]string` private field
- [ ] Type ID constants defined (11 for GeoKret: `GeokretTypeTraditional` through `GeokretTypeHidden`)
- [ ] Type ID constants defined (6 for Move: `MoveTypeDropped` through `MoveTypeDipped`)
- [ ] Constructor functions defined: `NewGeokretTypeRegistry()` and `NewMoveTypeRegistry()`
- [ ] Receiver methods defined on both structs: `Name()`, `IsValid()`, `All()`
- [ ] `All()` method returns a defensive copy of the type map (not the original mutable map)
- [ ] JSON marshaling methods implemented: `MarshalJSON()`, `UnmarshalJSON()`
- [ ] XML marshaling methods implemented: `MarshalXML()`, `UnmarshalXML()`, `UnmarshalXMLAttr()`
- [ ] CSV marshaling methods implemented: `MarshalCSV()`, `UnmarshalCSV()`
- [ ] YAML marshaling methods implemented: `MarshalYAML()`, `UnmarshalYAML()`
- [ ] Godoc comments present on all exported types and methods

### Backward Compatibility Validation

- [ ] Original `TypeName()` function behavior preserved (via wrapper function or direct replacement)
- [ ] All existing callers of `TypeName()` continue to work without code changes
- [ ] Identical output for all valid type IDs compared to original switch statement implementation
- [ ] Identical behavior for unknown/invalid type IDs (returns "Unknown" string)

### Test Coverage Validation

- [ ] Unit tests exist in `geokrety/geokrety/types_test.go` for `GeokretTypeRegistry`
- [ ] Unit tests exist in `geokrety/move/types_test.go` for `MoveTypeRegistry`
- [ ] All 11 valid GeoKret type IDs tested (0-10)
- [ ] All 6 valid Move type IDs tested (0-5)
- [ ] Invalid type IDs tested (e.g., -1, 99, 32767, -32768)
- [ ] `IsValid()` method tested for both valid and invalid IDs
- [ ] `All()` method tested to verify all types are present and defensive copy works
- [ ] JSON marshaling/unmarshaling tests verified
- [ ] XML marshaling/unmarshaling tests verified
- [ ] CSV marshaling/unmarshaling tests verified
- [ ] YAML marshaling/unmarshaling tests verified
- [ ] Concurrent access tested (optional but recommended: use `testing.T.Parallel()`)
- [ ] Code coverage analysis shows 100% coverage for new registry code: `go test -cover ./geokrety/geokrety ./geokrety/move`

### Code Quality Validation

- [ ] Code passes `go fmt` formatting checks
- [ ] Code passes `go vet` with no warnings or errors
- [ ] No unused code or variables present
- [ ] No magic numbers in production code (use constants)
- [ ] Godoc comments are well-formatted and explain purpose, parameters, and return values
- [ ] Examples included in godoc comments or test files

### Functional Acceptance Testing

**Test execution steps:**

1. **Run unit tests locally:**
   ```bash
   cd geokrety/geokrety && go test -v -cover ./...
   cd geokrety/move && go test -v -cover ./...
   ```

2. **Verify backward compatibility:**
   ```bash
   # Old code should still work:
   label := TypeName(0)
   assert label == "Traditional"
   ```

3. **Verify new API:**
   ```bash
   registry := NewGeokretTypeRegistry()
   label := registry.Name(10)
   assert label == "Hidden GeoKret"
   ```

4. **Verify JSON marshaling:**
   ```bash
   data, err := registry.MarshalJSON(0)
   // Should serialize to: "Traditional"
   ```

5. **Verify constant access:**
   ```bash
   // Constants should be available at compile time
   typeID := GeokretTypeBook
   ```

6. **Build the entire API project successfully:**
   ```bash
   cd api && go build ./...
   ```

### Deployment Readiness Checklist

- [ ] All existing tests pass (no test modifications required)
- [ ] All new tests pass with 100% coverage
- [ ] Code review completed and approved by maintainers
- [ ] Documentation updated (godoc comments, any API documentation)
- [ ] API instructions documentation updated (see below)
- [ ] No performance regressions detected (`go test -bench ./...` optional)
- [ ] Breaking changes addressed (if any) with migration guide
- [ ] Integration with rest of codebase verified (compile, existing functionality)

## 11. Related Specifications & Further Reading

### Codebase References

- **[GeokretId Type Implementation](../../geokrety/geokrety/gkid.go)** — The established pattern for struct-based type handling in the codebase; demonstrates idiomatic Go practices that inspired this refactoring
- **[Existing API Documentation](../endpoint-gkid/specification.md)** — API contracts and type definitions for reference
- **[API Instructions](../../.github/instructions/api.instructions.md)** — Document JSON/XML/CSV/YAML marshaling standards for all API-related type implementations

### External References

- **[Effective Go - Interfaces](https://golang.org/doc/effective_go#interfaces)** — Comprehensive guide to Go interface design and best practices
- **[Effective Go - Errors](https://golang.org/doc/effective_go#errors)** — Error handling patterns and idioms
- **[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)** — Community standards for Go code style and conventions
- **[Go Maps](https://golang.org/doc/effective_go#maps)** — Best practices for working with maps in Go
- **[Package Initialization](https://golang.org/doc/effective_go#init)** — Understanding package-level initialization and singletons

### Related Patterns & Practices

- **Struct-based registries**: Preferred pattern for encapsulating type mappings and related behavior
- **Constants over magic numbers**: Using named constants for type IDs enables compile-time checking and documentation
- **Table-driven tests**: Recommended testing pattern for covering multiple cases concisely
- **Defensive copying**: The `All()` method returns a defensive copy to mitigate unintended mutations
- **Data format marshaling**: JSON, XML, CSV, and YAML support for cross-platform data interchange
