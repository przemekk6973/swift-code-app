package util

import (
	"fmt"
	"net/http"
)

// AppError to ustandaryzowany błąd aplikacyjny z kodem HTTP
type AppError struct {
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

// NewError creates new AppError with code status
func NewError(message string, statusCode int) *AppError {
	return &AppError{Message: message, StatusCode: statusCode}
}

// WrapError wraps the communicate with AppError
func WrapError(base *AppError, format string, args ...interface{}) *AppError {
	return &AppError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: base.StatusCode,
	}
}

// base errors
var (
	ErrBadRequest = NewError("bad request", http.StatusBadRequest)
	ErrNotFound   = NewError("not found", http.StatusNotFound)
	ErrConflict   = NewError("conflict", http.StatusConflict)
	ErrInternal   = NewError("internal server error", http.StatusInternalServerError)
)

// StatusCodeFromError returns HTTP status for any error
func StatusCodeFromError(err error) int {
	if e, ok := err.(*AppError); ok {
		return e.StatusCode
	}
	return http.StatusInternalServerError
}

// Helpery to create new AppError with statement

// BadRequest creates AppError with 400 code
func BadRequest(format string, args ...interface{}) *AppError {
	return WrapError(ErrBadRequest, format, args...)
}

// NotFound creates AppError with 404 code
func NotFound(format string, args ...interface{}) *AppError {
	return WrapError(ErrNotFound, format, args...)
}

// Conflict creates AppError with 409 code
func Conflict(format string, args ...interface{}) *AppError {
	return WrapError(ErrConflict, format, args...)
}

// Internal creates AppError with 500 code
func Internal(format string, args ...interface{}) *AppError {
	return WrapError(ErrInternal, format, args...)
}
