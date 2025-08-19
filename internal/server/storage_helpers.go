package server

import (
	"encoding/json"

	"github.com/grantbirki/noot/internal/storage"
)

// itemWithNutritionToConsumption converts server response data to a consumption record
func itemWithNutritionToConsumption(userID string, transcript string, items []ItemWithNutrition, summary Summary) *storage.Consumption {
	// Serialize items to JSON
	itemsJSON, _ := json.Marshal(items)

	return &storage.Consumption{
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
