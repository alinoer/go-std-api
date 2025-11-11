package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/google/uuid"
)

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct {
	users               map[uuid.UUID]*models.User
	usersByUsername     map[string]*models.User
	createError         error
	getByIDError        error
	getByUsernameError  error
	listError           error
	listPaginatedError  error
	listPaginatedTotal  int64
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:           make(map[uuid.UUID]*models.User),
		usersByUsername: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.createError != nil {
		return m.createError
	}
	m.users[user.ID] = user
	m.usersByUsername[user.Username] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.getByUsernameError != nil {
		return nil, m.getByUsernameError
	}
	user, exists := m.usersByUsername[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) List(ctx context.Context) ([]*models.User, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	var users []*models.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockUserRepository) ListPaginated(ctx context.Context, pagination *models.PaginationParams) ([]*models.User, int64, error) {
	if m.listPaginatedError != nil {
		return nil, 0, m.listPaginatedError
	}
	
	var users []*models.User
	count := 0
	for _, user := range m.users {
		if count >= pagination.Offset && len(users) < pagination.PageSize {
			users = append(users, user)
		}
		count++
	}
	
	return users, m.listPaginatedTotal, nil
}

// Helper methods for setting up mock behavior
func (m *MockUserRepository) SetCreateError(err error) {
	m.createError = err
}

func (m *MockUserRepository) SetGetByIDError(err error) {
	m.getByIDError = err
}

func (m *MockUserRepository) SetGetByUsernameError(err error) {
	m.getByUsernameError = err
}

func (m *MockUserRepository) SetListError(err error) {
	m.listError = err
}

func (m *MockUserRepository) SetListPaginatedError(err error) {
	m.listPaginatedError = err
}

func (m *MockUserRepository) SetListPaginatedTotal(total int64) {
	m.listPaginatedTotal = total
}

func (m *MockUserRepository) AddUser(user *models.User) {
	m.users[user.ID] = user
	m.usersByUsername[user.Username] = user
}

func TestNewUserService(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	if service == nil {
		t.Error("expected non-nil service")
	}
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.CreateUserRequest
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name: "successful user creation",
			request: &models.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mock *MockUserRepository) {
				// No existing user
			},
			expectedError: "",
		},
		{
			name: "empty username",
			request: &models.CreateUserRequest{
				Username: "",
				Password: "password123",
			},
			setupMock:     func(mock *MockUserRepository) {},
			expectedError: "username is required",
		},
		{
			name: "empty password",
			request: &models.CreateUserRequest{
				Username: "testuser",
				Password: "",
			},
			setupMock:     func(mock *MockUserRepository) {},
			expectedError: "password is required",
		},
		{
			name: "username already exists",
			request: &models.CreateUserRequest{
				Username: "existinguser",
				Password: "password123",
			},
			setupMock: func(mock *MockUserRepository) {
				existingUser := &models.User{
					ID:       uuid.New(),
					Username: "existinguser",
				}
				mock.AddUser(existingUser)
			},
			expectedError: "username already exists",
		},
		{
			name: "repository create error",
			request: &models.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(mock *MockUserRepository) {
				mock.SetCreateError(fmt.Errorf("database error"))
			},
			expectedError: "failed to create user: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository()
			tt.setupMock(mockRepo)
			service := NewUserService(mockRepo)

			user, err := service.CreateUser(context.Background(), tt.request)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("expected non-nil user")
				return
			}

			if user.Username != tt.request.Username {
				t.Errorf("expected username %s, got %s", tt.request.Username, user.Username)
			}

			// Verify password hash
			hasher := sha256.New()
			hasher.Write([]byte(tt.request.Password))
			expectedHash := fmt.Sprintf("%x", hasher.Sum(nil))

			if user.PasswordHash != expectedHash {
				t.Errorf("password hash mismatch")
			}

			if user.ID == uuid.Nil {
				t.Error("expected non-nil user ID")
			}

			if user.CreatedAt.IsZero() {
				t.Error("expected non-zero CreatedAt")
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMock     func(*MockUserRepository)
		expectedUser  *models.User
		expectedError string
	}{
		{
			name:   "successful get user",
			userID: uuid.New(),
			setupMock: func(mock *MockUserRepository) {
				user := &models.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
				mock.AddUser(user)
			},
			expectedError: "",
		},
		{
			name:   "user not found",
			userID: uuid.New(),
			setupMock: func(mock *MockUserRepository) {
				// No user added
			},
			expectedError: "user not found",
		},
		{
			name:   "repository error",
			userID: uuid.New(),
			setupMock: func(mock *MockUserRepository) {
				mock.SetGetByIDError(fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository()
			tt.setupMock(mockRepo)
			service := NewUserService(mockRepo)

			// For the successful case, we need to use the actual user ID
			var testID uuid.UUID
			if tt.name == "successful get user" {
				for id := range mockRepo.users {
					testID = id
					break
				}
			} else {
				testID = tt.userID
			}

			user, err := service.GetUser(context.Background(), testID)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("expected non-nil user")
			}
		})
	}
}

