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
		user, err := store.GetUserBySubject(ctx, "github", "monalisa")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "monalisa@github.com", user.Email)

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
		user, err := store.GetUser(ctx, "non-existent-ulid")
		require.NoError(t, err)
		assert.Nil(t, user)

		// Test getting non-existent meal
		meal, err := store.GetMeal(ctx, "non-existent-ulid")
		require.NoError(t, err)
		assert.Nil(t, meal)

		// Test getting user by non-existent subject
		userBySubject, err := store.GetUserBySubject(ctx, "nonexistent", "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, userBySubject)
	})

	t.Run("CacheOperations", func(t *testing.T) {
		// Test creating and retrieving cached item
		item := &CachedItem{
			NormalizedName:   "apple_organic",
			NormalizedBrand:  "generic",
			OriginalName:     "Organic Apple",
			OriginalBrand:    "Generic",
			NutrientDataJSON: `{"calories": 95, "protein_g": 0.5}`,
		}

		// Create cached item
		err := store.UpsertCachedItem(ctx, item)
		require.NoError(t, err)
		assert.NotZero(t, item.ID)
		assert.False(t, item.FetchedAt.IsZero())
		assert.False(t, item.ExpiresAt.IsZero())
		assert.True(t, item.ExpiresAt.After(item.FetchedAt))

		// Get cached item
		retrieved, err := store.GetCachedItem(ctx, "apple_organic", "generic")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, item.NormalizedName, retrieved.NormalizedName)
		assert.Equal(t, item.NormalizedBrand, retrieved.NormalizedBrand)
		assert.Equal(t, item.OriginalName, retrieved.OriginalName)
		assert.Equal(t, item.NutrientDataJSON, retrieved.NutrientDataJSON)

		// Update cached item (upsert)
		item.OriginalName = "Updated Organic Apple"
		item.NutrientDataJSON = `{"calories": 100, "protein_g": 0.6}`
		err = store.UpsertCachedItem(ctx, item)
		require.NoError(t, err)

		// Verify update
		updated, err := store.GetCachedItem(ctx, "apple_organic", "generic")
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "Updated Organic Apple", updated.OriginalName)
		assert.Equal(t, `{"calories": 100, "protein_g": 0.6}`, updated.NutrientDataJSON)

		// Test non-existent cache item
		notFound, err := store.GetCachedItem(ctx, "nonexistent", "brand")
		require.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("DeleteUserCascade", func(t *testing.T) {
		// Create a user
		user := &User{
			Provider: "github",
			Subject:  "deleteme",
			Email:    "delete@example.com",
		}
		err := store.CreateUser(ctx, user)
		require.NoError(t, err)

		// Create some meals for the user
		meal1 := &Meal{
			UserID:        user.ID,
			Transcript:    "Meal 1",
			ItemsJSON:     `[{"name":"item1"}]`,
			TotalCalories: 100,
		}
		meal2 := &Meal{
			UserID:        user.ID,
			Transcript:    "Meal 2",
			ItemsJSON:     `[{"name":"item2"}]`,
			TotalCalories: 200,
		}
		err = store.CreateMeal(ctx, meal1)
		require.NoError(t, err)
		err = store.CreateMeal(ctx, meal2)
		require.NoError(t, err)

		// Verify user and meals exist
		retrievedUser, err := store.GetUser(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedUser)

		meals, err := store.GetMealsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, meals, 2)

		// Delete user cascade
		err = store.DeleteUserCascade(ctx, user.ID)
		require.NoError(t, err)

		// Verify user is deleted
		deletedUser, err := store.GetUser(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedUser)

		// Verify meals are deleted
		mealsAfterDelete, err := store.GetMealsByUser(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, mealsAfterDelete)

		// Test deleting non-existent user
		err = store.DeleteUserCascade(ctx, "non-existent-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}
