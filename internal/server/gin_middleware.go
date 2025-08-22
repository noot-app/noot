package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/storage"
)

// RequestIDMiddleware generates and sets request ID for each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware logs request details
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetString("request_id")

		LogDebug("Request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"request_id", requestID,
		)

		c.Next()

		duration := time.Since(start)

		LogInfo("Request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
		)
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := c.GetString("request_id")

				LogError("Panic recovered", fmt.Errorf("panic: %v", rec),
					"request_id", requestID,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
				)

				if isDebugMode() || isDevMode() {
					stack := captureStackGin(4)
					c.JSON(500, gin.H{
						"error":      "Internal server error (panic recovered)",
						"details":    stack,
						"request_id": requestID,
					})
				} else {
					c.JSON(500, gin.H{
						"error": "Internal server error",
					})
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}

// StoreMiddleware injects storage into Gin context
func StoreMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("store", store)
		c.Next()
	}
}

// CORSMiddleware adds CORS headers for development
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment variable, default to localhost:3000 for development
		allowedOrigins := getenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
		origins := strings.Split(allowedOrigins, ",")

		origin := c.Request.Header.Get("Origin")

		// Check if the origin is in the allowed list
		for _, allowedOrigin := range origins {
			if strings.TrimSpace(allowedOrigin) == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-dev-user-id")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// captureStackGin captures stack trace for Gin context
func captureStackGin(skip int) []string {
	var stack []string
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		fnName := "unknown"
		if fn != nil {
			fnName = fn.Name()
		}
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fnName))
	}
	return stack
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// DevAuthMiddleware handles development-only user switching via headers
// TODO: When implementing Supabase auth, this middleware should be replaced with
// a proper auth middleware that validates JWT tokens and extracts user context
func DevAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only allow dev user switching in development environment
		env := strings.ToLower(getEnv("ENV", "production"))
		if env != "development" {
			c.Next()
			return
		}

		// Check for development user override header
		devUserID := c.GetHeader("X-Dev-User-ID")
		if devUserID != "" {
			var user *storage.User
			var err error

			// Resolve user based on dev user ID
			switch devUserID {
			case "monalisa":
				user, err = store.GetUserBySubject(c.Request.Context(), DefaultUserProvider, DefaultUserSubject)
			case "alice":
				user, err = store.GetUserBySubject(c.Request.Context(), AliceUserProvider, AliceUserSubject)
			default:
				LogWarn("Invalid dev user ID requested", "user_id", devUserID)
				c.Next()
				return
			}

			if err != nil {
				LogError("Failed to get dev user", err, "user_id", devUserID)
				c.Next()
				return
			}

			if user != nil {
				// Set the user in context for handlers to use
				// TODO: When implementing Supabase auth, ensure this context key is consistent
				c.Set("dev_user", user)
				LogDebug("Dev user context set", "user_id", devUserID, "email", user.Email)
			}
		}

		c.Next()
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
