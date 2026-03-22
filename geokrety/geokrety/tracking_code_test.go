package geokrety

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTrackingCodeErrorFormatting(t *testing.T) {
	if got := (&TrackingCodeError{Reason: "broken"}).Error(); got != "broken" {
		t.Fatalf("Error() = %q, want broken", got)
	}
	if got := (&TrackingCodeError{Input: "abc", Reason: "broken"}).Error(); !strings.Contains(got, "abc") {
		t.Fatalf("Error() = %q, want input included", got)
	}
}

func TestNewTrackingCodeNormalizesAndMasks(t *testing.T) {
	code, err := NewTrackingCode("  abCDef  ")
	if err != nil {
		t.Fatalf("NewTrackingCode returned error: %v", err)
	}
	if got := code.RawForInternalUseOnly(); got != "ABCDEF" {
		t.Fatalf("Raw() = %q, want ABCDEF", got)
	}
	if got := code.Masked(); got != "A*****" {
		t.Fatalf("Masked() = %q, want A*****", got)
	}
	if got := code.String(); got != "A*****" {
		t.Fatalf("String() = %q, want A*****", got)
	}
	if !code.Valid() {
		t.Fatal("Valid() = false, want true")
	}
}

func TestNewTrackingCodeRejectsEmptyAndShortValues(t *testing.T) {
	for _, input := range []string{"", "   ", "A", "AB", "ABC"} {
		if _, err := NewTrackingCode(input); err == nil {
			t.Fatalf("NewTrackingCode(%q) should fail", input)
		}
	}
}

func TestNewTrackingCodeRejectsSpecialCharacters(t *testing.T) {
	for _, input := range []string{"ABC-D", "ABC D", "ABC\tD", "ABC.DEF", "ABC@D"} {
		if _, err := NewTrackingCode(input); err == nil {
			t.Fatalf("NewTrackingCode(%q) should fail (special character)", input)
		}
	}
}

func TestNewTrackingCodeAcceptsAlphanumeric(t *testing.T) {
	for _, input := range []string{"ABC1", "AB2D", "A123", "A0B1C2", "TEST123", "1234", "999TEST"} {
		code, err := NewTrackingCode(input)
		if err != nil {
			t.Fatalf("NewTrackingCode(%q) returned error: %v (should accept alphanumeric)", input, err)
		}
		if !code.Valid() {
			t.Fatalf("NewTrackingCode(%q) = invalid, want valid", input)
		}
	}
}

func TestNewTrackingCodeBoundaryCharacters(t *testing.T) {
	// Test ASCII boundary values: A-Z range [65-90], 0-9 range [48-57]
	// This tests that the validation logic correctly identifies allowed vs rejected characters
	testCases := []struct {
		input string
		want  bool // true = should pass, false = should fail
	}{
		{"@000", false}, // @ is ASCII 64 (just before A=65)
		{"A000", true},  // A is ASCII 65 (start of A-Z)
		{"Z000", true},  // Z is ASCII 90 (end of A-Z)
		{"[000", false}, // [ is ASCII 91 (just after Z=90)
		{"A/00", false}, // / is ASCII 47 (just before 0=48)
		{"A000", true},  // 0 is ASCII 48 (start of 0-9)
		{"A999", true},  // 9 is ASCII 57 (end of 0-9)
		{"A:00", false}, // : is ASCII 58 (just after 9=57)
	}
	for _, tc := range testCases {
		_, err := NewTrackingCode(tc.input)
		if tc.want && err != nil {
			t.Fatalf("NewTrackingCode(%q) should pass but got error: %v", tc.input, err)
		}
		if !tc.want && err == nil {
			t.Fatalf("NewTrackingCode(%q) should fail but passed", tc.input)
		}
	}
}

func TestNewTrackingCodeNormalizesLowercaseToUppercase(t *testing.T) {
	code, err := NewTrackingCode("abcd")
	if err != nil {
		t.Fatalf("NewTrackingCode(\"abcd\") returned error: %v", err)
	}
	if got := code.RawForInternalUseOnly(); got != "ABCD" {
		t.Fatalf("Raw() = %q, want ABCD", got)
	}
}

func TestNewTrackingCodeAcceptsMinimumLength(t *testing.T) {
	code, err := NewTrackingCode("ABCD")
	if err != nil {
		t.Fatalf("NewTrackingCode(\"ABCD\") returned error: %v", err)
	}
	if code == nil || code.RawForInternalUseOnly() != "ABCD" {
		t.Fatalf("NewTrackingCode(\"ABCD\") = %#v, want raw ABCD", code)
	}
}

