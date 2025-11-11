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

type MockPostService struct {
	createPostError               error
	getPostError                  error
	listPostsError                error
	listPostsPaginatedError       error
	getPostsByUserError           error
	getPostsByUserPaginatedError  error
	updatePostError               error
	deletePostError               error
	createdPost                   *models.Post
	retrievedPost                 *models.Post
	posts                         []*models.Post
	paginatedResponse             *models.PaginatedResponse
	updatedPost                   *models.Post
}

func (m *MockPostService) CreatePost(ctx context.Context, userID uuid.UUID, req *models.CreatePostRequest) (*models.Post, error) {
	if m.createPostError != nil {
		return nil, m.createPostError
	}
	return m.createdPost, nil
}

func (m *MockPostService) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	if m.getPostError != nil {
		return nil, m.getPostError
	}
	return m.retrievedPost, nil
}

func (m *MockPostService) ListPosts(ctx context.Context) ([]*models.Post, error) {
	if m.listPostsError != nil {
		return nil, m.listPostsError
	}
	return m.posts, nil
}

func (m *MockPostService) ListPostsPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	if m.listPostsPaginatedError != nil {
		return nil, m.listPostsPaginatedError
	}
	return m.paginatedResponse, nil
}

func (m *MockPostService) GetPostsByUser(ctx context.Context, userID uuid.UUID) ([]*models.Post, error) {
	if m.getPostsByUserError != nil {
		return nil, m.getPostsByUserError
	}
	return m.posts, nil
}

func (m *MockPostService) GetPostsByUserPaginated(ctx context.Context, userID uuid.UUID, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	if m.getPostsByUserPaginatedError != nil {
		return nil, m.getPostsByUserPaginatedError
	}
	return m.paginatedResponse, nil
}

func (m *MockPostService) UpdatePost(ctx context.Context, id uuid.UUID, req *models.UpdatePostRequest) (*models.Post, error) {
	if m.updatePostError != nil {
		return nil, m.updatePostError
	}
	return m.updatedPost, nil
}

func (m *MockPostService) DeletePost(ctx context.Context, id uuid.UUID) error {
	return m.deletePostError
}

type MockPostUserService struct {
	listUsersError error
	users          []*models.User
}

func (m *MockPostUserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	if m.listUsersError != nil {
		return nil, m.listUsersError
	}
	return m.users, nil
}

func (m *MockPostUserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	return nil, nil
}

func (m *MockPostUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, nil
}

func (m *MockPostUserService) ListUsersPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	return nil, nil
}

func (m *MockPostUserService) ValidateCredentials(ctx context.Context, username, password string) (*models.User, error) {
	return nil, nil
}

