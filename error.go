package ehttp

import (
	"fmt"
	"net/http"
)

// Common errors.
var (
	InternalServerError = NewErrorf(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	BadRequest          = NewErrorf(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	Unauthorized        = NewErrorf(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	NotFound            = NewErrorf(http.StatusNotFound, http.StatusText(http.StatusNotFound))
)

// Error is a basic error including the http return code.
type Error struct {
	code  int
	error error
}

// Code is an accessor for the error code.
func (e Error) Code() int {
	return e.code
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.error == nil {
		return "<nil>"
	}
	return e.error.Error()
}

// GetError exposes the underlying error.
func (e Error) GetError() error {
	return e.error
}

// NewErrorf creates a new http error including a status code.
func NewErrorf(code int, f string, args ...interface{}) error {
	return &Error{
		code:  code,
		error: fmt.Errorf(f, args...),
	}
}

// NewError creates a new http error including a status code.
func NewError(code int, err error) error {
	return &Error{
		code:  code,
		error: err,
	}
}
