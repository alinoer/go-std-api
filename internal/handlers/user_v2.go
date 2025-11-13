package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/alinoer/go-std-api/internal/errors"
	"github.com/alinoer/go-std-api/internal/logger"
	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/response"
	"github.com/alinoer/go-std-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// UserHandlerV2 demonstrates the new error handling system
type UserHandlerV2 struct {
	userService service.UserService
	logger      *logger.Logger
}

// NewUserHandlerV2 creates a new user handler with enhanced error handling
func NewUserHandlerV2(userService service.UserService) *UserHandlerV2 {
	return &UserHandlerV2{
		userService: userService,
		logger:      logger.GetLogger(),
	}
}

// CreateUser creates a new user with comprehensive error handling
func (h *UserHandlerV2) CreateUser(w http.ResponseWriter, r *http.Request) {
	resp := response.NewResponseWriter(w, r)
	ctx := r.Context()
	
	// Parse request
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.BadRequest("Invalid JSON payload")
		return
	}

	// Validate request
	if validationErr := h.validateCreateUserRequest(&req); validationErr != nil {
		resp.ValidationError(validationErr)
		return
	}

	// Create user
	user, err := h.userService.CreateUser(ctx, &req)
	if err != nil {
		// Convert service errors to appropriate HTTP errors
		if appErr := errors.AsAppError(err); appErr != nil {
			resp.Error(appErr)
		} else {
			resp.Error(errors.InternalError("Failed to create user").WithInternal(err))
		}
		return
	}

	// Log successful creation
	h.logger.WithContext(ctx).Info("User created successfully",
		"user_id", user.ID,
		"username", user.Username,
	)

	// Return success response
	resp.Created(map[string]interface{}{
		"user": user,
	})
}

// GetUser retrieves a user by ID
func (h *UserHandlerV2) GetUser(w http.ResponseWriter, r *http.Request) {
	resp := response.NewResponseWriter(w, r)
	ctx := r.Context()

	// Parse and validate user ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		resp.BadRequest("Invalid user ID format")
		return
	}

	// Get user from service
	user, err := h.userService.GetUser(ctx, id)
	if err != nil {
		// Handle different types of errors
		if appErr := errors.AsAppError(err); appErr != nil {
			switch appErr.Code {
			case errors.ErrCodeNotFound:
				resp.NotFound("user")
			default:
				resp.Error(appErr)
			}
		} else {
			// Convert unknown errors
			resp.Error(errors.InternalError("Failed to retrieve user").WithInternal(err))
		}
		return
	}

	// Return user data
	resp.Success(map[string]interface{}{
		"user": user,
	})
}

// ListUsers lists all users with pagination support
func (h *UserHandlerV2) ListUsers(w http.ResponseWriter, r *http.Request) {
	resp := response.NewResponseWriter(w, r)
	ctx := r.Context()

	// Parse pagination parameters
	pagination := h.parsePaginationParams(r)

	// Check if pagination is requested
	if pagination != nil {
		// Handle paginated response
		result, err := h.userService.ListUsersPaginated(ctx, pagination)
		if err != nil {
			resp.Error(errors.DatabaseError("list users", err))
			return
		}

		resp.JSONWithMeta(http.StatusOK, result.Data, "Users retrieved successfully", result.Pagination)
	} else {
		// Handle simple list
		users, err := h.userService.ListUsers(ctx)
		if err != nil {
			resp.Error(errors.DatabaseError("list users", err))
			return
		}

		resp.Success(map[string]interface{}{
			"users": users,
			"count": len(users),
		})
	}
}

