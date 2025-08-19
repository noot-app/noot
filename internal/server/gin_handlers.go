package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// APIServer implements the generated ServerInterface
type APIServer struct {
	store storage.Store
}

// NewAPIServer creates a new API server instance
func NewAPIServer(store storage.Store) *APIServer {
	return &APIServer{
		store: store,
	}
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
				LogInfo("Consumption saved to database", "consumption_id", consumption.ID, "user_id", user.ID, "request_id", requestID)
			}
		}
	}

	// Create the API response
	resp := api.ConsumptionResponse{
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

// OpenAPISpecHandler serves the OpenAPI specification
func (s *APIServer) OpenAPISpecHandler(c *gin.Context) {
	c.File("api/openapi.yaml")
}
