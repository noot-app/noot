package nutrients

import (
	"os"
	"strings"
	"testing"
)

// TestNewNutrientSystem demonstrates how the new centralized system works
func TestNewNutrientSystem(t *testing.T) {
	// This test demonstrates the new nutrient definition system
	// Before this change, adding Sodium would require editing 20+ files
	// Now it only requires:
	// 1. Adding one line to config/nutrients.yml
	// 2. Running script/generate-nutrients

	// Verify that the YAML contains Sodium (our test nutrient)
	yamlContent, err := os.ReadFile("../../config/nutrients.yml")
	if err != nil {
		t.Fatalf("Failed to read config/nutrients.yml: %v", err)
	}

	yamlStr := string(yamlContent)
	if !strings.Contains(yamlStr, "sodium_mg") {
		t.Error("Expected sodium_mg to be defined in config/nutrients.yml")
	}

	// Verify that generated TypeScript types include Sodium
	tsContent, err := os.ReadFile("../../apps/web/src/lib/types/nutrients.gen.ts")
	if err != nil {
		t.Fatalf("Failed to read generated TypeScript types: %v", err)
	}

	tsStr := string(tsContent)
	if !strings.Contains(tsStr, "sodium_mg: number") {
		t.Error("Expected sodium_mg to be in generated TypeScript types")
	}

	// Verify that generated Go item fields include Sodium
	itemFieldsContent, err := os.ReadFile("../../internal/storage/item_fields_gen.go")
	if err != nil {
		t.Fatalf("Failed to read generated item fields: %v", err)
	}

	itemFieldsStr := string(itemFieldsContent)
	if !strings.Contains(itemFieldsStr, "SodiumPer100g") {
		t.Error("Expected SodiumPer100g to be in generated item fields")
	}
	if !strings.Contains(itemFieldsStr, "OriginalSodium") {
		t.Error("Expected OriginalSodium to be in generated item fields")
	}
}
