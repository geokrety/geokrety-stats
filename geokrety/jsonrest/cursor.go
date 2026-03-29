package jsonrest

import (
	"encoding/base64"
	"encoding/json"
)

const CurrentCursorVersion = 1

type Cursor string

type CursorPayload struct {
	Version     int    `json:"v"`
	Offset      int64  `json:"o"`
	Fingerprint string `json:"f,omitempty"`
}

func EncodeCursor(version int, offset int64, fingerprint ...string) Cursor {
	payload := CursorPayload{Version: version, Offset: offset}
	if len(fingerprint) > 0 {
		payload.Fingerprint = fingerprint[0]
	}
	data, _ := json.Marshal(payload)
	return Cursor(base64.StdEncoding.EncodeToString(data))
}

func (c Cursor) Decode() (CursorPayload, error) {
	if c == "" {
		return CursorPayload{}, ErrInvalidCursor
	}
	decoded, err := base64.StdEncoding.DecodeString(string(c))
	if err != nil {
		return CursorPayload{}, ErrInvalidCursor
	}
	var payload CursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return CursorPayload{}, ErrInvalidCursor
	}
	if payload.Version != CurrentCursorVersion {
		return CursorPayload{}, ErrCursorVersionMismatch
	}
	if payload.Offset < 0 {
		return CursorPayload{}, ErrInvalidCursor
	}
	return payload, nil
}

func (c Cursor) String() string {
	return string(c)
}

func (c Cursor) IsEmpty() bool {
	return c == ""
}
