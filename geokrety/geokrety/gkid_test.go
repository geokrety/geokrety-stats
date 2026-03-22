package geokrety

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

func TestFromIntFormatsCanonicalGKID(t *testing.T) {
	testCases := []struct {
		name  string
		value int64
		want  string
	}{
		{name: "small", value: 1, want: "GK0001"},
		{name: "byte boundary", value: 255, want: "GK00FF"},
		{name: "wide value", value: 65535, want: "GKFFFF"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gkid, err := FromInt(tc.value)
			if err != nil {
				t.Fatalf("FromInt(%d) returned error: %v", tc.value, err)
			}
			if got := gkid.ToGKID(); got != tc.want {
				t.Fatalf("ToGKID() = %q, want %q", got, tc.want)
			}
			if got := gkid.Int(); got != tc.value {
				t.Fatalf("Int() = %d, want %d", got, tc.value)
			}
		})
	}
}

func TestNewParsesSupportedFormats(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "canonical prefix", input: "GK0001", want: 1},
		{name: "lowercase prefix", input: "gk0001", want: 1},
		{name: "bare hex", input: "00FF", want: 255},
		{name: "bare short hex", input: "FF", want: 255},
		{name: "decimal", input: "255", want: 255},
		{name: "decimal with spaces", input: " 255 ", want: 255},
		{name: "zero padded digits are hex", input: "0001", want: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gkid, err := New(tc.input)
			if err != nil {
				t.Fatalf("New(%q) returned error: %v", tc.input, err)
			}
			if got := gkid.Int(); got != tc.want {
				t.Fatalf("Int() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNewRejectsInvalidValues(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: invalidFormatReason},
		{name: "zero", input: "GK0000", want: "greater than zero"},
		{name: "negative", input: "-5", want: "positive"},
		{name: "junk", input: "XYZ123", want: invalidFormatReason},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.input)
			if err == nil {
				t.Fatalf("New(%q) expected error", tc.input)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestNewNullable(t *testing.T) {
	value, err := NewNullable("   ")
	if err != nil {
		t.Fatalf("NewNullable returned error: %v", err)
	}
	if value != nil {
		t.Fatalf("NewNullable returned %v, want nil", value)
	}
}

func TestNilHelpers(t *testing.T) {
	var gkid *GeokretId
	if got := gkid.String(); got != "nil" {
		t.Fatalf("String() = %q, want %q", got, "nil")
	}
	if got := gkid.IntOrZero(); got != 0 {
		t.Fatalf("IntOrZero() = %d, want 0", got)
	}
	if got := gkid.ToGKIDOrEmpty(); got != "" {
		t.Fatalf("ToGKIDOrEmpty() = %q, want empty string", got)
	}
}

func TestFormatSupportsStringAndDecimal(t *testing.T) {
	value, err := FromInt(255)
	if err != nil {
		t.Fatalf("FromInt returned error: %v", err)
	}
	if got := fmt.Sprintf("%s", value); got != "GK00FF" {
		t.Fatalf("%%s formatting = %q, want GK00FF", got)
	}
	if got := fmt.Sprintf("%v", value); got != "GK00FF" {
		t.Fatalf("%%v formatting = %q, want GK00FF", got)
	}
	if got := fmt.Sprintf("%d", value); got != "255" {
		t.Fatalf("%%d formatting = %q, want 255", got)
	}
}

func TestInvalidValueFormatting(t *testing.T) {
	value := GeokretId{}
	if got := fmt.Sprintf("%v", value); got != "invalid" {
		t.Fatalf("%%v formatting = %q, want invalid", got)
	}
	if got := fmt.Sprintf("%d", value); got != "0" {
		t.Fatalf("%%d formatting = %q, want 0", got)
	}
}

func TestMustValuePanicsOnNilAndZero(t *testing.T) {
	assertPanics(t, func() {
		var gkid *GeokretId
		_ = gkid.Int()
	})
	assertPanics(t, func() {
		gkid := &GeokretId{}
		_ = gkid.ToGKID()
	})
}

func TestJSONRoundTrip(t *testing.T) {
	type payload struct {
		GKID *GeokretId `json:"gkid"`
	}
	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "string", input: `{"gkid":"GK0001"}`, want: "GK0001"},
		{name: "number", input: `{"gkid":255}`, want: "GK00FF"},
		{name: "lowercase", input: `{"gkid":"gk00ff"}`, want: "GK00FF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var decoded payload
			if err := json.Unmarshal([]byte(tc.input), &decoded); err != nil {
				t.Fatalf("Unmarshal returned error: %v", err)
			}
			if decoded.GKID == nil {
				t.Fatalf("decoded GKID is nil")
			}
			if got := decoded.GKID.ToGKID(); got != tc.want {
				t.Fatalf("ToGKID() = %q, want %q", got, tc.want)
			}
			encoded, err := json.Marshal(decoded)
			if err != nil {
				t.Fatalf("Marshal returned error: %v", err)
			}
			if !strings.Contains(string(encoded), tc.want) {
				t.Fatalf("Marshal output %q does not contain %q", string(encoded), tc.want)
			}
		})
	}
}