func TestNewNullableTrackingCode(t *testing.T) {
	code, err := NewNullableTrackingCode(" ")
	if err != nil {
		t.Fatalf("NewNullableTrackingCode blank returned error: %v", err)
	}
	if code != nil {
		t.Fatalf("NewNullableTrackingCode blank = %#v, want nil", code)
	}

	code, err = NewNullableTrackingCode("GKAB")
	if err != nil {
		t.Fatalf("NewNullableTrackingCode value returned error: %v", err)
	}
	if code == nil || code.RawForInternalUseOnly() != "GKAB" {
		t.Fatalf("NewNullableTrackingCode value = %#v, want raw GKAB", code)
	}
}

func TestTrackingCodeMaskingFirstCharacterOnly(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want string
	}{
		{raw: "ABCD", want: "A***"},
		{raw: "ABCDEF", want: "A*****"},
		{raw: "SECRETCD", want: "S*******"},
		{raw: "GKABCD", want: "G*****"},
	} {
		code, err := NewTrackingCode(tc.raw)
		if err != nil {
			t.Fatalf("NewTrackingCode(%q) returned error: %v", tc.raw, err)
		}
		if got := code.Masked(); got != tc.want {
			t.Fatalf("Masked() for %q = %q, want %q", tc.raw, got, tc.want)
		}
	}
	if got := maskTrackingCode(""); got != "" {
		t.Fatalf("maskTrackingCode(\"\") = %q, want empty", got)
	}
	if got := maskTrackingCode("ABC"); got != "" {
		t.Fatalf("maskTrackingCode(\"ABC\") = %q, want empty (too short)", got)
	}
}

func TestTrackingCodeInvalidStringAndFormat(t *testing.T) {
	invalid := TrackingCode{}
	if got := invalid.String(); got != "invalid" {
		t.Fatalf("String() = %q, want invalid", got)
	}
	if got := invalid.RawForInternalUseOnly(); got != "" {
		t.Fatalf("Raw() = %q, want empty", got)
	}
	if invalid.Valid() {
		t.Fatal("Valid() = true, want false")
	}
	if got := fmt.Sprintf("%s", &invalid); got != "invalid" {
		t.Fatalf("%%s = %q, want invalid", got)
	}
	if got := fmt.Sprintf("%v", &invalid); got != "invalid" {
		t.Fatalf("%%v = %q, want invalid", got)
	}
	if got := fmt.Sprintf("%q", &invalid); got != `"invalid"` {
		t.Fatalf("%%q = %q, want quoted invalid", got)
	}
	var nilCode *TrackingCode
	if got := nilCode.String(); got != "nil" {
		t.Fatalf("nil String() = %q, want nil", got)
	}
	if got := nilCode.RawForInternalUseOnly(); got != "" {
		t.Fatalf("nil Raw() = %q, want empty", got)
	}
	if nilCode.Valid() {
		t.Fatal("nil Valid() = true, want false")
	}
	if got := fmt.Sprintf("%s", nilCode); got != "nil" {
		t.Fatalf("nil %%s = %q, want nil", got)
	}
	if got := fmt.Sprintf("%v", nilCode); got != "nil" {
		t.Fatalf("nil %%v = %q, want nil", got)
	}
	if got := fmt.Sprintf("%q", nilCode); got != `"nil"` {
		t.Fatalf("nil %%q = %q, want quoted nil", got)
	}
}

func TestTrackingCodeFormatMasksValue(t *testing.T) {
	code, err := NewTrackingCode("SECRETCD")
	if err != nil {
		t.Fatalf("NewTrackingCode returned error: %v", err)
	}
	for _, tc := range []struct {
		verb string
		want string
	}{
		{verb: "%s", want: "S*******"},
		{verb: "%v", want: "S*******"},
		{verb: "%q", want: `"S*******"`},
		{verb: "%x", want: "S*******"},
	} {
		if got := fmt.Sprintf(tc.verb, code); got != tc.want {
			t.Fatalf("fmt.Sprintf(%q) = %q, want %q", tc.verb, got, tc.want)
		}
	}
}

