package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectedError bool
		validateFunc  func(*Config) error
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"DATABASE_URL":   "",
				"API_SECRET_KEY": "",
				"SERVER_PORT":    "",
			},
			expectedError: false,
			validateFunc: func(c *Config) error {
				if c.DatabaseURL == "" {
					return fmt.Errorf("expected default DATABASE_URL")
				}
				if c.APISecretKey == "" {
					return fmt.Errorf("expected default API_SECRET_KEY")
				}
				if c.ServerPort == "" {
					return fmt.Errorf("expected default SERVER_PORT")
				}
				return nil
			},
		},
		{
			name: "custom values from environment",
			envVars: map[string]string{
				"DATABASE_URL":   "postgres://custom:password@localhost:5432/custom_db",
				"API_SECRET_KEY": "custom_secret_key",
				"SERVER_PORT":    "9090",
			},
			expectedError: false,
			validateFunc: func(c *Config) error {
				if c.DatabaseURL != "postgres://custom:password@localhost:5432/custom_db" {
					return fmt.Errorf("expected custom DATABASE_URL, got %s", c.DatabaseURL)
				}
				if c.APISecretKey != "custom_secret_key" {
					return fmt.Errorf("expected custom API_SECRET_KEY, got %s", c.APISecretKey)
				}
				if c.ServerPort != "9090" {
					return fmt.Errorf("expected custom SERVER_PORT, got %s", c.ServerPort)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup original env vars
			originalEnvs := make(map[string]string)
			for key := range tt.envVars {
				originalEnvs[key] = os.Getenv(key)
			}

			// Set test env vars
			for key, value := range tt.envVars {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			// Restore original env vars after test
			defer func() {
				for key, originalValue := range originalEnvs {
					if originalValue == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, originalValue)
					}
				}
			}()

			config, err := Load()

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Error("expected non-nil config")
				return
			}

			if tt.validateFunc != nil {
				if err := tt.validateFunc(config); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestConfig_validate(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedError bool
		errorContains string
	}{
		{
			name: "valid config",
			config: &Config{
				DatabaseURL:  "postgres://localhost:5432/test",
				APISecretKey: "secret",
				ServerPort:   "8080",
			},
			expectedError: false,
		},
		{
			name: "empty database URL",
			config: &Config{
				DatabaseURL:  "",
				APISecretKey: "secret",
				ServerPort:   "8080",
			},
			expectedError: true,
			errorContains: "DATABASE_URL is required",
		},
		{
			name: "empty API secret key",
			config: &Config{
				DatabaseURL:  "postgres://localhost:5432/test",
				APISecretKey: "",
				ServerPort:   "8080",
			},
			expectedError: true,
			errorContains: "API_SECRET_KEY is required",
		},
		{
			name: "empty server port",
			config: &Config{
				DatabaseURL:  "postgres://localhost:5432/test",
				APISecretKey: "secret",
				ServerPort:   "",
			},
			expectedError: true,
			errorContains: "SERVER_PORT is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env var set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "set_value",
			expected:     "set_value",
		},
		{
			name:         "env var not set",
			key:          "UNSET_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "empty env var",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup original value
			original := os.Getenv(tt.key)
			defer func() {
				if original == "" {
					os.Unsetenv(tt.key)
				} else {
					os.Setenv(tt.key, original)
				}
			}()

			// Set test value
			if tt.envValue == "" {
				os.Unsetenv(tt.key)
			} else {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func BenchmarkLoad(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load()
	}
}

func BenchmarkGetEnv(b *testing.B) {
	os.Setenv("BENCH_VAR", "bench_value")
	defer os.Unsetenv("BENCH_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnv("BENCH_VAR", "default")
	}
}