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
	"net/mail"
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

	// Cache for parsed public keys to avoid repeated conversion
	publicKeyCache      = make(map[string]interface{})
	publicKeyCacheMutex sync.RWMutex
)

// bytesToInt converts a byte slice to an int for RSA exponent parsing
// This handles variable-length exponent encoding more robustly than big.Int conversions
func bytesToInt(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("empty exponent")
	}
	var val int
	for _, x := range b {
		// Shift left 8 bits and add the next byte
		val = (val << 8) + int(x)
		// Guard against overflow (defensive; RSA exponents are typically small in practice)
		if val < 0 {
			return 0, fmt.Errorf("exponent overflow")
		}
	}
	return val, nil
}

// cachePublicKeysFromJWKS converts JWKS keys to public keys and caches them for faster lookup
func cachePublicKeysFromJWKS(jwks *JWKS) {
	keyMap := make(map[string]interface{})

	for _, key := range jwks.Keys {
		// Skip non-signing keys if use is specified
		if key.Use != "" && key.Use != "sig" {
			continue
		}

		var publicKey interface{}
		var err error

		switch key.Kty {
		case "EC":
			if key.Alg == "ES256" || key.Alg == "ES384" || key.Alg == "ES512" {
				publicKey, err = convertECJWKToPublicKey(&key)
			}
		case "RSA":
			if key.Alg == "RS256" || key.Alg == "RS384" || key.Alg == "RS512" {
				publicKey, err = convertRSAJWKToPublicKey(&key)
			}
		}

		if err == nil && publicKey != nil {
			keyMap[key.Kid] = publicKey
		}
	}

	publicKeyCacheMutex.Lock()
	publicKeyCache = keyMap
	publicKeyCacheMutex.Unlock()
}

// convertECJWKToPublicKey converts an EC JWK to an ECDSA public key
func convertECJWKToPublicKey(key *JWK) (*ecdsa.PublicKey, error) {
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

	// Determine the curve based on crv field first, then algorithm fallback
	var curve elliptic.Curve
	switch key.Crv {
	case "P-256", "P256":
		curve = elliptic.P256()
	case "P-384", "P384":
		curve = elliptic.P384()
	case "P-521", "P521":
		curve = elliptic.P521()
	default:
		// Fallback to algorithm mapping if crv field is absent or unrecognized
		switch key.Alg {
		case "ES256":
			curve = elliptic.P256()
		case "ES384":
			curve = elliptic.P384()
		case "ES512":
			curve = elliptic.P521()
		default:
			return nil, fmt.Errorf("unsupported ECDSA algorithm/curve: alg=%s crv=%s", key.Alg, key.Crv)
		}
	}

	// Create the ECDSA public key
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// convertRSAJWKToPublicKey converts an RSA JWK to an RSA public key
func convertRSAJWKToPublicKey(key *JWK) (*rsa.PublicKey, error) {
	// Decode the modulus (n) and exponent (e) from base64url
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RSA modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RSA exponent: %w", err)
	}

	// Convert modulus to big integer
	n := big.NewInt(0).SetBytes(nBytes)

	// Convert exponent bytes to int using robust parsing
	eInt, err := bytesToInt(eBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA exponent: %w", err)
	}

	// Create the RSA public key
	return &rsa.PublicKey{
		N: n,
		E: eInt,
	}, nil
}

// fetchJWKS fetches the JWKS from Supabase with double-check locking to prevent race conditions
func fetchJWKS(ctx context.Context, supabaseURL string) (*JWKS, error) {
	jwksCacheMutex.RLock()
	if jwksCache != nil && time.Since(jwksCacheTime) < jwksCacheTTL {
		defer jwksCacheMutex.RUnlock()
		return jwksCache, nil
	}
	jwksCacheMutex.RUnlock()

	// Acquire write lock and double-check (prevents duplicate network fetches under load)
	jwksCacheMutex.Lock()
	if jwksCache != nil && time.Since(jwksCacheTime) < jwksCacheTTL {
		// Someone else beat us to it while we were waiting for the write lock
		jwks := jwksCache
		jwksCacheMutex.Unlock()
		return jwks, nil
	}
	// Cache is still expired or empty - we need to fetch
	jwksCacheMutex.Unlock()

	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"

	// Use provided context with timeout for cancellation propagation
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(timeoutCtx, "GET", jwksURL, nil)
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

	// Cache the public keys for faster lookup
	cachePublicKeysFromJWKS(&jwks)

	return &jwks, nil
}

// fetchJWKSNoCache fetches JWKS without caching for key rotation scenarios
func fetchJWKSNoCache(ctx context.Context, supabaseURL string) (*JWKS, error) {
	jwksURL := strings.TrimSuffix(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"

	// Use provided context with timeout for cancellation propagation
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(timeoutCtx, "GET", jwksURL, nil)
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

	// Update cache
	jwksCacheMutex.Lock()
	jwksCache = &jwks
	jwksCacheTime = time.Now()
	jwksCacheMutex.Unlock()

	// Cache the public keys for faster lookup
	cachePublicKeysFromJWKS(&jwks)

	return &jwks, nil
}

// getPublicKeyFromJWKS extracts the public key for the given kid
func getPublicKeyFromJWKS(jwks *JWKS, kid string, tokenAlg string) (interface{}, error) {
	// First, try the public key cache for faster lookup
	publicKeyCacheMutex.RLock()
	if cachedKey, exists := publicKeyCache[kid]; exists {
		publicKeyCacheMutex.RUnlock()
		return cachedKey, nil
	}
	publicKeyCacheMutex.RUnlock()

	// Fallback to parsing from JWKS (handles case where cache wasn't populated)
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			// Skip non-signing keys if use is specified
			if key.Use != "" && key.Use != "sig" {
				continue
			}

			// Ensure JWK algorithm matches token algorithm for defense in depth
			if key.Alg != "" && key.Alg != tokenAlg {
				continue // Skip keys with mismatched algorithms
			}

			switch key.Kty {
			case "EC":
				// Handle Elliptic Curve keys (ES256, ES384, ES512)
				if key.Alg == "ES256" || key.Alg == "ES384" || key.Alg == "ES512" {
					return convertECJWKToPublicKey(&key)
				}
			case "RSA":
				// Handle RSA keys (RS256, RS384, RS512)
				if key.Alg == "RS256" || key.Alg == "RS384" || key.Alg == "RS512" {
					return convertRSAJWKToPublicKey(&key)
				}
			}
		}
	}
	return nil, fmt.Errorf("public key not found for kid: %s", kid)
}

