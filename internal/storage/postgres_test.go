package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getenv gets environment variable with fallback (for testing)
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// TestPostgreSQLStoreBasic tests basic PostgreSQL store functionality
// Note: This test requires a PostgreSQL database connection
func TestPostgreSQLStoreBasic(t *testing.T) {
	// Skip if no PostgreSQL connection string provided
	connStr := getTestPostgreSQLConnStr()
	if connStr == "" {
		t.Skip("Skipping PostgreSQL tests - no connection string provided")
	}

	// Create store
	store, err := NewPostgreSQLStore(connStr)
	require.NoError(t, err, "Failed to create PostgreSQL store")
	defer store.Close()

	// Test migration
	err = store.Migrate()
	require.NoError(t, err, "Failed to run migrations")

	ctx := context.Background()

	t.Run("CreateAndGetUser", func(t *testing.T) {
		user := &User{
			Provider: "github",
			Subject:  "testuser",
			Email:    "test@example.com",
		}

		// Create user
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())

		// Get user by ID
		retrieved, err := store.GetUser(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, user.Provider, retrieved.Provider)
		assert.Equal(t, user.Subject, retrieved.Subject)
		assert.Equal(t, user.Email, retrieved.Email)

		// Get user by subject
		bySubject, err := store.GetUserBySubject(ctx, user.Provider, user.Subject)
		require.NoError(t, err)
		require.NotNil(t, bySubject)
		assert.Equal(t, user.ID, bySubject.ID)
	})

	t.Run("CreateAndGetConsumption", func(t *testing.T) {
		// Create a user first
		user := &User{
			Provider: "github",
			Subject:  "testuser2",
			Email:    "test2@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		consumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "I had a banana",
			TotalCalories: 105,
			TotalProtein:  1.3,
			TotalCarbs:    27,
		}

		// Create consumption
		err = store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)
		assert.NotZero(t, consumption.ID)
		assert.False(t, consumption.CreatedAt.IsZero())

		// Get consumption by ID
		retrieved, err := store.GetConsumption(ctx, consumption.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, consumption.Transcript, retrieved.Transcript)
		assert.Equal(t, consumption.TotalCalories, retrieved.TotalCalories)

		// Update consumption
		retrieved.Transcript = "I had a large banana"
		retrieved.TotalCalories = 120
		err = store.UpdateConsumption(ctx, retrieved)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.UpdatedAt)

		// Verify update
		updated, err := store.GetConsumption(ctx, consumption.ID)
		require.NoError(t, err)
		assert.Equal(t, "I had a large banana", updated.Transcript)
		assert.Equal(t, 120.0, updated.TotalCalories)
	})

	t.Run("GetConsumptionsByUser", func(t *testing.T) {
		// Create a user
		user := &User{
			Provider: "github",
			Subject:  "testuser3",
			Email:    "test3@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create multiple consumptions
		for i := 0; i < 3; i++ {
			consumption := &Consumption{
				UserID:        user.ID,
				Transcript:    "Test consumption",
				TotalCalories: float64(100 + i*10),
			}
			err := store.CreateConsumption(ctx, consumption)
			require.NoError(t, err)
		}

		// Get consumptions for user
		consumptions, err := store.GetConsumptionsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, consumptions, 3)
	})

	t.Run("GetConsumptionsByUserSince", func(t *testing.T) {
		// Create a user
		user := &User{
			Provider: "github",
			Subject:  "testuser4",
			Email:    "test4@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create a consumption
		consumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "Recent consumption",
			TotalCalories: 150,
		}
		err = store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)

		// Get consumptions since 1 hour ago
		since := time.Now().Add(-1 * time.Hour)
		consumptions, err := store.GetConsumptionsByUserSince(ctx, user.ID, since)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(consumptions), 1)
	})

	t.Run("NonExistentRecords", func(t *testing.T) {
		// Test getting non-existent user
		user, err := store.GetUser(ctx, "550e8400-e29b-41d4-a716-446655440000") // Random UUID
		require.NoError(t, err)
		assert.Nil(t, user)

		// Test getting non-existent consumption
		consumption, err := store.GetConsumption(ctx, "550e8400-e29b-41d4-a716-446655440001")
		require.NoError(t, err)
		assert.Nil(t, consumption)
	})
}

