package server

import (
	"net/http"
)

// POST /api/ingest
func ingestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Accept up to ~100MB form size
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		httpError(w, http.StatusBadRequest, "Invalid multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		httpError(w, http.StatusBadRequest, "No audio file uploaded (field: audio)")
		return
	}
	defer file.Close()

	// Save to temp file
	tmpPath, mimeType, err := saveTempFile(file, header)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed to save upload: "+err.Error())
		return
	}
	defer removeFile(tmpPath)

	ctx := r.Context()

	// 1) Transcribe
	transcript, err := transcribeAudio(ctx, tmpPath, mimeType)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Transcription failed: "+err.Error())
		return
	}

	// 2) Parse items via Chat Completions
	parsed, err := parseItems(ctx, transcript)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Parsing failed: "+err.Error())
		return
	}

	// 3) Resolve FDC and nutrients per item (sequential for simplicity)
	var itemsWith []ItemWithNutrition
	for _, it := range parsed.Items {
		iw, err := resolveItemNutrition(ctx, it)
		if err != nil {
			iw = ItemWithNutrition{
				Item:  it,
				Note:  "Error: " + err.Error(),
				Nutri: nil,
				FDC:   nil,
			}
		}
		itemsWith = append(itemsWith, iw)
	}

	// 4) Summarize
	summary := summarize(itemsWith)

	resp := map[string]any{
		"transcript":  transcript,
		"parsedItems": parsed.Items,
		"items":       itemsWith,
		"summary":     summary,
	}

	writeJSON(w, http.StatusOK, resp)
}
