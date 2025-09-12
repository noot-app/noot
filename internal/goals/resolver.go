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
	Targets     map[string]float64 `json:"targets"`               // nutrient_key -> target amount
	UpperLimits map[string]float64 `json:"upper_limits"`          // nutrient_key -> upper limit
	Units       map[string]string  `json:"units"`                 // nutrient_key -> unit
	Source      string             `json:"source"`                // "dri", "custom", etc.
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
	Name              string             `json:"name,omitempty"`               // custom name for the goal set
	Targets           map[string]float64 `json:"targets,omitempty"`            // nutrient_key -> custom daily target
	UpperLimits       map[string]float64 `json:"upper_limits,omitempty"`       // nutrient_key -> custom upper limit
	DisabledNutrients []string           `json:"disabled_nutrients,omitempty"` // nutrient keys to hide from display and calculations
	// Legacy field for backward compatibility during transition
	Overrides map[string]float64 `json:"overrides,omitempty"` // deprecated: nutrient_key -> custom target
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
	if customOverrides != nil {
		hasOverrides := false

		// Apply target overrides
		if len(customOverrides.Targets) > 0 {
			for nutrient, value := range customOverrides.Targets {
				baseGoals.Targets[nutrient] = value
			}
			hasOverrides = true
		}

		// Apply upper limit overrides
		if len(customOverrides.UpperLimits) > 0 {
			for nutrient, value := range customOverrides.UpperLimits {
				baseGoals.UpperLimits[nutrient] = value
			}
			hasOverrides = true
		}

		// Handle legacy overrides field for backward compatibility
		if len(customOverrides.Overrides) > 0 {
			for nutrient, value := range customOverrides.Overrides {
				baseGoals.Targets[nutrient] = value
			}
			hasOverrides = true
		}

		// Only mark as custom if we actually have overrides
		if hasOverrides {
			baseGoals.Source = "custom"
			if customOverrides.Name != "" {
				baseGoals.CustomName = customOverrides.Name
			} else {
				baseGoals.CustomName = "Custom Goals"
			}
		}
	}

	// Filter out disabled nutrients if any are specified
	if customOverrides != nil && len(customOverrides.DisabledNutrients) > 0 {
		r.filterDisabledNutrients(baseGoals, customOverrides.DisabledNutrients)
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

	// Add nutrients from FDA Daily Values that aren't covered by DRI
	r.addDVNutrients(goals)

	return goals, nil
}

// addTargetIfMissing adds a target value for a nutrient if it doesn't already exist in targets or upper limits
func (r *GoalResolver) addTargetIfMissing(goals *Goals, nutrientKey string, value float64, unit string) {
	if _, existsInTargets := goals.Targets[nutrientKey]; !existsInTargets {
		if _, existsInUpperLimits := goals.UpperLimits[nutrientKey]; !existsInUpperLimits {
			goals.Targets[nutrientKey] = value
			goals.Units[nutrientKey] = unit
		}
	}
}

// addUpperLimitIfMissing adds an upper limit value for a nutrient if it doesn't already exist in targets or upper limits
func (r *GoalResolver) addUpperLimitIfMissing(goals *Goals, nutrientKey string, value float64, unit string) {
	if _, existsInTargets := goals.Targets[nutrientKey]; !existsInTargets {
		if _, existsInUpperLimits := goals.UpperLimits[nutrientKey]; !existsInUpperLimits {
			goals.UpperLimits[nutrientKey] = value
			goals.Units[nutrientKey] = unit
		}
	}
}

// addUpperLimitIfNotExists adds an upper limit value for a nutrient if it doesn't already exist as an upper limit
func (r *GoalResolver) addUpperLimitIfNotExists(goals *Goals, nutrientKey string, value float64, unit string) {
	if _, exists := goals.UpperLimits[nutrientKey]; !exists {
		goals.UpperLimits[nutrientKey] = value
		goals.Units[nutrientKey] = unit
	}
}

