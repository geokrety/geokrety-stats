package geokrety

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const invalidGeokretTypeFormatReason = "invalid geokret type format; expected registered label or integer id"

// UnknownTypeName is the fallback label returned for unsupported GeoKret types.
const UnknownTypeName = "Unknown"

// GeoKret type IDs as stored in the source GeoKrety data.
const (
	GeokretTypeTraditional int16 = 0
	GeokretTypeBook        int16 = 1
	GeokretTypeHumanPet    int16 = 2
	GeokretTypeCoin        int16 = 3
	GeokretTypeKretyPost   int16 = 4
	GeokretTypePebble      int16 = 5
	GeokretTypeCar         int16 = 6
	GeokretTypePlayingCard int16 = 7
	GeokretTypeDogTagPet   int16 = 8
	GeokretTypeJigsawPart  int16 = 9
	GeokretTypeHidden      int16 = 10
)

// GeokretTypeRegistry resolves GeoKret type IDs to their human-readable labels.
type GeokretTypeRegistry struct {
	types     map[int16]string
	labels    map[string]int16
	ordered   []int16
	unknown   string
	formatErr string
}

// GeokretTypeError describes invalid GeoKret type parsing or encoding input.
type GeokretTypeError struct {
	Input  string
	Reason string
}

// GeokretType stores a validated GeoKret type ID and provides serialization helpers.
type GeokretType struct {
	value int16
	valid bool
}

// DefaultGeokretTypeRegistry exposes the shared GeoKret type registry.
var DefaultGeokretTypeRegistry = NewGeokretTypeRegistry()

func (e *GeokretTypeError) Error() string {
	if strings.TrimSpace(e.Input) == "" {
		return e.Reason
	}
	return fmt.Sprintf("invalid geokret type %q: %s", e.Input, e.Reason)
}

// NewGeokretTypeRegistry builds the default GeoKret type-label registry.
func NewGeokretTypeRegistry() *GeokretTypeRegistry {
	entries := []struct {
		id    int16
		label string
	}{
		{id: GeokretTypeTraditional, label: "Traditional"},
		{id: GeokretTypeBook, label: "Book/CD/DVD..."},
		{id: GeokretTypeHumanPet, label: "Human/Pet"},
		{id: GeokretTypeCoin, label: "Coin"},
		{id: GeokretTypeKretyPost, label: "KretyPost"},
		{id: GeokretTypePebble, label: "Pebble"},
		{id: GeokretTypeCar, label: "Car"},
		{id: GeokretTypePlayingCard, label: "Playing card"},
		{id: GeokretTypeDogTagPet, label: "Dog tag/pet"},
		{id: GeokretTypeJigsawPart, label: "Jigsaw part"},
		{id: GeokretTypeHidden, label: "Hidden GeoKret"},
	}

	registry := &GeokretTypeRegistry{
		types:     make(map[int16]string, len(entries)),
		labels:    make(map[string]int16, len(entries)),
		ordered:   make([]int16, 0, len(entries)),
		unknown:   UnknownTypeName,
		formatErr: invalidGeokretTypeFormatReason,
	}

	for _, entry := range entries {
		registry.types[entry.id] = entry.label
		registry.labels[entry.label] = entry.id
		registry.ordered = append(registry.ordered, entry.id)
	}

	return registry
}

// NewType parses a label or integer string into a validated GeoKret type.
func NewType(raw string) (*GeokretType, error) {
	return DefaultGeokretTypeRegistry.Parse(raw)
}

// NewNullableType parses a label or integer string unless it is blank.
func NewNullableType(raw string) (*GeokretType, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	return NewType(raw)
}

// TypeFromInt validates a raw integer GeoKret type ID.
func TypeFromInt(value int16) (*GeokretType, error) {
	return DefaultGeokretTypeRegistry.FromInt(value)
}

// Name returns the human-readable label for a GeoKret type ID.
func (r *GeokretTypeRegistry) Name(typeID int16) string {
	if name, ok := r.types[typeID]; ok {
		return name
	}
	return r.unknown
}

// IsValid reports whether the provided GeoKret type ID exists in the registry.
func (r *GeokretTypeRegistry) IsValid(typeID int16) bool {
	_, ok := r.types[typeID]
	return ok
}

// All returns a defensive copy of the registered GeoKret type labels.
func (r *GeokretTypeRegistry) All() map[int16]string {
	copyOfTypes := make(map[int16]string, len(r.types))
	for _, id := range r.ordered {
		copyOfTypes[id] = r.types[id]
	}
	return copyOfTypes
}

// Parse converts a label or integer string to a validated GeoKret type.
func (r *GeokretTypeRegistry) Parse(raw string) (*GeokretType, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, &GeokretTypeError{Input: raw, Reason: r.formatErr}
	}
	if typeID, ok := r.labels[trimmed]; ok {
		return &GeokretType{value: typeID, valid: true}, nil
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 16)
	if err != nil {
		return nil, &GeokretTypeError{Input: raw, Reason: r.formatErr}
	}
	return r.FromInt(int16(parsed))
}

