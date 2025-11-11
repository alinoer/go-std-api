package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	apiKey := "test-api-key"
	
	tests := []struct {
		name               string
		authHeader         string
		expectedStatusCode int
		expectedBody       string
		expectContextValue bool
	}{
		{
			name:               "valid bearer token",
			authHeader:         "Bearer " + apiKey,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "OK",
			expectContextValue: true,
		},
		{
			name:               "missing authorization header",
			authHeader:         "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Authorization header required\n",
			expectContextValue: false,
		},
		{
			name:               "invalid authorization format",
			authHeader:         "Basic " + apiKey,
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Invalid authorization format. Use 'Bearer <token>'\n",
			expectContextValue: false,
		},
		{
			name:               "invalid api key",
			authHeader:         "Bearer wrong-key",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Invalid API key\n",
			expectContextValue: false,
		},
		{
			name:               "empty bearer token",
			authHeader:         "Bearer ",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Invalid API key\n",
			expectContextValue: false,
		},
		{
			name:               "bearer with space only",
			authHeader:         "Bearer",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Invalid authorization format. Use 'Bearer <token>'\n",
			expectContextValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that checks context
			var contextUserID string
			var contextOK bool
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contextUserID, contextOK = GetUserIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Create middleware
			middleware := AuthMiddleware(apiKey)
			handler := middleware(testHandler)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			
			w := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			// Check response body
			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}

			// Check context value
			if tt.expectContextValue {
				if !contextOK {
					t.Error("expected user ID in context, but not found")
				}
				if contextUserID != "authenticated-user" {
					t.Errorf("expected user ID 'authenticated-user', got %q", contextUserID)
				}
			} else {
				if contextOK {
					t.Errorf("expected no user ID in context, but found %q", contextUserID)
				}
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name           string
		context        context.Context
		expectedUserID string
		expectedOK     bool
	}{
		{
			name:           "context with user ID",
			context:        context.WithValue(context.Background(), UserIDKey, "test-user"),
			expectedUserID: "test-user",
			expectedOK:     true,
		},
		{
			name:           "context without user ID",
			context:        context.Background(),
			expectedUserID: "",
			expectedOK:     false,
		},
		{
			name:           "context with wrong type",
			context:        context.WithValue(context.Background(), UserIDKey, 123),
			expectedUserID: "",
			expectedOK:     false,
		},
		{
			name:           "context with empty user ID",
			context:        context.WithValue(context.Background(), UserIDKey, ""),
			expectedUserID: "",
			expectedOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, ok := GetUserIDFromContext(tt.context)

			if ok != tt.expectedOK {
				t.Errorf("expected ok %t, got %t", tt.expectedOK, ok)
			}

			if userID != tt.expectedUserID {
				t.Errorf("expected user ID %q, got %q", tt.expectedUserID, userID)
			}
		})
	}
}

func TestAuthMiddleware_Integration(t *testing.T) {
	apiKey := "integration-test-key"
	
	// Create a chain of middlewares with a test handler
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			t.Error("expected user ID in context for authenticated request")
			return
		}
		
		w.Header().Set("X-User-ID", userID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := AuthMiddleware(apiKey)
	handler := middleware(finalHandler)

	// Test successful authentication flow
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("X-User-ID") != "authenticated-user" {
		t.Errorf("expected X-User-ID header 'authenticated-user', got %q", w.Header().Get("X-User-ID"))
	}

	if w.Body.String() != "success" {
		t.Errorf("expected body 'success', got %q", w.Body.String())
	}
}