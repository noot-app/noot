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
		Grams: float32(internal.Grams),
	}

	// Include user display information if available
	if internal.UserQuantity != nil {
		qty := float32(*internal.UserQuantity)
		apiItem.UserQuantity = &qty
	}
	if internal.UserUnit != nil {
		apiItem.UserUnit = internal.UserUnit
	}

	if internal.Nutrients != nil {
		apiItem.Nutrients = convertInternalCompleteNutrientToAPI(*internal.Nutrients)
	}

	// Convert ingredients from storage format to API format
	if len(internal.Ingredients) > 0 {
		apiIngredients := make([]api.OFFIngredient, len(internal.Ingredients))
		for i, storageIngredient := range internal.Ingredients {
			apiIngredient := api.OFFIngredient{}

			if storageIngredient.ID != "" {
				apiIngredient.Id = &storageIngredient.ID
			}
			if storageIngredient.Text != "" {
				apiIngredient.Text = &storageIngredient.Text
			}
			if storageIngredient.PercentEstimate != nil {
				f32 := float32(*storageIngredient.PercentEstimate)
				apiIngredient.PercentEstimate = &f32
			}
			if storageIngredient.PercentMax != nil {
				f32 := float32(*storageIngredient.PercentMax)
				apiIngredient.PercentMax = &f32
			}
			if storageIngredient.PercentMin != nil {
				f32 := float32(*storageIngredient.PercentMin)
				apiIngredient.PercentMin = &f32
			}

			apiIngredients[i] = apiIngredient
		}
		apiItem.Ingredients = &apiIngredients
	}

	// Copy OFF URL
	if internal.OFFUrl != nil {
		apiItem.OffUrl = internal.OFFUrl
	}

	return apiItem
}

// convertInternalCompleteNutrientToAPI converts internal CompleteNutrient to API CompleteNutrient
func convertInternalCompleteNutrientToAPI(internal CompleteNutrient) *api.CompleteNutrient {
	return &api.CompleteNutrient{
		Calories:            RoundCaloriesUp(internal.Calories),
		ProteinG:            float32(internal.Protein),
		TotalFatG:           float32(internal.TotalFat),
		SaturatedFatG:       float32(internal.SaturatedFat),
		TransFatG:           float32(internal.TransFat),
		CholesterolMg:       float32(internal.Cholesterol),
		SodiumMg:            float32(internal.Sodium),
		TotalCarbsG:         float32(internal.TotalCarbs),
		DietaryFiberG:       float32(internal.DietaryFiber),
		TotalSugarsG:        float32(internal.TotalSugars),
		AddedSugarsG:        float32(internal.AddedSugars),
		VitaminAMcg:         float32(internal.VitaminA),
		VitaminCMg:          float32(internal.VitaminC),
		VitaminDMcg:         float32(internal.VitaminD),
		VitaminEMg:          float32(internal.VitaminE),
		VitaminKMcg:         float32(internal.VitaminK),
		ThiamineMg:          float32(internal.Thiamine),
		RiboflavinMg:        float32(internal.Riboflavin),
		NiacinMg:            float32(internal.Niacin),
		VitaminB6Mg:         float32(internal.VitaminB6),
		FolateMcg:           float32(internal.Folate),
		VitaminB12Mcg:       float32(internal.VitaminB12),
		CalciumMg:           float32(internal.Calcium),
		IronMg:              float32(internal.Iron),
		MagnesiumMg:         float32(internal.Magnesium),
		PhosphorusMg:        float32(internal.Phosphorus),
		PotassiumMg:         float32(internal.Potassium),
		ZincMg:              float32(internal.Zinc),
		CopperMg:            float32(internal.Copper),
		ManganeseMg:         float32(internal.Manganese),
		SeleniumMcg:         float32(internal.Selenium),
		IodineMcg:           float32(internal.Iodine),
		MolybdenumMcg:       float32(internal.Molybdenum),
		ChromiumMcg:         float32(internal.Chromium),
		FluorideMg:          float32(internal.Fluoride),
		ChlorideMg:          float32(internal.Chloride),
		BiotinMcg:           float32(internal.Biotin),
		PantothenicAcidMg:   float32(internal.PantothenicAcid),
		CholineMg:           float32(internal.Choline),
		MonounsaturatedFatG: float32(internal.MonounsaturatedFat),
		PolyunsaturatedFatG: float32(internal.PolyunsaturatedFat),
		Omega3AlaG:          float32(internal.Omega3Ala),
		Omega3EpaG:          float32(internal.Omega3Epa),
		Omega3DhaG:          float32(internal.Omega3Dha),
		Omega6G:             float32(internal.Omega6),
		AlcoholG:            float32(internal.Alcohol),
		CaffeineMg:          float32(internal.Caffeine),
		CreatineMg:          float32(internal.Creatine),
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
		Grams: float64(apiItem.Grams),
	}

	if apiItem.UserQuantity != nil {
		qty := float64(*apiItem.UserQuantity)
		internal.UserQuantity = &qty
	}
	if apiItem.UserUnit != nil {
		internal.UserUnit = apiItem.UserUnit
	}

	if apiItem.Nutrients != nil {
		internal.Nutrients = convertAPICompleteNutrientToInternal(*apiItem.Nutrients)
	}

	return internal
}

