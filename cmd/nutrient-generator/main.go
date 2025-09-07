package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
)

// Nutrient represents a single nutrient definition
type Nutrient struct {
	Key         string `yaml:"key"`
	Type        string `yaml:"type"`
	Unit        string `yaml:"unit"`
	Category    string `yaml:"category"`
	Per100g     bool   `yaml:"per_100g"`
	Original    bool   `yaml:"original"`
	GoFieldName string `yaml:"go_field_name"`
	JSONTag     string `yaml:"json_tag"`
	DisplayName string `yaml:"display_name"`
}

// NutrientConfig represents the entire nutrients configuration
type NutrientConfig struct {
	Nutrients []Nutrient `yaml:"nutrients"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <yaml-file> [output-dir]\n", os.Args[0])
		os.Exit(1)
	}

	yamlFile := os.Args[1]
	outputDir := "."
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}

	// Load nutrient definitions
	config, err := loadNutrientConfig(yamlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Generate Go code
	if err := generateGoTypes(config, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Go types: %v\n", err)
		os.Exit(1)
	}

	// Generate TypeScript types
	if err := generateTypeScriptTypes(config, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating TypeScript types: %v\n", err)
		os.Exit(1)
	}

	// Generate query builder mappings
	if err := generateQueryBuilderMappings(config, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating query builder mappings: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Code generation completed successfully!")
}

func loadNutrientConfig(filename string) (*NutrientConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var config NutrientConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	return &config, nil
}

func generateGoTypes(config *NutrientConfig, outputDir string) error {
	// Generate CompleteNutrient struct for internal/server/types.go
	if err := generateCompleteNutrientStruct(config, filepath.Join(outputDir, "internal/server")); err != nil {
		return fmt.Errorf("generating CompleteNutrient: %w", err)
	}

	// Generate nutrition service helpers for internal/server/
	if err := generateNutritionServiceHelpers(config, filepath.Join(outputDir, "internal/server")); err != nil {
		return fmt.Errorf("generating nutrition service helpers: %w", err)
	}

	// Generate Item struct fields for internal/storage/store.go
	if err := generateItemStructFields(config, filepath.Join(outputDir, "internal/storage")); err != nil {
		return fmt.Errorf("generating Item fields: %w", err)
	}

	return nil
}

func generateCompleteNutrientStruct(config *NutrientConfig, outputDir string) error {
	tmpl := `// CompleteNutrient represents all nutrition data for a food item (per serving)
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from config/nutrients.yml - see internal/nutrients/generator.go
type CompleteNutrient struct {
{{- range .Nutrients}}
{{- if and .Per100g (not (eq .Key "serving_grams"))}}
	{{.GoFieldName}} {{.Type}} ` + "`json:\"{{.JSONTag}}\"`" + ` // {{.DisplayName}} ({{.Unit}})
{{- end}}
{{- end}}
}
`

	t, err := template.New("completeNutrient").Parse(tmpl)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	outputFile := filepath.Join(outputDir, "nutrient_types_gen.go")
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write package header
	fmt.Fprintf(f, "package server\n\n")

	return t.Execute(f, config)
}

func generateItemStructFields(config *NutrientConfig, outputDir string) error {
	tmpl := `// Generated Item struct nutrition fields
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from config/nutrients.yml - see internal/nutrients/generator.go

package storage

// ItemNutrientFields contains all the nutrition fields for the Item struct
// Copy these into the Item struct definition
/*
// Original serving data fields
{{range .Nutrients}}
{{- if .Original}}
	{{if eq .Key "serving_grams"}}Original{{.GoFieldName}}{{else}}Original{{.GoFieldName}}{{end}} *{{.Type}} ` + "`json:\"original_{{.JSONTag}},omitempty\"`" + ` // {{.DisplayName}}
{{- end}}
{{- end}}

// Normalized nutrition data per 100g fields  
{{range .Nutrients}}
{{- if and .Per100g (not (eq .Key "serving_grams"))}}
	{{.GoFieldName}}Per100g {{.Type}} ` + "`json:\"{{.JSONTag}}_per_100g\"`" + ` // {{.DisplayName}} per 100g
{{- end}}
{{- end}}
*/
`

	t, err := template.New("itemFields").Parse(tmpl)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	outputFile := filepath.Join(outputDir, "item_fields_gen.go")
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, config)
}

