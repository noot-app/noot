package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/storage"
)

// DualAuthMiddleware validates both JWT and API key authentication
// This middleware supports both Bearer tokens (JWT) and X-API-Key headers
//
// Security Features:
// - Supports JWT authentication (existing functionality)
// - Supports API key authentication for Pro users
// - Enforces API key scope restrictions (read vs read_write)
// - Updates API key last_used_at timestamp
// - Loads secure user context for all downstream handlers
//
// Header Formats:
// - Authorization: Bearer <jwt_token>
// - X-API-Key: <api_key>
//
// Error Responses:
// - 401 "Authorization required" - Missing both Authorization and X-API-Key headers
// - 401 "Invalid API key" - Malformed or invalid API key
// - 401 "API key revoked or expired" - Key is no longer active
// - 403 "Insufficient scope" - API key doesn't have required permissions for the request
func DualAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Skip authentication for public endpoints
		if IsPublicEndpoint(c.Request.URL.Path) {
			LogDebug("Skipping auth for public endpoint", "path", c.Request.URL.Path)
			c.Next()
			return
		}

		// Check for API key first (X-API-Key header)
		apiKeyHeader := c.GetHeader("X-API-Key")
		if apiKeyHeader != "" {
			if err := authenticateAPIKey(c, apiKeyHeader, store); err != nil {
				LogWarn("API key authentication failed", "error", err.Error())
				handleAPIKeyAuthError(c, err)
				return
			}
			// API key authentication successful
			c.Next()
			return
		}

		// Fall back to JWT authentication
		jwtMiddleware := JWTAuthMiddleware(store)
		jwtMiddleware(c)
	}
}

// authenticateAPIKey handles API key authentication
func authenticateAPIKey(c *gin.Context, apiKey string, store storage.Store) error {
	ctx := c.Request.Context()

	// Extract prefix from the API key for lookup
	prefix, err := storage.ExtractPrefix(apiKey)
	if err != nil {
		return &APIKeyError{Type: "invalid_format", Message: "Invalid API key format"}
	}

	// Look up the API key by prefix
	keyRecord, err := store.GetAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		return &APIKeyError{Type: "lookup_failed", Message: "Failed to validate API key"}
	}

	if keyRecord == nil {
		return &APIKeyError{Type: "not_found", Message: "API key not found"}
	}

	// Check if the key is active (not revoked or expired)
	if !keyRecord.IsActive() {
		return &APIKeyError{Type: "inactive", Message: "API key is revoked or expired"}
	}

	// Verify the full API key matches the stored hash
	if !storage.VerifyAPIKey(apiKey, keyRecord.Hash) {
		return &APIKeyError{Type: "invalid", Message: "Invalid API key"}
	}

	// Get the user associated with this API key
	user, err := store.GetUser(ctx, keyRecord.UserID)
	if err != nil {
		return &APIKeyError{Type: "user_lookup_failed", Message: "Failed to get user for API key"}
	}

	if user == nil {
		return &APIKeyError{Type: "user_not_found", Message: "User not found for API key"}
	}

	// Verify user is still Pro (keys are only for Pro users)
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		return &APIKeyError{Type: "subscription_required", Message: "API key requires Pro subscription"}
	}

	// Check scope permissions
	if err := checkAPIKeyScope(c.Request.Method, keyRecord.Scope); err != nil {
		return err
	}

	// Update last used timestamp (async to not block the request)
	go func() {
		// Use a background context to avoid cancellation
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := store.UpdateAPIKeyLastUsed(bgCtx, keyRecord.ID, time.Now()); err != nil {
			LogWarn("Failed to update API key last used timestamp", "key_id", keyRecord.ID, "error", err.Error())
		}
	}()

	// Set authenticated user and API key info in context
	c.Set("auth_user", user)
	c.Set("api_key_id", keyRecord.ID)
	c.Set("api_key_scope", keyRecord.Scope)
	c.Set("auth_method", "api_key")

	LogDebug("User authenticated successfully via API key", "user_id", user.ID, "key_scope", keyRecord.Scope)
	return nil
}

// checkAPIKeyScope verifies that the API key has sufficient permissions for the request
func checkAPIKeyScope(method, scope string) error {
	switch scope {
	case "read":
		// Read scope only allows GET requests
		if method != http.MethodGet {
			return &APIKeyError{Type: "insufficient_scope", Message: "Read-only API key cannot perform write operations"}
		}
	case "read_write":
		// Read_write scope allows all methods
		return nil
	default:
		return &APIKeyError{Type: "invalid_scope", Message: "API key has invalid scope"}
	}
	return nil
}

// APIKeyError represents an API key authentication error
type APIKeyError struct {
	Type    string
	Message string
}

func (e *APIKeyError) Error() string {
	return e.Message
}

// handleAPIKeyAuthError handles API key authentication errors with appropriate HTTP responses
func handleAPIKeyAuthError(c *gin.Context, err error) {
	if apiErr, ok := err.(*APIKeyError); ok {
		switch apiErr.Type {
		case "invalid_format", "not_found", "invalid":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		case "inactive":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key revoked or expired"})
		case "insufficient_scope":
			c.JSON(http.StatusForbidden, gin.H{"error": apiErr.Message})
		case "subscription_required":
			c.JSON(http.StatusForbidden, gin.H{"error": "API key requires Pro subscription"})
		case "user_not_found", "user_lookup_failed":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		case "lookup_failed":
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service temporarily unavailable"})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
	}
	c.Abort()
}

// GetAuthenticatedMethod returns the authentication method used (jwt or api_key)
func GetAuthenticatedMethod(c *gin.Context) string {
	if method, exists := c.Get("auth_method"); exists {
		if methodStr, ok := method.(string); ok {
			return methodStr
		}
	}
	return "jwt" // Default to JWT if not specified
}

// GetAPIKeyScope returns the scope of the API key used for authentication
func GetAPIKeyScope(c *gin.Context) string {
	if scope, exists := c.Get("api_key_scope"); exists {
		if scopeStr, ok := scope.(string); ok {
			return scopeStr
		}
	}
	return ""
}

// GetAPIKeyID returns the ID of the API key used for authentication
func GetAPIKeyID(c *gin.Context) string {
	if id, exists := c.Get("api_key_id"); exists {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}
	return ""
}
