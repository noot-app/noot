package server

import (
	"net/http"
)

// POST /api/ingest
func ingestHandler(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

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

	// 2) Parse items via Chat Completions
	LogDebug("Starting item parsing", "request_id", requestID)
	parsed, err := parseItems(ctx, transcript)
	if err != nil {
		appErr := NewAppError("Parsing failed", http.StatusInternalServerError, err)
		handleAppError(w, appErr, requestID)
		return
	}
	LogDebug("Item parsing completed", "item_count", len(parsed.Items), "request_id", requestID)

	// 3) Resolve FDC and nutrients per item (sequential for simplicity)
	LogDebug("Starting nutrition lookup", "request_id", requestID)
	var itemsWith []ItemWithNutrition
	for i, it := range parsed.Items {
		LogDebug("Looking up nutrition", "item", it.Name, "index", i, "request_id", requestID)

		iw, err := resolveItemNutrition(ctx, it)
		if err != nil {
			LogWarn("Nutrition lookup failed", "item", it.Name, "error", err.Error(), "request_id", requestID)
			iw = ItemWithNutrition{
				Item:  it,
				Note:  "Error: " + err.Error(),
				Nutri: nil,
				FDC:   nil,
			}
		}
		itemsWith = append(itemsWith, iw)
	}
	LogDebug("Nutrition lookup completed", "request_id", requestID)

	// 4) Summarize
	summary := summarize(itemsWith)

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
		"total_calories", summary.Totals.EnergyKcal,
		"request_id", requestID,
	)

	writeJSON(w, http.StatusOK, resp)
}
