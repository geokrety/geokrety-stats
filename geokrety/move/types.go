package move

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const invalidMoveTypeFormatReason = "invalid move type format; expected registered label or integer id"

// UnknownTypeName is the fallback label returned for unsupported move types.
const UnknownTypeName = "Unknown"

// Move type IDs as stored in the source GeoKrety data.
const (
	MoveTypeDropped   int16 = 0
	MoveTypeGrabbed   int16 = 1
	MoveTypeCommented int16 = 2
	MoveTypeSeen      int16 = 3
	MoveTypeArchived  int16 = 4
	MoveTypeDipped    int16 = 5
)

// MoveTypeRegistry resolves move type IDs to their human-readable labels.
type MoveTypeRegistry struct {
	types     map[int16]string
	labels    map[string]int16
	ordered   []int16
	unknown   string
	formatErr string
}

// MoveTypeError describes invalid move type parsing or encoding input.
type MoveTypeError struct {
	Input  string
	Reason string
}

// MoveType stores a validated move type ID and provides serialization helpers.
type MoveType struct {
	value int16
	valid bool
}

// TypeRegistry is the backward-compatible alias for MoveTypeRegistry.
type TypeRegistry = MoveTypeRegistry

// Type is the backward-compatible alias for MoveType.
type Type = MoveType

// TypeError is the backward-compatible alias for MoveTypeError.
type TypeError = MoveTypeError

// DefaultMoveTypeRegistry exposes the shared move type registry.
var DefaultMoveTypeRegistry = NewMoveTypeRegistry()

// TypeName returns the label for a move type ID.
// Deprecated: Use DefaultMoveTypeRegistry.Name(typeID) instead.
func TypeName(typeID int16) string {
	return DefaultMoveTypeRegistry.Name(typeID)
}

// MoveTypeName returns the label for a move type ID.
// Deprecated: Use DefaultMoveTypeRegistry.Name(typeID) instead.
func MoveTypeName(typeID int16) string {
	return TypeName(typeID)
}

func (e *MoveTypeError) Error() string {
	if strings.TrimSpace(e.Input) == "" {
		return e.Reason
	}
	return fmt.Sprintf("invalid move type %q: %s", e.Input, e.Reason)
}

// NewMoveTypeRegistry builds the default move type-label registry.
func NewMoveTypeRegistry() *MoveTypeRegistry {
	entries := []struct {
		id    int16
		label string
	}{
		{id: MoveTypeDropped, label: "Dropped"},
		{id: MoveTypeGrabbed, label: "Grabbed"},
		{id: MoveTypeCommented, label: "Commented"},
		{id: MoveTypeSeen, label: "Seen"},
		{id: MoveTypeArchived, label: "Archived"},
		{id: MoveTypeDipped, label: "Dipped"},
	}

	registry := &MoveTypeRegistry{
		types:     make(map[int16]string, len(entries)),
		labels:    make(map[string]int16, len(entries)),
		ordered:   make([]int16, 0, len(entries)),
		unknown:   UnknownTypeName,
		formatErr: invalidMoveTypeFormatReason,
	}

	for _, entry := range entries {
		registry.types[entry.id] = entry.label
		registry.labels[entry.label] = entry.id
		registry.ordered = append(registry.ordered, entry.id)
	}

	return registry
}

// NewType parses a label or integer string into a validated move type.
func NewType(raw string) (*MoveType, error) {
	return DefaultMoveTypeRegistry.Parse(raw)
}

// NewNullableType parses a label or integer string unless it is blank.
func NewNullableType(raw string) (*MoveType, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	return NewType(raw)
}

// TypeFromInt validates a raw integer move type ID.
func TypeFromInt(value int16) (*MoveType, error) {
	return DefaultMoveTypeRegistry.FromInt(value)
}

// Name returns the human-readable label for a move type ID.
func (r *MoveTypeRegistry) Name(typeID int16) string {
	if name, ok := r.types[typeID]; ok {
		return name
	}
	return r.unknown
}

