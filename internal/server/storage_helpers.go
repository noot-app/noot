package server

import (
	"github.com/grantbirki/noot/internal/storage"
)

// CreateDatabaseConfig creates a database configuration from environment variables
// This eliminates duplication between main.go and server.go
func CreateDatabaseConfig() *storage.Config {
	config := &storage.Config{
		Type:     getenv("DATABASE_PROVIDER", "sqlite"),
		Database: getenv("DATABASE_PATH", "./noot.db"),
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenvInt("DB_PORT", 5432),
		Username: getenv("DB_USER", ""),
		Password: getenv("DB_PASS", ""),
		SSLMode:  getenv("DB_SSLMODE", "prefer"),
	}

	// For Supabase, use SUPABASE_DB_URL if provided
	if config.Type == "supabase" && getenv("SUPABASE_DB_URL", "") != "" {
		config.Database = getenv("SUPABASE_DB_URL", "")
	}

	return config
}

// itemWithNutritionToConsumption converts server response data to a consumption record
func itemWithNutritionToConsumption(userID string, transcript string, items []ItemWithNutrition, summary Summary) *storage.Consumption {
	return &storage.Consumption{
		UserID:        userID,
		Transcript:    transcript,
		TotalCalories: summary.Totals.Calories,
		TotalProtein:  summary.Totals.Protein,
		TotalFat:      summary.Totals.TotalFat,
		TotalCarbs:    summary.Totals.TotalCarbs,
		DietaryFiber:  summary.Totals.DietaryFiber,
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