func TestTrackingCodeJSONRoundTrip(t *testing.T) {
	code, err := NewTrackingCode("ABCDEFG")
	if err != nil {
		t.Fatalf("NewTrackingCode returned error: %v", err)
	}
	encoded, err := json.Marshal(code)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if got := string(encoded); got != `"A******"` {
		t.Fatalf("json.Marshal = %q, want %q", got, `"A******"`)
	}

	var decoded TrackingCode
	if err := decoded.UnmarshalJSON([]byte(`"XYZEABC"`)); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}
	if got := decoded.RawForInternalUseOnly(); got != "XYZEABC" {
		t.Fatalf("Raw() = %q, want XYZEABC", got)
	}
	if err := decoded.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Fatalf("UnmarshalJSON null returned error: %v", err)
	}
	if decoded.Valid() {
		t.Fatal("decoded.Valid() after null = true, want false")
	}
	for _, input := range []string{`123`, `""`, `"ABC"`, `"ab"`} {
		if err := decoded.UnmarshalJSON([]byte(input)); err == nil {
			t.Fatalf("UnmarshalJSON(%s) should fail", input)
		}
	}
	if _, err := (TrackingCode{}).MarshalJSON(); err == nil {
		t.Fatal("MarshalJSON invalid should fail")
	}
}

func TestTrackingCodeTextXMLAndXMLAttr(t *testing.T) {
	code, err := NewTrackingCode("abcdefg")
	if err != nil {
		t.Fatalf("NewTrackingCode returned error: %v", err)
	}
	text, err := code.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText returned error: %v", err)
	}
	if got := string(text); got != "A******" {
		t.Fatalf("MarshalText() = %q, want A******", got)
	}

	var decoded TrackingCode
	if err := decoded.UnmarshalText([]byte(" zzeabc ")); err != nil {
		t.Fatalf("UnmarshalText returned error: %v", err)
	}
	if got := decoded.RawForInternalUseOnly(); got != "ZZEABC" {
		t.Fatalf("Raw() = %q, want ZZEABC", got)
	}
	if err := decoded.UnmarshalText([]byte("")); err == nil {
		t.Fatal("UnmarshalText empty should fail")
	}

	type xmlPayload struct {
		XMLName xml.Name     `xml:"payload"`
		Code    TrackingCode `xml:"code"`
		Attr    TrackingCode `xml:"attr,attr"`
	}
	payload := xmlPayload{Code: *code, Attr: *code}
	encodedXML, err := xml.Marshal(payload)
	if err != nil {
		t.Fatalf("xml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(encodedXML), `<code>A******</code>`) || !strings.Contains(string(encodedXML), `attr="A******"`) {
		t.Fatalf("xml.Marshal output = %q, want masked values", string(encodedXML))
	}

	var decodedPayload xmlPayload
	if err := xml.Unmarshal([]byte(`<payload attr="xyzdef"><code>abcdefg</code></payload>`), &decodedPayload); err != nil {
		t.Fatalf("xml.Unmarshal returned error: %v", err)
	}
	if decodedPayload.Code.RawForInternalUseOnly() != "ABCDEFG" || decodedPayload.Attr.RawForInternalUseOnly() != "XYZDEF" {
		t.Fatalf("decoded payload = %#v, want normalized raw values", decodedPayload)
	}
	dec := xml.NewDecoder(strings.NewReader(`<code>`))
	token, err := dec.Token()
	if err != nil {
		t.Fatalf("Token returned error: %v", err)
	}
	start, ok := token.(xml.StartElement)
	if !ok {
		t.Fatalf("token = %T, want xml.StartElement", token)
	}
	if err := decoded.UnmarshalXML(dec, start); err == nil {
		t.Fatal("UnmarshalXML malformed element should fail")
	}
	if err := decoded.UnmarshalXMLAttr(xml.Attr{Name: xml.Name{Local: "code"}, Value: ""}); err == nil {
		t.Fatal("UnmarshalXMLAttr empty should fail")
	}
	if _, err := (TrackingCode{}).MarshalText(); err == nil {
		t.Fatal("MarshalText invalid should fail")
	}
	if err := (TrackingCode{}).MarshalXML(xml.NewEncoder(&strings.Builder{}), xml.StartElement{Name: xml.Name{Local: "code"}}); err == nil {
		t.Fatal("MarshalXML invalid should fail")
	}
	if _, err := (TrackingCode{}).MarshalXMLAttr(xml.Name{Local: "code"}); err == nil {
		t.Fatal("MarshalXMLAttr invalid should fail")
	}
}