// FromInt converts a raw integer type ID to a validated GeoKret type.
func (r *GeokretTypeRegistry) FromInt(value int16) (*GeokretType, error) {
	if !r.IsValid(value) {
		return nil, &GeokretTypeError{Input: strconv.FormatInt(int64(value), 10), Reason: fmt.Sprintf("unknown geokret type id: %d", value)}
	}
	return &GeokretType{value: value, valid: true}, nil
}

// EncodeJSON serializes a GeoKret type ID as its label.
func (r *GeokretTypeRegistry) EncodeJSON(typeID int16) ([]byte, error) {
	if !r.IsValid(typeID) {
		return nil, fmt.Errorf("invalid geokret type id: %d", typeID)
	}
	return json.Marshal(r.Name(typeID))
}

// DecodeJSON accepts either a string label or an integer ID.
func (r *GeokretTypeRegistry) DecodeJSON(data []byte) (int16, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" || trimmed == "" {
		return 0, &GeokretTypeError{Input: string(data), Reason: r.formatErr}
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
		return 0, &GeokretTypeError{Input: string(data), Reason: "invalid geokret type JSON value; expected string or integer"}
	}
	value, err := number.Int64()
	if err != nil {
		return 0, &GeokretTypeError{Input: number.String(), Reason: "geokret type must be an integer"}
	}
	parsed, err := r.fromInt64(value)
	if err != nil {
		return 0, err
	}
	return parsed.value, nil
}

// EncodeXML serializes a GeoKret type ID as element text.
func (r *GeokretTypeRegistry) EncodeXML(typeID int16, enc *xml.Encoder, start xml.StartElement) error {
	if !r.IsValid(typeID) {
		return fmt.Errorf("invalid geokret type id: %d", typeID)
	}
	return enc.EncodeElement(r.Name(typeID), start)
}

