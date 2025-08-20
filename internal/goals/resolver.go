package goals

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//go:embed data/*.json
var driDataFiles embed.FS

// GoalResolver handles nutrition goal resolution based on DRI data and user preferences
type GoalResolver struct {
	driData *DRIData
	dvData  *DVData
}

// DRIData represents the structured RDA/AI data by sex and age
type DRIData struct {
	Metadata    DRIMetadata            `json:"metadata"`
	AgeBrackets []string               `json:"ageBrackets"`
	BySex       map[string]interface{} `json:"bySex"`
}

// DRIMetadata contains information about the DRI data source
type DRIMetadata struct {
	Description string `json:"description"`
	Sources     []struct {
		Title     string `json:"title"`
		Source    string `json:"source"`
		Reference string `json:"referenceId"`
	} `json:"authoritativeSources"`
}

// DVData represents FDA Daily Value reference data
type DVData struct {
	Metadata       DVMetadata                 `json:"metadata"`
	FDADailyValues map[string]DailyValueEntry `json:"fdaDailyValues"`
}

// DVMetadata contains DV data source information
type DVMetadata struct {
	Description string `json:"description"`
}

// DailyValueEntry represents a single daily value entry
type DailyValueEntry struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Note  string  `json:"note,omitempty"`
}

// Goals represents resolved nutrition goals for a user
type Goals struct {
	Targets     map[string]float64 `json:"targets"`      // nutrient_key -> target amount
	UpperLimits map[string]float64 `json:"upper_limits"` // nutrient_key -> upper limit
	Units       map[string]string  `json:"units"`        // nutrient_key -> unit
	Source      string             `json:"source"`       // "dri", "custom", etc.
	CustomName  string             `json:"custom_name,omitempty"` // name of custom goal (only when source is "custom")
	LifeStage   LifeStage          `json:"life_stage"`
}

// LifeStage represents the demographic profile used for goal resolution
type LifeStage struct {
	Sex        string `json:"sex"`         // "male", "female", "unspecified"
	AgeBracket string `json:"age_bracket"` // e.g., "19-30", "31-50"
}

// UserOverrides represents custom goal overrides from Pro users
type UserOverrides struct {
	Name      string             `json:"name,omitempty"` // custom name for the goal set
	Overrides map[string]float64 `json:"overrides"`      // nutrient_key -> custom target
}

// NewGoalResolver creates a new goal resolver with embedded DRI data
func NewGoalResolver() (*GoalResolver, error) {
	resolver := &GoalResolver{}

	// Load DRI data
	driContent, err := driDataFiles.ReadFile("data/rda_ai.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read DRI data: %w", err)
	}

	if err := json.Unmarshal(driContent, &resolver.driData); err != nil {
		return nil, fmt.Errorf("failed to parse DRI data: %w", err)
	}

	// Load DV data
	dvContent, err := driDataFiles.ReadFile("data/dv.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read DV data: %w", err)
	}

	if err := json.Unmarshal(dvContent, &resolver.dvData); err != nil {
		return nil, fmt.Errorf("failed to parse DV data: %w", err)
	}

	return resolver, nil
}

// ResolveGoals determines nutrition goals for a user based on their profile and any custom overrides
func (r *GoalResolver) ResolveGoals(sex string, birthDate *time.Time, customOverrides *UserOverrides) (*Goals, error) {
	// Determine age bracket
	ageBracket := r.getAgeBracket(birthDate)

	// Get base goals from DRI data
	baseGoals, err := r.getDRIGoals(sex, ageBracket)
	if err != nil {
		return nil, fmt.Errorf("failed to get DRI goals: %w", err)
	}

	// Apply custom overrides if provided
	if customOverrides != nil && len(customOverrides.Overrides) > 0 {
		for nutrient, value := range customOverrides.Overrides {
			baseGoals.Targets[nutrient] = value
		}
		baseGoals.Source = "custom"
		if customOverrides.Name != "" {
			baseGoals.CustomName = customOverrides.Name
		} else {
			baseGoals.CustomName = "Custom Goals"
		}
	}

	return baseGoals, nil
}

// getAgeBracket determines the appropriate DRI age bracket based on birth date
func (r *GoalResolver) getAgeBracket(birthDate *time.Time) string {
	if birthDate == nil {
		return "19-30 y" // default adult bracket
	}

	ageYears := time.Since(*birthDate).Hours() / 24 / 365.25

	switch {
	case ageYears < 1:
		if ageYears < 0.5 { // 6 months
			return "0-6 mo"
		}
		return "7-12 mo"
	case ageYears < 4:
		return "1-3 y"
	case ageYears < 9:
		return "4-8 y"
	case ageYears < 14:
		return "9-13 y"
	case ageYears < 19:
		return "14-18 y"
	case ageYears < 31:
		return "19-30 y"
	case ageYears < 51:
		return "31-50 y"
	case ageYears < 71:
		return "51-70 y"
	default:
		return "71+ y"
	}
}

