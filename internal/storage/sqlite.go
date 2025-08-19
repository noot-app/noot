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
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rand.Reader).String()
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
		INSERT INTO users (id, provider, subject, email, subscription_tier, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now().UTC()
	user.ID = generateULID()
	user.CreatedAt = now

	// Set default subscription tier if not provided
	if user.SubscriptionTier == "" {
		user.SubscriptionTier = "free"
	}

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Provider, user.Subject, user.Email, user.SubscriptionTier, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *SQLiteStore) GetUser(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, provider, subject, email, subscription_tier, created_at FROM users WHERE id = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Provider, &user.Subject, &user.Email, &user.SubscriptionTier, &user.CreatedAt)
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
	query := `SELECT id, provider, subject, email, subscription_tier, created_at FROM users WHERE provider = ? AND subject = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, provider, subject).
		Scan(&user.ID, &user.Provider, &user.Subject, &user.Email, &user.SubscriptionTier, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by subject: %w", err)
	}

	return user, nil
}

// CreateConsumption creates a new consumption
func (s *SQLiteStore) CreateConsumption(ctx context.Context, consumption *Consumption) error {
	query := `
		INSERT INTO consumptions (id, user_id, transcript, items_json, total_calories, total_protein_g, 
						  total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().UTC()
	consumption.ID = generateULID()
	consumption.CreatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		consumption.ID, consumption.UserID, consumption.Transcript, consumption.ItemsJSON,
		consumption.TotalCalories, consumption.TotalProtein, consumption.TotalFat,
		consumption.TotalCarbs, consumption.TotalFiber, consumption.TotalSodium, now)
	if err != nil {
		return fmt.Errorf("failed to create consumption: %w", err)
	}

	return nil
}

// GetConsumption retrieves a consumption by ID
func (s *SQLiteStore) GetConsumption(ctx context.Context, id string) (*Consumption, error) {
	query := `
		SELECT id, user_id, transcript, items_json, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at
		FROM consumptions WHERE id = ?`

	consumption := &Consumption{}
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&consumption.ID, &consumption.UserID, &consumption.Transcript, &consumption.ItemsJSON,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.TotalFiber, &consumption.TotalSodium, &consumption.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Consumption not found
		}
		return nil, fmt.Errorf("failed to get consumption: %w", err)
	}

	return consumption, nil
}

// GetConsumptionsByUser retrieves consumptions for a user with pagination
func (s *SQLiteStore) GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*Consumption, error) {
	query := `
		SELECT id, user_id, transcript, items_json, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at
		FROM consumptions WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query consumptions: %w", err)
	}
	defer rows.Close()

	var consumptions []*Consumption
	for rows.Next() {
		consumption := &Consumption{}
		err := rows.Scan(&consumption.ID, &consumption.UserID, &consumption.Transcript, &consumption.ItemsJSON,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.TotalFiber, &consumption.TotalSodium, &consumption.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption: %w", err)
		}
		consumptions = append(consumptions, consumption)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating consumptions: %w", err)
	}

	return consumptions, nil
}

// GetConsumptionsByUserSince retrieves consumptions for a user since a specific time
func (s *SQLiteStore) GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Consumption, error) {
	query := `
		SELECT id, user_id, transcript, items_json, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, total_fiber_g, total_sodium_mg, created_at
		FROM consumptions WHERE user_id = ? AND created_at >= ?
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query consumptions since: %w", err)
	}
	defer rows.Close()

	var consumptions []*Consumption
	for rows.Next() {
		consumption := &Consumption{}
		err := rows.Scan(&consumption.ID, &consumption.UserID, &consumption.Transcript, &consumption.ItemsJSON,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.TotalFiber, &consumption.TotalSodium, &consumption.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption: %w", err)
		}
		consumptions = append(consumptions, consumption)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating consumptions: %w", err)
	}

	return consumptions, nil
}

