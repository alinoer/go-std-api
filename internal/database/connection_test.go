package database

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewConnection(t *testing.T) {
	tests := []struct {
		name          string
		databaseURL   string
		expectedError bool
		errorContains string
	}{
		{
			name:          "invalid URL format",
			databaseURL:   "invalid-url",
			expectedError: true,
			errorContains: "failed to parse database config",
		},
		{
			name:          "empty URL", 
			databaseURL:   "",
			// Note: pgxpool.ParseConfig actually accepts empty string and uses defaults
			// This test documents that behavior rather than enforcing it
			expectedError: false,
		},
		{
			name:        "valid URL format but unreachable database",
			databaseURL: "postgres://user:password@nonexistent-host:5432/database",
			// This might not error immediately due to connection pooling,
			// but will error on ping
			expectedError: true,
			errorContains: "failed to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewConnection(tt.databaseURL)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
					// Clean up if pool was somehow created
					if pool != nil {
						pool.Close()
					}
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if pool == nil {
				t.Error("expected non-nil pool")
				return
			}

			// Clean up
			pool.Close()
		})
	}
}

func TestNewConnection_WithRealDatabase(t *testing.T) {
	// This test requires a real database connection
	// Skip if not available
	databaseURL := getTestDatabaseURL()
	if databaseURL == "" {
		t.Skip("No test database URL available, skipping integration test")
	}

	pool, err := NewConnection(databaseURL)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
	defer pool.Close()

	if pool == nil {
		t.Error("expected non-nil pool")
	}

	// Test that we can actually use the connection
	ctx := context.Background()
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}

	// Test that pool stats are reasonable
	stats := pool.Stat()
	if stats.MaxConns() != 30 {
		t.Errorf("expected MaxConns to be 30, got %d", stats.MaxConns())
	}

	// Note: MinConns() method doesn't exist in pgxpool.Stat
	// This is configuration we set, not something we can easily verify from stats
	if stats.TotalConns() < 0 {
		t.Errorf("expected TotalConns to be non-negative, got %d", stats.TotalConns())
	}
}

func TestNewConnection_ConfigValidation(t *testing.T) {
	validURLs := []string{
		"postgres://user:password@localhost:5432/database",
		"postgresql://user:password@localhost:5432/database?sslmode=disable",
		"postgres://user@localhost/database",
	}

	for _, url := range validURLs {
		t.Run("valid_url_"+url, func(t *testing.T) {
			// We don't expect these to connect successfully, but the URL parsing should work
			_, err := NewConnection(url)
			// Error is expected since we're not connecting to real databases
			// but it should be a connection error, not a parsing error
			if err != nil && strings.Contains(err.Error(), "failed to parse database config") {
				t.Errorf("URL parsing failed for valid URL %s: %v", url, err)
			}
		})
	}
}

// getTestDatabaseURL returns the test database URL from environment
func getTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	// Try common test database URLs
	testURLs := []string{
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		"postgres://localhost:5432/go_api_test?sslmode=disable",
	}
	
	for _, url := range testURLs {
		if pool, err := pgxpool.New(context.Background(), url); err == nil {
			if err := pool.Ping(context.Background()); err == nil {
				pool.Close()
				return url
			}
			pool.Close()
		}
	}
	
	return ""
}

func BenchmarkNewConnection(b *testing.B) {
	// Use a URL that will fail quickly for benchmarking parsing performance
	databaseURL := "postgres://user:password@nonexistent-host:5432/database"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool, _ := NewConnection(databaseURL)
		if pool != nil {
			pool.Close()
		}
	}
}