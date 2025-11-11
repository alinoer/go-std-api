package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alinoer/go-std-api/internal/config"
	"github.com/alinoer/go-std-api/internal/database"
	"github.com/alinoer/go-std-api/internal/handlers"
	"github.com/alinoer/go-std-api/internal/middleware"
	"github.com/alinoer/go-std-api/internal/repository"
	"github.com/alinoer/go-std-api/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
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

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo, userRepo)
	authService := service.NewAuthService(cfg.APISecretKey)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	postHandler := handlers.NewPostHandler(postService, userService)
	authHandler := handlers.NewAuthHandler(userService, authService)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.LoggingMiddleware)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteMessage(w, "Server is healthy")
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Authentication routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Public routes
		r.Post("/users", userHandler.CreateUser) // Duplicate of register for backwards compatibility
		r.Get("/users", userHandler.ListUsers)
		r.Get("/users/{id}", userHandler.GetUser)
		r.Get("/users/{userId}/posts", postHandler.GetPostsByUser)

		// Public posts routes (read-only)
		r.Get("/posts", postHandler.ListPosts)
		r.Get("/posts/{id}", postHandler.GetPost)

		// Protected routes (require JWT authentication)
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuthMiddleware(authService))

			// Protected post routes
			r.Post("/posts", postHandler.CreatePost)
			r.Put("/posts/{id}", postHandler.UpdatePost)
			r.Delete("/posts/{id}", postHandler.DeletePost)
		})
	})

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
