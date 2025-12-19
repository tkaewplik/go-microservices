package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    string
	Message string
	Status  int
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeValidationFailed    = "VALIDATION_FAILED"
)

// Predefined errors
var (
	ErrBadRequest = &AppError{
		Code:    CodeBadRequest,
		Message: "bad request",
		Status:  http.StatusBadRequest,
	}

	ErrUnauthorized = &AppError{
		Code:    CodeUnauthorized,
		Message: "unauthorized",
		Status:  http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:    CodeForbidden,
		Message: "forbidden",
		Status:  http.StatusForbidden,
	}

	ErrNotFound = &AppError{
		Code:    CodeNotFound,
		Message: "not found",
		Status:  http.StatusNotFound,
	}

	ErrConflict = &AppError{
		Code:    CodeConflict,
		Message: "conflict",
		Status:  http.StatusConflict,
	}

	ErrInternalServer = &AppError{
		Code:    CodeInternalServerError,
		Message: "internal server error",
		Status:  http.StatusInternalServerError,
	}
)

// New creates a new AppError
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Wrap wraps an error with an AppError
func Wrap(err error, appErr *AppError) *AppError {
	return &AppError{
		Code:    appErr.Code,
		Message: appErr.Message,
		Status:  appErr.Status,
		Err:     err,
	}
}

// WithMessage returns a copy of the error with a custom message
func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: message,
		Status:  e.Status,
		Err:     e.Err,
	}
}

// Is checks if the target error is an AppError with the same code
func Is(err error, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == target.Code
	}
	return false
}

// GetStatus returns the HTTP status code from an error
func GetStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status
	}
	return http.StatusInternalServerError
}

// GetCode returns the error code from an error
func GetCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternalServerError
}
