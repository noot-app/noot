package server

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/grantbirki/noot/internal/version"
)

//go:embed frontend/*
var embeddedFS embed.FS

// StaticHandler serves the embedded frontend with appropriate cache headers.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(embeddedFS, "frontend")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle index.html specially to inject version for cache busting
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			serveIndexWithVersion(w, r)
			return
		}

		// Set cache headers based on file type
		setCacheHeaders(w, r.URL.Path)
		fileServer.ServeHTTP(w, r)
	})
}

// serveIndexWithVersion serves the index.html with cache-busting version numbers
func serveIndexWithVersion(w http.ResponseWriter, r *http.Request) {
	// Set no-cache headers for HTML
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Read the template file
	indexBytes, err := embeddedFS.ReadFile("frontend/index.html")
	if err != nil {
		http.Error(w, "Could not read index.html", http.StatusInternalServerError)
		return
	}

	// Replace static file references with versioned ones (only in production)
	indexContent := string(indexBytes)
	if !isDevMode() {
		// Production: Add version parameters for cache busting
		commitVersion := version.GetCommitShort()
		indexContent = strings.ReplaceAll(indexContent,
			`href="/styles.css"`,
			`href="/styles.css?v=`+commitVersion+`"`)
		indexContent = strings.ReplaceAll(indexContent,
			`src="/app.js"`,
			`src="/app.js?v=`+commitVersion+`"`)
	}
	// Development: Leave URLs unchanged since we have very short cache times

	w.Write([]byte(indexContent))
}

// setCacheHeaders sets appropriate Cache-Control headers based on the file type
func setCacheHeaders(w http.ResponseWriter, requestPath string) {
	// Clean the path and get the file extension
	cleanPath := path.Clean(requestPath)
	ext := strings.ToLower(path.Ext(cleanPath))

	// For root path or HTML files, use no-cache
	if cleanPath == "/" || cleanPath == "/index.html" || ext == ".html" {
		w.Header().Set("Cache-Control", "no-cache")
		return
	}

	// For static assets, use long-lived cache with immutable
	staticExtensions := map[string]bool{
		".css":   true,
		".js":    true,
		".json":  true,
		".svg":   true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".webp":  true,
		".ico":   true,
		".woff":  true,
		".woff2": true,
	}

	if staticExtensions[ext] {
		if isDevMode() {
			// Development: Very short cache to see changes immediately
			w.Header().Set("Cache-Control", "public, max-age=1") // 1 second
		} else {
			// Production: Long cache with version-based cache busting via query params
			w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
		}
	}
}
