package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/testutils"
	"github.com/google/uuid"
)

func TestPostRepository_Create(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	userRepo := NewUserRepository(testDB.DB)
	postRepo := NewPostRepository(testDB.DB)

	// Create a test user first
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := userRepo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name          string
		post          *models.Post
		expectedError bool
	}{
		{
			name: "successful create",
			post: &models.Post{
				ID:        uuid.New(),
				UserID:    testUser.ID,
				Title:     "Test Post",
				Content:   "This is a test post",
				CreatedAt: time.Now(),
			},
			expectedError: false,
		},
		{
			name: "invalid user ID",
			post: &models.Post{
				ID:        uuid.New(),
				UserID:    uuid.New(), // non-existent user
				Title:     "Test Post",
				Content:   "This is a test post",
				CreatedAt: time.Now(),
			},
			expectedError: true,
		},
		{
			name: "empty title",
			post: &models.Post{
				ID:        uuid.New(),
				UserID:    testUser.ID,
				Title:     "",
				Content:   "This is a test post",
				CreatedAt: time.Now(),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postRepo.Create(context.Background(), tt.post)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the post was actually created
			createdPost, err := postRepo.GetByID(context.Background(), tt.post.ID)
			if err != nil {
				t.Errorf("failed to get created post: %v", err)
				return
			}

			if createdPost.Title != tt.post.Title {
				t.Errorf("expected title %s, got %s", tt.post.Title, createdPost.Title)
			}
			if createdPost.Content != tt.post.Content {
				t.Errorf("expected content %s, got %s", tt.post.Content, createdPost.Content)
			}
			if createdPost.UserID != tt.post.UserID {
				t.Errorf("expected user ID %s, got %s", tt.post.UserID, createdPost.UserID)
			}
		})
	}
}

func TestPostRepository_GetByID(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	userRepo := NewUserRepository(testDB.DB)
	postRepo := NewPostRepository(testDB.DB)

	// Create a test user first
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := userRepo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create a test post
	testPost := &models.Post{
		ID:        uuid.New(),
		UserID:    testUser.ID,
		Title:     "Test Post",
		Content:   "This is a test post",
		CreatedAt: time.Now(),
	}
	err = postRepo.Create(context.Background(), testPost)
	if err != nil {
		t.Fatalf("failed to create test post: %v", err)
	}

	tests := []struct {
		name          string
		postID        uuid.UUID
		expectedError bool
	}{
		{
			name:          "existing post",
			postID:        testPost.ID,
			expectedError: false,
		},
		{
			name:          "non-existent post",
			postID:        uuid.New(),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := postRepo.GetByID(context.Background(), tt.postID)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if post.ID != tt.postID {
				t.Errorf("expected ID %s, got %s", tt.postID, post.ID)
			}
		})
	}
}

func TestPostRepository_Update(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	userRepo := NewUserRepository(testDB.DB)
	postRepo := NewPostRepository(testDB.DB)

	// Create a test user first
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := userRepo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create a test post
	testPost := &models.Post{
		ID:        uuid.New(),
		UserID:    testUser.ID,
		Title:     "Original Title",
		Content:   "Original Content",
		CreatedAt: time.Now(),
	}
	err = postRepo.Create(context.Background(), testPost)
	if err != nil {
		t.Fatalf("failed to create test post: %v", err)
	}

	tests := []struct {
		name          string
		postID        uuid.UUID
		updatedPost   *models.Post
		expectedError bool
	}{
		{
			name:   "successful update",
			postID: testPost.ID,
			updatedPost: &models.Post{
				ID:      testPost.ID,
				UserID:  testUser.ID,
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			expectedError: false,
		},
		{
			name:   "non-existent post",
			postID: uuid.New(),
			updatedPost: &models.Post{
				ID:      uuid.New(),
				UserID:  testUser.ID,
				Title:   "Updated Title",
				Content: "Updated Content",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postRepo.Update(context.Background(), tt.postID, tt.updatedPost)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the post was actually updated
			updatedPost, err := postRepo.GetByID(context.Background(), tt.postID)
			if err != nil {
				t.Errorf("failed to get updated post: %v", err)
				return
			}

			if updatedPost.Title != tt.updatedPost.Title {
				t.Errorf("expected title %s, got %s", tt.updatedPost.Title, updatedPost.Title)
			}
			if updatedPost.Content != tt.updatedPost.Content {
				t.Errorf("expected content %s, got %s", tt.updatedPost.Content, updatedPost.Content)
			}
		})
	}
}

func TestPostRepository_Delete(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	userRepo := NewUserRepository(testDB.DB)
	postRepo := NewPostRepository(testDB.DB)

	// Create a test user first
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := userRepo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name          string
		setupPost     bool
		expectedError bool
	}{
		{
			name:          "successful delete",
			setupPost:     true,
			expectedError: false,
		},
		{
			name:          "non-existent post",
			setupPost:     false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var postID uuid.UUID

			if tt.setupPost {
				// Create a test post
				testPost := &models.Post{
					ID:        uuid.New(),
					UserID:    testUser.ID,
					Title:     "Test Post",
					Content:   "This is a test post",
					CreatedAt: time.Now(),
				}
				err = postRepo.Create(context.Background(), testPost)
				if err != nil {
					t.Fatalf("failed to create test post: %v", err)
				}
				postID = testPost.ID
			} else {
				postID = uuid.New()
			}

			err := postRepo.Delete(context.Background(), postID)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the post was actually deleted
			_, err = postRepo.GetByID(context.Background(), postID)
			if err == nil {
				t.Error("expected error when getting deleted post, got nil")
			}
		})
	}
}