// DecodeXML deserializes a GeoKret type ID from element text.
func (r *GeokretTypeRegistry) DecodeXML(dec *xml.Decoder, start xml.StartElement) (int16, error) {
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

// EncodeXMLAttr serializes a GeoKret type ID as an XML attribute.
func (r *GeokretTypeRegistry) EncodeXMLAttr(typeID int16, name xml.Name) (xml.Attr, error) {
	if !r.IsValid(typeID) {
		return xml.Attr{}, fmt.Errorf("invalid geokret type id: %d", typeID)
	}
	return xml.Attr{Name: name, Value: r.Name(typeID)}, nil
}

// DecodeXMLAttr deserializes a GeoKret type ID from an XML attribute.
func (r *GeokretTypeRegistry) DecodeXMLAttr(attr xml.Attr) (int16, error) {
	parsed, err := r.Parse(attr.Value)
	if err != nil {
		return 0, err
	}
	return parsed.value, nil
}

// EncodeCSV serializes a GeoKret type ID as `ID,Label`.
func (r *GeokretTypeRegistry) EncodeCSV(typeID int16) (string, error) {
	if !r.IsValid(typeID) {
		return "", fmt.Errorf("invalid geokret type id: %d", typeID)
	}
	return fmt.Sprintf("%d,%s", typeID, r.Name(typeID)), nil
}

// DecodeCSV accepts `ID`, `Label`, or `ID,Label` input.
func (r *GeokretTypeRegistry) DecodeCSV(csvLine string) (int16, error) {
	parts := strings.Split(csvLine, ",")
	if len(parts) == 1 {
		parsed, err := r.Parse(parts[0])
		if err != nil {
			return 0, err
		}
		return parsed.value, nil
	}
	if len(parts) != 2 {
		return 0, &GeokretTypeError{Input: csvLine, Reason: "invalid CSV format for geokret type"}
	}

	idPart := strings.TrimSpace(parts[0])
	labelPart := strings.TrimSpace(parts[1])
	if idPart != "" {
		parsed, err := r.Parse(idPart)
		if err == nil {
			if labelPart != "" && r.Name(parsed.value) != labelPart {
				return 0, &GeokretTypeError{Input: csvLine, Reason: "geokret type CSV label does not match id"}
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
	return 0, &GeokretTypeError{Input: csvLine, Reason: "invalid CSV format for geokret type"}
}

// EncodeYAML serializes a GeoKret type ID as `{id, label}`.
func (r *GeokretTypeRegistry) EncodeYAML(typeID int16) (any, error) {
	if !r.IsValid(typeID) {
		return nil, fmt.Errorf("invalid geokret type id: %d", typeID)
	}
	return map[string]any{"id": typeID, "label": r.Name(typeID)}, nil
}

// DecodeYAML accepts YAML scalar or `{id, label}` forms.
func (r *GeokretTypeRegistry) DecodeYAML(data []byte) (int16, error) {
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

// ID returns the raw integer GeoKret type ID, or zero when invalid.
func (t *GeokretType) ID() int16 {
	if t == nil || !t.valid {
		return 0
	}
	return t.value
}

// Valid reports whether the GeoKret type holds a known value.
func (t *GeokretType) Valid() bool {
	return t != nil && t.valid
}

// Name returns the label for the type or `Unknown` when invalid.
func (t *GeokretType) Name() string {
	if t == nil || !t.valid {
		return UnknownTypeName
	}
	return DefaultGeokretTypeRegistry.Name(t.value)
}

// String renders the label for a valid type and a diagnostic marker otherwise.
func (t *GeokretType) String() string {
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
func (t GeokretType) MarshalJSON() ([]byte, error) {
	if !t.valid {
		return nil, fmt.Errorf("cannot marshal invalid geokret type")
	}
	return DefaultGeokretTypeRegistry.EncodeJSON(t.value)
}

// UnmarshalJSON accepts a string label or integer ID.
func (t *GeokretType) UnmarshalJSON(data []byte) error {
	if strings.TrimSpace(string(data)) == "null" {
		t.value = 0
		t.valid = false
		return nil
	}
	typeID, err := DefaultGeokretTypeRegistry.DecodeJSON(data)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalText serializes the type as its label.
func (t GeokretType) MarshalText() ([]byte, error) {
	if !t.valid {
		return nil, fmt.Errorf("cannot marshal invalid geokret type")
	}
	return []byte(t.Name()), nil
}

// UnmarshalText accepts a label or integer ID.
func (t *GeokretType) UnmarshalText(text []byte) error {
	parsed, err := DefaultGeokretTypeRegistry.Parse(string(text))
	if err != nil {
		return err
	}
	t.value = parsed.value
	t.valid = true
	return nil
}

// MarshalXML serializes the type as element text.
func (t GeokretType) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	if !t.valid {
		return fmt.Errorf("cannot marshal invalid geokret type")
	}
	return DefaultGeokretTypeRegistry.EncodeXML(t.value, enc, start)
}

// UnmarshalXML deserializes the type from element text.
func (t *GeokretType) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	typeID, err := DefaultGeokretTypeRegistry.DecodeXML(dec, start)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalXMLAttr serializes the type as an XML attribute.
func (t GeokretType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if !t.valid {
		return xml.Attr{}, fmt.Errorf("cannot marshal invalid geokret type")
	}
	return DefaultGeokretTypeRegistry.EncodeXMLAttr(t.value, name)
}

// UnmarshalXMLAttr deserializes the type from an XML attribute.
func (t *GeokretType) UnmarshalXMLAttr(attr xml.Attr) error {
	typeID, err := DefaultGeokretTypeRegistry.DecodeXMLAttr(attr)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalCSV serializes the type as `ID,Label`.
func (t GeokretType) MarshalCSV() (string, error) {
	if !t.valid {
		return "", fmt.Errorf("cannot marshal invalid geokret type")
	}
	return DefaultGeokretTypeRegistry.EncodeCSV(t.value)
}

// UnmarshalCSV accepts `ID`, `Label`, or `ID,Label` input.
func (t *GeokretType) UnmarshalCSV(csvLine string) error {
	typeID, err := DefaultGeokretTypeRegistry.DecodeCSV(csvLine)
	if err != nil {
		return err
	}
	t.value = typeID
	t.valid = true
	return nil
}

// MarshalYAML serializes the type as `{id, label}`.
func (t GeokretType) MarshalYAML() (any, error) {
	if !t.valid {
		return nil, fmt.Errorf("cannot marshal invalid geokret type")
	}
	return DefaultGeokretTypeRegistry.EncodeYAML(t.value)
}

// UnmarshalYAML accepts YAML scalar or `{id, label}` forms.
func (t *GeokretType) UnmarshalYAML(node *yaml.Node) error {
	parsed, err := DefaultGeokretTypeRegistry.parseYAMLNode(node)
	if err != nil {
		return err
	}
	t.value = parsed.value
	t.valid = true
	return nil
}

func (r *GeokretTypeRegistry) parseYAMLNode(node *yaml.Node) (*GeokretType, error) {
	current := unwrapYAMLNode(node)
	if current == nil {
		return nil, &GeokretTypeError{Reason: "invalid YAML structure for geokret type"}
	}

	switch current.Kind {
	case yaml.ScalarNode:
		return r.Parse(current.Value)
	case yaml.MappingNode:
		var idValue *GeokretType
		var labelValue *GeokretType
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
			return nil, &GeokretTypeError{Reason: "invalid YAML structure for geokret type"}
		}
		if idValue != nil && labelValue != nil && idValue.value != labelValue.value {
			return nil, &GeokretTypeError{Reason: "geokret type YAML id and label do not match"}
		}
		if idValue != nil {
			return idValue, nil
		}
		return labelValue, nil
	default:
		return nil, &GeokretTypeError{Reason: fmt.Sprintf("invalid YAML node kind %d for geokret type", current.Kind)}
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

func (r *GeokretTypeRegistry) fromInt64(value int64) (*GeokretType, error) {
	if value < -32768 || value > 32767 {
		return nil, &GeokretTypeError{Input: strconv.FormatInt(value, 10), Reason: "geokret type id is out of int16 range"}
	}
	return r.FromInt(int16(value))
}
