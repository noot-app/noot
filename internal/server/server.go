package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// Run configures routes and starts the HTTP server using Gin
func Run(ctx context.Context, port string) error {
	// Initialize storage using new config approach
	config := &storage.Config{
		Type:     getenv("DB_TYPE", "sqlite"),
		Database: getenv("DATABASE_PATH", "./noot.db"),
		// Future PostgreSQL support
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenvInt("DB_PORT", 5432),
		Username: getenv("DB_USER", ""),
		Password: getenv("DB_PASS", ""),
		SSLMode:  getenv("DB_SSLMODE", "prefer"),
	}

	store, err := storage.NewStore(config)
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

	// Set Gin mode
	if env == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Add middleware
	r.Use(RequestIDMiddleware())
	r.Use(LoggingMiddleware())
	r.Use(RecoveryMiddleware())
	r.Use(CORSMiddleware())
	r.Use(StoreMiddleware(store))

	// Create API server
	apiServer := NewAPIServer(store)

	// API v1 routes with OpenAPI generated routing
	v1 := r.Group("/api/v1")
	{
		// Use the generated server interface wrapper
		wrapper := &api.ServerInterfaceWrapper{
			Handler: apiServer,
		}

		v1.POST("/consumption", wrapper.CreateConsumption)
		v1.GET("/health", wrapper.GetHealth)

		// Development-only routes
		if env == "development" {
			v1.GET("/consumptions", wrapper.GetConsumptions)
			v1.GET("/nutrition-summary", wrapper.GetNutritionSummary)
			v1.GET("/docs", apiServer.SwaggerUIHandler)
			v1.GET("/openapi.yaml", apiServer.OpenAPISpecHandler)
		}
	}

	LogInfo("Starting server", "port", port, "debug", isDebugMode(), "env", getenv("ENV", "production"))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

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
