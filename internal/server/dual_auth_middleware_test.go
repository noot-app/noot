package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeyStore extends NullStore with configurable API key behaviors for testing
type TestAPIKeyStore struct {
	*NullStore
	apiKeys map[string]*storage.APIKey
	users   map[string]*storage.User
}

func NewTestAPIKeyStore() *TestAPIKeyStore {
	return &TestAPIKeyStore{
		NullStore: &NullStore{},
		apiKeys:   make(map[string]*storage.APIKey),
		users:     make(map[string]*storage.User),
	}
}

func (t *TestAPIKeyStore) SetAPIKey(prefix string, apiKey *storage.APIKey) {
	t.apiKeys[prefix] = apiKey
}

func (t *TestAPIKeyStore) SetUser(id string, user *storage.User) {
	t.users[id] = user
}

func (t *TestAPIKeyStore) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*storage.APIKey, error) {
	apiKey, exists := t.apiKeys[prefix]
	if !exists {
		return nil, nil
	}
	return apiKey, nil
}

func (t *TestAPIKeyStore) GetUser(ctx context.Context, id string) (*storage.User, error) {
	user, exists := t.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (t *TestAPIKeyStore) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	// No-op for testing
	return nil
}

func TestDualAuthMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Initialize logger for tests
	InitLogger()

	tests := []struct {
		name           string
		method         string
		path           string
		setupStore     func(*TestAPIKeyStore) string // returns API key to use
		setupHeaders   func(req *http.Request, apiKey string)
		expectedStatus int
		expectAuth     bool
	}{
		{
			name:           "OPTIONS request bypasses auth",
			method:         http.MethodOptions,
			path:           "/api/v1/test",
			setupStore:     func(store *TestAPIKeyStore) string { return "" },
			setupHeaders:   func(req *http.Request, apiKey string) {},
			expectedStatus: http.StatusOK,
			expectAuth:     false,
		},
		{
			name:           "Public endpoint bypasses auth",
			method:         http.MethodGet,
			path:           "/api/v1/health",
			setupStore:     func(store *TestAPIKeyStore) string { return "" },
			setupHeaders:   func(req *http.Request, apiKey string) {},
			expectedStatus: http.StatusOK,
			expectAuth:     false,
		},
		{
			name:   "Valid API key authenticates user",
			method: http.MethodPost,
			path:   "/api/v1/consumption",
			setupStore: func(store *TestAPIKeyStore) string {
				fullKey, prefix, hash, err := storage.GenerateAPIKey()
				require.NoError(t, err)

				apiKey := &storage.APIKey{
					ID:     "key-1",
					UserID: "user-1",
					Hash:   hash,
					Scope:  "read_write",
				}

				user := &storage.User{
					ID:               "user-1",
					Email:            "test@example.com",
					SubscriptionTier: storage.SubscriptionTierPro,
				}

				store.SetAPIKey(prefix, apiKey)
				store.SetUser("user-1", user)
				return fullKey
			},
			setupHeaders: func(req *http.Request, apiKey string) {
				req.Header.Set("X-API-Key", apiKey)
			},
			expectedStatus: http.StatusOK,
			expectAuth:     true,
		},
		{
			name:   "Invalid API key fails authentication",
			method: http.MethodPost,
			path:   "/api/v1/consumption",
			setupStore: func(store *TestAPIKeyStore) string {
				fullKey, _, _, err := storage.GenerateAPIKey()
				require.NoError(t, err)
				// Don't add to store
				return fullKey
			},
			setupHeaders: func(req *http.Request, apiKey string) {
				req.Header.Set("X-API-Key", apiKey)
			},
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewTestAPIKeyStore()
			apiKey := tt.setupStore(store)

			// Create router with middleware
			router := gin.New()
			router.Use(DualAuthMiddleware(store))
			router.Handle(tt.method, tt.path, func(c *gin.Context) {
				if tt.expectAuth {
					authUser, exists := c.Get("auth_user")
					assert.True(t, exists, "Auth user should be set")
					assert.NotNil(t, authUser, "Auth user should not be nil")
				}
				c.Status(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			tt.setupHeaders(req, apiKey)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCheckAPIKeyScope(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		scope       string
		shouldError bool
	}{
		{
			name:        "GET request with read scope",
			method:      "GET",
			scope:       "read",
			shouldError: false,
		},
		{
			name:        "GET request with read_write scope",
			method:      "GET",
			scope:       "read_write",
			shouldError: false,
		},
		{
			name:        "POST request with read_write scope",
			method:      "POST",
			scope:       "read_write",
			shouldError: false,
		},
		{
			name:        "POST request with read scope should fail",
			method:      "POST",
			scope:       "read",
			shouldError: true,
		},
		{
			name:        "PUT request with read scope should fail",
			method:      "PUT",
			scope:       "read",
			shouldError: true,
		},
		{
			name:        "DELETE request with read scope should fail",
			method:      "DELETE",
			scope:       "read",
			shouldError: true,
		},
		{
			name:        "PATCH request with read_write scope",
			method:      "PATCH",
			scope:       "read_write",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkAPIKeyScope(tt.method, tt.scope)
			if tt.shouldError {
				assert.Error(t, err, "Should return error for insufficient scope")
				assert.IsType(t, &APIKeyError{}, err, "Should return APIKeyError")
			} else {
				assert.NoError(t, err, "Should not return error for sufficient scope")
			}
		})
	}
}

func TestAPIKeyError(t *testing.T) {
	err := &APIKeyError{
		Type:    "test_error",
		Message: "Test error message",
	}

	assert.Equal(t, "Test error message", err.Error())
}

func TestGetAuthenticatedMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setMethod      string
		expectedMethod string
	}{
		{
			name:           "API key authentication",
			setMethod:      "api_key",
			expectedMethod: "api_key",
		},
		{
			name:           "JWT authentication",
			setMethod:      "jwt",
			expectedMethod: "jwt",
		},
		{
			name:           "No authentication method set defaults to jwt",
			setMethod:      "",
			expectedMethod: "jwt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.setMethod != "" {
				c.Set("auth_method", tt.setMethod)
			}

			method := GetAuthenticatedMethod(c)
			assert.Equal(t, tt.expectedMethod, method)
		})
	}
}

func TestGetAPIKeyScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with no scope set
	scope := GetAPIKeyScope(c)
	assert.Equal(t, "", scope)

	// Test with scope set
	c.Set("api_key_scope", "read_write")
	scope = GetAPIKeyScope(c)
	assert.Equal(t, "read_write", scope)
}

func TestGetAPIKeyID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with no ID set
	id := GetAPIKeyID(c)
	assert.Equal(t, "", id)

	// Test with ID set
	c.Set("api_key_id", "key-123")
	id = GetAPIKeyID(c)
	assert.Equal(t, "key-123", id)
}

func TestIsConsumptionEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/consumption/123", true},   // exactly 4 segments, starts with /api/v1/consumption/
		{"/api/v1/consumption", false},      // only 3 segments
		{"/api/v1/consumptions", false},     // different path
		{"/api/v1/consumptions/456", false}, // different path
		{"/api/v1/health", false},
		{"/api/v1/goals", false},
		{"/other/consumption", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isConsumptionEndpoint(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
