package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeItemNameForCache(t *testing.T) {
	tests := []struct {
		name           string
		itemName       string
		brand          *string
		expectedResult string
		description    string
	}{
		{
			name:           "Brand at beginning",
			itemName:       "Olipop cream soda",
			brand:          stringPtr("Olipop"),
			expectedResult: "cream soda",
			description:    "Should remove brand from beginning",
		},
		{
			name:           "Brand at end",
			itemName:       "cream soda Olipop",
			brand:          stringPtr("Olipop"),
			expectedResult: "cream soda",
			description:    "Should remove brand from end",
		},
		{
			name:           "Brand in different case",
			itemName:       "OLIPOP Cream Soda",
			brand:          stringPtr("olipop"),
			expectedResult: "cream soda",
			description:    "Should handle case-insensitive brand matching",
		},
		{
			name:           "Multi-word brand at beginning",
			itemName:       "Ben Jerry vanilla ice cream",
			brand:          stringPtr("Ben Jerry"),
			expectedResult: "vanilla ice cream",
			description:    "Should remove multi-word brand from beginning",
		},
		{
			name:           "Multi-word brand at end",
			itemName:       "vanilla ice cream Ben Jerry",
			brand:          stringPtr("Ben Jerry"),
			expectedResult: "vanilla ice cream",
			description:    "Should remove multi-word brand from end",
		},
		{
			name:           "No brand provided",
			itemName:       "vanilla ice cream",
			brand:          nil,
			expectedResult: "vanilla ice cream",
			description:    "Should return original when no brand provided",
		},
		{
			name:           "Empty brand",
			itemName:       "vanilla ice cream",
			brand:          stringPtr(""),
			expectedResult: "vanilla ice cream",
			description:    "Should return original when brand is empty",
		},
		{
			name:           "Brand not in name",
			itemName:       "vanilla ice cream",
			brand:          stringPtr("Haagen Daz"),
			expectedResult: "vanilla ice cream",
			description:    "Should return original when brand not found in name",
		},
		{
			name:           "Standalone brand word",
			itemName:       "Olipop",
			brand:          stringPtr("Olipop"),
			expectedResult: "olipop", // fallback to original since we can't remove everything
			description:    "Should fallback to original when removing brand would leave empty string",
		},
		{
			name:           "Brand appears multiple times",
			itemName:       "Olipop cream soda Olipop",
			brand:          stringPtr("Olipop"),
			expectedResult: "cream soda", // Should remove first occurrence
			description:    "Should remove brand from beginning when it appears multiple times",
		},
		{
			name:           "Complex case - Noosa",
			itemName:       "vanilla yogurt noosa",
			brand:          stringPtr("Noosa"),
			expectedResult: "vanilla yogurt",
			description:    "Should normalize Noosa brand correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeItemNameForCache(tt.itemName, tt.brand)
			assert.Equal(t, tt.expectedResult, result, tt.description)
		})
	}
}

func TestNormalizeItemNameWithQuantityForCache(t *testing.T) {
	tests := []struct {
		name                   string
		itemName               string
		brand                  *string
		expectedNormalizedName string
		expectedMultiplier     float64
		description            string
	}{
		{
			name:                   "Half can with brand at end",
			itemName:               "half can cream soda Olipop",
			brand:                  stringPtr("Olipop"),
			expectedNormalizedName: "cream soda",
			expectedMultiplier:     0.5,
			description:            "Should extract quantity and remove brand for cache key",
		},
		{
			name:                   "Quarter cup with brand at beginning",
			itemName:               "quarter cup Noosa vanilla yogurt",
			brand:                  stringPtr("Noosa"),
			expectedNormalizedName: "vanilla yogurt",
			expectedMultiplier:     0.25,
			description:            "Should extract quantity and remove brand from beginning",
		},
		{
			name:                   "Two containers with brand",
			itemName:               "2 containers Ben Jerry vanilla ice cream",
			brand:                  stringPtr("Ben Jerry"),
			expectedNormalizedName: "vanilla ice cream",
			expectedMultiplier:     2.0,
			description:            "Should extract multi-unit quantity and remove multi-word brand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizedName, quantityInfo := normalizeItemNameWithQuantityForCache(tt.itemName, tt.brand)
			assert.Equal(t, tt.expectedNormalizedName, normalizedName, tt.description+" (name)")
			assert.Equal(t, tt.expectedMultiplier, quantityInfo.Multiplier, tt.description+" (multiplier)")
		})
	}
}
