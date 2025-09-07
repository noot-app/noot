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
	cached := &storage.Item{
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
		Name:  "Test Food",
		Grams: 200.0,
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
		Name:  "Test Food",
		Grams: 200.0, // 200g serving
		Brand: stringPtr("Test Brand"),
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

	result := service.convertNutrientsToExactCache(item, nutrients, "test_key")

	// Check that values were converted to per-100g (should be halved)
	assert.Equal(t, "test_key", result.NormalizedName)    // Uses the exact cache key
	assert.Equal(t, "test brand", result.NormalizedBrand) // Normalized brand from "Test Brand"
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
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
	assert.True(t, result.UpdatedAt.After(result.CreatedAt.Add(-time.Second))) // Allow for very close timestamps

	// Check cache is fresh (less than 30 days old)
	expectedMaxAge := 30 * 24 * time.Hour
	actualAge := time.Since(result.UpdatedAt)
	assert.True(t, actualAge < expectedMaxAge)
}

func TestNutritionService_RoundingPrecision(t *testing.T) {
	service := NewNutritionService(nil)

	// Test with precise values that need rounding
	cached := &storage.Item{
		CaloriesPer100g:      123.456789,
		ProteinGPer100g:      12.3456789,
		ThiamineMgPer100g:    0.123456789, // Should round to 3 decimal places
		VitaminB12McgPer100g: 2.3456789,   // Should round to 2 decimal places
		ZincMgPer100g:        1.23456789,  // Should round to 2 decimal places
	}

	testItem := Item{
		Name:  "Test Food",
		Grams: 150.0, // 1.5x multiplier (150g serving)

	}
	result := service.convertCachedToNutrients(cached, testItem)

	// Check rounding precision - update expectations to match actual behavior
	assert.Equal(t, 185.185, result.Calories) // 123.456789 * 1.5 = 185.185... rounded to 3 decimal
	assert.Equal(t, 18.52, result.Protein)    // 12.3456789 * 1.5 = 18.518... rounded to 2 decimal
	assert.Equal(t, 0.185, result.Thiamine)   // 0.123456789 * 1.5 = 0.185... rounded to 3 decimals
	assert.Equal(t, 3.52, result.VitaminB12)  // 2.3456789 * 1.5 = 3.5185... rounded to 2 decimals
	assert.Equal(t, 1.85, result.Zinc)        // 1.23456789 * 1.5 = 1.8518... rounded to 2 decimals
}

// TestGeneratedNutrientIteration tests that generated helpers process all nutrient fields
func TestGeneratedNutrientIteration(t *testing.T) {
	// Test NutrientMeta contains expected nutrients
	assert.NotEmpty(t, NutrientMeta, "NutrientMeta should not be empty")
	
	// Check that we have the expected number of nutrients (should be 40+ nutrients)
	assert.Greater(t, len(NutrientMeta), 40, "Should have more than 40 nutrients")
	
	// Check that key nutrients exist
	foundCalories := false
	foundProtein := false
	foundVitaminB12 := false
	
	for _, nutrient := range NutrientMeta {
		switch nutrient.Key {
		case "calories":
			foundCalories = true
			assert.Equal(t, "Calories", nutrient.GoField)
			assert.Equal(t, 3, nutrient.Precision) // Calories should have 3 decimal precision
			assert.Equal(t, "kcal", nutrient.Unit)
		case "protein_g":
			foundProtein = true
			assert.Equal(t, "Protein", nutrient.GoField)
			assert.Equal(t, 2, nutrient.Precision)
			assert.Equal(t, "g", nutrient.Unit)
		case "vitamin_b12_mcg":
			foundVitaminB12 = true
			assert.Equal(t, "VitaminB12", nutrient.GoField)
			assert.Equal(t, 2, nutrient.Precision) // B12 should have 2 decimal precision
			assert.Equal(t, "mcg", nutrient.Unit)
		}
	}
	
	assert.True(t, foundCalories, "Should find calories in NutrientMeta")
	assert.True(t, foundProtein, "Should find protein in NutrientMeta")
	assert.True(t, foundVitaminB12, "Should find vitamin B12 in NutrientMeta")
}

