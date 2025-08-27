package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIServer_GetHealth(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create an in-memory SQLite store for testing
	config := &storage.Config{
		Type:     "sqlite",
		Database: ":memory:",
	}
	store, err := storage.NewStore(config)
	require.NoError(t, err)
	defer store.Close()

	apiServer, err := NewAPIServer(store)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/health", apiServer.GetHealth)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "health check returns 200",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Make request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response api.HealthResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "healthy", response.Status)
			assert.NotZero(t, response.Timestamp)
			assert.NotEmpty(t, response.Version)
		})
	}
}

func TestNewAPIServer(t *testing.T) {
	// Create an in-memory SQLite store for testing
	config := &storage.Config{
		Type:     "sqlite",
		Database: ":memory:",
	}
	store, err := storage.NewStore(config)
	require.NoError(t, err)
	defer store.Close()

	server, err := NewAPIServer(store)

	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.NotNil(t, server.store)
	assert.NotNil(t, server.goalResolver)
}

func TestCreateDatabaseConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupEnv   func()
		cleanupEnv func()
		expected   storage.Config
	}{
		{
			name: "default configuration",
			setupEnv: func() {
				clearDBEnvVars()
			},
			cleanupEnv: func() {},
			expected: storage.Config{
				Type:     "sqlite",
				Database: "./noot.db",
				Host:     "localhost",
				Port:     5432,
				Username: "",
				Password: "",
				SSLMode:  "prefer",
			},
		},
		{
			name: "postgres configuration",
			setupEnv: func() {
				clearDBEnvVars()
				os.Setenv("DATABASE_PROVIDER", "postgres")
				os.Setenv("DATABASE_PATH", "testdb")
				os.Setenv("DB_HOST", "testhost")
				os.Setenv("DB_PORT", "5433")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASS", "testpass")
				os.Setenv("DB_SSLMODE", "require")
			},
			cleanupEnv: func() {
				clearDBEnvVars()
			},
			expected: storage.Config{
				Type:     "postgres",
				Database: "testdb",
				Host:     "testhost",
				Port:     5433,
				Username: "testuser",
				Password: "testpass",
				SSLMode:  "require",
			},
		},
		{
			name: "supabase configuration with URL override",
			setupEnv: func() {
				clearDBEnvVars()
				os.Setenv("DATABASE_PROVIDER", "supabase")
				os.Setenv("DATABASE_PATH", "default_db")
				os.Setenv("SUPABASE_DB_URL", "postgresql://user:pass@host:port/db")
			},
			cleanupEnv: func() {
				clearDBEnvVars()
			},
			expected: storage.Config{
				Type:     "supabase",
				Database: "postgresql://user:pass@host:port/db", // Should be overridden by SUPABASE_DB_URL
				Host:     "localhost",
				Port:     5432,
				Username: "",
				Password: "",
				SSLMode:  "prefer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()
			defer tt.cleanupEnv()

			config := CreateDatabaseConfig()

			assert.Equal(t, tt.expected.Type, config.Type)
			assert.Equal(t, tt.expected.Database, config.Database)
			assert.Equal(t, tt.expected.Host, config.Host)
			assert.Equal(t, tt.expected.Port, config.Port)
			assert.Equal(t, tt.expected.Username, config.Username)
			assert.Equal(t, tt.expected.Password, config.Password)
			assert.Equal(t, tt.expected.SSLMode, config.SSLMode)
		})
	}
}

// Helper function to clear database environment variables
func clearDBEnvVars() {
	os.Unsetenv("DATABASE_PROVIDER")
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASS")
	os.Unsetenv("DB_SSLMODE")
	os.Unsetenv("SUPABASE_DB_URL")
}
