package server

import (
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

func TestParseSystemPrompt(t *testing.T) {
	prompt := parseSystemPrompt()
	
	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "JSON")
	assert.Contains(t, prompt, "nutrition")
	assert.Contains(t, prompt, "items")
	// Should contain structure definition
	assert.Contains(t, prompt, "calories")
	assert.Contains(t, prompt, "protein_g")
	assert.Contains(t, prompt, "nutrients")
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

// Helper function to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}