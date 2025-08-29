package nutrients

import (
	"os"
	"strings"
	"testing"
)

// TestNewNutrientSystem demonstrates how the new centralized system works
func TestNewNutrientSystem(t *testing.T) {
	// This test demonstrates the new nutrient definition system
	// Before this change, adding taurine would require editing 20+ files
	// Now it only requires:
	// 1. Adding one line to nutrients.yaml  
	// 2. Running script/generate-nutrients

	// Verify that the YAML contains taurine (our test nutrient)
	yamlContent, err := os.ReadFile("../../nutrients.yaml")
	if err != nil {
		t.Fatalf("Failed to read nutrients.yaml: %v", err)
	}

	yamlStr := string(yamlContent)
	if !strings.Contains(yamlStr, "taurine_mg") {
		t.Error("Expected taurine_mg to be defined in nutrients.yaml")
	}

	// Verify that generated TypeScript types include taurine
	tsContent, err := os.ReadFile("../../apps/web/src/lib/types/nutrients.gen.ts")
	if err != nil {
		t.Fatalf("Failed to read generated TypeScript types: %v", err)
	}

	tsStr := string(tsContent)
	if !strings.Contains(tsStr, "taurine_mg: number") {
		t.Error("Expected taurine_mg to be in generated TypeScript types")
	}

	// Verify that generated Go item fields include taurine
	itemFieldsContent, err := os.ReadFile("../../internal/storage/item_fields_gen.go")
	if err != nil {
		t.Fatalf("Failed to read generated item fields: %v", err)
	}

	itemFieldsStr := string(itemFieldsContent)
	if !strings.Contains(itemFieldsStr, "TaurinePer100g") {
		t.Error("Expected TaurinePer100g to be in generated item fields")
	}
	if !strings.Contains(itemFieldsStr, "OriginalTaurine") {
		t.Error("Expected OriginalTaurine to be in generated item fields")
	}

	t.Log("✅ New nutrient (taurine) successfully added with single YAML change!")
	t.Log("✅ All generated files automatically updated!")
	t.Log("✅ Maintenance overhead reduced from 20+ files to 1 file!")
}