func generateTypeScriptTypes(config *NutrientConfig, outputDir string) error {
	tmpl := `// Nutrient type definitions
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from nutrients.yaml

export interface CompleteNutrient {
{{- range .Nutrients}}
{{- if and .Per100g (not (eq .Key "serving_grams"))}}
  {{.JSONTag}}: number; // {{.DisplayName}} ({{.Unit}})
{{- end}}
{{- end}}
}

export interface ItemNutrients {
  // Original serving data
{{- range .Nutrients}}
{{- if .Original}}
  original_{{.JSONTag}}?: number | null; // {{.DisplayName}}
{{- end}}
{{- end}}

  // Per 100g normalized data
{{- range .Nutrients}}
{{- if and .Per100g (not (eq .Key "serving_grams"))}}
  {{.JSONTag}}_per_100g: number; // {{.DisplayName}} per 100g
{{- end}}
{{- end}}
}
`

	t, err := template.New("typescript").Parse(tmpl)
	if err != nil {
		return err
	}

	// Create TypeScript output directory
	tsOutputDir := filepath.Join(outputDir, "apps/web/src/lib/types")
	if err := os.MkdirAll(tsOutputDir, 0755); err != nil {
		return err
	}

	outputFile := filepath.Join(tsOutputDir, "nutrients.gen.ts")
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, config)
}

