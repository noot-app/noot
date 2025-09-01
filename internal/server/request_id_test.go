package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRequestID(t *testing.T) {
	t.Run("generates_unique_ids", func(t *testing.T) {
		id1 := generateRequestID()
		id2 := generateRequestID()

		// Should be different
		if id1 == id2 {
			t.Errorf("generateRequestID() should generate unique IDs, got same ID twice: %s", id1)
		}

		// Should be hex encoded (32 chars for 16 bytes)
		if len(id1) != 32 {
			t.Errorf("generateRequestID() should return 32 character hex string, got %d characters", len(id1))
		}

		// Should only contain hex characters
		for _, char := range id1 {
			if !strings.ContainsRune("0123456789abcdef", char) {
				t.Errorf("generateRequestID() should return hex string, found invalid character: %c", char)
			}
		}
	})

	t.Run("generates_multiple_unique_ids", func(t *testing.T) {
		ids := make(map[string]bool)

		// Generate 100 IDs and ensure they're all unique
		for i := 0; i < 100; i++ {
			id := generateRequestID()
			if ids[id] {
				t.Errorf("generateRequestID() generated duplicate ID: %s", id)
			}
			ids[id] = true
		}

		if len(ids) != 100 {
			t.Errorf("Expected 100 unique IDs, got %d", len(ids))
		}
	})
}

func TestRequestIDAlwaysIncluded(t *testing.T) {
	tests := []struct {
		name        string
		envSettings map[string]string
		expectStack bool
	}{
		{
			name: "Production environment",
			envSettings: map[string]string{
				"ENV":   "production",
				"DEBUG": "false",
			},
			expectStack: false,
		},
		{
			name: "Non-dev environment (staging)",
			envSettings: map[string]string{
				"ENV":   "staging",
				"DEBUG": "false",
			},
			expectStack: false,
		},
		{
			name: "Development environment",
			envSettings: map[string]string{
				"ENV":   "development",
				"DEBUG": "false",
			},
			expectStack: true, // isDevMode() returns true for "development"
		},
		{
			name: "Debug environment",
			envSettings: map[string]string{
				"ENV":   "development", // Need non-production ENV for debug to work
				"DEBUG": "true",
			},
			expectStack: true, // isDebugMode() returns true when DEBUG=true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envSettings {
				oldValue := os.Getenv(key)
				defer func(k, v string) {
					if v == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, v)
					}
				}(key, oldValue)
				os.Setenv(key, value)
			}

			// Test Recovery Middleware
			testRecoveryMiddleware(t, tt.expectStack)

			// Test App Error Handling via handleAppErrorGin
			testAppErrorHandling(t)

			// Test httpErrorWithDetails
			testHttpErrorWithDetails(t, tt.expectStack)
		})
	}
}

func testRecoveryMiddleware(t *testing.T, expectStack bool) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(RecoveryMiddleware())

	// Add a route that panics
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic for recovery")
	})

	// Make request
	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, 500, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// request_id should ALWAYS be present
	requestID, exists := response["request_id"]
	assert.True(t, exists, "request_id should always be present in recovery middleware response")
	assert.NotEmpty(t, requestID, "request_id should not be empty")

	// Stack should only be present in debug/dev mode
	_, hasStack := response["details"]
	if expectStack {
		assert.True(t, hasStack, "Stack details should be present in debug mode")
	} else {
		assert.False(t, hasStack, "Stack details should not be present in production/non-debug mode")
	}
}

func testAppErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	// Add a route that triggers an app error
	router.GET("/error", func(c *gin.Context) {
		requestID := getRequestID(c)
		appErr := NewAppError("Test application error", http.StatusBadRequest, fmt.Errorf("underlying error"))
		handleAppErrorGin(c, appErr, requestID)
	})

	// Make request
	req, _ := http.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, 400, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// trace_id (which is the request_id) should ALWAYS be present
	traceID, exists := response["trace_id"]
	assert.True(t, exists, "trace_id should always be present in app error response")
	assert.NotEmpty(t, traceID, "trace_id should not be empty")
}

func testHttpErrorWithDetails(t *testing.T, expectStack bool) {
	InitLogger() // Initialize logger for the test

	w := httptest.NewRecorder()
	stack := []string{"line1", "line2"}
	traceID := "test-trace-12345"

	httpErrorWithDetails(w, http.StatusInternalServerError, "detailed error", stack, traceID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "detailed error", response["error"])
	assert.Equal(t, float64(500), response["code"])

	// trace_id should ALWAYS be present when provided
	assert.Equal(t, "test-trace-12345", response["trace_id"], "trace_id should always be present when provided")
	assert.NotNil(t, response["timestamp"])

	// Check if stack is included based on environment
	if stackResponse, hasStack := response["stack"]; hasStack && stackResponse != nil {
		if expectStack {
			stackSlice := stackResponse.([]interface{})
			assert.Len(t, stackSlice, 2)
			assert.Equal(t, "line1", stackSlice[0])
			assert.Equal(t, "line2", stackSlice[1])
		} else {
			t.Error("Stack should not be present in production environment")
		}
	} else if expectStack {
		t.Error("Stack should be present in debug environment")
	}
}
