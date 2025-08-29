package server

import (
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

// itemWithNutritionToConsumptionItem converts an API ItemWithNutrition to a ConsumptionItem with nutrition snapshot
func itemWithNutritionToConsumptionItem(consumptionID string, item api.ItemWithNutrition) *storage.ConsumptionItem {
	// Get brand value, defaulting to empty string if nil
	brand := ""
	if item.Item.Brand != nil {
		brand = *item.Item.Brand
	}

	consumptionItem := &storage.ConsumptionItem{
		ConsumptionID: consumptionID,
		ItemID:        nil, // TODO: Link to global items cache if available
		Name:          item.Item.Name,
		Brand:         brand,
		Grams:         float64(item.Item.Grams),
		UserQuantity:  convertFloat32PtrToFloat64Ptr(item.Item.UserQuantity),
		UserUnit:      item.Item.UserUnit,
		Label:         item.Item.Label,
	}

	// Set note from ItemWithNutrition level or Item level
	if item.Note != nil && *item.Note != "" {
		consumptionItem.Note = item.Note
	} else if item.Item.Note != nil {
		consumptionItem.Note = item.Item.Note
	}

	// If the item has nutrients, snapshot them
	if item.Item.Nutrients != nil {
		nutrients := item.Item.Nutrients
		consumptionItem.Calories = float64(nutrients.Calories)
		consumptionItem.ProteinG = float64(nutrients.ProteinG)
		consumptionItem.TotalFatG = float64(nutrients.TotalFatG)
		consumptionItem.SaturatedFatG = float64(nutrients.SaturatedFatG)
		consumptionItem.TransFatG = float64(nutrients.TransFatG)
		consumptionItem.CholesterolMg = float64(nutrients.CholesterolMg)
		consumptionItem.SodiumMg = float64(nutrients.SodiumMg)
		consumptionItem.TotalCarbsG = float64(nutrients.TotalCarbsG)
		consumptionItem.DietaryFiberG = float64(nutrients.DietaryFiberG)
		consumptionItem.TotalSugarsG = float64(nutrients.TotalSugarsG)
		consumptionItem.AddedSugarsG = float64(nutrients.AddedSugarsG)
		consumptionItem.VitaminAMcg = float64(nutrients.VitaminAMcg)
		consumptionItem.VitaminCMg = float64(nutrients.VitaminCMg)
		consumptionItem.VitaminDMcg = float64(nutrients.VitaminDMcg)
		consumptionItem.VitaminEMg = float64(nutrients.VitaminEMg)
		consumptionItem.VitaminKMcg = float64(nutrients.VitaminKMcg)
		consumptionItem.ThiamineMg = float64(nutrients.ThiamineMg)
		consumptionItem.RiboflavinMg = float64(nutrients.RiboflavinMg)
		consumptionItem.NiacinMg = float64(nutrients.NiacinMg)
		consumptionItem.VitaminB6Mg = float64(nutrients.VitaminB6Mg)
		consumptionItem.FolateMcg = float64(nutrients.FolateMcg)
		consumptionItem.VitaminB12Mcg = float64(nutrients.VitaminB12Mcg)
		consumptionItem.BiotinMcg = float64(nutrients.BiotinMcg)
		consumptionItem.PantothenicAcidMg = float64(nutrients.PantothenicAcidMg)
		consumptionItem.CholineMg = float64(nutrients.CholineMg)
		consumptionItem.CalciumMg = float64(nutrients.CalciumMg)
		consumptionItem.IronMg = float64(nutrients.IronMg)
		consumptionItem.MagnesiumMg = float64(nutrients.MagnesiumMg)
		consumptionItem.PhosphorusMg = float64(nutrients.PhosphorusMg)
		consumptionItem.PotassiumMg = float64(nutrients.PotassiumMg)
		consumptionItem.ZincMg = float64(nutrients.ZincMg)
		consumptionItem.CopperMg = float64(nutrients.CopperMg)
		consumptionItem.ManganeseMg = float64(nutrients.ManganeseMg)
		consumptionItem.SeleniumMcg = float64(nutrients.SeleniumMcg)
		consumptionItem.IodineMcg = float64(nutrients.IodineMcg)
		consumptionItem.MolybdenumMcg = float64(nutrients.MolybdenumMcg)
		consumptionItem.ChromiumMcg = float64(nutrients.ChromiumMcg)
		consumptionItem.FluorideMg = float64(nutrients.FluorideMg)
		consumptionItem.ChlorideMg = float64(nutrients.ChlorideMg)
		consumptionItem.Omega3AlaG = float64(nutrients.Omega3AlaG)
		consumptionItem.Omega3EpaG = float64(nutrients.Omega3EpaG)
		consumptionItem.Omega3DhaG = float64(nutrients.Omega3DhaG)
		consumptionItem.Omega6G = float64(nutrients.Omega6G)
		consumptionItem.CreatineMg = float64(nutrients.CreatineMg)
		consumptionItem.CaffeineMg = float64(nutrients.CaffeineMg)
		consumptionItem.AlcoholG = float64(nutrients.AlcoholG)
		consumptionItem.PolyunsaturatedFatG = float64(nutrients.PolyunsaturatedFatG)
		consumptionItem.MonounsaturatedFatG = float64(nutrients.MonounsaturatedFatG)
	}

	return consumptionItem
}

// convertFloat32PtrToFloat64Ptr converts *float32 to *float64
func convertFloat32PtrToFloat64Ptr(f32 *float32) *float64 {
	if f32 == nil {
		return nil
	}
	f64 := float64(*f32)
	return &f64
}

// convertFloat64PtrToFloat32Ptr converts *float64 to *float32
func convertFloat64PtrToFloat32Ptr(f64 *float64) *float32 {
	if f64 == nil {
		return nil
	}
	f32 := float32(*f64)
	return &f32
}

// convertConsumptionItemToAPIItemWithNutrition converts storage.ConsumptionItem to api.ItemWithNutrition
func convertConsumptionItemToAPIItemWithNutrition(item *storage.ConsumptionItem) api.ItemWithNutrition {
	// Convert brand string to pointer (API expects *string for brand)
	var brand *string
	if item.Brand != "" {
		brand = &item.Brand
	}

	return api.ItemWithNutrition{
		Item: api.Item{
			Name:         item.Name,
			Brand:        brand,
			Grams:        float32(item.Grams),
			UserQuantity: convertFloat64PtrToFloat32Ptr(item.UserQuantity),
			UserUnit:     item.UserUnit,
			Label:        item.Label,
			Note:         item.Note,
			Nutrients: &api.CompleteNutrient{
				Calories:                float32(item.Calories),
				ProteinG:               float32(item.ProteinG),
				TotalFatG:              float32(item.TotalFatG),
				SaturatedFatG:          float32(item.SaturatedFatG),
				TransFatG:              float32(item.TransFatG),
				CholesterolMg:          float32(item.CholesterolMg),
				SodiumMg:               float32(item.SodiumMg),
				TotalCarbsG:            float32(item.TotalCarbsG),
				DietaryFiberG:          float32(item.DietaryFiberG),
				TotalSugarsG:           float32(item.TotalSugarsG),
				AddedSugarsG:           float32(item.AddedSugarsG),
				VitaminAMcg:            float32(item.VitaminAMcg),
				VitaminCMg:             float32(item.VitaminCMg),
				VitaminDMcg:            float32(item.VitaminDMcg),
				VitaminEMg:             float32(item.VitaminEMg),
				VitaminKMcg:            float32(item.VitaminKMcg),
				ThiamineMg:             float32(item.ThiamineMg),
				RiboflavinMg:           float32(item.RiboflavinMg),
				NiacinMg:               float32(item.NiacinMg),
				VitaminB6Mg:            float32(item.VitaminB6Mg),
				FolateMcg:              float32(item.FolateMcg),
				VitaminB12Mcg:          float32(item.VitaminB12Mcg),
				BiotinMcg:              float32(item.BiotinMcg),
				PantothenicAcidMg:      float32(item.PantothenicAcidMg),
				CholineMg:              float32(item.CholineMg),
				CalciumMg:              float32(item.CalciumMg),
				IronMg:                 float32(item.IronMg),
				MagnesiumMg:            float32(item.MagnesiumMg),
				PhosphorusMg:           float32(item.PhosphorusMg),
				PotassiumMg:            float32(item.PotassiumMg),
				ZincMg:                 float32(item.ZincMg),
				CopperMg:               float32(item.CopperMg),
				ManganeseMg:            float32(item.ManganeseMg),
				SeleniumMcg:            float32(item.SeleniumMcg),
				IodineMcg:              float32(item.IodineMcg),
				MolybdenumMcg:          float32(item.MolybdenumMcg),
				ChromiumMcg:            float32(item.ChromiumMcg),
				FluorideMg:             float32(item.FluorideMg),
				ChlorideMg:             float32(item.ChlorideMg),
				Omega3AlaG:             float32(item.Omega3AlaG),
				Omega3EpaG:             float32(item.Omega3EpaG),
				Omega3DhaG:             float32(item.Omega3DhaG),
				Omega6G:                float32(item.Omega6G),
				CreatineMg:             float32(item.CreatineMg),
				CaffeineMg:             float32(item.CaffeineMg),
				AlcoholG:               float32(item.AlcoholG),
				PolyunsaturatedFatG:    float32(item.PolyunsaturatedFatG),
				MonounsaturatedFatG:    float32(item.MonounsaturatedFatG),
			},
		},
		Note: nil, // Item-level note is in Item.Note, not at ItemWithNutrition level
	}
}

// convertStorageConsumptionToAPI converts storage.Consumption and its items to api.Consumption
func convertStorageConsumptionToAPI(consumption *storage.Consumption, items []*storage.ConsumptionItem) api.Consumption {
	// Convert consumption items
	apiItems := make([]api.ItemWithNutrition, len(items))
	for i, item := range items {
		apiItems[i] = convertConsumptionItemToAPIItemWithNutrition(item)
	}

	// Build summary from consumption totals
	summary := api.Summary{
		Totals: api.CompleteNutrient{
			Calories:                float32(consumption.TotalCalories),
			ProteinG:               float32(consumption.TotalProtein),
			TotalFatG:              float32(consumption.TotalFat),
			SaturatedFatG:          float32(consumption.SaturatedFat),
			TransFatG:              float32(consumption.TransFat),
			CholesterolMg:          float32(consumption.Cholesterol),
			SodiumMg:               float32(consumption.TotalSodium),
			TotalCarbsG:            float32(consumption.TotalCarbs),
			DietaryFiberG:          float32(consumption.DietaryFiber),
			TotalSugarsG:           float32(consumption.TotalSugars),
			AddedSugarsG:           float32(consumption.AddedSugars),
			VitaminAMcg:            float32(consumption.VitaminA),
			VitaminCMg:             float32(consumption.VitaminC),
			VitaminDMcg:            float32(consumption.VitaminD),
			VitaminEMg:             float32(consumption.VitaminE),
			VitaminKMcg:            float32(consumption.VitaminK),
			ThiamineMg:             float32(consumption.Thiamine),
			RiboflavinMg:           float32(consumption.Riboflavin),
			NiacinMg:               float32(consumption.Niacin),
			VitaminB6Mg:            float32(consumption.VitaminB6),
			FolateMcg:              float32(consumption.Folate),
			VitaminB12Mcg:          float32(consumption.VitaminB12),
			BiotinMcg:              float32(consumption.Biotin),
			PantothenicAcidMg:      float32(consumption.PantothenicAcid),
			CholineMg:              float32(consumption.Choline),
			CalciumMg:              float32(consumption.Calcium),
			IronMg:                 float32(consumption.Iron),
			MagnesiumMg:            float32(consumption.Magnesium),
			PhosphorusMg:           float32(consumption.Phosphorus),
			PotassiumMg:            float32(consumption.Potassium),
			ZincMg:                 float32(consumption.Zinc),
			CopperMg:               float32(consumption.Copper),
			ManganeseMg:            float32(consumption.Manganese),
			SeleniumMcg:            float32(consumption.Selenium),
			IodineMcg:              float32(consumption.Iodine),
			MolybdenumMcg:          float32(consumption.Molybdenum),
			ChromiumMcg:            float32(consumption.Chromium),
			FluorideMg:             float32(consumption.Fluoride),
			ChlorideMg:             float32(consumption.Chloride),
			Omega3AlaG:             float32(consumption.Omega3Ala),
			Omega3EpaG:             float32(consumption.Omega3Epa),
			Omega3DhaG:             float32(consumption.Omega3Dha),
			Omega6G:                float32(consumption.Omega6),
			CreatineMg:             float32(consumption.Creatine),
			CaffeineMg:             float32(consumption.Caffeine),
			AlcoholG:               float32(consumption.Alcohol),
			PolyunsaturatedFatG:    float32(consumption.PolyunsaturatedFat),
			MonounsaturatedFatG:    float32(consumption.MonounsaturatedFat),
		},
		// TODO: Add PercentOfDaily calculations based on user goals
		PercentOfDaily: make(map[string]int),
		DailyValues:    make(map[string]float32),
	}

	return api.Consumption{
		Id:        consumption.ID,
		CreatedAt: consumption.CreatedAt,
		Transcript: consumption.Transcript,
		Items:     apiItems,
		Summary:   summary,
		Label:     consumption.Label,
		Note:      consumption.Note,
	}
}
