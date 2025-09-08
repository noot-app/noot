// Generated nutrition service helpers
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from config/nutrients.yml

package server

import (
	"github.com/grantbirki/noot/internal/storage"
)

// NutrientDescriptor contains metadata about a nutrient field
type NutrientDescriptor struct {
	Key         string
	GoField     string
	Precision   int
	Category    string
	Unit        string
	DisplayName string
}

// NutrientMeta contains all nutrient field metadata
var NutrientMeta = []NutrientDescriptor{
	{
		Key:         "calories",
		GoField:     "Calories",
		Precision:   3,
		Category:    "macro",
		Unit:        "kcal",
		DisplayName: "Calories",
	},
	{
		Key:         "protein_g",
		GoField:     "Protein",
		Precision:   2,
		Category:    "macro",
		Unit:        "g",
		DisplayName: "Protein",
	},
	{
		Key:         "total_fat_g",
		GoField:     "TotalFat",
		Precision:   2,
		Category:    "macro",
		Unit:        "g",
		DisplayName: "Total Fat",
	},
	{
		Key:         "saturated_fat_g",
		GoField:     "SaturatedFat",
		Precision:   2,
		Category:    "fat",
		Unit:        "g",
		DisplayName: "Saturated Fat",
	},
	{
		Key:         "trans_fat_g",
		GoField:     "TransFat",
		Precision:   1,
		Category:    "fat",
		Unit:        "g",
		DisplayName: "Trans Fat",
	},
	{
		Key:         "monounsaturated_fat_g",
		GoField:     "MonounsaturatedFat",
		Precision:   2,
		Category:    "fat",
		Unit:        "g",
		DisplayName: "Monounsaturated Fat",
	},
	{
		Key:         "polyunsaturated_fat_g",
		GoField:     "PolyunsaturatedFat",
		Precision:   2,
		Category:    "fat",
		Unit:        "g",
		DisplayName: "Polyunsaturated Fat",
	},
	{
		Key:         "cholesterol_mg",
		GoField:     "Cholesterol",
		Precision:   1,
		Category:    "fat",
		Unit:        "mg",
		DisplayName: "Cholesterol",
	},
	{
		Key:         "sodium_mg",
		GoField:     "Sodium",
		Precision:   1,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Sodium",
	},
	{
		Key:         "total_carbs_g",
		GoField:     "TotalCarbs",
		Precision:   1,
		Category:    "carb",
		Unit:        "g",
		DisplayName: "Total Carbohydrates",
	},
	{
		Key:         "dietary_fiber_g",
		GoField:     "DietaryFiber",
		Precision:   1,
		Category:    "carb",
		Unit:        "g",
		DisplayName: "Dietary Fiber",
	},
	{
		Key:         "total_sugars_g",
		GoField:     "TotalSugars",
		Precision:   1,
		Category:    "carb",
		Unit:        "g",
		DisplayName: "Total Sugars",
	},
	{
		Key:         "added_sugars_g",
		GoField:     "AddedSugars",
		Precision:   1,
		Category:    "carb",
		Unit:        "g",
		DisplayName: "Added Sugars",
	},
	{
		Key:         "vitamin_a_mcg",
		GoField:     "VitaminA",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mcg",
		DisplayName: "Vitamin A",
	},
	{
		Key:         "vitamin_c_mg",
		GoField:     "VitaminC",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Vitamin C",
	},
	{
		Key:         "vitamin_d_mcg",
		GoField:     "VitaminD",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mcg",
		DisplayName: "Vitamin D",
	},
	{
		Key:         "vitamin_e_mg",
		GoField:     "VitaminE",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Vitamin E",
	},
	{
		Key:         "vitamin_k_mcg",
		GoField:     "VitaminK",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mcg",
		DisplayName: "Vitamin K",
	},
	{
		Key:         "thiamine_mg",
		GoField:     "Thiamine",
		Precision:   3,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Thiamine (B1)",
	},
	{
		Key:         "riboflavin_mg",
		GoField:     "Riboflavin",
		Precision:   3,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Riboflavin (B2)",
	},
	{
		Key:         "niacin_mg",
		GoField:     "Niacin",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Niacin (B3)",
	},
	{
		Key:         "vitamin_b6_mg",
		GoField:     "VitaminB6",
		Precision:   3,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Vitamin B6",
	},
	{
		Key:         "folate_mcg",
		GoField:     "Folate",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mcg",
		DisplayName: "Folate",
	},
	{
		Key:         "vitamin_b12_mcg",
		GoField:     "VitaminB12",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mcg",
		DisplayName: "Vitamin B12",
	},
	{
		Key:         "biotin_mcg",
		GoField:     "Biotin",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mcg",
		DisplayName: "Biotin",
	},
	{
		Key:         "pantothenic_acid_mg",
		GoField:     "PantothenicAcid",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Pantothenic Acid",
	},
	{
		Key:         "choline_mg",
		GoField:     "Choline",
		Precision:   2,
		Category:    "vitamin",
		Unit:        "mg",
		DisplayName: "Choline",
	},
	{
		Key:         "calcium_mg",
		GoField:     "Calcium",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Calcium",
	},
	{
		Key:         "iron_mg",
		GoField:     "Iron",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Iron",
	},
	{
		Key:         "magnesium_mg",
		GoField:     "Magnesium",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Magnesium",
	},
	{
		Key:         "phosphorus_mg",
		GoField:     "Phosphorus",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Phosphorus",
	},
	{
		Key:         "potassium_mg",
		GoField:     "Potassium",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Potassium",
	},
	{
		Key:         "zinc_mg",
		GoField:     "Zinc",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Zinc",
	},
	{
		Key:         "copper_mg",
		GoField:     "Copper",
		Precision:   3,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Copper",
	},
	{
		Key:         "manganese_mg",
		GoField:     "Manganese",
		Precision:   3,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Manganese",
	},
	{
		Key:         "selenium_mcg",
		GoField:     "Selenium",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mcg",
		DisplayName: "Selenium",
	},
	{
		Key:         "iodine_mcg",
		GoField:     "Iodine",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mcg",
		DisplayName: "Iodine",
	},
	{
		Key:         "molybdenum_mcg",
		GoField:     "Molybdenum",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mcg",
		DisplayName: "Molybdenum",
	},
	{
		Key:         "chromium_mcg",
		GoField:     "Chromium",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mcg",
		DisplayName: "Chromium",
	},
	{
		Key:         "fluoride_mg",
		GoField:     "Fluoride",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Fluoride",
	},
	{
		Key:         "chloride_mg",
		GoField:     "Chloride",
		Precision:   2,
		Category:    "mineral",
		Unit:        "mg",
		DisplayName: "Chloride",
	},
	{
		Key:         "omega3_ala_g",
		GoField:     "Omega3Ala",
		Precision:   3,
		Category:    "fatty_acid",
		Unit:        "g",
		DisplayName: "Omega-3 ALA",
	},
	{
		Key:         "omega3_epa_g",
		GoField:     "Omega3Epa",
		Precision:   3,
		Category:    "fatty_acid",
		Unit:        "g",
		DisplayName: "Omega-3 EPA",
	},
	{
		Key:         "omega3_dha_g",
		GoField:     "Omega3Dha",
		Precision:   3,
		Category:    "fatty_acid",
		Unit:        "g",
		DisplayName: "Omega-3 DHA",
	},
	{
		Key:         "omega6_g",
		GoField:     "Omega6",
		Precision:   2,
		Category:    "fatty_acid",
		Unit:        "g",
		DisplayName: "Omega-6",
	},
	{
		Key:         "alcohol_g",
		GoField:     "Alcohol",
		Precision:   2,
		Category:    "compound",
		Unit:        "g",
		DisplayName: "Alcohol",
	},
	{
		Key:         "caffeine_mg",
		GoField:     "Caffeine",
		Precision:   2,
		Category:    "compound",
		Unit:        "mg",
		DisplayName: "Caffeine",
	},
	{
		Key:         "creatine_mg",
		GoField:     "Creatine",
		Precision:   2,
		Category:    "compound",
		Unit:        "mg",
		DisplayName: "Creatine",
	},
}

