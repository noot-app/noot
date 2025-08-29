package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock store for testing that implements only the methods needed for auth testing
type mockStore struct {
	users map[string]*storage.User
}

func (m *mockStore) GetUser(ctx context.Context, userID string) (*storage.User, error) {
	if user, exists := m.users[userID]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *mockStore) CreateUser(ctx context.Context, user *storage.User) error {
	if m.users == nil {
		m.users = make(map[string]*storage.User)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockStore) UpdateUser(ctx context.Context, user *storage.User) error {
	if m.users == nil {
		m.users = make(map[string]*storage.User)
	}
	m.users[user.ID] = user
	return nil
}

// Mock implementations of other required methods with minimal implementations
func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) {
	return nil, nil
}
func (m *mockStore) CreateConsumption(ctx context.Context, consumption *storage.Consumption) error {
	return nil
}
func (m *mockStore) GetConsumption(ctx context.Context, id string) (*storage.Consumption, error) {
	return nil, nil
}
func (m *mockStore) UpdateConsumption(ctx context.Context, consumption *storage.Consumption) error {
	return nil
}
func (m *mockStore) DeleteConsumption(ctx context.Context, id string) error { return nil }
func (m *mockStore) GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*storage.Consumption, error) {
	return nil, nil
}
func (m *mockStore) GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*storage.Consumption, error) {
	return nil, nil
}
func (m *mockStore) GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*storage.NutritionSummary, error) {
	return nil, nil
}
func (m *mockStore) CreateConsumptionItem(ctx context.Context, item *storage.ConsumptionItem) error {
	return nil
}
func (m *mockStore) GetConsumptionItems(ctx context.Context, consumptionID string) ([]*storage.ConsumptionItem, error) {
	return nil, nil
}
func (m *mockStore) UpdateConsumptionItem(ctx context.Context, item *storage.ConsumptionItem) error {
	return nil
}
func (m *mockStore) DeleteConsumptionItem(ctx context.Context, id string) error { return nil }
func (m *mockStore) DeleteConsumptionItemsByConsumption(ctx context.Context, consumptionID string) error {
	return nil
}
func (m *mockStore) CreateItem(ctx context.Context, item *storage.Item) error      { return nil }
func (m *mockStore) GetItem(ctx context.Context, id string) (*storage.Item, error) { return nil, nil }
func (m *mockStore) GetItemByName(ctx context.Context, normalizedName, normalizedBrand string) (*storage.Item, error) {
	return nil, nil
}
func (m *mockStore) UpdateItem(ctx context.Context, item *storage.Item) error { return nil }
func (m *mockStore) GetStaleItems(ctx context.Context, staleAfter time.Time) ([]*storage.Item, error) {
	return nil, nil
}
func (m *mockStore) UpsertUserGoal(ctx context.Context, goal *storage.UserGoal) error { return nil }
func (m *mockStore) GetUserGoal(ctx context.Context, userID, name string) (*storage.UserGoal, error) {
	return nil, nil
}
func (m *mockStore) GetUserGoals(ctx context.Context, userID string) ([]*storage.UserGoal, error) {
	return nil, nil
}
func (m *mockStore) DeleteUserGoal(ctx context.Context, userID, name string) error    { return nil }
func (m *mockStore) SetActiveGoal(ctx context.Context, userID, goalName string) error { return nil }
func (m *mockStore) ClearActiveGoal(ctx context.Context, userID string) error         { return nil }
func (m *mockStore) GetActiveGoalName(ctx context.Context, userID string) (*string, error) {
	return nil, nil
}
func (m *mockStore) UpsertUserBiometrics(ctx context.Context, biometrics *storage.UserBiometrics) error {
	return nil
}
func (m *mockStore) GetUserBiometrics(ctx context.Context, userID string) (*storage.UserBiometrics, error) {
	return nil, nil
}
func (m *mockStore) DeleteUserBiometrics(ctx context.Context, userID string) error       { return nil }
func (m *mockStore) CreateItemAlias(ctx context.Context, alias *storage.ItemAlias) error { return nil }
func (m *mockStore) GetCanonicalName(ctx context.Context, aliasName, aliasBrand string) (canonicalName, canonicalBrand string, err error) {
	return "", "", nil
}
func (m *mockStore) Close() error { return nil }