// IsValid reports whether the provided move type ID exists in the registry.
func (r *MoveTypeRegistry) IsValid(typeID int16) bool {
	_, ok := r.types[typeID]
	return ok
}

// All returns a defensive copy of the registered move type labels.
func (r *MoveTypeRegistry) All() map[int16]string {
	copyOfTypes := make(map[int16]string, len(r.types))
	for _, id := range r.ordered {
		copyOfTypes[id] = r.types[id]
	}
	return copyOfTypes
}

// Parse converts a label or integer string to a validated move type.
func (r *MoveTypeRegistry) Parse(raw string) (*MoveType, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, &MoveTypeError{Input: raw, Reason: r.formatErr}
	}
	if typeID, ok := r.labels[trimmed]; ok {
		return &MoveType{value: typeID, valid: true}, nil
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 16)
	if err != nil {
		return nil, &MoveTypeError{Input: raw, Reason: r.formatErr}
	}
	return r.FromInt(int16(parsed))
}

// FromInt converts a raw integer type ID to a validated move type.
func (r *MoveTypeRegistry) FromInt(value int16) (*MoveType, error) {
	if !r.IsValid(value) {
		return nil, &MoveTypeError{Input: strconv.FormatInt(int64(value), 10), Reason: fmt.Sprintf("unknown move type id: %d", value)}
	}
	return &MoveType{value: value, valid: true}, nil
}

// EncodeJSON serializes a move type ID as its label.
func (r *MoveTypeRegistry) EncodeJSON(typeID int16) ([]byte, error) {
	if !r.IsValid(typeID) {
		return nil, fmt.Errorf("invalid move type id: %d", typeID)
	}
	return json.Marshal(r.Name(typeID))
}

// DecodeJSON accepts either a string label or an integer ID.
func (r *MoveTypeRegistry) DecodeJSON(data []byte) (int16, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" || trimmed == "" {
		return 0, &MoveTypeError{Input: string(data), Reason: r.formatErr}
	}
	if trimmed[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return 0, err
		}
		parsed, err := r.Parse(raw)
		if err != nil {
			return 0, err
		}
		return parsed.value, nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return 0, &MoveTypeError{Input: string(data), Reason: "invalid move type JSON value; expected string or integer"}
	}
	value, err := number.Int64()
	if err != nil {
		return 0, &MoveTypeError{Input: number.String(), Reason: "move type must be an integer"}
	}
	parsed, err := r.fromInt64(value)
	if err != nil {
		return 0, err
	}
	return parsed.value, nil
}

// EncodeXML serializes a move type ID as element text.
func (r *MoveTypeRegistry) EncodeXML(typeID int16, enc *xml.Encoder, start xml.StartElement) error {
	if !r.IsValid(typeID) {
		return fmt.Errorf("invalid move type id: %d", typeID)
	}
	return enc.EncodeElement(r.Name(typeID), start)
}

// DecodeXML deserializes a move type ID from element text.
func (r *MoveTypeRegistry) DecodeXML(dec *xml.Decoder, start xml.StartElement) (int16, error) {
	var raw string
	if err := dec.DecodeElement(&raw, &start); err != nil {
		return 0, err
	}
	parsed, err := r.Parse(raw)
	if err != nil {
		return 0, err
	}
	return parsed.value, nil
}

// EncodeXMLAttr serializes a move type ID as an XML attribute.
func (r *MoveTypeRegistry) EncodeXMLAttr(typeID int16, name xml.Name) (xml.Attr, error) {
	if !r.IsValid(typeID) {
		return xml.Attr{}, fmt.Errorf("invalid move type id: %d", typeID)
	}
	return xml.Attr{Name: name, Value: r.Name(typeID)}, nil
}

// DecodeXMLAttr deserializes a move type ID from an XML attribute.
func (r *MoveTypeRegistry) DecodeXMLAttr(attr xml.Attr) (int16, error) {
	parsed, err := r.Parse(attr.Value)
	if err != nil {
		return 0, err
	}
	return parsed.value, nil
}

