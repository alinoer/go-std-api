package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/repository"

	"github.com/google/uuid"
)

type UserService interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	ListUsers(ctx context.Context) ([]*models.User, error)
	ListUsersPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error)
	ValidateCredentials(ctx context.Context, username, password string) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Check if username already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash the password (simple SHA256 for demo - use bcrypt in production)
	hasher := sha256.New()
	hasher.Write([]byte(req.Password))
	passwordHash := fmt.Sprintf("%x", hasher.Sum(nil))

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *userService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return s.userRepo.List(ctx)
}

func (s *userService) ValidateCredentials(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Hash the provided password and compare
	hasher := sha256.New()
	hasher.Write([]byte(password))
	passwordHash := fmt.Sprintf("%x", hasher.Sum(nil))

	if user.PasswordHash != passwordHash {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

func (s *userService) ListUsersPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	users, total, err := s.userRepo.ListPaginated(ctx, pagination)
	if err != nil {
		return nil, err
	}

	meta := models.NewPaginationMeta(pagination.Page, pagination.PageSize, total)
	
	return &models.PaginatedResponse{
		Data:       users,
		Pagination: meta,
	}, nil
}