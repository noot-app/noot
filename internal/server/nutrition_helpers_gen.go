// Nutrition service helper functions and constants
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from config/nutrients.yml

package server

import (
	"github.com/grantbirki/noot/internal/storage"
	"math"
	"reflect"
)

// NutrientPrecision maps field names to their rounding precision
var NutrientPrecision = map[string]int{
	"Calories":           3,
	"Protein":            2,
	"TotalFat":           2,
	"SaturatedFat":       2,
	"TransFat":           1,
	"MonounsaturatedFat": 2,
	"PolyunsaturatedFat": 2,
	"Cholesterol":        1,
	"Sodium":             1,
	"TotalCarbs":         1,
	"DietaryFiber":       1,
	"TotalSugars":        1,
	"AddedSugars":        1,
	"VitaminA":           1,
	"VitaminC":           1,
	"VitaminD":           1,
	"VitaminE":           1,
	"VitaminK":           1,
	"Thiamine":           3,
	"Riboflavin":         3,
	"Niacin":             3,
	"VitaminB6":          3,
	"Folate":             3,
	"VitaminB12":         2,
	"Biotin":             3,
	"PantothenicAcid":    3,
	"Choline":            3,
	"Calcium":            1,
	"Iron":               1,
	"Magnesium":          1,
	"Phosphorus":         1,
	"Potassium":          1,
	"Zinc":               2,
	"Copper":             1,
	"Manganese":          1,
	"Selenium":           3,
	"Iodine":             3,
	"Molybdenum":         3,
	"Chromium":           3,
	"Fluoride":           1,
	"Chloride":           1,
	"Omega3Ala":          3,
	"Omega3Epa":          3,
	"Omega3Dha":          3,
	"Omega6":             2,
	"Alcohol":            2,
	"Caffeine":           1,
	"Creatine":           1,
}

// ScaleNutritionDataGenerated scales nutrition data using reflection for DRY approach
func ScaleNutritionDataGenerated(cached *storage.Item, factor float64) CompleteNutrient {
	var result CompleteNutrient

	cachedVal := reflect.ValueOf(cached).Elem()
	resultVal := reflect.ValueOf(&result).Elem()
	if field := cachedVal.FieldByName("CaloriesPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Calories"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Calories").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ProteinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Protein"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Protein").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["TotalFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("TotalFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["SaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("SaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TransFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["TransFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("TransFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MonounsaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["MonounsaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("MonounsaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PolyunsaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["PolyunsaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("PolyunsaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CholesterolPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Cholesterol"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Cholesterol").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SodiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Sodium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Sodium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalCarbsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["TotalCarbs"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("TotalCarbs").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("DietaryFiberPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["DietaryFiber"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("DietaryFiber").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalSugarsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["TotalSugars"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("TotalSugars").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("AddedSugarsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["AddedSugars"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("AddedSugars").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminAPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["VitaminA"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("VitaminA").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminCPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["VitaminC"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("VitaminC").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminDPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["VitaminD"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("VitaminD").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminEPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["VitaminE"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("VitaminE").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminKPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["VitaminK"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("VitaminK").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ThiaminePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Thiamine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Thiamine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("RiboflavinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Riboflavin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Riboflavin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("NiacinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Niacin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Niacin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminB6Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["VitaminB6"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("VitaminB6").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("FolatePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Folate"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Folate").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminB12Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["VitaminB12"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("VitaminB12").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("BiotinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Biotin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Biotin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PantothenicAcidPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["PantothenicAcid"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("PantothenicAcid").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CholinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Choline"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Choline").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CalciumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Calcium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Calcium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("IronPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Iron"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Iron").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MagnesiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Magnesium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Magnesium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PhosphorusPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Phosphorus"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Phosphorus").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PotassiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Potassium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Potassium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ZincPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Zinc"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Zinc").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CopperPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Copper"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Copper").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ManganesePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Manganese"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Manganese").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SeleniumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Selenium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Selenium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("IodinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Iodine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Iodine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MolybdenumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Molybdenum"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Molybdenum").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ChromiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Chromium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Chromium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("FluoridePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Fluoride"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Fluoride").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ChloridePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Chloride"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Chloride").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3AlaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Omega3Ala"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Omega3Ala").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3EpaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Omega3Epa"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Omega3Epa").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3DhaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Omega3Dha"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Omega3Dha").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega6Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Omega6"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Omega6").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("AlcoholPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Alcohol"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Alcohol").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CaffeinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Caffeine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Caffeine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CreatinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		scaled := field.Float() * factor
		precision, exists := NutrientPrecision["Creatine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(scaled, precision)
		resultVal.FieldByName("Creatine").SetFloat(rounded)
	}

	return result
}

// ConvertCachedToNutrientsGenerated converts cached data to CompleteNutrient
func ConvertCachedToNutrientsGenerated(cached *storage.Item) CompleteNutrient {
	var result CompleteNutrient

	cachedVal := reflect.ValueOf(cached).Elem()
	resultVal := reflect.ValueOf(&result).Elem()
	if field := cachedVal.FieldByName("CaloriesPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Calories"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Calories").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ProteinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Protein"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Protein").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TotalFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TotalFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["SaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("SaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TransFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TransFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TransFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MonounsaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["MonounsaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("MonounsaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PolyunsaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["PolyunsaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("PolyunsaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CholesterolPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Cholesterol"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Cholesterol").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SodiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Sodium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Sodium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalCarbsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TotalCarbs"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TotalCarbs").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("DietaryFiberPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["DietaryFiber"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("DietaryFiber").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalSugarsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TotalSugars"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TotalSugars").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("AddedSugarsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["AddedSugars"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("AddedSugars").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminAPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminA"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminA").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminCPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminC"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminC").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminDPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminD"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminD").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminEPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminE"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminE").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminKPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminK"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminK").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ThiaminePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Thiamine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Thiamine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("RiboflavinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Riboflavin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Riboflavin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("NiacinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Niacin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Niacin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminB6Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminB6"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminB6").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("FolatePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Folate"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Folate").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminB12Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminB12"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminB12").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("BiotinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Biotin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Biotin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PantothenicAcidPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["PantothenicAcid"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("PantothenicAcid").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CholinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Choline"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Choline").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CalciumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Calcium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Calcium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("IronPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Iron"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Iron").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MagnesiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Magnesium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Magnesium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PhosphorusPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Phosphorus"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Phosphorus").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PotassiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Potassium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Potassium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ZincPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Zinc"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Zinc").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CopperPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Copper"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Copper").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ManganesePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Manganese"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Manganese").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SeleniumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Selenium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Selenium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("IodinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Iodine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Iodine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MolybdenumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Molybdenum"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Molybdenum").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ChromiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Chromium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Chromium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("FluoridePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Fluoride"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Fluoride").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ChloridePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Chloride"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Chloride").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3AlaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega3Ala"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega3Ala").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3EpaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega3Epa"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega3Epa").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3DhaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega3Dha"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega3Dha").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega6Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega6"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega6").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("AlcoholPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Alcohol"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Alcohol").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CaffeinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Caffeine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Caffeine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CreatinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Creatine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Creatine").SetFloat(rounded)
	}

	return result
}

