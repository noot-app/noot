package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
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
	Email    string                 `json:"email"`
	UserRole string                 `json:"role"`
	AppData  map[string]interface{} `json:"app_metadata"`
	UserData map[string]interface{} `json:"user_metadata"`
}

// JWKS structures for ES256 key handling
type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	// ECC fields
	X   string `json:"x"`
	Y   string `json:"y"`
	Crv string `json:"crv"`
	// RSA fields
	N string `json:"n"`
	E string `json:"e"`
}

// Cache for JWKS to avoid frequent requests
var (
	jwksCache      *JWKS
	jwksCacheTime  time.Time
	jwksCacheMutex sync.RWMutex
	jwksCacheTTL   = 1 * time.Hour // Cache for 1 hour
)

// Rate limiter for user creation to prevent abuse
var (
	userCreationLimiter = make(map[string][]time.Time) // Track multiple attempts per email
	userCreationMutex   sync.RWMutex
)

// cleanupUserCreationLimiter removes old entries from the rate limiter
func cleanupUserCreationLimiter() {
	userCreationMutex.Lock()
	defer userCreationMutex.Unlock()

	cutoff := time.Now().Add(-time.Hour) // Keep entries for 1 hour
	for email, attempts := range userCreationLimiter {
		// Filter out old attempts
		var validAttempts []time.Time
		for _, attempt := range attempts {
			if attempt.After(cutoff) {
				validAttempts = append(validAttempts, attempt)
			}
		}

		if len(validAttempts) == 0 {
			delete(userCreationLimiter, email)
		} else {
			userCreationLimiter[email] = validAttempts
		}
	}
}

// fetchJWKS fetches the JWKS from Supabase
func fetchJWKS(supabaseURL string) (*JWKS, error) {
	jwksCacheMutex.RLock()
	if jwksCache != nil && time.Since(jwksCacheTime) < jwksCacheTTL {
		defer jwksCacheMutex.RUnlock()
		return jwksCache, nil
	}
	jwksCacheMutex.RUnlock()

	// Supabase JWKS endpoint is at /auth/v1/.well-known/jwks.json
	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	jwksCacheMutex.Lock()
	jwksCache = &jwks
	jwksCacheTime = time.Now()
	jwksCacheMutex.Unlock()

	return &jwks, nil
}

// getPublicKeyFromJWKS extracts the public key for the given kid
func getPublicKeyFromJWKS(jwks *JWKS, kid string) (interface{}, error) {
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			switch key.Kty {
			case "EC":
				// Handle Elliptic Curve keys (ES256, ES384, ES512)
				if key.Alg == "ES256" || key.Alg == "ES384" || key.Alg == "ES512" {
					// Decode the X and Y coordinates from base64url
					xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
					if err != nil {
						return nil, fmt.Errorf("failed to decode X coordinate: %w", err)
					}

					yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
					if err != nil {
						return nil, fmt.Errorf("failed to decode Y coordinate: %w", err)
					}

					// Convert to big integers
					x := big.NewInt(0).SetBytes(xBytes)
					y := big.NewInt(0).SetBytes(yBytes)

					// Determine the curve based on the algorithm
					var curve elliptic.Curve
					switch key.Alg {
					case "ES256":
						curve = elliptic.P256()
					case "ES384":
						curve = elliptic.P384()
					case "ES512":
						curve = elliptic.P521()
					default:
						return nil, fmt.Errorf("unsupported ECDSA algorithm: %s", key.Alg)
					}

					// Create the ECDSA public key
					publicKey := &ecdsa.PublicKey{
						Curve: curve,
						X:     x,
						Y:     y,
					}

					return publicKey, nil
				}
			case "RSA":
				// Handle RSA keys (RS256, RS384, RS512)
				if key.Alg == "RS256" || key.Alg == "RS384" || key.Alg == "RS512" {
					// Decode the modulus (n) and exponent (e) from base64url
					nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
					if err != nil {
						return nil, fmt.Errorf("failed to decode RSA modulus: %w", err)
					}

					eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
					if err != nil {
						return nil, fmt.Errorf("failed to decode RSA exponent: %w", err)
					}

					// Convert to big integers
					n := big.NewInt(0).SetBytes(nBytes)
					e := big.NewInt(0).SetBytes(eBytes)

					// Create the RSA public key
					publicKey := &rsa.PublicKey{
						N: n,
						E: int(e.Int64()),
					}

					return publicKey, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("public key not found for kid: %s", kid)
}

// getAsymmetricPublicKey handles fetching public keys for asymmetric JWT verification
func getAsymmetricPublicKey(token *jwt.Token, supabaseURL, jwtSecret string) (interface{}, error) {
	LogDebug("getAsymmetricPublicKey started", "hasSupabaseURL", supabaseURL != "")

	// First, try to get the kid from the token header
	kidInterface, ok := token.Header["kid"]
	if !ok {
		LogError("No kid in token header", nil)
		return nil, fmt.Errorf("no kid in token header for asymmetric signing")
	}

	kid, ok := kidInterface.(string)
	if !ok {
		LogError("Invalid kid type in token header", nil)
		return nil, fmt.Errorf("invalid kid in token header")
	}

	LogDebug("Found kid in token", "kid", kid)

	// For Supabase, try to fetch JWKS
	if supabaseURL != "" {
		LogDebug("Fetching JWKS from Supabase URL")
		jwks, err := fetchJWKS(supabaseURL)
		if err != nil {
			LogWarn("Failed to fetch JWKS, falling back to secret: " + err.Error())
			// Fall back to trying the secret as a key (this usually won't work for asymmetric)
			return []byte(jwtSecret), nil
		}

		LogDebug("JWKS fetched successfully", "keyCount", len(jwks.Keys))

		publicKey, err := getPublicKeyFromJWKS(jwks, kid)
		if err != nil {
			LogWarn("Failed to get public key from JWKS: " + err.Error())
			// Fall back to trying the secret
			return []byte(jwtSecret), nil
		}

		LogDebug("Public key extracted from JWKS successfully")
		return publicKey, nil
	}

	LogDebug("No Supabase URL - falling back to secret")
	// If no Supabase URL, fall back to secret
	return []byte(jwtSecret), nil
}

// isUserCreationRateLimited checks if user creation is rate limited
func isUserCreationRateLimited(email string) bool {
	userCreationMutex.RLock()
	defer userCreationMutex.RUnlock()

	attempts, exists := userCreationLimiter[email]
	if !exists {
		return false
	}

	// Count valid attempts within the last 10 minutes
	cutoff := time.Now().Add(-10 * time.Minute)
	validAttempts := 0
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			validAttempts++
		}
	}

	// Allow up to 2 user creation attempts per email per 10 minutes
	return validAttempts >= 2
}

