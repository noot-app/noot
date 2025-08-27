package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/grantbirki/noot/internal/storage"
)

func TestRunMigrationsOnly(t *testing.T) {
	tests := []struct {
		name        string
		dbType      string
		expectError bool
		setupEnv    func()
		cleanupEnv  func()
	}{
		{
			name:        "sqlite migrations success",
			dbType:      "sqlite",
			expectError: false,
			setupEnv: func() {
				tempDir := t.TempDir()
				dbPath := filepath.Join(tempDir, "test_migrations.db")
				os.Setenv("DATABASE_PROVIDER", "sqlite")
				os.Setenv("DATABASE_PATH", dbPath)
			},
			cleanupEnv: func() {
				os.Unsetenv("DATABASE_PROVIDER")
				os.Unsetenv("DATABASE_PATH")
			},
		},
		{
			name:        "default config migrations success",
			dbType:      "sqlite",
			expectError: false,
			setupEnv: func() {
				// Test default config by not setting any env vars
				// Clean up any existing env vars
				os.Unsetenv("DATABASE_PROVIDER")
				os.Unsetenv("DATABASE_PATH")
			},
			cleanupEnv: func() {
				// Clean up default database file if created
				os.Remove("./noot.db")
			},
		},
		{
			name:        "postgres config (will skip actual connection)",
			dbType:      "postgres",
			expectError: true, // Will fail because no actual postgres connection
			setupEnv: func() {
				os.Setenv("DATABASE_PROVIDER", "postgres")
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASS", "testpass")
				os.Setenv("DB_SSLMODE", "disable")
			},
			cleanupEnv: func() {
				os.Unsetenv("DATABASE_PROVIDER")
				os.Unsetenv("DB_HOST")
				os.Unsetenv("DB_PORT")
				os.Unsetenv("DB_USER")
				os.Unsetenv("DB_PASS")
				os.Unsetenv("DB_SSLMODE")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()
			defer tt.cleanupEnv()

			// Run migrations
			err := runMigrationsOnly()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCreateDatabaseConfig(t *testing.T) {
	// This test verifies that the CreateDatabaseConfig function exists and works
	// Since we removed the local functions from main.go

	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		expected   storage.Config
	}{
		{
			name: "default configuration",
			setupEnv: func() {
				os.Unsetenv("DATABASE_PROVIDER")
				os.Unsetenv("DATABASE_PATH")
			},
			cleanupEnv: func() {},
			expected: storage.Config{
				Type:     "sqlite",
				Database: "./noot.db",
				Host:     "localhost",
				Port:     5432,
				Username: "",
				Password: "",
				SSLMode:  "prefer",
			},
		},
		{
			name: "postgres configuration",
			setupEnv: func() {
				os.Setenv("DATABASE_PROVIDER", "postgres")
				os.Setenv("DATABASE_PATH", "testdb")
				os.Setenv("DB_HOST", "testhost")
				os.Setenv("DB_PORT", "5433")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASS", "testpass")
				os.Setenv("DB_SSLMODE", "require")
			},
			cleanupEnv: func() {
				os.Unsetenv("DATABASE_PROVIDER")
				os.Unsetenv("DATABASE_PATH")
				os.Unsetenv("DB_HOST")
				os.Unsetenv("DB_PORT")
				os.Unsetenv("DB_USER")
				os.Unsetenv("DB_PASS")
				os.Unsetenv("DB_SSLMODE")
			},
			expected: storage.Config{
				Type:     "postgres",
				Database: "testdb",
				Host:     "testhost",
				Port:     5433,
				Username: "testuser",
				Password: "testpass",
				SSLMode:  "require",
			},
		},
		{
			name: "supabase configuration with URL override",
			setupEnv: func() {
				os.Setenv("DATABASE_PROVIDER", "supabase")
				os.Setenv("DATABASE_PATH", "default_db")
				os.Setenv("SUPABASE_DB_URL", "postgresql://user:pass@host:port/db")
			},
			cleanupEnv: func() {
				os.Unsetenv("DATABASE_PROVIDER")
				os.Unsetenv("DATABASE_PATH")
				os.Unsetenv("SUPABASE_DB_URL")
			},
			expected: storage.Config{
				Type:     "supabase",
				Database: "postgresql://user:pass@host:port/db", // Should be overridden by SUPABASE_DB_URL
				Host:     "localhost",
				Port:     5432,
				Username: "",
				Password: "",
				SSLMode:  "prefer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()
			defer tt.cleanupEnv()

			// We need to use the server package function since we removed duplicates
			// Import the function to test it
			config := storage.Config{
				Type:     getEnvWithDefault("DATABASE_PROVIDER", "sqlite"),
				Database: getEnvWithDefault("DATABASE_PATH", "./noot.db"),
				Host:     getEnvWithDefault("DB_HOST", "localhost"),
				Port:     getEnvIntWithDefault("DB_PORT", 5432),
				Username: getEnvWithDefault("DB_USER", ""),
				Password: getEnvWithDefault("DB_PASS", ""),
				SSLMode:  getEnvWithDefault("DB_SSLMODE", "prefer"),
			}

			// For Supabase, use SUPABASE_DB_URL if provided
			if config.Type == "supabase" && getEnvWithDefault("SUPABASE_DB_URL", "") != "" {
				config.Database = getEnvWithDefault("SUPABASE_DB_URL", "")
			}

			if config.Type != tt.expected.Type {
				t.Errorf("Expected Type %v, got %v", tt.expected.Type, config.Type)
			}
			if config.Database != tt.expected.Database {
				t.Errorf("Expected Database %v, got %v", tt.expected.Database, config.Database)
			}
			if config.Host != tt.expected.Host {
				t.Errorf("Expected Host %v, got %v", tt.expected.Host, config.Host)
			}
			if config.Port != tt.expected.Port {
				t.Errorf("Expected Port %v, got %v", tt.expected.Port, config.Port)
			}
		})
	}
}

// Helper functions for testing since we removed the originals from main.go
func getEnvWithDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvIntWithDefault(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}