// EncodeCSV serializes a move type ID as `ID,Label`.
func (r *MoveTypeRegistry) EncodeCSV(typeID int16) (string, error) {
	if !r.IsValid(typeID) {
		return "", fmt.Errorf("invalid move type id: %d", typeID)
	}
	return fmt.Sprintf("%d,%s", typeID, r.Name(typeID)), nil
}

// DecodeCSV accepts `ID`, `Label`, or `ID,Label` input.
func (r *MoveTypeRegistry) DecodeCSV(csvLine string) (int16, error) {
	parts := strings.Split(csvLine, ",")
	if len(parts) == 1 {
		parsed, err := r.Parse(parts[0])
		if err != nil {
			return 0, err
		}
		return parsed.value, nil
	}
	if len(parts) != 2 {
		return 0, &MoveTypeError{Input: csvLine, Reason: "invalid CSV format for move type"}
	}

	idPart := strings.TrimSpace(parts[0])
	labelPart := strings.TrimSpace(parts[1])
	if idPart != "" {
		parsed, err := r.Parse(idPart)
		if err == nil {
			if labelPart != "" && r.Name(parsed.value) != labelPart {
				return 0, &MoveTypeError{Input: csvLine, Reason: "move type CSV label does not match id"}
			}
			return parsed.value, nil
		}
	}
	if labelPart != "" {
		parsed, err := r.Parse(labelPart)
		if err != nil {
			return 0, err
		}
		return parsed.value, nil
	}
	return 0, &MoveTypeError{Input: csvLine, Reason: "invalid CSV format for move type"}
}

// EncodeYAML serializes a move type ID as `{id, label}`.
func (r *MoveTypeRegistry) EncodeYAML(typeID int16) (any, error) {
	if !r.IsValid(typeID) {
		return nil, fmt.Errorf("invalid move type id: %d", typeID)
	}
	return map[string]any{"id": typeID, "label": r.Name(typeID)}, nil
}

// DecodeYAML accepts YAML scalar or `{id, label}` forms.
func (r *MoveTypeRegistry) DecodeYAML(data []byte) (int16, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return 0, err
	}
	parsed, err := r.parseYAMLNode(&node)
	if err != nil {
		return 0, err
	}
	return parsed.value, nil
}

// ID returns the raw integer move type ID, or zero when invalid.
func (t *MoveType) ID() int16 {
	if t == nil || !t.valid {
		return 0
	}
	return t.value
}

// Valid reports whether the move type holds a known value.
func (t *MoveType) Valid() bool {
	return t != nil && t.valid
}

// Name returns the label for the type or `Unknown` when invalid.
func (t *MoveType) Name() string {
	if t == nil || !t.valid {
		return UnknownTypeName
	}
	return DefaultMoveTypeRegistry.Name(t.value)
}

// String renders the label for a valid type and a diagnostic marker otherwise.
func (t *MoveType) String() string {
	switch {
	case t == nil:
		return "nil"
	case !t.valid:
		return "invalid"
	default:
		return t.Name()
	}
}

// MarshalJSON serializes the type as its label.
func (t MoveType) MarshalJSON() ([]byte, error) {
	if !t.valid {
		return nil, fmt.Errorf("cannot marshal invalid move type")
	}
	return DefaultMoveTypeRegistry.EncodeJSON(t.value)
}

// UnmarshalJSON accepts a string label or integer ID.
func (t *MoveType) UnmarshalJSON(data []byte) error {
	if strings.TrimSpace(string(data)) == "null" {
		t.value = 0
		t.valid = false
		return nil
	}
	typeID, err := DefaultMoveTypeRegistry.DecodeJSON(data)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalText serializes the type as its label.
func (t MoveType) MarshalText() ([]byte, error) {
	if !t.valid {
		return nil, fmt.Errorf("cannot marshal invalid move type")
	}
	return []byte(t.Name()), nil
}

// UnmarshalText accepts a label or integer ID.
func (t *MoveType) UnmarshalText(text []byte) error {
	parsed, err := DefaultMoveTypeRegistry.Parse(string(text))
	if err != nil {
		return err
	}
	t.value = parsed.value
	t.valid = true
	return nil
}

// MarshalXML serializes the type as element text.
func (t MoveType) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if !t.valid {
		return fmt.Errorf("cannot marshal invalid move type")
	}
	return DefaultMoveTypeRegistry.EncodeXML(t.value, enc, start)
}

