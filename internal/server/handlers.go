package server

import (
	"net/http"
)

// POST /api/ingest
func ingestHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())
	store := getStore(r.Context())

	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	LogDebug("Processing ingest request", "request_id", requestID)

	// Accept up to ~100MB form size
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		appErr := NewAppError("Invalid multipart form", http.StatusBadRequest, err)
		handleAppError(w, appErr, requestID)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		appErr := NewAppError("No audio file uploaded (field: audio)", http.StatusBadRequest, err)
		handleAppError(w, appErr, requestID)
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
		handleAppError(w, appErr, requestID)
		return
	}
	defer removeFile(tmpPath)

	LogDebug("Temp file created", "path", tmpPath, "mime_type", mimeType, "request_id", requestID)

	ctx := r.Context()

	// 1) Transcribe
	LogDebug("Starting transcription", "request_id", requestID)
	transcript, err := transcribeAudio(ctx, tmpPath, mimeType)
	if err != nil {
		appErr := NewAppError("Transcription failed", http.StatusInternalServerError, err)
		handleAppError(w, appErr, requestID)
		return
	}
	LogDebug("Transcription completed", "transcript_length", len(transcript), "request_id", requestID)

	// 2) Parse items (phase 1: extract items without nutrition)
	LogDebug("Starting item parsing (items only)", "request_id", requestID)
	parsed, err := parseItems(ctx, transcript)
	if err != nil {
		appErr := NewAppError("Item parsing failed", http.StatusInternalServerError, err)
		handleAppError(w, appErr, requestID)
		return
	}
	LogDebug("Item parsing completed", "item_count", len(parsed.Items), "request_id", requestID)

	// 3) Hydrate nutrition (phase 2: add nutrition data using cache + OpenAI)
	LogDebug("Starting nutrition hydration", "request_id", requestID)
	var hydratedItems []Item
	if store != nil {
		hydratedItems, err = hydrateNutrition(ctx, parsed.Items, store)
		if err != nil {
			appErr := NewAppError("Nutrition hydration failed", http.StatusInternalServerError, err)
			handleAppError(w, appErr, requestID)
			return
		}
	} else {
		// No store available - hydrate without cache (direct OpenAI calls)
		hydratedItems, err = hydrateNutritionWithoutCache(ctx, parsed.Items)
		if err != nil {
			appErr := NewAppError("Nutrition hydration failed", http.StatusInternalServerError, err)
			handleAppError(w, appErr, requestID)
			return
		}
	}
	LogDebug("Nutrition hydration completed", "hydrated_count", len(hydratedItems), "request_id", requestID)

	// 4) Convert to ItemWithNutrition format for response
	var itemsWith []ItemWithNutrition
	for _, it := range hydratedItems {
		iw := ItemWithNutrition{
			Item: it,
		}
		if it.Nutrients == nil {
			iw.Note = "Nutrition data unavailable"
		}
		itemsWith = append(itemsWith, iw)
	}

	// 5) Summarize
	summary := summarize(itemsWith)

	// 6) Save meal to database if store is available
	if store != nil {
		// For now, use a default user (monalisa) if no authentication
		// In the future, this would come from authentication middleware
		user, err := store.GetUserBySubject(ctx, "github", "monalisa")
		if err != nil {
			LogError("Failed to get user for meal storage", err)
		} else if user != nil {
			meal := itemWithNutritionToMeal(user.ID, transcript, itemsWith, summary)
			if err := store.CreateMeal(ctx, meal); err != nil {
				LogError("Failed to save meal to database", err)
				// Don't fail the request if storage fails
			} else {
				LogInfo("Meal saved to database", "meal_id", meal.ID, "user_id", user.ID, "request_id", requestID)
			}
		}
	}

	resp := map[string]any{
		"transcript":  transcript,
		"parsedItems": parsed.Items,
		"items":       itemsWith,
		"summary":     summary,
		"request_id":  requestID,
	}

	LogInfo("Ingest request completed successfully",
		"transcript_length", len(transcript),
		"items_count", len(itemsWith),
		"total_calories", summary.Totals.Calories,
		"request_id", requestID,
	)

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/meals - Development only endpoint to view stored meals
func mealsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	store := getStore(r.Context())
	if store == nil {
		httpError(w, http.StatusInternalServerError, "Storage not available")
		return
	}

	// For now, get meals for the default user (monalisa)
	user, err := store.GetUserBySubject(r.Context(), "github", "monalisa")
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}
	if user == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"meals": []interface{}{},
			"user":  nil,
		})
		return
	}

	// Get recent meals (last 50)
	meals, err := store.GetMealsByUser(r.Context(), user.ID, 50, 0)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed to get meals")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"meals": meals,
		"user":  user,
		"count": len(meals),
	})
}
