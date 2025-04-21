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

// NewError tworzy nowy AppError z danym kodem statusu
func NewError(message string, statusCode int) *AppError {
	return &AppError{Message: message, StatusCode: statusCode}
}

// WrapError opakowuje komunikat w bazowy AppError
func WrapError(base *AppError, format string, args ...interface{}) *AppError {
	return &AppError{
		Message:    fmt.Sprintf(format, args...),
		StatusCode: base.StatusCode,
	}
}

// Predefiniowane błędy bazowe
var (
	ErrBadRequest = NewError("bad request", http.StatusBadRequest)
	ErrNotFound   = NewError("not found", http.StatusNotFound)
	ErrConflict   = NewError("conflict", http.StatusConflict)
	ErrInternal   = NewError("internal server error", http.StatusInternalServerError)
)

// StatusCodeFromError zwraca HTTP status dla dowolnego błędu
func StatusCodeFromError(err error) int {
	if e, ok := err.(*AppError); ok {
		return e.StatusCode
	}
	return http.StatusInternalServerError
}

// Helpery do tworzenia nowych AppError z komunikatem

// BadRequest tworzy AppError z kodem 400
func BadRequest(format string, args ...interface{}) *AppError {
	return WrapError(ErrBadRequest, format, args...)
}

// NotFound tworzy AppError z kodem 404
func NotFound(format string, args ...interface{}) *AppError {
	return WrapError(ErrNotFound, format, args...)
}

// Conflict tworzy AppError z kodem 409
func Conflict(format string, args ...interface{}) *AppError {
	return WrapError(ErrConflict, format, args...)
}

// Internal tworzy AppError z kodem 500
func Internal(format string, args ...interface{}) *AppError {
	return WrapError(ErrInternal, format, args...)
}
