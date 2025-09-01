package storage

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// NutrientField represents a nutrition field with its database column name and struct field name
type NutrientField struct {
	DBColumn    string
	StructField string
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

	// Additional metadata fields
	additionalColumns := []string{
		"note", "ingredients", "url",
	}

	metaColumns := []string{
		"created_at", "updated_at",
	}

	// Combine all columns
	allColumns := make([]string, 0, len(baseColumns)+len(nutrientColumns)+len(additionalColumns)+len(metaColumns))
	allColumns = append(allColumns, baseColumns...)
	allColumns = append(allColumns, nutrientColumns...)
	allColumns = append(allColumns, additionalColumns...)
	allColumns = append(allColumns, metaColumns...)

	return allColumns
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

	// Additional metadata fields
	// Serialize ingredients to JSON for database storage
	var ingredientsJSON []byte
	if len(item.Ingredients) > 0 {
		var err error
		ingredientsJSON, err = json.Marshal(item.Ingredients)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal ingredients: %v", err))
		}
	}
	values = append(values, item.Note, ingredientsJSON, item.Url)

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

	// Meta columns - additional fields come before timestamps
	additionalIdx := 5 + len(nutrientFields)

	// Additional metadata fields (note, ingredients, url)
	var ingredientsJSON []byte
	scanArgs[additionalIdx] = &item.Note
	scanArgs[additionalIdx+1] = &ingredientsJSON // Will be deserialized after scan
	scanArgs[additionalIdx+2] = &item.Url

	// Timestamps
	createdAtIdx := additionalIdx + 3
	updatedAtIdx := createdAtIdx + 1

	scanArgs[createdAtIdx] = &item.CreatedAt
	scanArgs[updatedAtIdx] = &item.UpdatedAt

	// Perform the scan
	err := row.Scan(scanArgs...)
	if err != nil {
		return err
	}

	// Deserialize ingredients JSON if present
	if len(ingredientsJSON) > 0 {
		err = json.Unmarshal(ingredientsJSON, &item.Ingredients)
		if err != nil {
			return fmt.Errorf("failed to unmarshal ingredients: %w", err)
		}
	}

	return nil
}

// scannable interface to abstract sql.Row and sql.Rows
type scannable interface {
	Scan(dest ...interface{}) error
}
