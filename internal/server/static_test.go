package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticHandler(t *testing.T) {
	handler := StaticHandler()

	assert.NotNil(t, handler)

	// Test that it serves files (should return the handler without error)
	// Since we're testing the embedded frontend files, let's make a request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Should not panic
	assert.NotPanics(t, func() {
		handler.ServeHTTP(w, req)
	})

	// The handler should respond (status code should be set)
	// It might be 200 for index.html or 404 if no index.html exists
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
}

func TestStaticHandler_SpecificFile(t *testing.T) {
	handler := StaticHandler()

	// Try to access a specific file that should exist (assuming app.js exists)
	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should either serve the file (200) or not find it (404)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)

	if w.Code == http.StatusOK {
		// If the file exists, content type should be set
		contentType := w.Header().Get("Content-Type")
		assert.NotEmpty(t, contentType)
	}
}
