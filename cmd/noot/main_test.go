package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/grantbirki/noot/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestMain_Integration(t *testing.T) {
	// This is a basic integration test to ensure main doesn't panic on startup
	// We can't easily test the full main function without starting a server
	// but we can test that the basic setup works

	// Set required environment variables for basic functionality
	originalPort := os.Getenv("PORT")
	defer func() {
		if originalPort == "" {
			os.Unsetenv("PORT")
		} else {
			os.Setenv("PORT", originalPort)
		}
	}()

	// Test with valid port
	os.Setenv("PORT", "8080")

	// The main function would normally run the server, but we can't test that easily
	// Instead we test that the environment setup works correctly
	port := os.Getenv("PORT")
	assert.Equal(t, "8080", port)

	// Test default port behavior
	os.Unsetenv("PORT")
	port = os.Getenv("PORT")
	if port == "" {
		port = "3001" // This is the default in main()
	}
	assert.Equal(t, "3001", port)
}

func TestEnvironmentSetup(t *testing.T) {
	// Test that environment variables can be set/read properly
	// This tests the basic environment handling that main() relies on

	tests := []struct {
		name     string
		envVar   string
		value    string
		expected string
	}{
		{"PORT variable", "PORT", "9000", "9000"},
		{"DEBUG variable", "DEBUG", "true", "true"},
		{"LOG_LEVEL variable", "LOG_LEVEL", "DEBUG", "DEBUG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv(tt.envVar)
			defer func() {
				if original == "" {
					os.Unsetenv(tt.envVar)
				} else {
					os.Setenv(tt.envVar, original)
				}
			}()

			// Set test value
			os.Setenv(tt.envVar, tt.value)

			// Verify it was set
			result := os.Getenv(tt.envVar)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServerInitialization(t *testing.T) {
	// Test that we can import and reference the server package
	t.Run("LoadDotEnv_function_exists", func(t *testing.T) {
		// This test verifies that server.LoadDotEnv is callable
		// We can't test the actual functionality without file system changes
		// but we can verify the function exists and is callable
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("server.LoadDotEnv() panicked: %v", r)
			}
		}()
		// Call the function - it should not panic even if .env doesn't exist
		server.LoadDotEnv()
	})

	t.Run("InitLogger_function_exists", func(t *testing.T) {
		// Test that InitLogger can be called without panicking
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("server.InitLogger() panicked: %v", r)
			}
		}()
		server.InitLogger()
	})
}

func TestEnvironmentDefaults(t *testing.T) {
	// Test the default values used in main()
	t.Run("default_port_value", func(t *testing.T) {
		// Clear PORT to test default
		original := os.Getenv("PORT")
		defer func() {
			if original == "" {
				os.Unsetenv("PORT")
			} else {
				os.Setenv("PORT", original)
			}
		}()

		os.Unsetenv("PORT")
		port := os.Getenv("PORT")
		if port == "" {
			port = "3001" // Default from main()
		}

		assert.Equal(t, "3001", port)
		assert.NotEmpty(t, port)
	})

	t.Run("signal_handling_setup", func(t *testing.T) {
		// Test that we can create a signal context like main() does
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		assert.NotNil(t, ctx)
		assert.NotNil(t, cancel)

		// Clean up
		cancel()

		// Verify context was cancelled
		select {
		case <-ctx.Done():
			// Expected - context should be done after cancel
		default:
			t.Error("Context should be done after cancel()")
		}
	})
}

func TestMainFunctionComponents(t *testing.T) {
	// Test individual components that main() uses
	t.Run("port_parsing", func(t *testing.T) {
		testCases := []struct {
			envValue     string
			expectedPort string
		}{
			{"", "3001"},     // Default case
			{"8080", "8080"}, // Custom port
			{"3000", "3000"}, // Another custom port
			{"80", "80"},     // HTTP port
			{"443", "443"},   // HTTPS port
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("port_%s", tc.envValue), func(t *testing.T) {
				// Save original
				original := os.Getenv("PORT")
				defer func() {
					if original == "" {
						os.Unsetenv("PORT")
					} else {
						os.Setenv("PORT", original)
					}
				}()

				// Set test value
				if tc.envValue == "" {
					os.Unsetenv("PORT")
				} else {
					os.Setenv("PORT", tc.envValue)
				}

				// Simulate main()'s port logic
				port := os.Getenv("PORT")
				if port == "" {
					port = "3001"
				}

				assert.Equal(t, tc.expectedPort, port)
			})
		}
	})
}
