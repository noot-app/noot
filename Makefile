# Generate API types from OpenAPI specification
generate-api:
	oapi-codegen -generate types -package api -o internal/api/types.gen.go api/openapi.yaml

# Generate Gin server interfaces (optional validation middleware)
generate-gin-server:
	oapi-codegen -generate gin -package api -o internal/api/gin.gen.go api/openapi.yaml

# Generate all API code
generate: generate-api generate-gin-server

# Validate OpenAPI specification
validate-api:
	@echo "Validating OpenAPI specification..."
	@command -v openapi-generator-cli >/dev/null 2>&1 || { \
		echo "openapi-generator-cli not found. Install it via npm: npm install -g @openapitools/openapi-generator-cli"; \
		exit 1; \
	}
	openapi-generator-cli validate -i api/openapi.yaml

.PHONY: generate generate-api generate-gin-server validate-api