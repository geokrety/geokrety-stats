package move

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMoveTypeRegistryNameValidityAndWrapper(t *testing.T) {
	registry := NewMoveTypeRegistry()
	testCases := []struct {
		id      int16
		want    string
		isValid bool
	}{
		{id: MoveTypeDropped, want: "Dropped", isValid: true},
		{id: MoveTypeGrabbed, want: "Grabbed", isValid: true},
		{id: MoveTypeCommented, want: "Commented", isValid: true},
		{id: MoveTypeSeen, want: "Seen", isValid: true},
		{id: MoveTypeArchived, want: "Archived", isValid: true},
		{id: MoveTypeDipped, want: "Dipped", isValid: true},
		{id: -1, want: UnknownTypeName, isValid: false},
		{id: 99, want: UnknownTypeName, isValid: false},
	}

	for _, tc := range testCases {
		if got := registry.Name(tc.id); got != tc.want {
			t.Fatalf("Name(%d) = %q, want %q", tc.id, got, tc.want)
		}
		if got := registry.IsValid(tc.id); got != tc.isValid {
			t.Fatalf("IsValid(%d) = %t, want %t", tc.id, got, tc.isValid)
		}
	}

	if got := TypeName(MoveTypeDropped); got != "Dropped" {
		t.Fatalf("TypeName() = %q, want Dropped", got)
	}
	if got := MoveTypeName(MoveTypeDipped); got != "Dipped" {
		t.Fatalf("MoveTypeName() = %q, want Dipped", got)
	}

}

func TestMoveTypeRegistryAllReturnsCopy(t *testing.T) {
	registry := NewMoveTypeRegistry()
	all := registry.All()
	if len(all) != 6 {
		t.Fatalf("len(All()) = %d, want 6", len(all))
	}
	all[MoveTypeDropped] = "mutated"
	if got := registry.Name(MoveTypeDropped); got != "Dropped" {
		t.Fatalf("Name() after mutating copy = %q, want Dropped", got)
	}
}

func TestMoveTypeErrorFormatting(t *testing.T) {
	if got := (&MoveTypeError{Reason: "broken"}).Error(); got != "broken" {
		t.Fatalf("Error() = %q, want broken", got)
	}
	if got := (&MoveTypeError{Input: "oops", Reason: "broken"}).Error(); !strings.Contains(got, "oops") {
		t.Fatalf("Error() = %q, want input included", got)
	}
}

func TestMoveTypeSpecFacingAPI(t *testing.T) {
	registry := NewMoveTypeRegistry()
	var typedRegistry *MoveTypeRegistry = registry
	if got := typedRegistry.Name(MoveTypeSeen); got != "Seen" {
		t.Fatalf("MoveTypeRegistry.Name() = %q, want Seen", got)
	}

	parsed, err := typedRegistry.Parse("Dropped")
	if err != nil {
		t.Fatalf("MoveTypeRegistry.Parse returned error: %v", err)
	}
	var typedValue *MoveType = parsed
	if got := typedValue.Name(); got != "Dropped" {
		t.Fatalf("MoveType.Name() = %q, want Dropped", got)
	}

	_, err = typedRegistry.Parse("")
	if err == nil {
		t.Fatal("MoveTypeRegistry.Parse empty should fail")
	}
	var typedErr *MoveTypeError
	if !errors.As(err, &typedErr) {
		t.Fatalf("Parse error = %T, want *MoveTypeError", err)
	}
}