func TestJSONNullLeavesPointerUnset(t *testing.T) {
	type payload struct {
		GKID *GeokretId `json:"gkid"`
	}
	var decoded payload
	if err := json.Unmarshal([]byte(`{"gkid":null}`), &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if decoded.GKID != nil {
		t.Fatalf("decoded GKID = %v, want nil", decoded.GKID)
	}
}

func TestXMLRoundTrip(t *testing.T) {
	type payload struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid"`
	}
	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "canonical", input: `<payload><gkid>GK0001</gkid></payload>`, want: "GK0001"},
		{name: "decimal", input: `<payload><gkid>255</gkid></payload>`, want: "GK00FF"},
		{name: "lowercase", input: `<payload><gkid>gk00ff</gkid></payload>`, want: "GK00FF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var decoded payload
			if err := xml.Unmarshal([]byte(tc.input), &decoded); err != nil {
				t.Fatalf("Unmarshal returned error: %v", err)
			}
			if decoded.GKID == nil {
				t.Fatalf("decoded GKID is nil")
			}
			if got := decoded.GKID.ToGKID(); got != tc.want {
				t.Fatalf("ToGKID() = %q, want %q", got, tc.want)
			}
			encoded, err := xml.Marshal(decoded)
			if err != nil {
				t.Fatalf("Marshal returned error: %v", err)
			}
			if !strings.Contains(string(encoded), "<gkid>"+tc.want+"</gkid>") {
				t.Fatalf("Marshal output %q does not contain canonical element", string(encoded))
			}
		})
	}
}

func TestXMLAttributeRoundTrip(t *testing.T) {
	type payload struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid,attr"`
	}
	var decoded payload
	if err := xml.Unmarshal([]byte(`<payload gkid="255"></payload>`), &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if decoded.GKID == nil {
		t.Fatalf("decoded GKID is nil")
	}
	if got := decoded.GKID.ToGKID(); got != "GK00FF" {
		t.Fatalf("ToGKID() = %q, want GK00FF", got)
	}
	encoded, err := xml.Marshal(decoded)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if !strings.Contains(string(encoded), `gkid="GK00FF"`) {
		t.Fatalf("Marshal output %q does not contain canonical attribute", string(encoded))
	}
}

func TestXMLMissingElementLeavesPointerUnset(t *testing.T) {
	type payload struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid"`
	}
	var decoded payload
	if err := xml.Unmarshal([]byte(`<payload></payload>`), &decoded); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if decoded.GKID != nil {
		t.Fatalf("decoded GKID = %v, want nil", decoded.GKID)
	}
}

func TestXMLEmptyElementIsRejected(t *testing.T) {
	type payload struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid"`
	}
	var decoded payload
	err := xml.Unmarshal([]byte(`<payload><gkid></gkid></payload>`), &decoded)
	if err == nil {
		t.Fatal("expected error for empty gkid element")
	}
	if !strings.Contains(err.Error(), invalidFormatReason) {
		t.Fatalf("error %q does not contain %q", err.Error(), invalidFormatReason)
	}
}

func TestXMLEmptyAttributeIsRejected(t *testing.T) {
	type payload struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid,attr"`
	}
	var decoded payload
	err := xml.Unmarshal([]byte(`<payload gkid=""></payload>`), &decoded)
	if err == nil {
		t.Fatal("expected error for empty gkid attribute")
	}
	if !strings.Contains(err.Error(), invalidFormatReason) {
		t.Fatalf("error %q does not contain %q", err.Error(), invalidFormatReason)
	}
}

func TestScanAndValue(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  any
		want string
	}{
		{name: "int64", src: int64(255), want: "GK00FF"},
		{name: "string", src: "GK0001", want: "GK0001"},
		{name: "bytes", src: []byte("255"), want: "GK00FF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var value GeokretId
			if err := value.Scan(tc.src); err != nil {
				t.Fatalf("Scan returned error: %v", err)
			}
			if got := value.ToGKID(); got != tc.want {
				t.Fatalf("ToGKID() = %q, want %q", got, tc.want)
			}
		})
	}
	value, err := FromInt(255)
	if err != nil {
		t.Fatalf("FromInt returned error: %v", err)
	}
	stored, err := value.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if stored != int64(255) {
		t.Fatalf("Value() = %v, want 255", stored)
	}
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}
