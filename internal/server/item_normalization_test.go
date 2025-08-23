package server

import (
	"testing"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizationLogic(t *testing.T) {
	service := NewNutritionService(nil)

	t.Run("SingleUnitItem", func(t *testing.T) {
		item := Item{
			Name:         "Apple",
			Grams:        180.0,
			UserQuantity: floatPtr(1.0),
			UserUnit:     stringPtr("apple"),
			BaseQuantity: nil, // Single unit, no normalization needed
		}

		normalizedGrams := service.getNormalizedGrams(item)
		assert.Equal(t, 180.0, normalizedGrams)
	})

	t.Run("MultiUnitItemNormalization", func(t *testing.T) {
		item := Item{
			Name:         "Bananas",
			Grams:        240.0, // Total weight for 2 bananas
			UserQuantity: floatPtr(2.0),
			UserUnit:     stringPtr("bananas"),
			BaseQuantity: floatPtr(2.0), // Should normalize to single banana
		}

		normalizedGrams := service.getNormalizedGrams(item)
		assert.Equal(t, 120.0, normalizedGrams) // 240 / 2 = 120g per banana
	})

	t.Run("FractionalQuantityDoesNotNormalize", func(t *testing.T) {
		item := Item{
			Name:         "Pizza slice",
			Grams:        75.0, // Half a slice
			UserQuantity: floatPtr(0.5),
			UserUnit:     stringPtr("slice"),
			BaseQuantity: nil, // Fractional quantities don't set BaseQuantity
		}

		normalizedGrams := service.getNormalizedGrams(item)
		assert.Equal(t, 75.0, normalizedGrams) // No normalization
	})
}

func TestConvertNutrientsToExactCacheWithNormalization(t *testing.T) {
	service := NewNutritionService(nil)

	t.Run("NormalizeMultiUnitNutrition", func(t *testing.T) {
		// Simulate "2 cans of soda" - should store nutrition for 1 can
		item := Item{
			Name:         "Coca Cola",
			Grams:        710.0, // 2 cans worth
			UserQuantity: floatPtr(2.0),
			UserUnit:     stringPtr("cans"),
			BaseQuantity: floatPtr(2.0), // Normalize to single can
		}

		// Nutrition values for 2 cans (what the AI would return)
		nutrients := CompleteNutrient{
			Calories:     280, // Total for 2 cans
			Protein:      0,
			TotalFat:     0,
			SaturatedFat: 0,
			TransFat:     0,
			Cholesterol:  0,
			Sodium:       70, // Total for 2 cans
			TotalCarbs:   76, // Total for 2 cans
			DietaryFiber: 0,
			TotalSugars:  74, // Total for 2 cans
			AddedSugars:  74,
		}

		result := service.convertNutrientsToExactCache(item, nutrients, "test_key")

		// Should store normalized values (divided by 2)
		assert.Equal(t, 355.0, *result.OriginalServingGrams) // 710 / 2 = 355g per can
		assert.Equal(t, 140.0, *result.OriginalCalories)     // 280 / 2 = 140 calories per can
		assert.Equal(t, 35.0, *result.OriginalSodiumMg)      // 70 / 2 = 35mg per can
		assert.Equal(t, 38.0, *result.OriginalTotalCarbsG)   // 76 / 2 = 38g per can
		assert.Equal(t, 37.0, *result.OriginalTotalSugarsG)  // 74 / 2 = 37g per can
		assert.Equal(t, 37.0, *result.OriginalAddedSugarsG)  // 74 / 2 = 37g per can
	})

	t.Run("NoNormalizationForSingleUnit", func(t *testing.T) {
		// Single unit item should not be normalized
		item := Item{
			Name:         "Apple",
			Grams:        180.0,
			UserQuantity: floatPtr(1.0),
			UserUnit:     stringPtr("apple"),
			BaseQuantity: nil,
		}

		nutrients := CompleteNutrient{
			Calories:     95,
			Protein:      0.5,
			TotalFat:     0.3,
			SaturatedFat: 0.1,
			TransFat:     0,
			Cholesterol:  0,
			Sodium:       2,
			TotalCarbs:   25,
			DietaryFiber: 4,
			TotalSugars:  19,
			AddedSugars:  0,
		}

		result := service.convertNutrientsToExactCache(item, nutrients, "test_key")

		// Should store original values (no division)
		assert.Equal(t, 180.0, *result.OriginalServingGrams)
		assert.Equal(t, 95.0, *result.OriginalCalories)
		assert.Equal(t, 0.5, *result.OriginalProteinG)
		assert.Equal(t, 2.0, *result.OriginalSodiumMg)
		assert.Equal(t, 25.0, *result.OriginalTotalCarbsG)
		assert.Equal(t, 19.0, *result.OriginalTotalSugarsG)
	})
}

func TestOpenAIProviderBaseQuantityDetection(t *testing.T) {
	// Test that the OpenAI provider correctly detects multi-unit quantities
	t.Run("DetectMultiUnitQuantity", func(t *testing.T) {
		// Simulate parsing result that would come from OpenAI
		parsed := struct {
			Success bool    `json:"success"`
			Message *string `json:"message"`
			Items   []struct {
				Name         string   `json:"name"`
				Grams        *float64 `json:"grams"`
				UserQuantity *float64 `json:"user_quantity"`
				UserUnit     *string  `json:"user_unit"`
				Brand        *string  `json:"brand"`
			} `json:"items"`
		}{
			Success: true,
			Items: []struct {
				Name         string   `json:"name"`
				Grams        *float64 `json:"grams"`
				UserQuantity *float64 `json:"user_quantity"`
				UserUnit     *string  `json:"user_unit"`
				Brand        *string  `json:"brand"`
			}{
				{
					Name:         "Coca Cola",
					Grams:        floatPtr(710.0), // 2 cans
					UserQuantity: floatPtr(2.0),
					UserUnit:     stringPtr("cans"),
					Brand:        nil,
				},
				{
					Name:         "Apple",
					Grams:        floatPtr(180.0), // 1 apple
					UserQuantity: floatPtr(1.0),
					UserUnit:     stringPtr("apple"),
					Brand:        nil,
				},
			},
		}

		// Manually process the items like the OpenAI provider would
		var items []Item
		for _, i := range parsed.Items {
			var baseQuantity *float64
			if i.UserQuantity != nil && *i.UserQuantity > 1.0 {
				baseQuantity = i.UserQuantity
			}

			items = append(items, Item{
				Name:         i.Name,
				Grams:        *i.Grams,
				UserQuantity: i.UserQuantity,
				UserUnit:     i.UserUnit,
				Brand:        i.Brand,
				BaseQuantity: baseQuantity,
			})
		}

		require.Len(t, items, 2)

		// First item (2 cans) should have BaseQuantity set
		coca := items[0]
		assert.Equal(t, "Coca Cola", coca.Name)
		assert.Equal(t, 710.0, coca.Grams)
		assert.NotNil(t, coca.BaseQuantity)
		assert.Equal(t, 2.0, *coca.BaseQuantity)

		// Second item (1 apple) should NOT have BaseQuantity set
		apple := items[1]
		assert.Equal(t, "Apple", apple.Name)
		assert.Equal(t, 180.0, apple.Grams)
		assert.Nil(t, apple.BaseQuantity)
	})
}

func TestScalingFromNormalizedCache(t *testing.T) {
	// This test verifies that scaling works correctly from normalized cached values
	service := NewNutritionService(nil)

	t.Run("ScaleFromNormalizedBase", func(t *testing.T) {
		// Simulate cached item that was normalized (1 can)
		cachedItem := &storage.Item{
			OriginalServingGrams: floatPtr(355.0), // 1 can
			OriginalCalories:     floatPtr(140.0), // 1 can
			OriginalSodiumMg:     floatPtr(35.0),  // 1 can
			OriginalTotalCarbsG:  floatPtr(38.0),  // 1 can
			OriginalTotalSugarsG: floatPtr(37.0),  // 1 can
		}

		// User wants 3 cans
		fromGrams := 355.0 // Base: 1 can
		toGrams := 1065.0  // Target: 3 cans (355 * 3)

		result := service.scaleNutritionFromCachedServing(cachedItem, fromGrams, toGrams)

		// Should scale by factor of 3
		assert.Equal(t, 420.0, result.Calories)    // 140 * 3
		assert.Equal(t, 105.0, result.Sodium)      // 35 * 3
		assert.Equal(t, 114.0, result.TotalCarbs)  // 38 * 3
		assert.Equal(t, 111.0, result.TotalSugars) // 37 * 3
	})
}

func TestScalingMethodSelection(t *testing.T) {
	service := NewNutritionService(nil)

	t.Run("ShouldUse100gScalingForGramInput", func(t *testing.T) {
		item := Item{
			Name:         "Chicken breast",
			Grams:        250.0,
			UserQuantity: floatPtr(250.0),
			UserUnit:     stringPtr("g"), // User provided grams
		}

		cachedItem := &storage.Item{
			OriginalServingGrams: floatPtr(150.0), // Some base serving
		}

		should100g := service.shouldUse100gScaling(item, cachedItem)
		assert.True(t, should100g, "Should use 100g scaling when user provides grams")
	})

	t.Run("ShouldUseBaseScalingForLogicalUnits", func(t *testing.T) {
		item := Item{
			Name:         "Coca Cola",
			Grams:        355.0,
			UserQuantity: floatPtr(1.0),
			UserUnit:     stringPtr("can"), // Logical unit
		}

		cachedItem := &storage.Item{
			OriginalServingGrams: floatPtr(355.0), // Reliable base serving
		}

		should100g := service.shouldUse100gScaling(item, cachedItem)
		assert.False(t, should100g, "Should use base unit scaling for logical units with reliable cache")
	})

	t.Run("ShouldUse100gScalingWhenCacheUnreliable", func(t *testing.T) {
		item := Item{
			Name:         "Apple",
			Grams:        180.0,
			UserQuantity: floatPtr(1.0),
			UserUnit:     stringPtr("apple"), // Logical unit
		}

		cachedItem := &storage.Item{
			OriginalServingGrams: nil, // No reliable base serving data
		}

		should100g := service.shouldUse100gScaling(item, cachedItem)
		assert.True(t, should100g, "Should use 100g scaling when cache lacks reliable base serving data")
	})
}

func TestEndToEndNormalizationFlow(t *testing.T) {
	// This test demonstrates the complete normalization flow from parsing to caching to scaling
	service := NewNutritionService(nil)

	t.Run("CompleteNormalizationFlow", func(t *testing.T) {
		// Step 1: Simulate user input "2 cans of Coca Cola" being parsed
		item := Item{
			Name:         "Coca Cola",
			Grams:        710.0, // Total weight for 2 cans (355g each)
			UserQuantity: floatPtr(2.0),
			UserUnit:     stringPtr("cans"),
			BaseQuantity: floatPtr(2.0), // Detected as multi-unit
		}

		// Step 2: Simulate AI returning nutrition for 2 cans
		aiNutrients := CompleteNutrient{
			Calories:    280, // Total for 2 cans
			Sodium:      70,  // Total for 2 cans
			TotalCarbs:  76,  // Total for 2 cans
			TotalSugars: 74,  // Total for 2 cans
		}

		// Step 3: System should normalize for caching (divide by 2)
		normalizedGrams := service.getNormalizedGrams(item)
		assert.Equal(t, 355.0, normalizedGrams) // 710 / 2 = 355g per can

		// Step 4: Cache should store single-can values
		exactKey := service.makeExactServingKey("coca_cola", "", normalizedGrams)
		cachedItem := service.convertNutrientsToExactCache(item, aiNutrients, exactKey)

		// Verify cached values are normalized (single can)
		assert.Equal(t, 355.0, *cachedItem.OriginalServingGrams) // 710 / 2
		assert.Equal(t, 140.0, *cachedItem.OriginalCalories)     // 280 / 2
		assert.Equal(t, 35.0, *cachedItem.OriginalSodiumMg)      // 70 / 2
		assert.Equal(t, 38.0, *cachedItem.OriginalTotalCarbsG)   // 76 / 2
		assert.Equal(t, 37.0, *cachedItem.OriginalTotalSugarsG)  // 74 / 2

		// Step 5: When another user wants 3 cans, scaling should work correctly
		requestedGrams := 1065.0 // 3 cans (355 * 3)
		scaledNutrition := service.scaleNutritionFromCachedServing(
			cachedItem,
			355.0,          // Base: 1 can (from cache)
			requestedGrams, // Target: 3 cans
		)

		// Verify scaling is correct (3x single can)
		assert.Equal(t, 420.0, scaledNutrition.Calories)    // 140 * 3
		assert.Equal(t, 105.0, scaledNutrition.Sodium)      // 35 * 3
		assert.Equal(t, 114.0, scaledNutrition.TotalCarbs)  // 38 * 3
		assert.Equal(t, 111.0, scaledNutrition.TotalSugars) // 37 * 3

		// Step 6: Verify original user input is preserved for display
		assert.Equal(t, 2.0, *item.UserQuantity)
		assert.Equal(t, "cans", *item.UserUnit)
	})

	t.Run("FallbackTo100gScalingWhenAppropriate", func(t *testing.T) {
		// Create item where user provided grams
		item := Item{
			Name:         "Chicken breast",
			Grams:        250.0, // User provided 250g
			UserQuantity: floatPtr(250.0),
			UserUnit:     stringPtr("g"), // User specified grams
			BaseQuantity: nil,            // No normalization for gram inputs
		}

		// Mock cached item with unreliable base data
		cachedItem := &storage.Item{
			OriginalServingGrams: nil, // No reliable base serving
			// Per-100g data available
			CaloriesPer100g: 165.0,
			ProteinGPer100g: 31.0,
		}

		// Should choose 100g scaling over base scaling
		should100g := service.shouldUse100gScaling(item, cachedItem)
		assert.True(t, should100g)

		// Verify 100g scaling works
		nutrition := service.convertCachedToNutrients(cachedItem, item)
		expectedCalories := 165.0 * (250.0 / 100.0) // 165 * 2.5 = 412.5
		expectedProtein := 31.0 * (250.0 / 100.0)   // 31 * 2.5 = 77.5

		assert.InDelta(t, expectedCalories, nutrition.Calories, 0.1)
		assert.InDelta(t, expectedProtein, nutrition.Protein, 0.1)
	})
}

// Helper functions for creating pointers
func floatPtr(f float64) *float64 {
	return &f
}
