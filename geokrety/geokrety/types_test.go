package geokrety

import "testing"

func TestTypeName(t *testing.T) {
	if got := TypeName(4); got != "KretyPost" {
		t.Fatalf("TypeName() = %q, want KretyPost", got)
	}
	if got := TypeName(99); got != "Unknown" {
		t.Fatalf("TypeName() = %q, want Unknown", got)
	}
}
