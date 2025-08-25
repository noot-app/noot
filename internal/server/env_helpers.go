package server

import (
	"os"
	"strconv"
	"strings"
)

// getenv gets environment variable with fallback (consistent naming with server.go)
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getenvInt gets environment variable as int with fallback
func getenvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

// getEnv is an alias for getenv for consistency with gin_middleware.go
func getEnv(key, fallback string) string {
	return getenv(key, fallback)
}

// IsProduction checks if the application is running in production mode.
// This is the single source of truth for production detection.
// If ENV is set to "production" (case-insensitive, trimmed), returns true.
// If ENV is unset, defaults to production for safety.
// Any other ENV value (like "development", "staging", etc.) returns false.
func IsProduction() bool {
	env := strings.ToLower(strings.TrimSpace(getenv("ENV", "production")))
	return env == "production"
}