func TestTrackingCodeYAMLAndCSV(t *testing.T) {
	code, err := NewTrackingCode("abcdefg")
	if err != nil {
		t.Fatalf("NewTrackingCode returned error: %v", err)
	}
	encodedYAML, err := yaml.Marshal(code)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if got := strings.TrimSpace(string(encodedYAML)); got != "A******" {
		t.Fatalf("yaml.Marshal = %q, want A******", got)
	}

	var decoded TrackingCode
	if err := yaml.Unmarshal([]byte("xyzefgh\n"), &decoded); err != nil {
		t.Fatalf("yaml.Unmarshal scalar returned error: %v", err)
	}
	if got := decoded.RawForInternalUseOnly(); got != "XYZEFGH" {
		t.Fatalf("Raw() = %q, want XYZEFGH", got)
	}
	var nullNode yaml.Node
	if err := yaml.Unmarshal([]byte("null\n"), &nullNode); err != nil {
		t.Fatalf("yaml.Unmarshal node returned error: %v", err)
	}
	if err := decoded.UnmarshalYAML(&nullNode); err != nil {
		t.Fatalf("UnmarshalYAML null returned error: %v", err)
	}
	if decoded.Valid() {
		t.Fatal("decoded.Valid() after YAML null = true, want false")
	}
	if err := decoded.UnmarshalYAML(nil); err == nil {
		t.Fatal("UnmarshalYAML nil node should fail")
	}
	if err := yaml.Unmarshal([]byte("[]\n"), &decoded); err == nil {
		t.Fatal("yaml.Unmarshal sequence should fail")
	}
	if _, err := (TrackingCode{}).MarshalYAML(); err == nil {
		t.Fatal("MarshalYAML invalid should fail")
	}

	line, err := code.MarshalCSV()
	if err != nil {
		t.Fatalf("MarshalCSV returned error: %v", err)
	}
	if line != "A******" {
		t.Fatalf("MarshalCSV() = %q, want A******", line)
	}
	if err := decoded.UnmarshalCSV("xyzeabc"); err != nil {
		t.Fatalf("UnmarshalCSV returned error: %v", err)
	}
	if got := decoded.RawForInternalUseOnly(); got != "XYZEABC" {
		t.Fatalf("Raw() = %q, want XYZEABC", got)
	}
	if err := decoded.UnmarshalCSV(" "); err == nil {
		t.Fatal("UnmarshalCSV blank should fail")
	}
	if err := decoded.UnmarshalCSV("ABC"); err == nil {
		t.Fatal("UnmarshalCSV too short should fail")
	}
	if _, err := (TrackingCode{}).MarshalCSV(); err == nil {
		t.Fatal("MarshalCSV invalid should fail")
	}
	if unwrapTrackingCodeYAMLNode(nil) != nil {
		t.Fatal("unwrapTrackingCodeYAMLNode(nil) should return nil")
	}
}

func TestTrackingCodeScanAndValue(t *testing.T) {
	var decoded TrackingCode
	if err := decoded.Scan("abcdefg"); err != nil {
		t.Fatalf("Scan string returned error: %v", err)
	}
	if got := decoded.RawForInternalUseOnly(); got != "ABCDEFG" {
		t.Fatalf("Raw() = %q, want ABCDEFG", got)
	}
	if err := decoded.Scan([]byte("xyzefgh")); err != nil {
		t.Fatalf("Scan bytes returned error: %v", err)
	}
	if got := decoded.RawForInternalUseOnly(); got != "XYZEFGH" {
		t.Fatalf("Raw() = %q, want XYZEFGH", got)
	}
	if err := decoded.Scan(nil); err != nil {
		t.Fatalf("Scan nil returned error: %v", err)
	}
	if decoded.Valid() {
		t.Fatal("decoded.Valid() after nil scan = true, want false")
	}
	if err := decoded.Scan(123); err == nil {
		t.Fatal("Scan unsupported type should fail")
	}
	if err := decoded.Scan(" "); err == nil {
		t.Fatal("Scan blank string should fail")
	}
	if err := decoded.Scan("abc"); err == nil {
		t.Fatal("Scan too short should fail")
	}

	code, err := NewTrackingCode("abcdefg")
	if err != nil {
		t.Fatalf("NewTrackingCode returned error: %v", err)
	}
	value, err := code.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if value != "ABCDEFG" {
		t.Fatalf("Value() = %#v, want ABCDEFG", value)
	}
	if _, err := (TrackingCode{}).Value(); err == nil {
		t.Fatal("Value invalid should fail")
	}
}