// NutrientPrecisionMap maps nutrient keys to their precision values
var NutrientPrecisionMap = map[string]int{
	"calories":              3,
	"protein_g":             2,
	"total_fat_g":           2,
	"saturated_fat_g":       2,
	"trans_fat_g":           1,
	"monounsaturated_fat_g": 2,
	"polyunsaturated_fat_g": 2,
	"cholesterol_mg":        1,
	"sodium_mg":             1,
	"total_carbs_g":         1,
	"dietary_fiber_g":       1,
	"total_sugars_g":        1,
	"added_sugars_g":        1,
	"vitamin_a_mcg":         2,
	"vitamin_c_mg":          2,
	"vitamin_d_mcg":         2,
	"vitamin_e_mg":          2,
	"vitamin_k_mcg":         2,
	"thiamine_mg":           3,
	"riboflavin_mg":         3,
	"niacin_mg":             2,
	"vitamin_b6_mg":         3,
	"folate_mcg":            2,
	"vitamin_b12_mcg":       2,
	"biotin_mcg":            2,
	"pantothenic_acid_mg":   2,
	"choline_mg":            2,
	"calcium_mg":            2,
	"iron_mg":               2,
	"magnesium_mg":          2,
	"phosphorus_mg":         2,
	"potassium_mg":          2,
	"zinc_mg":               2,
	"copper_mg":             3,
	"manganese_mg":          3,
	"selenium_mcg":          2,
	"iodine_mcg":            2,
	"molybdenum_mcg":        2,
	"chromium_mcg":          2,
	"fluoride_mg":           2,
	"chloride_mg":           2,
	"omega3_ala_g":          3,
	"omega3_epa_g":          3,
	"omega3_dha_g":          3,
	"omega6_g":              2,
	"alcohol_g":             2,
	"caffeine_mg":           2,
	"creatine_mg":           2,
}

// Scale multiplies all nutrient values by the given factor
func (c *CompleteNutrient) Scale(factor float64) {
	c.Calories *= factor
	c.Protein *= factor
	c.TotalFat *= factor
	c.SaturatedFat *= factor
	c.TransFat *= factor
	c.MonounsaturatedFat *= factor
	c.PolyunsaturatedFat *= factor
	c.Cholesterol *= factor
	c.Sodium *= factor
	c.TotalCarbs *= factor
	c.DietaryFiber *= factor
	c.TotalSugars *= factor
	c.AddedSugars *= factor
	c.VitaminA *= factor
	c.VitaminC *= factor
	c.VitaminD *= factor
	c.VitaminE *= factor
	c.VitaminK *= factor
	c.Thiamine *= factor
	c.Riboflavin *= factor
	c.Niacin *= factor
	c.VitaminB6 *= factor
	c.Folate *= factor
	c.VitaminB12 *= factor
	c.Biotin *= factor
	c.PantothenicAcid *= factor
	c.Choline *= factor
	c.Calcium *= factor
	c.Iron *= factor
	c.Magnesium *= factor
	c.Phosphorus *= factor
	c.Potassium *= factor
	c.Zinc *= factor
	c.Copper *= factor
	c.Manganese *= factor
	c.Selenium *= factor
	c.Iodine *= factor
	c.Molybdenum *= factor
	c.Chromium *= factor
	c.Fluoride *= factor
	c.Chloride *= factor
	c.Omega3Ala *= factor
	c.Omega3Epa *= factor
	c.Omega3Dha *= factor
	c.Omega6 *= factor
	c.Alcohol *= factor
	c.Caffeine *= factor
	c.Creatine *= factor
}

