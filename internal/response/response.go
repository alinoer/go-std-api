package response

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alinoer/go-std-api/internal/errors"
	"github.com/alinoer/go-std-api/internal/logger"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorResponse represents an error response (defined in middleware but used here)
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Details   string                 `json:"details,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ResponseWriter provides methods to write standardized responses
type ResponseWriter struct {
	w http.ResponseWriter
	r *http.Request
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter, r *http.Request) *ResponseWriter {
	return &ResponseWriter{w: w, r: r}
}

// JSON writes a JSON response with the given status code and data
func (rw *ResponseWriter) JSON(statusCode int, data interface{}) error {
	return rw.JSONWithMessage(statusCode, data, "")
}

// JSONWithMessage writes a JSON response with status code, data, and message
func (rw *ResponseWriter) JSONWithMessage(statusCode int, data interface{}, message string) error {
	response := StandardResponse{
		Success:   statusCode < 400,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC(),
		RequestID: getRequestIDFromContext(rw.r),
	}

	return rw.writeJSON(statusCode, response)
}

// JSONWithMeta writes a JSON response with status code, data, message, and metadata
func (rw *ResponseWriter) JSONWithMeta(statusCode int, data interface{}, message string, meta interface{}) error {
	response := StandardResponse{
		Success:   statusCode < 400,
		Data:      data,
		Message:   message,
		Meta:      meta,
		Timestamp: time.Now().UTC(),
		RequestID: getRequestIDFromContext(rw.r),
	}

	return rw.writeJSON(statusCode, response)
}

// Success writes a successful response
func (rw *ResponseWriter) Success(data interface{}) error {
	return rw.JSON(http.StatusOK, data)
}

// Created writes a 201 Created response
func (rw *ResponseWriter) Created(data interface{}) error {
	return rw.JSONWithMessage(http.StatusCreated, data, "Resource created successfully")
}

// NoContent writes a 204 No Content response
func (rw *ResponseWriter) NoContent() error {
	rw.w.WriteHeader(http.StatusNoContent)
	return nil
}

// Error writes an error response using AppError
func (rw *ResponseWriter) Error(err error) error {
	appErr := errors.AsAppError(err)
	if appErr == nil {
		appErr = errors.InternalError("An unexpected error occurred")
	}

	// Add request ID if not present
	if appErr.RequestID == "" {
		appErr = appErr.WithRequestID(getRequestIDFromContext(rw.r))
	}

	// Log the error
	rw.logError(appErr)

	// Create error response
	errorResp := ErrorResponse{
		Success:   false,
		Error:     appErr.Message,
		Code:      string(appErr.Code),
		Details:   appErr.Details,
		Context:   appErr.Context,
		Timestamp: time.Now().UTC(),
		RequestID: appErr.RequestID,
	}

	return rw.writeJSON(appErr.HTTPStatus, errorResp)
}

// BadRequest writes a 400 Bad Request error
func (rw *ResponseWriter) BadRequest(message string) error {
	return rw.Error(errors.BadRequest(message))
}

// Unauthorized writes a 401 Unauthorized error
func (rw *ResponseWriter) Unauthorized(message string) error {
	return rw.Error(errors.Unauthorized(message))
}

// Forbidden writes a 403 Forbidden error
func (rw *ResponseWriter) Forbidden(message string) error {
	return rw.Error(errors.Forbidden(message))
}

// NotFound writes a 404 Not Found error
func (rw *ResponseWriter) NotFound(resource string) error {
	return rw.Error(errors.NotFound(resource))
}

// Conflict writes a 409 Conflict error
func (rw *ResponseWriter) Conflict(resource, details string) error {
	return rw.Error(errors.Conflict(resource, details))
}

// InternalError writes a 500 Internal Server Error
func (rw *ResponseWriter) InternalError(message string) error {
	return rw.Error(errors.InternalError(message))
}

// ValidationError writes a validation error response
func (rw *ResponseWriter) ValidationError(validationErrors *errors.ValidationErrors) error {
	if validationErrors == nil || !validationErrors.HasErrors() {
		return rw.BadRequest("Validation failed")
	}
	return rw.Error(validationErrors.ToAppError())
}

// writeJSON writes JSON response with proper headers
func (rw *ResponseWriter) writeJSON(statusCode int, data interface{}) error {
	rw.w.Header().Set("Content-Type", "application/json")
	rw.w.WriteHeader(statusCode)
	return json.NewEncoder(rw.w).Encode(data)
}

// logError logs the error appropriately
func (rw *ResponseWriter) logError(appErr *errors.AppError) {
	log := logger.GetLogger()
	
	// Create context with request information
	ctx := rw.r.Context()
	if appErr.RequestID != "" {
		ctx = WithValue(ctx, logger.RequestIDKey, appErr.RequestID)
	}
	if appErr.UserID != "" {
		ctx = WithValue(ctx, logger.UserIDKey, appErr.UserID)
	}

	logWithContext := log.WithContext(ctx)

	// Log based on error severity
	switch {
	case appErr.HTTPStatus >= 500:
		logWithContext.Error("Server error in response",
			appErr,
			"method", rw.r.Method,
			"path", rw.r.URL.Path,
			"status_code", appErr.HTTPStatus,
		)
	case appErr.HTTPStatus >= 400:
		logWithContext.Warn("Client error in response",
			"method", rw.r.Method,
			"path", rw.r.URL.Path,
			"status_code", appErr.HTTPStatus,
			"error_code", appErr.Code,
			"message", appErr.Message,
		)
	}
}

// getRequestIDFromContext extracts request ID from context or generates one
func getRequestIDFromContext(r *http.Request) string {
	if id := r.Context().Value(logger.RequestIDKey); id != nil {
		if str, ok := id.(string); ok && str != "" {
			return str
		}
	}
	
	// Try header as fallback
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	
	return ""
}

// Helper functions for context manipulation
func WithValue(ctx context.Context, key, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// Convenience functions that don't require ResponseWriter instance

// JSON writes a JSON response using the standard format
func JSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) error {
	return NewResponseWriter(w, r).JSON(statusCode, data)
}

// Success writes a success response
func Success(w http.ResponseWriter, r *http.Request, data interface{}) error {
	return NewResponseWriter(w, r).Success(data)
}

// Created writes a created response
func Created(w http.ResponseWriter, r *http.Request, data interface{}) error {
	return NewResponseWriter(w, r).Created(data)
}

// Error writes an error response
func Error(w http.ResponseWriter, r *http.Request, err error) error {
	return NewResponseWriter(w, r).Error(err)
}

// BadRequest writes a bad request error
func BadRequest(w http.ResponseWriter, r *http.Request, message string) error {
	return NewResponseWriter(w, r).BadRequest(message)
}

// Unauthorized writes an unauthorized error
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) error {
	return NewResponseWriter(w, r).Unauthorized(message)
}

// NotFound writes a not found error
func NotFound(w http.ResponseWriter, r *http.Request, resource string) error {
	return NewResponseWriter(w, r).NotFound(resource)
}

// InternalError writes an internal server error
func InternalError(w http.ResponseWriter, r *http.Request, message string) error {
	return NewResponseWriter(w, r).InternalError(message)
}