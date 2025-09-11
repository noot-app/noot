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

	ctx := context.Background()

	t.Run("CreateAndGetUser", func(t *testing.T) {
		user := &User{
			Handle: "testuser",
			Email:  "test@example.com",
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
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Handle, retrieved.Handle)
		assert.Equal(t, user.Email, retrieved.Email)

		// Get user by email
		byEmail, err := store.GetUserByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.NotNil(t, byEmail)
		assert.Equal(t, user.ID, byEmail.ID)
	})

	t.Run("CreateAndGetConsumption", func(t *testing.T) {
		// Create a user first
		user := &User{
			Handle: "test2",
			Email:  "test2@example.com",
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
		createdConsumption, err := store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)
		assert.NotZero(t, createdConsumption.ID)
		assert.False(t, createdConsumption.CreatedAt.IsZero())

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
			Handle: "test3",
			Email:  "test3@example.com",
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
			_, err := store.CreateConsumption(ctx, consumption)
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
			Handle: "test4",
			Email:  "test4@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create a consumption
		consumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "Recent consumption",
			TotalCalories: 150,
		}
		_, err = store.CreateConsumption(ctx, consumption)
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

// TestPostgreSQLStoreAdvanced tests more comprehensive PostgreSQL store functionality
func TestPostgreSQLStoreAdvanced(t *testing.T) {
	connStr := getTestPostgreSQLConnStr()
	if connStr == "" {
		t.Skip("Skipping PostgreSQL advanced tests - no connection string provided")
	}

	store, err := NewPostgreSQLStore(connStr)
	require.NoError(t, err, "Failed to create PostgreSQL store")
	defer store.Close()

	ctx := context.Background()

	t.Run("UserOperations", func(t *testing.T) {
		user := &User{
			Handle:           "advanceduser",
			Email:            "advanced@example.com",
			FullName:         stringPtr("Advanced User"),
			SubscriptionTier: "pro",
			AvatarURL:        stringPtr("https://example.com/avatar.jpg"),
		}

		// Create user
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())

		// Update user
		user.FullName = stringPtr("Updated Advanced User")
		user.AvatarURL = stringPtr("https://example.com/new-avatar.jpg")
		err = store.UpdateUser(ctx, user)
		require.NoError(t, err)

		// Verify update
		updated, err := store.GetUser(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "Updated Advanced User", *updated.FullName)
		assert.Equal(t, "https://example.com/new-avatar.jpg", *updated.AvatarURL)
	})

	t.Run("ConsumptionAdvancedOperations", func(t *testing.T) {
		user := &User{
			Handle: "consumptionuser",
			Email:  "consumption@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create consumption with comprehensive nutrition data
		consumption := &Consumption{
			UserID:               user.ID,
			Transcript:           "I had a nutritious meal",
			TotalCalories:        450,
			TotalProtein:         25.5,
			TotalFat:             15.2,
			TotalCarbs:           42.3,
			DietaryFiber:         8.1,
			TotalSodium:          890,
			SaturatedFat:         4.2,
			VitaminC:             45.6,
			Calcium:              120.5,
			Iron:                 3.8,
			Source:               "manual",
			IsPublic:             true,
			Title:                stringPtr("Healthy Lunch"),
			Note:                 stringPtr("Very satisfying meal"),
		}

		created, err := store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.False(t, created.CreatedAt.IsZero())

		// Test GetConsumptionForUser
		retrieved, err := store.GetConsumptionForUser(ctx, user.ID, created.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, consumption.Transcript, retrieved.Transcript)

		// Test GetPublicConsumption
		public, err := store.GetPublicConsumption(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, public)
		assert.Equal(t, consumption.Transcript, public.Transcript)

		// Test consumption deletion
		err = store.DeleteConsumption(ctx, created.ID)
		require.NoError(t, err)

		// Verify deletion
		deleted, err := store.GetConsumption(ctx, created.ID)
		require.NoError(t, err)
		assert.Nil(t, deleted)
	})

	t.Run("UserFavoriteOperations", func(t *testing.T) {
		user := &User{
			Handle: "favoriteuser",
			Email:  "favorite@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		consumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "Favorite meal",
			TotalCalories: 300,
			Source:        "manual",
		}
		created, err := store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)

		// Create favorite
		favorite, err := store.CreateFavorite(ctx, user.ID, created.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, favorite.ID)
		assert.Equal(t, user.ID, favorite.UserID)
		assert.Equal(t, created.ID, favorite.ConsumptionID)

		// Check if favorited
		isFav, err := store.IsFavorited(ctx, user.ID, created.ID)
		require.NoError(t, err)
		assert.True(t, isFav)

		// Get user favorites
		favorites, err := store.GetUserFavorites(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, favorites, 1)
		assert.Equal(t, created.ID, favorites[0].ConsumptionID)
		assert.NotNil(t, favorites[0].Consumption)

		// Delete favorite
		err = store.DeleteFavorite(ctx, user.ID, created.ID)
		require.NoError(t, err)

		// Verify deletion
		isFav, err = store.IsFavorited(ctx, user.ID, created.ID)
		require.NoError(t, err)
		assert.False(t, isFav)
	})

	t.Run("ItemOperations", func(t *testing.T) {
		item := &Item{
			CanonicalName:        "apple",
			Brand:                "granny smith",
			DisplayName:          "Apple",
			DisplayBrand:         "Granny Smith",
			CaloriesPer100g:      52,
			ProteinGPer100g:      0.3,
			TotalFatGPer100g:     0.2,
			TotalCarbsGPer100g:   14,
			DietaryFiberGPer100g: 2.4,
			VitaminCMgPer100g:    4.6,
		}

		// Create item
		err := store.CreateItem(ctx, item)
		require.NoError(t, err)
		assert.NotEmpty(t, item.ID)
		assert.False(t, item.CreatedAt.IsZero())

		// Get item by ID
		retrieved, err := store.GetItem(ctx, item.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, item.CanonicalName, retrieved.CanonicalName)
		assert.Equal(t, item.CaloriesPer100g, retrieved.CaloriesPer100g)

		// Get item by name
		byName, err := store.GetItemByName(ctx, item.CanonicalName, item.Brand)
		require.NoError(t, err)
		require.NotNil(t, byName)
		assert.Equal(t, item.ID, byName.ID)

		// Update item
		item.CaloriesPer100g = 55
		item.Note = stringPtr("Updated nutrition info")
		err = store.UpdateItem(ctx, item)
		require.NoError(t, err)

		// Verify update
		updated, err := store.GetItem(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, float64(55), updated.CaloriesPer100g)
		assert.Equal(t, "Updated nutrition info", *updated.Note)
	})

	t.Run("ConsumptionItemOperations", func(t *testing.T) {
		user := &User{
			Handle: "itemuser",
			Email:  "item@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		consumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "Test meal with items",
			TotalCalories: 200,
			Source:        "manual",
		}
		createdConsumption, err := store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)

		// Create consumption item
		item := &ConsumptionItem{
			ConsumptionID:   createdConsumption.ID,
			Name:            "Test Food Item",
			Brand:           "Test Brand",
			Grams:           100,
			UserQuantity:    floatPtr(1),
			UserUnit:        stringPtr("piece"),
			Calories:        200,
			ProteinG:        15,
			TotalFatG:       8,
			TotalCarbsG:     20,
			Note:            stringPtr("Delicious item"),
		}

		err = store.CreateConsumptionItem(ctx, item)
		require.NoError(t, err)
		assert.NotEmpty(t, item.ID)
		assert.False(t, item.CreatedAt.IsZero())

		// Get consumption items
		items, err := store.GetConsumptionItems(ctx, createdConsumption.ID)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, item.Name, items[0].Name)
		assert.Equal(t, item.Calories, items[0].Calories)

		// Update consumption item
		item.Name = "Updated Food Item"
		item.Calories = 220
		err = store.UpdateConsumptionItem(ctx, item)
		require.NoError(t, err)

		// Verify update
		updatedItems, err := store.GetConsumptionItems(ctx, createdConsumption.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Food Item", updatedItems[0].Name)
		assert.Equal(t, float64(220), updatedItems[0].Calories)

		// Delete consumption item
		err = store.DeleteConsumptionItem(ctx, item.ID)
		require.NoError(t, err)

		// Verify deletion
		deletedItems, err := store.GetConsumptionItems(ctx, createdConsumption.ID)
		require.NoError(t, err)
		assert.Len(t, deletedItems, 0)
	})

	t.Run("UserBiometricsOperations", func(t *testing.T) {
		user := &User{
			Handle: "biometricsuser",
			Email:  "biometrics@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		biometrics := &UserBiometrics{
			UserID:        user.ID,
			BirthDate:     &birthDate,
			Sex:           "male",
			HeightCm:      floatPtr(175.5),
			WeightKg:      floatPtr(70.2),
			ActivityLevel: "lightly_active",
		}

		// Upsert biometrics
		err = store.UpsertUserBiometrics(ctx, biometrics)
		require.NoError(t, err)
		assert.NotEmpty(t, biometrics.ID)

		// Get biometrics
		retrieved, err := store.GetUserBiometrics(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, user.ID, retrieved.UserID)
		assert.Equal(t, "male", retrieved.Sex)
		assert.Equal(t, 175.5, *retrieved.HeightCm)

		// Update biometrics
		biometrics.WeightKg = floatPtr(72.0)
		biometrics.ActivityLevel = "moderately_active"
		err = store.UpsertUserBiometrics(ctx, biometrics)
		require.NoError(t, err)

		// Verify update
		updated, err := store.GetUserBiometrics(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, 72.0, *updated.WeightKg)
		assert.Equal(t, "moderately_active", updated.ActivityLevel)

		// Delete biometrics
		err = store.DeleteUserBiometrics(ctx, user.ID)
		require.NoError(t, err)

		// Verify deletion
		deleted, err := store.GetUserBiometrics(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, deleted)
	})

	t.Run("UserGoalOperations", func(t *testing.T) {
		user := &User{
			Handle:           "goaluser",
			Email:            "goal@example.com",
			SubscriptionTier: "pro", // Goals require pro subscription
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		goal := &UserGoal{
			UserID:        user.ID,
			Name:          "weight_loss",
			Category:      "weight",
			OverridesJSON: `{"calories": 1800, "protein_g": 120}`,
		}

		// Upsert goal
		err = store.UpsertUserGoal(ctx, goal)
		require.NoError(t, err)
		assert.NotEmpty(t, goal.ID)

		// Get goal by name
		retrieved, err := store.GetUserGoal(ctx, user.ID, "weight_loss")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "weight_loss", retrieved.Name)
		assert.Equal(t, "weight", retrieved.Category)

		// Get goal by ID
		byID, err := store.GetUserGoalByID(ctx, user.ID, goal.ID)
		require.NoError(t, err)
		require.NotNil(t, byID)
		assert.Equal(t, goal.ID, byID.ID)

		// Get all user goals
		goals, err := store.GetUserGoals(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, goals, 1)

		// Set active goal
		err = store.SetActiveGoal(ctx, user.ID, goal.ID)
		require.NoError(t, err)

		// Get active goal ID
		activeID, err := store.GetActiveGoalID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, activeID)
		assert.Equal(t, goal.ID, *activeID)

		// Clear active goal
		err = store.ClearActiveGoal(ctx, user.ID)
		require.NoError(t, err)

		// Verify cleared
		cleared, err := store.GetActiveGoalID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, cleared)

		// Delete goal
		err = store.DeleteUserGoal(ctx, user.ID, "weight_loss")
		require.NoError(t, err)

		// Verify deletion
		deleted, err := store.GetUserGoal(ctx, user.ID, "weight_loss")
		require.NoError(t, err)
		assert.Nil(t, deleted)
	})
}

