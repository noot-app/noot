package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticHandler_CacheHeaders_Development(t *testing.T) {
	// Ensure we're in development mode
	originalEnv := os.Getenv("ENV")
	os.Setenv("ENV", "development")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	handler := StaticHandler()

	tests := []struct {
		name          string
		path          string
		expectedCache string
		description   string
	}{
		{
			name:          "CSS file",
			path:          "/styles.css",
			expectedCache: "public, max-age=1",
			description:   "CSS files should have very short cache in development",
		},
		{
			name:          "JavaScript file",
			path:          "/app.js",
			expectedCache: "public, max-age=1",
			description:   "JS files should have very short cache in development",
		},
		{
			name:          "JSON file",
			path:          "/simulation-data.json",
			expectedCache: "public, max-age=1",
			description:   "JSON files should have very short cache in development",
		},
		{
			name:          "Root path",
			path:          "/",
			expectedCache: "no-cache",
			description:   "Root path should have no-cache headers",
		},
		{
			name:          "HTML file",
			path:          "/index.html",
			expectedCache: "no-cache",
			description:   "HTML files should have no-cache headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Check that Cache-Control header is set correctly
			cacheControl := w.Header().Get("Cache-Control")
			assert.Equal(t, tt.expectedCache, cacheControl, tt.description)
		})
	}
}

func TestStaticHandler_CacheHeaders_Production(t *testing.T) {
	// Ensure we're in production mode
	originalEnv := os.Getenv("ENV")
	os.Setenv("ENV", "production")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	handler := StaticHandler()

	tests := []struct {
		name          string
		path          string
		expectedCache string
		description   string
	}{
		{
			name:          "CSS file",
			path:          "/styles.css",
			expectedCache: "public, max-age=31536000",
			description:   "CSS files should have long-lived cache headers in production",
		},
		{
			name:          "JavaScript file",
			path:          "/app.js",
			expectedCache: "public, max-age=31536000",
			description:   "JS files should have long-lived cache headers in production",
		},
		{
			name:          "JSON file",
			path:          "/simulation-data.json",
			expectedCache: "public, max-age=31536000",
			description:   "JSON files should have long-lived cache headers in production",
		},
		{
			name:          "Root path",
			path:          "/",
			expectedCache: "no-cache",
			description:   "Root path should have no-cache headers",
		},
		{
			name:          "HTML file",
			path:          "/index.html",
			expectedCache: "no-cache",
			description:   "HTML files should have no-cache headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Check that Cache-Control header is set correctly
			cacheControl := w.Header().Get("Cache-Control")
			assert.Equal(t, tt.expectedCache, cacheControl, tt.description)
		})
	}
}

func TestStaticHandler_StaticAssetTypes(t *testing.T) {
	// Test in development mode (default)
	originalEnv := os.Getenv("ENV")
	os.Setenv("ENV", "development")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	handler := StaticHandler()

	// Test with existing files from the embedded filesystem
	staticAssets := []string{
		"/styles.css",           // This file exists
		"/app.js",               // This file exists
		"/simulation-data.json", // This file exists
	}

	for _, asset := range staticAssets {
		t.Run(asset, func(t *testing.T) {
			req := httptest.NewRequest("GET", asset, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// All static assets should have short cache headers in development
			cacheControl := w.Header().Get("Cache-Control")
			assert.Equal(t, "public, max-age=1", cacheControl)
		})
	}
}

func TestStaticHandler_NoCacheTypes(t *testing.T) {
	handler := StaticHandler()

	noCacheAssets := []string{
		"/",
		"/index.html",
	}

	for _, asset := range noCacheAssets {
		t.Run(asset, func(t *testing.T) {
			req := httptest.NewRequest("GET", asset, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// HTML files and root should have no-cache headers
			cacheControl := w.Header().Get("Cache-Control")
			assert.Equal(t, "no-cache", cacheControl)
		})
	}
}

func TestStaticHandler_HeadersSetBeforeFileServer(t *testing.T) {
	// Test in development mode
	originalEnv := os.Getenv("ENV")
	os.Setenv("ENV", "development")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	handler := StaticHandler()

	// Test that headers are properly set for an existing CSS file
	req := httptest.NewRequest("GET", "/styles.css", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// CSS file should have development cache headers set and return 200
	cacheControl := w.Header().Get("Cache-Control")
	assert.Equal(t, "public, max-age=1", cacheControl)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStaticHandler_VersionInjection_Development(t *testing.T) {
	// Test in development mode (no version injection)
	originalEnv := os.Getenv("ENV")
	os.Setenv("ENV", "development")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	handler := StaticHandler()

	// Test that root path does NOT inject version numbers in development
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should have no-cache headers
	cacheControl := w.Header().Get("Cache-Control")
	assert.Equal(t, "no-cache", cacheControl)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	// Should NOT contain versioned asset links in development
	body := w.Body.String()
	assert.NotContains(t, body, "?v=", "HTML should NOT contain versioned asset links in development")

	// Should contain the original unversioned links
	assert.Contains(t, body, `href="/styles.css"`, "Original CSS link should be preserved in development")
	assert.Contains(t, body, `src="/app.js"`, "Original JS link should be preserved in development")
}

func TestStaticHandler_VersionInjection_Production(t *testing.T) {
	// Test in production mode (with version injection)
	originalEnv := os.Getenv("ENV")
	os.Setenv("ENV", "production")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENV")
		} else {
			os.Setenv("ENV", originalEnv)
		}
	}()

	handler := StaticHandler()

	// Test that root path injects version numbers into HTML in production
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should have no-cache headers
	cacheControl := w.Header().Get("Cache-Control")
	assert.Equal(t, "no-cache", cacheControl)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	// Should contain versioned asset links in production
	body := w.Body.String()
	assert.Contains(t, body, "?v=", "HTML should contain versioned asset links in production")

	// Should NOT contain the original unversioned links
	assert.NotContains(t, body, `href="/styles.css"`, "Original CSS link should be replaced in production")
	assert.NotContains(t, body, `src="/app.js"`, "Original JS link should be replaced in production")

	// Should contain the versioned links
	assert.Contains(t, body, `href="/styles.css?v=`, "CSS should have version parameter in production")
	assert.Contains(t, body, `src="/app.js?v=`, "JS should have version parameter in production")
}