func TestPostRepository_GetByUserID(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	userRepo := NewUserRepository(testDB.DB)
	postRepo := NewPostRepository(testDB.DB)

	// Create test users
	user1 := &models.User{
		ID:           uuid.New(),
		Username:     "user1",
		PasswordHash: "hash1",
		CreatedAt:    time.Now(),
	}
	user2 := &models.User{
		ID:           uuid.New(),
		Username:     "user2",
		PasswordHash: "hash2",
		CreatedAt:    time.Now(),
	}

	err := userRepo.Create(context.Background(), user1)
	if err != nil {
		t.Fatalf("failed to create user1: %v", err)
	}
	err = userRepo.Create(context.Background(), user2)
	if err != nil {
		t.Fatalf("failed to create user2: %v", err)
	}

	// Create posts for user1
	for i := 0; i < 3; i++ {
		post := &models.Post{
			ID:        uuid.New(),
			UserID:    user1.ID,
			Title:     fmt.Sprintf("Post %d for User 1", i+1),
			Content:   fmt.Sprintf("Content %d", i+1),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		err = postRepo.Create(context.Background(), post)
		if err != nil {
			t.Fatalf("failed to create post %d: %v", i+1, err)
		}
	}

	// Create posts for user2
	for i := 0; i < 2; i++ {
		post := &models.Post{
			ID:        uuid.New(),
			UserID:    user2.ID,
			Title:     fmt.Sprintf("Post %d for User 2", i+1),
			Content:   fmt.Sprintf("Content %d", i+1),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		err = postRepo.Create(context.Background(), post)
		if err != nil {
			t.Fatalf("failed to create post %d for user2: %v", i+1, err)
		}
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		expectedCount int
	}{
		{
			name:          "user1 posts",
			userID:        user1.ID,
			expectedCount: 3,
		},
		{
			name:          "user2 posts",
			userID:        user2.ID,
			expectedCount: 2,
		},
		{
			name:          "non-existent user",
			userID:        uuid.New(),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, err := postRepo.GetByUserID(context.Background(), tt.userID)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(posts) != tt.expectedCount {
				t.Errorf("expected %d posts, got %d", tt.expectedCount, len(posts))
			}

			// Verify all posts belong to the correct user
			for _, post := range posts {
				if post.UserID != tt.userID {
					t.Errorf("expected user ID %s, got %s", tt.userID, post.UserID)
				}
			}

			// Verify posts are ordered by created_at DESC
			for i := 1; i < len(posts); i++ {
				if posts[i-1].CreatedAt.Before(posts[i].CreatedAt) {
					t.Error("expected posts to be ordered by created_at DESC")
					break
				}
			}
		})
	}
}

func TestPostRepository_ListPaginated(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	userRepo := NewUserRepository(testDB.DB)
	postRepo := NewPostRepository(testDB.DB)

	// Create a test user
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := userRepo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create test posts
	numPosts := 15
	for i := 0; i < numPosts; i++ {
		post := &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Title:     fmt.Sprintf("Post %d", i+1),
			Content:   fmt.Sprintf("Content %d", i+1),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		err = postRepo.Create(context.Background(), post)
		if err != nil {
			t.Fatalf("failed to create post %d: %v", i+1, err)
		}
	}

	tests := []struct {
		name           string
		pagination     *models.PaginationParams
		expectedCount  int
		expectedTotal  int64
	}{
		{
			name: "first page",
			pagination: &models.PaginationParams{
				Page:     1,
				PageSize: 10,
				Offset:   0,
			},
			expectedCount: 10,
			expectedTotal: int64(numPosts),
		},
		{
			name: "second page",
			pagination: &models.PaginationParams{
				Page:     2,
				PageSize: 10,
				Offset:   10,
			},
			expectedCount: 5,
			expectedTotal: int64(numPosts),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts, total, err := postRepo.ListPaginated(context.Background(), tt.pagination)
			if err != nil {
				t.Fatalf("failed to list posts paginated: %v", err)
			}

			if len(posts) != tt.expectedCount {
				t.Errorf("expected %d posts, got %d", tt.expectedCount, len(posts))
			}

			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}

			// Verify posts are ordered by created_at DESC
			for i := 1; i < len(posts); i++ {
				if posts[i-1].CreatedAt.Before(posts[i].CreatedAt) {
					t.Error("expected posts to be ordered by created_at DESC")
					break
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkPostRepository_Create(b *testing.B) {
	testutils.SkipIfShort(b)

	testDB := testutils.SetupTestDB(b)
	defer testDB.Cleanup(b)

	userRepo := NewUserRepository(testDB.DB)
	postRepo := NewPostRepository(testDB.DB)

	// Create a test user
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "benchuser",
		PasswordHash: "benchhash",
		CreatedAt:    time.Now(),
	}
	userRepo.Create(context.Background(), testUser)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		post := &models.Post{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Title:     fmt.Sprintf("Bench Post %d", i),
			Content:   "Bench Content",
			CreatedAt: time.Now(),
		}
		postRepo.Create(context.Background(), post)
	}
}