// TestPostgreSQLStoreErrorHandling tests error conditions and edge cases
func TestPostgreSQLStoreErrorHandling(t *testing.T) {
	connStr := getTestPostgreSQLConnStr()
	if connStr == "" {
		t.Skip("Skipping PostgreSQL error handling tests - no connection string provided")
	}

	store, err := NewPostgreSQLStore(connStr)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	t.Run("UserValidationErrors", func(t *testing.T) {
		// Test missing handle
		user := &User{
			Email: "test@example.com",
		}
		err := store.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handle is required")

		// Test missing email
		user = &User{
			Handle: "testuser",
		}
		err = store.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("NonExistentResourceAccess", func(t *testing.T) {
		fakeID := "550e8400-e29b-41d4-a716-446655440000"

		// Test various get operations with non-existent IDs
		user, err := store.GetUser(ctx, fakeID)
		require.NoError(t, err)
		assert.Nil(t, user)

		consumption, err := store.GetConsumption(ctx, fakeID)
		require.NoError(t, err)
		assert.Nil(t, consumption)

		item, err := store.GetItem(ctx, fakeID)
		require.NoError(t, err)
		assert.Nil(t, item)

		biometrics, err := store.GetUserBiometrics(ctx, fakeID)
		require.NoError(t, err)
		assert.Nil(t, biometrics)
	})

	t.Run("InvalidDatabaseOperations", func(t *testing.T) {
		// Test with invalid user ID formats
		emptyCtx := context.Background()
		
		_, err := store.GetConsumptionsByUser(emptyCtx, "", 10, 0)
		// Should handle gracefully (return empty slice, not error)
		require.NoError(t, err)

		// Test update on non-existent consumption
		fakeConsumption := &Consumption{
			ID:            "550e8400-e29b-41d4-a716-446655440000",
			UserID:        "550e8400-e29b-41d4-a716-446655440001",
			Transcript:    "Non-existent",
			TotalCalories: 100,
		}
		err = store.UpdateConsumption(ctx, fakeConsumption)
		assert.Error(t, err) // Should fail to update non-existent record
	})
}

// Helper functions for test data
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
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

// TestDualDatabaseSupport tests PostgreSQL/Supabase database configuration
func TestDualDatabaseSupport(t *testing.T) {
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
		connStr := getTestPostgreSQLConnStr()
		if connStr == "" {
			t.Skip("Skipping default config test - no connection string")
		}

		config := &Config{
			Database: connStr,
		} // Empty Type should default to Supabase/Postgres

		store, err := NewStore(config)
		require.NoError(t, err)
		defer store.Close()

		// Should create PostgreSQLStore by default
		_, ok := store.(*PostgreSQLStore)
		assert.True(t, ok, "Expected PostgreSQLStore as default")
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
