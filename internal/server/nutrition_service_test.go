package server

import (
	"testing"
	"time"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewNutritionService(t *testing.T) {
	// Mock store
	var store storage.Store

	service := NewNutritionService(store)

	assert.NotNil(t, service)
	assert.NotNil(t, service.aiProvider)
	assert.NotNil(t, service.converter)
	assert.Equal(t, store, service.store)
}

func TestNutritionService_ConvertCachedToNutrients(t *testing.T) {
	service := NewNutritionService(nil)

	// Create mock cached data (per 100g values)
	cached := &storage.ItemCache{
		CaloriesPer100g:      100,
		ProteinGPer100g:      20,
		TotalFatGPer100g:     5,
		SaturatedFatGPer100g: 2,
		TransFatGPer100g:     0,
		CholesterolMgPer100g: 10,
		SodiumMgPer100g:      200,
		TotalCarbsGPer100g:   15,
		DietaryFiberGPer100g: 3,
		TotalSugarsGPer100g:  8,
		AddedSugarsGPer100g:  2,
		VitaminAMcgPer100g:   500,
		VitaminCMgPer100g:    60,
		CalciumMgPer100g:     150,
		IronMgPer100g:        2,
	}

	// Test item for 200g serving
	item := Item{
		Name:     "Test Food",
		Quantity: float64Ptr(200),
		Unit:     stringPtr("g"),
	}

	result := service.convertCachedToNutrients(cached, item)

	assert.Equal(t, 200.0, result.Calories)   // 100 * 2
	assert.Equal(t, 40.0, result.Protein)     // 20 * 2
	assert.Equal(t, 10.0, result.TotalFat)    // 5 * 2
	assert.Equal(t, 4.0, result.SaturatedFat) // 2 * 2
	assert.Equal(t, 0.0, result.TransFat)     // 0 * 2
	assert.Equal(t, 20.0, result.Cholesterol) // 10 * 2
	assert.Equal(t, 400.0, result.Sodium)     // 200 * 2
	assert.Equal(t, 30.0, result.TotalCarbs)  // 15 * 2
	assert.Equal(t, 6.0, result.DietaryFiber) // 3 * 2
	assert.Equal(t, 16.0, result.TotalSugars) // 8 * 2
	assert.Equal(t, 4.0, result.AddedSugars)  // 2 * 2
	assert.Equal(t, 1000.0, result.VitaminA)  // 500 * 2
	assert.Equal(t, 120.0, result.VitaminC)   // 60 * 2
	assert.Equal(t, 300.0, result.Calcium)    // 150 * 2
	assert.Equal(t, 4.0, result.Iron)         // 2 * 2
}

func TestNutritionService_ConvertNutrientsToCache(t *testing.T) {
	service := NewNutritionService(nil)

	// Create item with serving data
	item := Item{
		Name:     "Test Food",
		Quantity: stringToFloat64Ptr("200"), // 200g serving
		Unit:     stringPtr("g"),
		Brand:    stringPtr("Test Brand"),
	}

	// Nutrition data for 200g serving
	nutrients := CompleteNutrient{
		Calories:     200,
		Protein:      40,
		TotalFat:     10,
		SaturatedFat: 4,
		TransFat:     0,
		Cholesterol:  20,
		Sodium:       400,
		TotalCarbs:   30,
		DietaryFiber: 6,
		TotalSugars:  16,
		AddedSugars:  4,
		VitaminA:     1000,
		VitaminC:     120,
		Calcium:      300,
		Iron:         4,
	}

	result := service.convertNutrientsToCache(item, nutrients)

	// Check that values were converted to per-100g (should be halved)
	assert.Equal(t, "test food", result.NormalizedName)
	assert.Equal(t, "test brand", result.NormalizedBrand)
	assert.Equal(t, "Test Food", result.DisplayName)
	assert.Equal(t, "Test Brand", result.DisplayBrand)

	assert.Equal(t, 100.0, result.CaloriesPer100g)     // 200 / 2
	assert.Equal(t, 20.0, result.ProteinGPer100g)      // 40 / 2
	assert.Equal(t, 5.0, result.TotalFatGPer100g)      // 10 / 2
	assert.Equal(t, 2.0, result.SaturatedFatGPer100g)  // 4 / 2
	assert.Equal(t, 0.0, result.TransFatGPer100g)      // 0 / 2
	assert.Equal(t, 10.0, result.CholesterolMgPer100g) // 20 / 2
	assert.Equal(t, 200.0, result.SodiumMgPer100g)     // 400 / 2
	assert.Equal(t, 15.0, result.TotalCarbsGPer100g)   // 30 / 2
	assert.Equal(t, 3.0, result.DietaryFiberGPer100g)  // 6 / 2
	assert.Equal(t, 8.0, result.TotalSugarsGPer100g)   // 16 / 2
	assert.Equal(t, 2.0, result.AddedSugarsGPer100g)   // 4 / 2
	assert.Equal(t, 500.0, result.VitaminAMcgPer100g)  // 1000 / 2
	assert.Equal(t, 60.0, result.VitaminCMgPer100g)    // 120 / 2
	assert.Equal(t, 150.0, result.CalciumMgPer100g)    // 300 / 2
	assert.Equal(t, 2.0, result.IronMgPer100g)         // 4 / 2

	// Check timestamps are set
	assert.False(t, result.FetchedAt.IsZero())
	assert.False(t, result.ExpiresAt.IsZero())
	assert.True(t, result.ExpiresAt.After(result.FetchedAt))

	// Check TTL is approximately 30 days
	expectedTTL := 30 * 24 * time.Hour
	actualTTL := result.ExpiresAt.Sub(result.FetchedAt)
	assert.InDelta(t, expectedTTL.Seconds(), actualTTL.Seconds(), 60) // Within 1 minute
}

func TestNutritionService_RoundingPrecision(t *testing.T) {
	service := NewNutritionService(nil)

	// Test with precise values that need rounding
	cached := &storage.ItemCache{
		CaloriesPer100g:      123.456789,
		ProteinGPer100g:      12.3456789,
		ThiamineMgPer100g:    0.123456789, // Should round to 3 decimal places
		VitaminB12McgPer100g: 2.3456789,   // Should round to 2 decimal places
		ZincMgPer100g:        1.23456789,  // Should round to 2 decimal places
	}

	testItem := Item{
		Name:     "Test Food",
		Quantity: float64Ptr(150), // 1.5x multiplier (150g serving)
		Unit:     stringPtr("g"),
	}
	result := service.convertCachedToNutrients(cached, testItem)

	// Check rounding precision
	assert.Equal(t, 185.2, result.Calories)  // 123.456789 * 1.5 = 185.185... rounded to 1 decimal
	assert.Equal(t, 18.5, result.Protein)    // 12.3456789 * 1.5 = 18.518... rounded to 1 decimal
	assert.Equal(t, 0.185, result.Thiamine)  // 0.123456789 * 1.5 = 0.185... rounded to 3 decimals
	assert.Equal(t, 3.52, result.VitaminB12) // 2.3456789 * 1.5 = 3.5185... rounded to 2 decimals
	assert.Equal(t, 1.85, result.Zinc)       // 1.23456789 * 1.5 = 1.8518... rounded to 2 decimals
}

// Helper function to convert string to float64 pointer
func stringToFloat64Ptr(s string) *float64 {
	if s == "200" {
		v := 200.0
		return &v
	}
	return nil
}
