package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
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
		allowedOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
		
		// Handle empty CORS configuration
		if allowedOrigins == "" {
			if IsProduction() {
				// In production, crash if CORS_ALLOWED_ORIGINS is not set
				LogError("CORS_ALLOWED_ORIGINS must be set in production environment", fmt.Errorf("missing CORS configuration"))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
				c.Abort()
				return
			}
			// In development, default to localhost
			allowedOrigins = "http://localhost:3000"
		}
		
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
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")

			// Build allowed headers list - base headers always included
			baseHeaders := "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
			allowedHeaders := baseHeaders

			// Add development headers in development mode
			if !IsProduction() {
				allowedHeaders += ", X-Dev-User-ID"
			}

			// Add extra headers from environment variable (useful for testing production mode locally)
			extraHeaders := getEnv("EXTRA_ACCESS_CONTROL_ALLOW_HEADERS", "")
			if extraHeaders != "" {
				// Split by comma and add each header
				for _, header := range strings.Split(extraHeaders, ",") {
					header = strings.TrimSpace(header)
					if header != "" {
						allowedHeaders += ", " + header
					}
				}
				LogDebug("Added extra CORS headers", "extra_headers", extraHeaders)
			}

			c.Header("Access-Control-Allow-Headers", allowedHeaders)
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

	// In production, require HTTPS. In development, allow HTTP for localhost
	if IsProduction() {
		if strings.HasPrefix(origin, "http://") {
			return false
		}
	} else {
		// In development, allow HTTP only for localhost/127.0.0.1
		if strings.HasPrefix(origin, "http://") {
			if !strings.Contains(origin, "localhost") && !strings.Contains(origin, "127.0.0.1") {
				return false
			}
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
		// Security check - only allow dev auth in development
		if IsProduction() {
			LogWarn("Dev auth middleware called in production environment - blocking")
			c.Next()
			return
		}

		// Warn if production JWT config is set when using dev auth
		jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
		supabaseURL := getEnv("PUBLIC_SUPABASE_URL", "")
		if jwtSecret != "" || supabaseURL != "" {
			LogWarn("Production JWT configuration detected with dev auth - this may cause confusion")
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
		if IsProduction() {
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
