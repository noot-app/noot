package server

import (
	"context"
	"fmt"
	"net/http"
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

	// Initialize storage using shared config creation function
	config := CreateDatabaseConfig()

	store, err := storage.NewStore(config)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Set Gin mode
	if !IsProduction() {
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

	// Authentication middleware
	r.Use(JWTAuthMiddleware(store)) // Handles Supabase auth via JWT tokens

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
		if !IsProduction() {
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
	// Validate JWT configuration in production
	if IsProduction() {
		jwtSecret := getenv("SUPABASE_JWT_SECRET", "")
		supabaseURL := getenv("PUBLIC_SUPABASE_URL", "")

		// For modern Supabase, we need either the URL (for JWKS) or secret (for legacy)
		if jwtSecret == "" && supabaseURL == "" {
			return fmt.Errorf("either PUBLIC_SUPABASE_URL (for JWKS) or SUPABASE_JWT_SECRET (for legacy) is required in production")
		}

		// If using legacy JWT secret, validate its strength
		if jwtSecret != "" && len(jwtSecret) < 32 {
			LogWarn("JWT secret is shorter than recommended minimum of 32 characters")
		}

		// Recommend JWKS over legacy secrets
		if jwtSecret != "" && supabaseURL == "" {
			LogWarn("Using legacy JWT secret - consider upgrading to JWKS-based verification with PUBLIC_SUPABASE_URL")
		}

		// Validate CORS configuration in production
		corsOrigins := getenv("CORS_ALLOWED_ORIGINS", "")
		if corsOrigins == "" || corsOrigins == "http://localhost:3000" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS must be properly configured in production environment, cannot be empty or localhost")
		}
	}

	// Validate environment consistency
	if !IsProduction() {
		jwtSecret := getenv("SUPABASE_JWT_SECRET", "")
		if jwtSecret != "" {
			LogInfo("JWT secret configured - using Supabase authentication")
		}
	}

	env := getenv("ENV", "production") // For logging purposes only
	LogInfo("Security configuration validated", "env", env, "is_production", IsProduction())
	return nil
}