func TestMoveTypeConstructorsAndAccessors(t *testing.T) {
	parsed, err := NewType("Dropped")
	if err != nil {
		t.Fatalf("NewType returned error: %v", err)
	}
	if got := parsed.ID(); got != MoveTypeDropped {
		t.Fatalf("ID() = %d, want %d", got, MoveTypeDropped)
	}
	if got := parsed.Name(); got != "Dropped" {
		t.Fatalf("Name() = %q, want Dropped", got)
	}
	if !parsed.Valid() {
		t.Fatal("Valid() = false, want true")
	}

	parsed, err = NewType("5")
	if err != nil {
		t.Fatalf("NewType numeric returned error: %v", err)
	}
	if got := parsed.Name(); got != "Dipped" {
		t.Fatalf("Name() = %q, want Dipped", got)
	}

	nullable, err := NewNullableType("   ")
	if err != nil {
		t.Fatalf("NewNullableType returned error: %v", err)
	}
	if nullable != nil {
		t.Fatalf("NewNullableType = %v, want nil", nullable)
	}
	nullable, err = NewNullableType("Seen")
	if err != nil || nullable == nil || nullable.Name() != "Seen" {
		t.Fatalf("NewNullableType non-empty = %#v, %v, want Seen", nullable, err)
	}

	fromInt, err := TypeFromInt(MoveTypeSeen)
	if err != nil {
		t.Fatalf("TypeFromInt returned error: %v", err)
	}
	if got := fromInt.String(); got != "Seen" {
		t.Fatalf("String() = %q, want Seen", got)
	}

	if _, err := NewType(""); err == nil || !strings.Contains(err.Error(), invalidMoveTypeFormatReason) {
		t.Fatalf("NewType empty error = %v, want format error", err)
	}
	if _, err := NewType("99"); err == nil || !strings.Contains(err.Error(), "unknown move type id") {
		t.Fatalf("NewType unknown id error = %v, want unknown-id error", err)
	}
	if _, err := TypeFromInt(99); err == nil || !strings.Contains(err.Error(), "unknown move type id") {
		t.Fatalf("TypeFromInt invalid error = %v, want unknown-id error", err)
	}

	var nilType *MoveType
	if got := nilType.ID(); got != 0 {
		t.Fatalf("nil ID() = %d, want 0", got)
	}
	if nilType.Valid() {
		t.Fatal("nil Valid() = true, want false")
	}
	if got := nilType.Name(); got != UnknownTypeName {
		t.Fatalf("nil Name() = %q, want Unknown", got)
	}
	if got := nilType.String(); got != "nil" {
		t.Fatalf("nil String() = %q, want nil", got)
	}

	zero := MoveType{}
	if got := zero.String(); got != "invalid" {
		t.Fatalf("zero String() = %q, want invalid", got)
	}
}

func TestMoveTypeJSONHelpersAndRoundTrip(t *testing.T) {
	registry := NewMoveTypeRegistry()
	encoded, err := registry.EncodeJSON(MoveTypeDropped)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	if got := string(encoded); got != `"Dropped"` {
		t.Fatalf("MarshalJSON() = %q, want %q", got, `"Dropped"`)
	}
	if _, err := registry.EncodeJSON(99); err == nil {
		t.Fatal("MarshalJSON invalid id should fail")
	}

	for _, tc := range []struct {
		name  string
		input string
		want  int16
	}{
		{name: "label", input: `"Dropped"`, want: MoveTypeDropped},
		{name: "number", input: `5`, want: MoveTypeDipped},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := registry.DecodeJSON([]byte(tc.input))
			if err != nil {
				t.Fatalf("UnmarshalJSON returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("UnmarshalJSON() = %d, want %d", got, tc.want)
			}
		})
	}

	for _, input := range []string{`null`, `{}`, `1.5`} {
		if _, err := registry.DecodeJSON([]byte(input)); err == nil {
			t.Fatalf("UnmarshalJSON(%s) should fail", input)
		}
	}
	if _, err := registry.DecodeJSON([]byte(`"`)); err == nil {
		t.Fatal("UnmarshalJSON malformed quoted string should fail")
	}
	for _, input := range []string{`99`, `"Missing"`} {
		if _, err := registry.DecodeJSON([]byte(input)); err == nil {
			t.Fatalf("UnmarshalJSON(%s) should fail for unknown type", input)
		}
	}
	if _, err := registry.DecodeJSON([]byte(`32768`)); err == nil || !strings.Contains(err.Error(), "out of int16 range") {
		t.Fatalf("DecodeJSON overflow error = %v, want int16 range error", err)
	}
	if _, err := registry.DecodeJSON([]byte(`-32769`)); err == nil || !strings.Contains(err.Error(), "out of int16 range") {
		t.Fatalf("DecodeJSON negative overflow error = %v, want int16 range error", err)
	}

	type payload struct {
		Type *MoveType `json:"type"`
	}
	var decoded payload
	if err := json.Unmarshal([]byte(`{"type":"Grabbed"}`), &decoded); err != nil {
		t.Fatalf("json.Unmarshal label returned error: %v", err)
	}
	if decoded.Type == nil || decoded.Type.ID() != MoveTypeGrabbed {
		t.Fatalf("decoded.Type = %#v, want Grabbed", decoded.Type)
	}

	decoded = payload{}
	if err := json.Unmarshal([]byte(`{"type":2}`), &decoded); err != nil {
		t.Fatalf("json.Unmarshal number returned error: %v", err)
	}
	if decoded.Type == nil || decoded.Type.Name() != "Commented" {
		t.Fatalf("decoded.Type = %#v, want Commented", decoded.Type)
	}

	encodedPayload, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(encodedPayload), `"Commented"`) {
		t.Fatalf("json.Marshal output = %q, want Commented label", string(encodedPayload))
	}

	decoded = payload{}
	if err := json.Unmarshal([]byte(`{"type":null}`), &decoded); err != nil {
		t.Fatalf("json.Unmarshal null returned error: %v", err)
	}
	if decoded.Type != nil {
		t.Fatalf("decoded.Type = %#v, want nil", decoded.Type)
	}

	invalid := MoveType{}
	if _, err := invalid.MarshalJSON(); err == nil {
		t.Fatal("invalid MarshalJSON should fail")
	}
	if err := invalid.UnmarshalJSON([]byte(`99`)); err == nil {
		t.Fatal("invalid UnmarshalJSON should fail for unknown type")
	}
	if err := invalid.UnmarshalJSON([]byte(`32768`)); err == nil {
		t.Fatal("invalid UnmarshalJSON should fail for overflow")
	}
	if err := invalid.UnmarshalJSON([]byte(`-32769`)); err == nil {
		t.Fatal("invalid UnmarshalJSON should fail for negative overflow")
	}
	if err := invalid.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Fatalf("UnmarshalJSON null returned error: %v", err)
	}
	if invalid.Valid() {
		t.Fatal("UnmarshalJSON null should leave value invalid")
	}
}

