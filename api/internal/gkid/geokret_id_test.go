package gkid

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

func TestFromIntFormatsCanonicalGKID(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  string
	}{
		{name: "one", value: 1, want: "GK0001"},
		{name: "ff", value: 255, want: "GK00FF"},
		{name: "ffff", value: 65535, want: "GKFFFF"},
		{name: "beyond-padding", value: 65536, want: "GK10000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gkid, err := FromInt(tc.value)
			if err != nil {
				t.Fatalf("FromInt(%d) error = %v", tc.value, err)
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

func TestNewParsesMultipleFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "gk-prefixed", input: "GK0001", want: 1},
		{name: "gk-lowercase", input: "gk0001", want: 1},
		{name: "bare-hex-digits", input: "0001", want: 1},
		{name: "bare-hex-letters", input: "00ff", want: 255},
		{name: "decimal", input: "255", want: 255},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gkid, err := New(tc.input)
			if err != nil {
				t.Fatalf("New(%q) error = %v", tc.input, err)
			}
			if got := gkid.Int(); got != tc.want {
				t.Fatalf("Int() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNewRejectsInvalidValues(t *testing.T) {
	tests := []string{"", " ", "GK", "GK0000", "0", "-1", "GK-1", "GKZZ", "XYZ123", "1.5"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := New(input); err == nil {
				t.Fatalf("New(%q) expected error", input)
			}
		})
	}
}

func TestNewNullable(t *testing.T) {
	gkid, err := NewNullable("   ")
	if err != nil {
		t.Fatalf("NewNullable returned error: %v", err)
	}
	if gkid != nil {
		t.Fatalf("NewNullable expected nil, got %v", gkid)
	}
}

func TestNilSafeHelpers(t *testing.T) {
	var gkid *GeokretId
	if got := gkid.String(); got != "nil" {
		t.Fatalf("String() = %q, want nil", got)
	}
	if got := gkid.IntOrZero(); got != 0 {
		t.Fatalf("IntOrZero() = %d, want 0", got)
	}
	if got := gkid.ToGKIDOrEmpty(); got != "" {
		t.Fatalf("ToGKIDOrEmpty() = %q, want empty", got)
	}
}

func TestFmtFormattingUsesCanonicalGKID(t *testing.T) {
	gkid, err := FromInt(255)
	if err != nil {
		t.Fatalf("FromInt returned error: %v", err)
	}

	if got := fmt.Sprintf("%v", *gkid); got != "GK00FF" {
		t.Fatalf("fmt %%v on value = %q, want GK00FF", got)
	}
	if got := fmt.Sprintf("%s", *gkid); got != "GK00FF" {
		t.Fatalf("fmt %%s on value = %q, want GK00FF", got)
	}
	if got := fmt.Sprintf("%v", gkid); got != "GK00FF" {
		t.Fatalf("fmt %%v on pointer = %q, want GK00FF", got)
	}
	if got := fmt.Sprintf("%d", *gkid); got != "255" {
		t.Fatalf("fmt %%d on value = %q, want 255", got)
	}
	if got := fmt.Sprintf("%d", gkid); got != "255" {
		t.Fatalf("fmt %%d on pointer = %q, want 255", got)
	}
}

func TestPrimaryAccessorsPanicOnNil(t *testing.T) {
	assertPanic(t, func() {
		var gkid *GeokretId
		_ = gkid.Int()
	})
	assertPanic(t, func() {
		var gkid *GeokretId
		_ = gkid.ToGKID()
	})
}

func TestMarshalJSON(t *testing.T) {
	gkid, err := FromInt(255)
	if err != nil {
		t.Fatalf("FromInt returned error: %v", err)
	}
	encoded, err := json.Marshal(gkid)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if got := string(encoded); got != `"GK00FF"` {
		t.Fatalf("MarshalJSON = %s, want %q", encoded, `"GK00FF"`)
	}
}

func TestMarshalJSONNilPointer(t *testing.T) {
	var gkid *GeokretId
	encoded, err := json.Marshal(gkid)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if got := string(encoded); got != "null" {
		t.Fatalf("MarshalJSON(nil) = %s, want null", encoded)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{name: "gkid-string", json: `"GK0001"`, want: "GK0001"},
		{name: "bare-hex-string", json: `"00ff"`, want: "GK00FF"},
		{name: "decimal-number", json: `255`, want: "GK00FF"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gkid GeokretId
			if err := json.Unmarshal([]byte(tc.json), &gkid); err != nil {
				t.Fatalf("Unmarshal error = %v", err)
			}
			if got := gkid.ToGKID(); got != tc.want {
				t.Fatalf("ToGKID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUnmarshalJSONNullPointer(t *testing.T) {
	var payload struct {
		GKID *GeokretId `json:"gkid"`
	}
	if err := json.Unmarshal([]byte(`{"gkid":null}`), &payload); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if payload.GKID != nil {
		t.Fatalf("GKID expected nil, got %v", payload.GKID)
	}
}

func TestUnmarshalJSONRejectsInvalidValues(t *testing.T) {
	tests := []string{`0`, `-1`, `1.5`, `true`, `{}`, `[]`, `"GK0000"`, `"GKZZ"`}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			var gkid GeokretId
			if err := json.Unmarshal([]byte(input), &gkid); err == nil {
				t.Fatalf("Unmarshal(%s) expected error", input)
			}
		})
	}
}

func TestMarshalXML(t *testing.T) {
	payload := struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid,omitempty"`
	}{
		GKID: mustGeokretId(t, 255),
	}

	encoded, err := xml.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if got := string(encoded); got != `<payload><gkid>GK00FF</gkid></payload>` {
		t.Fatalf("MarshalXML = %s", encoded)
	}
}

func TestMarshalXMLAttribute(t *testing.T) {
	payload := struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid,attr,omitempty"`
	}{
		GKID: mustGeokretId(t, 255),
	}

	encoded, err := xml.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if got := string(encoded); got != `<payload gkid="GK00FF"></payload>` {
		t.Fatalf("MarshalXML attribute = %s", encoded)
	}
}