// CopyFrom copies all nutrient values from source to destination
func (dst *CompleteNutrient) CopyFrom(src *CompleteNutrient) {
	dst.Calories = src.Calories
	dst.Protein = src.Protein
	dst.TotalFat = src.TotalFat
	dst.SaturatedFat = src.SaturatedFat
	dst.TransFat = src.TransFat
	dst.MonounsaturatedFat = src.MonounsaturatedFat
	dst.PolyunsaturatedFat = src.PolyunsaturatedFat
	dst.Cholesterol = src.Cholesterol
	dst.Sodium = src.Sodium
	dst.TotalCarbs = src.TotalCarbs
	dst.DietaryFiber = src.DietaryFiber
	dst.TotalSugars = src.TotalSugars
	dst.AddedSugars = src.AddedSugars
	dst.VitaminA = src.VitaminA
	dst.VitaminC = src.VitaminC
	dst.VitaminD = src.VitaminD
	dst.VitaminE = src.VitaminE
	dst.VitaminK = src.VitaminK
	dst.Thiamine = src.Thiamine
	dst.Riboflavin = src.Riboflavin
	dst.Niacin = src.Niacin
	dst.VitaminB6 = src.VitaminB6
	dst.Folate = src.Folate
	dst.VitaminB12 = src.VitaminB12
	dst.Biotin = src.Biotin
	dst.PantothenicAcid = src.PantothenicAcid
	dst.Choline = src.Choline
	dst.Calcium = src.Calcium
	dst.Iron = src.Iron
	dst.Magnesium = src.Magnesium
	dst.Phosphorus = src.Phosphorus
	dst.Potassium = src.Potassium
	dst.Zinc = src.Zinc
	dst.Copper = src.Copper
	dst.Manganese = src.Manganese
	dst.Selenium = src.Selenium
	dst.Iodine = src.Iodine
	dst.Molybdenum = src.Molybdenum
	dst.Chromium = src.Chromium
	dst.Fluoride = src.Fluoride
	dst.Chloride = src.Chloride
	dst.Omega3Ala = src.Omega3Ala
	dst.Omega3Epa = src.Omega3Epa
	dst.Omega3Dha = src.Omega3Dha
	dst.Omega6 = src.Omega6
	dst.Alcohol = src.Alcohol
	dst.Caffeine = src.Caffeine
	dst.Creatine = src.Creatine
}

// ConvertPer100gToServing converts per-100g nutrition data to actual serving size
func ConvertPer100gToServing(cached *storage.Item, grams float64, converter *UnitConverter) CompleteNutrient {
	// Helper function to convert and round in one step
	convertAndRound := func(per100gValue float64, precision int) float64 {
		return RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(per100gValue, grams), precision)
	}

	return CompleteNutrient{
		Calories:           convertAndRound(cached.CaloriesPer100g, 3),
		Protein:            convertAndRound(cached.ProteinGPer100g, 2),
		TotalFat:           convertAndRound(cached.TotalFatGPer100g, 2),
		SaturatedFat:       convertAndRound(cached.SaturatedFatGPer100g, 2),
		TransFat:           convertAndRound(cached.TransFatGPer100g, 1),
		MonounsaturatedFat: convertAndRound(cached.MonounsaturatedFatGPer100g, 2),
		PolyunsaturatedFat: convertAndRound(cached.PolyunsaturatedFatGPer100g, 2),
		Cholesterol:        convertAndRound(cached.CholesterolMgPer100g, 1),
		Sodium:             convertAndRound(cached.SodiumMgPer100g, 1),
		TotalCarbs:         convertAndRound(cached.TotalCarbsGPer100g, 1),
		DietaryFiber:       convertAndRound(cached.DietaryFiberGPer100g, 1),
		TotalSugars:        convertAndRound(cached.TotalSugarsGPer100g, 1),
		AddedSugars:        convertAndRound(cached.AddedSugarsGPer100g, 1),
		VitaminA:           convertAndRound(cached.VitaminAMcgPer100g, 2),
		VitaminC:           convertAndRound(cached.VitaminCMgPer100g, 2),
		VitaminD:           convertAndRound(cached.VitaminDMcgPer100g, 2),
		VitaminE:           convertAndRound(cached.VitaminEMgPer100g, 2),
		VitaminK:           convertAndRound(cached.VitaminKMcgPer100g, 2),
		Thiamine:           convertAndRound(cached.ThiamineMgPer100g, 3),
		Riboflavin:         convertAndRound(cached.RiboflavinMgPer100g, 3),
		Niacin:             convertAndRound(cached.NiacinMgPer100g, 2),
		VitaminB6:          convertAndRound(cached.VitaminB6MgPer100g, 3),
		Folate:             convertAndRound(cached.FolateMcgPer100g, 2),
		VitaminB12:         convertAndRound(cached.VitaminB12McgPer100g, 2),
		Biotin:             convertAndRound(cached.BiotinMcgPer100g, 2),
		PantothenicAcid:    convertAndRound(cached.PantothenicAcidMgPer100g, 2),
		Choline:            convertAndRound(cached.CholineMgPer100g, 2),
		Calcium:            convertAndRound(cached.CalciumMgPer100g, 2),
		Iron:               convertAndRound(cached.IronMgPer100g, 2),
		Magnesium:          convertAndRound(cached.MagnesiumMgPer100g, 2),
		Phosphorus:         convertAndRound(cached.PhosphorusMgPer100g, 2),
		Potassium:          convertAndRound(cached.PotassiumMgPer100g, 2),
		Zinc:               convertAndRound(cached.ZincMgPer100g, 2),
		Copper:             convertAndRound(cached.CopperMgPer100g, 3),
		Manganese:          convertAndRound(cached.ManganeseMgPer100g, 3),
		Selenium:           convertAndRound(cached.SeleniumMcgPer100g, 2),
		Iodine:             convertAndRound(cached.IodineMcgPer100g, 2),
		Molybdenum:         convertAndRound(cached.MolybdenumMcgPer100g, 2),
		Chromium:           convertAndRound(cached.ChromiumMcgPer100g, 2),
		Fluoride:           convertAndRound(cached.FluorideMgPer100g, 2),
		Chloride:           convertAndRound(cached.ChlorideMgPer100g, 2),
		Omega3Ala:          convertAndRound(cached.Omega3AlaGPer100g, 3),
		Omega3Epa:          convertAndRound(cached.Omega3EpaGPer100g, 3),
		Omega3Dha:          convertAndRound(cached.Omega3DhaGPer100g, 3),
		Omega6:             convertAndRound(cached.Omega6GPer100g, 2),
		Alcohol:            convertAndRound(cached.AlcoholGPer100g, 2),
		Caffeine:           convertAndRound(cached.CaffeineMgPer100g, 2),
		Creatine:           convertAndRound(cached.CreatineMgPer100g, 2),
	}
}