// convertAPICompleteNutrientToInternal converts API CompleteNutrient to internal CompleteNutrient
func convertAPICompleteNutrientToInternal(api api.CompleteNutrient) *CompleteNutrient {
	return &CompleteNutrient{
		Calories:           float64(api.Calories),
		Protein:            float64(api.ProteinG),
		TotalFat:           float64(api.TotalFatG),
		SaturatedFat:       float64(api.SaturatedFatG),
		TransFat:           float64(api.TransFatG),
		Cholesterol:        float64(api.CholesterolMg),
		Sodium:             float64(api.SodiumMg),
		TotalCarbs:         float64(api.TotalCarbsG),
		DietaryFiber:       float64(api.DietaryFiberG),
		TotalSugars:        float64(api.TotalSugarsG),
		AddedSugars:        float64(api.AddedSugarsG),
		VitaminA:           float64(api.VitaminAMcg),
		VitaminC:           float64(api.VitaminCMg),
		VitaminD:           float64(api.VitaminDMcg),
		VitaminE:           float64(api.VitaminEMg),
		VitaminK:           float64(api.VitaminKMcg),
		Thiamine:           float64(api.ThiamineMg),
		Riboflavin:         float64(api.RiboflavinMg),
		Niacin:             float64(api.NiacinMg),
		VitaminB6:          float64(api.VitaminB6Mg),
		Folate:             float64(api.FolateMcg),
		VitaminB12:         float64(api.VitaminB12Mcg),
		Calcium:            float64(api.CalciumMg),
		Iron:               float64(api.IronMg),
		Magnesium:          float64(api.MagnesiumMg),
		Phosphorus:         float64(api.PhosphorusMg),
		Potassium:          float64(api.PotassiumMg),
		Zinc:               float64(api.ZincMg),
		Copper:             float64(api.CopperMg),
		Manganese:          float64(api.ManganeseMg),
		Selenium:           float64(api.SeleniumMcg),
		Iodine:             float64(api.IodineMcg),
		Molybdenum:         float64(api.MolybdenumMcg),
		Chromium:           float64(api.ChromiumMcg),
		Fluoride:           float64(api.FluorideMg),
		Chloride:           float64(api.ChlorideMg),
		Biotin:             float64(api.BiotinMcg),
		PantothenicAcid:    float64(api.PantothenicAcidMg),
		Choline:            float64(api.CholineMg),
		MonounsaturatedFat: float64(api.MonounsaturatedFatG),
		PolyunsaturatedFat: float64(api.PolyunsaturatedFatG),
		Omega3Ala:          float64(api.Omega3AlaG),
		Omega3Epa:          float64(api.Omega3EpaG),
		Omega3Dha:          float64(api.Omega3DhaG),
		Omega6:             float64(api.Omega6G),
		Alcohol:            float64(api.AlcoholG),
		Caffeine:           float64(api.CaffeineMg),
		Creatine:           float64(api.CreatineMg),
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

// convertStorageLabelsToAPI converts storage label array to API label array
func convertStorageLabelsToAPI(storageLabels []*storage.Label) *[]api.Label {
	if storageLabels == nil {
		return nil
	}

	apiLabels := make([]api.Label, len(storageLabels))
	for i, label := range storageLabels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}
	return &apiLabels
}

// convertMapFloat64ToFloat32 converts map[string]float64 to map[string]float32
func convertMapFloat64ToFloat32(input map[string]float64) map[string]float32 {
	result := make(map[string]float32)
	for k, v := range input {
		result[k] = float32(v)
	}
	return result
}
