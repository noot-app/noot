package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return def
}

// Enhanced error response structure
type ErrorResponse struct {
	Error   string    `json:"error"`
	Code    int       `json:"code"`
	Time    time.Time `json:"timestamp"`
	Stack   []string  `json:"stack,omitempty"`
	TraceID string    `json:"trace_id,omitempty"`
}

func httpError(w http.ResponseWriter, code int, msg string) {
	httpErrorWithDetails(w, code, msg, nil, "")
}

func httpErrorWithDetails(w http.ResponseWriter, code int, msg string, stack []string, traceID string) {
	LogError("HTTP Error", fmt.Errorf("status=%d message=%s", code, msg),
		"status_code", code, "trace_id", traceID)

	resp := ErrorResponse{
		Error: msg,
		Code:  code,
		Time:  time.Now().UTC(),
	}

	// Include stack trace in debug/dev mode
	if (isDebugMode() || isDevMode()) && len(stack) > 0 {
		resp.Stack = stack
	}

	if traceID != "" {
		resp.TraceID = traceID
	}

	writeJSON(w, code, resp)
}

func handleAppError(w http.ResponseWriter, err *AppError, traceID string) {
	httpErrorWithDetails(w, err.StatusCode, err.Message, err.Stack, traceID)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		LogError("Failed to encode JSON response", err)
	}
}

func saveTempFile(src multipart.File, header *multipart.FileHeader) (string, string, error) {
	// Get max upload size from environment, default to 100MB
	maxBytes := int64(100 * 1024 * 1024) // 100MB default
	if maxBytesStr := os.Getenv("MAX_UPLOAD_BYTES"); maxBytesStr != "" {
		if parsed, err := strconv.ParseInt(maxBytesStr, 10, 64); err == nil && parsed > 0 {
			maxBytes = parsed
		}
	}

	// Create a buffered reader to peek at the first 512 bytes for MIME detection
	reader := bufio.NewReader(src)

	// Peek up to 512 bytes to determine MIME type
	peekBytes := make([]byte, 512)
	n, err := reader.Read(peekBytes)
	if err != nil && err != io.EOF {
		return "", "", fmt.Errorf("failed to read file header: %w", err)
	}

	// Determine MIME type from the peeked bytes
	mime := http.DetectContentType(peekBytes[:n])
	ext := guessExtension(mime)

	// Create temp file with appropriate extension
	tmpFile, err := os.CreateTemp("", "audio-*"+ext)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}

	var cleanupFile = func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}

	// Write the peeked bytes first
	bytesWritten := int64(0)
	if n > 0 {
		written, err := tmpFile.Write(peekBytes[:n])
		if err != nil {
			cleanupFile()
			return "", "", fmt.Errorf("failed to write peeked bytes to temp file: %w", err)
		}
		bytesWritten = int64(written)
	}

	// Stream the remainder using io.Copy with size limit
	remainingBytes := maxBytes - bytesWritten
	if remainingBytes <= 0 {
		cleanupFile()
		return "", "", fmt.Errorf("file exceeds maximum upload size of %d bytes", maxBytes)
	}

	limitedReader := io.LimitReader(reader, remainingBytes)
	copied, err := io.Copy(tmpFile, limitedReader)
	if err != nil {
		cleanupFile()
		return "", "", fmt.Errorf("failed to stream file to temp file: %w", err)
	}

	bytesWritten += copied

	// Check if we hit the limit (which means the file was too large)
	if copied == remainingBytes {
		// Try to read one more byte to see if there's more data
		var oneByte [1]byte
		if extraRead, err := reader.Read(oneByte[:]); err == nil && extraRead > 0 {
			cleanupFile()
			return "", "", fmt.Errorf("file exceeds maximum upload size of %d bytes", maxBytes)
		}
	}

	// Close the file (but keep it on disk)
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", "", fmt.Errorf("failed to close temp file: %w", err)
	}

	return tmpFile.Name(), mime, nil
}

func guessExtension(mime string) string {
	l := strings.ToLower(mime)
	switch {
	case strings.Contains(l, "webm"):
		return ".webm"
	case strings.Contains(l, "ogg"):
		return ".ogg"
	case strings.Contains(l, "mp3"):
		return ".mp3"
	case strings.Contains(l, "mpeg"):
		return ".mp3"
	case strings.Contains(l, "wav"):
		return ".wav"
	default:
		return ".webm"
	}
}

func removeFile(p string) {
	_ = os.Remove(p)
}

