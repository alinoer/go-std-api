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
	"github.com/alinoer/go-std-api/internal/repository"
	"github.com/alinoer/go-std-api/internal/service"
	"github.com/alinoer/go-std-api/internal/testutils"
	"github.com/go-chi/chi/v5"
)

func TestUserHandler_Integration(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	// Setup test database
	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	// Setup real dependencies
	userRepo := repository.NewUserRepository(testDB.DB)
	userService := service.NewUserService(userRepo)
	handler := NewUserHandler(userService)

	t.Run("complete user workflow", func(t *testing.T) {
		// Test 1: Create user
		createReq := models.CreateUserRequest{
			Username: "integrationuser",
			Password: "password123",
		}
		
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		var createResp struct {
			User *models.User `json:"user"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &createResp); err != nil {
			t.Fatalf("failed to unmarshal create response: %v", err)
		}

		userID := createResp.User.ID

		// Test 2: Get user by ID
		req = httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
		w = httptest.NewRecorder()
		
		// Add router context for path parameter
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", userID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		
		handler.GetUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Test 3: List users (should contain our created user)
		req = httptest.NewRequest(http.MethodGet, "/users", nil)
		w = httptest.NewRecorder()

		handler.ListUsers(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var listResp struct {
			Users []*models.User `json:"users"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
			t.Fatalf("failed to unmarshal list response: %v", err)
		}

		found := false
		for _, user := range listResp.Users {
			if user.ID == userID {
				found = true
				break
			}
		}
		if !found {
			t.Error("created user not found in user list")
		}

		// Test 4: Test duplicate username (should fail)
		duplicateReq := models.CreateUserRequest{
			Username: "integrationuser", // same username
			Password: "differentpassword",
		}
		
		body, _ = json.Marshal(duplicateReq)
		req = httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		handler.CreateUser(w, req)

		if w.Code == http.StatusCreated {
			t.Error("expected duplicate username to fail, but it succeeded")
		}
	})
}

func TestUserHandler_ConcurrentCreation(t *testing.T) {
	testutils.SkipIfShort(t)
	testutils.SkipIfNoDatabase(t)

	testDB := testutils.SetupTestDB(t)
	defer testDB.Cleanup(t)

	userRepo := repository.NewUserRepository(testDB.DB)
	userService := service.NewUserService(userRepo)
	handler := NewUserHandler(userService)

	t.Run("concurrent user creation", func(t *testing.T) {
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				createReq := models.CreateUserRequest{
					Username: fmt.Sprintf("concurrentuser%d", id),
					Password: "password123",
				}
				
				body, _ := json.Marshal(createReq)
				req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				handler.CreateUser(w, req)

				if w.Code != http.StatusCreated {
					results <- fmt.Errorf("goroutine %d failed with status %d", id, w.Code)
					return
				}
				results <- nil
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			if err := <-results; err != nil {
				t.Error(err)
			}
		}
	})
}