package move

import "testing"

func TestTypeName(t *testing.T) {
	if got := TypeName(5); got != "Dipped" {
		t.Fatalf("TypeName() = %q, want Dipped", got)
	}
	if got := TypeName(99); got != "Unknown" {
		t.Fatalf("TypeName() = %q, want Unknown", got)
	}
}
