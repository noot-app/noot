package server

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeTranscriptOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "I ate an apple and banana",
			expected: "I ate an apple and banana",
		},
		{
			name:     "text with whitespace",
			input:    "  \n\tI ate an apple\r\n  ",
			expected: "I ate an apple",
		},
		{
			name:     "text with control characters",
			input:    "I ate\x00an\x01apple\x1F",
			expected: "I ateanapple",
		},
		{
			name:     "oversized text",
			input:    strings.Repeat("a", maxTranscriptLength+100),
			expected: strings.Repeat("a", maxTranscriptLength),
		},
		{
			name:     "empty text",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace and control chars",
			input:    "\x00\x01\t\n\r ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTranscriptOutput(tt.input)
			assert.Equal(t, tt.expected, result)

			// Additional security checks
			assert.LessOrEqual(t, len(result), maxTranscriptLength, "Result should not exceed max length")

			// Check for control characters (except allowed whitespace)
			for _, r := range result {
				if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' && r != ' ' {
					t.Errorf("Found unexpected control character in result: %U", r)
				}
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "normal text",
			input:    "Organic Apple",
			expected: "Organic Apple",
			desc:     "Normal text should pass through unchanged",
		},
		{
			name:     "text with control characters",
			input:    "Apple\x00\x01Brand",
			expected: "AppleBrand",
			desc:     "Control characters should be removed",
		},
		{
			name:     "potential SQL injection",
			input:    "Apple'; DROP TABLE users; --",
			expected: "Apple';  TABLE users; --",
			desc:     "SQL keywords should be sanitized",
		},
		{
			name:     "potential script injection",
			input:    "Apple<script>alert('xss')</script>",
			expected: "Apple",
			desc:     "Script tags should be completely removed",
		},
		{
			name:     "javascript handler",
			input:    "Apple onclick=alert('xss')",
			expected: "Apple  alert('xss')",
			desc:     "JavaScript handlers should be sanitized",
		},
		{
			name:     "case insensitive SQL",
			input:    "Apple UNION SELECT * FROM users",
			expected: "Apple   * FROM users",
			desc:     "Case insensitive SQL patterns should be caught",
		},
		{
			name:     "legitimate text with SQL words",
			input:    "I like to select apples from the store",
			expected: "I like to  apples from the store",
			desc:     "SQL words in context should be sanitized for safety",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeText(tt.input)
			assert.Equal(t, tt.expected, result, tt.desc)

			// Additional security checks - should not contain dangerous patterns
			dangerousPatterns := []string{"script", "javascript", "vbscript", "select ", "union ", "drop ", "insert ", "update ", "delete "}
			lowerResult := strings.ToLower(result)
			for _, pattern := range dangerousPatterns {
				assert.NotContains(t, lowerResult, pattern, "Result should not contain dangerous pattern: %s", pattern)
			}
		})
	}
}

func TestValidateParsedItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []Item
		expected int
		desc     string
	}{
		{
			name: "valid items",
			items: []Item{
				{Name: "Apple", Grams: 100.0},
				{Name: "Banana", Grams: 120.0},
			},
			expected: 2,
			desc:     "Valid items should be kept",
		},
		{
			name: "item with empty name",
			items: []Item{
				{Name: "", Grams: 100.0},
				{Name: "Apple", Grams: 100.0},
			},
			expected: 1,
			desc:     "Items with empty names should be filtered out",
		},
		{
			name: "item with oversized name",
			items: []Item{
				{Name: strings.Repeat("a", maxItemNameLength+10), Grams: 100.0},
			},
			expected: 1,
			desc:     "Oversized names should be truncated but item kept",
		},
		{
			name: "item with negative grams",
			items: []Item{
				{Name: "Apple", Grams: -10.0},
				{Name: "Banana", Grams: 120.0},
			},
			expected: 2,
			desc:     "Items with negative grams should be reset to 0",
		},
		{
			name: "item with excessive grams",
			items: []Item{
				{Name: "Apple", Grams: 50000.0}, // 50kg
				{Name: "Banana", Grams: 120.0},
			},
			expected: 2,
			desc:     "Items with excessive grams should be reset to 0",
		},
		{
			name: "item with invalid user quantity",
			items: []Item{
				{Name: "Apple", Grams: 100.0, UserQuantity: func() *float64 { v := -5.0; return &v }()},
				{Name: "Banana", Grams: 120.0, UserQuantity: func() *float64 { v := 2000.0; return &v }()},
			},
			expected: 2,
			desc:     "Invalid user quantities should be removed",
		},
		{
			name: "item with oversized brand",
			items: []Item{
				{Name: "Apple", Grams: 100.0, Brand: func() *string { v := strings.Repeat("b", maxBrandLength+10); return &v }()},
			},
			expected: 1,
			desc:     "Oversized brands should be truncated but item kept",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateParsedItems(tt.items)
			assert.Equal(t, tt.expected, len(result), tt.desc)

			// Additional validation checks
			for _, item := range result {
				assert.NotEmpty(t, item.Name, "All returned items should have non-empty names")
				assert.LessOrEqual(t, len(item.Name), maxItemNameLength, "Item names should not exceed max length")
				assert.GreaterOrEqual(t, item.Grams, 0.0, "Grams should be non-negative")
				assert.LessOrEqual(t, item.Grams, 10000.0, "Grams should be reasonable")

				if item.Brand != nil {
					assert.LessOrEqual(t, len(*item.Brand), maxBrandLength, "Brand names should not exceed max length")
				}

				if item.UserQuantity != nil {
					assert.GreaterOrEqual(t, *item.UserQuantity, 0.0, "User quantity should be non-negative")
					assert.LessOrEqual(t, *item.UserQuantity, 1000.0, "User quantity should be reasonable")
				}
			}
		})
	}
}

func TestTruncateForLog(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "exact length",
			input:    strings.Repeat("a", 50),
			expected: strings.Repeat("a", 50),
		},
		{
			name:     "long text",
			input:    strings.Repeat("a", 100),
			expected: strings.Repeat("a", 50) + "...",
		},
		{
			name:     "empty text",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateForLog(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 53, "Result should never exceed 50 chars + '...'")
		})
	}
}

func TestSecurityConstants(t *testing.T) {
	// Test that security constants are reasonable
	assert.Greater(t, maxTranscriptLength, 1000, "Max transcript length should allow reasonable transcripts")
	assert.Less(t, maxTranscriptLength, 100000, "Max transcript length should prevent abuse")

	assert.Greater(t, maxItemNameLength, 50, "Max item name length should allow reasonable food names")
	assert.Less(t, maxItemNameLength, 1000, "Max item name length should prevent abuse")

	assert.Greater(t, maxBrandLength, 20, "Max brand length should allow reasonable brand names")
	assert.Less(t, maxBrandLength, 500, "Max brand length should prevent abuse")
}
