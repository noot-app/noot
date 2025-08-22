package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStore(t *testing.T) {
	// Create temp database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Test migration
	err = store.Migrate()
	require.NoError(t, err)

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

		// Create consumption
		consumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "I had an apple and a banana",
			TotalCalories: 200,
			TotalProtein:  2,
			TotalFat:      0.5,
			TotalCarbs:    50,
			DietaryFiber:  8,
			TotalSodium:   2,
		}

		err = store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)
		assert.NotZero(t, consumption.ID)
		assert.False(t, consumption.CreatedAt.IsZero())

		// Get consumption by ID
		retrieved, err := store.GetConsumption(ctx, consumption.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, consumption.UserID, retrieved.UserID)
		assert.Equal(t, consumption.Transcript, retrieved.Transcript)
		assert.Equal(t, consumption.TotalCalories, retrieved.TotalCalories)
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
		consumptions := []*Consumption{
			{
				UserID:        user.ID,
				Transcript:    "Breakfast",
				TotalCalories: 150,
			},
			{
				UserID:        user.ID,
				Transcript:    "Lunch",
				TotalCalories: 300,
			},
		}

		for _, consumption := range consumptions {
			err := store.CreateConsumption(ctx, consumption)
			require.NoError(t, err)
		}

		// Get consumptions for user
		retrieved, err := store.GetConsumptionsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, retrieved, 2)

		// Consumptions should be ordered by created_at DESC
		assert.Equal(t, "Lunch", retrieved[0].Transcript)
		assert.Equal(t, "Breakfast", retrieved[1].Transcript)
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

		// Create consumptions with different timestamps
		oldConsumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "Old consumption",
			TotalCalories: 100,
		}
		err = store.CreateConsumption(ctx, oldConsumption)
		require.NoError(t, err)

		// Update the old consumption to have a past timestamp
		pastTime := time.Now().Add(-2 * time.Hour)
		_, err = store.db.Exec("UPDATE consumptions SET created_at = ? WHERE id = ?", pastTime, oldConsumption.ID)
		require.NoError(t, err)

		// Create a recent consumption
		recentConsumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "Recent consumption",
			TotalCalories: 200,
		}
		err = store.CreateConsumption(ctx, recentConsumption)
		require.NoError(t, err)

		// Get consumptions since 1 hour ago
		since := time.Now().Add(-1 * time.Hour)
		retrieved, err := store.GetConsumptionsByUserSince(ctx, user.ID, since)
		require.NoError(t, err)
		assert.Len(t, retrieved, 1)
		assert.Equal(t, "Recent consumption", retrieved[0].Transcript)
	})

	t.Run("Seed", func(t *testing.T) {
		// Reset database first
		err := store.Reset()
		require.NoError(t, err)

		// Run seed
		err = store.Seed()
		require.NoError(t, err)

		// Check that monalisa user was created
		user, err := store.GetUserBySubject(ctx, "email", "monalisa")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "monalisa@birki.io", user.Email)

		// Check that sample consumptions were created
		consumptions, err := store.GetConsumptionsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(consumptions), 3) // Should have at least 3 sample consumptions

		// Run seed again to ensure it doesn't create duplicates
		err = store.Seed()
		require.NoError(t, err)

		consumptionsAfterSecondSeed, err := store.GetConsumptionsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, len(consumptions), len(consumptionsAfterSecondSeed))
	})

	t.Run("NonExistentRecords", func(t *testing.T) {
		// Test getting non-existent user
		user, err := store.GetUser(ctx, "01JAPP9999XXXXXXXXXXXXXX") // Non-existent ULID
		require.NoError(t, err)
		assert.Nil(t, user)

		// Test getting non-existent consumption
		consumption, err := store.GetConsumption(ctx, "01JAPP9999XXXXXXXXXXXXXX") // Non-existent ULID
		require.NoError(t, err)
		assert.Nil(t, consumption)

		// Test getting user by non-existent subject
		userBySubject, err := store.GetUserBySubject(ctx, "nonexistent", "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, userBySubject)
	})

	t.Run("UpdateAndDeleteConsumption", func(t *testing.T) {
		// Create a user first
		user := &User{
			Provider: "github",
			Subject:  "updatedeleteuser",
			Email:    "updatedelete@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create consumption
		consumption := &Consumption{
			UserID:        user.ID,
			Transcript:    "I had a banana",
			TotalCalories: 105,
			TotalProtein:  1.3,
			TotalFat:      0.4,
			TotalCarbs:    27,
		}
		err = store.CreateConsumption(ctx, consumption)
		require.NoError(t, err)
		require.NotEmpty(t, consumption.ID)

		originalID := consumption.ID

		// Test UpdateConsumption - change to 2 bananas
		consumption.TotalCalories = 210
		consumption.TotalProtein = 2.6
		consumption.TotalFat = 0.8
		consumption.TotalCarbs = 54

		err = store.UpdateConsumption(ctx, consumption)
		require.NoError(t, err)

		// Verify the update
		updated, err := store.GetConsumption(ctx, originalID)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, float64(210), updated.TotalCalories)
		assert.Equal(t, 2.6, updated.TotalProtein)

		// Test updating non-existent consumption
		nonExistentConsumption := &Consumption{
			ID:            "01JAPP9999XXXXXXXXXXXXXX", // Non-existent ULID
			UserID:        user.ID,
			Transcript:    "test",
			TotalCalories: 100,
		}
		err = store.UpdateConsumption(ctx, nonExistentConsumption)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "consumption not found")

		// Test DeleteConsumption
		err = store.DeleteConsumption(ctx, originalID)
		require.NoError(t, err)

		// Verify deletion
		deleted, err := store.GetConsumption(ctx, originalID)
		require.NoError(t, err)
		assert.Nil(t, deleted)

		// Test deleting non-existent consumption
		err = store.DeleteConsumption(ctx, "01JAPP9999XXXXXXXXXXXXXX")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "consumption not found")
	})

	// DEPRECATED: ItemCacheOperations test - commenting out as methods are replaced by new Item methods
	/*
		t.Run("ItemCacheOperations", func(t *testing.T) {
			// Test getting non-existent item from cache
			item, err := store.GetItemFromCache(ctx, "apple", "generic")
			require.NoError(t, err)
			assert.Nil(t, item)

			// Create and upsert an item to cache
			now := time.Now().UTC()
			cacheItem := &ItemCache{
				NormalizedName:       "apple",
				NormalizedBrand:      "generic",
				DisplayName:          "Apple",
				DisplayBrand:         "Generic",
				CaloriesPer100g:      52,
				ProteinGPer100g:      0.3,
				TotalFatGPer100g:     0.2,
				TotalCarbsGPer100g:   14,
				DietaryFiberGPer100g: 2.4,
				SodiumMgPer100g:      1,
				VitaminCMgPer100g:    4.6,
				FetchedAt:            now,
				ExpiresAt:            now.AddDate(0, 0, 30),
			}

			err = store.UpsertItemCache(ctx, cacheItem)
			require.NoError(t, err)
			assert.NotEmpty(t, cacheItem.ID)

			// Retrieve the item from cache
			retrieved, err := store.GetItemFromCache(ctx, "apple", "generic")
			require.NoError(t, err)
			require.NotNil(t, retrieved)
			assert.Equal(t, "Apple", retrieved.DisplayName)
			assert.Equal(t, "Generic", retrieved.DisplayBrand)
			assert.Equal(t, float64(52), retrieved.CaloriesPer100g)
			assert.Equal(t, float64(0.3), retrieved.ProteinGPer100g)
			assert.Equal(t, float64(4.6), retrieved.VitaminCMgPer100g)

			// Test updating the same item (upsert existing)
			cacheItem.CaloriesPer100g = 55    // Updated calorie value
			cacheItem.VitaminCMgPer100g = 5.0 // Updated vitamin C value
			err = store.UpsertItemCache(ctx, cacheItem)
			require.NoError(t, err)

			// Retrieve updated item
			updated, err := store.GetItemFromCache(ctx, "apple", "generic")
			require.NoError(t, err)
			require.NotNil(t, updated)
			assert.Equal(t, float64(55), updated.CaloriesPer100g)
			assert.Equal(t, float64(5.0), updated.VitaminCMgPer100g)
			assert.Equal(t, retrieved.ID, updated.ID) // Same ID for update

			// Test refresh cache functionality
			refreshedItem := &ItemCache{
				DisplayName:          "Fresh Apple",
				DisplayBrand:         "Organic",
				CaloriesPer100g:      58,
				ProteinGPer100g:      0.4,
				TotalFatGPer100g:     0.1,
				TotalCarbsGPer100g:   15,
				DietaryFiberGPer100g: 2.8,
			}
			err = store.RefreshItemCache(ctx, "apple", "generic", refreshedItem)
			require.NoError(t, err)

			refreshed, err := store.GetItemFromCache(ctx, "apple", "generic")
			require.NoError(t, err)
			require.NotNil(t, refreshed)
			assert.Equal(t, "Fresh Apple", refreshed.DisplayName)
			assert.Equal(t, "Organic", refreshed.DisplayBrand)
			assert.Equal(t, float64(58), refreshed.CaloriesPer100g)
		})
	*/

	t.Run("ItemAliasOperations", func(t *testing.T) {
		// Test getting canonical name for non-existent alias
		canonical, brand, err := store.GetCanonicalName(ctx, "tomato", "")
		require.NoError(t, err)
		assert.Equal(t, "tomato", canonical) // Returns original if no alias found
		assert.Equal(t, "", brand)

		// Create an item alias
		alias := &ItemAlias{
			AliasName:      "tomato",
			AliasBrand:     "",
			CanonicalName:  "roma_tomato",
			CanonicalBrand: "fresh",
		}
		err = store.CreateItemAlias(ctx, alias)
		require.NoError(t, err)
		assert.NotEmpty(t, alias.ID)
		assert.False(t, alias.CreatedAt.IsZero())

		// Test retrieving canonical name
		canonical, brand, err = store.GetCanonicalName(ctx, "tomato", "")
		require.NoError(t, err)
		assert.Equal(t, "roma_tomato", canonical)
		assert.Equal(t, "fresh", brand)

		// Test another alias
		alias2 := &ItemAlias{
			AliasName:      "cherry_tomatoes",
			AliasBrand:     "organic",
			CanonicalName:  "cherry_tomato",
			CanonicalBrand: "organic",
		}
		err = store.CreateItemAlias(ctx, alias2)
		require.NoError(t, err)

		canonical2, brand2, err := store.GetCanonicalName(ctx, "cherry_tomatoes", "organic")
		require.NoError(t, err)
		assert.Equal(t, "cherry_tomato", canonical2)
		assert.Equal(t, "organic", brand2)
	})

	// DEPRECATED: CacheExpirationCheck test - commenting out as methods are replaced by new Item methods
	/*
		t.Run("CacheExpirationCheck", func(t *testing.T) {
			// Test nil item
			expired := store.IsItemCacheExpired(nil)
			assert.True(t, expired)

			// Test expired item
			now := time.Now().UTC()
			expiredItem := &ItemCache{
				ExpiresAt: now.Add(-1 * time.Hour), // Expired 1 hour ago
			}
			expired = store.IsItemCacheExpired(expiredItem)
			assert.True(t, expired)

			// Test valid item
			validItem := &ItemCache{
				ExpiresAt: now.Add(1 * time.Hour), // Expires in 1 hour
			}
			expired = store.IsItemCacheExpired(validItem)
			assert.False(t, expired)
		})
	*/

	t.Run("GetNutritionSummary", func(t *testing.T) {
		// Create a user
		user := &User{
			Provider: "github",
			Subject:  "testuser5",
			Email:    "test5@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create consumptions with nutrition data
		consumption1 := &Consumption{
			UserID:        user.ID,
			Transcript:    "Breakfast",
			TotalCalories: 300,
			TotalProtein:  10,
			TotalFat:      5,
			TotalCarbs:    60,
			DietaryFiber:  8,
			TotalSodium:   100,
		}
		err = store.CreateConsumption(ctx, consumption1)
		require.NoError(t, err)

		consumption2 := &Consumption{
			UserID:        user.ID,
			Transcript:    "Lunch",
			TotalCalories: 500,
			TotalProtein:  25,
			TotalFat:      20,
			TotalCarbs:    45,
			DietaryFiber:  5,
			TotalSodium:   800,
		}
		err = store.CreateConsumption(ctx, consumption2)
		require.NoError(t, err)

		// Get nutrition summary for the last 7 days
		endTime := time.Now().UTC()
		startTime := endTime.AddDate(0, 0, -7)

		summary, err := store.GetNutritionSummary(ctx, user.ID, startTime, endTime)
		require.NoError(t, err)
		require.NotNil(t, summary)

		// Check totals
		assert.Equal(t, 2, summary.ConsumptionCount)
		assert.Equal(t, float64(800), summary.TotalCalories)
		assert.Equal(t, float64(35), summary.TotalProtein)
		assert.Equal(t, float64(25), summary.TotalFat)
		assert.Equal(t, float64(105), summary.TotalCarbs)
		assert.Equal(t, float64(13), summary.DietaryFiber)
		assert.Equal(t, float64(900), summary.TotalSodium)

		// Check daily breakdown is populated
		assert.NotEmpty(t, summary.DailyBreakdown)
		assert.Equal(t, 1, len(summary.DailyBreakdown)) // All consumptions on same day

		dailySummary := summary.DailyBreakdown[0]
		assert.Equal(t, 2, dailySummary.ConsumptionCount)
		assert.Equal(t, float64(800), dailySummary.Calories)
	})

	t.Run("UserBiometricsOperations", func(t *testing.T) {
		// Create a user first
		user := &User{
			Provider: "github",
			Subject:  "biometrics_user",
			Email:    "biometrics@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Initially no biometrics should exist
		biometrics, err := store.GetUserBiometrics(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, biometrics)

		// Create biometrics
		birthDate := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		heightCm := 175.0
		weightKg := 70.5

		newBiometrics := &UserBiometrics{
			UserID:        user.ID,
			BirthDate:     &birthDate,
			Sex:           "male",
			HeightCm:      &heightCm,
			WeightKg:      &weightKg,
			ActivityLevel: "moderately_active",
		}

		err = store.UpsertUserBiometrics(ctx, newBiometrics)
		require.NoError(t, err)
		assert.NotZero(t, newBiometrics.ID)
		assert.False(t, newBiometrics.CreatedAt.IsZero())
		assert.False(t, newBiometrics.UpdatedAt.IsZero())

		// Get biometrics
		retrieved, err := store.GetUserBiometrics(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, newBiometrics.UserID, retrieved.UserID)
		assert.Equal(t, "male", retrieved.Sex)
		assert.Equal(t, birthDate, *retrieved.BirthDate)
		assert.Equal(t, heightCm, *retrieved.HeightCm)
		assert.Equal(t, weightKg, *retrieved.WeightKg)
		assert.Equal(t, "moderately_active", retrieved.ActivityLevel)

		// Update biometrics (should upsert)
		newWeight := 68.0
		retrieved.WeightKg = &newWeight
		retrieved.ActivityLevel = "very_active"

		err = store.UpsertUserBiometrics(ctx, retrieved)
		require.NoError(t, err)

		// Verify update
		updated, err := store.GetUserBiometrics(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, newWeight, *updated.WeightKg)
		assert.Equal(t, "very_active", updated.ActivityLevel)

		// Delete biometrics
		err = store.DeleteUserBiometrics(ctx, user.ID)
		require.NoError(t, err)

		// Verify deletion
		deleted, err := store.GetUserBiometrics(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, deleted)

		// Try to delete non-existent biometrics (should error)
		err = store.DeleteUserBiometrics(ctx, user.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user biometrics not found")
	})
}
