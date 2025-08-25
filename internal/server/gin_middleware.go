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

// CORSMiddleware adds CORS headers with enhanced security
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from environment variable
		allowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
		origins := strings.Split(allowedOrigins, ",")
		
		// Clean and validate origins
		var validOrigins []string
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" && isValidOrigin(origin) {
				validOrigins = append(validOrigins, origin)
			} else if origin != "" {
				LogWarn("Invalid CORS origin configured", "origin", origin)
			}
		}

		requestOrigin := c.Request.Header.Get("Origin")
		originAllowed := false

		// Check if the origin is in the allowed list
		for _, allowedOrigin := range validOrigins {
			if allowedOrigin == requestOrigin {
				c.Header("Access-Control-Allow-Origin", requestOrigin)
				originAllowed = true
				break
			}
		}
		
		// Only set CORS headers if origin is allowed
		if originAllowed {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Dev-User-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
		} else if requestOrigin != "" {
			LogWarn("CORS request from unauthorized origin", "origin", requestOrigin)
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isValidOrigin validates that a CORS origin is properly formatted
func isValidOrigin(origin string) bool {
	// Basic validation - must start with http:// or https://
	if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
		return false
	}
	
	// In production, require HTTPS except for localhost
	env := strings.ToLower(getEnv("ENV", "production"))
	if env == "production" {
		if strings.HasPrefix(origin, "http://") && !strings.Contains(origin, "localhost") {
			return false
		}
	}
	
	return true
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
// This middleware provides additional security safeguards to prevent accidental
// enablement in production environments
func DevAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Multi-layer environment validation for security
		env := strings.ToLower(getEnv("ENV", "production"))
		
		// Additional production detection safeguards
		if env != "development" {
			// Check for common production indicators
			if isProductionEnvironment() {
				LogWarn("Dev auth middleware called in production environment - blocking", "env", env)
				c.Next()
				return
			}
			
			// If environment is not explicitly "development" but also not clearly production,
			// be conservative and skip dev auth
			LogWarn("Ambiguous environment detected - skipping dev auth for security", "env", env)
			c.Next()
			return
		}
		
		// Ensure JWT secret is not set when using dev auth
		jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
		if jwtSecret != "" {
			LogWarn("JWT secret configured with dev auth - this is a security concern")
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

// isProductionEnvironment checks for common production environment indicators
func isProductionEnvironment() bool {
	// Check for common production environment variables
	prodIndicators := []string{
		"VERCEL",
		"NETLIFY", 
		"HEROKU",
		"AWS_LAMBDA_FUNCTION_NAME",
		"GOOGLE_CLOUD_PROJECT",
		"CF_PAGES", // Cloudflare Pages
		"RENDER",
	}
	
	for _, indicator := range prodIndicators {
		if getEnv(indicator, "") != "" {
			return true
		}
	}
	
	// Check for production-like domains or URLs
	serverName := getEnv("SERVER_NAME", "")
	if serverName != "" && !strings.Contains(serverName, "localhost") && !strings.Contains(serverName, "127.0.0.1") {
		return true
	}
	
	return false
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Prevent information disclosure
		c.Header("X-Powered-By", "") // Remove default server headers
		
		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy (basic)
		env := strings.ToLower(getEnv("ENV", "production"))
		if env == "production" {
			// Strict CSP for production
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https:; " +
				"style-src 'self' 'unsafe-inline' https:; " +
				"img-src 'self' data: https:; " +
				"font-src 'self' https:; " +
				"connect-src 'self' https:; " +
				"frame-ancestors 'none';"
			c.Header("Content-Security-Policy", csp)
			
			// HSTS for HTTPS
			if c.Request.TLS != nil {
				c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
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