func TestUserService_ValidateCredentials(t *testing.T) {
	hasher := sha256.New()
	hasher.Write([]byte("correctpassword"))
	correctHash := fmt.Sprintf("%x", hasher.Sum(nil))

	tests := []struct {
		name          string
		username      string
		password      string
		setupMock     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:     "valid credentials",
			username: "testuser",
			password: "correctpassword",
			setupMock: func(mock *MockUserRepository) {
				user := &models.User{
					ID:           uuid.New(),
					Username:     "testuser",
					PasswordHash: correctHash,
				}
				mock.AddUser(user)
			},
			expectedError: "",
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password",
			setupMock: func(mock *MockUserRepository) {
				// No user added
			},
			expectedError: "invalid credentials",
		},
		{
			name:     "wrong password",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func(mock *MockUserRepository) {
				user := &models.User{
					ID:           uuid.New(),
					Username:     "testuser",
					PasswordHash: correctHash,
				}
				mock.AddUser(user)
			},
			expectedError: "invalid credentials",
		},
		{
			name:     "repository error",
			username: "testuser",
			password: "password",
			setupMock: func(mock *MockUserRepository) {
				mock.SetGetByUsernameError(fmt.Errorf("database error"))
			},
			expectedError: "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository()
			tt.setupMock(mockRepo)
			service := NewUserService(mockRepo)

			user, err := service.ValidateCredentials(context.Background(), tt.username, tt.password)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("expected non-nil user")
				return
			}

			if user.Username != tt.username {
				t.Errorf("expected username %s, got %s", tt.username, user.Username)
			}
		})
	}
}

func TestUserService_ListUsersPaginated(t *testing.T) {
	tests := []struct {
		name          string
		pagination    *models.PaginationParams
		setupMock     func(*MockUserRepository)
		expectedError string
		expectedTotal int64
	}{
		{
			name: "successful pagination",
			pagination: &models.PaginationParams{
				Page:     1,
				PageSize: 10,
				Offset:   0,
			},
			setupMock: func(mock *MockUserRepository) {
				mock.SetListPaginatedTotal(25)
			},
			expectedError: "",
			expectedTotal: 25,
		},
		{
			name: "repository error",
			pagination: &models.PaginationParams{
				Page:     1,
				PageSize: 10,
				Offset:   0,
			},
			setupMock: func(mock *MockUserRepository) {
				mock.SetListPaginatedError(fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository()
			tt.setupMock(mockRepo)
			service := NewUserService(mockRepo)

			response, err := service.ListUsersPaginated(context.Background(), tt.pagination)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if response == nil {
				t.Error("expected non-nil response")
				return
			}

			if response.Pagination == nil {
				t.Error("expected non-nil pagination meta")
				return
			}

			if response.Pagination.Total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, response.Pagination.Total)
			}
		})
	}
}