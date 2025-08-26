package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/grantbirki/noot/internal/server"
	"github.com/grantbirki/noot/internal/storage"
)

// getenv gets environment variable with fallback
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getenvInt gets environment variable as int with fallback
func getenvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

// runMigrationsOnly runs database migrations without starting the server
func runMigrationsOnly() error {
	// Initialize storage using the same config approach as server.Run
	config := &storage.Config{
		Type:     getenv("DATABASE_PROVIDER", "sqlite"),
		Database: getenv("DATABASE_PATH", "./noot.db"),
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenvInt("DB_PORT", 5432),
		Username: getenv("DB_USER", ""),
		Password: getenv("DB_PASS", ""),
		SSLMode:  getenv("DB_SSLMODE", "prefer"),
	}

	// For Supabase, use SUPABASE_DB_URL if provided
	if config.Type == "supabase" && getenv("SUPABASE_DB_URL", "") != "" {
		config.Database = getenv("SUPABASE_DB_URL", "")
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

	return nil
}

func main() {
	// Optional .env loader
	server.LoadDotEnv()

	// Initialize logger after env variables are loaded
	server.InitLogger()

	// Check for migrate-only flag
	for _, arg := range os.Args[1:] {
		if arg == "--migrate-only" {
			if err := runMigrationsOnly(); err != nil {
				fmt.Printf("Migration failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Migrations completed successfully")
			return
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("noot-api running at http://localhost:%s\n", port)
	if err := server.Run(ctx, port); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
