package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	// Initialize logger
	InitLogger()

	tests := []struct {
		name         string
		method       string
		expectedCode int
	}{
		{"GET request", http.MethodGet, http.StatusOK},
		{"POST request", http.MethodPost, http.StatusMethodNotAllowed},
		{"PUT request", http.MethodPut, http.StatusMethodNotAllowed},
		{"DELETE request", http.MethodDelete, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/health", nil)
			w := httptest.NewRecorder()

			healthHandler(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				
				// Check response body contains expected fields
				body := w.Body.String()
				assert.Contains(t, body, "status")
				assert.Contains(t, body, "timestamp")
				assert.Contains(t, body, "version")
				assert.Contains(t, body, "healthy")
			}
		})
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     200,
	}

	rw.WriteHeader(404)

	assert.Equal(t, 404, rw.statusCode)
	assert.Equal(t, 404, w.Code)
}

func TestResponseWriter_DefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     200,
	}

	// Write something without explicitly setting status
	rw.Write([]byte("test"))

	assert.Equal(t, 200, rw.statusCode)
	assert.Equal(t, 200, w.Code)
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	// Should be 12 characters (6 bytes hex encoded)
	assert.Len(t, id1, 12)
	assert.Len(t, id2, 12)

	// Should be different
	assert.NotEqual(t, id1, id2)

	// Should only contain hex characters
	for _, r := range id1 {
		assert.True(t, (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	}
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "context with request ID",
			ctx:      context.WithValue(context.Background(), requestIDKey, "test123"),
			expected: "test123",
		},
		{
			name:     "context without request ID",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "context with wrong type",
			ctx:      context.WithValue(context.Background(), requestIDKey, 123),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRequestID(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	// Initialize logger
	InitLogger()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Check that request ID is in context
		requestID := getRequestID(r.Context())
		assert.NotEmpty(t, requestID)
		assert.Len(t, requestID, 12) // 6 bytes hex encoded
		w.WriteHeader(http.StatusOK)
	})

	middleware := requestIDMiddleware(handler)
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggingMiddleware(t *testing.T) {
	// Initialize logger
	InitLogger()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))
	})

	// Create a request with request ID in context
	req := httptest.NewRequest(http.MethodPost, "/test/path", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "test123"))
	w := httptest.NewRecorder()

	middleware := loggingMiddleware(handler)
	middleware.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestRecoveryMiddleware(t *testing.T) {
	// Initialize logger
	InitLogger()

	t.Run("normal handler", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("normal response"))
		})

		middleware := recoveryMiddleware(handler)
		
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "normal response", w.Body.String())
	})

	t.Run("panicking handler", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		middleware := recoveryMiddleware(handler)
		
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		// Should not panic
		assert.NotPanics(t, func() {
			middleware.ServeHTTP(w, req)
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		
		// Should contain error response
		body := w.Body.String()
		assert.Contains(t, body, "Internal server error")
		assert.Contains(t, body, "error")
	})
}

func TestMiddlewareChaining(t *testing.T) {
	// Initialize logger
	InitLogger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request ID is available (from requestIDMiddleware)
		requestID := getRequestID(r.Context())
		assert.NotEmpty(t, requestID)
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Chain middleware like in the server
	chained := requestIDMiddleware(
		loggingMiddleware(
			recoveryMiddleware(handler),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	chained.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}