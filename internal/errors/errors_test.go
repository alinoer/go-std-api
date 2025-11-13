package errors

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(ErrCodeNotFound, "Resource not found")

	if err.Code != ErrCodeNotFound {
		t.Errorf("expected code %s, got %s", ErrCodeNotFound, err.Code)
	}

	if err.Message != "Resource not found" {
		t.Errorf("expected message 'Resource not found', got %s", err.Message)
	}

	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}

	if time.Since(err.Timestamp) > time.Second {
		t.Error("timestamp should be recent")
	}

	if err.Context == nil {
		t.Error("context should be initialized")
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "basic error",
			err:      NewAppError(ErrCodeBadRequest, "Invalid input"),
			expected: "BAD_REQUEST: Invalid input",
		},
		{
			name: "error with details",
			err: NewAppError(ErrCodeValidation, "Validation failed").
				WithDetails("Username is required"),
			expected: "VALIDATION_ERROR: Validation failed - Username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestAppError_WithMethods(t *testing.T) {
	originalErr := errors.New("database connection failed")
	
	err := NewAppError(ErrCodeDatabaseError, "Database error").
		WithDetails("Connection timeout").
		WithInternal(originalErr).
		WithContext("table", "users").
		WithRequestID("req-123").
		WithUserID("user-456").
		WithHTTPStatus(http.StatusServiceUnavailable)

	if err.Details != "Connection timeout" {
		t.Errorf("expected details 'Connection timeout', got %s", err.Details)
	}

	if err.Internal != originalErr {
		t.Errorf("expected internal error to be set")
	}

	if err.Context["table"] != "users" {
		t.Errorf("expected context table to be 'users', got %v", err.Context["table"])
	}

	if err.RequestID != "req-123" {
		t.Errorf("expected request ID 'req-123', got %s", err.RequestID)
	}

	if err.UserID != "user-456" {
		t.Errorf("expected user ID 'user-456', got %s", err.UserID)
	}

	if err.HTTPStatus != http.StatusServiceUnavailable {
		t.Errorf("expected HTTP status %d, got %d", http.StatusServiceUnavailable, err.HTTPStatus)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewAppError(ErrCodeInternal, "Wrapped error").WithInternal(originalErr)

	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap should return the internal error")
	}

	// Test error chain compatibility
	if !errors.Is(appErr, originalErr) {
		t.Error("errors.Is should work with error chains")
	}
}

func TestPredefinedErrorConstructors(t *testing.T) {
	tests := []struct {
		name           string
		constructor    func() *AppError
		expectedCode   ErrorCode
		expectedStatus int
	}{
		{
			name:           "NotFound",
			constructor:    func() *AppError { return NotFound("user") },
			expectedCode:   ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "BadRequest",
			constructor:    func() *AppError { return BadRequest("invalid input") },
			expectedCode:   ErrCodeBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unauthorized",
			constructor:    func() *AppError { return Unauthorized("") },
			expectedCode:   ErrCodeUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Forbidden",
			constructor:    func() *AppError { return Forbidden("") },
			expectedCode:   ErrCodeForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Conflict",
			constructor:    func() *AppError { return Conflict("user", "already exists") },
			expectedCode:   ErrCodeConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "InternalError",
			constructor:    func() *AppError { return InternalError("server error") },
			expectedCode:   ErrCodeInternal,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, err.Code)
			}

			if err.HTTPStatus != tt.expectedStatus {
				t.Errorf("expected HTTP status %d, got %d", tt.expectedStatus, err.HTTPStatus)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("email", "must be a valid email address")

	if err.Code != ErrCodeValidation {
		t.Errorf("expected code %s, got %s", ErrCodeValidation, err.Code)
	}

	if err.Context["field"] != "email" {
		t.Errorf("expected field context to be 'email', got %v", err.Context["field"])
	}

	if err.Details != "must be a valid email address" {
		t.Errorf("expected details 'must be a valid email address', got %s", err.Details)
	}
}

func TestDatabaseError(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := DatabaseError("SELECT", originalErr)

	if err.Code != ErrCodeDatabaseError {
		t.Errorf("expected code %s, got %s", ErrCodeDatabaseError, err.Code)
	}

	if err.Internal != originalErr {
		t.Error("expected internal error to be set")
	}

	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP status %d, got %d", http.StatusInternalServerError, err.HTTPStatus)
	}
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "AppError",
			err:      NewAppError(ErrCodeNotFound, "not found"),
			expected: true,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAppError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "AppError",
			err:      NewAppError(ErrCodeNotFound, "not found"),
			expected: true,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: true, // Should be converted to AppError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AsAppError(tt.err)
			
			if tt.expected {
				if result == nil {
					t.Error("expected non-nil AppError")
				}
				
				if tt.err != nil && !IsAppError(tt.err) {
					// Standard error should be converted to internal error
					if result.Code != ErrCodeInternal {
						t.Errorf("expected code %s for converted error, got %s", ErrCodeInternal, result.Code)
					}
					if result.Internal != tt.err {
						t.Error("expected original error to be preserved as internal")
					}
				}
			} else {
				if result != nil {
					t.Error("expected nil AppError")
				}
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	ve := &ValidationErrors{}

	// Test empty validation errors
	if ve.HasErrors() {
		t.Error("empty validation errors should not have errors")
	}

	if ve.Error() != "validation failed" {
		t.Errorf("expected 'validation failed', got %s", ve.Error())
	}

	// Add validation errors
	ve.Add("email", "is required")
	ve.Add("password", "must be at least 8 characters")

	if !ve.HasErrors() {
		t.Error("should have validation errors after adding")
	}

	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(ve.Errors))
	}

	// Test conversion to AppError
	appErr := ve.ToAppError()
	if appErr == nil {
		t.Error("should return AppError when there are validation errors")
	}

	if appErr.Code != ErrCodeValidation {
		t.Errorf("expected code %s, got %s", ErrCodeValidation, appErr.Code)
	}

	// Test ToAppError with no errors
	emptyVE := &ValidationErrors{}
	if emptyVE.ToAppError() != nil {
		t.Error("should return nil when there are no validation errors")
	}
}

func TestGetHTTPStatusForCode(t *testing.T) {
	tests := []struct {
		code           ErrorCode
		expectedStatus int
	}{
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeBadRequest, http.StatusBadRequest},
		{ErrCodeValidation, http.StatusBadRequest},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeConflict, http.StatusConflict},
		{ErrCodeRateLimit, http.StatusTooManyRequests},
		{ErrCodeServiceUnavailable, http.StatusServiceUnavailable},
		{ErrCodeInternal, http.StatusInternalServerError},
		{ErrCodeDatabaseError, http.StatusInternalServerError},
		{ErrCodeExternalService, http.StatusInternalServerError},
		{"UNKNOWN_CODE", http.StatusInternalServerError}, // Default case
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			status := getHTTPStatusForCode(tt.code)
			if status != tt.expectedStatus {
				t.Errorf("expected status %d for code %s, got %d", tt.expectedStatus, tt.code, status)
			}
		})
	}
}