package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	InitLogger()

	t.Run("NewContext creates context with correct values", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		ctx := NewContext(w, req)

		assert.Equal(t, w, ctx.Writer)
		assert.Equal(t, req, ctx.Request)
		assert.Equal(t, http.MethodGet, ctx.Method())
		assert.Equal(t, "/test", ctx.Path())
	})

	t.Run("JSON response works correctly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		ctx := NewContext(w, req)

		data := map[string]string{"message": "hello"}
		ctx.JSON(200, data)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "hello", response["message"])
	})

	t.Run("ErrorJSON response works correctly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		ctx := NewContext(w, req)

		ctx.ErrorJSON(400, "test error")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test error", response.Error)
		assert.Equal(t, 400, response.Code)
		assert.NotZero(t, response.Time)
	})

	t.Run("HTML response works correctly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		ctx := NewContext(w, req)

		ctx.HTML(200, "<h1>Hello</h1>")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Equal(t, "<h1>Hello</h1>", w.Body.String())
	})

	t.Run("Query parameter extraction works", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test?name=value&empty=", nil)
		w := httptest.NewRecorder()
		ctx := NewContext(w, req)

		assert.Equal(t, "value", ctx.Query("name"))
		assert.Equal(t, "", ctx.Query("empty"))
		assert.Equal(t, "", ctx.Query("missing"))
	})
}

func TestV1HealthHandler(t *testing.T) {
	InitLogger()

	t.Run("GET returns health status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		handler := WrapHandler(v1HealthHandler)
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
		assert.Contains(t, response, "timestamp")
		assert.Contains(t, response, "version")
	})

	t.Run("POST returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
		w := httptest.NewRecorder()

		handler := WrapHandler(v1HealthHandler)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Method not allowed", response.Error)
		assert.Equal(t, 405, response.Code)
	})
}

func TestV1SwaggerUIHandler(t *testing.T) {
	InitLogger()

	t.Run("GET returns HTML page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil)
		w := httptest.NewRecorder()

		handler := WrapHandler(v1SwaggerUIHandler)
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))

		body := w.Body.String()
		assert.Contains(t, body, "Noot API Documentation")
		assert.Contains(t, body, "swagger-ui")
		assert.Contains(t, body, "/api/v1/openapi.yaml")
	})

	t.Run("POST returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/docs", nil)
		w := httptest.NewRecorder()

		handler := WrapHandler(v1SwaggerUIHandler)
		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Method not allowed", response.Error)
		assert.Equal(t, 405, response.Code)
	})
}

func TestWrapHandler(t *testing.T) {
	InitLogger()

	t.Run("WrapHandler creates proper http.HandlerFunc", func(t *testing.T) {
		testHandler := func(c *Context) {
			c.JSON(200, map[string]string{"test": "value"})
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler := WrapHandler(testHandler)
		wrappedHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "value", response["test"])
	})
}
