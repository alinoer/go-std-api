package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/google/uuid"
)

type MockPostRepository struct {
	posts                         map[uuid.UUID]*models.Post
	postsByUser                   map[uuid.UUID][]*models.Post
	createError                   error
	getByIDError                  error
	listError                     error
	listPaginatedError            error
	getByUserIDError              error
	getByUserIDPaginatedError     error
	updateError                   error
	deleteError                   error
	listPaginatedTotal            int64
	getByUserIDPaginatedTotal     int64
}

func NewMockPostRepository() *MockPostRepository {
	return &MockPostRepository{
		posts:       make(map[uuid.UUID]*models.Post),
		postsByUser: make(map[uuid.UUID][]*models.Post),
	}
}

func (m *MockPostRepository) Create(ctx context.Context, post *models.Post) error {
	if m.createError != nil {
		return m.createError
	}
	m.posts[post.ID] = post
	m.postsByUser[post.UserID] = append(m.postsByUser[post.UserID], post)
	return nil
}

func (m *MockPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	post, exists := m.posts[id]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

func (m *MockPostRepository) List(ctx context.Context) ([]*models.Post, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	var posts []*models.Post
	for _, post := range m.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func (m *MockPostRepository) ListPaginated(ctx context.Context, pagination *models.PaginationParams) ([]*models.Post, int64, error) {
	if m.listPaginatedError != nil {
		return nil, 0, m.listPaginatedError
	}
	
	var posts []*models.Post
	count := 0
	for _, post := range m.posts {
		if count >= pagination.Offset && len(posts) < pagination.PageSize {
			posts = append(posts, post)
		}
		count++
	}
	
	return posts, m.listPaginatedTotal, nil
}

func (m *MockPostRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Post, error) {
	if m.getByUserIDError != nil {
		return nil, m.getByUserIDError
	}
	return m.postsByUser[userID], nil
}

func (m *MockPostRepository) GetByUserIDPaginated(ctx context.Context, userID uuid.UUID, pagination *models.PaginationParams) ([]*models.Post, int64, error) {
	if m.getByUserIDPaginatedError != nil {
		return nil, 0, m.getByUserIDPaginatedError
	}
	
	posts := m.postsByUser[userID]
	var paginatedPosts []*models.Post
	
	start := pagination.Offset
	end := start + pagination.PageSize
	if start < len(posts) {
		if end > len(posts) {
			end = len(posts)
		}
		paginatedPosts = posts[start:end]
	}
	
	return paginatedPosts, m.getByUserIDPaginatedTotal, nil
}

func (m *MockPostRepository) Update(ctx context.Context, id uuid.UUID, post *models.Post) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, exists := m.posts[id]; !exists {
		return fmt.Errorf("post not found")
	}
	m.posts[id] = post
	return nil
}

func (m *MockPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, exists := m.posts[id]; !exists {
		return fmt.Errorf("post not found")
	}
	delete(m.posts, id)
	return nil
}

func (m *MockPostRepository) SetCreateError(err error) {
	m.createError = err
}

func (m *MockPostRepository) SetGetByIDError(err error) {
	m.getByIDError = err
}

func (m *MockPostRepository) SetListError(err error) {
	m.listError = err
}

func (m *MockPostRepository) SetListPaginatedError(err error) {
	m.listPaginatedError = err
}

func (m *MockPostRepository) SetGetByUserIDError(err error) {
	m.getByUserIDError = err
}

func (m *MockPostRepository) SetGetByUserIDPaginatedError(err error) {
	m.getByUserIDPaginatedError = err
}

func (m *MockPostRepository) SetUpdateError(err error) {
	m.updateError = err
}

func (m *MockPostRepository) SetDeleteError(err error) {
	m.deleteError = err
}

func (m *MockPostRepository) SetListPaginatedTotal(total int64) {
	m.listPaginatedTotal = total
}

func (m *MockPostRepository) SetGetByUserIDPaginatedTotal(total int64) {
	m.getByUserIDPaginatedTotal = total
}

func (m *MockPostRepository) AddPost(post *models.Post) {
	m.posts[post.ID] = post
	m.postsByUser[post.UserID] = append(m.postsByUser[post.UserID], post)
}

type MockPostUserRepository struct {
	users        map[uuid.UUID]*models.User
	getByIDError error
}

func NewMockPostUserRepository() *MockPostUserRepository {
	return &MockPostUserRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *MockPostUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *MockPostUserRepository) Create(ctx context.Context, user *models.User) error {
	return nil
}

func (m *MockPostUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (m *MockPostUserRepository) List(ctx context.Context) ([]*models.User, error) {
	return nil, nil
}

func (m *MockPostUserRepository) ListPaginated(ctx context.Context, pagination *models.PaginationParams) ([]*models.User, int64, error) {
	return nil, 0, nil
}

func (m *MockPostUserRepository) SetGetByIDError(err error) {
	m.getByIDError = err
}

func (m *MockPostUserRepository) AddUser(user *models.User) {
	m.users[user.ID] = user
}

func TestNewPostService(t *testing.T) {
	mockPostRepo := NewMockPostRepository()
	mockUserRepo := NewMockPostUserRepository()
	service := NewPostService(mockPostRepo, mockUserRepo)

	if service == nil {
		t.Error("expected non-nil service")
	}
}

func TestPostService_CreatePost(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		request       *models.CreatePostRequest
		setupMocks    func(*MockPostRepository, *MockPostUserRepository)
		expectedError string
	}{
		{
			name:   "successful post creation",
			userID: userID,
			request: &models.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {
				user := &models.User{
					ID:       userID,
					Username: "testuser",
				}
				userRepo.AddUser(user)
			},
			expectedError: "",
		},
		{
			name:   "empty title",
			userID: userID,
			request: &models.CreatePostRequest{
				Title:   "",
				Content: "This is a test post",
			},
			setupMocks:    func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {},
			expectedError: "title is required",
		},
		{
			name:   "empty content",
			userID: userID,
			request: &models.CreatePostRequest{
				Title:   "Test Post",
				Content: "",
			},
			setupMocks:    func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {},
			expectedError: "content is required",
		},
		{
			name:   "user not found",
			userID: userID,
			request: &models.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {
				userRepo.SetGetByIDError(fmt.Errorf("user not found"))
			},
			expectedError: "user not found",
		},
		{
			name:   "repository create error",
			userID: userID,
			request: &models.CreatePostRequest{
				Title:   "Test Post",
				Content: "This is a test post",
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {
				user := &models.User{
					ID:       userID,
					Username: "testuser",
				}
				userRepo.AddUser(user)
				postRepo.SetCreateError(fmt.Errorf("database error"))
			},
			expectedError: "failed to create post: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := NewMockPostRepository()
			mockUserRepo := NewMockPostUserRepository()
			tt.setupMocks(mockPostRepo, mockUserRepo)
			service := NewPostService(mockPostRepo, mockUserRepo)

			post, err := service.CreatePost(context.Background(), tt.userID, tt.request)

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

			if post == nil {
				t.Error("expected non-nil post")
				return
			}

			if post.Title != tt.request.Title {
				t.Errorf("expected title %s, got %s", tt.request.Title, post.Title)
			}

			if post.Content != tt.request.Content {
				t.Errorf("expected content %s, got %s", tt.request.Content, post.Content)
			}

			if post.UserID != tt.userID {
				t.Errorf("expected user ID %s, got %s", tt.userID, post.UserID)
			}

			if post.ID == uuid.Nil {
				t.Error("expected non-nil post ID")
			}

			if post.CreatedAt.IsZero() {
				t.Error("expected non-zero CreatedAt")
			}
		})
	}
}

func TestPostService_GetPost(t *testing.T) {
	postID := uuid.New()

	tests := []struct {
		name          string
		postID        uuid.UUID
		setupMock     func(*MockPostRepository)
		expectedError string
	}{
		{
			name:   "successful get post",
			postID: postID,
			setupMock: func(mock *MockPostRepository) {
				post := &models.Post{
					ID:      postID,
					Title:   "Test Post",
					Content: "Test Content",
				}
				mock.AddPost(post)
			},
			expectedError: "",
		},
		{
			name:   "post not found",
			postID: postID,
			setupMock: func(mock *MockPostRepository) {
				// No post added
			},
			expectedError: "post not found",
		},
		{
			name:   "repository error",
			postID: postID,
			setupMock: func(mock *MockPostRepository) {
				mock.SetGetByIDError(fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := NewMockPostRepository()
			mockUserRepo := NewMockPostUserRepository()
			tt.setupMock(mockPostRepo)
			service := NewPostService(mockPostRepo, mockUserRepo)

			post, err := service.GetPost(context.Background(), tt.postID)

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

			if post == nil {
				t.Error("expected non-nil post")
			}
		})
	}
}

func TestPostService_UpdatePost(t *testing.T) {
	postID := uuid.New()
	newTitle := "Updated Title"
	newContent := "Updated Content"

	tests := []struct {
		name          string
		postID        uuid.UUID
		request       *models.UpdatePostRequest
		setupMock     func(*MockPostRepository)
		expectedError string
	}{
		{
			name:   "successful update with both fields",
			postID: postID,
			request: &models.UpdatePostRequest{
				Title:   &newTitle,
				Content: &newContent,
			},
			setupMock: func(mock *MockPostRepository) {
				post := &models.Post{
					ID:        postID,
					Title:     "Original Title",
					Content:   "Original Content",
					CreatedAt: time.Now(),
				}
				mock.AddPost(post)
			},
			expectedError: "",
		},
		{
			name:   "successful update with only title",
			postID: postID,
			request: &models.UpdatePostRequest{
				Title: &newTitle,
			},
			setupMock: func(mock *MockPostRepository) {
				post := &models.Post{
					ID:        postID,
					Title:     "Original Title",
					Content:   "Original Content",
					CreatedAt: time.Now(),
				}
				mock.AddPost(post)
			},
			expectedError: "",
		},
		{
			name:   "post not found",
			postID: postID,
			request: &models.UpdatePostRequest{
				Title: &newTitle,
			},
			setupMock: func(mock *MockPostRepository) {
				mock.SetGetByIDError(fmt.Errorf("post not found"))
			},
			expectedError: "post not found",
		},
		{
			name:   "repository update error",
			postID: postID,
			request: &models.UpdatePostRequest{
				Title: &newTitle,
			},
			setupMock: func(mock *MockPostRepository) {
				post := &models.Post{
					ID:        postID,
					Title:     "Original Title",
					Content:   "Original Content",
					CreatedAt: time.Now(),
				}
				mock.AddPost(post)
				mock.SetUpdateError(fmt.Errorf("database error"))
			},
			expectedError: "failed to update post: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := NewMockPostRepository()
			mockUserRepo := NewMockPostUserRepository()
			tt.setupMock(mockPostRepo)
			service := NewPostService(mockPostRepo, mockUserRepo)

			post, err := service.UpdatePost(context.Background(), tt.postID, tt.request)

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

			if post == nil {
				t.Error("expected non-nil post")
				return
			}

			if tt.request.Title != nil && post.Title != *tt.request.Title {
				t.Errorf("expected title %s, got %s", *tt.request.Title, post.Title)
			}

			if tt.request.Content != nil && post.Content != *tt.request.Content {
				t.Errorf("expected content %s, got %s", *tt.request.Content, post.Content)
			}
		})
	}
}

func TestPostService_DeletePost(t *testing.T) {
	postID := uuid.New()

	tests := []struct {
		name          string
		postID        uuid.UUID
		setupMock     func(*MockPostRepository)
		expectedError string
	}{
		{
			name:   "successful delete",
			postID: postID,
			setupMock: func(mock *MockPostRepository) {
				post := &models.Post{
					ID: postID,
				}
				mock.AddPost(post)
			},
			expectedError: "",
		},
		{
			name:   "post not found",
			postID: postID,
			setupMock: func(mock *MockPostRepository) {
				mock.SetDeleteError(fmt.Errorf("post not found"))
			},
			expectedError: "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := NewMockPostRepository()
			mockUserRepo := NewMockPostUserRepository()
			tt.setupMock(mockPostRepo)
			service := NewPostService(mockPostRepo, mockUserRepo)

			err := service.DeletePost(context.Background(), tt.postID)

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
			}
		})
	}
}

