package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PostHandler struct {
	postService service.PostService
	userService service.UserService
}

func NewPostHandler(postService service.PostService, userService service.UserService) *PostHandler {
	return &PostHandler{
		postService: postService,
		userService: userService,
	}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// For this example, we'll use a hardcoded user ID since we don't have proper JWT auth
	// In a real application, you would extract the user ID from the JWT token
	// For now, let's get the first user from the database as an example
	users, err := h.userService.ListUsers(r.Context())
	if err != nil || len(users) == 0 {
		WriteError(w, http.StatusBadRequest, "No users found. Please create a user first")
		return
	}
	userID := users[0].ID

	post, err := h.postService.CreatePost(r.Context(), userID, &req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, SuccessResponse{Data: post})
}

func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	post, err := h.postService.GetPost(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Post not found")
		return
	}

	WriteSuccess(w, post)
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	// Check if pagination parameters are provided
	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("page_size") != "" {
		h.ListPostsPaginated(w, r)
		return
	}

	// Default non-paginated response
	posts, err := h.postService.ListPosts(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	WriteSuccess(w, posts)
}

func (h *PostHandler) ListPostsPaginated(w http.ResponseWriter, r *http.Request) {
	pagination := ParsePaginationParams(r)

	result, err := h.postService.ListPostsPaginated(r.Context(), pagination)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *PostHandler) GetPostsByUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check if pagination parameters are provided
	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("page_size") != "" {
		h.GetPostsByUserPaginated(w, r, userID)
		return
	}

	// Default non-paginated response
	posts, err := h.postService.GetPostsByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	WriteSuccess(w, posts)
}

func (h *PostHandler) GetPostsByUserPaginated(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	pagination := ParsePaginationParams(r)

	result, err := h.postService.GetPostsByUserPaginated(r.Context(), userID, pagination)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	var req models.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	post, err := h.postService.UpdatePost(r.Context(), id, &req)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	WriteSuccess(w, post)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	if err := h.postService.DeletePost(r.Context(), id); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	WriteMessage(w, "Post deleted successfully")
}