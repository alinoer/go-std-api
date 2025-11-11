# Go REST API Boilerplate

A simple, industry-standard REST API boilerplate built with Go's standard `net/http` library, Chi router, PostgreSQL, and pgx driver.

## Features

- **Clean Architecture**: Handler-Service-Repository pattern
- **Lightweight Routing**: Chi router with standard `net/http`
- **PostgreSQL Integration**: pgx v5 for efficient database operations
- **Database Migrations**: golang-migrate for schema management
- **Middleware**: Logging and authentication middleware
- **Configuration**: Environment-based configuration with godotenv
- **Graceful Shutdown**: Proper server shutdown handling

## Project Structure

```
├── cmd/
│   ├── server/main.go          # Main server application
│   └── seeder/main.go          # Database seeder utility
├── internal/
│   ├── config/config.go        # Configuration management
│   ├── database/connection.go  # Database connection setup
│   ├── handlers/               # HTTP handlers (controller layer)
│   ├── middleware/             # HTTP middleware
│   ├── models/                 # Data models
│   ├── repository/             # Data access layer
│   └── service/                # Business logic layer
├── migrations/                 # Database migration files
├── .env.example               # Environment variables template
└── go.mod                     # Go module definition
```

## Quick Start

### 1. Prerequisites

- Go 1.21+
- PostgreSQL 12+
- golang-migrate CLI (optional, for running migrations)

### 2. Setup

1. Clone and setup the project:
```bash
git clone <repository-url>
cd go-std-api
```

2. Install dependencies:
```bash
go mod tidy
```

3. Setup environment variables:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

4. Create PostgreSQL database:
```bash
createdb go_api_db
```

5. Run migrations:
```bash
# Install golang-migrate (if not already installed)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgres://user:password@localhost:5432/go_api_db?sslmode=disable" up
```

6. Seed the database:
```bash
go run cmd/seeder/main.go
```

7. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check
- `GET /health` - Server health check

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login with username and password

### Users
- `POST /api/v1/users` - Create a new user (same as register)
- `GET /api/v1/users` - List all users (supports pagination)
- `GET /api/v1/users/{id}` - Get user by ID
- `GET /api/v1/users/{userId}/posts` - Get posts by user (supports pagination)

### Posts
- `GET /api/v1/posts` - List all posts (supports pagination)
- `GET /api/v1/posts/{id}` - Get post by ID
- `POST /api/v1/posts` - Create a new post (requires authentication)
- `PUT /api/v1/posts/{id}` - Update post (requires authentication)
- `DELETE /api/v1/posts/{id}` - Delete post (requires authentication)

### Pagination
All list endpoints support pagination via query parameters:
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 10, max: 100)

Example: `GET /api/v1/posts?page=2&page_size=5`

## Authentication

Protected endpoints require a Bearer token in the Authorization header:

```bash
curl -H "Authorization: Bearer MY_SECRET_KEY" \
     -H "Content-Type: application/json" \
     -d '{"title":"New Post","content":"Post content"}' \
     http://localhost:8080/api/v1/posts
```

## Example Usage

### Register a new user:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'
```

### Login:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'
```

### Create a post (authenticated):
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer MY_SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d '{"title":"My First Post","content":"This is the content of my first post."}'
```

### Get all posts:
```bash
curl http://localhost:8080/api/v1/posts
```

### Get paginated posts:
```bash
# Get page 2 with 5 posts per page
curl "http://localhost:8080/api/v1/posts?page=2&page_size=5"
```

### Get posts by specific user (with pagination):
```bash
curl "http://localhost:8080/api/v1/users/{userId}/posts?page=1&page_size=3"
```

## Configuration

Environment variables (see `.env.example`):

- `DATABASE_URL`: PostgreSQL connection string
- `API_SECRET_KEY`: Secret key for API authentication
- `SERVER_PORT`: Server port (default: 8080)

## Development

### Running Tests
```bash
go test ./...
```

### Database Migrations

Create a new migration:
```bash
migrate create -ext sql -dir migrations -seq add_new_table
```

Run migrations:
```bash
migrate -path migrations -database $DATABASE_URL up
```

Rollback migrations:
```bash
migrate -path migrations -database $DATABASE_URL down 1
```

## Architecture

### Handler-Service-Repository Pattern

- **Handlers** (`internal/handlers/`): Handle HTTP requests/responses, input validation
- **Services** (`internal/service/`): Contain business logic, coordinate between handlers and repositories
- **Repositories** (`internal/repository/`): Handle database operations, data access layer

### Dependency Injection

Dependencies are injected through struct initialization:

```go
// Repository layer
userRepo := repository.NewUserRepository(db)
postRepo := repository.NewPostRepository(db)

// Service layer (depends on repositories)
userService := service.NewUserService(userRepo)
postService := service.NewPostService(postRepo, userRepo)

// Handler layer (depends on services)
userHandler := handlers.NewUserHandler(userService)
postHandler := handlers.NewPostHandler(postService, userService)
```

## Security Notes

- The current authentication is a simple API key for demonstration
- In production, implement proper JWT authentication
- The password hashing uses SHA256 for simplicity - use bcrypt in production
- Add rate limiting and input validation as needed
- Use HTTPS in production

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request
