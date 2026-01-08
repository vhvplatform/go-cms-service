package errors

import (
	"fmt"
)

// AppError represents an application error with context
type AppError struct {
	Code    string
	Message string
	Err     error
	Context map[string]interface{}
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

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// Common error codes
const (
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeInvalidInput   = "INVALID_INPUT"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeDatabase       = "DATABASE_ERROR"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotImplemented = "NOT_IMPLEMENTED"
)

// Common error constructors
func Internal(message string, err error) *AppError {
	return Wrap(err, ErrCodeInternal, message)
}

func NotFound(message string) *AppError {
	return New(ErrCodeNotFound, message)
}

func InvalidInput(message string) *AppError {
	return New(ErrCodeInvalidInput, message)
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message)
}

func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message)
}

func Database(message string, err error) *AppError {
	return Wrap(err, ErrCodeDatabase, message)
}

func Validation(message string) *AppError {
	return New(ErrCodeValidation, message)
}
