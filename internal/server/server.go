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
	// Validate environment and security configuration on startup
	if err := validateSecurityConfiguration(); err != nil {
		return fmt.Errorf("security configuration validation failed: %w", err)
	}
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
	r.Use(SecurityHeadersMiddleware()) // Add security headers
	r.Use(CORSMiddleware())
	r.Use(StoreMiddleware(store))
	
	// Authentication middleware - both dev and production
	r.Use(DevAuthMiddleware(store))  // Handles development auth via X-Dev-User-ID header
	r.Use(JWTAuthMiddleware(store))  // Handles production auth via JWT tokens

	// Create API server
	apiServer, err := NewAPIServer(store)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	// API v1 routes with OpenAPI generated routing
	v1 := r.Group("/api/v1")
	{
		// Use the generated RegisterHandlers to include all routes including biometrics
		api.RegisterHandlers(v1, apiServer)

		// Development-only routes that override or supplement the generated routes
		if env == "development" {
			// OpenAPI spec routes
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

// validateSecurityConfiguration performs startup security validation
func validateSecurityConfiguration() error {
	env := strings.ToLower(getenv("ENV", "production"))
	
	// Validate JWT configuration in production
	if env == "production" {
		jwtSecret := getenv("SUPABASE_JWT_SECRET", "")
		if jwtSecret == "" {
			return fmt.Errorf("SUPABASE_JWT_SECRET is required in production")
		}
		if len(jwtSecret) < 32 {
			LogWarn("JWT secret is shorter than recommended minimum of 32 characters")
		}
		
		// Validate CORS configuration in production
		corsOrigins := getenv("CORS_ALLOWED_ORIGINS", "")
		if corsOrigins == "" || corsOrigins == "http://localhost:3000" {
			LogWarn("CORS_ALLOWED_ORIGINS not configured for production - using default localhost")
		}
	}
	
	// Warn about dev auth in production-like environments
	if env != "development" {
		if isProductionEnvironmentForValidation() {
			LogWarn("Production environment detected - ensure dev auth is properly disabled")
		}
	}
	
	// Validate environment consistency
	if env == "development" {
		jwtSecret := getenv("SUPABASE_JWT_SECRET", "")
		if jwtSecret != "" {
			LogWarn("JWT secret configured in development - authentication will use JWT instead of dev auth")
		}
	}
	
	LogInfo("Security configuration validated", "env", env)
	return nil
}

// isProductionEnvironmentForValidation checks for production indicators during startup validation
func isProductionEnvironmentForValidation() bool {
	prodIndicators := []string{
		"VERCEL", "NETLIFY", "HEROKU", "AWS_LAMBDA_FUNCTION_NAME", 
		"GOOGLE_CLOUD_PROJECT", "CF_PAGES", "RENDER",
	}
	
	for _, indicator := range prodIndicators {
		if getenv(indicator, "") != "" {
			return true
		}
	}
	return false
}
