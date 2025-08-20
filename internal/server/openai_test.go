package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestFloat64Ptr(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "positive number",
			input:    123.45,
			expected: 123.45,
		},
		{
			name:     "negative number",
			input:    -67.89,
			expected: -67.89,
		},
		{
			name:     "zero",
			input:    0.0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := float64Ptr(tt.input)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// Helper function to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}