// Test JWT validation security patterns
func TestJWTSecurityValidation(t *testing.T) {
	// Save original env vars
	originalEnv := os.Getenv("ENV")
	originalIssuer := os.Getenv("SUPABASE_JWT_ISSUER")
	originalAudience := os.Getenv("SUPABASE_JWT_AUDIENCE")
	originalSecret := os.Getenv("SUPABASE_JWT_SECRET")

	defer func() {
		os.Setenv("ENV", originalEnv)
		os.Setenv("SUPABASE_JWT_ISSUER", originalIssuer)
		os.Setenv("SUPABASE_JWT_AUDIENCE", originalAudience)
		os.Setenv("SUPABASE_JWT_SECRET", originalSecret)
	}()

	t.Run("ProductionRequiresIssuerValidation", func(t *testing.T) {
		os.Setenv("ENV", "production")
		os.Setenv("SUPABASE_JWT_ISSUER", "")
		os.Setenv("SUPABASE_JWT_SECRET", "test_secret_that_is_longer_than_32_characters_for_security")

		claims := &SupabaseJWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "test-user-id",
				Issuer:    "test-issuer",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			Email: "test@example.com",
		}

		err := validateJWTClaims(claims)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SUPABASE_JWT_ISSUER must be configured in production")
	})

	t.Run("ProductionRequiresAudienceValidation", func(t *testing.T) {
		os.Setenv("ENV", "production")
		os.Setenv("SUPABASE_JWT_ISSUER", "test-issuer")
		os.Setenv("SUPABASE_JWT_AUDIENCE", "")
		os.Setenv("SUPABASE_JWT_SECRET", "test_secret_that_is_longer_than_32_characters_for_security")

		claims := &SupabaseJWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "test-user-id",
				Issuer:    "test-issuer",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			Email: "test@example.com",
		}

		err := validateJWTClaims(claims)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SUPABASE_JWT_AUDIENCE must be configured in production")
	})

	t.Run("ProductionRejectsShortJWTSecret", func(t *testing.T) {
		os.Setenv("ENV", "production")
		os.Setenv("SUPABASE_JWT_SECRET", "short_secret")

		store := &mockStore{}
		_, err := validateJWTAndGetUser(context.Background(), "invalid.jwt.token", store)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret must be at least 32 characters in production")
	})

	t.Run("ValidJWTClaimsPass", func(t *testing.T) {
		os.Setenv("ENV", "production")
		os.Setenv("SUPABASE_JWT_ISSUER", "test-issuer")
		os.Setenv("SUPABASE_JWT_AUDIENCE", "test-audience")

		claims := &SupabaseJWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "test-user-id",
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			Email: "test@example.com",
		}

		err := validateJWTClaims(claims)
		assert.NoError(t, err)
	})
}

