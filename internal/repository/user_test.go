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

func TestUserRepository_Create(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	repo := NewUserRepository(testDB.DB)

	tests := []struct {
		name          string
		user          *models.User
		expectedError bool
	}{
		{
			name: "successful create",
			user: &models.User{
				ID:           uuid.New(),
				Username:     "testuser1",
				PasswordHash: "hashedpassword",
				CreatedAt:    time.Now(),
			},
			expectedError: false,
		},
		{
			name: "duplicate username",
			user: &models.User{
				ID:           uuid.New(),
				Username:     "testuser1", // same username as above
				PasswordHash: "hashedpassword",
				CreatedAt:    time.Now(),
			},
			expectedError: true,
		},
		{
			name: "empty username",
			user: &models.User{
				ID:           uuid.New(),
				Username:     "",
				PasswordHash: "hashedpassword",
				CreatedAt:    time.Now(),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.user)

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

			// Verify the user was actually created
			createdUser, err := repo.GetByID(context.Background(), tt.user.ID)
			if err != nil {
				t.Errorf("failed to get created user: %v", err)
				return
			}

			if createdUser.Username != tt.user.Username {
				t.Errorf("expected username %s, got %s", tt.user.Username, createdUser.Username)
			}
			if createdUser.PasswordHash != tt.user.PasswordHash {
				t.Errorf("expected password hash %s, got %s", tt.user.PasswordHash, createdUser.PasswordHash)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	repo := NewUserRepository(testDB.DB)

	// Create a test user first
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := repo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		expectedError bool
	}{
		{
			name:          "existing user",
			userID:        testUser.ID,
			expectedError: false,
		},
		{
			name:          "non-existent user",
			userID:        uuid.New(),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByID(context.Background(), tt.userID)

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

			if user.ID != tt.userID {
				t.Errorf("expected ID %s, got %s", tt.userID, user.ID)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	repo := NewUserRepository(testDB.DB)

	// Create a test user first
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := repo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tests := []struct {
		name          string
		username      string
		expectedError bool
	}{
		{
			name:          "existing username",
			username:      testUser.Username,
			expectedError: false,
		},
		{
			name:          "non-existent username",
			username:      "nonexistent",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByUsername(context.Background(), tt.username)

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

			if user.Username != tt.username {
				t.Errorf("expected username %s, got %s", tt.username, user.Username)
			}
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	repo := NewUserRepository(testDB.DB)

	// Create test users
	testUsers := []*models.User{
		{
			ID:           uuid.New(),
			Username:     "user1",
			PasswordHash: "hash1",
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Username:     "user2",
			PasswordHash: "hash2",
			CreatedAt:    time.Now().Add(1 * time.Second),
		},
	}

	for _, user := range testUsers {
		err := repo.Create(context.Background(), user)
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
	}

	users, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}

	if len(users) != len(testUsers) {
		t.Errorf("expected %d users, got %d", len(testUsers), len(users))
	}

	// Verify users are ordered by created_at DESC
	if len(users) >= 2 {
		if users[0].CreatedAt.Before(users[1].CreatedAt) {
			t.Error("expected users to be ordered by created_at DESC")
		}
	}
}

func TestUserRepository_ListPaginated(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	repo := NewUserRepository(testDB.DB)

	// Create test users
	numUsers := 15
	for i := 0; i < numUsers; i++ {
		user := &models.User{
			ID:           uuid.New(),
			Username:     fmt.Sprintf("user%d", i),
			PasswordHash: fmt.Sprintf("hash%d", i),
			CreatedAt:    time.Now().Add(time.Duration(i) * time.Second),
		}
		err := repo.Create(context.Background(), user)
		if err != nil {
			t.Fatalf("failed to create test user %d: %v", i, err)
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
			expectedTotal: int64(numUsers),
		},
		{
			name: "second page",
			pagination: &models.PaginationParams{
				Page:     2,
				PageSize: 10,
				Offset:   10,
			},
			expectedCount: 5,
			expectedTotal: int64(numUsers),
		},
		{
			name: "larger page size",
			pagination: &models.PaginationParams{
				Page:     1,
				PageSize: 20,
				Offset:   0,
			},
			expectedCount: numUsers,
			expectedTotal: int64(numUsers),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, total, err := repo.ListPaginated(context.Background(), tt.pagination)
			if err != nil {
				t.Fatalf("failed to list users paginated: %v", err)
			}

			if len(users) != tt.expectedCount {
				t.Errorf("expected %d users, got %d", tt.expectedCount, len(users))
			}

			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}

			// Verify users are ordered by created_at DESC
			for i := 1; i < len(users); i++ {
				if users[i-1].CreatedAt.Before(users[i].CreatedAt) {
					t.Error("expected users to be ordered by created_at DESC")
					break
				}
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkUserRepository_Create(b *testing.B) {
	testutils.SkipIfShort(b)

	testDB := testutils.SetupTestDB(b)
	defer testDB.Cleanup(b)

	repo := NewUserRepository(testDB.DB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &models.User{
			ID:           uuid.New(),
			Username:     fmt.Sprintf("benchuser%d", i),
			PasswordHash: "benchhash",
			CreatedAt:    time.Now(),
		}
		repo.Create(context.Background(), user)
	}
}

func BenchmarkUserRepository_GetByID(b *testing.B) {
	testutils.SkipIfShort(b)

	testDB := testutils.SetupTestDB(b)
	defer testDB.Cleanup(b)

	repo := NewUserRepository(testDB.DB)

	// Create a test user
	testUser := &models.User{
		ID:           uuid.New(),
		Username:     "benchuser",
		PasswordHash: "benchhash",
		CreatedAt:    time.Now(),
	}
	repo.Create(context.Background(), testUser)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByID(context.Background(), testUser.ID)
	}
}