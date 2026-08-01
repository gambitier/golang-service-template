package domainerr

import (
	"errors"
	"fmt"
	"net/http"
)

// Code classifies an application error for HTTP mapping.
type Code string

const (
	CodeInvalidArgument Code = "invalid_argument"
	CodeNotFound        Code = "not_found"
	CodeConflict        Code = "conflict"
	CodeInternal        Code = "internal"
)

// Error is a structured domain/application error.
type Error struct {
	Code    Code
	Message string
	Fields  map[string]any
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatus maps the error code to an HTTP status.
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func InvalidArgument(message string, fields map[string]any) *Error {
	return &Error{Code: CodeInvalidArgument, Message: message, Fields: fields}
}

func NotFound(message string, fields map[string]any) *Error {
	return &Error{Code: CodeNotFound, Message: message, Fields: fields}
}

func Conflict(message string, fields map[string]any) *Error {
	return &Error{Code: CodeConflict, Message: message, Fields: fields}
}

func Internal(message string, err error) *Error {
	return &Error{Code: CodeInternal, Message: message, Err: err}
}

// As extracts *Error from an error chain.
func As(err error) (*Error, bool) {
	var de *Error
	if errors.As(err, &de) {
		return de, true
	}
	return nil, false
}
