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
	}

	// Always include request ID for easier debugging of user reports
	errorResp.TraceId = &requestID

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

// parseDateRangeParamsFromConsumptions converts GetConsumptionsParams to our internal DateRangeParams
func parseDateRangeParamsFromConsumptions(params api.GetConsumptionsParams) (*DateRangeParams, error) {
	// Check for ambiguous parameters
	hasDateRange := params.Start != nil || params.End != nil
	hasDays := params.Days != nil

	if hasDateRange && hasDays {
		return nil, NewAppError("Cannot specify both start/end dates and days parameter", http.StatusBadRequest, nil)
	}

	result := &DateRangeParams{}

	if hasDateRange {
		// Direct date range provided
		if params.Start != nil && params.End != nil {
			result.StartTime = *params.Start
			result.EndTime = *params.End

			// Calculate days for subscription validation
			diff := result.EndTime.Sub(result.StartTime)
			result.Days = int(diff.Hours()/24) + 1

			if result.Days <= 0 {
				return nil, NewAppError("Invalid date range: end date must be after start date", http.StatusBadRequest, nil)
			}
		} else {
			return nil, NewAppError("Both start and end dates must be provided when using date range", http.StatusBadRequest, nil)
		}
	} else if hasDays {
		// Days parameter provided
		result.Days = *params.Days
		if result.Days <= 0 {
			return nil, NewAppError("Days parameter must be greater than 0", http.StatusBadRequest, nil)
		}

		// Calculate date range using server UTC time
		result.EndTime = time.Now().UTC()
		result.StartTime = result.EndTime.AddDate(0, 0, -result.Days+1)

		// Set times to beginning/end of day for proper date range queries
		result.StartTime = time.Date(result.StartTime.Year(), result.StartTime.Month(), result.StartTime.Day(), 0, 0, 0, 0, time.UTC)
		result.EndTime = time.Date(result.EndTime.Year(), result.EndTime.Month(), result.EndTime.Day(), 23, 59, 59, 999999999, time.UTC)
	} else {
		// Default to 7 days (today view in most cases)
		result.Days = DefaultDays
		result.EndTime = time.Now().UTC()
		result.StartTime = result.EndTime.AddDate(0, 0, -result.Days+1)

		// Set times to beginning/end of day for proper date range queries
		result.StartTime = time.Date(result.StartTime.Year(), result.StartTime.Month(), result.StartTime.Day(), 0, 0, 0, 0, time.UTC)
		result.EndTime = time.Date(result.EndTime.Year(), result.EndTime.Month(), result.EndTime.Day(), 23, 59, 59, 999999999, time.UTC)
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

// normalizeConsumptionInput extracts and normalizes input from either JSON or multipart form
func normalizeConsumptionInput(c *gin.Context, requestID string, maxFormSize int64) (*ConsumptionInput, error) {
	contentType := c.GetHeader("Content-Type")

	LogDebug("Processing consumption input", "content_type", contentType, "request_id", requestID)

	// Handle JSON input (application/json)
	if strings.HasPrefix(contentType, "application/json") {
		var jsonInput struct {
			Text string `json:"text" binding:"required"`
		}

		if err := c.ShouldBindJSON(&jsonInput); err != nil {
			return nil, NewAppError("Invalid JSON or missing text field", http.StatusBadRequest, err)
		}

		// Sanitize the text input for security
		sanitizedText := sanitizeTranscriptOutput(jsonInput.Text)
		if len(strings.TrimSpace(sanitizedText)) == 0 {
			return nil, NewAppError("Text field cannot be empty", http.StatusBadRequest, nil)
		}

		LogDebug("JSON text input received", "text_length", len(sanitizedText), "request_id", requestID)

		return &ConsumptionInput{
			Text:      sanitizedText,
			Source:    "text",
			RequestID: requestID,
		}, nil
	}

	// Handle multipart form input (multipart/form-data) - existing behavior
	if err := c.Request.ParseMultipartForm(maxFormSize); err != nil {
		return nil, NewAppError("Invalid multipart form or file too large", http.StatusBadRequest, err)
	}

	// Check if text field is provided in multipart form (takes precedence over audio)
	if textValue := c.PostForm("text"); textValue != "" {
		sanitizedText := sanitizeTranscriptOutput(textValue)
		if len(strings.TrimSpace(sanitizedText)) == 0 {
			return nil, NewAppError("Text field cannot be empty", http.StatusBadRequest, nil)
		}

		LogDebug("Multipart text input received", "text_length", len(sanitizedText), "request_id", requestID)

		return &ConsumptionInput{
			Text:      sanitizedText,
			Source:    "text",
			RequestID: requestID,
		}, nil
	}

	// Handle audio file input
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		return nil, NewAppError("No audio file uploaded (field: audio) or text provided", http.StatusBadRequest, err)
	}
	defer file.Close()

	LogDebug("Audio file received",
		"filename", header.Filename,
		"size", header.Size,
		"content_type", header.Header.Get("Content-Type"),
		"request_id", requestID,
	)

	// Save to temp file
	tmpPath, mimeType, err := saveTempFile(file, header)
	if err != nil {
		return nil, NewAppError("Failed to save upload", http.StatusInternalServerError, err)
	}
	// Note: We don't defer removeFile here because the caller needs to handle cleanup

	LogDebug("Temp file created", "path", tmpPath, "mime_type", mimeType, "request_id", requestID)

	// Create nutrition service for transcription
	nutritionService := NewNutritionService(nil) // No store needed for transcription

	// Transcribe audio
	ctx := c.Request.Context()
	transcript, err := nutritionService.TranscribeAudio(ctx, tmpPath, mimeType)

	// Clean up temp file
	removeFile(tmpPath)

	if err != nil {
		return nil, NewAppError("Transcription failed", http.StatusInternalServerError, err)
	}

	LogDebug("Transcription completed", "transcript_length", len(transcript), "request_id", requestID)

	return &ConsumptionInput{
		Text:      transcript,
		Source:    "audio",
		RequestID: requestID,
	}, nil
}
