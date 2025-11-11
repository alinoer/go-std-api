package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	user, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, SuccessResponse{Data: user})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	WriteSuccess(w, user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Check if pagination parameters are provided
	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("page_size") != "" {
		h.ListUsersPaginated(w, r)
		return
	}

	// Default non-paginated response
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	WriteSuccess(w, users)
}

func (h *UserHandler) ListUsersPaginated(w http.ResponseWriter, r *http.Request) {
	pagination := ParsePaginationParams(r)

	result, err := h.userService.ListUsersPaginated(r.Context(), pagination)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	WriteJSON(w, http.StatusOK, result)
}