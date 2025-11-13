package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/alinoer/go-std-api/internal/errors"
	"github.com/alinoer/go-std-api/internal/logger"
	"github.com/google/uuid"
)

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Details   string                 `json:"details,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// ErrorHandlerConfig configures the error handling middleware
type ErrorHandlerConfig struct {
	// EnableStackTrace includes stack traces in development mode
	EnableStackTrace bool
	// EnableDetailedErrors includes detailed error information
	EnableDetailedErrors bool
	// DefaultMessage is used when no specific error message is available
	DefaultMessage string
	// Logger instance for error logging
	Logger *logger.Logger
}

// DefaultErrorHandlerConfig returns a default configuration
func DefaultErrorHandlerConfig() *ErrorHandlerConfig {
	return &ErrorHandlerConfig{
		EnableStackTrace:     false, // Enable in development
		EnableDetailedErrors: false, // Enable in development
		DefaultMessage:      "An error occurred while processing your request",
		Logger:              logger.GetLogger(),
	}
}

// ErrorHandler creates an error handling middleware
func ErrorHandler(config *ErrorHandlerConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultErrorHandlerConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom ResponseWriter to capture panics and errors
			ew := &errorResponseWriter{
				ResponseWriter: w,
				request:        r,
				config:         config,
			}

			// Set up panic recovery
			defer func() {
				if rec := recover(); rec != nil {
					err := fmt.Errorf("panic: %v", rec)
					stackTrace := string(debug.Stack())
					
					appErr := errors.InternalError("Internal server error").
						WithInternal(err).
						WithContext("panic", true).
						WithContext("stack_trace", stackTrace)
					
					ew.WriteError(appErr)
				}
			}()

			// Continue with the request
			next.ServeHTTP(ew, r)
		})
	}
}

// errorResponseWriter wraps http.ResponseWriter to handle errors
type errorResponseWriter struct {
	http.ResponseWriter
	request      *http.Request
	config       *ErrorHandlerConfig
	statusCode   int
	errorWritten bool
}

// WriteHeader captures the status code
func (ew *errorResponseWriter) WriteHeader(statusCode int) {
	ew.statusCode = statusCode
	if !ew.errorWritten && statusCode >= 400 {
		ew.ResponseWriter.WriteHeader(statusCode)
	} else if !ew.errorWritten {
		ew.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write captures the response
func (ew *errorResponseWriter) Write(data []byte) (int, error) {
	if ew.statusCode == 0 {
		ew.statusCode = http.StatusOK
	}
	return ew.ResponseWriter.Write(data)
}

// WriteError writes a structured error response
func (ew *errorResponseWriter) WriteError(err error) {
	if ew.errorWritten {
		return // Prevent double error writing
	}
	ew.errorWritten = true

	appErr := errors.AsAppError(err)
	if appErr == nil {
		appErr = errors.InternalError("An unexpected error occurred").WithInternal(err)
	}

	// Add request context to error
	if requestID := getRequestID(ew.request); requestID != "" {
		appErr = appErr.WithRequestID(requestID)
	}
	
	// Extract user ID from context if available
	if userID := getUserID(ew.request); userID != "" {
		appErr = appErr.WithUserID(userID)
	}

	// Log the error
	ew.logError(appErr)

	// Create error response
	response := ew.buildErrorResponse(appErr)

	// Set content type and status code
	ew.ResponseWriter.Header().Set("Content-Type", "application/json")
	ew.ResponseWriter.WriteHeader(appErr.HTTPStatus)

	// Write response
	if err := json.NewEncoder(ew.ResponseWriter).Encode(response); err != nil {
		// Fallback if JSON encoding fails
		http.Error(ew.ResponseWriter, "Internal server error", http.StatusInternalServerError)
	}
}

// buildErrorResponse creates a structured error response
func (ew *errorResponseWriter) buildErrorResponse(appErr *errors.AppError) *ErrorResponse {
	response := &ErrorResponse{
		Error:     appErr.Message,
		Code:      string(appErr.Code),
		Timestamp: appErr.Timestamp,
		RequestID: appErr.RequestID,
	}

	// Add details if enabled or for client errors
	if ew.config.EnableDetailedErrors || appErr.HTTPStatus < 500 {
		response.Details = appErr.Details
		response.Context = appErr.Context
	}

	// For production, sanitize internal server errors
	if appErr.HTTPStatus >= 500 && !ew.config.EnableDetailedErrors {
		response.Error = ew.config.DefaultMessage
		response.Code = string(errors.ErrCodeInternal)
		response.Details = ""
		response.Context = nil
	}

	return response
}

// logError logs the error with appropriate context
func (ew *errorResponseWriter) logError(appErr *errors.AppError) {
	ctx := ew.request.Context()
	
	// Add request information to context
	ctx = context.WithValue(ctx, logger.RequestIDKey, appErr.RequestID)
	if appErr.UserID != "" {
		ctx = context.WithValue(ctx, logger.UserIDKey, appErr.UserID)
	}

	// Create logger with context
	log := ew.config.Logger.WithContext(ctx)

	// Log based on severity
	switch {
	case appErr.HTTPStatus >= 500:
		log.Error("Server error occurred",
			appErr,
			"method", ew.request.Method,
			"path", ew.request.URL.Path,
			"status_code", appErr.HTTPStatus,
			"error_code", appErr.Code,
		)
	case appErr.HTTPStatus >= 400:
		log.Warn("Client error occurred",
			"method", ew.request.Method,
			"path", ew.request.URL.Path,
			"status_code", appErr.HTTPStatus,
			"error_code", appErr.Code,
			"message", appErr.Message,
		)
	}
}

// getRequestID extracts request ID from request context or headers
func getRequestID(r *http.Request) string {
	// Try to get from context first
	if id := r.Context().Value(logger.RequestIDKey); id != nil {
		if str, ok := id.(string); ok && str != "" {
			return str
		}
	}
	
	// Try to get from headers
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	
	// Generate new request ID
	return uuid.New().String()
}

// getUserID extracts user ID from request context
func getUserID(r *http.Request) string {
	if id := r.Context().Value(logger.UserIDKey); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// ErrorHandlerFunc is a handler function that can return an error
type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

// WithErrorHandler wraps an ErrorHandlerFunc to handle returned errors
func WithErrorHandler(config *ErrorHandlerConfig, handler ErrorHandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultErrorHandlerConfig()
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ew := &errorResponseWriter{
			ResponseWriter: w,
			request:        r,
			config:         config,
		}

		if err := handler(w, r); err != nil {
			ew.WriteError(err)
		}
	}
}

// HTTPError is a convenience function to create HTTP errors
func HTTPError(status int, message string) *errors.AppError {
	var code errors.ErrorCode
	switch status {
	case http.StatusBadRequest:
		code = errors.ErrCodeBadRequest
	case http.StatusUnauthorized:
		code = errors.ErrCodeUnauthorized
	case http.StatusForbidden:
		code = errors.ErrCodeForbidden
	case http.StatusNotFound:
		code = errors.ErrCodeNotFound
	case http.StatusConflict:
		code = errors.ErrCodeConflict
	case http.StatusTooManyRequests:
		code = errors.ErrCodeRateLimit
	case http.StatusInternalServerError:
		code = errors.ErrCodeInternal
	case http.StatusServiceUnavailable:
		code = errors.ErrCodeServiceUnavailable
	default:
		code = errors.ErrCodeInternal
	}

	return errors.NewAppError(code, message).WithHTTPStatus(status)
}

// Recovery middleware for panic recovery
func Recovery(config *ErrorHandlerConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultErrorHandlerConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					err := fmt.Errorf("panic: %v", rec)
					stackTrace := string(debug.Stack())
					
					// Log the panic
					ctx := context.WithValue(r.Context(), logger.RequestIDKey, getRequestID(r))
					log := config.Logger.WithContext(ctx)
					log.Error("Panic recovered",
						err,
						"method", r.Method,
						"path", r.URL.Path,
						"stack_trace", stackTrace,
					)

					// Create error response
					appErr := errors.InternalError("Internal server error").
						WithInternal(err).
						WithContext("panic", true)
					
					if config.EnableStackTrace {
						appErr = appErr.WithContext("stack_trace", stackTrace)
					}

					ew := &errorResponseWriter{
						ResponseWriter: w,
						request:        r,
						config:         config,
					}
					ew.WriteError(appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}