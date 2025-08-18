package server

import (
	"context"
	"strings"
	"time"
)

// Basic nutrition lookup for common foods
// This is a simple implementation to maintain functionality while improving OpenAI performance
// In the future, this could be replaced with proper USDA FDC API integration

var nutritionDatabase = map[string]CompleteNutrient{
	// Fruits
	"banana": {
		Calories: 105, Protein: 1.3, TotalFat: 0.4, TotalCarbs: 27,
		DietaryFiber: 3.1, TotalSugars: 14.4, VitaminC: 10.3, Potassium: 422,
		Calcium: 5, Iron: 0.3, Magnesium: 32,
	},
	"apple": {
		Calories: 95, Protein: 0.5, TotalFat: 0.3, TotalCarbs: 25,
		DietaryFiber: 4.4, TotalSugars: 19, VitaminC: 8.4, Potassium: 195,
		Calcium: 11, Iron: 0.2, Magnesium: 9,
	},
	"orange": {
		Calories: 62, Protein: 1.2, TotalFat: 0.2, TotalCarbs: 15.4,
		DietaryFiber: 3.1, TotalSugars: 12.2, VitaminC: 70, Potassium: 237,
		Calcium: 52, Iron: 0.1, Magnesium: 13,
	},

	// Proteins
	"egg": {
		Calories: 70, Protein: 6, TotalFat: 5, SaturatedFat: 1.6,
		Cholesterol: 186, Sodium: 70, TotalCarbs: 0.6,
		VitaminA: 270, VitaminD: 1.1, VitaminB12: 0.6, Selenium: 15.4,
	},
	"chicken breast": {
		Calories: 165, Protein: 31, TotalFat: 3.6, SaturatedFat: 1,
		Cholesterol: 85, Sodium: 74, Niacin: 14.8,
		Phosphorus: 228, Selenium: 27.6,
	},

	// Dairy
	"greek yogurt": {
		Calories: 130, Protein: 11, TotalFat: 5, SaturatedFat: 3.3,
		Cholesterol: 20, Sodium: 65, TotalCarbs: 9, TotalSugars: 9,
		Calcium: 150, VitaminB12: 0.8, Riboflavin: 0.3,
	},
	"milk": {
		Calories: 150, Protein: 8, TotalFat: 8, SaturatedFat: 4.6,
		Cholesterol: 24, Sodium: 105, TotalCarbs: 12, TotalSugars: 12,
		Calcium: 276, VitaminA: 395, VitaminD: 2.9, VitaminB12: 1.1,
	},

	// Grains
	"bread": {
		Calories: 80, Protein: 4, TotalFat: 1, SaturatedFat: 0.2,
		Sodium: 160, TotalCarbs: 15, DietaryFiber: 2, TotalSugars: 2,
		Iron: 1, Thiamine: 0.1, Niacin: 1.3, Folate: 43,
	},
	"rice": {
		Calories: 205, Protein: 4.3, TotalFat: 0.4, TotalCarbs: 45,
		DietaryFiber: 0.6, Sodium: 2, Niacin: 2.3,
		Magnesium: 19, Manganese: 1.1,
	},

	// Beverages
	"coffee": {
		Calories: 5, Protein: 0.3, TotalFat: 0, TotalCarbs: 1,
		Riboflavin: 0.2, Niacin: 0.5, Potassium: 116,
		Magnesium: 7,
	},
	"latte": {
		Calories: 190, Protein: 9, TotalFat: 7, SaturatedFat: 4.4,
		Cholesterol: 25, Sodium: 115, TotalCarbs: 19, TotalSugars: 18,
		Calcium: 290, VitaminA: 318, VitaminD: 2.4,
	},
}

// lookupNutrition attempts to find nutrition data for a food item
func lookupNutrition(ctx context.Context, foodName string) *CompleteNutrient {
	startTime := time.Now()
	defer func() {
		LogDebug("Nutrition lookup time", "duration_ms", time.Since(startTime).Milliseconds())
	}()

	// Normalize the food name for lookup
	normalized := strings.ToLower(strings.TrimSpace(foodName))

	// Direct lookup
	if nutrition, found := nutritionDatabase[normalized]; found {
		LogDebug("Nutrition found via direct lookup", "food", foodName)
		return &nutrition
	}

	// Fuzzy matching for common variations
	for key, nutrition := range nutritionDatabase {
		if strings.Contains(normalized, key) || strings.Contains(key, normalized) {
			LogDebug("Nutrition found via fuzzy match", "food", foodName, "matched", key)
			return &nutrition
		}
	}

	// Check for plurals (simple approach)
	if strings.HasSuffix(normalized, "s") {
		singular := normalized[:len(normalized)-1]
		if nutrition, found := nutritionDatabase[singular]; found {
			LogDebug("Nutrition found via singular form", "food", foodName, "matched", singular)
			return &nutrition
		}
	}

	LogDebug("No nutrition data found", "food", foodName)
	return nil
}

// enrichItemsWithNutrition adds nutrition data to parsed items
func enrichItemsWithNutrition(ctx context.Context, items []Item) []Item {
	startTime := time.Now()
	defer func() {
		LogDebug("Nutrition enrichment total time", "duration_ms", time.Since(startTime).Milliseconds(), "item_count", len(items))
	}()

	enriched := make([]Item, len(items))
	for i, item := range items {
		enriched[i] = item

		// Skip if already has nutrition data (shouldn't happen with new approach, but defensive)
		if item.Nutrients != nil {
			continue
		}

		// Look up nutrition data
		if nutrition := lookupNutrition(ctx, item.Name); nutrition != nil {
			// Scale nutrition data based on quantity if available
			scaled := *nutrition
			if item.Quantity != nil && *item.Quantity != 1.0 {
				scaleNutrition(&scaled, *item.Quantity)
			}
			enriched[i].Nutrients = &scaled
		}
	}

	return enriched
}

// scaleNutrition multiplies all nutrition values by the given factor
func scaleNutrition(nutrition *CompleteNutrient, factor float64) {
	nutrition.Calories *= factor
	nutrition.Protein *= factor
	nutrition.TotalFat *= factor
	nutrition.SaturatedFat *= factor
	nutrition.TransFat *= factor
	nutrition.Cholesterol *= factor
	nutrition.Sodium *= factor
	nutrition.TotalCarbs *= factor
	nutrition.DietaryFiber *= factor
	nutrition.TotalSugars *= factor
	nutrition.AddedSugars *= factor
	nutrition.VitaminA *= factor
	nutrition.VitaminC *= factor
	nutrition.VitaminD *= factor
	nutrition.VitaminE *= factor
	nutrition.VitaminK *= factor
	nutrition.Thiamine *= factor
	nutrition.Riboflavin *= factor
	nutrition.Niacin *= factor
	nutrition.VitaminB6 *= factor
	nutrition.Folate *= factor
	nutrition.VitaminB12 *= factor
	nutrition.Calcium *= factor
	nutrition.Iron *= factor
	nutrition.Magnesium *= factor
	nutrition.Phosphorus *= factor
	nutrition.Potassium *= factor
	nutrition.Zinc *= factor
	nutrition.Copper *= factor
	nutrition.Manganese *= factor
	nutrition.Selenium *= factor
}
