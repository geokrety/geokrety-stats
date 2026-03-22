package geokrety

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

const invalidFormatReason = "invalid gkid format; expected GK[0-9A-F]+, hexadecimal without prefix, or decimal integer"

type GeokretId struct {
	value int64
}

type GeokretIdError struct {
	Input  string
	Reason string
}

func (e *GeokretIdError) Error() string {
	if strings.TrimSpace(e.Input) == "" {
		return e.Reason
	}
	return fmt.Sprintf("invalid gkid %q: %s", e.Input, e.Reason)
}

func New(raw string) (*GeokretId, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, &GeokretIdError{Input: raw, Reason: invalidFormatReason}
	}
	return parseString(trimmed)
}

func NewNullable(raw string) (*GeokretId, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	return New(raw)
}

func FromInt(value int64) (*GeokretId, error) {
	switch {
	case value == 0:
		return nil, &GeokretIdError{Input: "0", Reason: "gkid must be greater than zero"}
	case value < 0:
		return nil, &GeokretIdError{Input: strconv.FormatInt(value, 10), Reason: "gkid must be positive"}
	default:
		return &GeokretId{value: value}, nil
	}
}

func parseString(raw string) (*GeokretId, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if strings.HasPrefix(normalized, "GK") {
		hexPart := strings.TrimPrefix(normalized, "GK")
		if hexPart == "" || !isHexString(hexPart) {
			return nil, &GeokretIdError{Input: raw, Reason: invalidFormatReason}
		}
		return parseBase(raw, hexPart, 16)
	}

	if isDecimalString(normalized) {
		if len(normalized) > 1 && normalized[0] == '0' {
			return parseBase(raw, normalized, 16)
		}
		return parseBase(raw, normalized, 10)
	}

	if isHexString(normalized) {
		return parseBase(raw, normalized, 16)
	}

	return nil, &GeokretIdError{Input: raw, Reason: invalidFormatReason}
}

func parseBase(input, raw string, base int) (*GeokretId, error) {
	parsed, err := strconv.ParseInt(raw, base, 64)
	if err != nil {
		return nil, &GeokretIdError{Input: input, Reason: "gkid is out of range"}
	}
	return FromInt(parsed)
}

func isDecimalString(raw string) bool {
	if raw == "" {
		return false
	}
	start := 0
	if raw[0] == '+' || raw[0] == '-' {
		if len(raw) == 1 {
			return false
		}
		start = 1
	}
	for _, ch := range raw[start:] {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func isHexString(raw string) bool {
	if raw == "" {
		return false
	}
	for _, ch := range raw {
		if (ch < '0' || ch > '9') && (ch < 'A' || ch > 'F') {
			return false
		}
	}
	return true
}

func formatGKID(value int64) string {
	return fmt.Sprintf("GK%04X", value)
}

func (g *GeokretId) mustValue() int64 {
	if g == nil {
		panic("nil GeokretId receiver")
	}
	if g.value <= 0 {
		panic("invalid zero-value GeokretId")
	}
	return g.value
}

func (g *GeokretId) Int() int64 {
	return g.mustValue()
}

func (g *GeokretId) ToGKID() string {
	return formatGKID(g.mustValue())
}

func (g *GeokretId) String() string {
	switch {
	case g == nil:
		return "nil"
	case g.value <= 0:
		return "invalid"
	default:
		return formatGKID(g.value)
	}
}

func (g GeokretId) Format(state fmt.State, verb rune) {
	formatted := "invalid"
	integerValue := "0"
	if g.value > 0 {
		formatted = formatGKID(g.value)
		integerValue = strconv.FormatInt(g.value, 10)
	}

	switch verb {
	case 'd':
		_, _ = state.Write([]byte(integerValue))
	case 'q':
		_, _ = fmt.Fprintf(state, "%q", formatted)
	case 's', 'v':
		_, _ = state.Write([]byte(formatted))
	default:
		_, _ = state.Write([]byte(formatted))
	}
}

func (g *GeokretId) IntOrZero() int64 {
	if g == nil || g.value <= 0 {
		return 0
	}
	return g.value
}

func (g *GeokretId) ToGKIDOrEmpty() string {
	if g == nil || g.value <= 0 {
		return ""
	}
	return formatGKID(g.value)
}

func (g GeokretId) MarshalJSON() ([]byte, error) {
	if g.value <= 0 {
		return nil, fmt.Errorf("cannot marshal invalid gkid")
	}
	return json.Marshal(formatGKID(g.value))
}

func (g GeokretId) MarshalText() ([]byte, error) {
	if g.value <= 0 {
		return nil, fmt.Errorf("cannot marshal invalid gkid")
	}
	return []byte(formatGKID(g.value)), nil
}

func (g *GeokretId) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" {
		g.value = 0
		return nil
	}
	if len(trimmed) == 0 {
		return &GeokretIdError{Input: string(data), Reason: invalidFormatReason}
	}
	if trimmed[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		parsed, err := New(raw)
		if err != nil {
			return err
		}
		g.value = parsed.value
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return &GeokretIdError{Input: string(data), Reason: "invalid gkid JSON value; expected string, integer, or null"}
	}
	value, err := number.Int64()
	if err != nil {
		return &GeokretIdError{Input: number.String(), Reason: "gkid must be an integer"}
	}
	parsed, err := FromInt(value)
	if err != nil {
		return err
	}
	g.value = parsed.value
	return nil
}

func (g *GeokretId) UnmarshalText(text []byte) error {
	trimmed := strings.TrimSpace(string(text))
	if trimmed == "" {
		return &GeokretIdError{Input: string(text), Reason: invalidFormatReason}
	}
	parsed, err := New(trimmed)
	if err != nil {
		return err
	}
	g.value = parsed.value
	return nil
}

func (g GeokretId) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	text, err := g.MarshalText()
	if err != nil {
		return err
	}
	return enc.EncodeElement(string(text), start)
}

func (g *GeokretId) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var raw string
	if err := dec.DecodeElement(&raw, &start); err != nil {
		return err
	}
	return g.UnmarshalText([]byte(raw))
}

func (g GeokretId) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	text, err := g.MarshalText()
	if err != nil {
		return xml.Attr{}, err
	}
	return xml.Attr{Name: name, Value: string(text)}, nil
}

func (g *GeokretId) UnmarshalXMLAttr(attr xml.Attr) error {
	return g.UnmarshalText([]byte(attr.Value))
}

func (g *GeokretId) Scan(src any) error {
	if src == nil {
		g.value = 0
		return nil
	}
	var parsed *GeokretId
	var err error
	switch value := src.(type) {
	case int64:
		parsed, err = FromInt(value)
	case int32:
		parsed, err = FromInt(int64(value))
	case int:
		parsed, err = FromInt(int64(value))
	case uint64:
		if value > uint64(^uint64(0)>>1) {
			return &GeokretIdError{Input: strconv.FormatUint(value, 10), Reason: "gkid is out of range"}
		}
		parsed, err = FromInt(int64(value))
	case []byte:
		parsed, err = New(string(value))
	case string:
		parsed, err = New(value)
	default:
		return fmt.Errorf("scan gkid from %T: unsupported source type", src)
	}
	if err != nil {
		return err
	}
	g.value = parsed.value
	return nil
}

func (g GeokretId) Value() (driver.Value, error) {
	if g.value <= 0 {
		return nil, fmt.Errorf("cannot store invalid gkid")
	}
	return g.value, nil
}
