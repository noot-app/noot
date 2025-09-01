package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
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

		LogDebug("Request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
		)
	}
}

// RecoveryMiddleware recovers from panics with secure error handling
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

				// In production, never expose stack traces or internal details
				if IsProduction() {
					c.JSON(500, gin.H{
						"error": "Internal server error",
					})
				} else if isDebugMode() || isDevMode() {
					// Only in debug/dev mode, include stack trace
					stack := captureStackGin(4)
					c.JSON(500, gin.H{
						"error":      "Internal server error (panic recovered)",
						"details":    stack,
						"request_id": requestID,
					})
				} else {
					// Development but not debug mode - include request ID but no stack
					c.JSON(500, gin.H{
						"error":      "Internal server error",
						"request_id": requestID,
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

		// Prepare allowed headers (needed for both regular and OPTIONS requests)
		allowedHeaders := "Authorization, Content-Type, Accept"
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
			c.Header("Access-Control-Allow-Headers", allowedHeaders)
			// Note: Access-Control-Allow-Credentials removed - not needed for bearer token auth
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
		} else if requestOrigin != "" {
			// Log unauthorized origins at different levels based on environment
			if IsProduction() {
				// could be noisy from bot traffic so moved to debug
				LogDebug("CORS request from unauthorized origin", "origin", requestOrigin)
			} else {
				LogWarn("CORS request from unauthorized origin", "origin", requestOrigin)
			}
		}

		if c.Request.Method == "OPTIONS" {
			// Set CORS headers for preflight requests before aborting
			if originAllowed {
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
				c.Header("Access-Control-Allow-Headers", allowedHeaders)
				c.Header("Access-Control-Max-Age", "86400")
			}
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isValidOrigin validates that a CORS origin is properly formatted and secure
func isValidOrigin(origin string) bool {
	// Parse the origin URL to validate structure and extract components
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Must be http or https scheme only
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// In production, require HTTPS only
	if IsProduction() {
		return u.Scheme == "https"
	}

	// In development, allow HTTP only for localhost/loopback addresses
	if u.Scheme == "http" {
		hostname := strings.ToLower(u.Hostname()) // Case-insensitive hostname check
		// Allow localhost, 127.0.0.1, and IPv6 loopback
		return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
	}

	// HTTPS is always allowed in development
	return true
}

// captureStackGin captures stack trace for Gin context
func captureStackGin(skip int) []string {
	var stack []string
	for i := skip; i < skip+20; i++ { // Increased from 10 to 20 for better debugging
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

// generateRequestID generates a random request ID with proper error handling
func generateRequestID() string {
	bytes := make([]byte, 16) // Increased from 6 to 16 bytes for stronger randomness
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		LogWarn("Failed to generate cryptographic request ID, using timestamp fallback", "error", err.Error())
		return fmt.Sprintf("ts_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing attacks
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking/UI redressing attacks
		c.Header("X-Frame-Options", "DENY")

		// Referrer policy - good balance between privacy and debugging
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS for HTTPS in production
		if IsProduction() && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}