func generateQueryBuilderMappings(config *NutrientConfig, outputDir string) error {
	tmpl := `// Nutrient field mappings for query builder
// THIS FILE IS GENERATED - DO NOT EDIT MANUALLY
// Generated from nutrients.yaml

package storage

// GetNutrientFields returns all nutrition-related database columns and their corresponding struct field names
func GetNutrientFields() []NutrientField {
	per100gFields := []NutrientField{
{{- range .Nutrients}}
{{- if and .Per100g (not (eq .Key "serving_grams"))}}
		{"{{.JSONTag}}_per_100g", "{{getItemFieldName . false}}Per100g"},
{{- end}}
{{- end}}
	}

	originalFields := []NutrientField{
{{- range .Nutrients}}
{{- if .Original}}
		{"original_{{.JSONTag}}", "Original{{getItemFieldName . true}}"},
{{- end}}
{{- end}}
	}

	// Combine both sets
	return append(per100gFields, originalFields...)
}
`

	// Helper function to get the correct field name for Item struct
	funcMap := template.FuncMap{
		"getItemFieldName": func(nutrient Nutrient, isOriginal bool) string {
			fieldName := nutrient.GoFieldName

			// Add unit suffix for both original and per-100g fields (except calories)
			switch {
			case contains(nutrient.JSONTag, "_g") && !strings.Contains(fieldName, "G"):
				if fieldName == "Calories" {
					return fieldName // Calories doesn't get G suffix
				}
				return fieldName + "G"
			case contains(nutrient.JSONTag, "_mg") && !strings.Contains(fieldName, "Mg"):
				return fieldName + "Mg"
			case contains(nutrient.JSONTag, "_mcg") && !strings.Contains(fieldName, "Mcg"):
				return fieldName + "Mcg"
			}
			return fieldName
		},
	}

	t, err := template.New("queryBuilder").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	storageOutputDir := filepath.Join(outputDir, "internal/storage")
	if err := os.MkdirAll(storageOutputDir, 0755); err != nil {
		return err
	}

	outputFile := filepath.Join(storageOutputDir, "query_builder_gen.go")
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, config)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// generateNutritionServiceHelpers generates helper functions for nutrition service
func generateNutritionServiceHelpers(config *NutrientConfig, outputDir string) error {
	// Get per-100g nutrients (excluding serving_grams)
	var per100gNutrients []Nutrient
	for _, nutrient := range config.Nutrients {
		if nutrient.Per100g && nutrient.Key != "serving_grams" {
			per100gNutrients = append(per100gNutrients, nutrient)
		}
	}

	tmpl := `// Generated nutrition service helpers
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
{{- range .Per100gNutrients}}
	{
		Key:         "{{.Key}}",
		GoField:     "{{.GoFieldName}}",
		Precision:   {{getPrecisionForNutrient .Key}},
		Category:    "{{.Category}}",
		Unit:        "{{.Unit}}",
		DisplayName: "{{.DisplayName}}",
	},
{{- end}}
}

// NutrientPrecisionMap maps nutrient keys to their precision values
var NutrientPrecisionMap = map[string]int{
{{- range .Per100gNutrients}}
	"{{.Key}}": {{getPrecisionForNutrient .Key}},
{{- end}}
}

// Scale multiplies all nutrient values by the given factor
func (c *CompleteNutrient) Scale(factor float64) {
{{- range .Per100gNutrients}}
	c.{{.GoFieldName}} *= factor
{{- end}}
}

// CopyFrom copies all nutrient values from source to destination
func (dst *CompleteNutrient) CopyFrom(src *CompleteNutrient) {
{{- range .Per100gNutrients}}
	dst.{{.GoFieldName}} = src.{{.GoFieldName}}
{{- end}}
}

// ConvertPer100gToServing converts per-100g nutrition data to actual serving size
func ConvertPer100gToServing(cached *storage.Item, grams float64, converter *UnitConverter) CompleteNutrient {
	// Helper function to convert and round in one step
	convertAndRound := func(per100gValue float64, precision int) float64 {
		return RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(per100gValue, grams), precision)
	}

	return CompleteNutrient{
{{- range .Per100gNutrients}}
		{{.GoFieldName}}: convertAndRound(cached.{{.GoFieldName}}{{getUnitSuffix .Unit}}Per100g, {{getPrecisionForNutrient .Key}}),
{{- end}}
	}
}

// ConvertServingToPer100g converts serving nutrition data to per-100g format
func ConvertServingToPer100g(nutrients CompleteNutrient, grams float64, converter *UnitConverter) Per100gSnapshot {
	// Helper function to convert and round in one step
	convertAndRound := func(servingValue float64, precision int) float64 {
		return RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(servingValue, grams), precision)
	}

	return Per100gSnapshot{
{{- range .Per100gNutrients}}
		{{.GoFieldName}}{{getUnitSuffix .Unit}}Per100g: convertAndRound(nutrients.{{.GoFieldName}}, {{getPrecisionForNutrient .Key}}),
{{- end}}
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
{{- range .Per100gNutrients}}
			{{.GoFieldName}}: floatValue(cached.Original{{.GoFieldName}}{{getUnitSuffix .Unit}}),
{{- end}}
		}
	}

	// Fallback to per-100g data (shouldn't happen with new schema)
	return CompleteNutrient{
{{- range .Per100gNutrients}}
		{{.GoFieldName}}: cached.{{.GoFieldName}}{{getUnitSuffix .Unit}}Per100g,
{{- end}}
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
{{- range .Per100gNutrients}}
		{{.GoFieldName}}: {{if eq .Key "calories"}}float64(RoundCaloriesUp(scaleValue(cachedItem.Original{{.GoFieldName}}{{getUnitSuffix .Unit}}))){{else}}scaleValue(cachedItem.Original{{.GoFieldName}}{{getUnitSuffix .Unit}}){{end}},
{{- end}}
	}
}

// ConvertNutrientsToExactCacheFields fills exact cache fields from nutrition data
func ConvertNutrientsToExactCacheFields(item *storage.Item, nutrients CompleteNutrient, grams float64) {
	// Helper to create float64 pointer
	ptr := func(val float64) *float64 { return &val }
	
	// Set original serving data
	item.OriginalServingGrams = ptr(grams)
{{- range .Per100gNutrients}}
	item.Original{{.GoFieldName}}{{getUnitSuffix .Unit}} = ptr(nutrients.{{.GoFieldName}})
{{- end}}
}

// ConvertNutrientsToPer100gCacheFields fills per-100g cache fields from nutrition data
func ConvertNutrientsToPer100gCacheFields(item *storage.Item, nutrients CompleteNutrient, grams float64, converter *UnitConverter) {
	// Helper function to convert and round
	convertAndRound := func(servingValue float64, precision int) float64 {
		return RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(servingValue, grams), precision)
	}
	
{{- range .Per100gNutrients}}
	item.{{.GoFieldName}}{{getUnitSuffix .Unit}}Per100g = convertAndRound(nutrients.{{.GoFieldName}}, {{getPrecisionForCategory .Category}})
{{- end}}
}

// Per100gSnapshot represents nutrition data in per-100g format
type Per100gSnapshot struct {
{{- range .Per100gNutrients}}
	{{.GoFieldName}}{{getUnitSuffix .Unit}}Per100g float64
{{- end}}
}

// RoundNutrient rounds a nutrient value to its specified precision
func RoundNutrient(value float64, precision int) float64 {
	return RoundToDecimalPlaces(value, precision)
}

// Ptr returns a pointer to the given value (generic helper)
func Ptr[T any](v T) *T {
	return &v
}
`

	// Create template with helper functions
	funcMap := template.FuncMap{
		"getPrecisionForCategory": getPrecisionForCategory,
		"getPrecisionForNutrient": getPrecisionForNutrient,
		"getUnitSuffix":           getUnitSuffix,
	}

	t, err := template.New("nutritionHelpers").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(outputDir, "nutrients_generated.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	data := struct {
		Per100gNutrients []Nutrient
	}{
		Per100gNutrients: per100gNutrients,
	}

	return t.Execute(file, data)
}

