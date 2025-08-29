package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateBrandAwareVariationsPhase2(t *testing.T) {
	tests := []struct {
		name        string
		itemName    string
		brand       *string
		expected    []string
		description string
	}{
		{
			name:        "No brand provided",
			itemName:    "vanilla yogurt",
			brand:       nil,
			expected:    []string{"vanilla yogurt"},
			description: "Should return only original name when no brand",
		},
		{
			name:        "Empty brand provided",
			itemName:    "greek yogurt",
			brand:       stringPtr(""),
			expected:    []string{"greek yogurt"},
			description: "Should return only original name when brand is empty",
		},
		{
			name:        "Simple brand and item",
			itemName:    "vanilla yogurt",
			brand:       stringPtr("noosa"),
			expected:    []string{"vanilla yogurt", "noosa vanilla yogurt", "vanilla yogurt noosa"},
			description: "Should generate brand + item and item + brand variations",
		},
		{
			name:        "Brand already in item name (prefix)",
			itemName:    "noosa vanilla yogurt",
			brand:       stringPtr("noosa"),
			expected:    []string{"noosa vanilla yogurt", "noosa noosa vanilla yogurt", "noosa vanilla yogurt noosa", "vanilla yogurt"},
			description: "Should generate variations and extract base product when brand is already in name",
		},
		{
			name:        "Brand already in item name (suffix)",
			itemName:    "vanilla yogurt noosa",
			brand:       stringPtr("noosa"),
			expected:    []string{"vanilla yogurt noosa", "noosa vanilla yogurt noosa", "vanilla yogurt noosa noosa", "vanilla yogurt"},
			description: "Should generate variations and extract base product when brand is at end",
		},
		{
			name:        "Multi-word brand",
			itemName:    "ice cream",
			brand:       stringPtr("ben jerry"),
			expected:    []string{"ice cream", "ben jerry ice cream", "ice cream ben jerry", "jerry ben ice cream"},
			description: "Should handle multi-word brands with reversed order variation",
		},
		{
			name:        "Case insensitive matching",
			itemName:    "Greek Yogurt",
			brand:       stringPtr("FAGE"),
			expected:    []string{"greek yogurt", "fage greek yogurt", "greek yogurt fage"},
			description: "Should normalize case for consistent matching",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateBrandAwareVariations(tt.itemName, tt.brand)
			assert.ElementsMatch(t, tt.expected, result, tt.description)
		})
	}
}

func TestBrandSafetyPhase2(t *testing.T) {
	tests := []struct {
		name     string
		itemName string
		brand1   *string
		brand2   *string
	}{
		{
			name:     "Different brands should not cross-contaminate",
			itemName: "vanilla yogurt",
			brand1:   stringPtr("noosa"),
			brand2:   stringPtr("chobani"),
		},
		{
			name:     "Generic vs branded items",
			itemName: "yogurt",
			brand1:   nil,
			brand2:   stringPtr("dannon"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variations1 := generateBrandAwareVariations(tt.itemName, tt.brand1)
			variations2 := generateBrandAwareVariations(tt.itemName, tt.brand2)

			// Ensure no cross-contamination between different brands
			for _, v1 := range variations1 {
				for _, v2 := range variations2 {
					if tt.brand1 != nil && tt.brand2 != nil {
						// If both have brands, variations should be completely different
						// except for the base item name case
						if v1 == v2 && v1 != tt.itemName {
							t.Errorf("Brand cross-contamination detected: %s appears in both %v and %v variations",
								v1, getBrandOrEmpty(tt.brand1), getBrandOrEmpty(tt.brand2))
						}
					}
				}
			}
		})
	}
}

func TestRealWorldBrandAwareExamples(t *testing.T) {
	tests := []struct {
		name          string
		itemName      string
		brand         *string
		shouldContain []string
		description   string
	}{
		{
			name:          "Noosa yogurt variations",
			itemName:      "vanilla yogurt",
			brand:         stringPtr("noosa"),
			shouldContain: []string{"vanilla yogurt", "noosa vanilla yogurt", "vanilla yogurt noosa"},
			description:   "Should generate common noosa yogurt variations for cache matching",
		},
		{
			name:          "Soda brand matching",
			itemName:      "cream soda",
			brand:         stringPtr("olipop"),
			shouldContain: []string{"cream soda", "olipop cream soda", "cream soda olipop"},
			description:   "Should help match 'half can cream soda Olipop' to cached 'cream soda olipop'",
		},
		{
			name:          "Multi-word brand handling",
			itemName:      "ice cream",
			brand:         stringPtr("ben jerry"),
			shouldContain: []string{"ice cream", "ben jerry ice cream", "ice cream ben jerry"},
			description:   "Should handle compound brands like Ben & Jerry's",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variations := generateBrandAwareVariations(tt.itemName, tt.brand)

			for _, expected := range tt.shouldContain {
				assert.Contains(t, variations, expected,
					"Expected '%s' to be in variations for %s", expected, tt.description)
			}
		})
	}
}