func TestMarshalXMLNilPointerOmitted(t *testing.T) {
	payload := struct {
		XMLName xml.Name   `xml:"payload"`
		GKID    *GeokretId `xml:"gkid,omitempty"`
	}{}

	encoded, err := xml.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if strings.Contains(string(encoded), "gkid") {
		t.Fatalf("expected gkid to be omitted, got %s", encoded)
	}
}

func TestUnmarshalXMLElement(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{name: "gkid-element", xml: `<payload><gkid>GK0001</gkid></payload>`, want: "GK0001"},
		{name: "decimal-element", xml: `<payload><gkid>255</gkid></payload>`, want: "GK00FF"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var payload struct {
				GKID *GeokretId `xml:"gkid,omitempty"`
			}
			if err := xml.Unmarshal([]byte(tc.xml), &payload); err != nil {
				t.Fatalf("Unmarshal returned error: %v", err)
			}
			if payload.GKID == nil {
				t.Fatalf("GKID expected non-nil")
			}
			if got := payload.GKID.ToGKID(); got != tc.want {
				t.Fatalf("ToGKID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUnmarshalXMLAttribute(t *testing.T) {
	var payload struct {
		GKID *GeokretId `xml:"gkid,attr,omitempty"`
	}
	if err := xml.Unmarshal([]byte(`<payload gkid="00ff"></payload>`), &payload); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if payload.GKID == nil {
		t.Fatalf("GKID expected non-nil")
	}
	if got := payload.GKID.ToGKID(); got != "GK00FF" {
		t.Fatalf("ToGKID() = %q, want GK00FF", got)
	}
}

func TestUnmarshalXMLMissingElementLeavesNil(t *testing.T) {
	var payload struct {
		GKID *GeokretId `xml:"gkid,omitempty"`
	}
	if err := xml.Unmarshal([]byte(`<payload></payload>`), &payload); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if payload.GKID != nil {
		t.Fatalf("GKID expected nil, got %v", payload.GKID)
	}
}

func TestUnmarshalXMLRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		attr bool
	}{
		{name: "invalid-zero-element", xml: `<payload><gkid>GK0000</gkid></payload>`},
		{name: "invalid-format-element", xml: `<payload><gkid>GKZZ</gkid></payload>`},
		{name: "invalid-zero-attr", xml: `<payload gkid="0"></payload>`, attr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.attr {
				var payload struct {
					GKID *GeokretId `xml:"gkid,attr,omitempty"`
				}
				if err := xml.Unmarshal([]byte(tc.xml), &payload); err == nil {
					t.Fatalf("Unmarshal(%s) expected error", tc.xml)
				}
				return
			}
			var payload struct {
				GKID *GeokretId `xml:"gkid,omitempty"`
			}
			if err := xml.Unmarshal([]byte(tc.xml), &payload); err == nil {
				t.Fatalf("Unmarshal(%s) expected error", tc.xml)
			}
		})
	}
}

func TestScanAndValue(t *testing.T) {
	var fromSQL GeokretId
	if err := fromSQL.Scan(int64(255)); err != nil {
		t.Fatalf("Scan(int64) error = %v", err)
	}
	if got := fromSQL.ToGKID(); got != "GK00FF" {
		t.Fatalf("ToGKID() = %q, want GK00FF", got)
	}
	if err := fromSQL.Scan([]byte("GK0001")); err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}
	if got := fromSQL.Int(); got != 1 {
		t.Fatalf("Int() = %d, want 1", got)
	}
	value, err := fromSQL.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	if got, ok := value.(int64); !ok || got != 1 {
		t.Fatalf("Value() = %#v, want int64(1)", value)
	}
	if err := fromSQL.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if got := fromSQL.IntOrZero(); got != 0 {
		t.Fatalf("IntOrZero() after Scan(nil) = %d, want 0", got)
	}
}

func TestValueImplementsDriverValuer(t *testing.T) {
	gkid, err := FromInt(42)
	if err != nil {
		t.Fatalf("FromInt returned error: %v", err)
	}
	var _ driver.Valuer = *gkid
}

func mustGeokretId(t *testing.T, value int64) *GeokretId {
	t.Helper()
	parsed, err := FromInt(value)
	if err != nil {
		t.Fatalf("FromInt(%d) returned error: %v", value, err)
	}
	return parsed
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
