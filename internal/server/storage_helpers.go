package server

import (
	"encoding/json"

	"github.com/grantbirki/noot/internal/storage"
)

// itemWithNutritionToMeal converts server response data to a meal record
func itemWithNutritionToMeal(userID string, transcript string, items []ItemWithNutrition, summary Summary) *storage.Meal {
	// Serialize items to JSON
	itemsJSON, _ := json.Marshal(items)

	return &storage.Meal{
		UserID:        userID,
		Transcript:    transcript,
		ItemsJSON:     string(itemsJSON),
		TotalCalories: summary.Totals.Calories,
		TotalProtein:  summary.Totals.Protein,
		TotalFat:      summary.Totals.TotalFat,
		TotalCarbs:    summary.Totals.TotalCarbs,
		TotalFiber:    summary.Totals.DietaryFiber,
		TotalSodium:   summary.Totals.Sodium,
	}
}