// getAsymmetricPublicKey handles fetching public keys for asymmetric JWT verification
func getAsymmetricPublicKey(ctx context.Context, token *jwt.Token, supabaseURL string) (interface{}, error) {
	kidVal, ok := token.Header["kid"].(string)
	if !ok || kidVal == "" {
		return nil, fmt.Errorf("missing or invalid kid in token header")
	}

	// Only use JWKS for asymmetric algs - don't fall back to secret
	if supabaseURL == "" {
		return nil, fmt.Errorf("PUBLIC_SUPABASE_URL required for asymmetric verification")
	}

	// Get the token algorithm for matching
	tokenAlg, _ := token.Header["alg"].(string)

	// Try cached JWKS first
	jwks, err := fetchJWKS(ctx, supabaseURL)
	if err == nil {
		if publicKey, err := getPublicKeyFromJWKS(jwks, kidVal, tokenAlg); err == nil {
			return publicKey, nil
		}
	}

	// Force refresh JWKS once if kid not found or fetch failed (handles key rotation)
	if jwks, err = fetchJWKSNoCache(ctx, supabaseURL); err == nil {
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
//
// Error Responses:
// - 401 "Authorization header required" - Missing Authorization header
// - 401 "Invalid authorization header format" - Malformed Authorization header
// - 401 "Invalid token" - Invalid JWT signature or malformed token
// - 401 "Token expired" - Valid token but expired
// - 404 "User not found" - Valid token but user doesn't exist in database
// - 500 "Authentication configuration error" - Missing required environment variables
// - 503 "Authentication service temporarily unavailable" - JWKS fetch failed
func JWTAuthMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip authentication for OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Skip authentication for public endpoints
		if IsPublicEndpoint(c.Request.URL.Path) {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format (case-insensitive)
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate JWT
		user, err := validateJWTAndGetUser(c.Request.Context(), tokenString, store)

		if err != nil {
			// Log validation failure without exposing sensitive error details
			LogWarn("JWT validation failed - authentication rejected")

			// Provide more specific error responses based on the error type
			errorMsg := err.Error()

			// Token expired
			if strings.Contains(errorMsg, "token expired") || strings.Contains(errorMsg, "expired") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				c.Abort()
				return
			}

			// Invalid signature or malformed token
			if strings.Contains(errorMsg, "signature is invalid") ||
				strings.Contains(errorMsg, "failed to parse JWT") ||
				strings.Contains(errorMsg, "invalid JWT token") ||
				strings.Contains(errorMsg, "'none' signing method") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			// User not found in database (valid token but no user record)
			if strings.Contains(errorMsg, "user profile not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				c.Abort()
				return
			}

			// Database connection/query errors
			if strings.Contains(errorMsg, "database error") {
				LogError("Database error during authentication - connection or query failed", nil)
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service temporarily unavailable"})
				c.Abort()
				return
			}

			// JWKS/key resolution issues (likely temporary)
			if strings.Contains(errorMsg, "unable to resolve public key") ||
				strings.Contains(errorMsg, "failed to fetch JWKS") {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Authentication service temporarily unavailable"})
				c.Abort()
				return
			}

			// Configuration issues
			if strings.Contains(errorMsg, "PUBLIC_SUPABASE_URL") ||
				strings.Contains(errorMsg, "SUPABASE_JWT_SECRET") {
				LogError("Authentication configuration error - missing required environment variables", nil)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication configuration error"})
				c.Abort()
				return
			}

			// Generic invalid token for other JWT-related errors
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// User is successfully authenticated at this point
		LogDebug("User authenticated successfully", "user_id", user.ID)

		// Set the authenticated user in context for handlers to use
		c.Set("auth_user", user)
		c.Next()
	}
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
	if expectedAudience == "" {
		// Default to "authenticated" for typical Supabase user tokens
		expectedAudience = "authenticated"
	}

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
			return getAsymmetricPublicKey(ctx, token, supabaseURL)
		case *jwt.SigningMethodRSA:
			// For RSA (RS256, RS384, RS512), we need to fetch the public key
			return getAsymmetricPublicKey(ctx, token, supabaseURL)
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
		LogError("JWT claims validation failed - invalid or missing required claims", nil)
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
			"user_id", claims.Subject)
		return nil, fmt.Errorf("user profile not found")
	}

	// Update user email if it changed in Supabase
	if user.Email != claims.Email {
		LogInfo("User email changed in Supabase, updating local record", "user_id", user.ID)
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

	// For Supabase user tokens, we strictly enforce role to be "authenticated"
	// This prevents admin/service tokens from being used for user operations
	if claims.UserRole != "" && claims.UserRole != "authenticated" {
		LogWarn("Non-user role rejected in JWT", "role", claims.UserRole)
		return fmt.Errorf("invalid role for user operation: %s", claims.UserRole)
	}

	return nil
}

// isValidEmail performs RFC-compliant email format validation using Go's built-in parser
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
