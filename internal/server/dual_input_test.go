package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDualConsumptionInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupRequest   func() (*http.Request, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "JSON text input",
			setupRequest: func() (*http.Request, error) {
				payload := map[string]string{
					"text": "I ate an apple and banana",
				}
				body, err := json.Marshal(payload)
				if err != nil {
					return nil, err
				}

				req, err := http.NewRequest("POST", "/consumption", bytes.NewReader(body))
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "apple")
				assert.Contains(t, body, "banana")
				assert.Contains(t, body, "input_source")
				assert.Contains(t, body, "text")
			},
		},
		{
			name: "Multipart text input",
			setupRequest: func() (*http.Request, error) {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)

				// Add text field (should take precedence over any audio)
				err := writer.WriteField("text", "I had a chicken salad")
				if err != nil {
					return nil, err
				}

				err = writer.Close()
				if err != nil {
					return nil, err
				}

				req, err := http.NewRequest("POST", "/consumption", &buf)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "chicken")
				assert.Contains(t, body, "salad")
				assert.Contains(t, body, "input_source")
				assert.Contains(t, body, "text")
			},
		},
		{
			name: "Invalid JSON",
			setupRequest: func() (*http.Request, error) {
				req, err := http.NewRequest("POST", "/consumption", strings.NewReader(`{"invalid": json}`))
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name: "Empty text JSON",
			setupRequest: func() (*http.Request, error) {
				payload := map[string]string{
					"text": "",
				}
				body, err := json.Marshal(payload)
				if err != nil {
					return nil, err
				}

				req, err := http.NewRequest("POST", "/consumption", bytes.NewReader(body))
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", "application/json")
				return req, nil
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid JSON or missing text field")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server without store (to avoid DB dependencies)
			server, err := NewAPIServer(nil)
			require.NoError(t, err)

			router := gin.New()
			// Add minimal middleware for request ID
			router.Use(func(c *gin.Context) {
				c.Set("request_id", "test-request-123")
				c.Next()
			})

			router.POST("/consumption", server.CreateConsumption)

			req, err := tt.setupRequest()
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w.Body.String())
		})
	}
}

func TestNormalizeConsumptionInput(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() (*gin.Context, *httptest.ResponseRecorder)
		expectedSource string
		expectError    bool
	}{
		{
			name: "JSON input normalization",
			setupContext: func() (*gin.Context, *httptest.ResponseRecorder) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				payload := map[string]string{
					"text": "test meal description",
				}
				body, _ := json.Marshal(payload)

				req, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				return c, w
			},
			expectedSource: "text",
			expectError:    false,
		},
		{
			name: "Multipart text input normalization",
			setupContext: func() (*gin.Context, *httptest.ResponseRecorder) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				_ = writer.WriteField("text", "test meal description")
				_ = writer.Close()

				req, _ := http.NewRequest("POST", "/", &buf)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				c.Request = req

				return c, w
			},
			expectedSource: "text",
			expectError:    false,
		},
		{
			name: "Empty JSON text",
			setupContext: func() (*gin.Context, *httptest.ResponseRecorder) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)

				payload := map[string]string{
					"text": "   ",
				}
				body, _ := json.Marshal(payload)

				req, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				c.Request = req

				return c, w
			},
			expectedSource: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := tt.setupContext()

			input, err := normalizeConsumptionInput(c, "test-request-123", 50<<20)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, input)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, input)
				assert.Equal(t, tt.expectedSource, input.Source)
				assert.NotEmpty(t, input.Text)
				assert.Equal(t, "test-request-123", input.RequestID)
			}
		})
	}
}
