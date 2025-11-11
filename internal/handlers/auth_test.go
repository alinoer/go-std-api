package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/service"
	"github.com/google/uuid"
)

type MockAuthUserService struct {
	createUserError         error
	validateCredentialsUser *models.User
	validateCredentialsError error
	createdUser             *models.User
}


func (m *MockAuthUserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	if m.createUserError != nil {
		return nil, m.createUserError
	}
	return m.createdUser, nil
}

func (m *MockAuthUserService) ValidateCredentials(ctx context.Context, username, password string) (*models.User, error) {
	if m.validateCredentialsError != nil {
		return nil, m.validateCredentialsError
	}
	return m.validateCredentialsUser, nil
}

func (m *MockAuthUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, nil
}

func (m *MockAuthUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (m *MockAuthUserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return nil, nil
}

func (m *MockAuthUserService) ListUsersPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	return nil, nil
}

func TestNewAuthHandler(t *testing.T) {
	mockUserService := &MockAuthUserService{}
	authService := service.NewAuthService("test-secret-key")
	handler := NewAuthHandler(mockUserService, authService)

	if handler == nil {
		t.Error("expected non-nil auth handler")
	}
	if handler.userService == nil {
		t.Error("expected non-nil user service")
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		setupMock          func(*MockAuthUserService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "successful registration",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mockUser *MockAuthUserService) {
				mockUser.createdUser = &models.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "invalid JSON",
			requestBody:        "invalid json",
			setupMock:          func(mockUser *MockAuthUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid JSON payload",
		},
		{
			name: "empty username",
			requestBody: models.RegisterRequest{
				Username: "",
				Password: "password123",
			},
			setupMock:          func(mockUser *MockAuthUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Username is required",
		},
		{
			name: "empty password",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Password: "",
			},
			setupMock:          func(mockUser *MockAuthUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Password is required",
		},
		{
			name: "password too short",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Password: "12345",
			},
			setupMock:          func(mockUser *MockAuthUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Password must be at least 6 characters long",
		},
		{
			name: "user service error",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mockUser *MockAuthUserService) {
				mockUser.createUserError = fmt.Errorf("username already exists")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "username already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := &MockAuthUserService{}
			authService := service.NewAuthService("test-secret-key")
			tt.setupMock(mockUserService)
			handler := NewAuthHandler(mockUserService, authService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			if tt.expectedError != "" {
				var errorResp ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
					t.Errorf("failed to unmarshal error response: %v", err)
				}
				if errorResp.Error != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, errorResp.Error)
				}
			} else {
				var response models.RegisterResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal success response: %v", err)
				}
				if response.User == nil {
					t.Error("expected non-nil user in response")
				}
				if response.Message == "" {
					t.Error("expected non-empty message in response")
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		setupMock          func(*MockAuthUserService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mockUser *MockAuthUserService) {
				mockUser.validateCredentialsUser = &models.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid JSON",
			requestBody:        "invalid json",
			setupMock:          func(mockUser *MockAuthUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid JSON payload",
		},
		{
			name: "empty username",
			requestBody: models.LoginRequest{
				Username: "",
				Password: "password123",
			},
			setupMock:          func(mockUser *MockAuthUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Username is required",
		},
		{
			name: "empty password",
			requestBody: models.LoginRequest{
				Username: "testuser",
				Password: "",
			},
			setupMock:          func(mockUser *MockAuthUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Password is required",
		},
		{
			name: "invalid credentials",
			requestBody: models.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMock: func(mockUser *MockAuthUserService) {
				mockUser.validateCredentialsError = fmt.Errorf("invalid credentials")
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedError:      "Invalid username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := &MockAuthUserService{}
			authService := service.NewAuthService("test-secret-key")
			tt.setupMock(mockUserService)
			handler := NewAuthHandler(mockUserService, authService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			if tt.expectedError != "" {
				var errorResp ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
					t.Errorf("failed to unmarshal error response: %v", err)
				}
				if errorResp.Error != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, errorResp.Error)
				}
			} else {
				var response models.LoginResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal success response: %v", err)
				}
				if response.User == nil {
					t.Error("expected non-nil user in response")
				}
				if response.AccessToken == "" {
					t.Error("expected non-empty access token in response")
				}
			}
		})
	}
}