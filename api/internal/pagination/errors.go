package pagination

import "errors"

var (
	ErrInvalidCursor            = errors.New("invalid cursor format")
	ErrUnsupportedCursorVersion = errors.New("unsupported cursor version")
	ErrInvalidLimit             = errors.New("invalid limit value")
	ErrInvalidOffset            = errors.New("invalid offset value")
	ErrAmbiguousPagination      = errors.New("cannot combine cursor with offset")
)
