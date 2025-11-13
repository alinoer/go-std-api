package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alinoer/go-std-api/internal/errors"
	"github.com/alinoer/go-std-api/internal/logger"
)

func TestErrorHandler(t *testing.T) {
	// Initialize logger for testing
	logger.Initialize("test-service", "1.0.0")

	config := &ErrorHandlerConfig{
		EnableStackTrace:     true,
		EnableDetailedErrors: true,
		DefaultMessage:      "Test error occurred",
		Logger:              logger.GetLogger(),
	}

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful request",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "panic recovery",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := ErrorHandler(config)
			handler := middleware(tt.handler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if !strings.Contains(response.Error, tt.expectedError) {
					t.Errorf("expected error to contain %q, got %q", tt.expectedError, response.Error)
				}
			}
		})
	}
}

func TestErrorResponseWriter_WriteError(t *testing.T) {
	config := &ErrorHandlerConfig{
		EnableStackTrace:     false,
		EnableDetailedErrors: true,
		DefaultMessage:      "Default error message",
		Logger:              logger.GetLogger(),
	}

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
		checkSanitized bool
	}{
		{
			name:           "AppError not found",
			err:            errors.NotFound("user"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   string(errors.ErrCodeNotFound),
		},
		{
			name:           "AppError bad request",
			err:            errors.BadRequest("Invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   string(errors.ErrCodeBadRequest),
		},
		{
			name:           "AppError internal error",
			err:            errors.InternalError("Database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   string(errors.ErrCodeInternal),
			checkSanitized: true,
		},
		{
			name:           "standard error",
			err:            errors.NewAppError(errors.ErrCodeInternal, "test").WithInternal(errors.New("standard error")),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   string(errors.ErrCodeInternal),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			ew := &errorResponseWriter{
				ResponseWriter: w,
				request:        req,
				config:         config,
			}

			ew.WriteError(tt.err)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check content type
			if w.Header().Get("Content-Type") != "application/json" {
				t.Error("expected Content-Type to be application/json")
			}

			// Decode response
			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Check error code
			if response.Code != tt.expectedCode {
				t.Errorf("expected code %q, got %q", tt.expectedCode, response.Code)
			}

			// Check if timestamp is set
			if response.Timestamp.IsZero() {
				t.Error("timestamp should be set")
			}
		})
	}
}

func TestErrorResponseWriter_ProductionMode(t *testing.T) {
	config := &ErrorHandlerConfig{
		EnableStackTrace:     false,
		EnableDetailedErrors: false, // Production mode
		DefaultMessage:      "An error occurred",
		Logger:              logger.GetLogger(),
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	ew := &errorResponseWriter{
		ResponseWriter: w,
		request:        req,
		config:         config,
	}

	// Create internal server error
	err := errors.InternalError("Sensitive database connection details")
	ew.WriteError(err)

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// In production mode, internal errors should be sanitized
	if response.Error != "An error occurred" {
		t.Errorf("expected sanitized error message, got %q", response.Error)
	}

	if response.Details != "" {
		t.Error("details should be empty in production mode for internal errors")
	}

	if response.Context != nil {
		t.Error("context should be nil in production mode for internal errors")
	}
}

func TestWithErrorHandler(t *testing.T) {
	config := DefaultErrorHandlerConfig()

	tests := []struct {
		name           string
		handler        ErrorHandlerFunc
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful handler",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
				return nil
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "handler with error",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errors.NotFound("resource")
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name: "handler with standard error",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("standard error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := WithErrorHandler(config, tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError {
				if w.Header().Get("Content-Type") != "application/json" {
					t.Error("expected JSON error response")
				}
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	tests := []struct {
		status       int
		expectedCode errors.ErrorCode
	}{
		{http.StatusBadRequest, errors.ErrCodeBadRequest},
		{http.StatusUnauthorized, errors.ErrCodeUnauthorized},
		{http.StatusForbidden, errors.ErrCodeForbidden},
		{http.StatusNotFound, errors.ErrCodeNotFound},
		{http.StatusConflict, errors.ErrCodeConflict},
		{http.StatusTooManyRequests, errors.ErrCodeRateLimit},
		{http.StatusInternalServerError, errors.ErrCodeInternal},
		{http.StatusServiceUnavailable, errors.ErrCodeServiceUnavailable},
		{999, errors.ErrCodeInternal}, // Unknown status
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			err := HTTPError(tt.status, "test message")

			if err.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, err.Code)
			}

			if err.HTTPStatus != tt.status {
				t.Errorf("expected HTTP status %d, got %d", tt.status, err.HTTPStatus)
			}

			if err.Message != "test message" {
				t.Errorf("expected message 'test message', got %s", err.Message)
			}
		})
	}
}

func TestRecovery(t *testing.T) {
	config := &ErrorHandlerConfig{
		EnableStackTrace: true,
		Logger:           logger.GetLogger(),
	}

	middleware := Recovery(config)

	// Test panic recovery
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Test normal request
	normalHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()

	normalHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d for normal request, got %d", http.StatusOK, w.Code)
	}
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func() *http.Request
		hasID    bool
	}{
		{
			name: "request with context ID",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				ctx := context.WithValue(req.Context(), logger.RequestIDKey, "ctx-123")
				return req.WithContext(ctx)
			},
			hasID: true,
		},
		{
			name: "request with header ID",
			setupReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("X-Request-ID", "header-123")
				return req
			},
			hasID: true,
		},
		{
			name: "request without ID",
			setupReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/test", nil)
			},
			hasID: true, // Should generate new ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			id := getRequestID(req)

			if tt.hasID && id == "" {
				t.Error("expected non-empty request ID")
			}
		})
	}
}