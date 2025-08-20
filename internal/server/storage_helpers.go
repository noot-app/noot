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
		// Additional micronutrients
		SaturatedFat: summary.Totals.SaturatedFat,
		TransFat:     summary.Totals.TransFat,
		Cholesterol:  summary.Totals.Cholesterol,
		TotalSugars:  summary.Totals.TotalSugars,
		AddedSugars:  summary.Totals.AddedSugars,
		// Vitamins
		VitaminA:   summary.Totals.VitaminA,
		VitaminC:   summary.Totals.VitaminC,
		VitaminD:   summary.Totals.VitaminD,
		VitaminE:   summary.Totals.VitaminE,
		VitaminK:   summary.Totals.VitaminK,
		Thiamine:   summary.Totals.Thiamine,
		Riboflavin: summary.Totals.Riboflavin,
		Niacin:     summary.Totals.Niacin,
		VitaminB6:  summary.Totals.VitaminB6,
		Folate:     summary.Totals.Folate,
		VitaminB12: summary.Totals.VitaminB12,
		// Minerals
		Calcium:    summary.Totals.Calcium,
		Iron:       summary.Totals.Iron,
		Magnesium:  summary.Totals.Magnesium,
		Phosphorus: summary.Totals.Phosphorus,
		Potassium:  summary.Totals.Potassium,
		Zinc:       summary.Totals.Zinc,
		Copper:     summary.Totals.Copper,
		Manganese:  summary.Totals.Manganese,
		Selenium:   summary.Totals.Selenium,
	}
}
