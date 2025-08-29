package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// handleAppErrorGin handles application errors in Gin context
func handleAppErrorGin(c *gin.Context, appErr *AppError, requestID string) {
	LogError(appErr.Message, appErr.Err,
		"request_id", requestID,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	// Use structured error response
	errorResp := api.ErrorResponse{
		Error:     appErr.Message,
		Code:      appErr.StatusCode,
		Timestamp: time.Now().UTC(),
	}

	// Add debug details in development mode
	if isDebugMode() || isDevMode() {
		if appErr.Err != nil {
			stack := []string{appErr.Err.Error()}
			errorResp.Stack = &stack
		}
		errorResp.TraceId = &requestID
	}

	c.JSON(appErr.StatusCode, errorResp)
}

// handleInternalServerError is a convenience wrapper for common internal server errors
func handleInternalServerError(c *gin.Context, message string, err error) {
	requestID := getRequestID(c)
	appErr := NewAppError(message, http.StatusInternalServerError, err)
	handleAppErrorGin(c, appErr, requestID)
}

// handleStorageUnavailableError handles the common "Storage not available" error case
func handleStorageUnavailableError(c *gin.Context) {
	handleInternalServerError(c, "Storage not available", nil)
}

// getRequestID retrieves the request ID from the Gin context
func getRequestID(c *gin.Context) string {
	return c.GetString("request_id")
}

// parseDateRangeParamsFromAPI converts generated API params to our internal DateRangeParams
func parseDateRangeParamsFromAPI(params api.GetNutritionSummaryParams) (*DateRangeParams, error) {
	// Convert the generated API params to our internal structure
	// This maintains compatibility with existing helper functions
	result := &DateRangeParams{}

	if params.Start != nil {
		result.StartTime = *params.Start
	} else {
		// Default to 7 days ago
		result.StartTime = time.Now().UTC().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	}

	if params.End != nil {
		result.EndTime = *params.End
	} else {
		// Default to now
		result.EndTime = time.Now().UTC()
	}

	// Handle days parameter if provided
	if params.Days != nil {
		result.Days = *params.Days
		// Override start time based on days
		result.StartTime = time.Now().UTC().AddDate(0, 0, -result.Days).Truncate(24 * time.Hour)
	} else {
		// Calculate days from start and end times
		diff := result.EndTime.Sub(result.StartTime)
		result.Days = int(diff.Hours()/24) + 1 // +1 to include partial days
	}

	if result.Days <= 0 {
		return nil, NewAppError("Invalid date range: end date must be after start date", http.StatusBadRequest, nil)
	}

	return result, nil
}

// convertUser converts a storage.User to api.User
func convertUser(user *storage.User) api.User {
	return api.User{
		Id:               user.ID,
		Handle:           user.Handle,
		FullName:         user.FullName,
		Email:            user.Email,
		SubscriptionTier: api.UserSubscriptionTier(user.SubscriptionTier),
		AvatarUrl:        user.AvatarURL,
		CreatedAt:        user.CreatedAt,
	}
}

// generateTimeSeries creates time series data from consumption records
func generateTimeSeries(consumptions []*storage.Consumption, metrics []string, start, end time.Time) map[string][]api.DataPoint {
	series := make(map[string][]api.DataPoint)

	// Initialize series for each metric
	for _, metric := range metrics {
		series[metric] = make([]api.DataPoint, 0)
	}

	// Group consumptions by date
	dailyTotals := make(map[string]map[string]float64)

	for _, consumption := range consumptions {
		dateKey := consumption.CreatedAt.Format("2006-01-02")

		if dailyTotals[dateKey] == nil {
			dailyTotals[dateKey] = make(map[string]float64)
		}

		// Add consumption values to daily totals
		for _, metric := range metrics {
			dailyTotals[dateKey][metric] += getConsumptionMetric(consumption, metric)
		}
	}

	// Generate data points for each day in the range
	current := start
	for current.Before(end) || current.Equal(end) {
		dateKey := current.Format("2006-01-02")

		for _, metric := range metrics {
			value := dailyTotals[dateKey][metric] // defaults to 0 if no data
			series[metric] = append(series[metric], api.DataPoint{
				Date:  current,
				Value: float32(value),
			})
		}

		current = current.Add(24 * time.Hour)
	}

	return series
}

// generateCSVExport creates CSV export data
func generateCSVExport(consumptions []*storage.Consumption, metrics []string, start, end time.Time) string {
	// Simple CSV generation - one row per date/metric combination
	csv := "date,metric,value\n"

	series := generateTimeSeries(consumptions, metrics, start, end)

	for metric, dataPoints := range series {
		for _, point := range dataPoints {
			csv += fmt.Sprintf("%s,%s,%.2f\n",
				point.Date.Format("2006-01-02"),
				metric,
				point.Value)
		}
	}

	return csv
}

// getConsumptionMetric extracts a specific metric value from a consumption record
func getConsumptionMetric(consumption *storage.Consumption, metric string) float64 {
	switch metric {
	case "calories":
		return consumption.TotalCalories
	case "protein_g":
		return consumption.TotalProtein
	case "total_fat_g":
		return consumption.TotalFat
	case "total_carbs_g":
		return consumption.TotalCarbs
	case "dietary_fiber_g":
		return consumption.DietaryFiber
	case "sodium_mg":
		return consumption.TotalSodium
	case "saturated_fat_g":
		return consumption.SaturatedFat
	case "trans_fat_g":
		return consumption.TransFat
	case "cholesterol_mg":
		return consumption.Cholesterol
	case "total_sugars_g":
		return consumption.TotalSugars
	case "added_sugars_g":
		return consumption.AddedSugars
	case "vitamin_a_mcg":
		return consumption.VitaminA
	case "vitamin_c_mg":
		return consumption.VitaminC
	case "vitamin_d_mcg":
		return consumption.VitaminD
	case "vitamin_e_mg":
		return consumption.VitaminE
	case "vitamin_k_mcg":
		return consumption.VitaminK
	case "thiamine_mg":
		return consumption.Thiamine
	case "riboflavin_mg":
		return consumption.Riboflavin
	case "niacin_mg":
		return consumption.Niacin
	case "vitamin_b6_mg":
		return consumption.VitaminB6
	case "folate_mcg":
		return consumption.Folate
	case "vitamin_b12_mcg":
		return consumption.VitaminB12
	case "calcium_mg":
		return consumption.Calcium
	case "iron_mg":
		return consumption.Iron
	case "magnesium_mg":
		return consumption.Magnesium
	case "phosphorus_mg":
		return consumption.Phosphorus
	case "potassium_mg":
		return consumption.Potassium
	case "zinc_mg":
		return consumption.Zinc
	case "copper_mg":
		return consumption.Copper
	case "manganese_mg":
		return consumption.Manganese
	case "selenium_mcg":
		return consumption.Selenium
	case "iodine_mcg":
		return consumption.Iodine
	case "molybdenum_mcg":
		return consumption.Molybdenum
	case "chromium_mcg":
		return consumption.Chromium
	case "fluoride_mg":
		return consumption.Fluoride
	case "chloride_mg":
		return consumption.Chloride
	case "biotin_mcg":
		return consumption.Biotin
	case "pantothenic_acid_mg":
		return consumption.PantothenicAcid
	case "choline_mg":
		return consumption.Choline
	case "monounsaturated_fat_g":
		return consumption.MonounsaturatedFat
	case "polyunsaturated_fat_g":
		return consumption.PolyunsaturatedFat
	case "omega3_ala_g":
		return consumption.Omega3Ala
	case "omega3_epa_g":
		return consumption.Omega3Epa
	case "omega3_dha_g":
		return consumption.Omega3Dha
	case "omega6_g":
		return consumption.Omega6
	case "alcohol_g":
		return consumption.Alcohol
	case "caffeine_mg":
		return consumption.Caffeine
	case "creatine_mg":
		return consumption.Creatine
	default:
		return 0
	}
}

// parseTrendsDateRangeParams converts trend API params to date range values
func parseTrendsDateRangeParams(start, end *time.Time, days *int) (time.Time, time.Time, int, error) {
	var startTime, endTime time.Time
	var numDays int

	if start != nil && end != nil {
		// Direct date range provided
		startTime = *start
		endTime = *end
		numDays = int(endTime.Sub(startTime).Hours()/24) + 1
	} else if days != nil {
		// Days parameter provided
		numDays = *days
		endTime = time.Now().UTC().Truncate(24 * time.Hour)
		startTime = endTime.AddDate(0, 0, -numDays)
	} else {
		// Default to 7 days
		numDays = 7
		endTime = time.Now().UTC().Truncate(24 * time.Hour)
		startTime = endTime.AddDate(0, 0, -numDays)
	}

	if startTime.After(endTime) {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("start date cannot be after end date")
	}

	return startTime, endTime, numDays, nil
}

// validateTrendsSubscriptionAccess checks if user has access based on subscription and date range
func validateTrendsSubscriptionAccess(subscriptionTier string, start, end time.Time) error {
	days := int(end.Sub(start).Hours()/24) + 1

	if days > 7 && strings.ToLower(subscriptionTier) != storage.SubscriptionTierPro {
		return fmt.Errorf("access to more than 7 days requires Pro subscription")
	}

	// Pro users can access up to 1 year of data
	if days > 365 {
		return fmt.Errorf("maximum date range is 365 days")
	}

	return nil
}