// UpdateUser updates user information
func (h *UserHandlerV2) UpdateUser(w http.ResponseWriter, r *http.Request) {
	resp := response.NewResponseWriter(w, r)

	// Parse user ID
	idStr := chi.URLParam(r, "id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		resp.BadRequest("Invalid user ID format")
		return
	}

	// Parse request body
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.BadRequest("Invalid JSON payload")
		return
	}

	// Validate request
	if validationErr := h.validateUpdateUserRequest(&req); validationErr != nil {
		resp.ValidationError(validationErr)
		return
	}

	// Update user (this would require implementing UpdateUser in service)
	// For now, we'll return a not implemented error
	resp.Error(errors.NewAppError(errors.ErrCodeInternal, "Update functionality not yet implemented").
		WithHTTPStatus(http.StatusNotImplemented).
		WithContext("feature", "user_update"))
}

// DeleteUser deletes a user
func (h *UserHandlerV2) DeleteUser(w http.ResponseWriter, r *http.Request) {
	resp := response.NewResponseWriter(w, r)
	
	// Parse user ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		resp.BadRequest("Invalid user ID format")
		return
	}

	// For now, return not implemented
	resp.Error(errors.NewAppError(errors.ErrCodeInternal, "Delete functionality not yet implemented").
		WithHTTPStatus(http.StatusNotImplemented).
		WithContext("user_id", id.String()))
}

// validateCreateUserRequest validates the create user request
func (h *UserHandlerV2) validateCreateUserRequest(req *models.CreateUserRequest) *errors.ValidationErrors {
	validationErrors := &errors.ValidationErrors{}

	// Username validation
	if req.Username == "" {
		validationErrors.Add("username", "Username is required")
	} else if len(req.Username) < 3 {
		validationErrors.Add("username", "Username must be at least 3 characters long")
	} else if len(req.Username) > 50 {
		validationErrors.Add("username", "Username must be less than 50 characters")
	}

	// Password validation
	if req.Password == "" {
		validationErrors.Add("password", "Password is required")
	} else if len(req.Password) < 6 {
		validationErrors.Add("password", "Password must be at least 6 characters long")
	} else if len(req.Password) > 100 {
		validationErrors.Add("password", "Password must be less than 100 characters")
	}

	// Security validations
	if h.containsSQLInjection(req.Username) {
		validationErrors.Add("username", "Username contains invalid characters")
	}

	if h.containsXSS(req.Username) {
		validationErrors.Add("username", "Username contains potentially dangerous content")
	}

	if !validationErrors.HasErrors() {
		return nil
	}

	return validationErrors
}

// validateUpdateUserRequest validates the update user request
func (h *UserHandlerV2) validateUpdateUserRequest(req *models.UpdateUserRequest) *errors.ValidationErrors {
	validationErrors := &errors.ValidationErrors{}

	// Add validation logic for update request
	// This is a placeholder for when UpdateUserRequest is defined

	if !validationErrors.HasErrors() {
		return nil
	}

	return validationErrors
}

// parsePaginationParams extracts pagination parameters from request
func (h *UserHandlerV2) parsePaginationParams(r *http.Request) *models.PaginationParams {
	// Check if pagination parameters are present
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	if pageStr == "" && pageSizeStr == "" {
		return nil // No pagination requested
	}

	// Parse parameters with defaults
	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := parsePositiveInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := parsePositiveInt(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return models.NewPaginationParams(page, pageSize)
}

// containsSQLInjection checks for basic SQL injection patterns
func (h *UserHandlerV2) containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_", 
		"DROP", "DELETE", "INSERT", "UPDATE", "SELECT", "UNION",
	}
	
	for _, pattern := range sqlPatterns {
		if contains(input, pattern) {
			return true
		}
	}
	return false
}

// containsXSS checks for basic XSS patterns
func (h *UserHandlerV2) containsXSS(input string) bool {
	xssPatterns := []string{
		"<script", "</script>", "<iframe", "javascript:", "onload=", "onerror=",
		"<img", "src=", "href=", "onclick=", "onmouseover=",
	}
	
	for _, pattern := range xssPatterns {
		if contains(input, pattern) {
			return true
		}
	}
	return false
}

// Helper functions

func parsePositiveInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, errors.NewAppError(errors.ErrCodeValidation, "Value must be positive")
	}
	return val, nil
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}