// LoadDotEnv tries common locations to load .env key=value into env
func LoadDotEnv() {
	candidates := []string{
		".env",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			loadEnvFile(p)
			return
		}
	}
}

func loadEnvFile(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), `"'`)
		// Only set if not already set (preserve command line env vars)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

// summarize calculates totals and daily value percentages for nutrition items
func summarize(items []ItemWithNutrition) Summary {
	totals := CompleteNutrient{}

	// Sum up all nutrients
	for _, item := range items {
		if item.Item.Nutrients != nil {
			n := item.Item.Nutrients
			totals.Calories += n.Calories
			totals.Protein += n.Protein
			totals.TotalFat += n.TotalFat
			totals.SaturatedFat += n.SaturatedFat
			totals.TransFat += n.TransFat
			totals.Cholesterol += n.Cholesterol
			totals.Sodium += n.Sodium
			totals.TotalCarbs += n.TotalCarbs
			totals.DietaryFiber += n.DietaryFiber
			totals.TotalSugars += n.TotalSugars
			totals.AddedSugars += n.AddedSugars
			totals.VitaminA += n.VitaminA
			totals.VitaminC += n.VitaminC
			totals.VitaminD += n.VitaminD
			totals.VitaminE += n.VitaminE
			totals.VitaminK += n.VitaminK
			totals.Thiamine += n.Thiamine
			totals.Riboflavin += n.Riboflavin
			totals.Niacin += n.Niacin
			totals.VitaminB6 += n.VitaminB6
			totals.Folate += n.Folate
			totals.VitaminB12 += n.VitaminB12
			totals.Calcium += n.Calcium
			totals.Iron += n.Iron
			totals.Magnesium += n.Magnesium
			totals.Phosphorus += n.Phosphorus
			totals.Potassium += n.Potassium
			totals.Zinc += n.Zinc
			totals.Copper += n.Copper
			totals.Manganese += n.Manganese
			totals.Selenium += n.Selenium
		}
	}

	// Daily values based on FDA recommendations for adults (2000 calorie diet)
	dailyValues := map[string]float64{
		"calories":        2000,
		"protein_g":       50,
		"total_fat_g":     78,
		"saturated_fat_g": 20,
		"cholesterol_mg":  300,
		"sodium_mg":       2300,
		"total_carbs_g":   275,
		"dietary_fiber_g": 28,
		"vitamin_a_mcg":   900,
		"vitamin_c_mg":    90,
		"vitamin_d_mcg":   20,
		"calcium_mg":      1300,
		"iron_mg":         18,
		"potassium_mg":    4700,
	}

	// Calculate percentages of daily values for key nutrients
	percentOfDaily := map[string]int{
		"calories":      int((totals.Calories / dailyValues["calories"]) * 100),
		"protein":       int((totals.Protein / dailyValues["protein_g"]) * 100),
		"total_fat":     int((totals.TotalFat / dailyValues["total_fat_g"]) * 100),
		"saturated_fat": int((totals.SaturatedFat / dailyValues["saturated_fat_g"]) * 100),
		"cholesterol":   int((totals.Cholesterol / dailyValues["cholesterol_mg"]) * 100),
		"sodium":        int((totals.Sodium / dailyValues["sodium_mg"]) * 100),
		"total_carbs":   int((totals.TotalCarbs / dailyValues["total_carbs_g"]) * 100),
		"dietary_fiber": int((totals.DietaryFiber / dailyValues["dietary_fiber_g"]) * 100),
		"vitamin_a":     int((totals.VitaminA / dailyValues["vitamin_a_mcg"]) * 100),
		"vitamin_c":     int((totals.VitaminC / dailyValues["vitamin_c_mg"]) * 100),
		"vitamin_d":     int((totals.VitaminD / dailyValues["vitamin_d_mcg"]) * 100),
		"calcium":       int((totals.Calcium / dailyValues["calcium_mg"]) * 100),
		"iron":          int((totals.Iron / dailyValues["iron_mg"]) * 100),
		"potassium":     int((totals.Potassium / dailyValues["potassium_mg"]) * 100),
	}

	return Summary{
		Totals:          totals,
		PercentOfDaily:  percentOfDaily,
		DailyValuesUsed: dailyValues,
	}
}

// Pointer utility functions for easier test data creation and API handling

// stringPtr returns a pointer to the string value
func stringPtr(s string) *string {
	return &s
}

// strPtrOrNil returns a trimmed string pointer or nil if empty
func strPtrOrNil(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