// getDRIGoals extracts goals from DRI data for the given sex and age bracket
func (r *GoalResolver) getDRIGoals(sex, ageBracket string) (*Goals, error) {
	// Handle unspecified sex by defaulting to male
	if sex == "unspecified" || sex == "" {
		sex = "male"
	}

	sexData, ok := r.driData.BySex[sex]
	if !ok {
		return nil, fmt.Errorf("sex '%s' not found in DRI data", sex)
	}

	sexMap, ok := sexData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid sex data format")
	}

	bracketData, ok := sexMap[ageBracket]
	if !ok {
		return nil, fmt.Errorf("age bracket '%s' not found for sex '%s'", ageBracket, sex)
	}

	bracketMap, ok := bracketData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid age bracket data format")
	}

	goals := &Goals{
		Targets:     make(map[string]float64),
		UpperLimits: make(map[string]float64), // TODO: implement UL data
		Units:       make(map[string]string),
		Source:      "dri",
		LifeStage: LifeStage{
			Sex:        sex,
			AgeBracket: ageBracket,
		},
	}

	// Extract macronutrients
	if macros, ok := bracketMap["macronutrients"].(map[string]interface{}); ok {
		r.extractNutrientValues(macros, goals, "macro")
	}

	// Extract vitamins
	if vitamins, ok := bracketMap["vitamins"].(map[string]interface{}); ok {
		r.extractNutrientValues(vitamins, goals, "vitamin")
	}

	// Extract minerals
	if minerals, ok := bracketMap["minerals"].(map[string]interface{}); ok {
		r.extractNutrientValues(minerals, goals, "mineral")
	}

	return goals, nil
}

// extractNutrientValues extracts nutrient values from DRI data
func (r *GoalResolver) extractNutrientValues(data map[string]interface{}, goals *Goals, category string) {
	for nutrient, valueData := range data {
		if valueMap, ok := valueData.(map[string]interface{}); ok {
			if value, ok := valueMap["value"].(float64); ok && value > 0 {
				// Map DRI nutrient names to API nutrient keys
				apiKey := r.mapNutrientToAPIKey(nutrient)
				if apiKey != "" {
					goals.Targets[apiKey] = value

					// Extract unit
					if unit, ok := valueMap["unit"].(string); ok {
						goals.Units[apiKey] = r.normalizeUnit(unit)
					}
				}
			}
		}
	}
}

// mapNutrientToAPIKey maps DRI nutrient names to CompleteNutrient API field names
func (r *GoalResolver) mapNutrientToAPIKey(driName string) string {
	// Create mapping from DRI naming to API naming
	mapping := map[string]string{
		// Macronutrients
		"energy":       "calories",
		"carbohydrate": "total_carbs_g",
		"protein":      "protein_g",
		"fat":          "total_fat_g",
		"fiber":        "dietary_fiber_g",

		// Vitamins - map to API field names
		"vitamin_A":     "vitamin_a_mcg",
		"vitamin_C":     "vitamin_c_mg",
		"vitamin_D":     "vitamin_d_mcg",
		"vitamin_E":     "vitamin_e_mg",
		"vitamin_K":     "vitamin_k_mcg",
		"thiamin_B1":    "thiamine_mg",
		"riboflavin_B2": "riboflavin_mg",
		"niacin_B3":     "niacin_mg",
		"vitamin_B6":    "vitamin_b6_mg",
		"folate_B9":     "folate_mcg",
		"vitamin_B12":   "vitamin_b12_mcg",

		// Minerals - map to API field names
		"calcium":    "calcium_mg",
		"iron":       "iron_mg",
		"magnesium":  "magnesium_mg",
		"phosphorus": "phosphorus_mg",
		"potassium":  "potassium_mg",
		"zinc":       "zinc_mg",
		"copper":     "copper_mg",
		"manganese":  "manganese_mg",
		"selenium":   "selenium_mcg",
		"sodium":     "sodium_mg",
	}

	if apiKey, exists := mapping[driName]; exists {
		return apiKey
	}

	// Try direct mapping for cases where names already match
	if strings.Contains(driName, "_mg") || strings.Contains(driName, "_mcg") || strings.Contains(driName, "_g") {
		return driName
	}

	return "" // No mapping found
}

// normalizeUnit standardizes units to match API expectations
func (r *GoalResolver) normalizeUnit(unit string) string {
	// Remove extra text like "(400 IU)" and standardize
	unit = strings.Split(unit, " ")[0] // Take first part before space
	unit = strings.ToLower(unit)

	switch unit {
	case "µg", "mcg":
		return "mcg"
	case "mg":
		return "mg"
	case "g":
		return "g"
	case "kcal":
		return "kcal"
	default:
		return unit
	}
}
