package server

import (
	"math"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// BiometricsCalculations contains calculated metrics from biometric data
type BiometricsCalculations struct {
	AgeYears *int     `json:"age_years,omitempty"`
	BMR      *float64 `json:"bmr,omitempty"`
	TDEE     *float64 `json:"tdee,omitempty"`
	BMI      *float64 `json:"bmi,omitempty"`
}

// CalculateMetrics computes age, BMR, TDEE, and BMI from biometric data
func CalculateMetrics(biometrics *storage.UserBiometrics) *BiometricsCalculations {
	if biometrics == nil {
		return &BiometricsCalculations{}
	}

	calculations := &BiometricsCalculations{}

	// Calculate age if birth date is available
	if biometrics.BirthDate != nil {
		age := calculateAge(*biometrics.BirthDate)
		calculations.AgeYears = &age
	}

	// Calculate BMR if we have required data
	if calculations.AgeYears != nil && biometrics.HeightCm != nil && biometrics.WeightKg != nil {
		bmr := calculateBMR(*calculations.AgeYears, *biometrics.HeightCm, *biometrics.WeightKg, biometrics.Sex)
		calculations.BMR = &bmr

		// Calculate TDEE if BMR is available
		tdee := calculateTDEE(bmr, biometrics.ActivityLevel)
		calculations.TDEE = &tdee
	}

	// Calculate BMI if height and weight are available
	if biometrics.HeightCm != nil && biometrics.WeightKg != nil {
		bmi := calculateBMI(*biometrics.HeightCm, *biometrics.WeightKg)
		calculations.BMI = &bmi
	}

	return calculations
}

// calculateAge calculates age in years from birth date
func calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// Adjust if birthday hasn't occurred this year
	if now.Month() < birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}

// calculateBMR calculates Basal Metabolic Rate using Mifflin-St Jeor equation
func calculateBMR(age int, heightCm, weightKg float64, sex string) float64 {
	// Mifflin-St Jeor equation:
	// Men: BMR = 10 × weight(kg) + 6.25 × height(cm) - 5 × age(years) + 5
	// Women: BMR = 10 × weight(kg) + 6.25 × height(cm) - 5 × age(years) - 161

	bmr := 10*weightKg + 6.25*heightCm - 5*float64(age)

	if sex == "female" {
		bmr -= 161
	} else {
		// Default to male equation for male, other, prefer_not_to_say, or unspecified
		bmr += 5
	}

	return math.Round(bmr*100) / 100 // Round to 2 decimal places
}

// calculateTDEE calculates Total Daily Energy Expenditure from BMR and activity level
func calculateTDEE(bmr float64, activityLevel string) float64 {
	var multiplier float64

	switch activityLevel {
	case "sedentary":
		multiplier = 1.2
	case "lightly_active":
		multiplier = 1.375
	case "moderately_active":
		multiplier = 1.55
	case "very_active":
		multiplier = 1.725
	case "extra_active":
		multiplier = 1.9
	default:
		multiplier = 1.375 // Default to lightly active
	}

	tdee := bmr * multiplier
	return math.Round(tdee*100) / 100 // Round to 2 decimal places
}

// calculateBMI calculates Body Mass Index
func calculateBMI(heightCm, weightKg float64) float64 {
	heightM := heightCm / 100 // Convert cm to meters
	bmi := weightKg / (heightM * heightM)
	return math.Round(bmi*100) / 100 // Round to 2 decimal places
}
