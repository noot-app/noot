package server

import (
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// convertInternalItemToAPI converts internal Item to API Item
func convertInternalItemToAPI(internal Item) api.Item {
	apiItem := api.Item{
		Name:  internal.Name,
		Brand: internal.Brand,
		Unit:  internal.Unit,
	}

	if internal.Quantity != nil {
		qty := float32(*internal.Quantity)
		apiItem.Quantity = &qty
	}

	if internal.Nutrients != nil {
		apiItem.Nutrients = convertInternalCompleteNutrientToAPI(*internal.Nutrients)
	}

	return apiItem
}

// convertInternalCompleteNutrientToAPI converts internal CompleteNutrient to API CompleteNutrient
func convertInternalCompleteNutrientToAPI(internal CompleteNutrient) *api.CompleteNutrient {
	return &api.CompleteNutrient{
		Calories:      float32(internal.Calories),
		ProteinG:      float32(internal.Protein),
		TotalFatG:     float32(internal.TotalFat),
		SaturatedFatG: float32(internal.SaturatedFat),
		TransFatG:     float32(internal.TransFat),
		CholesterolMg: float32(internal.Cholesterol),
		SodiumMg:      float32(internal.Sodium),
		TotalCarbsG:   float32(internal.TotalCarbs),
		DietaryFiberG: float32(internal.DietaryFiber),
		TotalSugarsG:  float32(internal.TotalSugars),
		AddedSugarsG:  float32(internal.AddedSugars),
		VitaminAMcg:   float32(internal.VitaminA),
		VitaminCMg:    float32(internal.VitaminC),
		VitaminDMcg:   float32(internal.VitaminD),
		VitaminEMg:    float32(internal.VitaminE),
		VitaminKMcg:   float32(internal.VitaminK),
		ThiamineMg:    float32(internal.Thiamine),
		RiboflavinMg:  float32(internal.Riboflavin),
		NiacinMg:      float32(internal.Niacin),
		VitaminB6Mg:   float32(internal.VitaminB6),
		FolateMcg:     float32(internal.Folate),
		VitaminB12Mcg: float32(internal.VitaminB12),
		CalciumMg:     float32(internal.Calcium),
		IronMg:        float32(internal.Iron),
		MagnesiumMg:   float32(internal.Magnesium),
		PhosphorusMg:  float32(internal.Phosphorus),
		PotassiumMg:   float32(internal.Potassium),
		ZincMg:        float32(internal.Zinc),
		CopperMg:      float32(internal.Copper),
		ManganeseMg:   float32(internal.Manganese),
		SeleniumMcg:   float32(internal.Selenium),
	}
}

// convertInternalSummaryToAPI converts internal Summary to API Summary
func convertInternalSummaryToAPI(internal Summary) api.Summary {
	return api.Summary{
		Totals:         *convertInternalCompleteNutrientToAPI(internal.Totals),
		PercentOfDaily: internal.PercentOfDaily,
		DailyValues:    convertMapFloat64ToFloat32(internal.DailyValuesUsed),
	}
}

// convertAPIItemsToInternal converts API ItemWithNutrition slice to internal ItemWithNutrition slice
func convertAPIItemsToInternal(apiItems []api.ItemWithNutrition) []ItemWithNutrition {
	var internal []ItemWithNutrition
	for _, apiItem := range apiItems {
		item := convertAPIItemToInternal(apiItem.Item)
		internalItem := ItemWithNutrition{
			Item: item,
		}
		if apiItem.Note != nil {
			internalItem.Note = *apiItem.Note
		}
		internal = append(internal, internalItem)
	}
	return internal
}

// convertAPIItemToInternal converts API Item to internal Item
func convertAPIItemToInternal(apiItem api.Item) Item {
	internal := Item{
		Name:  apiItem.Name,
		Brand: apiItem.Brand,
		Unit:  apiItem.Unit,
	}

	if apiItem.Quantity != nil {
		qty := float64(*apiItem.Quantity)
		internal.Quantity = &qty
	}

	if apiItem.Nutrients != nil {
		internal.Nutrients = convertAPICompleteNutrientToInternal(*apiItem.Nutrients)
	}

	return internal
}

// convertAPICompleteNutrientToInternal converts API CompleteNutrient to internal CompleteNutrient
func convertAPICompleteNutrientToInternal(api api.CompleteNutrient) *CompleteNutrient {
	return &CompleteNutrient{
		Calories:     float64(api.Calories),
		Protein:      float64(api.ProteinG),
		TotalFat:     float64(api.TotalFatG),
		SaturatedFat: float64(api.SaturatedFatG),
		TransFat:     float64(api.TransFatG),
		Cholesterol:  float64(api.CholesterolMg),
		Sodium:       float64(api.SodiumMg),
		TotalCarbs:   float64(api.TotalCarbsG),
		DietaryFiber: float64(api.DietaryFiberG),
		TotalSugars:  float64(api.TotalSugarsG),
		AddedSugars:  float64(api.AddedSugarsG),
		VitaminA:     float64(api.VitaminAMcg),
		VitaminC:     float64(api.VitaminCMg),
		VitaminD:     float64(api.VitaminDMcg),
		VitaminE:     float64(api.VitaminEMg),
		VitaminK:     float64(api.VitaminKMcg),
		Thiamine:     float64(api.ThiamineMg),
		Riboflavin:   float64(api.RiboflavinMg),
		Niacin:       float64(api.NiacinMg),
		VitaminB6:    float64(api.VitaminB6Mg),
		Folate:       float64(api.FolateMcg),
		VitaminB12:   float64(api.VitaminB12Mcg),
		Calcium:      float64(api.CalciumMg),
		Iron:         float64(api.IronMg),
		Magnesium:    float64(api.MagnesiumMg),
		Phosphorus:   float64(api.PhosphorusMg),
		Potassium:    float64(api.PotassiumMg),
		Zinc:         float64(api.ZincMg),
		Copper:       float64(api.CopperMg),
		Manganese:    float64(api.ManganeseMg),
		Selenium:     float64(api.SeleniumMcg),
	}
}

// convertInternalNutritionSummaryToAPI converts internal storage NutritionSummary to API NutritionSummary
func convertInternalNutritionSummaryToAPI(internal *storage.NutritionSummary) api.NutritionSummary {
	var dailyBreakdown []api.DailySummary
	for _, day := range internal.DailyBreakdown {
		dailyBreakdown = append(dailyBreakdown, api.DailySummary{
			Date:             day.Date,
			Calories:         float32(day.Calories),
			ProteinG:         float32(day.Protein),
			TotalCarbsG:      float32(day.Carbs),
			TotalFatG:        float32(day.Fat),
			FiberG:           float32(day.Fiber),
			SodiumMg:         float32(day.Sodium),
			ConsumptionCount: day.ConsumptionCount,
		})
	}

	return api.NutritionSummary{
		UserId:             internal.UserID,
		StartDate:          internal.StartDate,
		EndDate:            internal.EndDate,
		TotalCalories:      float32(internal.TotalCalories),
		TotalProteinG:      float32(internal.TotalProtein),
		TotalCarbsG:        float32(internal.TotalCarbs),
		TotalFatG:          float32(internal.TotalFat),
		TotalFiberG:        float32(internal.TotalFiber),
		TotalSodiumMg:      float32(internal.TotalSodium),
		TotalSaturatedFatG: float32(internal.TotalSaturatedFat),
		TotalTransFatG:     float32(internal.TotalTransFat),
		TotalCholesterolMg: float32(internal.TotalCholesterol),
		TotalSugarsG:       float32(internal.TotalSugars),
		TotalAddedSugarsG:  float32(internal.TotalAddedSugars),
		TotalVitaminAMcg:   float32(internal.TotalVitaminA),
		TotalVitaminCMg:    float32(internal.TotalVitaminC),
		TotalVitaminDMcg:   float32(internal.TotalVitaminD),
		TotalVitaminEMg:    float32(internal.TotalVitaminE),
		TotalVitaminKMcg:   float32(internal.TotalVitaminK),
		TotalThiamineMg:    float32(internal.TotalThiamine),
		TotalRiboflavinMg:  float32(internal.TotalRiboflavin),
		TotalNiacinMg:      float32(internal.TotalNiacin),
		TotalVitaminB6Mg:   float32(internal.TotalVitaminB6),
		TotalFolateMcg:     float32(internal.TotalFolate),
		TotalVitaminB12Mcg: float32(internal.TotalVitaminB12),
		TotalCalciumMg:     float32(internal.TotalCalcium),
		TotalIronMg:        float32(internal.TotalIron),
		TotalMagnesiumMg:   float32(internal.TotalMagnesium),
		TotalPhosphorusMg:  float32(internal.TotalPhosphorus),
		TotalPotassiumMg:   float32(internal.TotalPotassium),
		TotalZincMg:        float32(internal.TotalZinc),
		TotalCopperMg:      float32(internal.TotalCopper),
		TotalManganeseMg:   float32(internal.TotalManganese),
		TotalSeleniumMcg:   float32(internal.TotalSelenium),
		AvgCaloriesPerDay:  float32(internal.AvgCaloriesPerDay),
		AvgProteinPerDay:   float32(internal.AvgProteinPerDay),
		AvgCarbsPerDay:     float32(internal.AvgCarbsPerDay),
		AvgFatPerDay:       float32(internal.AvgFatPerDay),
		AvgFiberPerDay:     float32(internal.AvgFiberPerDay),
		AvgSodiumPerDay:    float32(internal.AvgSodiumPerDay),
		ConsumptionCount:   internal.ConsumptionCount,
		DailyBreakdown:     dailyBreakdown,
	}
}

// convertInternalUserToAPI converts internal storage User to API User
func convertInternalUserToAPI(internal *storage.User) api.User {
	subscriptionTier := api.Free
	if internal.SubscriptionTier == "pro" {
		subscriptionTier = api.Pro
	}

	return api.User{
		Id:               internal.ID,
		Email:            internal.Email,
		SubscriptionTier: subscriptionTier,
		CreatedAt:        internal.CreatedAt,
	}
}

// convertMapFloat64ToFloat32 converts map[string]float64 to map[string]float32
func convertMapFloat64ToFloat32(input map[string]float64) map[string]float32 {
	result := make(map[string]float32)
	for k, v := range input {
		result[k] = float32(v)
	}
	return result
}
