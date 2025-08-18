package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
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

func saveTempFile(src multipart.File, _ *multipart.FileHeader) (string, string, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		return "", "", err
	}
	b := buf.Bytes()
	mime := http.DetectContentType(b)
	ext := guessExtension(mime)
	tmpFile, err := os.CreateTemp("", "audio-*"+ext)
	if err != nil {
		return "", "", err
	}
	defer tmpFile.Close()
	if _, err := tmpFile.Write(b); err != nil {
		return "", "", err
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
	case strings.Contains(l, "wav"):
		return ".wav"
	default:
		return ".webm"
	}
}

func isFinite(f float64) bool {
	return !((f != f) || (f > 1e308) || (f < -1e308))
}

func maxFloat(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
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
