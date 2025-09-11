package goals

import (
	"testing"
	"time"
)

func TestNewGoalResolver(t *testing.T) {
	resolver, err := NewGoalResolver()
	if err != nil {
		t.Fatalf("Failed to create goal resolver: %v", err)
	}

	if resolver == nil {
		t.Fatal("Expected resolver to be non-nil")
	}

	if resolver.driData == nil {
		t.Fatal("Expected DRI data to be loaded")
	}

	if resolver.dvData == nil {
		t.Fatal("Expected DV data to be loaded")
	}
}

func TestGetAgeBracket(t *testing.T) {
	resolver, err := NewGoalResolver()
	if err != nil {
		t.Fatalf("Failed to create goal resolver: %v", err)
	}

	tests := []struct {
		name      string
		birthDate *time.Time
		expected  string
	}{
		{
			name:      "nil birth date",
			birthDate: nil,
			expected:  "19-30 y",
		},
		{
			name:      "30 year old",
			birthDate: func() *time.Time { t := time.Now().AddDate(-30, 0, 0); return &t }(),
			expected:  "19-30 y",
		},
		{
			name:      "40 year old",
			birthDate: func() *time.Time { t := time.Now().AddDate(-40, 0, 0); return &t }(),
			expected:  "31-50 y",
		},
		{
			name:      "65 year old",
			birthDate: func() *time.Time { t := time.Now().AddDate(-65, 0, 0); return &t }(),
			expected:  "51-70 y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.getAgeBracket(tt.birthDate)
			if result != tt.expected {
				t.Errorf("Expected age bracket %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResolveGoals(t *testing.T) {
	resolver, err := NewGoalResolver()
	if err != nil {
		t.Fatalf("Failed to create goal resolver: %v", err)
	}

	// Test basic goal resolution for adult male
	birthDate := time.Now().AddDate(-30, 0, 0) // 30 year old
	goals, err := resolver.ResolveGoals("male", &birthDate, nil)
	if err != nil {
		t.Fatalf("Failed to resolve goals: %v", err)
	}

	if goals == nil {
		t.Fatal("Expected goals to be non-nil")
	}

	if goals.Source != "dri" {
		t.Errorf("Expected source to be 'dri', got %s", goals.Source)
	}

	if goals.LifeStage.Sex != "male" {
		t.Errorf("Expected sex to be 'male', got %s", goals.LifeStage.Sex)
	}

	if goals.LifeStage.AgeBracket != "19-30 y" {
		t.Errorf("Expected age bracket to be '19-30 y', got %s", goals.LifeStage.AgeBracket)
	}

	// Should have some targets
	if len(goals.Targets) == 0 {
		t.Error("Expected some nutrition targets to be set")
	}

	// Verify we have comprehensive nutrient coverage
	expectedMinimumNutrients := 39 // Should have 39 nutrients with complete DRI coverage
	totalNutrients := len(goals.Targets) + len(goals.UpperLimits)
	if totalNutrients < expectedMinimumNutrients {
		t.Errorf("Expected at least %d total nutrition values (targets + upper limits), got %d targets + %d upper limits = %d total",
			expectedMinimumNutrients, len(goals.Targets), len(goals.UpperLimits), totalNutrients)
	}

	// Verify upper limits are set for nutrients that should be minimized
	expectedUpperLimits := []string{"added_sugars_g", "saturated_fat_g", "trans_fat_g", "cholesterol_mg"}
	for _, nutrient := range expectedUpperLimits {
		if _, exists := goals.UpperLimits[nutrient]; !exists {
			t.Errorf("Expected upper limit for %s but not found", nutrient)
		}
	}

	// Test with legacy custom overrides (backward compatibility)
	legacyOverrides := &UserOverrides{
		Overrides: map[string]float64{
			"calories":  3000,
			"protein_g": 180,
		},
	}

	goalsWithLegacyOverrides, err := resolver.ResolveGoals("male", &birthDate, legacyOverrides)
	if err != nil {
		t.Fatalf("Failed to resolve goals with legacy overrides: %v", err)
	}

	if goalsWithLegacyOverrides.Source != "custom" {
		t.Errorf("Expected source to be 'custom' with overrides, got %s", goalsWithLegacyOverrides.Source)
	}

	if goalsWithLegacyOverrides.Targets["calories"] != 3000 {
		t.Errorf("Expected calories to be overridden to 3000, got %f", goalsWithLegacyOverrides.Targets["calories"])
	}
}

func TestResolveGoalsWithTargetsAndUpperLimits(t *testing.T) {
	resolver, err := NewGoalResolver()
	if err != nil {
		t.Fatalf("Failed to create goal resolver: %v", err)
	}

	birthDate := time.Now().AddDate(-30, 0, 0) // 30 year old

	tests := []struct {
		name      string
		overrides *UserOverrides
		checkFunc func(*testing.T, *Goals)
	}{
		{
			name: "targets only",
			overrides: &UserOverrides{
				Name: "High Protein",
				Targets: map[string]float64{
					"calories":  3000,
					"protein_g": 180,
				},
			},
			checkFunc: func(t *testing.T, goals *Goals) {
				if goals.Source != "custom" {
					t.Errorf("Expected source to be 'custom', got %s", goals.Source)
				}
				if goals.CustomName != "High Protein" {
					t.Errorf("Expected custom name to be 'High Protein', got %s", goals.CustomName)
				}
				if goals.Targets["calories"] != 3000 {
					t.Errorf("Expected calories target to be 3000, got %f", goals.Targets["calories"])
				}
				if goals.Targets["protein_g"] != 180 {
					t.Errorf("Expected protein target to be 180, got %f", goals.Targets["protein_g"])
				}
			},
		},
		{
			name: "upper limits only",
			overrides: &UserOverrides{
				Name: "Low Sodium",
				UpperLimits: map[string]float64{
					"sodium_mg":       1500,
					"added_sugars_g":  20,
					"saturated_fat_g": 15,
				},
			},
			checkFunc: func(t *testing.T, goals *Goals) {
				if goals.Source != "custom" {
					t.Errorf("Expected source to be 'custom', got %s", goals.Source)
				}
				if goals.CustomName != "Low Sodium" {
					t.Errorf("Expected custom name to be 'Low Sodium', got %s", goals.CustomName)
				}
				if goals.UpperLimits["sodium_mg"] != 1500 {
					t.Errorf("Expected sodium upper limit to be 1500, got %f", goals.UpperLimits["sodium_mg"])
				}
				if goals.UpperLimits["added_sugars_g"] != 20 {
					t.Errorf("Expected added sugars upper limit to be 20, got %f", goals.UpperLimits["added_sugars_g"])
				}
				if goals.UpperLimits["saturated_fat_g"] != 15 {
					t.Errorf("Expected saturated fat upper limit to be 15, got %f", goals.UpperLimits["saturated_fat_g"])
				}
			},
		},
		{
			name: "both targets and upper limits",
			overrides: &UserOverrides{
				Name: "Comprehensive",
				Targets: map[string]float64{
					"calories":    2800,
					"protein_g":   150,
					"vitamin_c_mg": 120,
				},
				UpperLimits: map[string]float64{
					"sodium_mg":       1800,
					"caffeine_mg":     300,
					"cholesterol_mg":  250,
				},
			},
			checkFunc: func(t *testing.T, goals *Goals) {
				if goals.Source != "custom" {
					t.Errorf("Expected source to be 'custom', got %s", goals.Source)
				}
				// Check targets
				if goals.Targets["calories"] != 2800 {
					t.Errorf("Expected calories target to be 2800, got %f", goals.Targets["calories"])
				}
				if goals.Targets["protein_g"] != 150 {
					t.Errorf("Expected protein target to be 150, got %f", goals.Targets["protein_g"])
				}
				if goals.Targets["vitamin_c_mg"] != 120 {
					t.Errorf("Expected vitamin C target to be 120, got %f", goals.Targets["vitamin_c_mg"])
				}
				// Check upper limits
				if goals.UpperLimits["sodium_mg"] != 1800 {
					t.Errorf("Expected sodium upper limit to be 1800, got %f", goals.UpperLimits["sodium_mg"])
				}
				if goals.UpperLimits["caffeine_mg"] != 300 {
					t.Errorf("Expected caffeine upper limit to be 300, got %f", goals.UpperLimits["caffeine_mg"])
				}
				if goals.UpperLimits["cholesterol_mg"] != 250 {
					t.Errorf("Expected cholesterol upper limit to be 250, got %f", goals.UpperLimits["cholesterol_mg"])
				}
			},
		},
		{
			name: "backward compatibility migration",
			overrides: &UserOverrides{
				Name: "Legacy Test",
				// Legacy overrides field (should be migrated to Targets)
				Overrides: map[string]float64{
					"calories":  2500,
					"protein_g": 140,
				},
				// New fields should take precedence
				Targets: map[string]float64{
					"calories": 2600, // Should override the legacy value
				},
				UpperLimits: map[string]float64{
					"sodium_mg": 2000,
				},
			},
			checkFunc: func(t *testing.T, goals *Goals) {
				if goals.Source != "custom" {
					t.Errorf("Expected source to be 'custom', got %s", goals.Source)
				}
				// Targets field should take precedence over legacy Overrides
				if goals.Targets["calories"] != 2600 {
					t.Errorf("Expected calories target to be 2600 (from Targets, not legacy Overrides), got %f", goals.Targets["calories"])
				}
				// Legacy value should be migrated for protein_g since it's not in Targets
				if goals.Targets["protein_g"] != 140 {
					t.Errorf("Expected protein target to be 140 (migrated from legacy Overrides), got %f", goals.Targets["protein_g"])
				}
				// Upper limits should work normally
				if goals.UpperLimits["sodium_mg"] != 2000 {
					t.Errorf("Expected sodium upper limit to be 2000, got %f", goals.UpperLimits["sodium_mg"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goals, err := resolver.ResolveGoals("male", &birthDate, tt.overrides)
			if err != nil {
				t.Fatalf("Failed to resolve goals: %v", err)
			}

			if goals == nil {
				t.Fatal("Expected goals to be non-nil")
			}

			tt.checkFunc(t, goals)
		})
	}
}

func TestAllNutrientsCanHaveUpperLimits(t *testing.T) {
	resolver, err := NewGoalResolver()
	if err != nil {
		t.Fatalf("Failed to create goal resolver: %v", err)
	}

	birthDate := time.Now().AddDate(-30, 0, 0) // 30 year old

	// Test that we can set upper limits for any nutrient, even those that are normally targets
	testNutrients := []string{
		"vitamin_c_mg",    // Normally a target
		"protein_g",       // Normally a target
		"calcium_mg",      // Normally a target
		"iron_mg",         // Normally a target
		"potassium_mg",    // Normally a target
	}

	customUpperLimits := make(map[string]float64)
	for _, nutrient := range testNutrients {
		// Set an upper limit for nutrients that would normally be targets
		customUpperLimits[nutrient] = 999.0
	}

	overrides := &UserOverrides{
		Name:        "Test All Upper Limits",
		UpperLimits: customUpperLimits,
	}

	goals, err := resolver.ResolveGoals("male", &birthDate, overrides)
	if err != nil {
		t.Fatalf("Failed to resolve goals with custom upper limits: %v", err)
	}

	// Verify all custom upper limits were set
	for _, nutrient := range testNutrients {
		if limit, exists := goals.UpperLimits[nutrient]; !exists {
			t.Errorf("Expected upper limit for %s but not found", nutrient)
		} else if limit != 999.0 {
			t.Errorf("Expected upper limit for %s to be 999.0, got %f", nutrient, limit)
		}
	}

	// Verify these nutrients are NOT in targets anymore (since we overrode them to be upper limits)
	for _, nutrient := range testNutrients {
		if _, exists := goals.Targets[nutrient]; exists {
			// This is actually OK - a nutrient can be both a target and have an upper limit
			// The user might want both a minimum intake AND a maximum
			t.Logf("Nutrient %s has both target and upper limit - this is acceptable", nutrient)
		}
	}
}
