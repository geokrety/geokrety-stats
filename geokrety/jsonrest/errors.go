package jsonrest

type RequestError struct {
	Code    string
	Message string
}

func (e *RequestError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

var (
	ErrInvalidPaginationMode = &RequestError{Code: "INVALID_PAGINATION_MODE", Message: "invalid pagination mode"}
	ErrInvalidPage           = &RequestError{Code: "INVALID_PAGE", Message: "invalid page"}
	ErrOutOfBounds           = &RequestError{Code: "OUT_OF_BOUNDS", Message: "page is out of bounds"}
	ErrLimitExceeded         = &RequestError{Code: "LIMIT_EXCEEDED", Message: "limit exceeds maximum"}
	ErrInvalidCursor         = &RequestError{Code: "INVALID_CURSOR", Message: "invalid cursor"}
	ErrCursorVersionMismatch = &RequestError{Code: "CURSOR_VERSION_MISMATCH", Message: "unsupported cursor version"}
)

func IsRequestError(err error, target *RequestError) bool {
	if err == nil || target == nil {
		return false
	}
	requestErr, ok := err.(*RequestError)
	if !ok {
		return false
	}
	return requestErr.Code == target.Code
}
