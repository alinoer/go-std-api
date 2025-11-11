package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MockUserHandlerService struct {
	createUserError           error
	getUserError              error
	listUsersError            error
	listUsersPaginatedError   error
	createdUser               *models.User
	retrievedUser             *models.User
	users                     []*models.User
	paginatedResponse         *models.PaginatedResponse
}

func (m *MockUserHandlerService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	if m.createUserError != nil {
		return nil, m.createUserError
	}
	return m.createdUser, nil
}

func (m *MockUserHandlerService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getUserError != nil {
		return nil, m.getUserError
	}
	return m.retrievedUser, nil
}

func (m *MockUserHandlerService) ListUsers(ctx context.Context) ([]*models.User, error) {
	if m.listUsersError != nil {
		return nil, m.listUsersError
	}
	return m.users, nil
}

func (m *MockUserHandlerService) ListUsersPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	if m.listUsersPaginatedError != nil {
		return nil, m.listUsersPaginatedError
	}
	return m.paginatedResponse, nil
}

func (m *MockUserHandlerService) ValidateCredentials(ctx context.Context, username, password string) (*models.User, error) {
	return nil, nil
}

func (m *MockUserHandlerService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func TestNewUserHandler(t *testing.T) {
	mockService := &MockUserHandlerService{}
	handler := NewUserHandler(mockService)

	if handler == nil {
		t.Error("expected non-nil user handler")
	}
	if handler.userService == nil {
		t.Error("expected non-nil user service")
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		setupMock          func(*MockUserHandlerService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "successful user creation",
			requestBody: models.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mock *MockUserHandlerService) {
				mock.createdUser = &models.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "invalid JSON",
			requestBody:        "invalid json",
			setupMock:          func(mock *MockUserHandlerService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid JSON payload",
		},
		{
			name: "service error",
			requestBody: models.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mock *MockUserHandlerService) {
				mock.createUserError = fmt.Errorf("username already exists")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "username already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserHandlerService{}
			tt.setupMock(mockService)
			handler := NewUserHandler(mockService)

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

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateUser(w, req)

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
				var response SuccessResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal success response: %v", err)
				}
				if response.Data == nil {
					t.Error("expected non-nil data in response")
				}
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	validUserID := uuid.New()

	tests := []struct {
		name               string
		userID             string
		setupMock          func(*MockUserHandlerService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:   "successful get user",
			userID: validUserID.String(),
			setupMock: func(mock *MockUserHandlerService) {
				mock.retrievedUser = &models.User{
					ID:       validUserID,
					Username: "testuser",
				}
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid user ID",
			userID:             "invalid-uuid",
			setupMock:          func(mock *MockUserHandlerService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid user ID",
		},
		{
			name:   "user not found",
			userID: validUserID.String(),
			setupMock: func(mock *MockUserHandlerService) {
				mock.getUserError = fmt.Errorf("user not found")
			},
			expectedStatusCode: http.StatusNotFound,
			expectedError:      "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserHandlerService{}
			tt.setupMock(mockService)
			handler := NewUserHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.GetUser(w, req)

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
				var response SuccessResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal success response: %v", err)
				}
				if response.Data == nil {
					t.Error("expected non-nil data in response")
				}
			}
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name               string
		queryParams        map[string]string
		setupMock          func(*MockUserHandlerService)
		expectedStatusCode int
		expectedError      string
		expectPaginated    bool
	}{
		{
			name: "successful list users without pagination",
			setupMock: func(mock *MockUserHandlerService) {
				mock.users = []*models.User{
					{ID: uuid.New(), Username: "user1"},
					{ID: uuid.New(), Username: "user2"},
				}
			},
			expectedStatusCode: http.StatusOK,
			expectPaginated:    false,
		},
		{
			name: "successful list users with pagination",
			queryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
			setupMock: func(mock *MockUserHandlerService) {
				mock.paginatedResponse = &models.PaginatedResponse{
					Data: []*models.User{
						{ID: uuid.New(), Username: "user1"},
					},
					Pagination: &models.PaginationMeta{
						Page:      1,
						PageSize:  10,
						Total:     1,
					},
				}
			},
			expectedStatusCode: http.StatusOK,
			expectPaginated:    true,
		},
		{
			name: "service error",
			setupMock: func(mock *MockUserHandlerService) {
				mock.listUsersError = fmt.Errorf("database error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "Failed to retrieve users",
		},
		{
			name: "paginated service error",
			queryParams: map[string]string{
				"page": "1",
			},
			setupMock: func(mock *MockUserHandlerService) {
				mock.listUsersPaginatedError = fmt.Errorf("database error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "Failed to retrieve users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserHandlerService{}
			tt.setupMock(mockService)
			handler := NewUserHandler(mockService)

			reqURL := "/users"
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for k, v := range tt.queryParams {
					values.Add(k, v)
				}
				reqURL += "?" + values.Encode()
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			handler.ListUsers(w, req)

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
				if tt.expectPaginated {
					var response models.PaginatedResponse
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("failed to unmarshal paginated response: %v", err)
					}
					if response.Data == nil {
						t.Error("expected non-nil data in response")
					}
					if response.Pagination == nil {
						t.Error("expected non-nil pagination in response")
					}
				} else {
					var response SuccessResponse
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("failed to unmarshal success response: %v", err)
					}
					if response.Data == nil {
						t.Error("expected non-nil data in response")
					}
				}
			}
		})
	}
}