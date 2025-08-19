package server

import (
	"net/http"
	"time"
)

// v1HealthHandler handles GET /api/v1/health using the new context pattern
func v1HealthHandler(c *Context) {
	if c.Method() != http.MethodGet {
		c.ErrorJSON(http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   getenv("VERSION", "dev"),
	})
}

// v1ConsumptionHandler handles POST /api/v1/consumption using the new context pattern
func v1ConsumptionHandler(c *Context) {
	// For now, delegate to the existing handler
	// This preserves all existing logic while using the new pattern
	ingestHandler(c.Writer, c.Request)
}

// v1ConsumptionsHandler handles GET /api/v1/consumptions using the new context pattern
func v1ConsumptionsHandler(c *Context) {
	// For now, delegate to the existing handler
	consumptionsHandler(c.Writer, c.Request)
}

// v1NutritionSummaryHandler handles GET /api/v1/nutrition-summary using the new context pattern
func v1NutritionSummaryHandler(c *Context) {
	// For now, delegate to the existing handler
	nutritionSummaryHandler(c.Writer, c.Request)
}

// v1SwaggerUIHandler serves Swagger UI using the new context pattern
func v1SwaggerUIHandler(c *Context) {
	if c.Method() != http.MethodGet {
		c.ErrorJSON(http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

	c.HTML(http.StatusOK, html)
}

// v1OpenAPISpecHandler serves the OpenAPI specification using the new context pattern
func v1OpenAPISpecHandler(c *Context) {
	if c.Method() != http.MethodGet {
		c.ErrorJSON(http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	c.ServeFile("api/openapi.yaml")
}