// ConvertServingToPer100g converts serving nutrition data to per-100g format
func ConvertServingToPer100g(nutrients CompleteNutrient, grams float64, converter *UnitConverter) Per100gSnapshot {
	// Helper function to convert and round in one step
	convertAndRound := func(servingValue float64, precision int) float64 {
		return RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(servingValue, grams), precision)
	}

	return Per100gSnapshot{
		CaloriesPer100g:            convertAndRound(nutrients.Calories, 3),
		ProteinGPer100g:            convertAndRound(nutrients.Protein, 2),
		TotalFatGPer100g:           convertAndRound(nutrients.TotalFat, 2),
		SaturatedFatGPer100g:       convertAndRound(nutrients.SaturatedFat, 2),
		TransFatGPer100g:           convertAndRound(nutrients.TransFat, 1),
		MonounsaturatedFatGPer100g: convertAndRound(nutrients.MonounsaturatedFat, 2),
		PolyunsaturatedFatGPer100g: convertAndRound(nutrients.PolyunsaturatedFat, 2),
		CholesterolMgPer100g:       convertAndRound(nutrients.Cholesterol, 1),
		SodiumMgPer100g:            convertAndRound(nutrients.Sodium, 1),
		TotalCarbsGPer100g:         convertAndRound(nutrients.TotalCarbs, 1),
		DietaryFiberGPer100g:       convertAndRound(nutrients.DietaryFiber, 1),
		TotalSugarsGPer100g:        convertAndRound(nutrients.TotalSugars, 1),
		AddedSugarsGPer100g:        convertAndRound(nutrients.AddedSugars, 1),
		VitaminAMcgPer100g:         convertAndRound(nutrients.VitaminA, 2),
		VitaminCMgPer100g:          convertAndRound(nutrients.VitaminC, 2),
		VitaminDMcgPer100g:         convertAndRound(nutrients.VitaminD, 2),
		VitaminEMgPer100g:          convertAndRound(nutrients.VitaminE, 2),
		VitaminKMcgPer100g:         convertAndRound(nutrients.VitaminK, 2),
		ThiamineMgPer100g:          convertAndRound(nutrients.Thiamine, 3),
		RiboflavinMgPer100g:        convertAndRound(nutrients.Riboflavin, 3),
		NiacinMgPer100g:            convertAndRound(nutrients.Niacin, 2),
		VitaminB6MgPer100g:         convertAndRound(nutrients.VitaminB6, 3),
		FolateMcgPer100g:           convertAndRound(nutrients.Folate, 2),
		VitaminB12McgPer100g:       convertAndRound(nutrients.VitaminB12, 2),
		BiotinMcgPer100g:           convertAndRound(nutrients.Biotin, 2),
		PantothenicAcidMgPer100g:   convertAndRound(nutrients.PantothenicAcid, 2),
		CholineMgPer100g:           convertAndRound(nutrients.Choline, 2),
		CalciumMgPer100g:           convertAndRound(nutrients.Calcium, 2),
		IronMgPer100g:              convertAndRound(nutrients.Iron, 2),
		MagnesiumMgPer100g:         convertAndRound(nutrients.Magnesium, 2),
		PhosphorusMgPer100g:        convertAndRound(nutrients.Phosphorus, 2),
		PotassiumMgPer100g:         convertAndRound(nutrients.Potassium, 2),
		ZincMgPer100g:              convertAndRound(nutrients.Zinc, 2),
		CopperMgPer100g:            convertAndRound(nutrients.Copper, 3),
		ManganeseMgPer100g:         convertAndRound(nutrients.Manganese, 3),
		SeleniumMcgPer100g:         convertAndRound(nutrients.Selenium, 2),
		IodineMcgPer100g:           convertAndRound(nutrients.Iodine, 2),
		MolybdenumMcgPer100g:       convertAndRound(nutrients.Molybdenum, 2),
		ChromiumMcgPer100g:         convertAndRound(nutrients.Chromium, 2),
		FluorideMgPer100g:          convertAndRound(nutrients.Fluoride, 2),
		ChlorideMgPer100g:          convertAndRound(nutrients.Chloride, 2),
		Omega3AlaGPer100g:          convertAndRound(nutrients.Omega3Ala, 3),
		Omega3EpaGPer100g:          convertAndRound(nutrients.Omega3Epa, 3),
		Omega3DhaGPer100g:          convertAndRound(nutrients.Omega3Dha, 3),
		Omega6GPer100g:             convertAndRound(nutrients.Omega6, 2),
		AlcoholGPer100g:            convertAndRound(nutrients.Alcohol, 2),
		CaffeineMgPer100g:          convertAndRound(nutrients.Caffeine, 2),
		CreatineMgPer100g:          convertAndRound(nutrients.Creatine, 2),
	}
}

