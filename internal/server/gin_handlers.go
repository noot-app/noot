package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

// CreateConsumption implements ServerInterface.CreateConsumption
//
// Security Notes:
// - Requires authentication via JWT middleware
// - AI inputs/outputs are validated and sanitized
// - File uploads limited to prevent abuse
// - User context enforced for data storage
func (s *APIServer) CreateConsumption(c *gin.Context) {
	requestID := c.GetString("request_id")
	startTime := time.Now()

	LogDebug("Processing consumption request", "request_id", requestID)

	// Parse size limits - reduced to 50MB for security
	maxFormSize := int64(50 << 20) // 50MB
	if maxBytesStr := getenv("MAX_UPLOAD_BYTES", ""); maxBytesStr != "" {
		if parsed, err := strconv.ParseInt(maxBytesStr, 10, 64); err == nil && parsed > 0 {
			// Cap at 100MB even if environment requests more
			if parsed > 100<<20 {
				maxFormSize = 100 << 20
			} else {
				maxFormSize = parsed
			}
		}
	}

	// 1) Normalize input (either audio transcription or text)
	inputStart := time.Now()
	input, err := normalizeConsumptionInput(c, requestID, maxFormSize)
	inputDuration := time.Since(inputStart).Milliseconds()
	LogInfo("metric_timing", "component", "normalize_input", "duration_ms", inputDuration, "request_id", requestID)

	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			appErr := NewAppError("Failed to process input", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
		}
		return
	}

	ctx := c.Request.Context()

	LogDebug("Input normalized", "source", input.Source, "request_id", requestID)

	// Handle consumption duplication (skip AI/LLM processing)
	if input.Source == "duplicate" && input.ConsumptionID != "" {
		if s.store == nil {
			appErr := NewAppError("Storage not available for consumption duplication", http.StatusInternalServerError, nil)
			s.handleAppError(c, appErr, requestID)
			return
		}

		// Get authenticated user for authorization
		user, err := getCurrentUser(c)
		if err != nil {
			if appErr, ok := err.(*AppError); ok {
				s.handleAppError(c, appErr, requestID)
			} else {
				appErr := NewAppError("Failed to get user", http.StatusInternalServerError, err)
				s.handleAppError(c, appErr, requestID)
			}
			return
		}

		LogDebug("Processing consumption duplication", "consumption_id", input.ConsumptionID, "user_id", user.ID, "request_id", requestID)

		// Fetch the existing consumption with ownership check
		existingConsumption, err := s.store.GetConsumptionForUser(ctx, user.ID, input.ConsumptionID)
		if err != nil {
			appErr := NewAppError("Failed to get consumption for duplication", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
		if existingConsumption == nil {
			appErr := NewAppError("Consumption not found or access denied", http.StatusNotFound, nil)
			s.handleAppError(c, appErr, requestID)
			return
		}

		// Convert the existing consumption to API format to get all items and nutrition data
		existingAPIConsumption, err := storageConsumptionToAPI(ctx, s.store, existingConsumption)
		if err != nil {
			appErr := NewAppError("Failed to convert existing consumption", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}

		// Create a new consumption using the existing data
		// Calculate summary from the existing items
		internalItems := convertAPIItemsToInternal(existingAPIConsumption.Items)
		summary := summarize(internalItems)

		// Determine the source based on input and context
		source := determineConsumptionSource(c, input)

		// Create new consumption record (same transcript, but new timestamp)
		newConsumption := itemWithNutritionToConsumption(user.ID, existingConsumption.Transcript, summary, source)

		// Copy the title and note from the original consumption
		newConsumption.Title = existingConsumption.Title
		newConsumption.Note = existingConsumption.Note

		createdConsumption, err := s.store.CreateConsumption(ctx, newConsumption)
		if err != nil {
			appErr := NewAppError("Failed to save duplicated consumption", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}

		LogInfo("Consumption duplicated successfully", "original_consumption_id", input.ConsumptionID, "new_consumption_id", createdConsumption.ID, "user_id", user.ID, "request_id", requestID)

		// Save individual consumption items by duplicating the existing ones
		for _, itemWithNutrition := range existingAPIConsumption.Items {
			// Try to find existing item in global cache for linking (optional)
			var itemID *string
			// Use the same normalization as the original creation logic
			normalizedName := normalizeItemNameForCache(itemWithNutrition.Item.Name, itemWithNutrition.Item.Brand)
			normalizedBrand := normalizeItemName(getBrandOrEmpty(itemWithNutrition.Item.Brand))

			// Create exact serving key to match how items are stored
			exactKey := fmt.Sprintf("%s|%s|%.1fg", normalizedName, normalizedBrand, itemWithNutrition.Item.Grams)
			if existingItem, err := s.store.GetItemByName(ctx, exactKey, normalizedBrand); err == nil && existingItem != nil {
				itemID = &existingItem.ID
			}

			consumptionItem := apiItemWithNutritionToConsumptionItem(createdConsumption.ID, itemWithNutrition, itemID)
			if err := s.store.CreateConsumptionItem(ctx, consumptionItem); err != nil {
				LogError("Failed to save duplicated consumption item", err, "item_name", itemWithNutrition.Item.Name, "consumption_id", createdConsumption.ID)
				// Continue with other items even if one fails
			}
		}

		// Copy labels from the original consumption to the new one
		if len(existingConsumption.Labels) > 0 {
			labelIDs := make([]string, len(existingConsumption.Labels))
			for i, label := range existingConsumption.Labels {
				labelIDs[i] = label.ID
			}

			if err := s.store.AssignConsumptionLabels(ctx, user.ID, createdConsumption.ID, labelIDs); err != nil {
				LogError("Failed to copy labels to duplicated consumption", err, "original_consumption_id", input.ConsumptionID, "new_consumption_id", createdConsumption.ID, "label_count", len(labelIDs))
				// Continue even if label copying fails - the consumption itself was created successfully
			} else {
				LogDebug("Successfully copied labels to duplicated consumption", "label_count", len(labelIDs), "new_consumption_id", createdConsumption.ID)
			}
		}

		// Convert to API format for response using the created consumption directly
		apiConsumption, err := storageConsumptionToAPI(ctx, s.store, createdConsumption)
		if err != nil {
			appErr := NewAppError("Failed to convert duplicated consumption", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}

		LogInfo("Consumption duplication completed successfully",
			"original_consumption_id", input.ConsumptionID,
			"new_consumption_id", createdConsumption.ID,
			"items_count", len(existingAPIConsumption.Items),
			"total_calories", summary.Totals.Calories,
			"request_id", requestID,
		)

		c.JSON(http.StatusOK, apiConsumption)
		return
	}

	// Original AI/LLM processing flow for text/audio input
	transcript := input.Text

	LogDebug("Starting AI/LLM processing", "source", input.Source, "text_length", len(transcript), "request_id", requestID)

	// Create nutrition service
	nutritionService := NewNutritionService(s.store)

	// 2) Parse items (phase 1: extract items without nutrition)
	parseStart := time.Now()
	parsed, err := nutritionService.ParseItems(ctx, transcript)
	parseDuration := time.Since(parseStart).Milliseconds()
	LogInfo("metric_timing", "component", "parse_items", "duration_ms", parseDuration, "request_id", requestID)

	if err != nil {
		appErr := NewAppError("Item parsing failed", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	LogDebug("Item parsing completed", "item_count", len(parsed.Items), "request_id", requestID)

	// 3) Hydrate nutrition (phase 2: add nutrition data using cache + AI)
	hydrateStart := time.Now()
	var hydratedItems []Item
	if s.store != nil {
		hydratedItems, err = nutritionService.HydrateNutrition(ctx, parsed.Items, parsed.Transcript)
		if err != nil {
			appErr := NewAppError("Nutrition hydration failed", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
	} else {
		// No store available - hydrate without cache (direct AI calls)
		hydratedItems, err = nutritionService.HydrateNutritionWithoutCache(ctx, parsed.Items, parsed.Transcript)
		if err != nil {
			appErr := NewAppError("Nutrition hydration failed", http.StatusInternalServerError, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
	}
	hydrateDuration := time.Since(hydrateStart).Milliseconds()
	LogInfo("metric_timing", "component", "hydrate_nutrition", "duration_ms", hydrateDuration, "request_id", requestID)

	LogDebug("Nutrition hydration completed", "hydrated_count", len(hydratedItems), "request_id", requestID)

	// 4) Convert to API types
	var itemsWithNutrition []api.ItemWithNutrition

	for _, it := range hydratedItems {
		// Convert internal Item to API Item
		apiItem := convertInternalItemToAPI(it)

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
		user, err := getCurrentUser(c)
		if err != nil {
			LogError("Failed to get user for consumption storage", err)
		} else if user != nil {
			// Determine the source based on input and context
			source := determineConsumptionSource(c, input)

			dbStart := time.Now()
			consumption := itemWithNutritionToConsumption(user.ID, transcript, summary, source)
			createdConsumption, err := s.store.CreateConsumption(ctx, consumption)
			if err != nil {
				LogError("Failed to save consumption to database", err)
				// Don't fail the request if storage fails
			} else {
				LogInfo("Consumption saved to database", "consumption_id", createdConsumption.ID, "user_id", user.ID, "request_id", requestID)

				// Save individual consumption items for historic breakdown
				itemsStart := time.Now()
				for i, itemWithNutrition := range itemsWithNutrition {
					// Try to find existing item in global cache for linking (optional)
					var itemID *string

					// Use the corresponding hydrated item for canonical name lookup
					if i < len(hydratedItems) {
						hydratedItem := hydratedItems[i]
						canonicalName := normalizeCanonicalName(hydratedItem.CanonicalName)
						if canonicalName == "" {
							canonicalName = normalizeCanonicalName(hydratedItem.Name)
						}
						normalizedBrand := normalizeItemName(getBrandOrEmpty(hydratedItem.Brand))

						// Look up item using canonical name and brand
						if existingItem, err := s.store.GetItemByName(ctx, canonicalName, normalizedBrand); err == nil && existingItem != nil {
							itemID = &existingItem.ID
						}
					}

					consumptionItem := apiItemWithNutritionToConsumptionItem(createdConsumption.ID, itemWithNutrition, itemID)
					if err := s.store.CreateConsumptionItem(ctx, consumptionItem); err != nil {
						LogError("Failed to save consumption item", err, "item_name", itemWithNutrition.Item.Name, "consumption_id", createdConsumption.ID)
						// Continue with other items even if one fails
					}
				}
				itemsDuration := time.Since(itemsStart).Milliseconds()
				LogInfo("metric_timing", "component", "db_save_items", "duration_ms", itemsDuration, "request_id", requestID)

				dbDuration := time.Since(dbStart).Milliseconds()
				LogInfo("metric_timing", "component", "db_save_total", "duration_ms", dbDuration, "request_id", requestID)

				// Convert to API format for response using the created consumption directly
				apiConsumption, err := storageConsumptionToAPI(ctx, s.store, createdConsumption)
				if err != nil {
					appErr := NewAppError("Failed to convert created consumption", http.StatusInternalServerError, err)
					s.handleAppError(c, appErr, requestID)
					return
				}

				totalDuration := time.Since(startTime).Milliseconds()
				LogInfo("metric_timing", "component", "request_total", "duration_ms", totalDuration, "request_id", requestID)

				LogInfo("Consumption request completed successfully",
					"transcript_length", len(transcript),
					"items_count", len(itemsWithNutrition),
					"total_calories", apiSummary.Totals.Calories,
					"request_id", requestID,
				)

				c.JSON(http.StatusOK, apiConsumption)
				return
			}
		}
	}
}

// GetConsumption implements ServerInterface.GetConsumption
func (s *APIServer) GetConsumption(c *gin.Context, id string) {
	if s.store == nil {
		handleStorageUnavailableError(c)
		return
	}

	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	var cons *storage.Consumption
	var err error

	// Try to get authenticated user (may or may not be present)
	user := getCurrentUserOptional(c)

	if user != nil {
		// User is authenticated - they can see their own private or public consumptions
		cons, err = s.store.GetConsumptionForUser(ctx, user.ID, id)
		if err != nil {
			handleInternalServerError(c, "Failed to get consumption", err)
			return
		}
	} else {
		// No authenticated user - can only see public consumptions
		cons, err = s.store.GetPublicConsumption(ctx, id)

		// If the consumption cannot be found, or it is not set to public, `GetPublicConsumption()` will return an error
		if err != nil {
			handleInternalServerError(c, "Failed to get public consumption", err)
			return
		}
	}

	if cons == nil {
		appErr := NewAppError("Consumption not found", http.StatusNotFound, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API including items
	apiConsumption, err := storageConsumptionToAPI(ctx, s.store, cons)
	if err != nil {
		handleInternalServerError(c, "Failed to convert consumption", err)
		return
	}

	// If not the owner (unauthenticated or different user), remove private information
	if user == nil || user.ID != cons.UserID {
		apiConsumption.Labels = nil
		apiConsumption.Note = nil

		// Also remove labels from items
		if apiConsumption.Items != nil {
			for i := range apiConsumption.Items {
				if apiConsumption.Items[i].Item.Labels != nil {
					apiConsumption.Items[i].Item.Labels = nil
				}
			}
		}
	}

	c.JSON(http.StatusOK, apiConsumption)
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

	// Get authenticated user for authorization
	user, err := getCurrentUser(c)
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			handleInternalServerError(c, "Failed to get user", err)
		}
		return
	}

	// Parse the request body
	var updateReq api.UpdateConsumptionRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Get the existing consumption with ownership check
	existingConsumption, err := s.store.GetConsumptionForUser(ctx, user.ID, id)
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

	// Update the consumption record (keep original transcript, user_id, and source)
	updatedConsumption := itemWithNutritionToConsumption(existingConsumption.UserID, existingConsumption.Transcript, summary, existingConsumption.Source)
	updatedConsumption.ID = existingConsumption.ID

	// Handle timestamp update - use new timestamp if provided, otherwise keep existing created_at
	if updateReq.ConsumedAt != nil {
		updatedConsumption.ConsumedAt = *updateReq.ConsumedAt
	} else {
		updatedConsumption.CreatedAt = existingConsumption.CreatedAt
	}

	// Handle note update - use new note if provided, otherwise keep existing note
	if updateReq.Note != nil {
		updatedConsumption.Note = updateReq.Note
	} else {
		updatedConsumption.Note = existingConsumption.Note
	}

	// Handle title update - use new title if provided, otherwise keep existing title
	if updateReq.Title != nil {
		updatedConsumption.Title = updateReq.Title
	} else {
		updatedConsumption.Title = existingConsumption.Title
	}

	// Handle is_public update - use new value if provided, otherwise keep existing value
	if updateReq.IsPublic != nil {
		updatedConsumption.IsPublic = *updateReq.IsPublic
	} else {
		updatedConsumption.IsPublic = existingConsumption.IsPublic
	}

	if err := s.store.UpdateConsumption(ctx, updatedConsumption); err != nil {
		appErr := NewAppError("Failed to update consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Update consumption items - replace existing with new ones
	// First, delete all existing consumption items
	if err := s.store.DeleteConsumptionItemsByConsumption(ctx, updatedConsumption.ID); err != nil {
		LogError("Failed to delete existing consumption items", err, "consumption_id", updatedConsumption.ID)
		// Continue - this is not critical to fail the request
	}

	// Then, create new consumption items
	for _, itemWithNutrition := range updateReq.Items {
		// Try to find existing item in global cache for linking (optional)
		var itemID *string
		if s.store != nil {
			// Use the same normalization as the caching logic for consistent keys
			normalizedName := normalizeItemNameForCache(itemWithNutrition.Item.Name, itemWithNutrition.Item.Brand)
			normalizedBrand := normalizeItemName(getBrandOrEmpty(itemWithNutrition.Item.Brand))

			// Create exact serving key to match how items are stored
			exactKey := fmt.Sprintf("%s|%s|%.1fg", normalizedName, normalizedBrand, itemWithNutrition.Item.Grams)
			if existingItem, err := s.store.GetItemByName(ctx, exactKey, normalizedBrand); err == nil && existingItem != nil {
				itemID = &existingItem.ID
			}
		}

		consumptionItem := apiItemWithNutritionToConsumptionItem(updatedConsumption.ID, itemWithNutrition, itemID)
		if err := s.store.CreateConsumptionItem(ctx, consumptionItem); err != nil {
			LogError("Failed to create consumption item during update", err, "item_name", itemWithNutrition.Item.Name, "consumption_id", updatedConsumption.ID)
			// Continue with other items even if one fails
		}
	}

	// After updating the consumption and items, get the full updated consumption data
	// This ensures we have all the latest data including labels and proper timestamps
	updatedConsumptionWithItems, err := s.store.GetConsumptionForUser(ctx, user.ID, updatedConsumption.ID)
	if err != nil {
		appErr := NewAppError("Failed to retrieve updated consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API format for response
	apiConsumption, err := storageConsumptionToAPI(ctx, s.store, updatedConsumptionWithItems)
	if err != nil {
		appErr := NewAppError("Failed to convert updated consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	LogInfo("Consumption updated successfully", "consumption_id", id, "request_id", requestID)
	c.JSON(http.StatusOK, apiConsumption)
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

	// Get authenticated user for authorization
	user, err := getCurrentUser(c)
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			handleInternalServerError(c, "Failed to get user", err)
		}
		return
	}

	// Check if consumption exists and user owns it
	existingConsumption, err := s.store.GetConsumptionForUser(ctx, user.ID, id)
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
func (s *APIServer) GetConsumptions(c *gin.Context, params api.GetConsumptionsParams) {
	if s.store == nil {
		handleStorageUnavailableError(c)
		return
	}

	requestID := c.GetString("request_id")

	// For now, get consumptions for the default user
	user, err := getCurrentUser(c)
	if err != nil {
		handleInternalServerError(c, "Failed to get user", err)
		return
	}
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"consumptions": []interface{}{},
			"user":         nil,
		})
		return
	}

	// Parse date range parameters if provided
	var dateParams *DateRangeParams
	hasDateFiltering := params.Start != nil || params.End != nil || params.Days != nil

	if hasDateFiltering {
		dateParams, err = parseDateRangeParamsFromConsumptions(params)
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

		// Validate subscription access for date range
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
	}

	// Check for label filtering parameters from params
	var labelsParam string
	var matchParam string

	if params.Labels != nil {
		labelsParam = *params.Labels
	}
	if params.Match != nil {
		matchParam = string(*params.Match)
	}

	var consumptions []*storage.Consumption
	ctx := c.Request.Context()

	// Get pagination parameters
	limit := MaxConsumptions // default
	offset := 0              // default

	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	if labelsParam != "" {
		// Filter by labels
		labelNames := strings.Split(labelsParam, ",")
		// Trim whitespace from each label name
		for i, name := range labelNames {
			labelNames[i] = strings.TrimSpace(name)
		}

		matchAll := matchParam == "all"
		consumptions, err = s.store.GetConsumptionsByLabels(ctx, user.ID, labelNames, matchAll, limit, offset)
		if err != nil {
			handleInternalServerError(c, "Failed to get consumptions by labels", err)
			return
		}
	} else if hasDateFiltering {
		// Get consumptions filtered by date range
		consumptions, err = s.store.GetConsumptionsByUserDateRange(ctx, user.ID, dateParams.StartTime, dateParams.EndTime, limit, offset)
		if err != nil {
			handleInternalServerError(c, "Failed to get consumptions by date range", err)
			return
		}
	} else {
		// Get recent consumptions without filtering
		consumptions, err = s.store.GetConsumptionsByUser(ctx, user.ID, limit, offset)
		if err != nil {
			handleInternalServerError(c, "Failed to get consumptions", err)
			return
		}
	}

	// Convert storage consumptions to API format with consumption items
	apiConsumptions := make([]api.Consumption, len(consumptions))
	for i, consumption := range consumptions {
		apiConsumption, err := storageConsumptionToAPI(c.Request.Context(), s.store, consumption)
		if err != nil {
			LogError("Failed to convert consumption to API format", err, "consumption_id", consumption.ID)
			// Continue with other consumptions if one fails to convert
			continue
		}
		apiConsumptions[i] = *apiConsumption
	}

	response := gin.H{
		"consumptions": apiConsumptions,
		"user":         user,
		"count":        len(apiConsumptions),
	}

	// Include date range information in response if filtering was applied
	if hasDateFiltering {
		response["date_range"] = gin.H{
			"start": dateParams.StartTime,
			"end":   dateParams.EndTime,
			"days":  dateParams.Days,
		}
	}

	c.JSON(http.StatusOK, response)
}

// handleAppError handles application errors in Gin context
func (s *APIServer) handleAppError(c *gin.Context, appErr *AppError, requestID string) {
	handleAppErrorGin(c, appErr, requestID)
}

// GetGoals implements ServerInterface.GetGoals
func (s *APIServer) GetGoals(c *gin.Context, params api.GetGoalsParams) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Determine goal source behavior
	forceDRI := false
	if params.Source != nil && *params.Source == api.Dri {
		forceDRI = true
	}

	// Get user's custom goals if they're a Pro user and not forcing DRI
	var customOverrides *goals.UserOverrides
	if !forceDRI && user.SubscriptionTier == storage.SubscriptionTierPro {
		// Check if a specific goal was requested via query parameter
		goalName := ""
		if params.GoalName != nil {
			goalName = *params.GoalName
		}

		// If no specific goal requested, use the active goal
		if goalName == "" && user.ActiveGoalID != nil {
			// Get the active goal by ID to find its name
			activeGoal, err := s.store.GetUserGoalByID(ctx, user.ID, *user.ActiveGoalID)
			if err != nil {
				LogError("Failed to get active goal", err, "user_id", user.ID, "goal_id", *user.ActiveGoalID)
			} else if activeGoal != nil {
				goalName = activeGoal.Name
			}
		}

		// Only try to get goals if we have a goal name
		if goalName != "" {
			userGoal, err := s.store.GetUserGoal(ctx, user.ID, goalName)
			if err != nil {
				LogError("Failed to get user goal", err, "user_id", user.ID, "goal_name", goalName)
			} else if userGoal != nil {
				customOverrides = &goals.UserOverrides{}
				if err := json.Unmarshal([]byte(userGoal.OverridesJSON), &customOverrides); err != nil {
					LogError("Failed to parse user goal overrides", err, "user_id", user.ID, "goal_name", goalName)
					customOverrides = nil
				}
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

	// Set disabled nutrients if available and we have custom overrides
	if customOverrides != nil && len(customOverrides.DisabledNutrients) > 0 {
		apiGoals.DisabledNutrients = &customOverrides.DisabledNutrients
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
	user, err := getCurrentUser(c)
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

	// Validate that at least one override type is provided
	hasTargets := req.Targets != nil && len(*req.Targets) > 0
	hasUpperLimits := req.UpperLimits != nil && len(*req.UpperLimits) > 0
	hasLegacyOverrides := req.Overrides != nil && len(*req.Overrides) > 0
	hasDisabledNutrients := req.DisabledNutrients != nil && len(*req.DisabledNutrients) > 0

	if !hasTargets && !hasUpperLimits && !hasLegacyOverrides && !hasDisabledNutrients {
		appErr := NewAppError("At least one target, upper limit, override, or disabled nutrient must be provided", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Validate goal name is required
	if req.Name == "" {
		appErr := NewAppError("Goal name is required", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	if len(req.Name) > 50 {
		appErr := NewAppError("Goal name must be 50 characters or less", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Create the user overrides structure with support for both targets and upper limits
	userOverrides := goals.UserOverrides{
		Name: req.Name,
	}

	// Handle targets
	if req.Targets != nil && len(*req.Targets) > 0 {
		userOverrides.Targets = make(map[string]float64)
		for k, v := range *req.Targets {
			userOverrides.Targets[k] = float64(v)
		}
	}

	// Handle upper limits
	if req.UpperLimits != nil && len(*req.UpperLimits) > 0 {
		userOverrides.UpperLimits = make(map[string]float64)
		for k, v := range *req.UpperLimits {
			userOverrides.UpperLimits[k] = float64(v)
		}
	}

	// Handle legacy overrides for backward compatibility
	if req.Overrides != nil && len(*req.Overrides) > 0 {
		if userOverrides.Targets == nil {
			userOverrides.Targets = make(map[string]float64)
		}
		// Legacy overrides are treated as targets
		for k, v := range *req.Overrides {
			userOverrides.Targets[k] = float64(v)
		}
	}

	// Handle disabled nutrients
	if req.DisabledNutrients != nil {
		userOverrides.DisabledNutrients = *req.DisabledNutrients
	}

	// Store the complete structure as JSON
	overridesJSON, err := json.Marshal(userOverrides)
	if err != nil {
		appErr := NewAppError("Failed to serialize goal overrides", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Set category with default value
	category := "custom"
	if req.Category != nil {
		category = string(*req.Category)
	}

	userGoal := &storage.UserGoal{
		UserID:        user.ID,
		Name:          req.Name, // Use the actual goal name from request
		Category:      category,
		OverridesJSON: string(overridesJSON),
	}

	if err := s.store.UpsertUserGoal(ctx, userGoal); err != nil {
		appErr := NewAppError("Failed to save custom goals", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Set this as the active goal if the user doesn't have one set yet
	if user.ActiveGoalID == nil {
		if err := s.store.SetActiveGoal(ctx, user.ID, userGoal.ID); err != nil {
			LogError("Failed to set active goal for new user", err, "user_id", user.ID, "goal_id", userGoal.ID)
		}
	}

	// Return updated goals
	s.GetGoals(c, api.GetGoalsParams{})
}

// GetUserBiometrics retrieves user biometrics data
func (s *APIServer) GetUserBiometrics(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getCurrentUser(c)
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
	user, err := getCurrentUser(c)
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
	user, err := getCurrentUser(c)
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

// GetGoalSets implements ServerInterface.GetGoalSets
func (s *APIServer) GetGoalSets(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// For non-Pro users, return empty goal sets but still include user data
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		response := api.GoalSetsResponse{
			GoalSets:       []api.GoalSetSummary{},
			ActiveGoalName: "",
			User:           convertUser(user),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Get all goal sets for the user
	userGoals, err := s.store.GetUserGoals(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to get goal sets", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API format
	goalSets := make([]api.GoalSetSummary, len(userGoals))
	for i, goal := range userGoals {
		goalSets[i] = api.GoalSetSummary{
			Name:      goal.Name,
			Category:  api.GoalSetSummaryCategory(goal.Category),
			CreatedAt: goal.CreatedAt,
			UpdatedAt: goal.UpdatedAt,
		}
	}

	// Determine active goal name by looking up the goal ID
	activeGoalName := ""
	if user.ActiveGoalID != nil {
		// Find the goal with the active ID to get its name
		for _, goal := range userGoals {
			if goal.ID == *user.ActiveGoalID {
				activeGoalName = goal.Name
				break
			}
		}
	}

	response := api.GoalSetsResponse{
		GoalSets:       goalSets,
		ActiveGoalName: activeGoalName,
		User:           convertUser(user),
	}

	c.JSON(http.StatusOK, response)
}

// SetActiveGoalSet implements ServerInterface.SetActiveGoalSet
func (s *APIServer) SetActiveGoalSet(c *gin.Context) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check subscription tier
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		appErr := NewAppError("Pro subscription required for goal sets management", http.StatusForbidden, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse request body
	var req api.SetActiveGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Validate the goal exists and get its ID
	goal, err := s.store.GetUserGoal(ctx, user.ID, req.Name)
	if err != nil {
		appErr := NewAppError("Failed to verify goal exists", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if goal == nil {
		appErr := NewAppError("Goal set not found", http.StatusNotFound, fmt.Errorf("goal '%s' not found", req.Name))
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Set the active goal using the goal ID
	if err := s.store.SetActiveGoal(ctx, user.ID, goal.ID); err != nil {
		appErr := NewAppError("Failed to set active goal", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Return the updated goals (which will now use the new active goal)
	s.GetGoals(c, api.GetGoalsParams{})
}

// DeleteGoalSet implements ServerInterface.DeleteGoalSet
func (s *APIServer) DeleteGoalSet(c *gin.Context, name string) {
	requestID := c.GetString("request_id")
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check subscription tier
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		appErr := NewAppError("Pro subscription required for goal sets management", http.StatusForbidden, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check if we're deleting the active goal
	isActiveGoal := false
	if user.ActiveGoalID != nil {
		// Get the goal being deleted to compare its ID with the active goal ID
		goalToDelete, err := s.store.GetUserGoal(ctx, user.ID, name)
		if err == nil && goalToDelete != nil && goalToDelete.ID == *user.ActiveGoalID {
			isActiveGoal = true
		}
	}

	// Get all user goals to check if this is the last one
	allGoals, err := s.store.GetUserGoals(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to get user goals", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Delete the goal set
	if err := s.store.DeleteUserGoal(ctx, user.ID, name); err != nil {
		if err.Error() == "user goal not found" {
			appErr := NewAppError("Goal set not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
		appErr := NewAppError("Failed to delete goal set", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// If we deleted the active goal and it was the last goal, clear the active goal name
	// This allows the system to fall back to DRI defaults
	if isActiveGoal && len(allGoals) <= 1 {
		if err := s.store.ClearActiveGoal(ctx, user.ID); err != nil {
			LogError("Failed to clear active goal after deleting last goal", err, "user_id", user.ID, "goal_name", name)
			// Don't fail the request - the goal was deleted successfully
		}
	}

	c.Status(http.StatusNoContent)
}

// Label handlers

// GetConsumptionLabels retrieves labels assigned to a consumption
func (s *APIServer) GetConsumptionLabels(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	labels, err := s.store.ListConsumptionLabels(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to retrieve consumption labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiLabels := make([]api.Label, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"labels": apiLabels})
}

// AssignConsumptionLabels assigns labels to a consumption
func (s *APIServer) AssignConsumptionLabels(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.AssignLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.AssignConsumptionLabels(ctx, user.ID, id, req.Ids)
	if err != nil {
		if err.Error() == "consumption not found" || err.Error() == "access denied" {
			appErr := NewAppError("Consumption not found or access denied", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to assign labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Return current labels
	labels, err := s.store.ListConsumptionLabels(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to retrieve updated labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiLabels := make([]api.Label, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"labels": apiLabels})
}

// UnassignConsumptionLabel removes a label assignment from a consumption
func (s *APIServer) UnassignConsumptionLabel(c *gin.Context, id string, labelId string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.UnassignConsumptionLabel(ctx, user.ID, id, labelId)
	if err != nil {
		appErr := NewAppError("Assignment not found or access denied", http.StatusNotFound, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetConsumptionItemLabels retrieves labels assigned to a consumption item
func (s *APIServer) GetConsumptionItemLabels(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	labels, err := s.store.ListConsumptionItemLabels(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to retrieve consumption item labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiLabels := make([]api.Label, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"labels": apiLabels})
}

// AssignConsumptionItemLabels assigns labels to a consumption item
func (s *APIServer) AssignConsumptionItemLabels(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.AssignLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.AssignConsumptionItemLabels(ctx, user.ID, id, req.Ids)
	if err != nil {
		if err.Error() == "consumption item not found" || err.Error() == "access denied" {
			appErr := NewAppError("Consumption item not found or access denied", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to assign labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Return current labels
	labels, err := s.store.ListConsumptionItemLabels(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to retrieve updated labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiLabels := make([]api.Label, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"labels": apiLabels})
}

// UnassignConsumptionItemLabel removes a label assignment from a consumption item
func (s *APIServer) UnassignConsumptionItemLabel(c *gin.Context, id string, labelId string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.UnassignConsumptionItemLabel(ctx, user.ID, id, labelId)
	if err != nil {
		appErr := NewAppError("Assignment not found or access denied", http.StatusNotFound, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	c.Status(http.StatusNoContent)
}

// Event handlers

// GetEvents retrieves events for the current user with optional filtering
func (s *APIServer) GetEvents(c *gin.Context, params api.GetEventsParams) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Parse query parameters into EventListOptions
	options := storage.EventListOptions{
		Limit:  50, // Default limit
		Offset: 0,
	}

	if params.Limit != nil {
		options.Limit = *params.Limit
	}
	if params.Offset != nil {
		options.Offset = *params.Offset
	}
	if params.StartDate != nil {
		options.StartDate = params.StartDate
	}
	if params.EndDate != nil {
		options.EndDate = params.EndDate
	}
	if params.EventTypeId != nil {
		options.EventTypeID = params.EventTypeId
	}
	if params.LevelMin != nil {
		options.LevelMin = params.LevelMin
	}
	if params.LevelMax != nil {
		options.LevelMax = params.LevelMax
	}
	if params.Labels != nil {
		options.Labels = strings.Split(*params.Labels, ",")
	}
	if params.Match != nil && *params.Match == api.GetEventsParamsMatchAll {
		options.MatchAll = true
	}

	ctx := c.Request.Context()
	events, err := s.store.ListEvents(ctx, user.ID, options)
	if err != nil {
		appErr := NewAppError("Failed to retrieve events", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiEvents := make([]api.Event, len(events))
	for i, event := range events {
		apiEvents[i] = api.Event{
			Id:          event.ID,
			UserId:      event.UserID,
			Name:        event.Name,
			EventTypeId: event.EventTypeID,
			StartedAt:   event.StartedAt,
			EndedAt:     event.EndedAt,
			Level:       event.Level,
			Note:        event.Note,
			Color:       event.Color,
			CreatedAt:   event.CreatedAt,
			UpdatedAt:   event.UpdatedAt,
		}
	}

	response := api.EventsResponse{
		Events: apiEvents,
	}

	c.JSON(http.StatusOK, response)
}

// CreateEvent creates a new event for the current user
func (s *APIServer) CreateEvent(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.EventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Validate that event dates are not in the future
	now := time.Now().UTC()
	if req.StartedAt.After(now) {
		appErr := NewAppError("Event start time cannot be in the future", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Calculate end time if duration_minutes is provided
	var endedAt *time.Time
	if req.EndedAt != nil {
		if req.EndedAt.After(now) {
			appErr := NewAppError("Event end time cannot be in the future", http.StatusBadRequest, nil)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		endedAt = req.EndedAt
	} else if req.DurationMinutes != nil {
		endTime := req.StartedAt.Add(time.Duration(*req.DurationMinutes) * time.Minute)
		if endTime.After(now) {
			appErr := NewAppError("Event end time cannot be in the future", http.StatusBadRequest, nil)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		endedAt = &endTime
	}

	event := &storage.Event{
		UserID:      user.ID,
		Name:        req.Name,
		EventTypeID: req.EventTypeId,
		StartedAt:   req.StartedAt,
		EndedAt:     endedAt,
		Level:       req.Level,
		Note:        req.Note,
		Color:       req.Color,
	}

	ctx := c.Request.Context()
	err = s.store.CreateEvent(ctx, event)
	if err != nil {
		if strings.Contains(err.Error(), "event_limit_exceeded") {
			appErr := NewAppError("Event limit exceeded (1000 events per user)", http.StatusForbidden, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to create event", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	apiEvent := api.Event{
		Id:          event.ID,
		UserId:      event.UserID,
		Name:        event.Name,
		EventTypeId: event.EventTypeID,
		StartedAt:   event.StartedAt,
		EndedAt:     event.EndedAt,
		Level:       event.Level,
		Note:        event.Note,
		Color:       event.Color,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}

	c.JSON(http.StatusCreated, apiEvent)
}

// GetEvent retrieves a specific event by ID
func (s *APIServer) GetEvent(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	event, err := s.store.GetEvent(ctx, user.ID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			appErr := NewAppError("Event not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to retrieve event", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Load labels and links for the event
	labels, err := s.store.ListEventLabels(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to load event labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	links, err := s.store.ListEventLinks(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to load event links", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiLabels := make([]api.Label, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}

	apiLinks := make([]api.EventLink, len(links))
	for i, link := range links {
		apiLinks[i] = api.EventLink{
			Id:                link.ID,
			EventId:           link.EventID,
			ConsumptionId:     link.ConsumptionID,
			ConsumptionItemId: link.ConsumptionItemID,
			CreatedAt:         link.CreatedAt,
		}
	}

	response := api.EventWithDetails{
		Id:          event.ID,
		UserId:      event.UserID,
		Name:        event.Name,
		EventTypeId: event.EventTypeID,
		StartedAt:   event.StartedAt,
		EndedAt:     event.EndedAt,
		Level:       event.Level,
		Note:        event.Note,
		Color:       event.Color,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
		Labels:      apiLabels,
		Links:       apiLinks,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateEvent updates an existing event
func (s *APIServer) UpdateEvent(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.EventUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Get existing event to preserve fields not being updated
	ctx := c.Request.Context()
	event, err := s.store.GetEvent(ctx, user.ID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			appErr := NewAppError("Event not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to retrieve event", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Update fields if provided
	now := time.Now().UTC()

	if req.Name != nil {
		event.Name = *req.Name
	}
	if req.EventTypeId != nil {
		event.EventTypeID = req.EventTypeId
	}
	if req.StartedAt != nil {
		if req.StartedAt.After(now) {
			appErr := NewAppError("Event start time cannot be in the future", http.StatusBadRequest, nil)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		event.StartedAt = *req.StartedAt
	}
	if req.EndedAt != nil {
		if req.EndedAt.After(now) {
			appErr := NewAppError("Event end time cannot be in the future", http.StatusBadRequest, nil)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		event.EndedAt = req.EndedAt
	}
	if req.Level != nil {
		event.Level = req.Level
	}
	if req.Note != nil {
		event.Note = req.Note
	}
	if req.Color != nil {
		event.Color = req.Color
	}

	err = s.store.UpdateEvent(ctx, event)
	if err != nil {
		appErr := NewAppError("Failed to update event", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	apiEvent := api.Event{
		Id:          event.ID,
		UserId:      event.UserID,
		Name:        event.Name,
		EventTypeId: event.EventTypeID,
		StartedAt:   event.StartedAt,
		EndedAt:     event.EndedAt,
		Level:       event.Level,
		Note:        event.Note,
		Color:       event.Color,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}

	c.JSON(http.StatusOK, apiEvent)
}

// DeleteEvent deletes an event and all its associations
func (s *APIServer) DeleteEvent(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.DeleteEvent(ctx, user.ID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			appErr := NewAppError("Event not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to delete event", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetEventLabels retrieves all labels assigned to an event
func (s *APIServer) GetEventLabels(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	labels, err := s.store.ListEventLabels(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to retrieve event labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	apiLabels := make([]api.Label, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}

	response := struct {
		Labels []api.Label `json:"labels"`
	}{
		Labels: apiLabels,
	}

	c.JSON(http.StatusOK, response)
}

// AssignEventLabels assigns labels to an event
func (s *APIServer) AssignEventLabels(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.AssignLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.AssignEventLabels(ctx, user.ID, id, req.Ids)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			appErr := NewAppError("Event or label not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to assign labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Return updated labels
	labels, err := s.store.ListEventLabels(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to retrieve updated labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	apiLabels := make([]api.Label, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.Label{
			Id:          label.ID,
			Name:        label.Name,
			Description: label.Description,
			Color:       label.Color,
			CreatedAt:   label.CreatedAt,
			UpdatedAt:   label.UpdatedAt,
		}
	}

	response := struct {
		Labels []api.Label `json:"labels"`
	}{
		Labels: apiLabels,
	}

	c.JSON(http.StatusOK, response)
}

// UnassignEventLabel removes a label assignment from an event
func (s *APIServer) UnassignEventLabel(c *gin.Context, id string, labelId string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.UnassignEventLabel(ctx, user.ID, id, labelId)
	if err != nil {
		appErr := NewAppError("Assignment not found or access denied", http.StatusNotFound, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetEventLinks retrieves consumption/item links for an event
func (s *APIServer) GetEventLinks(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	links, err := s.store.ListEventLinks(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Failed to retrieve event links", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	apiLinks := make([]api.EventLink, len(links))
	for i, link := range links {
		apiLinks[i] = api.EventLink{
			Id:                link.ID,
			EventId:           link.EventID,
			ConsumptionId:     link.ConsumptionID,
			ConsumptionItemId: link.ConsumptionItemID,
			CreatedAt:         link.CreatedAt,
		}
	}

	response := struct {
		Links []api.EventLink `json:"links"`
	}{
		Links: apiLinks,
	}

	c.JSON(http.StatusOK, response)
}

// CreateEventLink creates a manual link between an event and consumption/item
func (s *APIServer) CreateEventLink(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.EventLinkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Validate exactly one target is specified
	if (req.ConsumptionId == nil && req.ConsumptionItemId == nil) || (req.ConsumptionId != nil && req.ConsumptionItemId != nil) {
		appErr := NewAppError("Exactly one of consumption_id or consumption_item_id must be specified", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	link := &storage.EventLink{
		EventID:           id,
		ConsumptionID:     req.ConsumptionId,
		ConsumptionItemID: req.ConsumptionItemId,
	}

	ctx := c.Request.Context()
	// Verify event ownership before creating link
	_, err = s.store.GetEvent(ctx, user.ID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			appErr := NewAppError("Event not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to verify event ownership", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	err = s.store.CreateEventLink(ctx, link)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			appErr := NewAppError("Link already exists", http.StatusConflict, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to create event link", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	apiLink := api.EventLink{
		Id:                link.ID,
		EventId:           link.EventID,
		ConsumptionId:     link.ConsumptionID,
		ConsumptionItemId: link.ConsumptionItemID,
		CreatedAt:         link.CreatedAt,
	}

	c.JSON(http.StatusCreated, apiLink)
}

// DeleteEventLink removes a link between an event and consumption/item
func (s *APIServer) DeleteEventLink(c *gin.Context, id string, linkId string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.DeleteEventLink(ctx, user.ID, linkId)
	if err != nil {
		appErr := NewAppError("Link not found or access denied", http.StatusNotFound, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	c.Status(http.StatusNoContent)
}

// Event Type handlers

// GetEventTypes retrieves all event types for the current user
func (s *APIServer) GetEventTypes(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	eventTypes, err := s.store.ListEventTypes(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to retrieve event types", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Get event counts for each event type
	eventCounts, err := s.store.GetEventTypeCounts(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to retrieve event type counts", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiEventTypes := make([]api.EventTypeWithUsage, len(eventTypes))
	for i, eventType := range eventTypes {
		eventCount := eventCounts[eventType.ID] // Will be 0 if not found
		apiEventTypes[i] = api.EventTypeWithUsage{
			Id:          eventType.ID,
			UserId:      eventType.UserID,
			Name:        eventType.Name,
			Description: eventType.Description,
			DefaultName: eventType.DefaultName,
			Color:       eventType.Color,
			CreatedAt:   eventType.CreatedAt,
			UpdatedAt:   eventType.UpdatedAt,
			EventCount:  eventCount,
		}
	}

	response := api.EventTypesResponse{
		EventTypes: apiEventTypes,
	}

	c.JSON(http.StatusOK, response)
}

// CreateEventType creates a new event type for the current user
func (s *APIServer) CreateEventType(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.EventTypeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Validate required fields
	if req.Name == "" {
		appErr := NewAppError("Event type name is required", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	if req.Color == "" {
		appErr := NewAppError("Event type color is required", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Create storage event type
	eventType := &storage.EventType{
		UserID:      user.ID,
		Name:        req.Name,
		Description: req.Description,
		DefaultName: req.DefaultName,
		Color:       req.Color,
	}

	ctx := c.Request.Context()
	err = s.store.CreateEventType(ctx, eventType)
	if err != nil {
		if strings.Contains(err.Error(), "event_type_limit_exceeded") {
			appErr := NewAppError("Event type limit exceeded (50 event types per user)", http.StatusForbidden, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		if strings.Contains(err.Error(), "duplicate") {
			appErr := NewAppError("Event type with this name already exists", http.StatusConflict, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to create event type", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiEventType := api.EventType{
		Id:          eventType.ID,
		Name:        eventType.Name,
		Description: eventType.Description,
		DefaultName: eventType.DefaultName,
		Color:       eventType.Color,
		CreatedAt:   eventType.CreatedAt,
		UpdatedAt:   eventType.UpdatedAt,
	}

	c.JSON(http.StatusCreated, apiEventType)
}

// UpdateEventType updates an existing event type
func (s *APIServer) UpdateEventType(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	var req api.EventTypeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()

	// Get existing event type to ensure it exists and user owns it
	existingEventType, err := s.store.GetEventType(ctx, user.ID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			appErr := NewAppError("Event type not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to retrieve event type", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Update fields if provided
	if req.Name != nil {
		existingEventType.Name = *req.Name
	}
	if req.Description != nil {
		existingEventType.Description = req.Description
	}
	if req.DefaultName != nil {
		existingEventType.DefaultName = req.DefaultName
	}
	if req.Color != nil {
		existingEventType.Color = *req.Color
	}

	err = s.store.UpdateEventType(ctx, existingEventType)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			appErr := NewAppError("Event type with this name already exists", http.StatusConflict, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to update event type", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	// Convert to API format
	apiEventType := api.EventType{
		Id:          existingEventType.ID,
		Name:        existingEventType.Name,
		Description: existingEventType.Description,
		DefaultName: existingEventType.DefaultName,
		Color:       existingEventType.Color,
		CreatedAt:   existingEventType.CreatedAt,
		UpdatedAt:   existingEventType.UpdatedAt,
	}

	c.JSON(http.StatusOK, apiEventType)
}

// DeleteEventType deletes an event type and updates related events
func (s *APIServer) DeleteEventType(c *gin.Context, id string) {
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	ctx := c.Request.Context()
	err = s.store.DeleteEventType(ctx, user.ID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			appErr := NewAppError("Event type not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, c.GetString("request_id"))
			return
		}
		appErr := NewAppError("Failed to delete event type", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, c.GetString("request_id"))
		return
	}

	c.Status(http.StatusNoContent)
}