// Test authentication middleware security patterns
func TestAuthMiddlewareSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original env vars
	originalEnv := os.Getenv("ENV")
	defer func() {
		os.Setenv("ENV", originalEnv)
	}()

	t.Run("ProductionRejectsRequestsWithoutAuth", func(t *testing.T) {
		os.Setenv("ENV", "production")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		store := &mockStore{}
		middleware := JWTAuthMiddleware(store)
		middleware(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Authorization header required", response["error"])
	})

	t.Run("RejectsInvalidAuthHeaderFormat", func(t *testing.T) {
		os.Setenv("ENV", "production")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat token")

		store := &mockStore{}
		middleware := JWTAuthMiddleware(store)
		middleware(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid authorization header format", response["error"])
	})

	t.Run("AllowsPublicEndpoints", func(t *testing.T) {
		os.Setenv("ENV", "production")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/health", nil)

		store := &mockStore{}
		middleware := JWTAuthMiddleware(store)

		// Set up a handler to verify the middleware passes through
		called := false
		testHandler := func(c *gin.Context) {
			called = true
		}

		// Create a router and add our test handler
		router := gin.New()
		router.GET("/api/v1/health", middleware, testHandler)

		// Make the request
		router.ServeHTTP(w, c.Request)

		assert.True(t, called, "Next handler should be called for public endpoints")
	})

	t.Run("DevelopmentModeSkipsAuthWithoutSecret", func(t *testing.T) {
		os.Setenv("ENV", "development")
		os.Setenv("SUPABASE_JWT_SECRET", "")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		store := &mockStore{}
		middleware := JWTAuthMiddleware(store)

		// Set up a handler to verify the middleware passes through
		called := false
		testHandler := func(c *gin.Context) {
			called = true
		}

		// Create a router and add our test handler
		router := gin.New()
		router.GET("/api/v1/test", middleware, testHandler)

		// Make the request
		router.ServeHTTP(w, c.Request)

		assert.True(t, called, "Next handler should be called in development without JWT secret")
	})
}

// Test email validation security
func TestEmailValidationSecurity(t *testing.T) {
	testCases := []struct {
		name     string
		email    string
		expected bool
	}{
		{"ValidEmail", "user@example.com", true},
		{"ValidEmailWithSubdomain", "user@mail.example.com", true},
		{"EmptyEmail", "", false},
		{"NoAtSymbol", "userexample.com", false},
		{"MultipleAtSymbols", "user@@example.com", false},
		{"NoLocalPart", "@example.com", false},
		{"NoDomainPart", "user@", false},
		{"NoTLD", "user@example", false},
		{"TooLong", strings.Repeat("a", 250) + "@example.com", false},
		{"LocalPartTooLong", strings.Repeat("a", 70) + "@example.com", false},
		{"WithSpaces", "user @example.com", false},
		{"WithControlChars", "user\n@example.com", false},
		{"WithHighASCII", "user@éxample.com", false},
		{"SQLInjectionAttempt", "'; DROP TABLE users; --@example.com", false},
		{"XSSAttempt", "<script>alert('xss')</script>@example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidEmail(tc.email)
			assert.Equal(t, tc.expected, result, "Email validation failed for: %s", tc.email)
		})
	}
}

// Test public endpoint detection
func TestPublicEndpointSecurity(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/health", true},
		{"/api/v1/docs", true},
		{"/api/v1/openapi.yaml", true},
		{"/api/v1/users", false},
		{"/api/v1/meals", false},
		{"/api/v1/health/extended", false}, // Should not match partial paths
		{"health", false},                  // Must be exact match
		{"/health", false},                 // Must include API prefix
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Path_%s", strings.ReplaceAll(tc.path, "/", "_")), func(t *testing.T) {
			result := isPublicEndpoint(tc.path)
			assert.Equal(t, tc.expected, result, "Public endpoint detection failed for: %s", tc.path)
		})
	}
}

// Test rate limiting for user creation
func TestUserCreationRateLimiting(t *testing.T) {
	email := "test@example.com"

	// Clean up any existing state
	userCreationMutex.Lock()
	delete(userCreationLimiter, email)
	userCreationMutex.Unlock()

	t.Run("AllowsInitialAttempts", func(t *testing.T) {
		// First attempt should be allowed
		assert.False(t, isUserCreationRateLimited(email))
		recordUserCreationAttempt(email)

		// Second attempt should be allowed
		assert.False(t, isUserCreationRateLimited(email))
		recordUserCreationAttempt(email)
	})

	t.Run("BlocksAfterLimit", func(t *testing.T) {
		// Third attempt should be blocked
		assert.True(t, isUserCreationRateLimited(email))
	})

	t.Run("CleansUpOldEntries", func(t *testing.T) {
		// Simulate old entries by directly modifying the limiter
		userCreationMutex.Lock()
		oldTime := time.Now().Add(-2 * time.Hour)
		userCreationLimiter[email] = []time.Time{oldTime, oldTime}
		userCreationMutex.Unlock()

		cleanupUserCreationLimiter()

		// After cleanup, new attempts should be allowed
		assert.False(t, isUserCreationRateLimited(email))
	})
}
