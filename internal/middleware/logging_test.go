package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		handlerStatusCode  int
		handlerBody        string
		expectedStatusCode int
	}{
		{
			name:               "GET request with 200 status",
			method:             http.MethodGet,
			path:               "/users",
			handlerStatusCode:  http.StatusOK,
			handlerBody:        "success",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "POST request with 201 status",
			method:             http.MethodPost,
			path:               "/users",
			handlerStatusCode:  http.StatusCreated,
			handlerBody:        "created",
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "GET request with 404 status",
			method:             http.MethodGet,
			path:               "/not-found",
			handlerStatusCode:  http.StatusNotFound,
			handlerBody:        "not found",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "PUT request with 500 status",
			method:             http.MethodPut,
			path:               "/users/1",
			handlerStatusCode:  http.StatusInternalServerError,
			handlerBody:        "error",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var logBuffer bytes.Buffer
			log.SetOutput(&logBuffer)
			defer log.SetOutput(os.Stderr) // Restore default output

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatusCode)
				w.Write([]byte(tt.handlerBody))
			})

			// Apply logging middleware
			handler := LoggingMiddleware(testHandler)

			// Create test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Record start time for duration check
			start := time.Now()
			
			// Execute request
			handler.ServeHTTP(w, req)
			
			// Calculate approximate duration
			duration := time.Since(start)

			// Check response
			if w.Code != tt.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			if w.Body.String() != tt.handlerBody {
				t.Errorf("expected body %q, got %q", tt.handlerBody, w.Body.String())
			}

			// Check log output
			logOutput := logBuffer.String()
			
			// Verify log contains expected elements
			expectedElements := []string{
				"[" + tt.method + "]",
				tt.path,
				string(rune('0' + tt.handlerStatusCode/100)), // First digit of status code
			}

			for _, element := range expectedElements {
				if !strings.Contains(logOutput, element) {
					t.Errorf("expected log to contain %q, got: %s", element, logOutput)
				}
			}

			// Verify duration is logged (should contain time unit)
			if !strings.Contains(logOutput, "µs") && !strings.Contains(logOutput, "ms") && !strings.Contains(logOutput, "s") {
				t.Errorf("expected log to contain duration, got: %s", logOutput)
			}

			// Verify duration is reasonable (should be less than 1 second for this simple test)
			if duration > time.Second {
				t.Errorf("request took too long: %v", duration)
			}
		})
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "status OK",
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "status not found",
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "status internal server error",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			rw := &responseWriter{
				ResponseWriter: recorder,
				statusCode:     http.StatusOK, // Default
			}

			rw.WriteHeader(tt.statusCode)

			if rw.statusCode != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, rw.statusCode)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("expected recorder status code %d, got %d", tt.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestResponseWriter_DefaultStatusCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
	}

	// Write without calling WriteHeader (should use default 200)
	rw.Write([]byte("test"))

	if rw.statusCode != http.StatusOK {
		t.Errorf("expected default status code %d, got %d", http.StatusOK, rw.statusCode)
	}
}

func TestLoggingMiddleware_Integration(t *testing.T) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)

	// Create a handler that simulates some work
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond) // Simulate some work
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "created"}`))
	})

	// Apply middleware
	handler := LoggingMiddleware(testHandler)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"username": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	expectedBody := `{"message": "created"}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}

	// Verify log
	logOutput := logBuffer.String()
	expectedLogElements := []string{
		"[POST]",
		"/api/v1/users",
		"201",
	}

	for _, element := range expectedLogElements {
		if !strings.Contains(logOutput, element) {
			t.Errorf("expected log to contain %q, got: %s", element, logOutput)
		}
	}
}