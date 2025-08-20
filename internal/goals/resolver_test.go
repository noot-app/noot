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

	// Test with custom overrides
	overrides := &UserOverrides{
		Overrides: map[string]float64{
			"calories":  3000,
			"protein_g": 180,
		},
	}

	goalsWithOverrides, err := resolver.ResolveGoals("male", &birthDate, overrides)
	if err != nil {
		t.Fatalf("Failed to resolve goals with overrides: %v", err)
	}

	if goalsWithOverrides.Source != "custom" {
		t.Errorf("Expected source to be 'custom' with overrides, got %s", goalsWithOverrides.Source)
	}

	if goalsWithOverrides.Targets["calories"] != 3000 {
		t.Errorf("Expected calories to be overridden to 3000, got %f", goalsWithOverrides.Targets["calories"])
	}
}
