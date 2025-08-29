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
