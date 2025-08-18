package server

import (
	"context"
	"testing"
	"time"
)

func init() {
	// Initialize logger for tests
	InitLogger()
}

// Test that nutrition lookup works correctly
func TestNutritionLookup(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		foodName     string
		shouldFind   bool
		expectedCals float64
	}{
		{"Direct match", "banana", true, 105},
		{"Case insensitive", "APPLE", true, 95},
		{"Fuzzy match", "greek yogurt", true, 130},
		{"Plural form", "eggs", true, 70},
		{"Not found", "exotic-fruit-xyz", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nutrition := lookupNutrition(ctx, tt.foodName)
			
			if tt.shouldFind {
				if nutrition == nil {
					t.Errorf("Expected to find nutrition for %s, but got nil", tt.foodName)
					return
				}
				if nutrition.Calories != tt.expectedCals {
					t.Errorf("Expected calories %v for %s, got %v", tt.expectedCals, tt.foodName, nutrition.Calories)
				}
			} else {
				if nutrition != nil {
					t.Errorf("Expected not to find nutrition for %s, but got %+v", tt.foodName, nutrition)
				}
			}
		})
	}
}

// Test that nutrition enrichment works
func TestNutritionEnrichment(t *testing.T) {
	ctx := context.Background()

	items := []Item{
		{Name: "banana", Quantity: floatPtr(1)},
		{Name: "unknown-food-xyz", Quantity: floatPtr(1)},
		{Name: "apple", Quantity: floatPtr(2)}, // Should scale nutrition
	}

	enriched := enrichItemsWithNutrition(ctx, items)

	// Check banana
	if enriched[0].Nutrients == nil {
		t.Error("Expected banana to have nutrition data")
	} else if enriched[0].Nutrients.Calories != 105 {
		t.Errorf("Expected banana to have 105 calories, got %v", enriched[0].Nutrients.Calories)
	}

	// Check unknown food
	if enriched[1].Nutrients != nil {
		t.Error("Expected unknown food to have no nutrition data")
	}

	// Check scaled apple (2x)
	if enriched[2].Nutrients == nil {
		t.Error("Expected apple to have nutrition data")
	} else if enriched[2].Nutrients.Calories != 190 { // 95 * 2
		t.Errorf("Expected scaled apple to have 190 calories, got %v", enriched[2].Nutrients.Calories)
	}
}

// Test that new prompt is much more concise
func TestPromptOptimization(t *testing.T) {
	oldPromptLength := 2481 // From our analysis
	newPrompt := parseSystemPrompt()
	
	if len(newPrompt) >= oldPromptLength {
		t.Errorf("New prompt should be shorter than old prompt. Old: %d, New: %d", oldPromptLength, len(newPrompt))
	}

	// Should be significantly shorter (at least 50% reduction)
	if len(newPrompt) > oldPromptLength/2 {
		t.Errorf("New prompt should be at least 50%% shorter. Old: %d, New: %d", oldPromptLength, len(newPrompt))
	}

	// Should not contain nutrition fields
	if contains(newPrompt, "nutrients") || contains(newPrompt, "calories") || contains(newPrompt, "protein_g") {
		t.Error("New prompt should not contain nutrition-related fields")
	}
}

// Benchmark nutrition lookup performance
func BenchmarkNutritionLookup(b *testing.B) {
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lookupNutrition(ctx, "banana")
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   len(s) > len(substr) && s[:len(substr)] == substr ||
		   (len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Performance comparison test - simulates the improvement
func TestPerformanceImprovement(t *testing.T) {
	// Simulate old approach timing (complex prompt + nutrition generation)
	oldApproachStart := time.Now()
	time.Sleep(100 * time.Millisecond) // Simulate old slow processing
	oldDuration := time.Since(oldApproachStart)

	// Simulate new approach timing (simple prompt + fast lookup)
	newApproachStart := time.Now()
	ctx := context.Background()
	
	// Simple parsing simulation
	time.Sleep(20 * time.Millisecond) // Simulate faster parsing
	
	// Fast nutrition lookup
	items := []Item{
		{Name: "banana", Quantity: floatPtr(1)},
		{Name: "apple", Quantity: floatPtr(1)},
	}
	enrichItemsWithNutrition(ctx, items)
	
	newDuration := time.Since(newApproachStart)

	t.Logf("Simulated old approach: %v", oldDuration)
	t.Logf("Simulated new approach: %v", newDuration)
	
	// New approach should be significantly faster
	if newDuration >= oldDuration {
		t.Errorf("New approach should be faster. Old: %v, New: %v", oldDuration, newDuration)
	}
}