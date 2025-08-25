package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

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

// Rate limiter for user creation to prevent abuse
var (
	userCreationLimiter = make(map[string]time.Time)
	userCreationMutex   sync.RWMutex
)

// cleanupUserCreationLimiter removes old entries from the rate limiter
func cleanupUserCreationLimiter() {
	userCreationMutex.Lock()
	defer userCreationMutex.Unlock()
	
	cutoff := time.Now().Add(-time.Hour) // Keep entries for 1 hour
	for email, lastAttempt := range userCreationLimiter {
		if lastAttempt.Before(cutoff) {
			delete(userCreationLimiter, email)
		}
	}
}

// isUserCreationRateLimited checks if user creation is rate limited
func isUserCreationRateLimited(email string) bool {
	userCreationMutex.RLock()
	defer userCreationMutex.RUnlock()
	
	lastAttempt, exists := userCreationLimiter[email]
	if !exists {
		return false
	}
	
	// Allow one user creation per email per 10 minutes
	return time.Since(lastAttempt) < 10*time.Minute
}

// recordUserCreationAttempt records a user creation attempt
func recordUserCreationAttempt(email string) {
	userCreationMutex.Lock()
	defer userCreationMutex.Unlock()
	
	userCreationLimiter[email] = time.Now()
	
	// Periodically cleanup old entries
	if len(userCreationLimiter) > 100 {
		go cleanupUserCreationLimiter()
	}
}

// JWTAuthMiddleware validates Supabase JWT tokens and creates/loads user context
// This middleware is used in production to authenticate API requests
func JWTAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Production environment validation with multiple safeguards
		env := strings.ToLower(getEnv("ENV", "production"))
		isProduction := env == "production"
		
		// Additional production validation checks
		if !isProduction {
			// Allow development only if explicitly configured and JWT secret is not set
			if env == "development" {
				jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
				if jwtSecret != "" {
					// If JWT secret is configured, we should use JWT auth even in development
					LogWarn("JWT secret configured in development - enforcing JWT authentication")
					isProduction = true
				} else {
					c.Next()
					return
				}
			} else {
				// Unknown environment - default to production security
				LogWarn("Unknown environment detected, defaulting to production security", "env", env)
				isProduction = true
			}
		}
		
		if !isProduction {
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
	
	// Validate JWT secret strength (minimum 32 characters)
	if len(jwtSecret) < 32 {
		LogWarn("JWT secret is shorter than recommended minimum of 32 characters")
	}

	// Parse and validate JWT with comprehensive options
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method - only allow HMAC
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
	
	// Enhanced claim validation
	if err := validateJWTClaims(claims); err != nil {
		return nil, fmt.Errorf("JWT claims validation failed: %w", err)
	}

	// Try to get existing user by Supabase ID (subject)
	user, err := store.GetUserBySubject(ctx, "supabase", claims.Subject)
	if err != nil {
		// Check rate limiting before creating new user
		if isUserCreationRateLimited(claims.Email) {
			return nil, fmt.Errorf("user creation rate limited for email: %s", claims.Email)
		}
		
		// Record the creation attempt
		recordUserCreationAttempt(claims.Email)
		
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
			LogError("Failed to create user", err, "supabase_id", claims.Subject, "email", claims.Email)
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

// validateJWTClaims performs comprehensive validation of JWT claims
func validateJWTClaims(claims *SupabaseJWTClaims) error {
	now := time.Now()
	
	// Validate expiration time
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(now) {
		return fmt.Errorf("token expired")
	}
	
	// Validate not before time
	if claims.NotBefore != nil && claims.NotBefore.Time.After(now) {
		return fmt.Errorf("token not valid yet")
	}
	
	// Validate issued at time (not too far in the future)
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(now.Add(5*time.Minute)) {
		return fmt.Errorf("token issued too far in the future")
	}
	
	// Validate subject exists
	if claims.Subject == "" {
		return fmt.Errorf("subject claim is required")
	}
	
	// Validate email exists and is reasonable
	if claims.Email == "" {
		return fmt.Errorf("email claim is required")
	}
	if !isValidEmail(claims.Email) {
		return fmt.Errorf("invalid email format in claims")
	}
	
	// Optional issuer validation (if configured)
	expectedIssuer := getEnv("SUPABASE_JWT_ISSUER", "")
	if expectedIssuer != "" && claims.Issuer != expectedIssuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}
	
	// Optional audience validation (if configured)
	expectedAudience := getEnv("SUPABASE_JWT_AUDIENCE", "")
	if expectedAudience != "" {
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == expectedAudience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return fmt.Errorf("invalid audience: expected %s", expectedAudience)
		}
	}
	
	return nil
}

// isValidEmail performs basic email format validation
func isValidEmail(email string) bool {
	// Basic validation - contains @ and has reasonable length
	return strings.Contains(email, "@") && len(email) > 3 && len(email) < 255
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