package testutils

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB represents a test database instance
type TestDB struct {
	DB       *pgxpool.Pool
	ConnStr  string
	teardown func()
}

// SetupTestDB creates and returns a test database instance
func SetupTestDB(t TestingInterface) *TestDB {
	// Check if we should skip database tests
	if testing.Short() {
		t.Skip("Skipping database tests in short mode")
	}

	// Get test database connection string from environment
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		// Default test database connection string
		connStr = "postgres://localhost/go_api_test?sslmode=disable"
	}

	db, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		t.Skipf("Cannot ping test database: %v", err)
	}

	// Create a unique schema for this test
	schemaName := fmt.Sprintf("test_schema_%s", sanitizeTestName(t.Name()))
	_, err = db.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Set search path to use our test schema
	_, err = db.Exec(context.Background(), fmt.Sprintf("SET search_path TO %s", schemaName))
	if err != nil {
		t.Fatalf("Failed to set search path: %v", err)
	}

	// Create tables
	createTables(t, db)

	return &TestDB{
		DB:      db,
		ConnStr: connStr,
		teardown: func() {
			// Clean up the test schema
			db.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
			db.Close()
		},
	}
}

// Cleanup cleans up the test database
func (tdb *TestDB) Cleanup(t TestingInterface) {
	if tdb.teardown != nil {
		tdb.teardown()
	}
}

// TestingInterface is an interface that both *testing.T and *testing.B implement
type TestingInterface interface {
	Skip(...interface{})
	Skipf(string, ...interface{})
	Fatalf(string, ...interface{})
	Name() string
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t TestingInterface) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// SkipIfNoDatabase skips the test if no test database is available
func SkipIfNoDatabase(t *testing.T) {
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://localhost/go_api_test?sslmode=disable"
	}

	db, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Skipf("No test database available: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
}

// sanitizeTestName removes special characters from test names for schema names
func sanitizeTestName(name string) string {
	// Replace special characters with underscore
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else {
			result += "_"
		}
	}
	// Limit length
	if len(result) > 20 {
		result = result[:20]
	}
	return result
}

// createTables creates the necessary tables for testing
func createTables(t TestingInterface, db *pgxpool.Pool) {
	// Create users table
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create posts table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS posts (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}
}