func TestPostService_ListPostsPaginated(t *testing.T) {
	tests := []struct {
		name          string
		pagination    *models.PaginationParams
		setupMock     func(*MockPostRepository)
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
			setupMock: func(mock *MockPostRepository) {
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
			setupMock: func(mock *MockPostRepository) {
				mock.SetListPaginatedError(fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := NewMockPostRepository()
			mockUserRepo := NewMockPostUserRepository()
			tt.setupMock(mockPostRepo)
			service := NewPostService(mockPostRepo, mockUserRepo)

			response, err := service.ListPostsPaginated(context.Background(), tt.pagination)

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

func TestPostService_GetPostsByUserPaginated(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		pagination    *models.PaginationParams
		setupMocks    func(*MockPostRepository, *MockPostUserRepository)
		expectedError string
		expectedTotal int64
	}{
		{
			name:   "successful pagination",
			userID: userID,
			pagination: &models.PaginationParams{
				Page:     1,
				PageSize: 10,
				Offset:   0,
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {
				user := &models.User{
					ID:       userID,
					Username: "testuser",
				}
				userRepo.AddUser(user)
				postRepo.SetGetByUserIDPaginatedTotal(10)
			},
			expectedError: "",
			expectedTotal: 10,
		},
		{
			name:   "user not found",
			userID: userID,
			pagination: &models.PaginationParams{
				Page:     1,
				PageSize: 10,
				Offset:   0,
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {
				userRepo.SetGetByIDError(fmt.Errorf("user not found"))
			},
			expectedError: "user not found",
		},
		{
			name:   "repository error",
			userID: userID,
			pagination: &models.PaginationParams{
				Page:     1,
				PageSize: 10,
				Offset:   0,
			},
			setupMocks: func(postRepo *MockPostRepository, userRepo *MockPostUserRepository) {
				user := &models.User{
					ID:       userID,
					Username: "testuser",
				}
				userRepo.AddUser(user)
				postRepo.SetGetByUserIDPaginatedError(fmt.Errorf("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := NewMockPostRepository()
			mockUserRepo := NewMockPostUserRepository()
			tt.setupMocks(mockPostRepo, mockUserRepo)
			service := NewPostService(mockPostRepo, mockUserRepo)

			response, err := service.GetPostsByUserPaginated(context.Background(), tt.userID, tt.pagination)

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