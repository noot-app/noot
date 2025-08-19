package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitLogger(t *testing.T) {
	// Test that InitLogger doesn't panic
	assert.NotPanics(t, func() {
		InitLogger()
	})

	// Test that logger is set after initialization
	assert.NotNil(t, logger)
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected string
	}{
		{"DEBUG", "DEBUG", "DEBUG"},
		{"INFO", "INFO", "INFO"},
		{"WARN", "WARN", "WARN"},
		{"ERROR", "ERROR", "ERROR"},
		{"invalid", "INVALID", "INFO"},  // defaults to INFO
		{"empty", "", "INFO"},           // defaults to INFO
		{"lowercase", "debug", "DEBUG"}, // should be uppercase
	}

	originalLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		if originalLevel == "" {
			os.Unsetenv("LOG_LEVEL")
		} else {
			os.Setenv("LOG_LEVEL", originalLevel)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.logLevel == "" {
				os.Unsetenv("LOG_LEVEL")
			} else {
				os.Setenv("LOG_LEVEL", tt.logLevel)
			}

			level := getLogLevel()
			assert.Equal(t, tt.expected, strings.ToUpper(level.String()))
		})
	}
}

func TestIsDebugMode(t *testing.T) {
	tests := []struct {
		name     string
		debug    string
		expected bool
	}{
		{"true", "true", true},
		{"True", "True", true},
		{"TRUE", "TRUE", true},
		{"false", "false", false},
		{"False", "False", false},
		{"FALSE", "FALSE", false},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"1", "1", true},
		{"0", "0", false},
	}

	originalDebug := os.Getenv("DEBUG")
	defer func() {
		if originalDebug == "" {
			os.Unsetenv("DEBUG")
		} else {
			os.Setenv("DEBUG", originalDebug)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.debug == "" {
				os.Unsetenv("DEBUG")
			} else {
				os.Setenv("DEBUG", tt.debug)
			}

			result := isDebugMode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDevMode(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{"dev", "dev", true},
		{"development", "development", true},
		{"local", "local", true},
		{"Dev", "Dev", true},                 // case insensitive
		{"DEVELOPMENT", "DEVELOPMENT", true}, // case insensitive
		{"production", "production", false},
		{"prod", "prod", false},
		{"staging", "staging", false},
		{"empty", "", false}, // defaults to production
	}

	originalEnv := os.Getenv("ENV")
	originalEnvironment := os.Getenv("ENVIRONMENT")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
		if originalEnvironment == "" {
			os.Unsetenv("ENVIRONMENT")
		} else {
			os.Setenv("ENVIRONMENT", originalEnvironment)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear both ENV vars first
			os.Unsetenv("ENV")
			os.Unsetenv("ENVIRONMENT")

			if tt.env != "" {
				os.Setenv("ENV", tt.env)
			}

			result := isDevMode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDevMode_Environment(t *testing.T) {
	// Test ENVIRONMENT variable when ENV is not set
	originalEnv := os.Getenv("ENV")
	originalEnvironment := os.Getenv("ENVIRONMENT")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
		if originalEnvironment == "" {
			os.Unsetenv("ENVIRONMENT")
		} else {
			os.Setenv("ENVIRONMENT", originalEnvironment)
		}
	}()

	os.Unsetenv("ENV")
	os.Setenv("ENVIRONMENT", "development")

	result := isDevMode()
	assert.True(t, result)
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		err      error
		expected string
	}{
		{
			name:     "error with cause",
			message:  "test message",
			err:      errors.New("cause error"),
			expected: "test message: cause error",
		},
		{
			name:     "error without cause",
			message:  "test message only",
			err:      nil,
			expected: "test message only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := &AppError{
				Message: tt.message,
				Err:     tt.err,
			}

			result := appErr.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := &AppError{
		Message: "wrapped error",
		Err:     originalErr,
	}

	unwrapped := appErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	appErr := &AppError{
		Message: "no cause",
		Err:     nil,
	}

	unwrapped := appErr.Unwrap()
	assert.Nil(t, unwrapped)
}

func TestNewAppError(t *testing.T) {
	originalDebug := os.Getenv("DEBUG")
	originalEnv := os.Getenv("ENV")
	defer func() {
		if originalDebug == "" {
			os.Unsetenv("DEBUG")
		} else {
			os.Setenv("DEBUG", originalDebug)
		}
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	t.Run("debug mode includes stack", func(t *testing.T) {
		os.Setenv("DEBUG", "true")
		os.Unsetenv("ENV")

		err := errors.New("test error")
		appErr := NewAppError("test message", 500, err)

		assert.Equal(t, "test message", appErr.Message)
		assert.Equal(t, err, appErr.Err)
		assert.Equal(t, 500, appErr.StatusCode)
		assert.NotEmpty(t, appErr.Stack)
	})

	t.Run("dev mode includes stack", func(t *testing.T) {
		os.Unsetenv("DEBUG")
		os.Setenv("ENV", "development")

		err := errors.New("test error")
		appErr := NewAppError("test message", 500, err)

		assert.Equal(t, "test message", appErr.Message)
		assert.Equal(t, err, appErr.Err)
		assert.Equal(t, 500, appErr.StatusCode)
		assert.NotEmpty(t, appErr.Stack)
	})

	t.Run("production mode no stack", func(t *testing.T) {
		os.Unsetenv("DEBUG")
		os.Setenv("ENV", "production")

		err := errors.New("test error")
		appErr := NewAppError("test message", 500, err)

		assert.Equal(t, "test message", appErr.Message)
		assert.Equal(t, err, appErr.Err)
		assert.Equal(t, 500, appErr.StatusCode)
		assert.Empty(t, appErr.Stack)
	})
}

func TestCaptureStack(t *testing.T) {
	stack := captureStack(1)
	assert.NotEmpty(t, stack)

	// Should contain this test function
	found := false
	for _, frame := range stack {
		if strings.Contains(frame, "TestCaptureStack") {
			found = true
			break
		}
	}
	assert.True(t, found, "Stack should contain the test function")
}

func TestLogHelpers(t *testing.T) {
	// Initialize logger
	InitLogger()

	// These tests mainly ensure the functions don't panic
	// In a real test environment, you might capture log output

	t.Run("LogInfo", func(t *testing.T) {
		assert.NotPanics(t, func() {
			LogInfo("test info", "key", "value")
		})
	})

	t.Run("LogError", func(t *testing.T) {
		assert.NotPanics(t, func() {
			LogError("test error", errors.New("test"), "key", "value")
		})
	})

	t.Run("LogWarn", func(t *testing.T) {
		assert.NotPanics(t, func() {
			LogWarn("test warning", "key", "value")
		})
	})

	t.Run("LogDebug", func(t *testing.T) {
		assert.NotPanics(t, func() {
			LogDebug("test debug", "key", "value")
		})
	})
}

func TestHandleAppError(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Set debug mode to test stack inclusion
	originalDebug := os.Getenv("DEBUG")
	os.Setenv("DEBUG", "true")
	defer func() {
		if originalDebug == "" {
			os.Unsetenv("DEBUG")
		} else {
			os.Setenv("DEBUG", originalDebug)
		}
	}()

	w := httptest.NewRecorder()
	appErr := NewAppError("test app error", 422, fmt.Errorf("underlying error"))
	traceID := "trace456"

	handleAppError(w, appErr, traceID)

	assert.Equal(t, 422, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test app error", response["error"])
	assert.Equal(t, float64(422), response["code"])
	assert.Equal(t, "trace456", response["trace_id"])
}
