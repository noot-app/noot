// Generated Item struct nutrition fields
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from config/nutrients.yml - see internal/nutrients/generator.go

package storage

// ItemNutrientFields contains all the nutrition fields for the Item struct
// Copy these into the Item struct definition
/*
// Original serving data fields

	OriginalServingGrams *float64 `json:"original_serving_grams,omitempty"` // Serving Size
	OriginalCalories *float64 `json:"original_calories,omitempty"` // Calories
	OriginalProtein *float64 `json:"original_protein_g,omitempty"` // Protein
	OriginalTotalFat *float64 `json:"original_total_fat_g,omitempty"` // Total Fat
	OriginalSaturatedFat *float64 `json:"original_saturated_fat_g,omitempty"` // Saturated Fat
	OriginalTransFat *float64 `json:"original_trans_fat_g,omitempty"` // Trans Fat
	OriginalMonounsaturatedFat *float64 `json:"original_monounsaturated_fat_g,omitempty"` // Monounsaturated Fat
	OriginalPolyunsaturatedFat *float64 `json:"original_polyunsaturated_fat_g,omitempty"` // Polyunsaturated Fat
	OriginalCholesterol *float64 `json:"original_cholesterol_mg,omitempty"` // Cholesterol
	OriginalSodium *float64 `json:"original_sodium_mg,omitempty"` // Sodium
	OriginalTotalCarbs *float64 `json:"original_total_carbs_g,omitempty"` // Total Carbohydrates
	OriginalDietaryFiber *float64 `json:"original_dietary_fiber_g,omitempty"` // Dietary Fiber
	OriginalTotalSugars *float64 `json:"original_total_sugars_g,omitempty"` // Total Sugars
	OriginalAddedSugars *float64 `json:"original_added_sugars_g,omitempty"` // Added Sugars
	OriginalVitaminA *float64 `json:"original_vitamin_a_mcg,omitempty"` // Vitamin A
	OriginalVitaminC *float64 `json:"original_vitamin_c_mg,omitempty"` // Vitamin C
	OriginalVitaminD *float64 `json:"original_vitamin_d_mcg,omitempty"` // Vitamin D
	OriginalVitaminE *float64 `json:"original_vitamin_e_mg,omitempty"` // Vitamin E
	OriginalVitaminK *float64 `json:"original_vitamin_k_mcg,omitempty"` // Vitamin K
	OriginalThiamine *float64 `json:"original_thiamine_mg,omitempty"` // Thiamine (B1)
	OriginalRiboflavin *float64 `json:"original_riboflavin_mg,omitempty"` // Riboflavin (B2)
	OriginalNiacin *float64 `json:"original_niacin_mg,omitempty"` // Niacin (B3)
	OriginalVitaminB6 *float64 `json:"original_vitamin_b6_mg,omitempty"` // Vitamin B6
	OriginalFolate *float64 `json:"original_folate_mcg,omitempty"` // Folate
	OriginalVitaminB12 *float64 `json:"original_vitamin_b12_mcg,omitempty"` // Vitamin B12
	OriginalBiotin *float64 `json:"original_biotin_mcg,omitempty"` // Biotin
	OriginalPantothenicAcid *float64 `json:"original_pantothenic_acid_mg,omitempty"` // Pantothenic Acid
	OriginalCholine *float64 `json:"original_choline_mg,omitempty"` // Choline
	OriginalCalcium *float64 `json:"original_calcium_mg,omitempty"` // Calcium
	OriginalIron *float64 `json:"original_iron_mg,omitempty"` // Iron
	OriginalMagnesium *float64 `json:"original_magnesium_mg,omitempty"` // Magnesium
	OriginalPhosphorus *float64 `json:"original_phosphorus_mg,omitempty"` // Phosphorus
	OriginalPotassium *float64 `json:"original_potassium_mg,omitempty"` // Potassium
	OriginalZinc *float64 `json:"original_zinc_mg,omitempty"` // Zinc
	OriginalCopper *float64 `json:"original_copper_mg,omitempty"` // Copper
	OriginalManganese *float64 `json:"original_manganese_mg,omitempty"` // Manganese
	OriginalSelenium *float64 `json:"original_selenium_mcg,omitempty"` // Selenium
	OriginalIodine *float64 `json:"original_iodine_mcg,omitempty"` // Iodine
	OriginalMolybdenum *float64 `json:"original_molybdenum_mcg,omitempty"` // Molybdenum
	OriginalChromium *float64 `json:"original_chromium_mcg,omitempty"` // Chromium
	OriginalFluoride *float64 `json:"original_fluoride_mg,omitempty"` // Fluoride
	OriginalChloride *float64 `json:"original_chloride_mg,omitempty"` // Chloride
	OriginalOmega3Ala *float64 `json:"original_omega3_ala_g,omitempty"` // Omega-3 ALA
	OriginalOmega3Epa *float64 `json:"original_omega3_epa_g,omitempty"` // Omega-3 EPA
	OriginalOmega3Dha *float64 `json:"original_omega3_dha_g,omitempty"` // Omega-3 DHA
	OriginalOmega6 *float64 `json:"original_omega6_g,omitempty"` // Omega-6
	OriginalAlcohol *float64 `json:"original_alcohol_g,omitempty"` // Alcohol
	OriginalCaffeine *float64 `json:"original_caffeine_mg,omitempty"` // Caffeine
	OriginalCreatine *float64 `json:"original_creatine_mg,omitempty"` // Creatine

// Normalized nutrition data per 100g fields  

	CaloriesPer100g float64 `json:"calories_per_100g"` // Calories per 100g
	ProteinPer100g float64 `json:"protein_g_per_100g"` // Protein per 100g
	TotalFatPer100g float64 `json:"total_fat_g_per_100g"` // Total Fat per 100g
	SaturatedFatPer100g float64 `json:"saturated_fat_g_per_100g"` // Saturated Fat per 100g
	TransFatPer100g float64 `json:"trans_fat_g_per_100g"` // Trans Fat per 100g
	MonounsaturatedFatPer100g float64 `json:"monounsaturated_fat_g_per_100g"` // Monounsaturated Fat per 100g
	PolyunsaturatedFatPer100g float64 `json:"polyunsaturated_fat_g_per_100g"` // Polyunsaturated Fat per 100g
	CholesterolPer100g float64 `json:"cholesterol_mg_per_100g"` // Cholesterol per 100g
	SodiumPer100g float64 `json:"sodium_mg_per_100g"` // Sodium per 100g
	TotalCarbsPer100g float64 `json:"total_carbs_g_per_100g"` // Total Carbohydrates per 100g
	DietaryFiberPer100g float64 `json:"dietary_fiber_g_per_100g"` // Dietary Fiber per 100g
	TotalSugarsPer100g float64 `json:"total_sugars_g_per_100g"` // Total Sugars per 100g
	AddedSugarsPer100g float64 `json:"added_sugars_g_per_100g"` // Added Sugars per 100g
	VitaminAPer100g float64 `json:"vitamin_a_mcg_per_100g"` // Vitamin A per 100g
	VitaminCPer100g float64 `json:"vitamin_c_mg_per_100g"` // Vitamin C per 100g
	VitaminDPer100g float64 `json:"vitamin_d_mcg_per_100g"` // Vitamin D per 100g
	VitaminEPer100g float64 `json:"vitamin_e_mg_per_100g"` // Vitamin E per 100g
	VitaminKPer100g float64 `json:"vitamin_k_mcg_per_100g"` // Vitamin K per 100g
	ThiaminePer100g float64 `json:"thiamine_mg_per_100g"` // Thiamine (B1) per 100g
	RiboflavinPer100g float64 `json:"riboflavin_mg_per_100g"` // Riboflavin (B2) per 100g
	NiacinPer100g float64 `json:"niacin_mg_per_100g"` // Niacin (B3) per 100g
	VitaminB6Per100g float64 `json:"vitamin_b6_mg_per_100g"` // Vitamin B6 per 100g
	FolatePer100g float64 `json:"folate_mcg_per_100g"` // Folate per 100g
	VitaminB12Per100g float64 `json:"vitamin_b12_mcg_per_100g"` // Vitamin B12 per 100g
	BiotinPer100g float64 `json:"biotin_mcg_per_100g"` // Biotin per 100g
	PantothenicAcidPer100g float64 `json:"pantothenic_acid_mg_per_100g"` // Pantothenic Acid per 100g
	CholinePer100g float64 `json:"choline_mg_per_100g"` // Choline per 100g
	CalciumPer100g float64 `json:"calcium_mg_per_100g"` // Calcium per 100g
	IronPer100g float64 `json:"iron_mg_per_100g"` // Iron per 100g
	MagnesiumPer100g float64 `json:"magnesium_mg_per_100g"` // Magnesium per 100g
	PhosphorusPer100g float64 `json:"phosphorus_mg_per_100g"` // Phosphorus per 100g
	PotassiumPer100g float64 `json:"potassium_mg_per_100g"` // Potassium per 100g
	ZincPer100g float64 `json:"zinc_mg_per_100g"` // Zinc per 100g
	CopperPer100g float64 `json:"copper_mg_per_100g"` // Copper per 100g
	ManganesePer100g float64 `json:"manganese_mg_per_100g"` // Manganese per 100g
	SeleniumPer100g float64 `json:"selenium_mcg_per_100g"` // Selenium per 100g
	IodinePer100g float64 `json:"iodine_mcg_per_100g"` // Iodine per 100g
	MolybdenumPer100g float64 `json:"molybdenum_mcg_per_100g"` // Molybdenum per 100g
	ChromiumPer100g float64 `json:"chromium_mcg_per_100g"` // Chromium per 100g
	FluoridePer100g float64 `json:"fluoride_mg_per_100g"` // Fluoride per 100g
	ChloridePer100g float64 `json:"chloride_mg_per_100g"` // Chloride per 100g
	Omega3AlaPer100g float64 `json:"omega3_ala_g_per_100g"` // Omega-3 ALA per 100g
	Omega3EpaPer100g float64 `json:"omega3_epa_g_per_100g"` // Omega-3 EPA per 100g
	Omega3DhaPer100g float64 `json:"omega3_dha_g_per_100g"` // Omega-3 DHA per 100g
	Omega6Per100g float64 `json:"omega6_g_per_100g"` // Omega-6 per 100g
	AlcoholPer100g float64 `json:"alcohol_g_per_100g"` // Alcohol per 100g
	CaffeinePer100g float64 `json:"caffeine_mg_per_100g"` // Caffeine per 100g
	CreatinePer100g float64 `json:"creatine_mg_per_100g"` // Creatine per 100g
*/
