package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		data           interface{}
		expectedStatus int
		expectedData   interface{}
	}{
		{
			name:           "write success response",
			statusCode:     http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedData:   map[string]interface{}{"message": "success"},
		},
		{
			name:           "write error response",
			statusCode:     http.StatusBadRequest,
			data:           ErrorResponse{Error: "bad request"},
			expectedStatus: http.StatusBadRequest,
			expectedData:   map[string]interface{}{"error": "bad request"},
		},
		{
			name:           "write string data",
			statusCode:     http.StatusCreated,
			data:           "created successfully",
			expectedStatus: http.StatusCreated,
			expectedData:   "created successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteJSON(w, tt.statusCode, tt.data)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}

			var responseData interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &responseData); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			expectedJSON, _ := json.Marshal(tt.expectedData)
			actualJSON, _ := json.Marshal(responseData)

			if string(expectedJSON) != string(actualJSON) {
				t.Errorf("expected response %s, got %s", string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		message        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "bad request error",
			statusCode:     http.StatusBadRequest,
			message:        "invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:           "internal server error",
			statusCode:     http.StatusInternalServerError,
			message:        "database connection failed",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database connection failed",
		},
		{
			name:           "not found error",
			statusCode:     http.StatusNotFound,
			message:        "user not found",
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteError(w, tt.statusCode, tt.message)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			var errorResp ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			if errorResp.Error != tt.expectedError {
				t.Errorf("expected error message %s, got %s", tt.expectedError, errorResp.Error)
			}
		})
	}
}

func TestWriteSuccess(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		expectedData interface{}
	}{
		{
			name:         "write user data",
			data:         map[string]interface{}{"id": 1, "username": "test"},
			expectedData: map[string]interface{}{"id": float64(1), "username": "test"},
		},
		{
			name:         "write array data",
			data:         []string{"item1", "item2"},
			expectedData: []interface{}{"item1", "item2"},
		},
		{
			name:         "write nil data",
			data:         nil,
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteSuccess(w, tt.data)

			if w.Code != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
			}

			var successResp SuccessResponse
			if err := json.Unmarshal(w.Body.Bytes(), &successResp); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			expectedJSON, _ := json.Marshal(tt.expectedData)
			actualJSON, _ := json.Marshal(successResp.Data)

			if string(expectedJSON) != string(actualJSON) {
				t.Errorf("expected data %s, got %s", string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func TestWriteMessage(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "success message",
			message:         "operation completed successfully",
			expectedMessage: "operation completed successfully",
		},
		{
			name:            "empty message",
			message:         "",
			expectedMessage: "",
		},
		{
			name:            "unicode message",
			message:         "操作成功",
			expectedMessage: "操作成功",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			WriteMessage(w, tt.message)

			if w.Code != http.StatusOK {
				t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
			}

			var successResp SuccessResponse
			if err := json.Unmarshal(w.Body.Bytes(), &successResp); err != nil {
				t.Errorf("failed to unmarshal response: %v", err)
			}

			if successResp.Message != tt.expectedMessage {
				t.Errorf("expected message %s, got %s", tt.expectedMessage, successResp.Message)
			}

			if successResp.Data != nil {
				t.Errorf("expected nil data, got %v", successResp.Data)
			}
		})
	}
}