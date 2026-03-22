package geokrety

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGeokretTypeRegistryNameValidityAndWrapper(t *testing.T) {
	registry := NewGeokretTypeRegistry()
	testCases := []struct {
		id      int16
		want    string
		isValid bool
	}{
		{id: GeokretTypeTraditional, want: "Traditional", isValid: true},
		{id: GeokretTypeBook, want: "Book/CD/DVD...", isValid: true},
		{id: GeokretTypeHumanPet, want: "Human/Pet", isValid: true},
		{id: GeokretTypeCoin, want: "Coin", isValid: true},
		{id: GeokretTypeKretyPost, want: "KretyPost", isValid: true},
		{id: GeokretTypePebble, want: "Pebble", isValid: true},
		{id: GeokretTypeCar, want: "Car", isValid: true},
		{id: GeokretTypePlayingCard, want: "Playing card", isValid: true},
		{id: GeokretTypeDogTagPet, want: "Dog tag/pet", isValid: true},
		{id: GeokretTypeJigsawPart, want: "Jigsaw part", isValid: true},
		{id: GeokretTypeHidden, want: "Hidden GeoKret", isValid: true},
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

	if got := TypeName(GeokretTypeTraditional); got != "Traditional" {
		t.Fatalf("TypeName() = %q, want Traditional", got)
	}
	if got := GeokretTypeName(GeokretTypeHidden); got != "Hidden GeoKret" {
		t.Fatalf("GeokretTypeName() = %q, want Hidden GeoKret", got)
	}

}

func TestGeokretTypeRegistryAllReturnsCopy(t *testing.T) {
	registry := NewGeokretTypeRegistry()
	all := registry.All()
	if len(all) != 11 {
		t.Fatalf("len(All()) = %d, want 11", len(all))
	}
	all[GeokretTypeTraditional] = "mutated"
	if got := registry.Name(GeokretTypeTraditional); got != "Traditional" {
		t.Fatalf("Name() after mutating copy = %q, want Traditional", got)
	}
}

func TestGeokretTypeErrorFormatting(t *testing.T) {
	if got := (&GeokretTypeError{Reason: "broken"}).Error(); got != "broken" {
		t.Fatalf("Error() = %q, want broken", got)
	}
	if got := (&GeokretTypeError{Input: "oops", Reason: "broken"}).Error(); !strings.Contains(got, "oops") {
		t.Fatalf("Error() = %q, want input included", got)
	}
}

func TestGeokretTypeConstructorsAndAccessors(t *testing.T) {
	parsed, err := NewType("Traditional")
	if err != nil {
		t.Fatalf("NewType returned error: %v", err)
	}
	if got := parsed.ID(); got != GeokretTypeTraditional {
		t.Fatalf("ID() = %d, want %d", got, GeokretTypeTraditional)
	}
	if got := parsed.Name(); got != "Traditional" {
		t.Fatalf("Name() = %q, want Traditional", got)
	}
	if !parsed.Valid() {
		t.Fatal("Valid() = false, want true")
	}

	parsed, err = NewType("10")
	if err != nil {
		t.Fatalf("NewType numeric returned error: %v", err)
	}
	if got := parsed.Name(); got != "Hidden GeoKret" {
		t.Fatalf("Name() = %q, want Hidden GeoKret", got)
	}

	nullable, err := NewNullableType("   ")
	if err != nil {
		t.Fatalf("NewNullableType returned error: %v", err)
	}
	if nullable != nil {
		t.Fatalf("NewNullableType = %v, want nil", nullable)
	}
	nullable, err = NewNullableType("Coin")
	if err != nil || nullable == nil || nullable.Name() != "Coin" {
		t.Fatalf("NewNullableType non-empty = %#v, %v, want Coin", nullable, err)
	}

	fromInt, err := TypeFromInt(GeokretTypeCoin)
	if err != nil {
		t.Fatalf("TypeFromInt returned error: %v", err)
	}
	if got := fromInt.String(); got != "Coin" {
		t.Fatalf("String() = %q, want Coin", got)
	}

	if _, err := NewType(""); err == nil || !strings.Contains(err.Error(), invalidGeokretTypeFormatReason) {
		t.Fatalf("NewType empty error = %v, want format error", err)
	}
	if _, err := NewType("99"); err == nil || !strings.Contains(err.Error(), "unknown geokret type id") {
		t.Fatalf("NewType unknown id error = %v, want unknown-id error", err)
	}
	if _, err := TypeFromInt(99); err == nil || !strings.Contains(err.Error(), "unknown geokret type id") {
		t.Fatalf("TypeFromInt invalid error = %v, want unknown-id error", err)
	}

	var nilType *GeokretType
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

	zero := GeokretType{}
	if got := zero.String(); got != "invalid" {
		t.Fatalf("zero String() = %q, want invalid", got)
	}
}

func TestGeokretTypeJSONHelpersAndRoundTrip(t *testing.T) {
	registry := NewGeokretTypeRegistry()
	encoded, err := registry.EncodeJSON(GeokretTypeTraditional)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	if got := string(encoded); got != `"Traditional"` {
		t.Fatalf("MarshalJSON() = %q, want %q", got, `"Traditional"`)
	}
	if _, err := registry.EncodeJSON(99); err == nil {
		t.Fatal("MarshalJSON invalid id should fail")
	}

	for _, tc := range []struct {
		name  string
		input string
		want  int16
	}{
		{name: "label", input: `"Traditional"`, want: GeokretTypeTraditional},
		{name: "number", input: `10`, want: GeokretTypeHidden},
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
		Type *GeokretType `json:"type"`
	}
	var decoded payload
	if err := json.Unmarshal([]byte(`{"type":"Hidden GeoKret"}`), &decoded); err != nil {
		t.Fatalf("json.Unmarshal label returned error: %v", err)
	}
	if decoded.Type == nil || decoded.Type.ID() != GeokretTypeHidden {
		t.Fatalf("decoded.Type = %#v, want Hidden GeoKret", decoded.Type)
	}

	decoded = payload{}
	if err := json.Unmarshal([]byte(`{"type":3}`), &decoded); err != nil {
		t.Fatalf("json.Unmarshal number returned error: %v", err)
	}
	if decoded.Type == nil || decoded.Type.Name() != "Coin" {
		t.Fatalf("decoded.Type = %#v, want Coin", decoded.Type)
	}

	encodedPayload, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(encodedPayload), `"Coin"`) {
		t.Fatalf("json.Marshal output = %q, want Coin label", string(encodedPayload))
	}

	decoded = payload{}
	if err := json.Unmarshal([]byte(`{"type":null}`), &decoded); err != nil {
		t.Fatalf("json.Unmarshal null returned error: %v", err)
	}
	if decoded.Type != nil {
		t.Fatalf("decoded.Type = %#v, want nil", decoded.Type)
	}

	invalid := GeokretType{}
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

func TestGeokretTypeTextXMLAndXMLAttr(t *testing.T) {
	registry := NewGeokretTypeRegistry()
	valid, err := TypeFromInt(GeokretTypePlayingCard)
	if err != nil {
		t.Fatalf("TypeFromInt returned error: %v", err)
	}

	text, err := valid.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText returned error: %v", err)
	}
	if got := string(text); got != "Playing card" {
		t.Fatalf("MarshalText() = %q, want Playing card", got)
	}

	var fromText GeokretType
	if err := fromText.UnmarshalText([]byte("8")); err != nil {
		t.Fatalf("UnmarshalText returned error: %v", err)
	}
	if got := fromText.Name(); got != "Dog tag/pet" {
		t.Fatalf("Name() = %q, want Dog tag/pet", got)
	}
	if err := fromText.UnmarshalText([]byte("")); err == nil {
		t.Fatal("UnmarshalText empty should fail")
	}

	buf := bytes.Buffer{}
	enc := xml.NewEncoder(&buf)
	if err := registry.EncodeXML(GeokretTypeTraditional, enc, xml.StartElement{Name: xml.Name{Local: "type"}}); err != nil {
		t.Fatalf("registry MarshalXML returned error: %v", err)
	}
	if err := enc.Flush(); err != nil {
		t.Fatalf("xml encoder flush returned error: %v", err)
	}
	if got := buf.String(); got != "<type>Traditional</type>" {
		t.Fatalf("registry MarshalXML output = %q, want <type>Traditional</type>", got)
	}
	if err := registry.EncodeXML(99, xml.NewEncoder(&bytes.Buffer{}), xml.StartElement{Name: xml.Name{Local: "type"}}); err == nil {
		t.Fatal("registry MarshalXML invalid id should fail")
	}

	dec, start := newXMLStart(t, `<type>Hidden GeoKret</type>`)
	decodedID, err := registry.DecodeXML(dec, start)
	if err != nil {
		t.Fatalf("registry UnmarshalXML returned error: %v", err)
	}
	if decodedID != GeokretTypeHidden {
		t.Fatalf("registry UnmarshalXML() = %d, want %d", decodedID, GeokretTypeHidden)
	}
	dec, start = newXMLStart(t, `<type>10</type>`)
	decodedID, err = registry.DecodeXML(dec, start)
	if err != nil {
		t.Fatalf("registry UnmarshalXML numeric returned error: %v", err)
	}
	if decodedID != GeokretTypeHidden {
		t.Fatalf("registry UnmarshalXML() = %d, want %d for numeric element", decodedID, GeokretTypeHidden)
	}
	dec, start = newXMLStart(t, `<type></type>`)
	if _, err := registry.DecodeXML(dec, start); err == nil {
		t.Fatal("registry UnmarshalXML empty element should fail")
	}
	dec, start = newXMLStart(t, `<type>`)
	if _, err := registry.DecodeXML(dec, start); err == nil {
		t.Fatal("registry UnmarshalXML malformed element should fail")
	}

	attr, err := registry.EncodeXMLAttr(GeokretTypeCoin, xml.Name{Local: "type"})
	if err != nil {
		t.Fatalf("registry MarshalXMLAttr returned error: %v", err)
	}
	if attr.Value != "Coin" {
		t.Fatalf("registry MarshalXMLAttr value = %q, want Coin", attr.Value)
	}
	if _, err := registry.EncodeXMLAttr(99, xml.Name{Local: "type"}); err == nil {
		t.Fatal("registry MarshalXMLAttr invalid id should fail")
	}
	decodedID, err = registry.DecodeXMLAttr(xml.Attr{Name: xml.Name{Local: "type"}, Value: "4"})
	if err != nil {
		t.Fatalf("registry UnmarshalXMLAttr returned error: %v", err)
	}
	if decodedID != GeokretTypeKretyPost {
		t.Fatalf("registry UnmarshalXMLAttr() = %d, want %d", decodedID, GeokretTypeKretyPost)
	}
	if _, err := registry.DecodeXMLAttr(xml.Attr{Name: xml.Name{Local: "type"}, Value: ""}); err == nil {
		t.Fatal("registry UnmarshalXMLAttr empty should fail")
	}

	type xmlPayload struct {
		XMLName xml.Name     `xml:"payload"`
		Type    *GeokretType `xml:"type"`
		Attr    *GeokretType `xml:"attr,attr"`
	}
	var payload xmlPayload
	if err := xml.Unmarshal([]byte(`<payload attr="3"><type>Traditional</type></payload>`), &payload); err != nil {
		t.Fatalf("xml.Unmarshal returned error: %v", err)
	}
	if payload.Type == nil || payload.Type.Name() != "Traditional" {
		t.Fatalf("payload.Type = %#v, want Traditional", payload.Type)
	}
	if payload.Attr == nil || payload.Attr.Name() != "Coin" {
		t.Fatalf("payload.Attr = %#v, want Coin", payload.Attr)
	}
	marshaledXML, err := xml.Marshal(payload)
	if err != nil {
		t.Fatalf("xml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(marshaledXML), `<type>Traditional</type>`) || !strings.Contains(string(marshaledXML), `attr="Coin"`) {
		t.Fatalf("xml.Marshal output = %q, want canonical element and attribute labels", string(marshaledXML))
	}
	invalid := GeokretType{}
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

func TestGeokretTypeCSVAndYAML(t *testing.T) {
	registry := NewGeokretTypeRegistry()
	line, err := registry.EncodeCSV(GeokretTypeTraditional)
	if err != nil {
		t.Fatalf("MarshalCSV returned error: %v", err)
	}
	if line != "0,Traditional" {
		t.Fatalf("MarshalCSV() = %q, want 0,Traditional", line)
	}
	if _, err := registry.EncodeCSV(99); err == nil {
		t.Fatal("MarshalCSV invalid id should fail")
	}

	for _, tc := range []struct {
		input string
		want  int16
	}{
		{input: "Traditional", want: GeokretTypeTraditional},
		{input: "10", want: GeokretTypeHidden},
		{input: "10,Hidden GeoKret", want: GeokretTypeHidden},
		{input: ",Coin", want: GeokretTypeCoin},
		{input: "oops,Coin", want: GeokretTypeCoin},
	} {
		got, err := registry.DecodeCSV(tc.input)
		if err != nil {
			t.Fatalf("UnmarshalCSV(%q) returned error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("UnmarshalCSV(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
	for _, input := range []string{"1,Traditional", "1,Traditional,extra", ","} {
		if _, err := registry.DecodeCSV(input); err == nil {
			t.Fatalf("UnmarshalCSV(%q) should fail", input)
		}
	}
	for _, input := range []string{"oops", ",Missing"} {
		if _, err := registry.DecodeCSV(input); err == nil {
			t.Fatalf("UnmarshalCSV(%q) should fail", input)
		}
	}

	value, err := TypeFromInt(GeokretTypeDogTagPet)
	if err != nil {
		t.Fatalf("TypeFromInt returned error: %v", err)
	}
	line, err = value.MarshalCSV()
	if err != nil {
		t.Fatalf("Type MarshalCSV returned error: %v", err)
	}
	if line != "8,Dog tag/pet" {
		t.Fatalf("Type MarshalCSV() = %q, want 8,Dog tag/pet", line)
	}
	var fromCSV GeokretType
	if err := fromCSV.UnmarshalCSV("9,Jigsaw part"); err != nil {
		t.Fatalf("Type UnmarshalCSV returned error: %v", err)
	}
	if fromCSV.Name() != "Jigsaw part" {
		t.Fatalf("Type UnmarshalCSV Name() = %q, want Jigsaw part", fromCSV.Name())
	}
	invalid := GeokretType{}
	if _, err := invalid.MarshalCSV(); err == nil {
		t.Fatal("invalid MarshalCSV should fail")
	}
	if err := fromCSV.UnmarshalCSV("oops,wrong,label"); err == nil {
		t.Fatal("invalid UnmarshalCSV should fail")
	}

	marshaledYAML, err := registry.EncodeYAML(GeokretTypeHidden)
	if err != nil {
		t.Fatalf("MarshalYAML returned error: %v", err)
	}
	serialized, err := yaml.Marshal(marshaledYAML)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(serialized), "label: Hidden GeoKret") {
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
		{name: "scalar label", input: "Traditional\n", want: GeokretTypeTraditional},
		{name: "scalar id", input: "10\n", want: GeokretTypeHidden},
		{name: "mapping id only", input: "id: 3\n", want: GeokretTypeCoin},
		{name: "mapping label only", input: "label: Coin\n", want: GeokretTypeCoin},
		{name: "mapping", input: "id: 3\nlabel: Coin\n", want: GeokretTypeCoin},
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
	for _, input := range []string{"[]\n", "other: value\n", "id: 1\nlabel: Traditional\n"} {
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
		Type GeokretType `yaml:"type"`
	}
	var payload yamlPayload
	if err := yaml.Unmarshal([]byte("type: Hidden GeoKret\n"), &payload); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}
	if payload.Type.Name() != "Hidden GeoKret" {
		t.Fatalf("payload.Type.Name() = %q, want Hidden GeoKret", payload.Type.Name())
	}
	encodedPayload, err := yaml.Marshal(payload)
	if err != nil {
		t.Fatalf("yaml.Marshal payload returned error: %v", err)
	}
	if !strings.Contains(string(encodedPayload), "label: Hidden GeoKret") {
		t.Fatalf("yaml.Marshal payload output = %q, want label", string(encodedPayload))
	}
	if err := yaml.Unmarshal([]byte("type: []\n"), &payload); err == nil {
		t.Fatal("yaml.Unmarshal invalid payload should fail")
	}
	if _, err := invalid.MarshalYAML(); err == nil {
		t.Fatal("invalid MarshalYAML should fail")
	}
	if _, err := DefaultGeokretTypeRegistry.parseYAMLNode(nil); err == nil {
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