// TestGeneratedScaleFunction tests the generated Scale method
func TestGeneratedScaleFunction(t *testing.T) {
	original := CompleteNutrient{
		Calories: 100,
		Protein:  20,
		TotalFat: 5,
		VitaminB12: 2.5,
	}
	
	// Scale by 2.0
	scaled := original
	scaled.Scale(2.0)
	
	assert.Equal(t, 200.0, scaled.Calories)
	assert.Equal(t, 40.0, scaled.Protein)
	assert.Equal(t, 10.0, scaled.TotalFat)
	assert.Equal(t, 5.0, scaled.VitaminB12)
	
	// Original should be unchanged
	assert.Equal(t, 100.0, original.Calories)
	assert.Equal(t, 20.0, original.Protein)
}

// TestGeneratedCopyFromFunction tests the generated CopyFrom method
func TestGeneratedCopyFromFunction(t *testing.T) {
	source := CompleteNutrient{
		Calories: 150,
		Protein:  25,
		TotalFat: 8,
		VitaminB12: 3.2,
	}
	
	var dest CompleteNutrient
	dest.CopyFrom(&source)
	
	assert.Equal(t, 150.0, dest.Calories)
	assert.Equal(t, 25.0, dest.Protein)
	assert.Equal(t, 8.0, dest.TotalFat)
	assert.Equal(t, 3.2, dest.VitaminB12)
}

// TestGeneratedConversions tests the generated conversion functions
func TestGeneratedConversions(t *testing.T) {
	service := NewNutritionService(nil)
	
	// Create a mock cached item with per-100g values
	cached := &storage.Item{
		CaloriesPer100g:      100,
		ProteinGPer100g:      20,
		VitaminB12McgPer100g: 2.5,
		OriginalServingGrams: Ptr(150.0), // 150g serving
		OriginalCalories:     Ptr(150.0),
		OriginalProteinG:     Ptr(30.0),
		OriginalVitaminB12Mcg: Ptr(3.75),
	}
	
	// Test ConvertPer100gToServing
	result := ConvertPer100gToServing(cached, 200.0, service.converter)
	assert.Equal(t, 200.0, result.Calories)   // 100 * 2.0
	assert.Equal(t, 40.0, result.Protein)     // 20 * 2.0
	assert.Equal(t, 5.0, result.VitaminB12)   // 2.5 * 2.0
	
	// Test ConvertExactCachedToNutrients (should use original values)
	exactResult := ConvertExactCachedToNutrients(cached)
	assert.Equal(t, 150.0, exactResult.Calories)   // Original serving values
	assert.Equal(t, 30.0, exactResult.Protein)
	assert.Equal(t, 3.75, exactResult.VitaminB12)
}

// TestPtrHelper tests the generic Ptr helper function
func TestPtrHelper(t *testing.T) {
	// Test with float64
	f := 42.5
	ptrF := Ptr(f)
	assert.NotNil(t, ptrF)
	assert.Equal(t, 42.5, *ptrF)
	
	// Test with string
	s := "test"
	ptrS := Ptr(s)
	assert.NotNil(t, ptrS)
	assert.Equal(t, "test", *ptrS)
	
	// Test with int
	i := 123
	ptrI := Ptr(i)
	assert.NotNil(t, ptrI)
	assert.Equal(t, 123, *ptrI)
}

// TestUtilityHelpers tests the new utility helper functions
func TestUtilityHelpers(t *testing.T) {
	// Test isFresh
	now := time.Now()
	assert.True(t, isFresh(now, time.Hour))
	assert.True(t, isFresh(now.Add(-30*time.Minute), time.Hour))
	assert.False(t, isFresh(now.Add(-2*time.Hour), time.Hour))
	
	// Test isGramUnit
	assert.True(t, isGramUnit(Ptr("g")))
	assert.True(t, isGramUnit(Ptr("grams")))
	assert.True(t, isGramUnit(Ptr("gram")))
	assert.False(t, isGramUnit(Ptr("kg")))
	assert.False(t, isGramUnit(Ptr("oz")))
	assert.False(t, isGramUnit(nil))
}
