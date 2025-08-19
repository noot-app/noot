package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
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
