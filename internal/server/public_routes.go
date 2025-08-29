package server

import (
	"slices"
	"strings"
)

// publicEndpoints defines the endpoints that don't require authentication
var publicEndpoints = []string{
	"/api/v1/health",
	"/api/v1/openapi.yaml",
}

// docsPrefix defines the prefix for documentation endpoints that don't require authentication
const docsPrefix = "/api/v1/docs"

// IsPublicEndpoint checks if the given path is a public endpoint that doesn't require authentication
func IsPublicEndpoint(path string) bool {
	// Exact match for specific endpoints
	if slices.Contains(publicEndpoints, path) {
		return true
	}

	// Prefix match for docs endpoints to handle subpaths
	if strings.HasPrefix(path, docsPrefix) {
		return true
	}

	return false
}
