package storage

import (
	"fmt"
	"reflect"
	"strings"
)

// NutrientField represents a nutrition field with its database column name and struct field name
type NutrientField struct {
	DBColumn    string
	StructField string
}

// GetNutrientFields returns all nutrition-related database columns and their corresponding struct field names
func GetNutrientFields() []NutrientField {
	per100gFields := []NutrientField{
		{"calories_per_100g", "CaloriesPer100g"},
		{"protein_g_per_100g", "ProteinGPer100g"},
		{"total_fat_g_per_100g", "TotalFatGPer100g"},
		{"saturated_fat_g_per_100g", "SaturatedFatGPer100g"},
		{"trans_fat_g_per_100g", "TransFatGPer100g"},
		{"cholesterol_mg_per_100g", "CholesterolMgPer100g"},
		{"sodium_mg_per_100g", "SodiumMgPer100g"},
		{"total_carbs_g_per_100g", "TotalCarbsGPer100g"},
		{"dietary_fiber_g_per_100g", "DietaryFiberGPer100g"},
		{"total_sugars_g_per_100g", "TotalSugarsGPer100g"},
		{"added_sugars_g_per_100g", "AddedSugarsGPer100g"},
		{"vitamin_a_mcg_per_100g", "VitaminAMcgPer100g"},
		{"vitamin_c_mg_per_100g", "VitaminCMgPer100g"},
		{"vitamin_d_mcg_per_100g", "VitaminDMcgPer100g"},
		{"vitamin_e_mg_per_100g", "VitaminEMgPer100g"},
		{"vitamin_k_mcg_per_100g", "VitaminKMcgPer100g"},
		{"thiamine_mg_per_100g", "ThiamineMgPer100g"},
		{"riboflavin_mg_per_100g", "RiboflavinMgPer100g"},
		{"niacin_mg_per_100g", "NiacinMgPer100g"},
		{"vitamin_b6_mg_per_100g", "VitaminB6MgPer100g"},
		{"folate_mcg_per_100g", "FolateMcgPer100g"},
		{"vitamin_b12_mcg_per_100g", "VitaminB12McgPer100g"},
		{"biotin_mcg_per_100g", "BiotinMcgPer100g"},
		{"pantothenic_acid_mg_per_100g", "PantothenicAcidMgPer100g"},
		{"choline_mg_per_100g", "CholineMgPer100g"},
		{"calcium_mg_per_100g", "CalciumMgPer100g"},
		{"iron_mg_per_100g", "IronMgPer100g"},
		{"magnesium_mg_per_100g", "MagnesiumMgPer100g"},
		{"phosphorus_mg_per_100g", "PhosphorusMgPer100g"},
		{"potassium_mg_per_100g", "PotassiumMgPer100g"},
		{"zinc_mg_per_100g", "ZincMgPer100g"},
		{"copper_mg_per_100g", "CopperMgPer100g"},
		{"manganese_mg_per_100g", "ManganeseMgPer100g"},
		{"selenium_mcg_per_100g", "SeleniumMcgPer100g"},
		{"iodine_mcg_per_100g", "IodineMcgPer100g"},
		{"molybdenum_mcg_per_100g", "MolybdenumMcgPer100g"},
		{"chromium_mcg_per_100g", "ChromiumMcgPer100g"},
		{"fluoride_mg_per_100g", "FluorideMgPer100g"},
		{"chloride_mg_per_100g", "ChlorideMgPer100g"},
		{"monounsaturated_fat_g_per_100g", "MonounsaturatedFatGPer100g"},
		{"polyunsaturated_fat_g_per_100g", "PolyunsaturatedFatGPer100g"},
		{"omega3_ala_g_per_100g", "Omega3AlaGPer100g"},
		{"omega3_epa_g_per_100g", "Omega3EpaGPer100g"},
		{"omega3_dha_g_per_100g", "Omega3DhaGPer100g"},
		{"omega6_g_per_100g", "Omega6GPer100g"},
		{"alcohol_g_per_100g", "AlcoholGPer100g"},
		{"caffeine_mg_per_100g", "CaffeineMgPer100g"},
		{"creatine_mg_per_100g", "CreatineMgPer100g"},
	}

	originalFields := []NutrientField{
		{"original_serving_grams", "OriginalServingGrams"},
		{"original_calories", "OriginalCalories"},
		{"original_protein_g", "OriginalProteinG"},
		{"original_total_fat_g", "OriginalTotalFatG"},
		{"original_saturated_fat_g", "OriginalSaturatedFatG"},
		{"original_trans_fat_g", "OriginalTransFatG"},
		{"original_cholesterol_mg", "OriginalCholesterolMg"},
		{"original_sodium_mg", "OriginalSodiumMg"},
		{"original_total_carbs_g", "OriginalTotalCarbsG"},
		{"original_dietary_fiber_g", "OriginalDietaryFiberG"},
		{"original_total_sugars_g", "OriginalTotalSugarsG"},
		{"original_added_sugars_g", "OriginalAddedSugarsG"},
		{"original_vitamin_a_mcg", "OriginalVitaminAMcg"},
		{"original_vitamin_c_mg", "OriginalVitaminCMg"},
		{"original_vitamin_d_mcg", "OriginalVitaminDMcg"},
		{"original_vitamin_e_mg", "OriginalVitaminEMg"},
		{"original_vitamin_k_mcg", "OriginalVitaminKMcg"},
		{"original_thiamine_mg", "OriginalThiamineMg"},
		{"original_riboflavin_mg", "OriginalRiboflavinMg"},
		{"original_niacin_mg", "OriginalNiacinMg"},
		{"original_vitamin_b6_mg", "OriginalVitaminB6Mg"},
		{"original_folate_mcg", "OriginalFolateMcg"},
		{"original_vitamin_b12_mcg", "OriginalVitaminB12Mcg"},
		{"original_biotin_mcg", "OriginalBiotinMcg"},
		{"original_pantothenic_acid_mg", "OriginalPantothenicAcidMg"},
		{"original_choline_mg", "OriginalCholineMg"},
		{"original_calcium_mg", "OriginalCalciumMg"},
		{"original_iron_mg", "OriginalIronMg"},
		{"original_magnesium_mg", "OriginalMagnesiumMg"},
		{"original_phosphorus_mg", "OriginalPhosphorusMg"},
		{"original_potassium_mg", "OriginalPotassiumMg"},
		{"original_zinc_mg", "OriginalZincMg"},
		{"original_copper_mg", "OriginalCopperMg"},
		{"original_manganese_mg", "OriginalManganeseMg"},
		{"original_selenium_mcg", "OriginalSeleniumMcg"},
		{"original_iodine_mcg", "OriginalIodineMcg"},
		{"original_molybdenum_mcg", "OriginalMolybdenumMcg"},
		{"original_chromium_mcg", "OriginalChromiumMcg"},
		{"original_fluoride_mg", "OriginalFluorideMg"},
		{"original_chloride_mg", "OriginalChlorideMg"},
		{"original_monounsaturated_fat_g", "OriginalMonounsaturatedFatG"},
		{"original_polyunsaturated_fat_g", "OriginalPolyunsaturatedFatG"},
		{"original_omega3_ala_g", "OriginalOmega3AlaG"},
		{"original_omega3_epa_g", "OriginalOmega3EpaG"},
		{"original_omega3_dha_g", "OriginalOmega3DhaG"},
		{"original_omega6_g", "OriginalOmega6G"},
		{"original_alcohol_g", "OriginalAlcoholG"},
		{"original_caffeine_mg", "OriginalCaffeineMg"},
		{"original_creatine_mg", "OriginalCreatineMg"},
	}

	// Combine both sets
	return append(per100gFields, originalFields...)
}

