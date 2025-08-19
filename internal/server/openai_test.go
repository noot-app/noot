package server

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranscriptionPrompt(t *testing.T) {
	prompt := transcriptionPrompt()

	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "audio")
	assert.Contains(t, prompt, "foods")
	assert.Contains(t, prompt, "drinks")
	// Should contain guidance about preserving brand names
	assert.Contains(t, prompt, "brand")
}

func TestParseItemsSystemPrompt(t *testing.T) {
	prompt := parseItemsSystemPrompt()

	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "JSON")
	assert.Contains(t, prompt, "items")
	// Should contain structure definition
	assert.Contains(t, prompt, "name")
	assert.Contains(t, prompt, "quantity")
	assert.Contains(t, prompt, "unit")
	assert.Contains(t, prompt, "brand")
	// Should contain serving size instructions
	assert.Contains(t, prompt, "serving")
}

func TestNutritionSystemPrompt(t *testing.T) {
	prompt := nutritionSystemPrompt()

	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "JSON")
	assert.Contains(t, prompt, "nutrition")
	assert.Contains(t, prompt, "nutrients")
	// Should contain structure definition
	assert.Contains(t, prompt, "calories")
	assert.Contains(t, prompt, "protein_g")
}

func TestStrPtrOrNil(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected *string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty string",
			input:    stringPtr(""),
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    stringPtr("   "),
			expected: nil,
		},
		{
			name:     "valid string",
			input:    stringPtr("test"),
			expected: stringPtr("test"),
		},
		{
			name:     "string with whitespace",
			input:    stringPtr("  test  "),
			expected: stringPtr("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strPtrOrNil(tt.input)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestGetNutritionFromOpenAI_MissingAPIKey(t *testing.T) {
	// Save current env var
	original := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if original == "" {
			os.Unsetenv("OPENAI_API_KEY")
		} else {
			os.Setenv("OPENAI_API_KEY", original)
		}
	}()

	// Unset API key
	os.Unsetenv("OPENAI_API_KEY")

	ctx := context.Background()
	item := Item{Name: "Apple", Quantity: float64Ptr(1), Unit: stringPtr("medium")}
	_, err := getNutritionFromOpenAI(ctx, item)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY not configured")
}

// Helper function to create float64 pointers for tests
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}
