package server

import (
	"os"
	"testing"
)

func TestIsProduction(t *testing.T) {
	// Save original ENV value to restore later
	originalEnv := os.Getenv("ENV")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "ENV=production (lowercase)",
			envValue: "production",
			expected: true,
		},
		{
			name:     "ENV=PRODUCTION (uppercase)",
			envValue: "PRODUCTION",
			expected: true,
		},
		{
			name:     "ENV=Production (mixed case)",
			envValue: "Production",
			expected: true,
		},
		{
			name:     "ENV= production (with spaces)",
			envValue: " production ",
			expected: true,
		},
		{
			name:     "ENV=development",
			envValue: "development",
			expected: false,
		},
		{
			name:     "ENV=staging",
			envValue: "staging",
			expected: false,
		},
		{
			name:     "ENV=test",
			envValue: "test",
			expected: false,
		},
		{
			name:     "ENV=local",
			envValue: "local",
			expected: false,
		},
		{
			name:     "ENV unset (defaults to production for safety)",
			envValue: "",
			expected: true,
		},
		{
			name:     "ENV=empty_string",
			envValue: "",
			expected: true,
		},
		{
			name:     "ENV=random_value",
			envValue: "random_value",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up ENV
			if tt.envValue == "" {
				os.Unsetenv("ENV")
			} else {
				os.Setenv("ENV", tt.envValue)
			}

			// Test the function
			result := IsProduction()
			if result != tt.expected {
				t.Errorf("IsProduction() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
