package jsonrest

import "testing"

func TestCursorRoundTrip(t *testing.T) {
	cursor := EncodeCursor(CurrentCursorVersion, 42)
	payload, err := cursor.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if payload.Version != CurrentCursorVersion {
		t.Fatalf("Version = %d, want %d", payload.Version, CurrentCursorVersion)
	}
	if payload.Offset != 42 {
		t.Fatalf("Offset = %d, want 42", payload.Offset)
	}
}

func TestCursorDecodeErrors(t *testing.T) {
	tests := []struct {
		name   string
		cursor Cursor
		want   *RequestError
	}{
		{name: "empty", cursor: "", want: ErrInvalidCursor},
		{name: "garbage", cursor: "%%%", want: ErrInvalidCursor},
		{name: "unsupported version", cursor: EncodeCursor(CurrentCursorVersion+1, 10), want: ErrCursorVersionMismatch},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.cursor.Decode()
			if !IsRequestError(err, tc.want) {
				t.Fatalf("Decode() error = %v, want %v", err, tc.want)
			}
		})
	}
}
