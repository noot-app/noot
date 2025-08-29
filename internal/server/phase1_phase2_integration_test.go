package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhase1And2Integration(t *testing.T) {
	// Test that Phase 1 (quantity) and Phase 2 (brand-aware) work together
	tests := []struct {
		name               string
		itemName           string
		brand              *string
		expectedQuantity   float64
		expectedCleanName  string
		expectedVariations int // minimum number of variations expected
		description        string
	}{
		{
			name:               "Quantity + Brand Integration",
			itemName:           "half can cream soda",
			brand:              stringPtr("olipop"),
			expectedQuantity:   0.5,
			expectedCleanName:  "cream soda",
			expectedVariations: 3, // original + brand variations
			description:        "Should extract quantity AND generate brand variations",
		},
		{
			name:               "Multi-word brand with quantity",
			itemName:           "2 pints ice cream",
			brand:              stringPtr("ben jerry"),
			expectedQuantity:   2.0,
			expectedCleanName:  "ice cream",
			expectedVariations: 4, // original + brand variations + reversed brand
			description:        "Should handle complex quantity + multi-word brand",
		},
		{
			name:               "Size descriptor + brand",
			itemName:           "large greek yogurt",
			brand:              stringPtr("fage"),
			expectedQuantity:   1.3,
			expectedCleanName:  "greek yogurt",
			expectedVariations: 3, // original + brand variations
			description:        "Should extract size multiplier AND generate brand variations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Phase 1: Quantity extraction
			quantityInfo := extractQuantityFromName(tt.itemName)
			assert.Equal(t, tt.expectedQuantity, quantityInfo.Multiplier,
				"Phase 1: Expected quantity multiplier %.2f for '%s'", tt.expectedQuantity, tt.itemName)
			assert.Equal(t, tt.expectedCleanName, quantityInfo.CleanName,
				"Phase 1: Expected clean name '%s' for '%s'", tt.expectedCleanName, tt.itemName)

			// Test Phase 2: Brand-aware variations using the clean name from Phase 1
			variations := generateBrandAwareVariations(quantityInfo.CleanName, tt.brand)
			assert.GreaterOrEqual(t, len(variations), tt.expectedVariations,
				"Phase 2: Expected at least %d variations for clean name '%s' with brand '%s', got %d: %v",
				tt.expectedVariations, quantityInfo.CleanName, getBrandOrEmpty(tt.brand), len(variations), variations)

			// Verify the clean name is always in the variations
			assert.Contains(t, variations, quantityInfo.CleanName,
				"Phase 2: Clean name '%s' should be in brand variations", quantityInfo.CleanName)

			t.Logf("Integration test '%s': quantity=%.2f, clean='%s', variations=%v",
				tt.name, quantityInfo.Multiplier, quantityInfo.CleanName, variations)
		})
	}
}

func TestRealWorldIntegrationExamples(t *testing.T) {
	// Test real user input scenarios that benefit from both phases
	tests := []struct {
		userInput   string
		brand       *string
		description string
	}{
		{
			userInput:   "half can cream soda Olipop",
			brand:       stringPtr("olipop"),
			description: "User says brand in item name, LLM extracts it - should work with both quantity and brand matching",
		},
		{
			userInput:   "2 containers noosa vanilla yogurt",
			brand:       stringPtr("noosa"),
			description: "Multi-container branded item - quantity and brand intelligence working together",
		},
		{
			userInput:   "large bowl ben jerry ice cream",
			brand:       stringPtr("ben jerry"),
			description: "Size + container + multi-word brand - full integration test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.userInput, func(t *testing.T) {
			// Phase 1: Extract quantity and clean the name
			quantityInfo := extractQuantityFromName(tt.userInput)

			// Phase 2: Generate brand-aware variations for the clean name
			brandVariations := generateBrandAwareVariations(quantityInfo.CleanName, tt.brand)

			// Verify we get meaningful results from both phases
			assert.NotEqual(t, 1.0, quantityInfo.Multiplier,
				"Should extract non-default quantity from '%s'", tt.userInput)
			assert.NotEqual(t, tt.userInput, quantityInfo.CleanName,
				"Should clean the item name from '%s'", tt.userInput)
			assert.Greater(t, len(brandVariations), 1,
				"Should generate multiple brand variations for clean name '%s'", quantityInfo.CleanName)

			t.Logf("Real-world test '%s': quantity=%.2f, clean='%s', brand_variations=%v",
				tt.userInput, quantityInfo.Multiplier, quantityInfo.CleanName, brandVariations)
		})
	}
}
