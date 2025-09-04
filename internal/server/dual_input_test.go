package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Note: Full integration tests are commented out as they require OpenAI API access
// which is not available in the CI environment. The dual input functionality is tested
// through the normalization tests below and through manual testing.

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
