package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealth(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func()
		cleanupEnv     func()
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "health_check_with_default_version",
			setupEnv: func() {
				os.Unsetenv("VERSION")
			},
			cleanupEnv:     func() {},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "timestamp", "version"},
		},
		{
			name: "health_check_with_custom_version",
			setupEnv: func() {
				os.Setenv("VERSION", "v1.2.3")
			},
			cleanupEnv: func() {
				os.Unsetenv("VERSION")
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "timestamp", "version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer tt.cleanupEnv()

			// Create test server - we can use a nil store for health check as it doesn't use storage
			gin.SetMode(gin.TestMode)
			router := gin.New()

			server := &APIServer{} // Health endpoint doesn't need store or goalResolver
			router.GET("/health", server.GetHealth)

			// Create request
			req, err := http.NewRequest(http.MethodGet, "/health", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response api.HealthResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check required fields are present
			assert.Equal(t, "healthy", response.Status)
			assert.NotZero(t, response.Timestamp)

			if os.Getenv("VERSION") != "" {
				assert.Equal(t, os.Getenv("VERSION"), response.Version)
			} else {
				assert.Equal(t, "dev", response.Version)
			}

			// Verify timestamp is recent (within last 5 seconds)
			timeDiff := time.Since(response.Timestamp)
			assert.True(t, timeDiff < 5*time.Second, "Timestamp should be recent")
		})
	}
}

// TestExportData is commented out for now as it requires complex mocking
// This would be better as an integration test with a test database
// func TestExportData(t *testing.T) {
//   // Complex handler requiring database, auth, and business logic
//   // Better to focus on unit testing individual components first
// }
