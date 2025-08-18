package server

type ParsedItems struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name      string            `json:"name"`
	Quantity  *float64          `json:"quantity"` // null -> nil
	Unit      *string           `json:"unit"`
	Brand     *string           `json:"brand"`
	Nutrients *CompleteNutrient `json:"nutrients,omitempty"`
}

type CompleteNutrient struct {
	// Basic macronutrients (per serving)
	Calories     float64 `json:"calories"`
	Protein      float64 `json:"protein_g"`
	TotalFat     float64 `json:"total_fat_g"`
	SaturatedFat float64 `json:"saturated_fat_g"`
	TransFat     float64 `json:"trans_fat_g"`
	Cholesterol  float64 `json:"cholesterol_mg"`
	Sodium       float64 `json:"sodium_mg"`
	TotalCarbs   float64 `json:"total_carbs_g"`
	DietaryFiber float64 `json:"dietary_fiber_g"`
	TotalSugars  float64 `json:"total_sugars_g"`
	AddedSugars  float64 `json:"added_sugars_g"`

	// Key vitamins
	VitaminA   float64 `json:"vitamin_a_mcg"`
	VitaminC   float64 `json:"vitamin_c_mg"`
	VitaminD   float64 `json:"vitamin_d_mcg"`
	VitaminE   float64 `json:"vitamin_e_mg"`
	VitaminK   float64 `json:"vitamin_k_mcg"`
	Thiamine   float64 `json:"thiamine_mg"`   // B1
	Riboflavin float64 `json:"riboflavin_mg"` // B2
	Niacin     float64 `json:"niacin_mg"`     // B3
	VitaminB6  float64 `json:"vitamin_b6_mg"`
	Folate     float64 `json:"folate_mcg"`
	VitaminB12 float64 `json:"vitamin_b12_mcg"`

	// Key minerals
	Calcium    float64 `json:"calcium_mg"`
	Iron       float64 `json:"iron_mg"`
	Magnesium  float64 `json:"magnesium_mg"`
	Phosphorus float64 `json:"phosphorus_mg"`
	Potassium  float64 `json:"potassium_mg"`
	Zinc       float64 `json:"zinc_mg"`
	Copper     float64 `json:"copper_mg"`
	Manganese  float64 `json:"manganese_mg"`
	Selenium   float64 `json:"selenium_mcg"`
}

type ItemWithNutrition struct {
	Item Item   `json:"item"`
	Note string `json:"note,omitempty"`
}

type Summary struct {
	Totals          CompleteNutrient   `json:"totals"`
	PercentOfDaily  map[string]int     `json:"percent_of_daily"`
	DailyValuesUsed map[string]float64 `json:"daily_values"`
}
