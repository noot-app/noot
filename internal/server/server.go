package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// Define custom context key types to avoid collisions
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	storeKey     contextKey = "store"
)

// Run configures routes and starts the HTTP server
func Run(ctx context.Context, port string) error {
	// Initialize storage
	dbPath := getenv("DATABASE_PATH", "./noot.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Run migrations
	if err := store.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Run seeding in development
	env := strings.ToLower(getenv("ENV", "production"))
	devSeed := strings.ToLower(getenv("DEV_DB_SEED", "false")) == "true"
	if env == "development" || devSeed {
		if err := store.Seed(); err != nil {
			LogError("Failed to seed database", err)
			// Don't fail startup on seed error, just log it
		}
	}

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/ingest", ingestHandler)
	mux.HandleFunc("/api/health", healthHandler)

	// Development-only meals API
	if env == "development" {
		mux.HandleFunc("/api/meals", mealsHandler)
	}

	// Static frontend (embedded)
	mux.Handle("/", StaticHandler())

	// Chain middleware with store injection
	handler := requestIDMiddleware(
		loggingMiddleware(
			recoveryMiddleware(
				storeMiddleware(store, mux),
			),
		),
	)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	LogInfo("Starting server", "port", port, "debug", isDebugMode(), "env", getenv("ENV", "production"))

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		LogInfo("Shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			LogError("Server shutdown error", err)
		}
		LogInfo("Server stopped")
		return nil
	case err := <-errCh:
		LogError("Server error", err)
		return err
	}
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID
		requestID := generateRequestID()

		// Add to context and response header
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := getRequestID(r.Context())

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		LogDebug("Request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"request_id", requestID,
		)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		LogInfo("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := getRequestID(r.Context())

				LogError("Panic recovered", fmt.Errorf("panic: %v", rec),
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
				)

				if isDebugMode() || isDevMode() {
					stack := captureStack(4)
					httpErrorWithDetails(w, http.StatusInternalServerError,
						"Internal server error (panic recovered)", stack, requestID)
				} else {
					httpError(w, http.StatusInternalServerError, "Internal server error")
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   getenv("VERSION", "dev"),
	})
}

// Helper types and functions
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func generateRequestID() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func getStore(ctx context.Context) storage.Store {
	if store, ok := ctx.Value(storeKey).(storage.Store); ok {
		return store
	}
	return nil
}

func storeMiddleware(store storage.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), storeKey, store)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