// getPrecisionForCategory returns appropriate precision based on nutrient category
func getPrecisionForCategory(category string) int {
	switch category {
	case "macro":
		return 2 // Protein, carbs, fat
	case "fat":
		return 2 // All fats
	case "mineral":
		return 1 // Most minerals
	case "vitamin":
		return 1 // Most vitamins
	case "compound":
		return 2 // Alcohol, caffeine, creatine
	case "fatty_acid":
		return 3 // Omega fatty acids
	case "carb":
		return 1 // Carbs, fiber, sugars
	default:
		return 2 // Default precision
	}
}

// getPrecisionForNutrient returns precise precision based on specific nutrient
func getPrecisionForNutrient(key string) int {
	switch key {
	case "calories":
		return 3
	case "protein_g":
		return 2
	case "total_fat_g", "saturated_fat_g", "monounsaturated_fat_g", "polyunsaturated_fat_g":
		return 2
	case "trans_fat_g":
		return 1
	case "cholesterol_mg":
		return 1
	case "sodium_mg":
		return 1
	case "total_carbs_g", "dietary_fiber_g", "total_sugars_g", "added_sugars_g":
		return 1
	case "thiamine_mg", "riboflavin_mg", "vitamin_b6_mg":
		return 3 // B vitamins need higher precision
	case "vitamin_b12_mcg":
		return 2
	case "zinc_mg":
		return 2
	case "omega3_ala_g", "omega3_epa_g", "omega3_dha_g":
		return 3
	case "omega6_g":
		return 2
	case "alcohol_g":
		return 2
	case "copper_mg", "manganese_mg":
		return 3 // Trace elements
	default:
		// Use category-based precision for others
		return getPrecisionForCategory("default")
	}
}

// getUnitSuffix returns the appropriate field suffix based on unit
func getUnitSuffix(unit string) string {
	switch unit {
	case "g", "grams":
		return "G"
	case "mg":
		return "Mg"
	case "mcg":
		return "Mcg"
	case "kcal":
		return "" // Calories has no suffix
	default:
		return ""
	}
}
