package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/goals"
	"github.com/grantbirki/noot/internal/storage"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// APIServer implements the generated ServerInterface
type APIServer struct {
	store        storage.Store
	goalResolver *goals.GoalResolver
}

// NewAPIServer creates a new API server instance
func NewAPIServer(store storage.Store) (*APIServer, error) {
	goalResolver, err := goals.NewGoalResolver()
	if err != nil {
		return nil, err
	}

	return &APIServer{
		store:        store,
		goalResolver: goalResolver,
	}, nil
}

// GetHealth implements ServerInterface.GetHealth
func (s *APIServer) GetHealth(c *gin.Context) {
	response := api.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   getenv("VERSION", "dev"),
	}
	c.JSON(http.StatusOK, response)
}

// CreateConsumption implements ServerInterface.CreateConsumption
func (s *APIServer) CreateConsumption(c *gin.Context) {
	requestID := c.GetString("request_id")

	LogDebug("Processing consumption request", "request_id", requestID)

	// Accept up to ~100MB form size
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil {
		appErr := NewAppError("Invalid multipart form", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		appErr := NewAppError("No audio file uploaded (field: audio)", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
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
		appErr := NewAppError("Failed to save upload", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	defer removeFile(tmpPath)

	LogDebug("Temp file created", "path", tmpPath, "mime_type", mimeType, "request_id", requestID)

	ctx := c.Request.Context()

	// Create nutrition service
	nutritionService := NewNutritionService(s.store)

	// 1) Transcribe
	LogDebug("Starting transcription", "request_id", requestID)
	transcript, err := nutritionService.TranscribeAudio(ctx, tmpPath, mimeType)
	if err != nil {
		appErr := NewAppError("Transcription failed", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	LogDebug("Transcription completed", "transcript_length", len(transcript), "request_id", requestID)

	// 2) Parse items (phase 1: extract items without nutrition)
	LogDebug("Starting item parsing (items only)", "request_id", requestID)
	parsed, err := nutritionService.ParseItems(ctx, transcript)
	if err != nil {
		appErr := NewAppError("Item parsing failed", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	LogDebug("Item parsing completed", "item_count", len(parsed.Items), "request_id", requestID)

	// 3) Hydrate nutrition (phase 2: add nutrition data using cache + AI)
	LogDebug("Starting nutrition hydration", "request_id", requestID)
	var hydratedItems []Item
	if s.store != nil {
		hydratedItems, err = nutritionService.HydrateNutrition(ctx, parsed.Items)
		if err != nil {
			appErr := NewAppError("Nutrition hydration failed", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
	} else {
		// No store available - hydrate without cache (direct AI calls)
		hydratedItems, err = nutritionService.HydrateNutritionWithoutCache(ctx, parsed.Items)
		if err != nil {
			appErr := NewAppError("Nutrition hydration failed", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
	}
	LogDebug("Nutrition hydration completed", "hydrated_count", len(hydratedItems), "request_id", requestID)

	// 4) Convert to API types
	var apiItems []api.Item
	var itemsWithNutrition []api.ItemWithNutrition

	for _, it := range hydratedItems {
		// Convert internal Item to API Item
		apiItem := convertInternalItemToAPI(it)
		apiItems = append(apiItems, apiItem)

		// Create ItemWithNutrition
		iw := api.ItemWithNutrition{
			Item: apiItem,
		}
		if it.Nutrients == nil {
			note := "Nutrition data unavailable"
			iw.Note = &note
		}
		itemsWithNutrition = append(itemsWithNutrition, iw)
	}

	// 5) Summarize using internal types and convert to API
	summary := summarize(convertAPIItemsToInternal(itemsWithNutrition))
	apiSummary := convertInternalSummaryToAPI(summary)

	// 6) Save consumption to database if store is available
	var consumptionID string
	if s.store != nil {
		// For now, use the default user if no authentication
		// In the future, this would come from authentication middleware
		user, err := getDefaultUser(ctx, s.store)
		if err != nil {
			LogError("Failed to get user for consumption storage", err)
		} else if user != nil {
			// Convert back to internal types for storage
			internalItems := convertAPIItemsToInternal(itemsWithNutrition)
			consumption := itemWithNutritionToConsumption(user.ID, transcript, internalItems, summary)
			if err := s.store.CreateConsumption(ctx, consumption); err != nil {
				LogError("Failed to save consumption to database", err)
				// Don't fail the request if storage fails
			} else {
				consumptionID = consumption.ID
				LogInfo("Consumption saved to database", "consumption_id", consumption.ID, "user_id", user.ID, "request_id", requestID)
			}
		}
	}

	// Create the API response
	resp := api.ConsumptionResponse{
		Id:          consumptionID, // Include consumption ID for editing
		Transcript:  transcript,
		ParsedItems: apiItems,
		Items:       itemsWithNutrition,
		Summary:     apiSummary,
		RequestId:   requestID,
	}

	LogInfo("Consumption request completed successfully",
		"transcript_length", len(transcript),
		"items_count", len(itemsWithNutrition),
		"total_calories", apiSummary.Totals.Calories,
		"request_id", requestID,
	)

	c.JSON(http.StatusOK, resp)
}

// UpdateConsumption implements ServerInterface.UpdateConsumption
func (s *APIServer) UpdateConsumption(c *gin.Context, id string) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	LogInfo("Update consumption request received", "consumption_id", id, "request_id", requestID)

	if s.store == nil {
		appErr := NewAppError("Storage not available", http.StatusInternalServerError, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse the request body
	var updateReq api.UpdateConsumptionRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Get the existing consumption
	existingConsumption, err := s.store.GetConsumption(ctx, id)
	if err != nil {
		appErr := NewAppError("Failed to get consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if existingConsumption == nil {
		appErr := NewAppError("Consumption not found", http.StatusNotFound, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert API items to internal format and calculate new summary
	internalItems := convertAPIItemsToInternal(updateReq.Items)

	// Recalculate summary from updated items
	summary := summarize(internalItems)

	// Update the consumption record (keep original transcript, user_id, created_at)
	updatedConsumption := itemWithNutritionToConsumption(existingConsumption.UserID, existingConsumption.Transcript, internalItems, summary)
	updatedConsumption.ID = existingConsumption.ID
	updatedConsumption.CreatedAt = existingConsumption.CreatedAt

	if err := s.store.UpdateConsumption(ctx, updatedConsumption); err != nil {
		appErr := NewAppError("Failed to update consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert updated consumption back to API format for response
	apiItems := make([]api.Item, len(internalItems))
	for i, item := range internalItems {
		apiItems[i] = convertInternalItemToAPI(item.Item)
	}

	apiSummary := convertInternalSummaryToAPI(summary)

	// Create the API response
	resp := api.ConsumptionResponse{
		Id:          updatedConsumption.ID,
		Transcript:  updatedConsumption.Transcript,
		ParsedItems: apiItems,
		Items:       updateReq.Items,
		Summary:     apiSummary,
		RequestId:   requestID,
	}

	LogInfo("Consumption updated successfully", "consumption_id", id, "request_id", requestID)
	c.JSON(http.StatusOK, resp)
}

// DeleteConsumption implements ServerInterface.DeleteConsumption
func (s *APIServer) DeleteConsumption(c *gin.Context, id string) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	LogInfo("Delete consumption request received", "consumption_id", id, "request_id", requestID)

	if s.store == nil {
		appErr := NewAppError("Storage not available", http.StatusInternalServerError, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check if consumption exists before trying to delete
	existingConsumption, err := s.store.GetConsumption(ctx, id)
	if err != nil {
		appErr := NewAppError("Failed to get consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if existingConsumption == nil {
		appErr := NewAppError("Consumption not found", http.StatusNotFound, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Delete the consumption
	if err := s.store.DeleteConsumption(ctx, id); err != nil {
		appErr := NewAppError("Failed to delete consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Create success response
	resp := api.DeleteResponse{
		Message: "Consumption deleted successfully",
		Id:      id,
	}

	LogInfo("Consumption deleted successfully", "consumption_id", id, "request_id", requestID)
	c.JSON(http.StatusOK, resp)
}

// GetConsumptions implements ServerInterface.GetConsumptions
func (s *APIServer) GetConsumptions(c *gin.Context) {
	if s.store == nil {
		errorResp := api.ErrorResponse{
			Error:     "Storage not available",
			Code:      http.StatusInternalServerError,
			Timestamp: time.Now().UTC(),
		}
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	// For now, get consumptions for the default user
	user, err := getDefaultUser(c.Request.Context(), s.store)
	if err != nil {
		errorResp := api.ErrorResponse{
			Error:     "Failed to get user",
			Code:      http.StatusInternalServerError,
			Timestamp: time.Now().UTC(),
		}
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"consumptions": []interface{}{},
			"user":         nil,
		})
		return
	}

	// Get recent consumptions
	consumptions, err := s.store.GetConsumptionsByUser(c.Request.Context(), user.ID, MaxConsumptions, 0)
	if err != nil {
		errorResp := api.ErrorResponse{
			Error:     "Failed to get consumptions",
			Code:      http.StatusInternalServerError,
			Timestamp: time.Now().UTC(),
		}
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"consumptions": consumptions,
		"user":         user,
		"count":        len(consumptions),
	})
}

// GetNutritionSummary implements ServerInterface.GetNutritionSummary
func (s *APIServer) GetNutritionSummary(c *gin.Context, params api.GetNutritionSummaryParams) {
	if s.store == nil {
		errorResp := api.ErrorResponse{
			Error:     "Storage not available",
			Code:      http.StatusInternalServerError,
			Timestamp: time.Now().UTC(),
		}
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	requestID := c.GetString("request_id")

	// Get the default user
	user, err := getDefaultUser(c.Request.Context(), s.store)
	if err != nil {
		errorResp := api.ErrorResponse{
			Error:     "Failed to get user",
			Code:      http.StatusInternalServerError,
			Timestamp: time.Now().UTC(),
		}
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}
	if user == nil {
		errorResp := api.ErrorResponse{
			Error:     "User not found",
			Code:      http.StatusNotFound,
			Timestamp: time.Now().UTC(),
		}
		c.JSON(http.StatusNotFound, errorResp)
		return
	}

	// Parse date range parameters from the generated params
	dateParams, err := parseDateRangeParamsFromAPI(params)
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			errorResp := api.ErrorResponse{
				Error:     "Invalid date parameters",
				Code:      http.StatusBadRequest,
				Timestamp: time.Now().UTC(),
			}
			c.JSON(http.StatusBadRequest, errorResp)
		}
		return
	}

	// Validate subscription access
	if err := validateSubscriptionAccess(user, dateParams.Days); err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			errorResp := api.ErrorResponse{
				Error:     "Access denied",
				Code:      http.StatusForbidden,
				Timestamp: time.Now().UTC(),
			}
			c.JSON(http.StatusForbidden, errorResp)
		}
		return
	}

	// Apply performance limit
	dateParams.Days = applyDaysLimit(dateParams.Days)

	summary, err := s.store.GetNutritionSummary(c.Request.Context(), user.ID, dateParams.StartTime, dateParams.EndTime)
	if err != nil {
		appErr := NewAppError("Failed to get nutrition summary", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API response format
	response := api.NutritionSummaryResponse{
		Summary: convertInternalNutritionSummaryToAPI(summary),
		User:    convertInternalUserToAPI(user),
		Days:    dateParams.Days,
		DateRange: struct {
			End   *time.Time `json:"end,omitempty"`
			Start *time.Time `json:"start,omitempty"`
		}{
			Start: &dateParams.StartTime,
			End:   &dateParams.EndTime,
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleAppError handles application errors in Gin context
func (s *APIServer) handleAppError(c *gin.Context, appErr *AppError, requestID string) {
	handleAppErrorGin(c, appErr, requestID)
}

// Development-only handlers

// SwaggerUIHandler serves Swagger UI for API documentation
func (s *APIServer) SwaggerUIHandler(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Noot API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/api/v1/openapi.yaml',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// GetGoals implements ServerInterface.GetGoals
func (s *APIServer) GetGoals(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getDefaultUser(ctx, s.store)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Get user's custom goals if they're a Pro user
	var customOverrides *goals.UserOverrides
	if user.SubscriptionTier == storage.SubscriptionTierPro {
		userGoal, err := s.store.GetUserGoal(ctx, user.ID, "custom")
		if err != nil {
			LogError("Failed to get user goal", err, "user_id", user.ID)
		} else if userGoal != nil {
			customOverrides = &goals.UserOverrides{}
			if err := json.Unmarshal([]byte(userGoal.OverridesJSON), &customOverrides); err != nil {
				LogError("Failed to parse user goal overrides", err, "user_id", user.ID)
				customOverrides = nil
			}
		}
	}

	// Get user's biometrics for personalized goals
	userBiometrics, err := s.store.GetUserBiometrics(ctx, user.ID)
	if err != nil {
		LogError("Failed to get user biometrics", err, "user_id", user.ID)
		// Continue with default values
	}

	// Extract sex and birth_date from biometrics, with defaults
	sex := "male" // Default fallback
	var birthDate *time.Time
	if userBiometrics != nil {
		if userBiometrics.Sex != "" && userBiometrics.Sex != "prefer_not_to_say" {
			sex = userBiometrics.Sex
		}
		birthDate = userBiometrics.BirthDate
	}

	// Resolve goals based on biometrics or defaults
	resolvedGoals, err := s.goalResolver.ResolveGoals(sex, birthDate, customOverrides)
	if err != nil {
		appErr := NewAppError("Failed to resolve nutrition goals", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API response format
	apiTargets := make(map[string]float32)
	for k, v := range resolvedGoals.Targets {
		apiTargets[k] = float32(v)
	}

	apiUpperLimits := make(map[string]float32)
	for k, v := range resolvedGoals.UpperLimits {
		apiUpperLimits[k] = float32(v)
	}

	apiGoals := api.Goals{
		Targets:     apiTargets,
		UpperLimits: apiUpperLimits,
		Units:       resolvedGoals.Units,
		Source:      api.GoalsSource(resolvedGoals.Source),
		LifeStage: api.LifeStage{
			Sex:        api.LifeStageSex(resolvedGoals.LifeStage.Sex),
			AgeBracket: resolvedGoals.LifeStage.AgeBracket,
		},
	}

	// Set custom name if available
	if resolvedGoals.CustomName != "" {
		apiGoals.CustomName = &resolvedGoals.CustomName
	}

	response := api.GoalsResponse{
		Goals: apiGoals,
		User:  convertUser(user),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateGoals implements ServerInterface.UpdateGoals
func (s *APIServer) UpdateGoals(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getDefaultUser(ctx, s.store)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check subscription tier
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		appErr := NewAppError("Pro subscription required for custom goals", http.StatusForbidden, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse request body
	var req api.UpdateGoalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Validate overrides (basic validation)
	if len(req.Overrides) == 0 {
		appErr := NewAppError("At least one override must be provided", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Validate and set goal name
	goalDisplayName := "Custom Goals" // default display name
	if req.Name != nil && *req.Name != "" {
		if len(*req.Name) > 50 {
			appErr := NewAppError("Goal name must be 50 characters or less", http.StatusBadRequest, nil)
			s.handleAppError(c, appErr, requestID)
			return
		}
		goalDisplayName = *req.Name
	}

	// Create the user overrides structure with both name and overrides
	userOverrides := goals.UserOverrides{
		Name:      goalDisplayName,
		Overrides: make(map[string]float64),
	}

	// Convert from float32 to float64
	for k, v := range req.Overrides {
		userOverrides.Overrides[k] = float64(v)
	}

	// Store the complete structure as JSON
	overridesJSON, err := json.Marshal(userOverrides)
	if err != nil {
		appErr := NewAppError("Failed to serialize goal overrides", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	userGoal := &storage.UserGoal{
		UserID:        user.ID,
		Name:          "custom", // Keep as "custom" for database constraint
		OverridesJSON: string(overridesJSON),
	}

	if err := s.store.UpsertUserGoal(ctx, userGoal); err != nil {
		appErr := NewAppError("Failed to save custom goals", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Return updated goals
	s.GetGoals(c)
}

// GetTrends implements ServerInterface.GetTrends
func (s *APIServer) GetTrends(c *gin.Context, params api.GetTrendsParams) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getDefaultUser(ctx, s.store)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse date range parameters
	start, end, days, err := parseTrendsDateRangeParams(params.Start, params.End, params.Days)
	if err != nil {
		appErr := NewAppError("Invalid date range parameters", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Apply subscription-based limits
	if err := validateTrendsSubscriptionAccess(user.SubscriptionTier, start, end); err != nil {
		appErr := NewAppError(err.Error(), http.StatusForbidden, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Get consumption data
	consumptions, err := s.store.GetConsumptionsByUserSince(ctx, user.ID, start)
	if err != nil {
		appErr := NewAppError("Failed to get consumptions", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse requested metrics (default to common macros)
	metrics := []string{"calories", "protein_g", "total_fat_g", "total_carbs_g"}
	if params.Metrics != nil && *params.Metrics != "" {
		// Parse comma-separated metrics
		// TODO: Implement proper parsing and validation
	}

	// Generate time series data
	series := generateTimeSeries(consumptions, metrics, start, end)

	response := api.TrendsResponse{
		Series: series,
		User:   convertUser(user),
		DateRange: struct {
			End   *time.Time `json:"end,omitempty"`
			Start *time.Time `json:"start,omitempty"`
		}{
			Start: &start,
			End:   &end,
		},
		Days: days,
	}

	c.JSON(http.StatusOK, response)
}

// ExportData implements ServerInterface.ExportData
func (s *APIServer) ExportData(c *gin.Context, params api.ExportDataParams) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getDefaultUser(ctx, s.store)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check subscription tier
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		appErr := NewAppError("Pro subscription required for data export", http.StatusForbidden, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse date range parameters
	start, end, _, err := parseTrendsDateRangeParams(params.Start, params.End, nil)
	if err != nil {
		appErr := NewAppError("Invalid date range parameters", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Get consumption data
	consumptions, err := s.store.GetConsumptionsByUserSince(ctx, user.ID, start)
	if err != nil {
		appErr := NewAppError("Failed to get consumptions", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse requested metrics (default to common macros)
	metrics := []string{"calories", "protein_g", "total_fat_g", "total_carbs_g"}
	if params.Metrics != nil && *params.Metrics != "" {
		// TODO: Implement proper parsing and validation
	}

	if params.Format == "csv" {
		// Generate CSV export
		csvData := generateCSVExport(consumptions, metrics, start, end)
		c.Header("Content-Disposition", "attachment; filename=\"nutrition-export.csv\"")
		c.Data(http.StatusOK, "text/csv", []byte(csvData))
	} else {
		// Generate JSON export (same as trends response)
		series := generateTimeSeries(consumptions, metrics, start, end)
		response := api.ExportResponse{
			Series: series,
			User:   convertUser(user),
			DateRange: struct {
				End   *time.Time `json:"end,omitempty"`
				Start *time.Time `json:"start,omitempty"`
			}{
				Start: &start,
				End:   &end,
			},
			Format: api.ExportResponseFormat(params.Format),
		}
		c.JSON(http.StatusOK, response)
	}
}

// GetUserBiometrics retrieves user biometrics data
func (s *APIServer) GetUserBiometrics(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getDefaultUser(ctx, s.store)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if user == nil {
		appErr := NewAppError("User not found", http.StatusNotFound, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Get user biometrics
	biometrics, err := s.store.GetUserBiometrics(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to get user biometrics", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Calculate metrics from biometrics
	calculations := CalculateMetrics(biometrics)

	// Convert to API response format
	var apiBiometrics *api.UserBiometrics
	if biometrics != nil {
		apiBiometrics = &api.UserBiometrics{}

		if biometrics.BirthDate != nil {
			apiDate := openapi_types.Date{Time: *biometrics.BirthDate}
			apiBiometrics.BirthDate = &apiDate
		}

		if biometrics.Sex != "" {
			switch biometrics.Sex {
			case "male":
				sex := api.Male
				apiBiometrics.Sex = &sex
			case "female":
				sex := api.Female
				apiBiometrics.Sex = &sex
			case "other":
				sex := api.Other
				apiBiometrics.Sex = &sex
			case "prefer_not_to_say":
				sex := api.PreferNotToSay
				apiBiometrics.Sex = &sex
			}
		}

		if biometrics.HeightCm != nil {
			heightFloat32 := float32(*biometrics.HeightCm)
			apiBiometrics.HeightCm = &heightFloat32
		}
		if biometrics.WeightKg != nil {
			weightFloat32 := float32(*biometrics.WeightKg)
			apiBiometrics.WeightKg = &weightFloat32
		}

		if biometrics.ActivityLevel != "" {
			switch biometrics.ActivityLevel {
			case "sedentary":
				level := api.UserBiometricsActivityLevelSedentary
				apiBiometrics.ActivityLevel = &level
			case "lightly_active":
				level := api.UserBiometricsActivityLevelLightlyActive
				apiBiometrics.ActivityLevel = &level
			case "moderately_active":
				level := api.UserBiometricsActivityLevelModeratelyActive
				apiBiometrics.ActivityLevel = &level
			case "very_active":
				level := api.UserBiometricsActivityLevelVeryActive
				apiBiometrics.ActivityLevel = &level
			case "extra_active":
				level := api.UserBiometricsActivityLevelExtraActive
				apiBiometrics.ActivityLevel = &level
			}
		}
	}

	// Build calculated metrics inline struct
	var calculatedMetrics *struct {
		AgeYears *int     `json:"age_years,omitempty"`
		Bmi      *float32 `json:"bmi,omitempty"`
		Bmr      *float32 `json:"bmr,omitempty"`
		Tdee     *float32 `json:"tdee,omitempty"`
	}
	if calculations != nil {
		calculatedMetrics = &struct {
			AgeYears *int     `json:"age_years,omitempty"`
			Bmi      *float32 `json:"bmi,omitempty"`
			Bmr      *float32 `json:"bmr,omitempty"`
			Tdee     *float32 `json:"tdee,omitempty"`
		}{}

		calculatedMetrics.AgeYears = calculations.AgeYears
		if calculations.BMR != nil {
			bmrFloat32 := float32(*calculations.BMR)
			calculatedMetrics.Bmr = &bmrFloat32
		}
		if calculations.TDEE != nil {
			tdeeFloat32 := float32(*calculations.TDEE)
			calculatedMetrics.Tdee = &tdeeFloat32
		}
		if calculations.BMI != nil {
			bmiFloat32 := float32(*calculations.BMI)
			calculatedMetrics.Bmi = &bmiFloat32
		}
	}

	response := api.BiometricsResponse{
		Biometrics:        apiBiometrics,
		CalculatedMetrics: calculatedMetrics,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserBiometrics creates or updates user biometrics
func (s *APIServer) UpdateUserBiometrics(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getDefaultUser(ctx, s.store)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if user == nil {
		appErr := NewAppError("User not found", http.StatusNotFound, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse request body
	var req api.UpdateBiometricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert API request to storage biometrics
	biometrics := &storage.UserBiometrics{
		UserID: user.ID,
	}

	// Convert birth date
	if req.BirthDate != nil {
		biometrics.BirthDate = &req.BirthDate.Time
	}

	// Convert sex
	if req.Sex != nil {
		switch *req.Sex {
		case api.UpdateBiometricsRequestSexMale:
			biometrics.Sex = "male"
		case api.UpdateBiometricsRequestSexFemale:
			biometrics.Sex = "female"
		case api.UpdateBiometricsRequestSexOther:
			biometrics.Sex = "other"
		case api.UpdateBiometricsRequestSexPreferNotToSay:
			biometrics.Sex = "prefer_not_to_say"
		}
	} else {
		biometrics.Sex = "prefer_not_to_say" // Default
	}

	// Convert height and weight
	if req.HeightCm != nil {
		heightFloat64 := float64(*req.HeightCm)
		biometrics.HeightCm = &heightFloat64
	}
	if req.WeightKg != nil {
		weightFloat64 := float64(*req.WeightKg)
		biometrics.WeightKg = &weightFloat64
	}

	// Convert activity level
	if req.ActivityLevel != nil {
		switch *req.ActivityLevel {
		case api.UpdateBiometricsRequestActivityLevelSedentary:
			biometrics.ActivityLevel = "sedentary"
		case api.UpdateBiometricsRequestActivityLevelLightlyActive:
			biometrics.ActivityLevel = "lightly_active"
		case api.UpdateBiometricsRequestActivityLevelModeratelyActive:
			biometrics.ActivityLevel = "moderately_active"
		case api.UpdateBiometricsRequestActivityLevelVeryActive:
			biometrics.ActivityLevel = "very_active"
		case api.UpdateBiometricsRequestActivityLevelExtraActive:
			biometrics.ActivityLevel = "extra_active"
		}
	} else {
		biometrics.ActivityLevel = "lightly_active" // Default
	}

	// Upsert biometrics
	err = s.store.UpsertUserBiometrics(ctx, biometrics)
	if err != nil {
		appErr := NewAppError("Failed to save user biometrics", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Return the updated biometrics with calculations
	// Re-fetch to get the complete data with timestamps
	updatedBiometrics, err := s.store.GetUserBiometrics(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to retrieve updated biometrics", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	calculations := CalculateMetrics(updatedBiometrics)

	// Convert to API response (same logic as GetUserBiometrics)
	var apiBiometrics *api.UserBiometrics
	if updatedBiometrics != nil {
		apiBiometrics = &api.UserBiometrics{}

		if updatedBiometrics.BirthDate != nil {
			apiDate := openapi_types.Date{Time: *updatedBiometrics.BirthDate}
			apiBiometrics.BirthDate = &apiDate
		}

		if updatedBiometrics.Sex != "" {
			switch updatedBiometrics.Sex {
			case "male":
				sex := api.Male
				apiBiometrics.Sex = &sex
			case "female":
				sex := api.Female
				apiBiometrics.Sex = &sex
			case "other":
				sex := api.Other
				apiBiometrics.Sex = &sex
			case "prefer_not_to_say":
				sex := api.PreferNotToSay
				apiBiometrics.Sex = &sex
			}
		}

		if updatedBiometrics.HeightCm != nil {
			heightFloat32 := float32(*updatedBiometrics.HeightCm)
			apiBiometrics.HeightCm = &heightFloat32
		}
		if updatedBiometrics.WeightKg != nil {
			weightFloat32 := float32(*updatedBiometrics.WeightKg)
			apiBiometrics.WeightKg = &weightFloat32
		}

		if updatedBiometrics.ActivityLevel != "" {
			switch updatedBiometrics.ActivityLevel {
			case "sedentary":
				level := api.UserBiometricsActivityLevelSedentary
				apiBiometrics.ActivityLevel = &level
			case "lightly_active":
				level := api.UserBiometricsActivityLevelLightlyActive
				apiBiometrics.ActivityLevel = &level
			case "moderately_active":
				level := api.UserBiometricsActivityLevelModeratelyActive
				apiBiometrics.ActivityLevel = &level
			case "very_active":
				level := api.UserBiometricsActivityLevelVeryActive
				apiBiometrics.ActivityLevel = &level
			case "extra_active":
				level := api.UserBiometricsActivityLevelExtraActive
				apiBiometrics.ActivityLevel = &level
			}
		}
	}

	var calculatedMetrics *struct {
		AgeYears *int     `json:"age_years,omitempty"`
		Bmi      *float32 `json:"bmi,omitempty"`
		Bmr      *float32 `json:"bmr,omitempty"`
		Tdee     *float32 `json:"tdee,omitempty"`
	}
	if calculations != nil {
		calculatedMetrics = &struct {
			AgeYears *int     `json:"age_years,omitempty"`
			Bmi      *float32 `json:"bmi,omitempty"`
			Bmr      *float32 `json:"bmr,omitempty"`
			Tdee     *float32 `json:"tdee,omitempty"`
		}{}

		calculatedMetrics.AgeYears = calculations.AgeYears
		if calculations.BMR != nil {
			bmrFloat32 := float32(*calculations.BMR)
			calculatedMetrics.Bmr = &bmrFloat32
		}
		if calculations.TDEE != nil {
			tdeeFloat32 := float32(*calculations.TDEE)
			calculatedMetrics.Tdee = &tdeeFloat32
		}
		if calculations.BMI != nil {
			bmiFloat32 := float32(*calculations.BMI)
			calculatedMetrics.Bmi = &bmiFloat32
		}
	}

	response := api.BiometricsResponse{
		Biometrics:        apiBiometrics,
		CalculatedMetrics: calculatedMetrics,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUserBiometrics deletes user biometrics data
func (s *APIServer) DeleteUserBiometrics(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getDefaultUser(ctx, s.store)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if user == nil {
		appErr := NewAppError("User not found", http.StatusNotFound, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Delete biometrics
	err = s.store.DeleteUserBiometrics(ctx, user.ID)
	if err != nil {
		if err.Error() == "user biometrics not found" {
			appErr := NewAppError("User biometrics not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
		appErr := NewAppError("Failed to delete user biometrics", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	response := api.DeleteResponse{
		Message: "User biometrics deleted successfully",
		Id:      user.ID,
	}

	c.JSON(http.StatusOK, response)
}

// OpenAPISpecHandler serves the OpenAPI specification
func (s *APIServer) OpenAPISpecHandler(c *gin.Context) {
	c.File("api/openapi.yaml")
}
