package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBrandTag(t *testing.T) {
	client := &OFFClient{}

	tests := []struct {
		name     string
		brand    string
		expected string
	}{
		{
			name:     "Simple brand",
			brand:    "Olipop",
			expected: "olipop",
		},
		{
			name:     "Brand with space",
			brand:    "Clif Bar",
			expected: "clif-bar",
		},
		{
			name:     "Brand with apostrophe",
			brand:    "Kellogg's",
			expected: "kelloggs",
		},
		{
			name:     "Brand with ampersand and spaces",
			brand:    "Ben & Jerry's",
			expected: "ben-jerrys",
		},
		{
			name:     "Brand with accent",
			brand:    "Nestlé Pure Life",
			expected: "nestle-pure-life",
		},
		{
			name:     "Complex brand with multiple special chars",
			brand:    "H&M® Fashion Co.",
			expected: "hm-fashion-co",
		},
		{
			name:     "Brand with parentheses",
			brand:    "Coca-Cola (Original)",
			expected: "coca-cola-original",
		},
		{
			name:     "Brand with forward slash",
			brand:    "Arm/Hammer",
			expected: "arm-hammer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.formatBrandTag(tt.brand)
			assert.Equal(t, tt.expected, result, "Brand tag formatting for '%s' should be '%s' but got '%s'", tt.brand, tt.expected, result)
		})
	}
}
