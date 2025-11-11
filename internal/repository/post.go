package repository

import (
	"context"
	"fmt"

	"github.com/alinoer/go-std-api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error)
	List(ctx context.Context) ([]*models.Post, error)
	ListPaginated(ctx context.Context, pagination *models.PaginationParams) ([]*models.Post, int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Post, error)
	GetByUserIDPaginated(ctx context.Context, userID uuid.UUID, pagination *models.PaginationParams) ([]*models.Post, int64, error)
	Update(ctx context.Context, id uuid.UUID, post *models.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	query := `
		INSERT INTO posts (id, user_id, title, content, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query, post.ID, post.UserID, post.Title, post.Content, post.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	return nil
}

func (r *postRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at
		FROM posts
		WHERE id = $1`

	var post models.Post
	err := r.db.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &post, nil
}

func (r *postRepository) List(ctx context.Context) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at
		FROM posts
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, nil
}

func (r *postRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by user ID: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, nil
}

func (r *postRepository) Update(ctx context.Context, id uuid.UUID, post *models.Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2
		WHERE id = $3`

	result, err := r.db.Exec(ctx, query, post.Title, post.Content, id)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *postRepository) ListPaginated(ctx context.Context, pagination *models.PaginationParams) ([]*models.Post, int64, error) {
	// First, get the total count
	countQuery := `SELECT COUNT(*) FROM posts`
	var total int64
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// Then get the paginated results
	query := `
		SELECT id, user_id, title, content, created_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, pagination.PageSize, pagination.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list posts with pagination: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, total, nil
}

func (r *postRepository) GetByUserIDPaginated(ctx context.Context, userID uuid.UUID, pagination *models.PaginationParams) ([]*models.Post, int64, error) {
	// First, get the total count for this user
	countQuery := `SELECT COUNT(*) FROM posts WHERE user_id = $1`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts for user: %w", err)
	}

	// Then get the paginated results
	query := `
		SELECT id, user_id, title, content, created_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, pagination.PageSize, pagination.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts by user ID with pagination: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, total, nil
}