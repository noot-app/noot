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

	t.Run("CreateAndGetMeal", func(t *testing.T) {
		// Create a user first
		user := &User{
			Provider: "github",
			Subject:  "testuser2",
			Email:    "test2@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create meal
		meal := &Meal{
			UserID:        user.ID,
			Transcript:    "I had an apple and a banana",
			ItemsJSON:     `[{"name":"apple","quantity":1,"unit":"medium"},{"name":"banana","quantity":1,"unit":"medium"}]`,
			TotalCalories: 200,
			TotalProtein:  2,
			TotalFat:      0.5,
			TotalCarbs:    50,
			TotalFiber:    8,
			TotalSodium:   2,
		}

		err = store.CreateMeal(ctx, meal)
		require.NoError(t, err)
		assert.NotZero(t, meal.ID)
		assert.False(t, meal.CreatedAt.IsZero())

		// Get meal by ID
		retrieved, err := store.GetMeal(ctx, meal.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, meal.UserID, retrieved.UserID)
		assert.Equal(t, meal.Transcript, retrieved.Transcript)
		assert.Equal(t, meal.ItemsJSON, retrieved.ItemsJSON)
		assert.Equal(t, meal.TotalCalories, retrieved.TotalCalories)
	})

	t.Run("GetMealsByUser", func(t *testing.T) {
		// Create a user
		user := &User{
			Provider: "github",
			Subject:  "testuser3",
			Email:    "test3@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create multiple meals
		meals := []*Meal{
			{
				UserID:        user.ID,
				Transcript:    "Breakfast",
				ItemsJSON:     `[{"name":"toast","quantity":2,"unit":"slice"}]`,
				TotalCalories: 150,
			},
			{
				UserID:        user.ID,
				Transcript:    "Lunch",
				ItemsJSON:     `[{"name":"sandwich","quantity":1,"unit":"sandwich"}]`,
				TotalCalories: 300,
			},
		}

		for _, meal := range meals {
			err := store.CreateMeal(ctx, meal)
			require.NoError(t, err)
		}

		// Get meals for user
		retrieved, err := store.GetMealsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, retrieved, 2)

		// Meals should be ordered by created_at DESC
		assert.Equal(t, "Lunch", retrieved[0].Transcript)
		assert.Equal(t, "Breakfast", retrieved[1].Transcript)
	})

	t.Run("GetMealsByUserSince", func(t *testing.T) {
		// Create a user
		user := &User{
			Provider: "github",
			Subject:  "testuser4",
			Email:    "test4@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create meals with different timestamps
		oldMeal := &Meal{
			UserID:        user.ID,
			Transcript:    "Old meal",
			ItemsJSON:     `[{"name":"old","quantity":1}]`,
			TotalCalories: 100,
		}
		err = store.CreateMeal(ctx, oldMeal)
		require.NoError(t, err)

		// Update the old meal to have a past timestamp
		pastTime := time.Now().Add(-2 * time.Hour)
		_, err = store.db.Exec("UPDATE meals SET created_at = ? WHERE id = ?", pastTime, oldMeal.ID)
		require.NoError(t, err)

		// Create a recent meal
		recentMeal := &Meal{
			UserID:        user.ID,
			Transcript:    "Recent meal",
			ItemsJSON:     `[{"name":"recent","quantity":1}]`,
			TotalCalories: 200,
		}
		err = store.CreateMeal(ctx, recentMeal)
		require.NoError(t, err)

		// Get meals since 1 hour ago
		since := time.Now().Add(-1 * time.Hour)
		retrieved, err := store.GetMealsByUserSince(ctx, user.ID, since)
		require.NoError(t, err)
		assert.Len(t, retrieved, 1)
		assert.Equal(t, "Recent meal", retrieved[0].Transcript)
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

		// Check that sample meals were created
		meals, err := store.GetMealsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(meals), 3) // Should have at least 3 sample meals

		// Run seed again to ensure it doesn't create duplicates
		err = store.Seed()
		require.NoError(t, err)

		mealsAfterSecondSeed, err := store.GetMealsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, len(meals), len(mealsAfterSecondSeed))
	})

	t.Run("NonExistentRecords", func(t *testing.T) {
		// Test getting non-existent user
		user, err := store.GetUser(ctx, "01JAPP9999XXXXXXXXXXXXXX") // Non-existent ULID
		require.NoError(t, err)
		assert.Nil(t, user)

		// Test getting non-existent meal
		meal, err := store.GetMeal(ctx, "01JAPP9999XXXXXXXXXXXXXX") // Non-existent ULID
		require.NoError(t, err)
		assert.Nil(t, meal)

		// Test getting user by non-existent subject
		userBySubject, err := store.GetUserBySubject(ctx, "nonexistent", "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, userBySubject)
	})

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

	t.Run("GetNutritionSummary", func(t *testing.T) {
		// Create a user
		user := &User{
			Provider: "github",
			Subject:  "testuser5",
			Email:    "test5@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create meals with nutrition data
		meal1 := &Meal{
			UserID:        user.ID,
			Transcript:    "Breakfast",
			ItemsJSON:     `[{"name":"oatmeal","quantity":1,"unit":"cup"}]`,
			TotalCalories: 300,
			TotalProtein:  10,
			TotalFat:      5,
			TotalCarbs:    60,
			TotalFiber:    8,
			TotalSodium:   100,
		}
		err = store.CreateMeal(ctx, meal1)
		require.NoError(t, err)

		meal2 := &Meal{
			UserID:        user.ID,
			Transcript:    "Lunch",
			ItemsJSON:     `[{"name":"sandwich","quantity":1,"unit":"sandwich"}]`,
			TotalCalories: 500,
			TotalProtein:  25,
			TotalFat:      20,
			TotalCarbs:    45,
			TotalFiber:    5,
			TotalSodium:   800,
		}
		err = store.CreateMeal(ctx, meal2)
		require.NoError(t, err)

		// Get nutrition summary for the last 7 days
		endTime := time.Now().UTC()
		startTime := endTime.AddDate(0, 0, -7)

		summary, err := store.GetNutritionSummary(ctx, user.ID, startTime, endTime)
		require.NoError(t, err)
		require.NotNil(t, summary)

		// Check totals
		assert.Equal(t, 2, summary.MealCount)
		assert.Equal(t, float64(800), summary.TotalCalories)
		assert.Equal(t, float64(35), summary.TotalProtein)
		assert.Equal(t, float64(25), summary.TotalFat)
		assert.Equal(t, float64(105), summary.TotalCarbs)
		assert.Equal(t, float64(13), summary.TotalFiber)
		assert.Equal(t, float64(900), summary.TotalSodium)

		// Check daily breakdown is populated
		assert.NotEmpty(t, summary.DailyBreakdown)
		assert.Equal(t, 1, len(summary.DailyBreakdown)) // All meals on same day

		dailySummary := summary.DailyBreakdown[0]
		assert.Equal(t, 2, dailySummary.MealCount)
		assert.Equal(t, float64(800), dailySummary.Calories)
	})
}
