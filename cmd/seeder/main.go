package main

import (
	"context"
	"log"

	"github.com/alinoer/go-std-api/internal/config"
	"github.com/alinoer/go-std-api/internal/database"
	"github.com/alinoer/go-std-api/internal/models"
	"github.com/alinoer/go-std-api/internal/repository"
	"github.com/alinoer/go-std-api/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Connected to database successfully")

	// Initialize repositories and services
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userRepo)

	ctx := context.Background()

	// Create sample users
	log.Println("Creating sample users...")
	
	sampleUsers := []models.CreateUserRequest{
		{Username: "john_doe", Password: "password123"},
		{Username: "alice_smith", Password: "alice2023"},
		{Username: "bob_wilson", Password: "bob456"},
		{Username: "sarah_jones", Password: "sarah789"},
		{Username: "mike_brown", Password: "mike321"},
	}

	var users []*models.User
	
	for _, userReq := range sampleUsers {
		user, err := userService.CreateUser(ctx, &userReq)
		if err != nil {
			log.Printf("Failed to create user %s (may already exist): %v", userReq.Username, err)
			// Try to get existing user
			user, err = userService.GetUserByUsername(ctx, userReq.Username)
			if err != nil {
				log.Printf("Failed to get existing user %s: %v", userReq.Username, err)
				continue
			}
			log.Printf("Using existing user: %s", user.Username)
		} else {
			log.Printf("Created user: %s (ID: %s)", user.Username, user.ID)
		}
		users = append(users, user)
	}

	// Create sample posts
	log.Println("Creating sample posts...")
	
	if len(users) == 0 {
		log.Println("No users available, skipping post creation")
	} else {
		samplePosts := []struct {
			Title   string
			Content string
			UserIdx int // Index of user in users array
		}{
			{
				Title:   "Welcome to Go API",
				Content: "This is a sample post created by the seeder. It demonstrates how to create posts using the Go REST API with PostgreSQL backend.",
				UserIdx: 0,
			},
			{
				Title:   "Building REST APIs with Go",
				Content: "Go (Golang) is an excellent choice for building REST APIs. With its powerful standard library, particularly the net/http package, you can create robust and performant web services.",
				UserIdx: 1,
			},
			{
				Title:   "Database Integration with pgx",
				Content: "The pgx library provides excellent PostgreSQL integration for Go applications. It offers connection pooling, prepared statements, and efficient data scanning.",
				UserIdx: 0,
			},
			{
				Title:   "Advanced Go Patterns",
				Content: "Exploring advanced Go patterns like interfaces, context, channels, and goroutines for building concurrent applications.",
				UserIdx: 2,
			},
			{
				Title:   "Testing in Go",
				Content: "Best practices for testing Go applications, including unit tests, integration tests, and table-driven tests.",
				UserIdx: 3,
			},
			{
				Title:   "Go Performance Tips",
				Content: "Optimizing Go applications for better performance: profiling, benchmarking, and memory management.",
				UserIdx: 4,
			},
			{
				Title:   "Microservices with Go",
				Content: "Building scalable microservices architecture using Go, including service discovery and load balancing.",
				UserIdx: 1,
			},
			{
				Title:   "Go and Docker",
				Content: "Containerizing Go applications with Docker: multi-stage builds, optimization, and deployment strategies.",
				UserIdx: 2,
			},
			{
				Title:   "Error Handling in Go",
				Content: "Effective error handling patterns in Go: custom errors, error wrapping, and error propagation.",
				UserIdx: 0,
			},
			{
				Title:   "Concurrency Patterns",
				Content: "Deep dive into Go's concurrency primitives: goroutines, channels, select statements, and sync package.",
				UserIdx: 3,
			},
		}

		for _, postData := range samplePosts {
			if postData.UserIdx >= len(users) {
				continue
			}
			
			postReq := &models.CreatePostRequest{
				Title:   postData.Title,
				Content: postData.Content,
			}
			
			post, err := postService.CreatePost(ctx, users[postData.UserIdx].ID, postReq)
			if err != nil {
				log.Printf("Failed to create post '%s': %v", postReq.Title, err)
				continue
			}
			log.Printf("Created post: %s by %s (ID: %s)", post.Title, users[postData.UserIdx].Username, post.ID)
		}
	}

	log.Println("Seeding completed successfully!")
	log.Printf("You can now:")
	log.Printf("1. Start the server: go run cmd/server/main.go")
	log.Printf("2. View posts: GET http://localhost:8080/api/v1/posts")
	log.Printf("3. Create protected posts: POST http://localhost:8080/api/v1/posts (with Authorization: Bearer %s)", cfg.APISecretKey)
}