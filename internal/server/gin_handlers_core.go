package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// GetHealth implements ServerInterface.GetHealth
func (s *APIServer) GetHealth(c *gin.Context) {
	response := api.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   getenv("VERSION", "dev"),
	}
	c.JSON(http.StatusOK, response)
}

// ExportData implements ServerInterface.ExportData
func (s *APIServer) ExportData(c *gin.Context, params api.ExportDataParams) {
	requestID := getRequestID(c)
	ctx := c.Request.Context()

	// Get the default user for now (in production, get from auth)
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Failed to get user", http.StatusNotFound, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check subscription tier
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		appErr := NewAppError("Pro subscription required for data export", http.StatusForbidden, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse date range parameters
	start, end, _, err := parseTrendsDateRangeParams(params.Start, params.End, nil)
	if err != nil {
		appErr := NewAppError("Invalid date range parameters", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Get consumption data
	consumptions, err := s.store.GetConsumptionsByUserSince(ctx, user.ID, start)
	if err != nil {
		appErr := NewAppError("Failed to get consumptions", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse requested metrics (default to common macros)
	metrics := []string{"calories", "protein_g", "total_fat_g", "total_carbs_g"}
	if params.Metrics != nil && *params.Metrics != "" {
		// TODO: Implement proper parsing and validation
	}

	if params.Format == "csv" {
		// Generate CSV export
		csvData := generateCSVExport(consumptions, metrics, start, end)
		c.Header("Content-Disposition", "attachment; filename=\"nutrition-export.csv\"")
		c.Data(http.StatusOK, "text/csv", []byte(csvData))
	} else {
		// Generate JSON export (same as trends response)
		series := generateTimeSeries(consumptions, metrics, start, end)
		response := api.ExportResponse{
			Series: series,
			User:   convertUser(user),
			DateRange: struct {
				End   *time.Time `json:"end,omitempty"`
				Start *time.Time `json:"start,omitempty"`
			}{
				Start: &start,
				End:   &end,
			},
			Format: api.ExportResponseFormat(params.Format),
		}
		c.JSON(http.StatusOK, response)
	}
}

// Development-only handlers

// SwaggerUIHandler serves Swagger UI for API documentation
func (s *APIServer) SwaggerUIHandler(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Noot API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/api/v1/openapi.yaml',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// OpenAPISpecHandler serves the OpenAPI specification
func (s *APIServer) OpenAPISpecHandler(c *gin.Context) {
	c.File("api/openapi.yaml")
}