// ConvertExactCachedToNutrientsGenerated converts exact cached data to CompleteNutrient
func ConvertExactCachedToNutrientsGenerated(cached *storage.Item) CompleteNutrient {
	var result CompleteNutrient

	cachedVal := reflect.ValueOf(cached).Elem()
	resultVal := reflect.ValueOf(&result).Elem()
	if field := cachedVal.FieldByName("CaloriesPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Calories"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Calories").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ProteinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Protein"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Protein").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TotalFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TotalFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["SaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("SaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TransFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TransFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TransFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MonounsaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["MonounsaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("MonounsaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PolyunsaturatedFatPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["PolyunsaturatedFat"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("PolyunsaturatedFat").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CholesterolPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Cholesterol"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Cholesterol").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SodiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Sodium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Sodium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalCarbsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TotalCarbs"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TotalCarbs").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("DietaryFiberPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["DietaryFiber"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("DietaryFiber").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("TotalSugarsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["TotalSugars"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("TotalSugars").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("AddedSugarsPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["AddedSugars"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("AddedSugars").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminAPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminA"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminA").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminCPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminC"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminC").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminDPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminD"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminD").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminEPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminE"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminE").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminKPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminK"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminK").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ThiaminePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Thiamine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Thiamine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("RiboflavinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Riboflavin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Riboflavin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("NiacinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Niacin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Niacin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminB6Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminB6"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminB6").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("FolatePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Folate"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Folate").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("VitaminB12Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["VitaminB12"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("VitaminB12").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("BiotinPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Biotin"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Biotin").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PantothenicAcidPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["PantothenicAcid"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("PantothenicAcid").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CholinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Choline"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Choline").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CalciumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Calcium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Calcium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("IronPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Iron"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Iron").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MagnesiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Magnesium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Magnesium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PhosphorusPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Phosphorus"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Phosphorus").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("PotassiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Potassium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Potassium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ZincPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Zinc"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Zinc").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CopperPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Copper"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Copper").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ManganesePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Manganese"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Manganese").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("SeleniumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Selenium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Selenium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("IodinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Iodine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Iodine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("MolybdenumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Molybdenum"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Molybdenum").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ChromiumPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Chromium"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Chromium").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("FluoridePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Fluoride"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Fluoride").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("ChloridePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Chloride"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Chloride").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3AlaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega3Ala"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega3Ala").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3EpaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega3Epa"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega3Epa").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega3DhaPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega3Dha"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega3Dha").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("Omega6Per100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Omega6"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Omega6").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("AlcoholPer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Alcohol"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Alcohol").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CaffeinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Caffeine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Caffeine").SetFloat(rounded)
	}
	if field := cachedVal.FieldByName("CreatinePer100g"); field.IsValid() && field.Kind() == reflect.Float64 {
		value := field.Float()
		precision, exists := NutrientPrecision["Creatine"]
		if !exists {
			precision = 2 // default precision
		}
		rounded := convertAndRound(value, precision)
		resultVal.FieldByName("Creatine").SetFloat(rounded)
	}

	return result
}

// convertAndRound rounds a float64 to the specified decimal places
func convertAndRound(value float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(value*multiplier) / multiplier
}
