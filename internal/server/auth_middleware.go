package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"slices"
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
	httpClient     = &http.Client{Timeout: 5 * time.Second}
)

// fetchJWKS fetches the JWKS from Supabase
func fetchJWKS(supabaseURL string) (*JWKS, error) {
	jwksCacheMutex.RLock()
	if jwksCache != nil && time.Since(jwksCacheTime) < jwksCacheTTL {
		defer jwksCacheMutex.RUnlock()
		return jwksCache, nil
	}
	jwksCacheMutex.RUnlock()

	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	jwksCacheMutex.Lock()
	jwksCache = &jwks
	jwksCacheTime = time.Now()
	jwksCacheMutex.Unlock()

	return &jwks, nil
}

// fetchJWKSNoCache fetches JWKS without caching for key rotation scenarios
func fetchJWKSNoCache(supabaseURL string) (*JWKS, error) {
	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"

	req, err := http.NewRequest("GET", jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	// Add context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Update cache
	jwksCacheMutex.Lock()
	jwksCache = &jwks
	jwksCacheTime = time.Now()
	jwksCacheMutex.Unlock()

	return &jwks, nil
}

// getPublicKeyFromJWKS extracts the public key for the given kid
func getPublicKeyFromJWKS(jwks *JWKS, kid string, tokenAlg string) (interface{}, error) {
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			// Ensure JWK algorithm matches token algorithm for defense in depth
			if key.Alg != "" && key.Alg != tokenAlg {
				continue // Skip keys with mismatched algorithms
			}

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
func getAsymmetricPublicKey(token *jwt.Token, supabaseURL string) (interface{}, error) {
	kidVal, ok := token.Header["kid"].(string)
	if !ok || kidVal == "" {
		return nil, fmt.Errorf("missing or invalid kid in token header")
	}

	// Only use JWKS for asymmetric algs - don't fall back to secret
	if supabaseURL == "" {
		return nil, fmt.Errorf("PUBLIC_SUPABASE_URL required for asymmetric verification")
	}

	// Get the token algorithm for matching
	tokenAlg, ok := token.Header["alg"].(string)
	if !ok {
		tokenAlg = "" // Allow empty alg for fallback
	}

	// Try cached JWKS first
	jwks, err := fetchJWKS(supabaseURL)
	if err == nil {
		if publicKey, err := getPublicKeyFromJWKS(jwks, kidVal, tokenAlg); err == nil {
			return publicKey, nil
		}
	}

	// Force refresh JWKS once if kid not found or fetch failed (handles key rotation)
	if jwks, err = fetchJWKSNoCache(supabaseURL); err == nil {
		if publicKey, err := getPublicKeyFromJWKS(jwks, kidVal, tokenAlg); err == nil {
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("unable to resolve public key for kid %s", kidVal)
}

// JWTAuthMiddleware validates Supabase JWT tokens and loads user context
// This middleware is used in production to authenticate API requests
//
// Security Features:
// - Validates JWT signature using JWKS from Supabase
// - Enforces token expiration and claim validation
// - Loads secure user context for all downstream handlers
// - Defaults to production security mode for unknown environments
// - Skips authentication for public endpoints like health checks
func JWTAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip authentication for OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Skip authentication for public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			LogDebug("Skipping JWT auth for public endpoint", "path", c.Request.URL.Path)
			c.Next()
			return
		}

		// Production environment validation with multiple safeguards
		env := strings.ToLower(getEnv("ENV", "production"))
		isProduction := env == "production"

		// Additional production validation checks
		if !isProduction {
			// Allow development only if explicitly configured and JWT secret is not set
			if env == "development" {
				jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
				if jwtSecret != "" {
					// If JWT secret is configured in development, use JWT auth
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

		// Get JWT from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			LogWarn("No Authorization header found in production mode")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			LogWarn("Invalid Authorization header format")
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
		c.Next()
	}
}

// isPublicEndpoint checks if the given path is a public endpoint that doesn't require authentication
func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/api/v1/health",
		"/api/v1/openapi.yaml",
	}

	// Exact match for specific endpoints
	if slices.Contains(publicEndpoints, path) {
		return true
	}

	// Prefix match for docs endpoints to handle subpaths
	if strings.HasPrefix(path, "/api/v1/docs") {
		return true
	}

	return false
}

// validateJWTAndGetUser validates a Supabase JWT and returns the corresponding user
func validateJWTAndGetUser(ctx context.Context, tokenString string, store storage.Store) (*storage.User, error) {
	// Get Supabase URL for JWKS fetching (required for modern Supabase)
	supabaseURL := getEnv("PUBLIC_SUPABASE_URL", "")

	// Get JWT secret (only for legacy shared secret approach)
	jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")

	// For modern Supabase, we only need the URL for JWKS
	// For legacy setups, we need the JWT secret
	if supabaseURL == "" && jwtSecret == "" {
		return nil, fmt.Errorf("either PUBLIC_SUPABASE_URL (for JWKS) or SUPABASE_JWT_SECRET (for legacy) must be set")
	}

	// Get expected issuer and audience for validation
	expectedIssuer := getEnv("SUPABASE_JWT_ISSUER", "")
	if expectedIssuer == "" && supabaseURL != "" {
		// Derive default issuer from Supabase URL
		expectedIssuer = strings.TrimSuffix(supabaseURL, "/") + "/auth/v1"
	}
	expectedAudience := getEnv("SUPABASE_JWT_AUDIENCE", "")

	// Create parser with validation options and leeway
	parserOptions := []jwt.ParserOption{
		jwt.WithLeeway(60 * time.Second),
		jwt.WithValidMethods([]string{
			"HS256", "HS384", "HS512",
			"RS256", "RS384", "RS512",
			"ES256", "ES384", "ES512",
		}),
	}
	if expectedIssuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(expectedIssuer))
	}
	if expectedAudience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(expectedAudience))
	}

	parser := jwt.NewParser(parserOptions...)

	// Parse and validate JWT with comprehensive options
	token, err := parser.ParseWithClaims(tokenString, &SupabaseJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Explicitly reject "none" algorithm
		if token.Method == jwt.SigningMethodNone {
			return nil, fmt.Errorf("'none' signing method is not allowed")
		}

		// Check the signing method
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			// For HMAC (HS256, HS384, HS512), use the JWT secret
			if jwtSecret == "" {
				return nil, fmt.Errorf("SUPABASE_JWT_SECRET required for HMAC verification")
			}
			return []byte(jwtSecret), nil
		case *jwt.SigningMethodECDSA:
			// For ECDSA (ES256, ES384, ES512), we need to fetch the public key
			return getAsymmetricPublicKey(token, supabaseURL)
		case *jwt.SigningMethodRSA:
			// For RSA (RS256, RS384, RS512), we need to fetch the public key
			return getAsymmetricPublicKey(token, supabaseURL)
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	claims, ok := token.Claims.(*SupabaseJWTClaims)
	if !ok {
		LogError("Failed to cast JWT claims to SupabaseJWTClaims", nil)
		return nil, fmt.Errorf("invalid JWT claims")
	}

	// Enhanced claim validation
	if err := validateJWTClaims(claims); err != nil {
		LogError("JWT claims validation failed", err)
		return nil, fmt.Errorf("JWT claims validation failed: %w", err)
	}

	if ctx == nil {
		LogError("Context is nil", nil)
		return nil, fmt.Errorf("context is nil")
	}

	if store == nil {
		LogError("Store is nil", nil)
		return nil, fmt.Errorf("storage store is nil")
	}

	user, err := store.GetUser(ctx, claims.Subject)

	if err != nil {
		LogError("Database error while looking up user", err, "user_id", claims.Subject)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user == nil {
		LogError("User profile not found despite valid JWT - data inconsistency detected", nil,
			"user_id", claims.Subject, "email", claims.Email)
		return nil, fmt.Errorf("user profile not found")
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
// Note: exp, nbf, iat, issuer, and audience are now validated by jwt.Parser with options
func validateJWTClaims(claims *SupabaseJWTClaims) error {
	if claims == nil {
		LogError("Claims is nil", nil)
		return fmt.Errorf("claims cannot be nil")
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

	// For Supabase user tokens, we typically expect role to be "authenticated"
	// This helps prevent admin/service tokens from being used for user operations
	if claims.UserRole != "" && claims.UserRole != "authenticated" {
		LogWarn("Non-user role detected in JWT", "role", claims.UserRole)
		// Optionally return error here if you want to strictly enforce user tokens only
		// return fmt.Errorf("invalid role for user operation: %s", claims.UserRole)
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

	// Basic character validation - no spaces, must be printable ASCII, no dangerous chars
	for _, r := range email {
		if r == ' ' || r < 32 || r > 126 || r == '<' || r == '>' {
			return false
		}
	}

	return true
}

// GetAuthenticatedUser extracts the authenticated user from Gin context
func GetAuthenticatedUser(c *gin.Context) *storage.User {
	// Check for JWT auth user
	if user, exists := c.Get("auth_user"); exists {
		if authUser, ok := user.(*storage.User); ok {
			return authUser
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
		tierNames := append([]string(nil), allowedTiers...)

		errorMsg := "Subscription required: " + strings.Join(tierNames, " or ")
		c.JSON(http.StatusForbidden, gin.H{"error": errorMsg})
		c.Abort()
	}
}

// RequireProSubscription middleware ensures user has pro subscription (backward compatibility)
func RequireProSubscription() gin.HandlerFunc {
	return RequireSubscriptionTiers(storage.SubscriptionTierPro)
}