// GetItemColumns returns all column names for the items table in the correct order
func GetItemColumns() []string {
	baseColumns := []string{
		"id", "normalized_name", "normalized_brand", "display_name", "display_brand",
	}

	nutrientFields := GetNutrientFields()
	nutrientColumns := make([]string, len(nutrientFields))
	for i, field := range nutrientFields {
		nutrientColumns[i] = field.DBColumn
	}

	metaColumns := []string{
		"created_at", "updated_at",
	}

	// Combine all columns
	allColumns := make([]string, 0, len(baseColumns)+len(nutrientColumns)+len(metaColumns))
	allColumns = append(allColumns, baseColumns...)
	allColumns = append(allColumns, nutrientColumns...)
	allColumns = append(allColumns, metaColumns...)

	return allColumns
}

// buildInsertQuery builds the INSERT query for items table
func buildInsertQuery() string {
	columns := GetItemColumns()
	columnsList := strings.Join(columns, ", ")
	placeholders := strings.Repeat("?, ", len(columns)-1) + "?"

	return fmt.Sprintf(`
		INSERT INTO items (%s) VALUES (%s)`, columnsList, placeholders)
}

// buildSelectQuery builds the SELECT query for items table
func buildSelectQuery(where string) string {
	columns := GetItemColumns()
	columnsList := strings.Join(columns, ", ")

	query := fmt.Sprintf("SELECT %s FROM items", columnsList)
	if where != "" {
		query += " WHERE " + where
	}
	return query
}