// ConvertExactCachedToNutrients converts cached exact serving data to nutrients
func ConvertExactCachedToNutrients(cached *storage.Item) CompleteNutrient {
	// Helper function to safely dereference pointers
	floatValue := func(f *float64) float64 {
		if f == nil {
			return 0
		}
		return *f
	}

	// If we have original serving data, use it directly
	if cached.OriginalServingGrams != nil {
		return CompleteNutrient{
			Calories:           floatValue(cached.OriginalCalories),
			Protein:            floatValue(cached.OriginalProteinG),
			TotalFat:           floatValue(cached.OriginalTotalFatG),
			SaturatedFat:       floatValue(cached.OriginalSaturatedFatG),
			TransFat:           floatValue(cached.OriginalTransFatG),
			MonounsaturatedFat: floatValue(cached.OriginalMonounsaturatedFatG),
			PolyunsaturatedFat: floatValue(cached.OriginalPolyunsaturatedFatG),
			Cholesterol:        floatValue(cached.OriginalCholesterolMg),
			Sodium:             floatValue(cached.OriginalSodiumMg),
			TotalCarbs:         floatValue(cached.OriginalTotalCarbsG),
			DietaryFiber:       floatValue(cached.OriginalDietaryFiberG),
			TotalSugars:        floatValue(cached.OriginalTotalSugarsG),
			AddedSugars:        floatValue(cached.OriginalAddedSugarsG),
			VitaminA:           floatValue(cached.OriginalVitaminAMcg),
			VitaminC:           floatValue(cached.OriginalVitaminCMg),
			VitaminD:           floatValue(cached.OriginalVitaminDMcg),
			VitaminE:           floatValue(cached.OriginalVitaminEMg),
			VitaminK:           floatValue(cached.OriginalVitaminKMcg),
			Thiamine:           floatValue(cached.OriginalThiamineMg),
			Riboflavin:         floatValue(cached.OriginalRiboflavinMg),
			Niacin:             floatValue(cached.OriginalNiacinMg),
			VitaminB6:          floatValue(cached.OriginalVitaminB6Mg),
			Folate:             floatValue(cached.OriginalFolateMcg),
			VitaminB12:         floatValue(cached.OriginalVitaminB12Mcg),
			Biotin:             floatValue(cached.OriginalBiotinMcg),
			PantothenicAcid:    floatValue(cached.OriginalPantothenicAcidMg),
			Choline:            floatValue(cached.OriginalCholineMg),
			Calcium:            floatValue(cached.OriginalCalciumMg),
			Iron:               floatValue(cached.OriginalIronMg),
			Magnesium:          floatValue(cached.OriginalMagnesiumMg),
			Phosphorus:         floatValue(cached.OriginalPhosphorusMg),
			Potassium:          floatValue(cached.OriginalPotassiumMg),
			Zinc:               floatValue(cached.OriginalZincMg),
			Copper:             floatValue(cached.OriginalCopperMg),
			Manganese:          floatValue(cached.OriginalManganeseMg),
			Selenium:           floatValue(cached.OriginalSeleniumMcg),
			Iodine:             floatValue(cached.OriginalIodineMcg),
			Molybdenum:         floatValue(cached.OriginalMolybdenumMcg),
			Chromium:           floatValue(cached.OriginalChromiumMcg),
			Fluoride:           floatValue(cached.OriginalFluorideMg),
			Chloride:           floatValue(cached.OriginalChlorideMg),
			Omega3Ala:          floatValue(cached.OriginalOmega3AlaG),
			Omega3Epa:          floatValue(cached.OriginalOmega3EpaG),
			Omega3Dha:          floatValue(cached.OriginalOmega3DhaG),
			Omega6:             floatValue(cached.OriginalOmega6G),
			Alcohol:            floatValue(cached.OriginalAlcoholG),
			Caffeine:           floatValue(cached.OriginalCaffeineMg),
			Creatine:           floatValue(cached.OriginalCreatineMg),
		}
	}

	// Fallback to per-100g data (shouldn't happen with new schema)
	return CompleteNutrient{
		Calories:           cached.CaloriesPer100g,
		Protein:            cached.ProteinGPer100g,
		TotalFat:           cached.TotalFatGPer100g,
		SaturatedFat:       cached.SaturatedFatGPer100g,
		TransFat:           cached.TransFatGPer100g,
		MonounsaturatedFat: cached.MonounsaturatedFatGPer100g,
		PolyunsaturatedFat: cached.PolyunsaturatedFatGPer100g,
		Cholesterol:        cached.CholesterolMgPer100g,
		Sodium:             cached.SodiumMgPer100g,
		TotalCarbs:         cached.TotalCarbsGPer100g,
		DietaryFiber:       cached.DietaryFiberGPer100g,
		TotalSugars:        cached.TotalSugarsGPer100g,
		AddedSugars:        cached.AddedSugarsGPer100g,
		VitaminA:           cached.VitaminAMcgPer100g,
		VitaminC:           cached.VitaminCMgPer100g,
		VitaminD:           cached.VitaminDMcgPer100g,
		VitaminE:           cached.VitaminEMgPer100g,
		VitaminK:           cached.VitaminKMcgPer100g,
		Thiamine:           cached.ThiamineMgPer100g,
		Riboflavin:         cached.RiboflavinMgPer100g,
		Niacin:             cached.NiacinMgPer100g,
		VitaminB6:          cached.VitaminB6MgPer100g,
		Folate:             cached.FolateMcgPer100g,
		VitaminB12:         cached.VitaminB12McgPer100g,
		Biotin:             cached.BiotinMcgPer100g,
		PantothenicAcid:    cached.PantothenicAcidMgPer100g,
		Choline:            cached.CholineMgPer100g,
		Calcium:            cached.CalciumMgPer100g,
		Iron:               cached.IronMgPer100g,
		Magnesium:          cached.MagnesiumMgPer100g,
		Phosphorus:         cached.PhosphorusMgPer100g,
		Potassium:          cached.PotassiumMgPer100g,
		Zinc:               cached.ZincMgPer100g,
		Copper:             cached.CopperMgPer100g,
		Manganese:          cached.ManganeseMgPer100g,
		Selenium:           cached.SeleniumMcgPer100g,
		Iodine:             cached.IodineMcgPer100g,
		Molybdenum:         cached.MolybdenumMcgPer100g,
		Chromium:           cached.ChromiumMcgPer100g,
		Fluoride:           cached.FluorideMgPer100g,
		Chloride:           cached.ChlorideMgPer100g,
		Omega3Ala:          cached.Omega3AlaGPer100g,
		Omega3Epa:          cached.Omega3EpaGPer100g,
		Omega3Dha:          cached.Omega3DhaGPer100g,
		Omega6:             cached.Omega6GPer100g,
		Alcohol:            cached.AlcoholGPer100g,
		Caffeine:           cached.CaffeineMgPer100g,
		Creatine:           cached.CreatineMgPer100g,
	}
}