func TestMoveTypeTextXMLAndXMLAttr(t *testing.T) {
	registry := NewMoveTypeRegistry()
	valid, err := TypeFromInt(MoveTypeArchived)
	if err != nil {
		t.Fatalf("TypeFromInt returned error: %v", err)
	}

	text, err := valid.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText returned error: %v", err)
	}
	if got := string(text); got != "Archived" {
		t.Fatalf("MarshalText() = %q, want Archived", got)
	}

	var fromText MoveType
	if err := fromText.UnmarshalText([]byte("4")); err != nil {
		t.Fatalf("UnmarshalText returned error: %v", err)
	}
	if got := fromText.Name(); got != "Archived" {
		t.Fatalf("Name() = %q, want Archived", got)
	}
	if err := fromText.UnmarshalText([]byte("")); err == nil {
		t.Fatal("UnmarshalText empty should fail")
	}

	buf := bytes.Buffer{}
	enc := xml.NewEncoder(&buf)
	if err := registry.EncodeXML(MoveTypeDropped, enc, xml.StartElement{Name: xml.Name{Local: "type"}}); err != nil {
		t.Fatalf("registry MarshalXML returned error: %v", err)
	}
	if err := enc.Flush(); err != nil {
		t.Fatalf("xml encoder flush returned error: %v", err)
	}
	if got := buf.String(); got != "<type>Dropped</type>" {
		t.Fatalf("registry MarshalXML output = %q, want <type>Dropped</type>", got)
	}
	if err := registry.EncodeXML(99, xml.NewEncoder(&bytes.Buffer{}), xml.StartElement{Name: xml.Name{Local: "type"}}); err == nil {
		t.Fatal("registry MarshalXML invalid id should fail")
	}

	dec, start := newXMLStart(t, `<type>Dipped</type>`)
	decodedID, err := registry.DecodeXML(dec, start)
	if err != nil {
		t.Fatalf("registry UnmarshalXML returned error: %v", err)
	}
	if decodedID != MoveTypeDipped {
		t.Fatalf("registry UnmarshalXML() = %d, want %d", decodedID, MoveTypeDipped)
	}
	dec, start = newXMLStart(t, `<type>5</type>`)
	decodedID, err = registry.DecodeXML(dec, start)
	if err != nil {
		t.Fatalf("registry UnmarshalXML numeric returned error: %v", err)
	}
	if decodedID != MoveTypeDipped {
		t.Fatalf("registry UnmarshalXML() = %d, want %d for numeric element", decodedID, MoveTypeDipped)
	}
	dec, start = newXMLStart(t, `<type></type>`)
	if _, err := registry.DecodeXML(dec, start); err == nil {
		t.Fatal("registry UnmarshalXML empty element should fail")
	}
	dec, start = newXMLStart(t, `<type>`)
	if _, err := registry.DecodeXML(dec, start); err == nil {
		t.Fatal("registry UnmarshalXML malformed element should fail")
	}

	attr, err := registry.EncodeXMLAttr(MoveTypeSeen, xml.Name{Local: "type"})
	if err != nil {
		t.Fatalf("registry MarshalXMLAttr returned error: %v", err)
	}
	if attr.Value != "Seen" {
		t.Fatalf("registry MarshalXMLAttr value = %q, want Seen", attr.Value)
	}
	if _, err := registry.EncodeXMLAttr(99, xml.Name{Local: "type"}); err == nil {
		t.Fatal("registry MarshalXMLAttr invalid id should fail")
	}
	decodedID, err = registry.DecodeXMLAttr(xml.Attr{Name: xml.Name{Local: "type"}, Value: "1"})
	if err != nil {
		t.Fatalf("registry UnmarshalXMLAttr returned error: %v", err)
	}
	if decodedID != MoveTypeGrabbed {
		t.Fatalf("registry UnmarshalXMLAttr() = %d, want %d", decodedID, MoveTypeGrabbed)
	}
	if _, err := registry.DecodeXMLAttr(xml.Attr{Name: xml.Name{Local: "type"}, Value: ""}); err == nil {
		t.Fatal("registry UnmarshalXMLAttr empty should fail")
	}

	type xmlPayload struct {
		XMLName xml.Name  `xml:"payload"`
		Type    *MoveType `xml:"type"`
		Attr    *MoveType `xml:"attr,attr"`
	}
	var payload xmlPayload
	if err := xml.Unmarshal([]byte(`<payload attr="3"><type>Dropped</type></payload>`), &payload); err != nil {
		t.Fatalf("xml.Unmarshal returned error: %v", err)
	}
	if payload.Type == nil || payload.Type.Name() != "Dropped" {
		t.Fatalf("payload.Type = %#v, want Dropped", payload.Type)
	}
	if payload.Attr == nil || payload.Attr.Name() != "Seen" {
		t.Fatalf("payload.Attr = %#v, want Seen", payload.Attr)
	}
	marshaledXML, err := xml.Marshal(payload)
	if err != nil {
		t.Fatalf("xml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(marshaledXML), `<type>Dropped</type>`) || !strings.Contains(string(marshaledXML), `attr="Seen"`) {
		t.Fatalf("xml.Marshal output = %q, want canonical element and attribute labels", string(marshaledXML))
	}
	invalid := MoveType{}
	dec, start = newXMLStart(t, `<type></type>`)
	if err := invalid.UnmarshalXML(dec, start); err == nil {
		t.Fatal("invalid UnmarshalXML should fail on empty element")
	}
	if err := invalid.UnmarshalXMLAttr(xml.Attr{Name: xml.Name{Local: "type"}, Value: ""}); err == nil {
		t.Fatal("invalid UnmarshalXMLAttr should fail on empty attr")
	}
	if _, err := invalid.MarshalText(); err == nil {
		t.Fatal("invalid MarshalText should fail")
	}
	if err := invalid.MarshalXML(xml.NewEncoder(&bytes.Buffer{}), xml.StartElement{Name: xml.Name{Local: "type"}}); err == nil {
		t.Fatal("invalid MarshalXML should fail")
	}
	if _, err := invalid.MarshalXMLAttr(xml.Name{Local: "type"}); err == nil {
		t.Fatal("invalid MarshalXMLAttr should fail")
	}
}

