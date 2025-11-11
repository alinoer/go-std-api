package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	APISecretKey string
	ServerPort   string
}

func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	config := &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/go_api_db?sslmode=disable"),
		APISecretKey: getEnv("API_SECRET_KEY", "MY_SECRET_KEY"),
		ServerPort:   getEnv("SERVER_PORT", "8080"),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.APISecretKey == "" {
		return fmt.Errorf("API_SECRET_KEY is required")
	}
	if c.ServerPort == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}