// UnmarshalXML deserializes the type from element text.
func (t *MoveType) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	typeID, err := DefaultMoveTypeRegistry.DecodeXML(dec, start)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalXMLAttr serializes the type as an XML attribute.
func (t MoveType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if !t.valid {
		return xml.Attr{}, fmt.Errorf("cannot marshal invalid move type")
	}
	return DefaultMoveTypeRegistry.EncodeXMLAttr(t.value, name)
}

// UnmarshalXMLAttr deserializes the type from an XML attribute.
func (t *MoveType) UnmarshalXMLAttr(attr xml.Attr) error {
	typeID, err := DefaultMoveTypeRegistry.DecodeXMLAttr(attr)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalCSV serializes the type as `ID,Label`.
func (t MoveType) MarshalCSV() (string, error) {
	if !t.valid {
		return "", fmt.Errorf("cannot marshal invalid move type")
	}
	return DefaultMoveTypeRegistry.EncodeCSV(t.value)
}

// UnmarshalCSV accepts `ID`, `Label`, or `ID,Label` input.
func (t *MoveType) UnmarshalCSV(csvLine string) error {
	typeID, err := DefaultMoveTypeRegistry.DecodeCSV(csvLine)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalYAML serializes the type as `{id, label}`.
func (t MoveType) MarshalYAML() (any, error) {
	if !t.valid {
		return nil, fmt.Errorf("cannot marshal invalid move type")
	}
	return DefaultMoveTypeRegistry.EncodeYAML(t.value)
}

// UnmarshalYAML accepts YAML scalar or `{id, label}` forms.
func (t *MoveType) UnmarshalYAML(node *yaml.Node) error {
	parsed, err := DefaultMoveTypeRegistry.parseYAMLNode(node)
	if err != nil {
		return err
	}
	t.value = parsed.value
	t.valid = true
	return nil
}

func (r *MoveTypeRegistry) parseYAMLNode(node *yaml.Node) (*MoveType, error) {
	current := unwrapYAMLNode(node)
	if current == nil {
		return nil, &MoveTypeError{Reason: "invalid YAML structure for move type"}
	}

	switch current.Kind {
	case yaml.ScalarNode:
		return r.Parse(current.Value)
	case yaml.MappingNode:
		var idValue *MoveType
		var labelValue *MoveType
		for i := 0; i+1 < len(current.Content); i += 2 {
			key := strings.TrimSpace(current.Content[i].Value)
			valueNode := unwrapYAMLNode(current.Content[i+1])
			switch key {
			case "id":
				parsed, err := r.Parse(valueNode.Value)
				if err != nil {
					return nil, err
				}
				idValue = parsed
			case "label":
				parsed, err := r.Parse(valueNode.Value)
				if err != nil {
					return nil, err
				}
				labelValue = parsed
			}
		}
		if idValue == nil && labelValue == nil {
			return nil, &MoveTypeError{Reason: "invalid YAML structure for move type"}
		}
		if idValue != nil && labelValue != nil && idValue.value != labelValue.value {
			return nil, &MoveTypeError{Reason: "move type YAML id and label do not match"}
		}
		if idValue != nil {
			return idValue, nil
		}
		return labelValue, nil
	default:
		return nil, &MoveTypeError{Reason: fmt.Sprintf("invalid YAML node kind %d for move type", current.Kind)}
	}
}

func unwrapYAMLNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	current := node
	for current.Kind == yaml.DocumentNode && len(current.Content) > 0 {
		current = current.Content[0]
	}
	return current
}

func (r *MoveTypeRegistry) fromInt64(value int64) (*MoveType, error) {
	if value < -32768 || value > 32767 {
		return nil, &MoveTypeError{Input: strconv.FormatInt(value, 10), Reason: "move type id is out of int16 range"}
	}
	return r.FromInt(int16(value))
}