// recordUserCreationAttempt records a user creation attempt
func recordUserCreationAttempt(email string) {
	userCreationMutex.Lock()
	defer userCreationMutex.Unlock()

	now := time.Now()
	attempts, exists := userCreationLimiter[email]
	if !exists {
		attempts = make([]time.Time, 0, 2)
	}

	// Add current attempt
	attempts = append(attempts, now)
	userCreationLimiter[email] = attempts

	// Periodically cleanup old entries
	if len(userCreationLimiter) > 100 {
		go cleanupUserCreationLimiter()
	}
}

// JWTAuthMiddleware validates Supabase JWT tokens and creates/loads user context
// This middleware is used in production to authenticate API requests
func JWTAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		LogDebug("JWT auth middleware started")

		// Production environment validation with multiple safeguards
		env := strings.ToLower(getEnv("ENV", "production"))
		isProduction := env == "production"

		LogDebug("JWT middleware environment check", "env", env, "isProduction", isProduction)

		// Additional production validation checks
		if !isProduction {
			// Allow development only if explicitly configured and JWT secret is not set
			if env == "development" {
				jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
				if jwtSecret != "" {
					// Check if we already have a dev user set by DevAuthMiddleware
					if devUser, exists := c.Get("dev_user"); exists {
						// We have dev auth active - use that instead of JWT validation
						LogDebug("Using dev auth user instead of JWT validation")
						c.Set("auth_user", devUser)
						c.Next()
						return
					}
					// If JWT secret is configured and no dev user, use JWT auth
					LogWarn("JWT secret configured in development - enforcing JWT authentication")
					isProduction = true
				} else {
					LogDebug("Development mode with no JWT secret - skipping JWT auth")
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
			LogDebug("Not production mode - skipping JWT auth")
			c.Next()
			return
		}

		LogDebug("Production mode detected - enforcing JWT auth")

		// Get JWT from Authorization header
		authHeader := c.GetHeader("Authorization")
		LogDebug("Authorization header check", "hasHeader", authHeader != "")

		if authHeader == "" {
			LogWarn("No Authorization header found in production mode")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			LogWarn("Invalid Authorization header format", "header", authHeader[:min(20, len(authHeader))]+"...")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		LogDebug("JWT token extracted", "tokenLength", len(tokenString), "tokenPrefix", tokenString[:min(20, len(tokenString))]+"...")

		// Validate JWT
		LogDebug("Starting JWT validation")
		user, err := validateJWTAndGetUser(c.Request.Context(), tokenString, store)
		LogDebug("JWT validation completed", "hasError", err != nil, "hasUser", user != nil)

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
	LogDebug("validateJWTAndGetUser started")

	// Get Supabase URL for JWKS fetching (required for modern Supabase)
	supabaseURL := getEnv("PUBLIC_SUPABASE_URL", "")
	LogDebug("Got Supabase URL", "hasURL", supabaseURL != "")

	// Get JWT secret (only for legacy shared secret approach)
	jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
	LogDebug("Got JWT secret", "hasSecret", jwtSecret != "")

	// For modern Supabase, we only need the URL for JWKS
	// For legacy setups, we need the JWT secret
	if supabaseURL == "" && jwtSecret == "" {
		return nil, fmt.Errorf("either PUBLIC_SUPABASE_URL (for JWKS) or SUPABASE_JWT_SECRET (for legacy) must be set")
	}

	// Validate JWT secret strength if using legacy approach
	if jwtSecret != "" && len(jwtSecret) < 32 {
		LogWarn("JWT secret is shorter than recommended minimum of 32 characters")
	}

	LogDebug("Starting JWT parsing")
	// Parse and validate JWT with comprehensive options
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		LogDebug("JWT key func called", "method", fmt.Sprintf("%T", token.Method))
		// Check the signing method
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			LogDebug("Using HMAC signing method")
			// For HMAC (HS256, HS384, HS512), use the JWT secret
			return []byte(jwtSecret), nil
		case *jwt.SigningMethodECDSA:
			LogDebug("Using ECDSA signing method - fetching public key")
			// For ECDSA (ES256, ES384, ES512), we need to fetch the public key
			return getAsymmetricPublicKey(token, supabaseURL, jwtSecret)
		case *jwt.SigningMethodRSA:
			LogDebug("Using RSA signing method - fetching public key")
			// For RSA (RS256, RS384, RS512), we need to fetch the public key
			return getAsymmetricPublicKey(token, supabaseURL, jwtSecret)
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})

	LogDebug("JWT parsing completed", "hasError", err != nil, "isValid", token != nil && token.Valid)

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	LogDebug("About to extract claims from token")
	claims, ok := token.Claims.(*SupabaseJWTClaims)
	LogDebug("Claims extraction result", "ok", ok, "claims", claims != nil)
	if !ok {
		LogError("Failed to cast JWT claims to SupabaseJWTClaims", nil)
		return nil, fmt.Errorf("invalid JWT claims")
	}

	LogDebug("Claims extracted successfully", "subject", claims.Subject, "email", claims.Email)

	// Enhanced claim validation
	LogDebug("Starting JWT claims validation")
	if err := validateJWTClaims(claims); err != nil {
		LogError("JWT claims validation failed", err)
		return nil, fmt.Errorf("JWT claims validation failed: %w", err)
	}
	LogDebug("JWT claims validation passed")

	// Try to get existing user by Supabase ID (direct UUID lookup)
	LogDebug("Looking up user by UUID", "user_id", claims.Subject)

	if ctx == nil {
		LogError("Context is nil", nil)
		return nil, fmt.Errorf("context is nil")
	}

	if store == nil {
		LogError("Store is nil", nil)
		return nil, fmt.Errorf("storage store is nil")
	}

	LogDebug("Store is valid, calling GetUser")
	user, err := store.GetUser(ctx, claims.Subject)
	LogDebug("GetUser completed", "hasError", err != nil, "hasUser", user != nil)

	if err != nil {
		LogError("Database error while looking up user", err, "user_id", claims.Subject)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		// User doesn't exist, create them
		LogDebug("User not found, creating new user", "user_id", claims.Subject, "email", claims.Email)
		LogInfo("Creating new user from Supabase JWT", "user_id", claims.Subject, "email", claims.Email)

		// Create new user with Supabase auth.users.id directly as the ID
		user = &storage.User{
			ID:               claims.Subject,        // Use auth.users.id directly
			Email:            claims.Email,
			SubscriptionTier: storage.SubscriptionTierFree, // Default to free tier
		}

		// Extract handle and full_name from user_metadata
		if handle, ok := claims.UserData["handle"].(string); ok && handle != "" {
			user.Handle = handle
		} else {
			// Handle is required - this should not happen with proper signup flow
			LogError("User creation failed: handle is required", nil, "user_id", claims.Subject, "email", claims.Email)
			return nil, fmt.Errorf("user handle is required but not provided in user metadata")
		}

		if fullName, ok := claims.UserData["full_name"].(string); ok && fullName != "" {
			user.FullName = &fullName
		}

		// Save to database
		LogDebug("About to call CreateUser", "user", user != nil)
		if err := store.CreateUser(ctx, user); err != nil {
			LogError("Failed to create user", err, "supabase_id", claims.Subject, "email", claims.Email)
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		LogInfo("Successfully created new user", "user_id", user.ID, "email", user.Email)
		return user, nil
	}

	LogDebug("Found existing user", "user_id", user.ID, "email", user.Email)

	// Update user email if it changed in Supabase
	if user.Email != claims.Email {
		LogInfo("User email changed in Supabase, updating local record",
			"user_id", user.ID, "old_email", user.Email, "new_email", claims.Email)
		user.Email = claims.Email
		if err := store.UpdateUser(ctx, user); err != nil {
			LogWarn("Failed to update user email", "user_id", user.ID, "error", err.Error())
		}
	}

	return user, nil
}

// validateJWTClaims performs comprehensive validation of JWT claims
func validateJWTClaims(claims *SupabaseJWTClaims) error {
	LogDebug("validateJWTClaims started", "claims", claims != nil)

	if claims == nil {
		LogError("Claims is nil", nil)
		return fmt.Errorf("claims cannot be nil")
	}

	now := time.Now()
	LogDebug("Validating claims", "subject", claims.Subject, "email", claims.Email)

	// Validate expiration time
	LogDebug("Checking expiration time", "expiresAt", claims.ExpiresAt != nil)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(now) {
		return fmt.Errorf("token expired")
	}

	// Validate not before time
	LogDebug("Checking not before time", "notBefore", claims.NotBefore != nil)
	if claims.NotBefore != nil && claims.NotBefore.Time.After(now) {
		return fmt.Errorf("token not valid yet")
	}

	// Validate issued at time (not too far in the future)
	LogDebug("Checking issued at time", "issuedAt", claims.IssuedAt != nil)
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(now.Add(5*time.Minute)) {
		return fmt.Errorf("token issued too far in the future")
	}

	// Validate subject exists
	LogDebug("Checking subject", "subject", claims.Subject)
	if claims.Subject == "" {
		return fmt.Errorf("subject claim is required")
	}

	// Validate email exists and is reasonable
	LogDebug("Checking email", "email", claims.Email)
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

// isValidEmail performs robust email format validation
func isValidEmail(email string) bool {
	// Check basic length requirements
	if len(email) < 5 || len(email) > 254 {
		return false
	}

	// Must contain exactly one @ symbol
	atIndex := strings.Index(email, "@")
	if atIndex == -1 || atIndex != strings.LastIndex(email, "@") {
		return false
	}

	// Split into local and domain parts
	localPart := email[:atIndex]
	domainPart := email[atIndex+1:]

	// Validate local part
	if len(localPart) < 1 || len(localPart) > 64 {
		return false
	}

	// Validate domain part
	if len(domainPart) < 1 || len(domainPart) > 253 {
		return false
	}

	// Domain must contain at least one dot and end with a valid TLD
	dotIndex := strings.LastIndex(domainPart, ".")
	if dotIndex == -1 || dotIndex == 0 || dotIndex == len(domainPart)-1 {
		return false
	}

	// TLD must be at least 2 characters
	tld := domainPart[dotIndex+1:]
	if len(tld) < 2 {
		return false
	}

	// Basic character validation - no spaces, must be printable ASCII
	for _, r := range email {
		if r == ' ' || r < 32 || r > 126 {
			return false
		}
	}

	return true
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

	// Fall back to dev auth user (development only)
	if !IsProduction() {
		if user, exists := c.Get("dev_user"); exists {
			if devUser, ok := user.(*storage.User); ok {
				return devUser
			}
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

// RequireSubscriptionTiers middleware ensures user has one of the specified subscription tiers
func RequireSubscriptionTiers(allowedTiers ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthenticatedUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if user's tier is in the allowed list
		for _, tier := range allowedTiers {
			if user.SubscriptionTier == tier {
				c.Next()
				return
			}
		}

		// Build error message with allowed tiers
		var tierNames []string
		for _, tier := range allowedTiers {
			tierNames = append(tierNames, tier)
		}

		errorMsg := "Subscription required: " + strings.Join(tierNames, " or ")
		c.JSON(http.StatusForbidden, gin.H{"error": errorMsg})
		c.Abort()
	}
}

// RequireProSubscription middleware ensures user has pro subscription (backward compatibility)
func RequireProSubscription() gin.HandlerFunc {
	return RequireSubscriptionTiers(storage.SubscriptionTierPro)
}
