package geokrety

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const invalidTrackingCodeReason = "tracking code must be 4+ characters and contain only uppercase letters and digits"

// TrackingCodeError describes invalid tracking-code input.
type TrackingCodeError struct {
	Input  string
	Reason string
}

// TrackingCode stores a normalized raw tracking code and exposes masked textual behavior.
type TrackingCode struct {
	value string
}

func (e *TrackingCodeError) Error() string {
	if strings.TrimSpace(e.Input) == "" {
		return e.Reason
	}
	return fmt.Sprintf("invalid tracking code %q: %s", e.Input, e.Reason)
}

// NewTrackingCode normalizes and validates a raw tracking code.
func NewTrackingCode(raw string) (*TrackingCode, error) {
	normalized, err := normalizeTrackingCode(raw)
	if err != nil {
		return nil, err
	}
	return &TrackingCode{value: normalized}, nil
}

// NewNullableTrackingCode normalizes a raw tracking code unless it is blank.
func NewNullableTrackingCode(raw string) (*TrackingCode, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	return NewTrackingCode(raw)
}

// RawForInternalUseOnly returns the normalized raw tracking code.
// This method should be used ONLY for internal application logic and database persistence.
// DO NOT pass the raw value to external systems, logs, or API responses—use Masked() instead.
func (t *TrackingCode) RawForInternalUseOnly() string {
	if !t.Valid() {
		return ""
	}
	return t.value
}

// Valid reports whether the tracking code contains a normalized raw value.

func (t *TrackingCode) Valid() bool {
	if t == nil {
		return false
	}
	return t.value != ""
}

// Masked returns the obfuscated textual form of the tracking code.
func (t *TrackingCode) Masked() string {
	return maskTrackingCode(t.RawForInternalUseOnly())
}

// String returns the masked form for valid values and `invalid` otherwise.

func (t *TrackingCode) String() string {
	if t == nil {
		return "nil"
	}
	if !t.Valid() {
		return "invalid"
	}
	return t.Masked()
}

// Format keeps fmt string rendering masked by default.
func (t *TrackingCode) Format(state fmt.State, verb rune) {
	formatted := t.String()

	switch verb {
	case 'q':
		_, _ = fmt.Fprintf(state, "%q", formatted)
	case 's', 'v':
		_, _ = state.Write([]byte(formatted))
	default:
		_, _ = state.Write([]byte(formatted))
	}
}

// MarshalJSON serializes the masked value.
func (t TrackingCode) MarshalJSON() ([]byte, error) {
	if !t.Valid() {
		return nil, fmt.Errorf("cannot marshal invalid tracking code")
	}
	return json.Marshal(t.Masked())
}

// UnmarshalJSON accepts a raw string or null.
func (t *TrackingCode) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" {
		t.value = ""
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return &TrackingCodeError{Input: string(data), Reason: "invalid tracking code JSON value; expected string or null"}
	}
	normalized, err := normalizeTrackingCode(raw)
	if err != nil {
		return err
	}
	t.value = normalized
	return nil
}

// MarshalText serializes the masked value.
func (t TrackingCode) MarshalText() ([]byte, error) {
	if !t.Valid() {
		return nil, fmt.Errorf("cannot marshal invalid tracking code")
	}
	return []byte(t.Masked()), nil
}

// UnmarshalText accepts a raw tracking-code string.
func (t *TrackingCode) UnmarshalText(text []byte) error {
	normalized, err := normalizeTrackingCode(string(text))
	if err != nil {
		return err
	}
	t.value = normalized
	return nil
}

// MarshalXML serializes the masked value as XML element text.
func (t TrackingCode) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	text, err := t.MarshalText()
	if err != nil {
		return err
	}
	return enc.EncodeElement(string(text), start)
}

// UnmarshalXML accepts a raw tracking-code string from XML element text.
func (t *TrackingCode) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var raw string
	if err := dec.DecodeElement(&raw, &start); err != nil {
		return err
	}
	return t.UnmarshalText([]byte(raw))
}

// MarshalXMLAttr serializes the masked value as an XML attribute.
func (t TrackingCode) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	text, err := t.MarshalText()
	if err != nil {
		return xml.Attr{}, err
	}
	return xml.Attr{Name: name, Value: string(text)}, nil
}

// UnmarshalXMLAttr accepts a raw tracking-code string from an XML attribute.
func (t *TrackingCode) UnmarshalXMLAttr(attr xml.Attr) error {
	return t.UnmarshalText([]byte(attr.Value))
}

// MarshalYAML serializes the masked value.
func (t TrackingCode) MarshalYAML() (any, error) {
	if !t.Valid() {
		return nil, fmt.Errorf("cannot marshal invalid tracking code")
	}
	return t.Masked(), nil
}

// UnmarshalYAML accepts a raw tracking-code string or null.
func (t *TrackingCode) UnmarshalYAML(node *yaml.Node) error {
	current := unwrapTrackingCodeYAMLNode(node)
	if current == nil {
		return &TrackingCodeError{Reason: "invalid YAML structure for tracking code"}
	}
	if current.Tag == "!!null" {
		t.value = ""
		return nil
	}
	if current.Kind != yaml.ScalarNode {
		return &TrackingCodeError{Reason: fmt.Sprintf("invalid YAML node kind %d for tracking code", current.Kind)}
	}
	return t.UnmarshalText([]byte(current.Value))
}

// MarshalCSV serializes the masked value.
func (t TrackingCode) MarshalCSV() (string, error) {
	if !t.Valid() {
		return "", fmt.Errorf("cannot marshal invalid tracking code")
	}
	return t.Masked(), nil
}

// UnmarshalCSV accepts a raw tracking-code string.
func (t *TrackingCode) UnmarshalCSV(value string) error {
	return t.UnmarshalText([]byte(value))
}

// Scan loads the raw tracking code from a database value.
func (t *TrackingCode) Scan(src any) error {
	if src == nil {
		t.value = ""
		return nil
	}
	var raw string
	switch value := src.(type) {
	case string:
		raw = value
	case []byte:
		raw = string(value)
	default:
		return fmt.Errorf("scan tracking code from %T: unsupported source type", src)
	}
	normalized, err := normalizeTrackingCode(raw)
	if err != nil {
		return err
	}
	t.value = normalized
	return nil
}

// Value returns the normalized raw tracking code for database persistence.
func (t TrackingCode) Value() (driver.Value, error) {
	if !t.Valid() {
		return nil, fmt.Errorf("cannot store invalid tracking code")
	}
	return t.value, nil
}

func normalizeTrackingCode(raw string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if len(normalized) < 4 {
		return "", &TrackingCodeError{Input: raw, Reason: invalidTrackingCodeReason}
	}
	// Verify only uppercase letters (A-Z) and digits (0-9)
	for _, ch := range normalized {
		if (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') {
			return "", &TrackingCodeError{Input: raw, Reason: invalidTrackingCodeReason}
		}
	}
	return normalized, nil
}

func maskTrackingCode(raw string) string {
	if len(raw) < 4 {
		return ""
	}
	return raw[:1] + strings.Repeat("*", len(raw)-1)
}

func unwrapTrackingCodeYAMLNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	current := node
	for current.Kind == yaml.DocumentNode && len(current.Content) > 0 {
		current = current.Content[0]
	}
	return current
}