// GetNutritionSummary retrieves aggregated nutrition data for a user over a time period
func (s *SQLiteStore) GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*NutritionSummary, error) {
	// First get the overall totals
	totalQuery := `
		SELECT 
			COUNT(*) as consumption_count,
			COALESCE(SUM(total_calories), 0) as total_calories,
			COALESCE(SUM(total_protein_g), 0) as total_protein,
			COALESCE(SUM(total_fat_g), 0) as total_fat,
			COALESCE(SUM(total_carbs_g), 0) as total_carbs,
			COALESCE(SUM(total_fiber_g), 0) as total_fiber,
			COALESCE(SUM(total_sodium_mg), 0) as total_sodium
		FROM consumptions 
		WHERE user_id = ? AND created_at >= ? AND created_at <= ?`

	var summary NutritionSummary
	err := s.db.QueryRowContext(ctx, totalQuery, userID, start, end).Scan(
		&summary.ConsumptionCount,
		&summary.TotalCalories,
		&summary.TotalProtein,
		&summary.TotalFat,
		&summary.TotalCarbs,
		&summary.TotalFiber,
		&summary.TotalSodium,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get nutrition summary totals: %w", err)
	}

	// Calculate number of days in the period
	daysDiff := end.Sub(start).Hours() / 24
	if daysDiff < 1 {
		daysDiff = 1 // Minimum 1 day for average calculations
	}

	// Set summary metadata
	summary.UserID = userID
	summary.StartDate = start
	summary.EndDate = end

	// Calculate daily averages
	summary.AvgCaloriesPerDay = summary.TotalCalories / daysDiff
	summary.AvgProteinPerDay = summary.TotalProtein / daysDiff
	summary.AvgFatPerDay = summary.TotalFat / daysDiff
	summary.AvgCarbsPerDay = summary.TotalCarbs / daysDiff
	summary.AvgFiberPerDay = summary.TotalFiber / daysDiff
	summary.AvgSodiumPerDay = summary.TotalSodium / daysDiff

	// Get daily breakdown for charts
	dailyQuery := `
		SELECT 
			substr(created_at, 1, 10) as date,
			COUNT(*) as consumption_count,
			COALESCE(SUM(total_calories), 0) as calories,
			COALESCE(SUM(total_protein_g), 0) as protein,
			COALESCE(SUM(total_fat_g), 0) as fat,
			COALESCE(SUM(total_carbs_g), 0) as carbs,
			COALESCE(SUM(total_fiber_g), 0) as fiber,
			COALESCE(SUM(total_sodium_mg), 0) as sodium
		FROM consumptions 
		WHERE user_id = ? AND created_at >= ? AND created_at <= ?
		GROUP BY substr(created_at, 1, 10)
		ORDER BY date ASC`

	rows, err := s.db.QueryContext(ctx, dailyQuery, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily breakdown: %w", err)
	}
	defer rows.Close()

	var dailyBreakdown []DailySummary
	for rows.Next() {
		var daily DailySummary
		var dateStr string

		err := rows.Scan(&dateStr, &daily.ConsumptionCount, &daily.Calories,
			&daily.Protein, &daily.Fat, &daily.Carbs, &daily.Fiber, &daily.Sodium)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily summary: %w", err)
		}

		// Parse the date string
		daily.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		dailyBreakdown = append(dailyBreakdown, daily)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily breakdown: %w", err)
	}

	summary.DailyBreakdown = dailyBreakdown
	return &summary, nil
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
			Provider:         "email",
			Subject:          "monalisa",
			Email:            "monalisa@birki.io",
			SubscriptionTier: "pro", // Give the seed user pro access
		}
		if err := s.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("failed to create seed user: %w", err)
		}
	}

	// Sample consumption data with realistic nutrition - dates relative to today
	sampleConsumptions := []struct {
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
			daysAgo:       0, // Today
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
			daysAgo:       1, // Yesterday
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
			daysAgo:       2, // 2 days ago
		},
		{
			transcript:    "I had salmon with quinoa and roasted vegetables",
			itemsJSON:     `[{"name":"salmon fillet","quantity":1,"unit":"fillet","nutrients":{"calories":280,"protein_g":39,"total_fat_g":12,"total_carbs_g":0,"dietary_fiber_g":0,"sodium_mg":85}},{"name":"quinoa","quantity":0.5,"unit":"cup","nutrients":{"calories":110,"protein_g":4,"total_fat_g":1.8,"total_carbs_g":20,"dietary_fiber_g":2.5,"sodium_mg":7}},{"name":"roasted vegetables","quantity":1,"unit":"cup","nutrients":{"calories":80,"protein_g":3,"total_fat_g":3,"total_carbs_g":12,"dietary_fiber_g":4,"sodium_mg":250}}]`,
			totalCalories: 470,
			totalProtein:  46,
			totalFat:      16.8,
			totalCarbs:    32,
			totalFiber:    6.5,
			totalSodium:   342,
			daysAgo:       3, // 3 days ago
		},
		{
			transcript:    "I had a protein smoothie with spinach and berries after workout",
			itemsJSON:     `[{"name":"protein powder","quantity":1,"unit":"scoop","nutrients":{"calories":120,"protein_g":25,"total_fat_g":1,"total_carbs_g":3,"dietary_fiber_g":1,"sodium_mg":180}},{"name":"spinach","quantity":1,"unit":"cup","nutrients":{"calories":7,"protein_g":0.9,"total_fat_g":0.1,"total_carbs_g":1.1,"dietary_fiber_g":0.7,"sodium_mg":24}},{"name":"mixed berries","quantity":1,"unit":"cup","nutrients":{"calories":70,"protein_g":1,"total_fat_g":0.5,"total_carbs_g":17,"dietary_fiber_g":6,"sodium_mg":1}},{"name":"almond milk","quantity":1,"unit":"cup","nutrients":{"calories":40,"protein_g":1,"total_fat_g":3,"total_carbs_g":2,"dietary_fiber_g":1,"sodium_mg":170}}]`,
			totalCalories: 237,
			totalProtein:  27.9,
			totalFat:      4.6,
			totalCarbs:    23.1,
			totalFiber:    8.7,
			totalSodium:   375,
			daysAgo:       4, // 4 days ago
		},
		{
			transcript:    "I had pasta with marinara sauce and grilled chicken breast",
			itemsJSON:     `[{"name":"pasta","quantity":2,"unit":"oz","nutrients":{"calories":200,"protein_g":7,"total_fat_g":1,"total_carbs_g":42,"dietary_fiber_g":2,"sodium_mg":0}},{"name":"marinara sauce","quantity":0.5,"unit":"cup","nutrients":{"calories":35,"protein_g":2,"total_fat_g":0,"total_carbs_g":8,"dietary_fiber_g":2,"sodium_mg":430}},{"name":"grilled chicken breast","quantity":4,"unit":"oz","nutrients":{"calories":185,"protein_g":35,"total_fat_g":4,"total_carbs_g":0,"dietary_fiber_g":0,"sodium_mg":84}}]`,
			totalCalories: 420,
			totalProtein:  44,
			totalFat:      5,
			totalCarbs:    50,
			totalFiber:    4,
			totalSodium:   514,
			daysAgo:       5, // 5 days ago
		},
	}

	// Create sample consumptions with different timestamps
	for _, sample := range sampleConsumptions {
		// Check if similar consumption already exists for this user (avoid duplicates)
		existingConsumptions, err := s.GetConsumptionsByUser(ctx, user.ID, 10, 0)
		if err != nil {
			return fmt.Errorf("failed to check existing consumptions: %w", err)
		}

		// Skip if a consumption with same transcript already exists
		exists := false
		for _, existing := range existingConsumptions {
			if existing.Transcript == sample.transcript {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		consumption := &Consumption{
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

		if err := s.CreateConsumption(ctx, consumption); err != nil {
			return fmt.Errorf("failed to create seed consumption: %w", err)
		}

		// Update the created_at timestamp to simulate different days
		if sample.daysAgo > 0 {
			pastTime := time.Now().UTC().AddDate(0, 0, -sample.daysAgo)
			updateQuery := `UPDATE consumptions SET created_at = ? WHERE id = ?`
			if _, err := s.db.Exec(updateQuery, pastTime, consumption.ID); err != nil {
				return fmt.Errorf("failed to update consumption timestamp: %w", err)
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