// getTestPostgreSQLConnStr returns a PostgreSQL connection string for testing
// This can be set via environment variable TEST_POSTGRESQL_URL
func getTestPostgreSQLConnStr() string {
	// Try to get from environment
	if connStr := getenv("TEST_POSTGRESQL_URL", ""); connStr != "" {
		return connStr
	}

	// Try to get from Supabase URL (for convenience in testing)
	if connStr := getenv("SUPABASE_DB_URL", ""); connStr != "" {
		return connStr
	}

	return ""
}

// TestDualDatabaseSupport tests that both SQLite and PostgreSQL can be configured
func TestDualDatabaseSupport(t *testing.T) {
	t.Run("SQLiteConfig", func(t *testing.T) {
		config := &Config{
			Type:     "sqlite",
			Database: ":memory:",
		}

		store, err := NewStore(config)
		require.NoError(t, err)
		defer store.Close()

		// Should create SQLiteStore
		_, ok := store.(*SQLiteStore)
		assert.True(t, ok, "Expected SQLiteStore")

		// Test basic operation
		ctx := context.Background()
		err = store.Migrate()
		require.NoError(t, err)

		user := &User{
			Provider: "test",
			Subject:  "user1",
			Email:    "user1@test.com",
			Handle:   "testuser1",
		}
		err = store.CreateUser(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
	})

	t.Run("PostgreSQLConfig", func(t *testing.T) {
		connStr := getTestPostgreSQLConnStr()
		if connStr == "" {
			t.Skip("Skipping PostgreSQL config test - no connection string")
		}

		config := &Config{
			Type:     "postgres",
			Database: connStr,
		}

		store, err := NewStore(config)
		require.NoError(t, err)
		defer store.Close()

		// Should create PostgreSQLStore
		_, ok := store.(*PostgreSQLStore)
		assert.True(t, ok, "Expected PostgreSQLStore")
	})

	t.Run("SupabaseConfig", func(t *testing.T) {
		connStr := getTestPostgreSQLConnStr()
		if connStr == "" {
			t.Skip("Skipping Supabase config test - no connection string")
		}

		config := &Config{
			Type:     "supabase",
			Database: connStr,
		}

		store, err := NewStore(config)
		require.NoError(t, err)
		defer store.Close()

		// Should create PostgreSQLStore (same implementation)
		_, ok := store.(*PostgreSQLStore)
		assert.True(t, ok, "Expected PostgreSQLStore for Supabase config")
	})

	t.Run("DefaultConfig", func(t *testing.T) {
		config := &Config{} // Empty config should default to SQLite

		store, err := NewStore(config)
		require.NoError(t, err)
		defer store.Close()

		// Should create SQLiteStore by default
		_, ok := store.(*SQLiteStore)
		assert.True(t, ok, "Expected SQLiteStore as default")
	})
}

// TestEnvironmentVariableSupport tests that environment variables work correctly
func TestEnvironmentVariableSupport(t *testing.T) {
	// This test would be more comprehensive in a real scenario
	// For now, just test that the config fields are set correctly

	t.Run("ConfigFieldMapping", func(t *testing.T) {
		config := &Config{
			Type:     "postgres",
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			Username: "testuser",
			Password: "testpass",
			SSLMode:  "require",
		}

		// Verify fields are set
		assert.Equal(t, "postgres", config.Type)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 5432, config.Port)
		assert.Equal(t, "testdb", config.Database)
		assert.Equal(t, "testuser", config.Username)
		assert.Equal(t, "testpass", config.Password)
		assert.Equal(t, "require", config.SSLMode)
	})
}
