package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// CacheTTL is the default cache time-to-live (30 days)
const CacheTTL = 30 * 24 * time.Hour

//go:embed migrations/*.sql
var migrationFiles embed.FS

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// generateULID generates a new ULID string
func generateULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// NewSQLiteStore creates a new SQLite-based store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath+"?_fk=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &SQLiteStore{db: db}
	return store, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Migrate applies all pending migrations
func (s *SQLiteStore) Migrate() error {
	// Get list of migration files
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Sort migration files by name
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// Apply each migration
	for _, file := range files {
		content, err := migrationFiles.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := s.db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}
	}

	return nil
}

// Reset drops all tables and re-applies migrations
func (s *SQLiteStore) Reset() error {
	// Drop tables in reverse dependency order
	tables := []string{"item_aliases", "items_cache", "meals", "users"}
	for _, table := range tables {
		if _, err := s.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Re-apply migrations
	return s.Migrate()
}

// CreateUser creates a new user
func (s *SQLiteStore) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, provider, subject, email, created_at)
		VALUES (?, ?, ?, ?, ?)`

	now := time.Now().UTC()
	user.ID = generateULID()
	user.CreatedAt = now

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Provider, user.Subject, user.Email, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *SQLiteStore) GetUser(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, provider, subject, email, created_at FROM users WHERE id = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Provider, &user.Subject, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserBySubject retrieves a user by provider and subject
func (s *SQLiteStore) GetUserBySubject(ctx context.Context, provider, subject string) (*User, error) {
	query := `SELECT id, provider, subject, email, created_at FROM users WHERE provider = ? AND subject = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, provider, subject).
		Scan(&user.ID, &user.Provider, &user.Subject, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by subject: %w", err)
	}

	return user, nil
}

// CreateMeal creates a new meal
func (s *SQLiteStore) CreateMeal(ctx context.Context, meal *Meal) error {
	query := `
		INSERT INTO meals (id, user_id, transcript, items_json, total_calories, total_protein_g, 
						  total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().UTC()
	meal.ID = generateULID()
	meal.CreatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		meal.ID, meal.UserID, meal.Transcript, meal.ItemsJSON,
		meal.TotalCalories, meal.TotalProtein, meal.TotalFat,
		meal.TotalCarbs, meal.TotalFiber, meal.TotalSodium, now)
	if err != nil {
		return fmt.Errorf("failed to create meal: %w", err)
	}

	return nil
}

// GetMeal retrieves a meal by ID
func (s *SQLiteStore) GetMeal(ctx context.Context, id string) (*Meal, error) {
	query := `
		SELECT id, user_id, transcript, items_json, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at
		FROM meals WHERE id = ?`

	meal := &Meal{}
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&meal.ID, &meal.UserID, &meal.Transcript, &meal.ItemsJSON,
			&meal.TotalCalories, &meal.TotalProtein, &meal.TotalFat,
			&meal.TotalCarbs, &meal.TotalFiber, &meal.TotalSodium, &meal.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Meal not found
		}
		return nil, fmt.Errorf("failed to get meal: %w", err)
	}

	return meal, nil
}

// GetMealsByUser retrieves meals for a user with pagination
func (s *SQLiteStore) GetMealsByUser(ctx context.Context, userID string, limit, offset int) ([]*Meal, error) {
	query := `
		SELECT id, user_id, transcript, items_json, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at
		FROM meals WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query meals: %w", err)
	}
	defer rows.Close()

	var meals []*Meal
	for rows.Next() {
		meal := &Meal{}
		err := rows.Scan(&meal.ID, &meal.UserID, &meal.Transcript, &meal.ItemsJSON,
			&meal.TotalCalories, &meal.TotalProtein, &meal.TotalFat,
			&meal.TotalCarbs, &meal.TotalFiber, &meal.TotalSodium, &meal.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meal: %w", err)
		}
		meals = append(meals, meal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating meals: %w", err)
	}

	return meals, nil
}

