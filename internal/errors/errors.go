package errors

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents a unique error code for tracking and categorization
type ErrorCode string

const (
	// Client Errors (4xx)
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeRateLimit    ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Server Errors (5xx)
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCodeExternalService    ErrorCode = "EXTERNAL_SERVICE_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Internal   error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the internal error for error chain compatibility
func (e *AppError) Unwrap() error {
	return e.Internal
}

// NewAppError creates a new AppError with the given code and message
func NewAppError(code ErrorCode, message string) *AppError {
	httpStatus := getHTTPStatusForCode(code)
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now().UTC(),
		Context:    make(map[string]interface{}),
	}
}

// WithDetails adds detailed information to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithInternal adds an internal error for debugging
func (e *AppError) WithInternal(err error) *AppError {
	e.Internal = err
	return e
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRequestID adds a request ID for tracing
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithUserID adds a user ID for user context
func (e *AppError) WithUserID(userID string) *AppError {
	e.UserID = userID
	return e
}

// WithHTTPStatus overrides the default HTTP status
func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// getHTTPStatusForCode maps error codes to HTTP status codes
func getHTTPStatusForCode(code ErrorCode) int {
	switch code {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeBadRequest, ErrCodeValidation:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeRateLimit:
		return http.StatusTooManyRequests
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeInternal, ErrCodeDatabaseError, ErrCodeExternalService:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Predefined error constructors for common cases
func NotFound(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

func BadRequest(message string) *AppError {
	return NewAppError(ErrCodeBadRequest, message)
}

func ValidationError(field string, message string) *AppError {
	return NewAppError(ErrCodeValidation, fmt.Sprintf("Validation failed for field '%s'", field)).
		WithDetails(message).
		WithContext("field", field)
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return NewAppError(ErrCodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return NewAppError(ErrCodeForbidden, message)
}

func Conflict(resource string, details string) *AppError {
	return NewAppError(ErrCodeConflict, fmt.Sprintf("%s already exists", resource)).
		WithDetails(details)
}

func InternalError(message string) *AppError {
	return NewAppError(ErrCodeInternal, message)
}

func DatabaseError(operation string, err error) *AppError {
	return NewAppError(ErrCodeDatabaseError, fmt.Sprintf("Database operation failed: %s", operation)).
		WithInternal(err)
}

func ExternalServiceError(service string, err error) *AppError {
	return NewAppError(ErrCodeExternalService, fmt.Sprintf("External service '%s' error", service)).
		WithInternal(err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to AppError if possible, otherwise creates a generic internal error
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Convert standard errors to AppError
	return InternalError("An unexpected error occurred").WithInternal(err)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors struct {
	Errors []*AppError `json:"errors"`
}

// Error implements the error interface for ValidationErrors
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %d errors", len(ve.Errors))
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError(field, message))
}

// HasErrors returns true if there are any validation errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// ToAppError converts ValidationErrors to a single AppError
func (ve *ValidationErrors) ToAppError() *AppError {
	if !ve.HasErrors() {
		return nil
	}

	return NewAppError(ErrCodeValidation, "Multiple validation errors").
		WithContext("validation_errors", ve.Errors)
}
