package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/service"
)

type AuthHandler struct {
	userService service.UserService
	authService *service.AuthService
}

func NewAuthHandler(userService service.UserService, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.Username == "" {
		WriteError(w, http.StatusBadRequest, "Username is required")
		return
	}
	if req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Password is required")
		return
	}
	if len(req.Password) < 6 {
		WriteError(w, http.StatusBadRequest, "Password must be at least 6 characters long")
		return
	}

	// Convert to CreateUserRequest
	createReq := &models.CreateUserRequest{
		Username: req.Username,
		Password: req.Password,
	}

	user, err := h.userService.CreateUser(r.Context(), createReq)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := models.RegisterResponse{
		User:    user,
		Message: "User registered successfully",
	}

	WriteJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate input
	if req.Username == "" {
		WriteError(w, http.StatusBadRequest, "Username is required")
		return
	}
	if req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Validate credentials
	user, err := h.userService.ValidateCredentials(r.Context(), req.Username, req.Password)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate JWT token
	token, expiresIn, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response := models.LoginResponse{
		User:        user,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}

	WriteJSON(w, http.StatusOK, response)
}