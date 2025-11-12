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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func BenchmarkUserHandler_CreateUser(b *testing.B) {
	mockService := &MockUserHandlerService{
		createdUser: &models.User{
			ID:       uuid.New(),
			Username: "benchuser",
		},
	}
	handler := NewUserHandler(mockService)

	requestBody := models.CreateUserRequest{
		Username: "benchuser",
		Password: "password123",
	}
	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateUser(w, req)
	}
}

func BenchmarkUserHandler_GetUser(b *testing.B) {
	mockService := &MockUserHandlerService{
		retrievedUser: &models.User{
			ID:       uuid.New(),
			Username: "benchuser",
		},
	}
	handler := NewUserHandler(mockService)

	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
		w := httptest.NewRecorder()

		// Add route context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", userID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetUser(w, req)
	}
}

func BenchmarkUserHandler_ListUsers(b *testing.B) {
	// Create test users for the mock
	users := make([]*models.User, 100)
	for i := 0; i < 100; i++ {
		users[i] = &models.User{
			ID:       uuid.New(),
			Username: fmt.Sprintf("user%d", i),
		}
	}

	mockService := &MockUserHandlerService{
		users: users,
	}
	handler := NewUserHandler(mockService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		handler.ListUsers(w, req)
	}
}

func BenchmarkAuthHandler_Register(b *testing.B) {
	mockUserService := &MockAuthUserService{
		createdUser: &models.User{
			ID:       uuid.New(),
			Username: "benchuser",
		},
	}
	authService := service.NewAuthService("test-secret-key")
	handler := NewAuthHandler(mockUserService, authService)

	requestBody := models.RegisterRequest{
		Username: "benchuser",
		Password: "password123",
	}
	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Register(w, req)
	}
}

func BenchmarkAuthHandler_Login(b *testing.B) {
	mockUserService := &MockAuthUserService{
		validateCredentialsUser: &models.User{
			ID:       uuid.New(),
			Username: "benchuser",
		},
	}
	authService := service.NewAuthService("test-secret-key")
	handler := NewAuthHandler(mockUserService, authService)

	requestBody := models.LoginRequest{
		Username: "benchuser",
		Password: "password123",
	}
	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)
	}
}