// extractNutrientValues extracts nutrient values from DRI data
func (r *GoalResolver) extractNutrientValues(data map[string]interface{}, goals *Goals, category string) {
	// Define nutrients that should default to upper limits (minimize intake)
	// Users can still override these with custom targets if desired
	defaultUpperLimitNutrients := map[string]bool{
		"added_sugars_g":  true,
		"saturated_fat_g": true,
		"trans_fat_g":     true,
		"cholesterol_mg":  true,
		"alcohol_g":       true,
		"caffeine_mg":     true,
	}

	for nutrient, valueData := range data {
		if valueMap, ok := valueData.(map[string]interface{}); ok {
			if value, ok := valueMap["value"].(float64); ok && value > 0 {
				// Map DRI nutrient names to API nutrient keys
				apiKey := r.mapNutrientToAPIKey(nutrient)
				if apiKey != "" {
					// Check if this should default to an upper limit or a target
					// Note: Users can override this preference with custom goals
					if defaultUpperLimitNutrients[apiKey] {
						goals.UpperLimits[apiKey] = value
					} else {
						goals.Targets[apiKey] = value
					}

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

		// Additional macronutrients (using FDA DV data)
		"saturated_fat": "saturated_fat_g",
		"trans_fat":     "trans_fat_g",
		"cholesterol":   "cholesterol_mg",
		"total_sugars":  "total_sugars_g",
		"added_sugars":  "added_sugars_g",

		// Vitamins - map to API field names
		"vitamin_A":           "vitamin_a_mcg",
		"vitamin_C":           "vitamin_c_mg",
		"vitamin_D":           "vitamin_d_mcg",
		"vitamin_E":           "vitamin_e_mg",
		"vitamin_K":           "vitamin_k_mcg",
		"thiamin_B1":          "thiamine_mg",
		"riboflavin_B2":       "riboflavin_mg",
		"niacin_B3":           "niacin_mg",
		"vitamin_B6":          "vitamin_b6_mg",
		"folate_B9":           "folate_mcg",
		"vitamin_B12":         "vitamin_b12_mcg",
		"biotin_B7":           "biotin_mcg",
		"pantothenic_acid_B5": "pantothenic_acid_mg",
		"choline":             "choline_mg",

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
		"chloride":   "chloride_mg",
		"chromium":   "chromium_mcg",
		"fluoride":   "fluoride_mg",
		"iodine":     "iodine_mcg",
		"molybdenum": "molybdenum_mcg",
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

// addDVNutrients adds nutrients from FDA Daily Values that aren't covered by DRI
func (r *GoalResolver) addDVNutrients(goals *Goals) {
	// Nutrients to add from DV data that typically aren't in DRI
	dvNutrients := map[string]string{
		"calories":      "calories", // Standard calorie target for adults
		"fat_total":     "total_fat_g",
		"saturated_fat": "saturated_fat_g",
		"trans_fat":     "trans_fat_g",
		"cholesterol":   "cholesterol_mg",
		"total_sugars":  "total_sugars_g",
		"added_sugars":  "added_sugars_g",
		"chloride":      "chloride_mg",
	}

	// Define nutrients that should default to upper limits (minimize intake)
	// Users can always override these with custom goals if they prefer different treatment
	defaultUpperLimitNutrients := map[string]bool{
		"added_sugars_g":  true,
		"saturated_fat_g": true,
		"trans_fat_g":     true,
		"cholesterol_mg":  true,
		"alcohol_g":       true,
		"caffeine_mg":     true,
	}

	for dvKey, apiKey := range dvNutrients {
		// Only add if not already present from DRI data (in either targets or upper limits)
		if _, existsInTargets := goals.Targets[apiKey]; !existsInTargets {
			if _, existsInUpperLimits := goals.UpperLimits[apiKey]; !existsInUpperLimits {
				if entry, exists := r.dvData.FDADailyValues[dvKey]; exists {
					// Check if this should default to an upper limit or a target
					// Note: Users can override this preference with custom goals
					if defaultUpperLimitNutrients[apiKey] {
						goals.UpperLimits[apiKey] = entry.Value
					} else {
						goals.Targets[apiKey] = entry.Value
					}
					goals.Units[apiKey] = r.normalizeUnit(entry.Unit)
				}
			}
		}
	}

	// Special handling for calories - use a standard 2000 kcal for adults if not present
	if _, exists := goals.Targets["calories"]; !exists {
		goals.Targets["calories"] = 2000
		goals.Units["calories"] = "kcal"
	}

	// Add nutrients that don't have FDA DV but should be tracked
	// These are nutrients in CompleteNutrient that need values for completeness
	r.addTargetIfMissing(goals, "total_sugars_g", 0, "g") // No specific recommendation, track for awareness

	// Add creatine with a reasonable target based on common supplementation recommendations
	// 3-5g per day is the typical maintenance dose for those who supplement
	r.addTargetIfMissing(goals, "creatine_mg", 3000, "mg") // 3g maintenance dose in mg

	// Add fat type targets that aren't in DRI but are important for tracking
	// Based on healthy fat distribution recommendations
	r.addTargetIfMissing(goals, "monounsaturated_fat_g", 27, "g") // ~10-15% of calories, using 2000 kcal = 22-33g, target middle at 27g
	r.addTargetIfMissing(goals, "polyunsaturated_fat_g", 17, "g") // ~5-10% of calories, using 2000 kcal = 11-22g, target middle at 17g

	// Add omega fatty acid targets based on DRI/health recommendations
	r.addTargetIfMissing(goals, "omega3_ala_g", 1.4, "g")  // DRI AI: 1.6g for men, 1.1g for women, using 1.4g as middle ground
	r.addTargetIfMissing(goals, "omega3_epa_g", 0.25, "g") // Health organizations recommend 250-500mg combined EPA+DHA, split evenly
	r.addTargetIfMissing(goals, "omega3_dha_g", 0.25, "g") // Health organizations recommend 250-500mg combined EPA+DHA, split evenly
	r.addTargetIfMissing(goals, "omega6_g", 11, "g")       // Balance with omega-3, typically 4:1 to 10:1 ratio, target ~11g for balance

	// Trans fat should be an upper limit with small buffer (minimize intake, but allow for LLM estimation errors)
	r.addUpperLimitIfMissing(goals, "trans_fat_g", 0.1, "g")

	// Add default upper limits for nutrients that should be minimized if not already set
	r.addUpperLimitIfNotExists(goals, "added_sugars_g", 50, "g")   // WHO/IOM recommendation: <10% of total calories, using 2000 kcal = 50g
	r.addUpperLimitIfNotExists(goals, "saturated_fat_g", 22, "g")  // AHA recommendation: <10% of total calories, using 2000 kcal = ~22g
	r.addUpperLimitIfNotExists(goals, "cholesterol_mg", 300, "mg") // AHA recommendation: <300mg per day

	// Add upper limits for functional compounds
	r.addUpperLimitIfNotExists(goals, "alcohol_g", 14, "g")     // Moderate drinking guidelines: up to 14g per day for women, 28g for men (using 14g as conservative limit)
	r.addUpperLimitIfNotExists(goals, "caffeine_mg", 400, "mg") // FDA guideline: up to 400mg per day for healthy adults
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

// filterDisabledNutrients removes disabled nutrients from goals
func (r *GoalResolver) filterDisabledNutrients(goals *Goals, disabledNutrients []string) {
	if len(disabledNutrients) == 0 {
		return
	}

	// Create a set for faster lookup
	disabled := make(map[string]bool)
	for _, nutrient := range disabledNutrients {
		disabled[nutrient] = true
	}

	// Remove disabled nutrients from targets
	for nutrient := range goals.Targets {
		if disabled[nutrient] {
			delete(goals.Targets, nutrient)
		}
	}

	// Remove disabled nutrients from upper limits
	for nutrient := range goals.UpperLimits {
		if disabled[nutrient] {
			delete(goals.UpperLimits, nutrient)
		}
	}

	// Remove disabled nutrients from units
	for nutrient := range goals.Units {
		if disabled[nutrient] {
			delete(goals.Units, nutrient)
		}
	}
}
