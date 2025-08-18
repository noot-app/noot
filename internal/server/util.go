package server

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func httpError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{
		"error": msg,
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
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
		_ = os.Setenv(key, val)
	}
}
