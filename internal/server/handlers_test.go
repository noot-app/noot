package server

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIngestHandler_MethodNotAllowed(t *testing.T) {
	// Initialize logger
	InitLogger()

	req := httptest.NewRequest(http.MethodGet, "/api/ingest", nil)
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestIngestHandler_InvalidMultipartForm(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Create request with invalid content type
	req := httptest.NewRequest(http.MethodPost, "/api/ingest", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/json") // Wrong content type
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid multipart form")
}

func TestIngestHandler_NoAudioFile(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Create multipart form without audio field
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("other", "value")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "No audio file uploaded")
}

func TestIngestHandler_AudioFileTooBig(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Set max upload bytes to very small value
	originalEnv := os.Getenv("MAX_UPLOAD_BYTES")
	os.Setenv("MAX_UPLOAD_BYTES", "100") // 100 bytes
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("MAX_UPLOAD_BYTES")
		} else {
			os.Setenv("MAX_UPLOAD_BYTES", originalEnv)
		}
	}()

	// Create multipart form with large audio file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a file larger than 100 bytes
	part, err := writer.CreateFormFile("audio", "test.webm")
	assert.NoError(t, err)

	largeData := bytes.Repeat([]byte("x"), 200) // 200 bytes
	part.Write(largeData)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to save upload")
}

func TestIngestHandler_MissingOpenAIKey(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Unset OpenAI API key to test error handling
	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		}
	}()

	// Create valid multipart form with small audio file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("audio", "test.webm")
	assert.NoError(t, err)

	// Add WebM header to make it valid
	webmHeader := []byte{0x1A, 0x45, 0xDF, 0xA3}
	smallData := append(webmHeader, bytes.Repeat([]byte("x"), 50)...)
	part.Write(smallData)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/ingest", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Transcription failed")
}

func TestTranscribeAudio_MissingAPIKey(t *testing.T) {
	// Initialize logger
	InitLogger()

	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		}
	}()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-audio-*.webm")
	assert.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	ctx := context.Background()
	_, err = transcribeAudio(ctx, tmpFile.Name(), "audio/webm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY not configured")
}

func TestParseItems_MissingAPIKey(t *testing.T) {
	// Initialize logger
	InitLogger()

	originalKey := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENAI_API_KEY", originalKey)
		}
	}()

	ctx := context.Background()
	_, err := parseItems(ctx, "I had an apple")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OPENAI_API_KEY not configured")
}