func TestMoveTypeCSVAndYAML(t *testing.T) {
	registry := NewMoveTypeRegistry()
	line, err := registry.EncodeCSV(MoveTypeDropped)
	if err != nil {
		t.Fatalf("MarshalCSV returned error: %v", err)
	}
	if line != "0,Dropped" {
		t.Fatalf("MarshalCSV() = %q, want 0,Dropped", line)
	}
	if _, err := registry.EncodeCSV(99); err == nil {
		t.Fatal("MarshalCSV invalid id should fail")
	}

	for _, tc := range []struct {
		input string
		want  int16
	}{
		{input: "Dropped", want: MoveTypeDropped},
		{input: "5", want: MoveTypeDipped},
		{input: "5,Dipped", want: MoveTypeDipped},
		{input: ",Seen", want: MoveTypeSeen},
		{input: "oops,Seen", want: MoveTypeSeen},
	} {
		got, err := registry.DecodeCSV(tc.input)
		if err != nil {
			t.Fatalf("UnmarshalCSV(%q) returned error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("UnmarshalCSV(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
	for _, input := range []string{"1,Dropped", "1,Dropped,extra", ","} {
		if _, err := registry.DecodeCSV(input); err == nil {
			t.Fatalf("UnmarshalCSV(%q) should fail", input)
		}
	}
	for _, input := range []string{"oops", ",Missing"} {
		if _, err := registry.DecodeCSV(input); err == nil {
			t.Fatalf("UnmarshalCSV(%q) should fail", input)
		}
	}

	value, err := TypeFromInt(MoveTypeCommented)
	if err != nil {
		t.Fatalf("TypeFromInt returned error: %v", err)
	}
	line, err = value.MarshalCSV()
	if err != nil {
		t.Fatalf("Type MarshalCSV returned error: %v", err)
	}
	if line != "2,Commented" {
		t.Fatalf("Type MarshalCSV() = %q, want 2,Commented", line)
	}
	var fromCSV MoveType
	if err := fromCSV.UnmarshalCSV("4,Archived"); err != nil {
		t.Fatalf("Type UnmarshalCSV returned error: %v", err)
	}
	if fromCSV.Name() != "Archived" {
		t.Fatalf("Type UnmarshalCSV Name() = %q, want Archived", fromCSV.Name())
	}
	invalid := MoveType{}
	if _, err := invalid.MarshalCSV(); err == nil {
		t.Fatal("invalid MarshalCSV should fail")
	}
	if err := fromCSV.UnmarshalCSV("oops,wrong,label"); err == nil {
		t.Fatal("invalid UnmarshalCSV should fail")
	}

	marshaledYAML, err := registry.EncodeYAML(MoveTypeDipped)
	if err != nil {
		t.Fatalf("MarshalYAML returned error: %v", err)
	}
	serialized, err := yaml.Marshal(marshaledYAML)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(serialized), "label: Dipped") {
		t.Fatalf("yaml.Marshal output = %q, want label", string(serialized))
	}
	if _, err := registry.EncodeYAML(99); err == nil {
		t.Fatal("MarshalYAML invalid id should fail")
	}

	for _, tc := range []struct {
		name  string
		input string
		want  int16
	}{
		{name: "scalar label", input: "Dropped\n", want: MoveTypeDropped},
		{name: "scalar id", input: "5\n", want: MoveTypeDipped},
		{name: "mapping id only", input: "id: 2\n", want: MoveTypeCommented},
		{name: "mapping label only", input: "label: Commented\n", want: MoveTypeCommented},
		{name: "mapping", input: "id: 2\nlabel: Commented\n", want: MoveTypeCommented},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := registry.DecodeYAML([]byte(tc.input))
			if err != nil {
				t.Fatalf("UnmarshalYAML returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("UnmarshalYAML() = %d, want %d", got, tc.want)
			}
		})
	}
	for _, input := range []string{"[]\n", "other: value\n", "id: 1\nlabel: Dropped\n"} {
		if _, err := registry.DecodeYAML([]byte(input)); err == nil {
			t.Fatalf("UnmarshalYAML(%q) should fail", input)
		}
	}
	for _, input := range []string{"id: Missing\n", "label: \n"} {
		if _, err := registry.DecodeYAML([]byte(input)); err == nil {
			t.Fatalf("UnmarshalYAML(%q) should fail", input)
		}
	}
	if _, err := registry.DecodeYAML([]byte("{")); err == nil {
		t.Fatal("UnmarshalYAML malformed document should fail")
	}

	type yamlPayload struct {
		Type Type `yaml:"type"`
	}
	var payload yamlPayload
	if err := yaml.Unmarshal([]byte("type: Dipped\n"), &payload); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}
	if payload.Type.Name() != "Dipped" {
		t.Fatalf("payload.Type.Name() = %q, want Dipped", payload.Type.Name())
	}
	encodedPayload, err := yaml.Marshal(payload)
	if err != nil {
		t.Fatalf("yaml.Marshal payload returned error: %v", err)
	}
	if !strings.Contains(string(encodedPayload), "label: Dipped") {
		t.Fatalf("yaml.Marshal payload output = %q, want label", string(encodedPayload))
	}
	if err := yaml.Unmarshal([]byte("type: []\n"), &payload); err == nil {
		t.Fatal("yaml.Unmarshal invalid payload should fail")
	}
	if _, err := invalid.MarshalYAML(); err == nil {
		t.Fatal("invalid MarshalYAML should fail")
	}
	if _, err := DefaultMoveTypeRegistry.parseYAMLNode(nil); err == nil {
		t.Fatal("parseYAMLNode(nil) should fail")
	}

	if unwrapYAMLNode(nil) != nil {
		t.Fatal("unwrapYAMLNode(nil) should return nil")
	}
}

func newXMLStart(t *testing.T, raw string) (*xml.Decoder, xml.StartElement) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(raw))
	token, err := dec.Token()
	if err != nil {
		t.Fatalf("Token returned error: %v", err)
	}
	start, ok := token.(xml.StartElement)
	if !ok {
		t.Fatalf("token = %T, want xml.StartElement", token)
	}
	return dec, start
}