// buildUpdateQuery builds the UPDATE query for items table
func buildUpdateQuery() string {
	columns := GetItemColumns()
	// Skip id, created_at for updates
	updateColumns := make([]string, 0, len(columns)-2)
	for _, col := range columns {
		if col != "id" && col != "created_at" {
			updateColumns = append(updateColumns, col+" = ?")
		}
	}

	return fmt.Sprintf(`
		UPDATE items SET %s WHERE id = ?`, strings.Join(updateColumns, ", "))
}

// extractItemValues extracts values from an Item struct in the correct order for SQL operations
func extractItemValues(item *Item, includeID bool, includeCreatedAt bool) []interface{} {
	values := make([]interface{}, 0, 88) // Pre-allocate for all columns

	// Base columns
	if includeID {
		values = append(values, item.ID)
	}
	values = append(values,
		item.NormalizedName, item.NormalizedBrand, item.DisplayName, item.DisplayBrand,
	)

	// Use reflection to get nutrition field values
	itemValue := reflect.ValueOf(item).Elem() // Dereference pointer
	nutrientFields := GetNutrientFields()

	for _, field := range nutrientFields {
		fieldValue := itemValue.FieldByName(field.StructField)
		if !fieldValue.IsValid() {
			panic(fmt.Sprintf("field %s not found in Item struct", field.StructField))
		}
		values = append(values, fieldValue.Interface())
	}

	// Meta columns
	if includeCreatedAt {
		values = append(values, item.CreatedAt)
	}
	values = append(values, item.UpdatedAt)

	return values
}

// scanItemRow scans a database row into an Item struct
func scanItemRow(row scannable, item *Item) error {
	// Create slice to hold all the scan targets
	columns := GetItemColumns()
	scanArgs := make([]interface{}, len(columns))

	// Base columns
	scanArgs[0] = &item.ID
	scanArgs[1] = &item.NormalizedName
	scanArgs[2] = &item.NormalizedBrand
	scanArgs[3] = &item.DisplayName
	scanArgs[4] = &item.DisplayBrand

	// Use reflection to set nutrition field scan targets
	itemValue := reflect.ValueOf(item).Elem()
	nutrientFields := GetNutrientFields()

	for i, field := range nutrientFields {
		fieldValue := itemValue.FieldByName(field.StructField)
		if !fieldValue.IsValid() {
			return fmt.Errorf("field %s not found in Item struct", field.StructField)
		}
		if !fieldValue.CanAddr() {
			return fmt.Errorf("field %s cannot be addressed", field.StructField)
		}
		scanArgs[5+i] = fieldValue.Addr().Interface()
	}

	// Meta columns
	createdAtIdx := 5 + len(nutrientFields)
	updatedAtIdx := createdAtIdx + 1

	scanArgs[createdAtIdx] = &item.CreatedAt
	scanArgs[updatedAtIdx] = &item.UpdatedAt

	return row.Scan(scanArgs...)
}

// scannable interface to abstract sql.Row and sql.Rows
type scannable interface {
	Scan(dest ...interface{}) error
}
