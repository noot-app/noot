package server

import (
	"context"

	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// CreateDatabaseConfig creates a database configuration from environment variables
// This eliminates duplication between main.go and server.go
func CreateDatabaseConfig() *storage.Config {
	config := &storage.Config{
		Type:     getenv("DATABASE_PROVIDER", "supabase"),
		Database: getenv("SUPABASE_DB_URL", ""),
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
func itemWithNutritionToConsumption(userID string, transcript string, summary Summary) *storage.Consumption {
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
		Iodine:     summary.Totals.Iodine,
		Molybdenum: summary.Totals.Molybdenum,
		Chromium:   summary.Totals.Chromium,
		Fluoride:   summary.Totals.Fluoride,
		Chloride:   summary.Totals.Chloride,
		// Additional vitamins
		Biotin:          summary.Totals.Biotin,
		PantothenicAcid: summary.Totals.PantothenicAcid,
		Choline:         summary.Totals.Choline,
		// New fatty acids
		MonounsaturatedFat: summary.Totals.MonounsaturatedFat,
		PolyunsaturatedFat: summary.Totals.PolyunsaturatedFat,
		Omega3Ala:          summary.Totals.Omega3Ala,
		Omega3Epa:          summary.Totals.Omega3Epa,
		Omega3Dha:          summary.Totals.Omega3Dha,
		Omega6:             summary.Totals.Omega6,
		// Functional compounds
		Alcohol:  summary.Totals.Alcohol,
		Caffeine: summary.Totals.Caffeine,
		Creatine: summary.Totals.Creatine,
	}
}

// apiItemWithNutritionToConsumptionItem converts API response data to a consumption item record
func apiItemWithNutritionToConsumptionItem(consumptionID string, item api.ItemWithNutrition, itemID *string) *storage.ConsumptionItem {
	brandStr := ""
	if item.Item.Brand != nil {
		brandStr = *item.Item.Brand
	}

	var nutrients api.CompleteNutrient
	if item.Item.Nutrients != nil {
		nutrients = *item.Item.Nutrients
	}

	var userQty *float64
	if item.Item.UserQuantity != nil {
		f64val := float64(*item.Item.UserQuantity)
		userQty = &f64val
	}

	return &storage.ConsumptionItem{
		ConsumptionID: consumptionID,
		ItemID:        itemID,
		Name:          item.Item.Name,
		Brand:         brandStr,
		Grams:         float64(item.Item.Grams),
		UserQuantity:  userQty,
		UserUnit:      item.Item.UserUnit,
		Note:          item.Item.Note,
		// Nutrition snapshot for this serving
		Calories:            float64(nutrients.Calories),
		ProteinG:            float64(nutrients.ProteinG),
		TotalFatG:           float64(nutrients.TotalFatG),
		SaturatedFatG:       float64(nutrients.SaturatedFatG),
		TransFatG:           float64(nutrients.TransFatG),
		CholesterolMg:       float64(nutrients.CholesterolMg),
		SodiumMg:            float64(nutrients.SodiumMg),
		TotalCarbsG:         float64(nutrients.TotalCarbsG),
		DietaryFiberG:       float64(nutrients.DietaryFiberG),
		TotalSugarsG:        float64(nutrients.TotalSugarsG),
		AddedSugarsG:        float64(nutrients.AddedSugarsG),
		VitaminAMcg:         float64(nutrients.VitaminAMcg),
		VitaminCMg:          float64(nutrients.VitaminCMg),
		VitaminDMcg:         float64(nutrients.VitaminDMcg),
		VitaminEMg:          float64(nutrients.VitaminEMg),
		VitaminKMcg:         float64(nutrients.VitaminKMcg),
		ThiamineMg:          float64(nutrients.ThiamineMg),
		RiboflavinMg:        float64(nutrients.RiboflavinMg),
		NiacinMg:            float64(nutrients.NiacinMg),
		VitaminB6Mg:         float64(nutrients.VitaminB6Mg),
		FolateMcg:           float64(nutrients.FolateMcg),
		VitaminB12Mcg:       float64(nutrients.VitaminB12Mcg),
		BiotinMcg:           float64(nutrients.BiotinMcg),
		PantothenicAcidMg:   float64(nutrients.PantothenicAcidMg),
		CholineMg:           float64(nutrients.CholineMg),
		CalciumMg:           float64(nutrients.CalciumMg),
		IronMg:              float64(nutrients.IronMg),
		MagnesiumMg:         float64(nutrients.MagnesiumMg),
		PhosphorusMg:        float64(nutrients.PhosphorusMg),
		PotassiumMg:         float64(nutrients.PotassiumMg),
		ZincMg:              float64(nutrients.ZincMg),
		CopperMg:            float64(nutrients.CopperMg),
		ManganeseMg:         float64(nutrients.ManganeseMg),
		SeleniumMcg:         float64(nutrients.SeleniumMcg),
		IodineMcg:           float64(nutrients.IodineMcg),
		MolybdenumMcg:       float64(nutrients.MolybdenumMcg),
		ChromiumMcg:         float64(nutrients.ChromiumMcg),
		FluorideMg:          float64(nutrients.FluorideMg),
		ChlorideMg:          float64(nutrients.ChlorideMg),
		Omega3AlaG:          float64(nutrients.Omega3AlaG),
		Omega3EpaG:          float64(nutrients.Omega3EpaG),
		Omega3DhaG:          float64(nutrients.Omega3DhaG),
		Omega6G:             float64(nutrients.Omega6G),
		CreatineMg:          float64(nutrients.CreatineMg),
		CaffeineMg:          float64(nutrients.CaffeineMg),
		AlcoholG:            float64(nutrients.AlcoholG),
		PolyunsaturatedFatG: float64(nutrients.PolyunsaturatedFatG),
		MonounsaturatedFatG: float64(nutrients.MonounsaturatedFatG),
		// Ingredients and OFF URL (historical snapshot)
		Ingredients: convertAPIIngredientsToStorage(item.Item.Ingredients), // Convert and copy ingredients from API item
		Url:         item.Item.Url,                                         // Copy OFF URL from API item
	}
}

// storageConsumptionToAPI converts storage.Consumption with its items to API format
func storageConsumptionToAPI(ctx context.Context, store storage.Store, consumption *storage.Consumption) (*api.Consumption, error) {
	// Get consumption items for historic breakdown
	consumptionItems, err := store.GetConsumptionItems(ctx, consumption.ID)
	if err != nil {
		LogError("Failed to get consumption items", err, "consumption_id", consumption.ID)
		// Continue without items if this fails - fallback to empty array
		consumptionItems = []*storage.ConsumptionItem{}
	}

	// Convert consumption items to API format
	apiItems := make([]api.ItemWithNutrition, len(consumptionItems))
	for i, ci := range consumptionItems {
		var brand *string
		if ci.Brand != "" {
			brand = &ci.Brand
		}

		nutrients := &api.CompleteNutrient{
			Calories:            float32(ci.Calories),
			ProteinG:            float32(ci.ProteinG),
			TotalFatG:           float32(ci.TotalFatG),
			SaturatedFatG:       float32(ci.SaturatedFatG),
			TransFatG:           float32(ci.TransFatG),
			CholesterolMg:       float32(ci.CholesterolMg),
			SodiumMg:            float32(ci.SodiumMg),
			TotalCarbsG:         float32(ci.TotalCarbsG),
			DietaryFiberG:       float32(ci.DietaryFiberG),
			TotalSugarsG:        float32(ci.TotalSugarsG),
			AddedSugarsG:        float32(ci.AddedSugarsG),
			VitaminAMcg:         float32(ci.VitaminAMcg),
			VitaminCMg:          float32(ci.VitaminCMg),
			VitaminDMcg:         float32(ci.VitaminDMcg),
			VitaminEMg:          float32(ci.VitaminEMg),
			VitaminKMcg:         float32(ci.VitaminKMcg),
			ThiamineMg:          float32(ci.ThiamineMg),
			RiboflavinMg:        float32(ci.RiboflavinMg),
			NiacinMg:            float32(ci.NiacinMg),
			VitaminB6Mg:         float32(ci.VitaminB6Mg),
			FolateMcg:           float32(ci.FolateMcg),
			VitaminB12Mcg:       float32(ci.VitaminB12Mcg),
			BiotinMcg:           float32(ci.BiotinMcg),
			PantothenicAcidMg:   float32(ci.PantothenicAcidMg),
			CholineMg:           float32(ci.CholineMg),
			CalciumMg:           float32(ci.CalciumMg),
			IronMg:              float32(ci.IronMg),
			MagnesiumMg:         float32(ci.MagnesiumMg),
			PhosphorusMg:        float32(ci.PhosphorusMg),
			PotassiumMg:         float32(ci.PotassiumMg),
			ZincMg:              float32(ci.ZincMg),
			CopperMg:            float32(ci.CopperMg),
			ManganeseMg:         float32(ci.ManganeseMg),
			SeleniumMcg:         float32(ci.SeleniumMcg),
			IodineMcg:           float32(ci.IodineMcg),
			MolybdenumMcg:       float32(ci.MolybdenumMcg),
			ChromiumMcg:         float32(ci.ChromiumMcg),
			FluorideMg:          float32(ci.FluorideMg),
			ChlorideMg:          float32(ci.ChlorideMg),
			Omega3AlaG:          float32(ci.Omega3AlaG),
			Omega3EpaG:          float32(ci.Omega3EpaG),
			Omega3DhaG:          float32(ci.Omega3DhaG),
			Omega6G:             float32(ci.Omega6G),
			CreatineMg:          float32(ci.CreatineMg),
			CaffeineMg:          float32(ci.CaffeineMg),
			AlcoholG:            float32(ci.AlcoholG),
			PolyunsaturatedFatG: float32(ci.PolyunsaturatedFatG),
			MonounsaturatedFatG: float32(ci.MonounsaturatedFatG),
		}

		var userQty *float32
		if ci.UserQuantity != nil {
			f32val := float32(*ci.UserQuantity)
			userQty = &f32val
		}

		apiItems[i] = api.ItemWithNutrition{
			Item: api.Item{
				Name:         ci.Name,
				Grams:        float32(ci.Grams),
				UserQuantity: userQty,
				UserUnit:     ci.UserUnit,
				Brand:        brand,
				Note:         ci.Note,
				Labels:       convertStorageLabelsToAPI(ci.Labels),
				Nutrients:    nutrients,
			},
		}
	}

	// Create summary from storage totals
	summary := api.Summary{
		Totals: api.CompleteNutrient{
			Calories:            float32(consumption.TotalCalories),
			ProteinG:            float32(consumption.TotalProtein),
			TotalFatG:           float32(consumption.TotalFat),
			SaturatedFatG:       float32(consumption.SaturatedFat),
			TransFatG:           float32(consumption.TransFat),
			CholesterolMg:       float32(consumption.Cholesterol),
			SodiumMg:            float32(consumption.TotalSodium),
			TotalCarbsG:         float32(consumption.TotalCarbs),
			DietaryFiberG:       float32(consumption.DietaryFiber),
			TotalSugarsG:        float32(consumption.TotalSugars),
			AddedSugarsG:        float32(consumption.AddedSugars),
			VitaminAMcg:         float32(consumption.VitaminA),
			VitaminCMg:          float32(consumption.VitaminC),
			VitaminDMcg:         float32(consumption.VitaminD),
			VitaminEMg:          float32(consumption.VitaminE),
			VitaminKMcg:         float32(consumption.VitaminK),
			ThiamineMg:          float32(consumption.Thiamine),
			RiboflavinMg:        float32(consumption.Riboflavin),
			NiacinMg:            float32(consumption.Niacin),
			VitaminB6Mg:         float32(consumption.VitaminB6),
			FolateMcg:           float32(consumption.Folate),
			VitaminB12Mcg:       float32(consumption.VitaminB12),
			BiotinMcg:           float32(consumption.Biotin),
			PantothenicAcidMg:   float32(consumption.PantothenicAcid),
			CholineMg:           float32(consumption.Choline),
			CalciumMg:           float32(consumption.Calcium),
			IronMg:              float32(consumption.Iron),
			MagnesiumMg:         float32(consumption.Magnesium),
			PhosphorusMg:        float32(consumption.Phosphorus),
			PotassiumMg:         float32(consumption.Potassium),
			ZincMg:              float32(consumption.Zinc),
			CopperMg:            float32(consumption.Copper),
			ManganeseMg:         float32(consumption.Manganese),
			SeleniumMcg:         float32(consumption.Selenium),
			IodineMcg:           float32(consumption.Iodine),
			MolybdenumMcg:       float32(consumption.Molybdenum),
			ChromiumMcg:         float32(consumption.Chromium),
			FluorideMg:          float32(consumption.Fluoride),
			ChlorideMg:          float32(consumption.Chloride),
			Omega3AlaG:          float32(consumption.Omega3Ala),
			Omega3EpaG:          float32(consumption.Omega3Epa),
			Omega3DhaG:          float32(consumption.Omega3Dha),
			Omega6G:             float32(consumption.Omega6),
			CreatineMg:          float32(consumption.Creatine),
			CaffeineMg:          float32(consumption.Caffeine),
			AlcoholG:            float32(consumption.Alcohol),
			PolyunsaturatedFatG: float32(consumption.PolyunsaturatedFat),
			MonounsaturatedFatG: float32(consumption.MonounsaturatedFat),
		},
	}

	return &api.Consumption{
		Id:         consumption.ID,
		UserId:     consumption.UserID,
		Transcript: consumption.Transcript,
		Note:       consumption.Note,
		Labels:     convertStorageLabelsToAPI(consumption.Labels),
		Items:      apiItems,
		Summary:    summary,
		CreatedAt:  consumption.ConsumedAt,
	}, nil
}

// convertAPIIngredientsToStorage converts API ingredients to storage format
func convertAPIIngredientsToStorage(apiIngredients *[]api.OFFIngredient) []storage.OFFIngredient {
	if apiIngredients == nil {
		return nil
	}

	storageIngredients := make([]storage.OFFIngredient, len(*apiIngredients))
	for i, apiIngredient := range *apiIngredients {
		storageIngredient := storage.OFFIngredient{}

		if apiIngredient.Id != nil {
			storageIngredient.ID = *apiIngredient.Id
		}
		if apiIngredient.Text != nil {
			storageIngredient.Text = *apiIngredient.Text
		}
		if apiIngredient.PercentEstimate != nil {
			f64 := float64(*apiIngredient.PercentEstimate)
			storageIngredient.PercentEstimate = &f64
		}
		if apiIngredient.PercentMax != nil {
			f64 := float64(*apiIngredient.PercentMax)
			storageIngredient.PercentMax = &f64
		}
		if apiIngredient.PercentMin != nil {
			f64 := float64(*apiIngredient.PercentMin)
			storageIngredient.PercentMin = &f64
		}

		storageIngredients[i] = storageIngredient
	}

	return storageIngredients
}