// GetMealsByUserSince retrieves meals for a user since a specific time
func (s *SQLiteStore) GetMealsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Meal, error) {
	query := `
		SELECT id, user_id, transcript, items_json, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at
		FROM meals WHERE user_id = ? AND created_at >= ?
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query meals since: %w", err)
	}
	defer rows.Close()

	var meals []*Meal
	for rows.Next() {
		meal := &Meal{}
		err := rows.Scan(&meal.ID, &meal.UserID, &meal.Transcript, &meal.ItemsJSON,
			&meal.TotalCalories, &meal.TotalProtein, &meal.TotalFat,
			&meal.TotalCarbs, &meal.TotalFiber, &meal.TotalSodium, &meal.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meal: %w", err)
		}
		meals = append(meals, meal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating meals: %w", err)
	}

	return meals, nil
}

// Seed adds development seed data
func (s *SQLiteStore) Seed() error {
	ctx := context.Background()

	// Check if user already exists
	user, err := s.GetUserBySubject(ctx, "email", "monalisa")
	if err != nil {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	// Create monalisa user if it doesn't exist
	if user == nil {
		user = &User{
			Provider: "email",
			Subject:  "monalisa",
			Email:    "monalisa@birki.io",
		}
		if err := s.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("failed to create seed user: %w", err)
		}
	}

	// Sample meal data with realistic nutrition
	sampleMeals := []struct {
		transcript    string
		itemsJSON     string
		totalCalories float64
		totalProtein  float64
		totalFat      float64
		totalCarbs    float64
		totalFiber    float64
		totalSodium   float64
		daysAgo       int
	}{
		{
			transcript:    "I had a latte with organic whole milk and Greek yogurt with blueberries",
			itemsJSON:     `[{"name":"latte","quantity":1,"unit":"cup","nutrients":{"calories":150,"protein_g":8,"total_fat_g":8,"total_carbs_g":12,"dietary_fiber_g":0,"sodium_mg":150}},{"name":"Greek yogurt","quantity":1,"unit":"cup","nutrients":{"calories":130,"protein_g":23,"total_fat_g":0,"total_carbs_g":9,"dietary_fiber_g":0,"sodium_mg":65}},{"name":"blueberries","quantity":0.5,"unit":"cup","nutrients":{"calories":42,"protein_g":0.5,"total_fat_g":0.2,"total_carbs_g":11,"dietary_fiber_g":1.8,"sodium_mg":1}}]`,
			totalCalories: 322,
			totalProtein:  31.5,
			totalFat:      8.2,
			totalCarbs:    32,
			totalFiber:    1.8,
			totalSodium:   216,
			daysAgo:       1,
		},
		{
			transcript:    "I had a chicken salad sandwich with avocado and an apple",
			itemsJSON:     `[{"name":"chicken salad sandwich","quantity":1,"unit":"sandwich","nutrients":{"calories":350,"protein_g":25,"total_fat_g":18,"total_carbs_g":28,"dietary_fiber_g":3,"sodium_mg":650}},{"name":"avocado","quantity":0.5,"unit":"medium","nutrients":{"calories":160,"protein_g":2,"total_fat_g":15,"total_carbs_g":9,"dietary_fiber_g":7,"sodium_mg":7}},{"name":"apple","quantity":1,"unit":"medium","nutrients":{"calories":95,"protein_g":0.5,"total_fat_g":0.3,"total_carbs_g":25,"dietary_fiber_g":4,"sodium_mg":1}}]`,
			totalCalories: 605,
			totalProtein:  27.5,
			totalFat:      33.3,
			totalCarbs:    62,
			totalFiber:    14,
			totalSodium:   658,
			daysAgo:       2,
		},
		{
			transcript:    "I had oatmeal with banana and walnuts for breakfast",
			itemsJSON:     `[{"name":"oatmeal","quantity":1,"unit":"cup","nutrients":{"calories":150,"protein_g":5,"total_fat_g":3,"total_carbs_g":27,"dietary_fiber_g":4,"sodium_mg":2}},{"name":"banana","quantity":1,"unit":"medium","nutrients":{"calories":105,"protein_g":1.3,"total_fat_g":0.4,"total_carbs_g":27,"dietary_fiber_g":3.1,"sodium_mg":1}},{"name":"walnuts","quantity":0.25,"unit":"cup","nutrients":{"calories":163,"protein_g":4,"total_fat_g":16,"total_carbs_g":3,"dietary_fiber_g":2,"sodium_mg":1}}]`,
			totalCalories: 418,
			totalProtein:  10.3,
			totalFat:      19.4,
			totalCarbs:    57,
			totalFiber:    9.1,
			totalSodium:   4,
			daysAgo:       3,
		},
	}

	// Create sample meals with different timestamps
	for _, sample := range sampleMeals {
		// Check if similar meal already exists for this user (avoid duplicates)
		existingMeals, err := s.GetMealsByUser(ctx, user.ID, 10, 0)
		if err != nil {
			return fmt.Errorf("failed to check existing meals: %w", err)
		}

		// Skip if a meal with same transcript already exists
		exists := false
		for _, existing := range existingMeals {
			if existing.Transcript == sample.transcript {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		meal := &Meal{
			UserID:        user.ID,
			Transcript:    sample.transcript,
			ItemsJSON:     sample.itemsJSON,
			TotalCalories: sample.totalCalories,
			TotalProtein:  sample.totalProtein,
			TotalFat:      sample.totalFat,
			TotalCarbs:    sample.totalCarbs,
			TotalFiber:    sample.totalFiber,
			TotalSodium:   sample.totalSodium,
		}

		if err := s.CreateMeal(ctx, meal); err != nil {
			return fmt.Errorf("failed to create seed meal: %w", err)
		}

		// Update the created_at timestamp to simulate different days
		if sample.daysAgo > 0 {
			pastTime := time.Now().AddDate(0, 0, -sample.daysAgo)
			updateQuery := `UPDATE meals SET created_at = ? WHERE id = ?`
			if _, err := s.db.Exec(updateQuery, pastTime, meal.ID); err != nil {
				return fmt.Errorf("failed to update meal timestamp: %w", err)
			}
		}
	}

	return nil
}

// GetItemFromCache retrieves an item from cache, checking if it's expired
func (s *SQLiteStore) GetItemFromCache(ctx context.Context, normalizedName, normalizedBrand string) (*ItemCache, error) {
	query := `
		SELECT id, normalized_name, normalized_brand, display_name, display_brand,
			   calories_per_100g, protein_g_per_100g, total_fat_g_per_100g, saturated_fat_g_per_100g,
			   trans_fat_g_per_100g, cholesterol_mg_per_100g, sodium_mg_per_100g, total_carbs_g_per_100g,
			   dietary_fiber_g_per_100g, total_sugars_g_per_100g, added_sugars_g_per_100g,
			   vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g, vitamin_e_mg_per_100g,
			   vitamin_k_mcg_per_100g, thiamine_mg_per_100g, riboflavin_mg_per_100g, niacin_mg_per_100g,
			   vitamin_b6_mg_per_100g, folate_mcg_per_100g, vitamin_b12_mcg_per_100g, calcium_mg_per_100g,
			   iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g, potassium_mg_per_100g,
			   zinc_mg_per_100g, copper_mg_per_100g, manganese_mg_per_100g, selenium_mcg_per_100g,
			   fetched_at, expires_at, created_at, updated_at
		FROM items_cache WHERE normalized_name = ? AND normalized_brand = ?`

	item := &ItemCache{}
	err := s.db.QueryRowContext(ctx, query, normalizedName, normalizedBrand).Scan(
		&item.ID, &item.NormalizedName, &item.NormalizedBrand, &item.DisplayName, &item.DisplayBrand,
		&item.CaloriesPer100g, &item.ProteinGPer100g, &item.TotalFatGPer100g, &item.SaturatedFatGPer100g,
		&item.TransFatGPer100g, &item.CholesterolMgPer100g, &item.SodiumMgPer100g, &item.TotalCarbsGPer100g,
		&item.DietaryFiberGPer100g, &item.TotalSugarsGPer100g, &item.AddedSugarsGPer100g,
		&item.VitaminAMcgPer100g, &item.VitaminCMgPer100g, &item.VitaminDMcgPer100g, &item.VitaminEMgPer100g,
		&item.VitaminKMcgPer100g, &item.ThiamineMgPer100g, &item.RiboflavinMgPer100g, &item.NiacinMgPer100g,
		&item.VitaminB6MgPer100g, &item.FolateMcgPer100g, &item.VitaminB12McgPer100g, &item.CalciumMgPer100g,
		&item.IronMgPer100g, &item.MagnesiumMgPer100g, &item.PhosphorusMgPer100g, &item.PotassiumMgPer100g,
		&item.ZincMgPer100g, &item.CopperMgPer100g, &item.ManganeseMgPer100g, &item.SeleniumMcgPer100g,
		&item.FetchedAt, &item.ExpiresAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Item not found in cache
		}
		return nil, fmt.Errorf("failed to get item from cache: %w", err)
	}

	return item, nil
}

// UpsertItemCache inserts or updates an item in the cache
func (s *SQLiteStore) UpsertItemCache(ctx context.Context, item *ItemCache) error {
	now := time.Now().UTC()

	// Check if item exists
	existing, err := s.GetItemFromCache(ctx, item.NormalizedName, item.NormalizedBrand)
	if err != nil {
		return fmt.Errorf("failed to check existing cache item: %w", err)
	}

	if existing == nil {
		// Insert new item
		item.ID = generateULID()
		item.CreatedAt = now
		item.UpdatedAt = now
		if item.FetchedAt.IsZero() {
			item.FetchedAt = now
		}
		if item.ExpiresAt.IsZero() {
			item.ExpiresAt = now.Add(CacheTTL)
		}

		query := `
			INSERT INTO items_cache (
				id, normalized_name, normalized_brand, display_name, display_brand,
				calories_per_100g, protein_g_per_100g, total_fat_g_per_100g, saturated_fat_g_per_100g,
				trans_fat_g_per_100g, cholesterol_mg_per_100g, sodium_mg_per_100g, total_carbs_g_per_100g,
				dietary_fiber_g_per_100g, total_sugars_g_per_100g, added_sugars_g_per_100g,
				vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g, vitamin_e_mg_per_100g,
				vitamin_k_mcg_per_100g, thiamine_mg_per_100g, riboflavin_mg_per_100g, niacin_mg_per_100g,
				vitamin_b6_mg_per_100g, folate_mcg_per_100g, vitamin_b12_mcg_per_100g, calcium_mg_per_100g,
				iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g, potassium_mg_per_100g,
				zinc_mg_per_100g, copper_mg_per_100g, manganese_mg_per_100g, selenium_mcg_per_100g,
				fetched_at, expires_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = s.db.ExecContext(ctx, query,
			item.ID, item.NormalizedName, item.NormalizedBrand, item.DisplayName, item.DisplayBrand,
			item.CaloriesPer100g, item.ProteinGPer100g, item.TotalFatGPer100g, item.SaturatedFatGPer100g,
			item.TransFatGPer100g, item.CholesterolMgPer100g, item.SodiumMgPer100g, item.TotalCarbsGPer100g,
			item.DietaryFiberGPer100g, item.TotalSugarsGPer100g, item.AddedSugarsGPer100g,
			item.VitaminAMcgPer100g, item.VitaminCMgPer100g, item.VitaminDMcgPer100g, item.VitaminEMgPer100g,
			item.VitaminKMcgPer100g, item.ThiamineMgPer100g, item.RiboflavinMgPer100g, item.NiacinMgPer100g,
			item.VitaminB6MgPer100g, item.FolateMcgPer100g, item.VitaminB12McgPer100g, item.CalciumMgPer100g,
			item.IronMgPer100g, item.MagnesiumMgPer100g, item.PhosphorusMgPer100g, item.PotassiumMgPer100g,
			item.ZincMgPer100g, item.CopperMgPer100g, item.ManganeseMgPer100g, item.SeleniumMcgPer100g,
			item.FetchedAt, item.ExpiresAt, item.CreatedAt, item.UpdatedAt,
		)
	} else {
		// Update existing item
		item.ID = existing.ID
		item.CreatedAt = existing.CreatedAt
		item.UpdatedAt = now
		if item.FetchedAt.IsZero() {
			item.FetchedAt = now
		}
		if item.ExpiresAt.IsZero() {
			item.ExpiresAt = now.Add(CacheTTL)
		}

		query := `
			UPDATE items_cache SET
				display_name = ?, display_brand = ?,
				calories_per_100g = ?, protein_g_per_100g = ?, total_fat_g_per_100g = ?, saturated_fat_g_per_100g = ?,
				trans_fat_g_per_100g = ?, cholesterol_mg_per_100g = ?, sodium_mg_per_100g = ?, total_carbs_g_per_100g = ?,
				dietary_fiber_g_per_100g = ?, total_sugars_g_per_100g = ?, added_sugars_g_per_100g = ?,
				vitamin_a_mcg_per_100g = ?, vitamin_c_mg_per_100g = ?, vitamin_d_mcg_per_100g = ?, vitamin_e_mg_per_100g = ?,
				vitamin_k_mcg_per_100g = ?, thiamine_mg_per_100g = ?, riboflavin_mg_per_100g = ?, niacin_mg_per_100g = ?,
				vitamin_b6_mg_per_100g = ?, folate_mcg_per_100g = ?, vitamin_b12_mcg_per_100g = ?, calcium_mg_per_100g = ?,
				iron_mg_per_100g = ?, magnesium_mg_per_100g = ?, phosphorus_mg_per_100g = ?, potassium_mg_per_100g = ?,
				zinc_mg_per_100g = ?, copper_mg_per_100g = ?, manganese_mg_per_100g = ?, selenium_mcg_per_100g = ?,
				fetched_at = ?, expires_at = ?, updated_at = ?
			WHERE normalized_name = ? AND normalized_brand = ?`

		_, err = s.db.ExecContext(ctx, query,
			item.DisplayName, item.DisplayBrand,
			item.CaloriesPer100g, item.ProteinGPer100g, item.TotalFatGPer100g, item.SaturatedFatGPer100g,
			item.TransFatGPer100g, item.CholesterolMgPer100g, item.SodiumMgPer100g, item.TotalCarbsGPer100g,
			item.DietaryFiberGPer100g, item.TotalSugarsGPer100g, item.AddedSugarsGPer100g,
			item.VitaminAMcgPer100g, item.VitaminCMgPer100g, item.VitaminDMcgPer100g, item.VitaminEMgPer100g,
			item.VitaminKMcgPer100g, item.ThiamineMgPer100g, item.RiboflavinMgPer100g, item.NiacinMgPer100g,
			item.VitaminB6MgPer100g, item.FolateMcgPer100g, item.VitaminB12McgPer100g, item.CalciumMgPer100g,
			item.IronMgPer100g, item.MagnesiumMgPer100g, item.PhosphorusMgPer100g, item.PotassiumMgPer100g,
			item.ZincMgPer100g, item.CopperMgPer100g, item.ManganeseMgPer100g, item.SeleniumMcgPer100g,
			item.FetchedAt, item.ExpiresAt, item.UpdatedAt,
			item.NormalizedName, item.NormalizedBrand,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to upsert item cache: %w", err)
	}

	return nil
}

// RefreshItemCache refreshes an expired cache item with new data
func (s *SQLiteStore) RefreshItemCache(ctx context.Context, normalizedName, normalizedBrand string, item *ItemCache) error {
	// This is essentially an upsert operation for expired items
	item.NormalizedName = normalizedName
	item.NormalizedBrand = normalizedBrand
	return s.UpsertItemCache(ctx, item)
}

// CreateItemAlias creates a new item alias
func (s *SQLiteStore) CreateItemAlias(ctx context.Context, alias *ItemAlias) error {
	query := `
		INSERT INTO item_aliases (id, alias_name, alias_brand, canonical_name, canonical_brand, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now().UTC()
	alias.ID = generateULID()
	alias.CreatedAt = now

	_, err := s.db.ExecContext(ctx, query, alias.ID, alias.AliasName, alias.AliasBrand,
		alias.CanonicalName, alias.CanonicalBrand, now)
	if err != nil {
		return fmt.Errorf("failed to create item alias: %w", err)
	}

	return nil
}

// GetCanonicalName retrieves canonical name and brand for an alias
func (s *SQLiteStore) GetCanonicalName(ctx context.Context, aliasName, aliasBrand string) (canonicalName, canonicalBrand string, err error) {
	query := `SELECT canonical_name, canonical_brand FROM item_aliases WHERE alias_name = ? AND alias_brand = ?`

	err = s.db.QueryRowContext(ctx, query, aliasName, aliasBrand).
		Scan(&canonicalName, &canonicalBrand)
	if err != nil {
		if err == sql.ErrNoRows {
			// No alias found, return the original name and brand
			return aliasName, aliasBrand, nil
		}
		return "", "", fmt.Errorf("failed to get canonical name: %w", err)
	}

	return canonicalName, canonicalBrand, nil
}

// IsItemCacheExpired checks if a cache item has expired
func (s *SQLiteStore) IsItemCacheExpired(item *ItemCache) bool {
	if item == nil {
		return true
	}
	return time.Now().UTC().After(item.ExpiresAt)
}