func (m *MockPostUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func TestNewPostHandler(t *testing.T) {
	mockPostService := &MockPostService{}
	mockUserService := &MockPostUserService{}
	handler := NewPostHandler(mockPostService, mockUserService)

	if handler == nil {
		t.Error("expected non-nil post handler")
	}
	if handler.postService == nil {
		t.Error("expected non-nil post service")
	}
	if handler.userService == nil {
		t.Error("expected non-nil user service")
	}
}

func TestPostHandler_CreatePost(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		setupMocks         func(*MockPostService, *MockPostUserService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name: "successful post creation",
			requestBody: models.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(postMock *MockPostService, userMock *MockPostUserService) {
				userMock.users = []*models.User{
					{ID: uuid.New(), Username: "testuser"},
				}
				postMock.createdPost = &models.Post{
					ID:      uuid.New(),
					Title:   "Test Post",
					Content: "This is a test post",
				}
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "invalid JSON",
			requestBody:        "invalid json",
			setupMocks:         func(postMock *MockPostService, userMock *MockPostUserService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid JSON payload",
		},
		{
			name: "no users found",
			requestBody: models.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(postMock *MockPostService, userMock *MockPostUserService) {
				userMock.users = []*models.User{}
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "No users found. Please create a user first",
		},
		{
			name: "user service error",
			requestBody: models.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(postMock *MockPostService, userMock *MockPostUserService) {
				userMock.listUsersError = fmt.Errorf("database error")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "No users found. Please create a user first",
		},
		{
			name: "post service error",
			requestBody: models.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(postMock *MockPostService, userMock *MockPostUserService) {
				userMock.users = []*models.User{
					{ID: uuid.New(), Username: "testuser"},
				}
				postMock.createPostError = fmt.Errorf("validation error")
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostService := &MockPostService{}
			mockUserService := &MockPostUserService{}
			tt.setupMocks(mockPostService, mockUserService)
			handler := NewPostHandler(mockPostService, mockUserService)

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

			req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreatePost(w, req)

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

func TestPostHandler_GetPost(t *testing.T) {
	validPostID := uuid.New()

	tests := []struct {
		name               string
		postID             string
		setupMock          func(*MockPostService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:   "successful get post",
			postID: validPostID.String(),
			setupMock: func(mock *MockPostService) {
				mock.retrievedPost = &models.Post{
					ID:      validPostID,
					Title:   "Test Post",
					Content: "Test Content",
				}
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid post ID",
			postID:             "invalid-uuid",
			setupMock:          func(mock *MockPostService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid post ID",
		},
		{
			name:   "post not found",
			postID: validPostID.String(),
			setupMock: func(mock *MockPostService) {
				mock.getPostError = fmt.Errorf("post not found")
			},
			expectedStatusCode: http.StatusNotFound,
			expectedError:      "Post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostService := &MockPostService{}
			mockUserService := &MockPostUserService{}
			tt.setupMock(mockPostService)
			handler := NewPostHandler(mockPostService, mockUserService)

			req := httptest.NewRequest(http.MethodGet, "/posts/"+tt.postID, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.postID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.GetPost(w, req)

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

func TestPostHandler_UpdatePost(t *testing.T) {
	validPostID := uuid.New()
	title := "Updated Title"
	content := "Updated Content"

	tests := []struct {
		name               string
		postID             string
		requestBody        interface{}
		setupMock          func(*MockPostService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:   "successful update post",
			postID: validPostID.String(),
			requestBody: models.UpdatePostRequest{
				Title:   &title,
				Content: &content,
			},
			setupMock: func(mock *MockPostService) {
				mock.updatedPost = &models.Post{
					ID:      validPostID,
					Title:   title,
					Content: content,
				}
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid post ID",
			postID:             "invalid-uuid",
			requestBody:        models.UpdatePostRequest{},
			setupMock:          func(mock *MockPostService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid post ID",
		},
		{
			name:               "invalid JSON",
			postID:             validPostID.String(),
			requestBody:        "invalid json",
			setupMock:          func(mock *MockPostService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid JSON payload",
		},
		{
			name:   "post not found",
			postID: validPostID.String(),
			requestBody: models.UpdatePostRequest{
				Title: &title,
			},
			setupMock: func(mock *MockPostService) {
				mock.updatePostError = fmt.Errorf("post not found")
			},
			expectedStatusCode: http.StatusNotFound,
			expectedError:      "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostService := &MockPostService{}
			mockUserService := &MockPostUserService{}
			tt.setupMock(mockPostService)
			handler := NewPostHandler(mockPostService, mockUserService)

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

			req := httptest.NewRequest(http.MethodPut, "/posts/"+tt.postID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.postID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.UpdatePost(w, req)

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

func TestPostHandler_DeletePost(t *testing.T) {
	validPostID := uuid.New()

	tests := []struct {
		name               string
		postID             string
		setupMock          func(*MockPostService)
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "successful delete post",
			postID:             validPostID.String(),
			setupMock:          func(mock *MockPostService) {},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "invalid post ID",
			postID:             "invalid-uuid",
			setupMock:          func(mock *MockPostService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid post ID",
		},
		{
			name:   "post not found",
			postID: validPostID.String(),
			setupMock: func(mock *MockPostService) {
				mock.deletePostError = fmt.Errorf("post not found")
			},
			expectedStatusCode: http.StatusNotFound,
			expectedError:      "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostService := &MockPostService{}
			mockUserService := &MockPostUserService{}
			tt.setupMock(mockPostService)
			handler := NewPostHandler(mockPostService, mockUserService)

			req := httptest.NewRequest(http.MethodDelete, "/posts/"+tt.postID, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.postID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.DeletePost(w, req)

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
				if response.Message == "" {
					t.Error("expected non-empty message in response")
				}
			}
		})
	}
}

func TestPostHandler_GetPostsByUser(t *testing.T) {
	validUserID := uuid.New()

	tests := []struct {
		name               string
		userID             string
		queryParams        map[string]string
		setupMock          func(*MockPostService)
		expectedStatusCode int
		expectedError      string
		expectPaginated    bool
	}{
		{
			name:   "successful get posts by user without pagination",
			userID: validUserID.String(),
			setupMock: func(mock *MockPostService) {
				mock.posts = []*models.Post{
					{ID: uuid.New(), Title: "Post 1"},
					{ID: uuid.New(), Title: "Post 2"},
				}
			},
			expectedStatusCode: http.StatusOK,
			expectPaginated:    false,
		},
		{
			name:   "successful get posts by user with pagination",
			userID: validUserID.String(),
			queryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
			setupMock: func(mock *MockPostService) {
				mock.paginatedResponse = &models.PaginatedResponse{
					Data: []*models.Post{
						{ID: uuid.New(), Title: "Post 1"},
					},
					Pagination: &models.PaginationMeta{
						Page:     1,
						PageSize: 10,
						Total:    1,
					},
				}
			},
			expectedStatusCode: http.StatusOK,
			expectPaginated:    true,
		},
		{
			name:               "invalid user ID",
			userID:             "invalid-uuid",
			setupMock:          func(mock *MockPostService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid user ID",
		},
		{
			name:   "service error",
			userID: validUserID.String(),
			setupMock: func(mock *MockPostService) {
				mock.getPostsByUserError = fmt.Errorf("database error")
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "Failed to retrieve posts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostService := &MockPostService{}
			mockUserService := &MockPostUserService{}
			tt.setupMock(mockPostService)
			handler := NewPostHandler(mockPostService, mockUserService)

			reqURL := "/users/" + tt.userID + "/posts"
			if len(tt.queryParams) > 0 {
				values := url.Values{}
				for k, v := range tt.queryParams {
					values.Add(k, v)
				}
				reqURL += "?" + values.Encode()
			}

			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userId", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.GetPostsByUser(w, req)

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