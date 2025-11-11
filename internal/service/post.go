package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/repository"

	"github.com/google/uuid"
)

type PostService interface {
	CreatePost(ctx context.Context, userID uuid.UUID, req *models.CreatePostRequest) (*models.Post, error)
	GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error)
	ListPosts(ctx context.Context) ([]*models.Post, error)
	ListPostsPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error)
	GetPostsByUser(ctx context.Context, userID uuid.UUID) ([]*models.Post, error)
	GetPostsByUserPaginated(ctx context.Context, userID uuid.UUID, pagination *models.PaginationParams) (*models.PaginatedResponse, error)
	UpdatePost(ctx context.Context, id uuid.UUID, req *models.UpdatePostRequest) (*models.Post, error)
	DeletePost(ctx context.Context, id uuid.UUID) error
}

type postService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) PostService {
	return &postService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

func (s *postService) CreatePost(ctx context.Context, userID uuid.UUID, req *models.CreatePostRequest) (*models.Post, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	post := &models.Post{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (s *postService) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	return s.postRepo.GetByID(ctx, id)
}

func (s *postService) ListPosts(ctx context.Context) ([]*models.Post, error) {
	return s.postRepo.List(ctx)
}

func (s *postService) GetPostsByUser(ctx context.Context, userID uuid.UUID) ([]*models.Post, error) {
	return s.postRepo.GetByUserID(ctx, userID)
}

func (s *postService) UpdatePost(ctx context.Context, id uuid.UUID, req *models.UpdatePostRequest) (*models.Post, error) {
	// Get existing post
	existingPost, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		existingPost.Title = *req.Title
	}
	if req.Content != nil {
		existingPost.Content = *req.Content
	}

	if err := s.postRepo.Update(ctx, id, existingPost); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return existingPost, nil
}

func (s *postService) DeletePost(ctx context.Context, id uuid.UUID) error {
	return s.postRepo.Delete(ctx, id)
}

func (s *postService) ListPostsPaginated(ctx context.Context, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.ListPaginated(ctx, pagination)
	if err != nil {
		return nil, err
	}

	meta := models.NewPaginationMeta(pagination.Page, pagination.PageSize, total)
	
	return &models.PaginatedResponse{
		Data:       posts,
		Pagination: meta,
	}, nil
}

func (s *postService) GetPostsByUserPaginated(ctx context.Context, userID uuid.UUID, pagination *models.PaginationParams) (*models.PaginatedResponse, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	posts, total, err := s.postRepo.GetByUserIDPaginated(ctx, userID, pagination)
	if err != nil {
		return nil, err
	}

	meta := models.NewPaginationMeta(pagination.Page, pagination.PageSize, total)
	
	return &models.PaginatedResponse{
		Data:       posts,
		Pagination: meta,
	}, nil
}