// ScaleNutritionFromCachedServing scales nutrition data from one serving to another
func ScaleNutritionFromCachedServing(cachedItem *storage.Item, fromGrams, toGrams float64) CompleteNutrient {
	scalingFactor := toGrams / fromGrams

	// Helper function to safely dereference pointers and scale
	scaleValue := func(ptr *float64) float64 {
		if ptr == nil {
			return 0.0
		}
		return *ptr * scalingFactor
	}

	return CompleteNutrient{
		Calories:           float64(RoundCaloriesUp(scaleValue(cachedItem.OriginalCalories))),
		Protein:            scaleValue(cachedItem.OriginalProteinG),
		TotalFat:           scaleValue(cachedItem.OriginalTotalFatG),
		SaturatedFat:       scaleValue(cachedItem.OriginalSaturatedFatG),
		TransFat:           scaleValue(cachedItem.OriginalTransFatG),
		MonounsaturatedFat: scaleValue(cachedItem.OriginalMonounsaturatedFatG),
		PolyunsaturatedFat: scaleValue(cachedItem.OriginalPolyunsaturatedFatG),
		Cholesterol:        scaleValue(cachedItem.OriginalCholesterolMg),
		Sodium:             scaleValue(cachedItem.OriginalSodiumMg),
		TotalCarbs:         scaleValue(cachedItem.OriginalTotalCarbsG),
		DietaryFiber:       scaleValue(cachedItem.OriginalDietaryFiberG),
		TotalSugars:        scaleValue(cachedItem.OriginalTotalSugarsG),
		AddedSugars:        scaleValue(cachedItem.OriginalAddedSugarsG),
		VitaminA:           scaleValue(cachedItem.OriginalVitaminAMcg),
		VitaminC:           scaleValue(cachedItem.OriginalVitaminCMg),
		VitaminD:           scaleValue(cachedItem.OriginalVitaminDMcg),
		VitaminE:           scaleValue(cachedItem.OriginalVitaminEMg),
		VitaminK:           scaleValue(cachedItem.OriginalVitaminKMcg),
		Thiamine:           scaleValue(cachedItem.OriginalThiamineMg),
		Riboflavin:         scaleValue(cachedItem.OriginalRiboflavinMg),
		Niacin:             scaleValue(cachedItem.OriginalNiacinMg),
		VitaminB6:          scaleValue(cachedItem.OriginalVitaminB6Mg),
		Folate:             scaleValue(cachedItem.OriginalFolateMcg),
		VitaminB12:         scaleValue(cachedItem.OriginalVitaminB12Mcg),
		Biotin:             scaleValue(cachedItem.OriginalBiotinMcg),
		PantothenicAcid:    scaleValue(cachedItem.OriginalPantothenicAcidMg),
		Choline:            scaleValue(cachedItem.OriginalCholineMg),
		Calcium:            scaleValue(cachedItem.OriginalCalciumMg),
		Iron:               scaleValue(cachedItem.OriginalIronMg),
		Magnesium:          scaleValue(cachedItem.OriginalMagnesiumMg),
		Phosphorus:         scaleValue(cachedItem.OriginalPhosphorusMg),
		Potassium:          scaleValue(cachedItem.OriginalPotassiumMg),
		Zinc:               scaleValue(cachedItem.OriginalZincMg),
		Copper:             scaleValue(cachedItem.OriginalCopperMg),
		Manganese:          scaleValue(cachedItem.OriginalManganeseMg),
		Selenium:           scaleValue(cachedItem.OriginalSeleniumMcg),
		Iodine:             scaleValue(cachedItem.OriginalIodineMcg),
		Molybdenum:         scaleValue(cachedItem.OriginalMolybdenumMcg),
		Chromium:           scaleValue(cachedItem.OriginalChromiumMcg),
		Fluoride:           scaleValue(cachedItem.OriginalFluorideMg),
		Chloride:           scaleValue(cachedItem.OriginalChlorideMg),
		Omega3Ala:          scaleValue(cachedItem.OriginalOmega3AlaG),
		Omega3Epa:          scaleValue(cachedItem.OriginalOmega3EpaG),
		Omega3Dha:          scaleValue(cachedItem.OriginalOmega3DhaG),
		Omega6:             scaleValue(cachedItem.OriginalOmega6G),
		Alcohol:            scaleValue(cachedItem.OriginalAlcoholG),
		Caffeine:           scaleValue(cachedItem.OriginalCaffeineMg),
		Creatine:           scaleValue(cachedItem.OriginalCreatineMg),
	}
}

