package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grantbirki/noot/internal/storage"
)

// SupabaseJWTClaims represents the claims in a Supabase JWT
type SupabaseJWTClaims struct {
	jwt.RegisteredClaims
	Email    string `json:"email"`
	UserRole string `json:"role"`
	AppData  map[string]interface{} `json:"app_metadata"`
	UserData map[string]interface{} `json:"user_metadata"`
}

// JWTAuthMiddleware validates Supabase JWT tokens and creates/loads user context
// This middleware is used in production to authenticate API requests
func JWTAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip JWT auth in development mode - use DevAuthMiddleware instead
		env := strings.ToLower(getEnv("ENV", "production"))
		if env == "development" {
			c.Next()
			return
		}

		// Get JWT from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate JWT
		user, err := validateJWTAndGetUser(c.Request.Context(), tokenString, store)
		if err != nil {
			LogWarn("JWT validation failed", "error", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if user == nil {
			LogWarn("JWT validated but user not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set the authenticated user in context for handlers to use
		c.Set("auth_user", user)
		LogDebug("JWT auth successful", "user_id", user.ID, "email", user.Email)

		c.Next()
	}
}

// validateJWTAndGetUser validates a Supabase JWT and returns the corresponding user
func validateJWTAndGetUser(ctx context.Context, tokenString string, store storage.Store) (*storage.User, error) {
	// Get JWT secret from environment
	jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET environment variable not set")
	}

	// Parse and validate JWT
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	claims, ok := token.Claims.(*SupabaseJWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	// Try to get existing user by Supabase ID (subject)
	user, err := store.GetUserBySubject(ctx, "supabase", claims.Subject)
	if err != nil {
		// If user doesn't exist, create them
		LogInfo("Creating new user from Supabase JWT", "supabase_id", claims.Subject, "email", claims.Email)
		
		// Create new user with Supabase ID mapping
		user = &storage.User{
			Email:            claims.Email,
			Provider:         "supabase",
			Subject:          claims.Subject,
			SubscriptionTier: storage.SubscriptionTierFree, // Default to free tier
		}

		// Save to database
		if err := store.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		LogInfo("Successfully created new user", "user_id", user.ID, "email", user.Email)
		return user, nil
	}

	// Update user email if it changed in Supabase
	if user.Email != claims.Email {
		LogWarn("User email changed in Supabase but cannot update local record", 
			"user_id", user.ID, "old_email", user.Email, "new_email", claims.Email)
		// TODO: Implement UpdateUser method in Store interface if email updates are needed
	}

	return user, nil
}

// GetAuthenticatedUser extracts the authenticated user from Gin context
// Works with both dev auth and JWT auth middleware
func GetAuthenticatedUser(c *gin.Context) *storage.User {
	// Check for JWT auth user first (production)
	if user, exists := c.Get("auth_user"); exists {
		if authUser, ok := user.(*storage.User); ok {
			return authUser
		}
	}

	// Fall back to dev auth user (development)
	if user, exists := c.Get("dev_user"); exists {
		if devUser, ok := user.(*storage.User); ok {
			return devUser
		}
	}

	return nil
}

// RequireAuth middleware ensures a user is authenticated
// Can be used on individual routes that need authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthenticatedUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireProSubscription middleware ensures user has pro subscription
func RequireProSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthenticatedUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		if user.SubscriptionTier != storage.SubscriptionTierPro {
			c.JSON(http.StatusForbidden, gin.H{"error": "Pro subscription required"})
			c.Abort()
			return
		}
		c.Next()
	}
}