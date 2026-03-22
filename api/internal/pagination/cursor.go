package pagination

import (
	"encoding/base64"
	"encoding/json"
)

const CurrentCursorVersion = 1

type Cursor string

type CursorPayload struct {
	V int   `json:"V"`
	O int64 `json:"O"`
}

func EncodeCursor(version int, offset int64) Cursor {
	payload := CursorPayload{V: version, O: offset}
	data, _ := json.Marshal(payload)
	return Cursor(base64.StdEncoding.EncodeToString(data))
}

func (c Cursor) Decode() (version int, offset int64, err error) {
	if c == "" {
		return 0, 0, ErrInvalidCursor
	}
	decoded, err := base64.StdEncoding.DecodeString(string(c))
	if err != nil {
		return 0, 0, ErrInvalidCursor
	}
	var payload CursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return 0, 0, ErrInvalidCursor
	}
	if payload.V != CurrentCursorVersion {
		return 0, 0, ErrUnsupportedCursorVersion
	}
	if payload.O < 0 {
		return 0, 0, ErrInvalidCursor
	}
	return payload.V, payload.O, nil
}

func (c Cursor) IsEmpty() bool {
	return c == ""
}

func (c Cursor) String() string {
	return string(c)
}