// ConvertNutrientsToExactCacheFields fills exact cache fields from nutrition data
func ConvertNutrientsToExactCacheFields(item *storage.Item, nutrients CompleteNutrient, grams float64) {
	// Helper to create float64 pointer
	ptr := func(val float64) *float64 { return &val }

	// Set original serving data
	item.OriginalServingGrams = ptr(grams)
	item.OriginalCalories = ptr(nutrients.Calories)
	item.OriginalProteinG = ptr(nutrients.Protein)
	item.OriginalTotalFatG = ptr(nutrients.TotalFat)
	item.OriginalSaturatedFatG = ptr(nutrients.SaturatedFat)
	item.OriginalTransFatG = ptr(nutrients.TransFat)
	item.OriginalMonounsaturatedFatG = ptr(nutrients.MonounsaturatedFat)
	item.OriginalPolyunsaturatedFatG = ptr(nutrients.PolyunsaturatedFat)
	item.OriginalCholesterolMg = ptr(nutrients.Cholesterol)
	item.OriginalSodiumMg = ptr(nutrients.Sodium)
	item.OriginalTotalCarbsG = ptr(nutrients.TotalCarbs)
	item.OriginalDietaryFiberG = ptr(nutrients.DietaryFiber)
	item.OriginalTotalSugarsG = ptr(nutrients.TotalSugars)
	item.OriginalAddedSugarsG = ptr(nutrients.AddedSugars)
	item.OriginalVitaminAMcg = ptr(nutrients.VitaminA)
	item.OriginalVitaminCMg = ptr(nutrients.VitaminC)
	item.OriginalVitaminDMcg = ptr(nutrients.VitaminD)
	item.OriginalVitaminEMg = ptr(nutrients.VitaminE)
	item.OriginalVitaminKMcg = ptr(nutrients.VitaminK)
	item.OriginalThiamineMg = ptr(nutrients.Thiamine)
	item.OriginalRiboflavinMg = ptr(nutrients.Riboflavin)
	item.OriginalNiacinMg = ptr(nutrients.Niacin)
	item.OriginalVitaminB6Mg = ptr(nutrients.VitaminB6)
	item.OriginalFolateMcg = ptr(nutrients.Folate)
	item.OriginalVitaminB12Mcg = ptr(nutrients.VitaminB12)
	item.OriginalBiotinMcg = ptr(nutrients.Biotin)
	item.OriginalPantothenicAcidMg = ptr(nutrients.PantothenicAcid)
	item.OriginalCholineMg = ptr(nutrients.Choline)
	item.OriginalCalciumMg = ptr(nutrients.Calcium)
	item.OriginalIronMg = ptr(nutrients.Iron)
	item.OriginalMagnesiumMg = ptr(nutrients.Magnesium)
	item.OriginalPhosphorusMg = ptr(nutrients.Phosphorus)
	item.OriginalPotassiumMg = ptr(nutrients.Potassium)
	item.OriginalZincMg = ptr(nutrients.Zinc)
	item.OriginalCopperMg = ptr(nutrients.Copper)
	item.OriginalManganeseMg = ptr(nutrients.Manganese)
	item.OriginalSeleniumMcg = ptr(nutrients.Selenium)
	item.OriginalIodineMcg = ptr(nutrients.Iodine)
	item.OriginalMolybdenumMcg = ptr(nutrients.Molybdenum)
	item.OriginalChromiumMcg = ptr(nutrients.Chromium)
	item.OriginalFluorideMg = ptr(nutrients.Fluoride)
	item.OriginalChlorideMg = ptr(nutrients.Chloride)
	item.OriginalOmega3AlaG = ptr(nutrients.Omega3Ala)
	item.OriginalOmega3EpaG = ptr(nutrients.Omega3Epa)
	item.OriginalOmega3DhaG = ptr(nutrients.Omega3Dha)
	item.OriginalOmega6G = ptr(nutrients.Omega6)
	item.OriginalAlcoholG = ptr(nutrients.Alcohol)
	item.OriginalCaffeineMg = ptr(nutrients.Caffeine)
	item.OriginalCreatineMg = ptr(nutrients.Creatine)
}

