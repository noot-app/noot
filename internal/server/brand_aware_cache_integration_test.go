package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrandAwareCacheKeyGeneration(t *testing.T) {
	service := &NutritionService{}

	tests := []struct {
		name        string
		scenarios   []cacheTestScenario
		description string
	}{
		{
			name: "Olipop cream soda cache consolidation",
			scenarios: []cacheTestScenario{
				{
					itemName:    "Olipop cream soda",
					brand:       stringPtr("Olipop"),
					grams:       355.0,
					expectedKey: "cream soda|olipop|355.0g",
					description: "Full can with brand at beginning",
				},
				{
					itemName:    "cream soda Olipop",
					brand:       stringPtr("Olipop"),
					grams:       355.0,
					expectedKey: "cream soda|olipop|355.0g",
					description: "Full can with brand at end - should generate SAME cache key",
				},
				{
					itemName:    "cream soda olipop", // different case
					brand:       stringPtr("Olipop"),
					grams:       355.0,
					expectedKey: "cream soda|olipop|355.0g",
					description: "Different case - should generate SAME cache key",
				},
			},
			description: "All Olipop cream soda variations should generate the same cache key regardless of brand position or case",
		},
		{
			name: "Ben Jerry ice cream cache consolidation",
			scenarios: []cacheTestScenario{
				{
					itemName:    "Ben Jerry vanilla ice cream",
					brand:       stringPtr("Ben Jerry"),
					grams:       473.0,
					expectedKey: "vanilla ice cream|ben jerry|473.0g",
					description: "Multi-word brand at beginning",
				},
				{
					itemName:    "vanilla ice cream Ben Jerry",
					brand:       stringPtr("Ben Jerry"),
					grams:       473.0,
					expectedKey: "vanilla ice cream|ben jerry|473.0g",
					description: "Multi-word brand at end - should generate SAME cache key",
				},
			},
			description: "Multi-word brands should be normalized consistently",
		},
		{
			name: "Noosa yogurt cache consolidation",
			scenarios: []cacheTestScenario{
				{
					itemName:    "vanilla yogurt noosa",
					brand:       stringPtr("Noosa"),
					grams:       170.0,
					expectedKey: "vanilla yogurt|noosa|170.0g",
					description: "Brand at end",
				},
				{
					itemName:    "noosa vanilla yogurt",
					brand:       stringPtr("Noosa"),
					grams:       170.0,
					expectedKey: "vanilla yogurt|noosa|170.0g",
					description: "Brand at beginning - should generate SAME cache key",
				},
			},
			description: "Noosa yogurt variations should generate the same cache key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var firstKey string

			for i, scenario := range tt.scenarios {
				// Use brand-aware normalization
				normalizedName := normalizeItemNameForCache(scenario.itemName, scenario.brand)
				normalizedBrand := normalizeItemName(getBrandOrEmpty(scenario.brand))

				actualKey := service.makeExactServingKey(normalizedName, normalizedBrand, scenario.grams)

				// Verify each scenario generates expected key
				assert.Equal(t, scenario.expectedKey, actualKey,
					"Scenario %d (%s): %s", i+1, scenario.description, scenario.itemName)

				// Verify all scenarios in the test case generate the SAME key
				if i == 0 {
					firstKey = actualKey
				} else {
					assert.Equal(t, firstKey, actualKey,
						"All scenarios should generate the same cache key. Scenario: %s", scenario.description)
				}
			}
		})
	}
}

func TestBrandAwareCachePreventsDuplication(t *testing.T) {
	t.Run("Before fix - would create duplicate cache entries", func(t *testing.T) {
		service := &NutritionService{}

		// OLD way (without brand-aware normalization) - would create different keys
		oldKey1 := service.makeExactServingKey("olipop cream soda", "olipop", 355.0)
		oldKey2 := service.makeExactServingKey("cream soda olipop", "olipop", 355.0)

		// These would be different, causing cache duplication
		assert.NotEqual(t, oldKey1, oldKey2, "Old method creates different cache keys for same product")
		assert.Equal(t, "olipop cream soda|olipop|355.0g", oldKey1)
		assert.Equal(t, "cream soda olipop|olipop|355.0g", oldKey2)
	})

	t.Run("After fix - creates unified cache entries", func(t *testing.T) {
		service := &NutritionService{}

		// NEW way (with brand-aware normalization) - creates same key
		name1 := normalizeItemNameForCache("Olipop cream soda", stringPtr("Olipop"))
		name2 := normalizeItemNameForCache("cream soda Olipop", stringPtr("Olipop"))

		newKey1 := service.makeExactServingKey(name1, "olipop", 355.0)
		newKey2 := service.makeExactServingKey(name2, "olipop", 355.0)

		// These should be identical, preventing cache duplication
		assert.Equal(t, newKey1, newKey2, "New method creates same cache key for same product")
		assert.Equal(t, "cream soda|olipop|355.0g", newKey1)
		assert.Equal(t, "cream soda|olipop|355.0g", newKey2)
	})
}

func TestBrandAwareCacheIntegrationWithQuantities(t *testing.T) {
	service := &NutritionService{}

	tests := []struct {
		name        string
		itemName    string
		brand       *string
		grams       float64
		expectedKey string
		description string
	}{
		{
			name:        "Half can Olipop - brand at end",
			itemName:    "half can cream soda Olipop",
			brand:       stringPtr("Olipop"),
			grams:       177.5,
			expectedKey: "cream soda|olipop|177.5g",
			description: "Fractional quantity with brand at end",
		},
		{
			name:        "Half can Olipop - brand at beginning",
			itemName:    "half can Olipop cream soda",
			brand:       stringPtr("Olipop"),
			grams:       177.5,
			expectedKey: "cream soda|olipop|177.5g",
			description: "Fractional quantity with brand at beginning - SAME key as above",
		},
		{
			name:        "Quarter cup Noosa - brand at end",
			itemName:    "quarter cup vanilla yogurt Noosa",
			brand:       stringPtr("Noosa"),
			grams:       42.5,
			expectedKey: "vanilla yogurt|noosa|42.5g",
			description: "Quarter quantity with brand at end",
		},
		{
			name:        "Quarter cup Noosa - brand at beginning",
			itemName:    "quarter cup Noosa vanilla yogurt",
			brand:       stringPtr("Noosa"),
			grams:       42.5,
			expectedKey: "vanilla yogurt|noosa|42.5g",
			description: "Quarter quantity with brand at beginning - SAME key as above",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract quantity and normalize name for cache
			normalizedName, quantityInfo := normalizeItemNameWithQuantityForCache(tt.itemName, tt.brand)
			normalizedBrand := normalizeItemName(getBrandOrEmpty(tt.brand))

			actualKey := service.makeExactServingKey(normalizedName, normalizedBrand, tt.grams)

			assert.Equal(t, tt.expectedKey, actualKey, tt.description)

			// Verify quantity extraction still works
			assert.Greater(t, quantityInfo.Multiplier, 0.0, "Should extract valid quantity multiplier")
		})
	}
}

type cacheTestScenario struct {
	itemName    string
	brand       *string
	grams       float64
	expectedKey string
	description string
}
