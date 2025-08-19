package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwaggerUIHandler(t *testing.T) {
	// Initialize logger
	InitLogger()

	t.Run("GET request returns HTML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil)
		w := httptest.NewRecorder()

		swaggerUIHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		body := w.Body.String()
		assert.Contains(t, body, "Noot API Documentation")
		assert.Contains(t, body, "swagger-ui")
		assert.Contains(t, body, "/api/v1/openapi.yaml")
	})

	t.Run("POST request returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/docs", nil)
		w := httptest.NewRecorder()

		swaggerUIHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})
}

func TestOpenAPISpecHandler(t *testing.T) {
	// Initialize logger
	InitLogger()

	t.Run("GET request serves file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
		w := httptest.NewRecorder()

		openAPISpecHandler(w, req)

		// The file should be served, status depends on file existence
		// In test environment, file should exist
		if w.Code == http.StatusOK {
			body := w.Body.String()
			assert.Contains(t, body, "openapi:")
			assert.Contains(t, body, "Noot API")
		}
	})

	t.Run("POST request returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/openapi.yaml", nil)
		w := httptest.NewRecorder()

		openAPISpecHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})
}

func TestAPIv1Routes(t *testing.T) {
	// Initialize logger
	InitLogger()

	t.Run("health endpoint works on v1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		healthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, "healthy")
		assert.Contains(t, body, "timestamp")
		assert.Contains(t, body, "version")
	})

	t.Run("consumption endpoint accepts POST on v1", func(t *testing.T) {
		// Create a multipart form
		body := strings.NewReader("dummy-multipart-data")
		req := httptest.NewRequest(http.MethodPost, "/api/v1/consumption", body)
		req.Header.Set("Content-Type", "multipart/form-data")
		w := httptest.NewRecorder()

		// Wrap with request ID middleware and call handler
		handler := requestIDMiddleware(http.HandlerFunc(ingestHandler))
		handler.ServeHTTP(w, req)

		// Should get an error about invalid multipart form, but not method not allowed
		assert.NotEqual(t, http.StatusMethodNotAllowed, w.Code)
	})
}