// ConvertNutrientsToPer100gCacheFields fills per-100g cache fields from nutrition data
func ConvertNutrientsToPer100gCacheFields(item *storage.Item, nutrients CompleteNutrient, grams float64, converter *UnitConverter) {
	// Helper function to convert and round
	convertAndRound := func(servingValue float64, precision int) float64 {
		return RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(servingValue, grams), precision)
	}
	item.CaloriesPer100g = convertAndRound(nutrients.Calories, 2)
	item.ProteinGPer100g = convertAndRound(nutrients.Protein, 2)
	item.TotalFatGPer100g = convertAndRound(nutrients.TotalFat, 2)
	item.SaturatedFatGPer100g = convertAndRound(nutrients.SaturatedFat, 2)
	item.TransFatGPer100g = convertAndRound(nutrients.TransFat, 2)
	item.MonounsaturatedFatGPer100g = convertAndRound(nutrients.MonounsaturatedFat, 2)
	item.PolyunsaturatedFatGPer100g = convertAndRound(nutrients.PolyunsaturatedFat, 2)
	item.CholesterolMgPer100g = convertAndRound(nutrients.Cholesterol, 2)
	item.SodiumMgPer100g = convertAndRound(nutrients.Sodium, 1)
	item.TotalCarbsGPer100g = convertAndRound(nutrients.TotalCarbs, 1)
	item.DietaryFiberGPer100g = convertAndRound(nutrients.DietaryFiber, 1)
	item.TotalSugarsGPer100g = convertAndRound(nutrients.TotalSugars, 1)
	item.AddedSugarsGPer100g = convertAndRound(nutrients.AddedSugars, 1)
	item.VitaminAMcgPer100g = convertAndRound(nutrients.VitaminA, 1)
	item.VitaminCMgPer100g = convertAndRound(nutrients.VitaminC, 1)
	item.VitaminDMcgPer100g = convertAndRound(nutrients.VitaminD, 1)
	item.VitaminEMgPer100g = convertAndRound(nutrients.VitaminE, 1)
	item.VitaminKMcgPer100g = convertAndRound(nutrients.VitaminK, 1)
	item.ThiamineMgPer100g = convertAndRound(nutrients.Thiamine, 1)
	item.RiboflavinMgPer100g = convertAndRound(nutrients.Riboflavin, 1)
	item.NiacinMgPer100g = convertAndRound(nutrients.Niacin, 1)
	item.VitaminB6MgPer100g = convertAndRound(nutrients.VitaminB6, 1)
	item.FolateMcgPer100g = convertAndRound(nutrients.Folate, 1)
	item.VitaminB12McgPer100g = convertAndRound(nutrients.VitaminB12, 1)
	item.BiotinMcgPer100g = convertAndRound(nutrients.Biotin, 1)
	item.PantothenicAcidMgPer100g = convertAndRound(nutrients.PantothenicAcid, 1)
	item.CholineMgPer100g = convertAndRound(nutrients.Choline, 1)
	item.CalciumMgPer100g = convertAndRound(nutrients.Calcium, 1)
	item.IronMgPer100g = convertAndRound(nutrients.Iron, 1)
	item.MagnesiumMgPer100g = convertAndRound(nutrients.Magnesium, 1)
	item.PhosphorusMgPer100g = convertAndRound(nutrients.Phosphorus, 1)
	item.PotassiumMgPer100g = convertAndRound(nutrients.Potassium, 1)
	item.ZincMgPer100g = convertAndRound(nutrients.Zinc, 1)
	item.CopperMgPer100g = convertAndRound(nutrients.Copper, 1)
	item.ManganeseMgPer100g = convertAndRound(nutrients.Manganese, 1)
	item.SeleniumMcgPer100g = convertAndRound(nutrients.Selenium, 1)
	item.IodineMcgPer100g = convertAndRound(nutrients.Iodine, 1)
	item.MolybdenumMcgPer100g = convertAndRound(nutrients.Molybdenum, 1)
	item.ChromiumMcgPer100g = convertAndRound(nutrients.Chromium, 1)
	item.FluorideMgPer100g = convertAndRound(nutrients.Fluoride, 1)
	item.ChlorideMgPer100g = convertAndRound(nutrients.Chloride, 1)
	item.Omega3AlaGPer100g = convertAndRound(nutrients.Omega3Ala, 3)
	item.Omega3EpaGPer100g = convertAndRound(nutrients.Omega3Epa, 3)
	item.Omega3DhaGPer100g = convertAndRound(nutrients.Omega3Dha, 3)
	item.Omega6GPer100g = convertAndRound(nutrients.Omega6, 3)
	item.AlcoholGPer100g = convertAndRound(nutrients.Alcohol, 2)
	item.CaffeineMgPer100g = convertAndRound(nutrients.Caffeine, 2)
	item.CreatineMgPer100g = convertAndRound(nutrients.Creatine, 2)
}

// Per100gSnapshot represents nutrition data in per-100g format
type Per100gSnapshot struct {
	CaloriesPer100g            float64
	ProteinGPer100g            float64
	TotalFatGPer100g           float64
	SaturatedFatGPer100g       float64
	TransFatGPer100g           float64
	MonounsaturatedFatGPer100g float64
	PolyunsaturatedFatGPer100g float64
	CholesterolMgPer100g       float64
	SodiumMgPer100g            float64
	TotalCarbsGPer100g         float64
	DietaryFiberGPer100g       float64
	TotalSugarsGPer100g        float64
	AddedSugarsGPer100g        float64
	VitaminAMcgPer100g         float64
	VitaminCMgPer100g          float64
	VitaminDMcgPer100g         float64
	VitaminEMgPer100g          float64
	VitaminKMcgPer100g         float64
	ThiamineMgPer100g          float64
	RiboflavinMgPer100g        float64
	NiacinMgPer100g            float64
	VitaminB6MgPer100g         float64
	FolateMcgPer100g           float64
	VitaminB12McgPer100g       float64
	BiotinMcgPer100g           float64
	PantothenicAcidMgPer100g   float64
	CholineMgPer100g           float64
	CalciumMgPer100g           float64
	IronMgPer100g              float64
	MagnesiumMgPer100g         float64
	PhosphorusMgPer100g        float64
	PotassiumMgPer100g         float64
	ZincMgPer100g              float64
	CopperMgPer100g            float64
	ManganeseMgPer100g         float64
	SeleniumMcgPer100g         float64
	IodineMcgPer100g           float64
	MolybdenumMcgPer100g       float64
	ChromiumMcgPer100g         float64
	FluorideMgPer100g          float64
	ChlorideMgPer100g          float64
	Omega3AlaGPer100g          float64
	Omega3EpaGPer100g          float64
	Omega3DhaGPer100g          float64
	Omega6GPer100g             float64
	AlcoholGPer100g            float64
	CaffeineMgPer100g          float64
	CreatineMgPer100g          float64
}

// RoundNutrient rounds a nutrient value to its specified precision
func RoundNutrient(value float64, precision int) float64 {
	return RoundToDecimalPlaces(value, precision)
}

// Ptr returns a pointer to the given value (generic helper)
func Ptr[T any](v T) *T {
	return &v
}
