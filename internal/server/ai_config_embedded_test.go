package server

import (
	"os"
	"testing"
)

func TestEmbeddedAIRequestBuilder(t *testing.T) {
	tests := []struct {
		name   string
		aiType string
	}{
		{
			name:   "GetNutrition embedded config",
			aiType: "GetNutrition",
		},
		{
			name:   "ParseItems embedded config",
			aiType: "ParseItems",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewEmbeddedAIRequestBuilder(tt.aiType)
			if err != nil {
				t.Fatalf("Failed to create embedded AI request builder: %v", err)
			}

			if builder == nil {
				t.Fatal("Builder should not be nil")
			}

			// Test building a request
			payload, err := builder.BuildRequestPayload("test input")
			if err != nil {
				t.Fatalf("Failed to build request payload: %v", err)
			}

			if payload == nil {
				t.Fatal("Payload should not be nil")
			}

			if payload.Model == "" {
				t.Error("Model should not be empty")
			}

			if len(payload.Input) == 0 {
				t.Error("Input should not be empty")
			}
		})
	}
}

func TestEmbeddedConfigFallback(t *testing.T) {
	// Test that the builders can initialize with embedded configs
	nutritionBuilder, err := GetNutritionBuilder()
	if err != nil {
		t.Fatalf("Failed to get nutrition builder: %v", err)
	}
	if nutritionBuilder == nil {
		t.Fatal("Nutrition builder should not be nil")
	}

	parseItemsBuilder, err := ParseItemsBuilder()
	if err != nil {
		t.Fatalf("Failed to get parse items builder: %v", err)
	}
	if parseItemsBuilder == nil {
		t.Fatal("Parse items builder should not be nil")
	}
}

func TestEnvironmentVariableExpansion(t *testing.T) {
	// Set a test environment variable
	testToken := "test-token-12345"
	os.Setenv("OPENFOODFACTS_MCP_TOKEN", testToken)
	defer os.Unsetenv("OPENFOODFACTS_MCP_TOKEN")

	// Create a builder for GetNutrition which uses the MCP token
	builder, err := NewEmbeddedAIRequestBuilder("GetNutrition")
	if err != nil {
		t.Fatalf("Failed to create embedded AI request builder: %v", err)
	}

	// Build a request to check if the config is loaded with expanded env vars
	payload, err := builder.BuildRequestPayload("test input")
	if err != nil {
		t.Fatalf("Failed to build request payload: %v", err)
	}

	// Check if the authorization field in tools contains the expanded token
	found := false
	for _, tool := range payload.Tools {
		if tool.Type == "mcp" && tool.Authorization == testToken {
			found = true
			break
		}
	}

	if !found {
		t.Error("Environment variable OPENFOODFACTS_MCP_TOKEN was not properly expanded in embedded config")
	}
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "Valid environment variable",
			input:    "${TEST_VAR}",
			envVar:   "TEST_VAR",
			envValue: "test-value",
			expected: "test-value",
		},
		{
			name:     "Non-existent environment variable",
			input:    "${NON_EXISTENT}",
			envVar:   "",
			envValue: "",
			expected: "${NON_EXISTENT}", // Should return original
		},
		{
			name:     "Not an environment variable format",
			input:    "regular-string",
			envVar:   "",
			envValue: "",
			expected: "regular-string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable if needed
			if tt.envVar != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
