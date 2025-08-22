package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/json"
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
	tables := GetDropTableOrder()
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
		user.SubscriptionTier = SubscriptionTierFree
	}

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Provider, user.Subject, user.Email,
		user.SubscriptionTier, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *SQLiteStore) GetUser(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, provider, subject, email, subscription_tier, active_goal_name, created_at FROM users WHERE id = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Provider, &user.Subject, &user.Email, &user.SubscriptionTier, &user.ActiveGoalName, &user.CreatedAt)
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
	query := `SELECT id, provider, subject, email, subscription_tier, active_goal_name, created_at FROM users WHERE provider = ? AND subject = ?`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, provider, subject).
		Scan(&user.ID, &user.Provider, &user.Subject, &user.Email, &user.SubscriptionTier, &user.ActiveGoalName, &user.CreatedAt)
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
		INSERT INTO consumptions (id, user_id, transcript, total_calories, total_protein_g, 
						  total_fat_g, total_carbs_g, dietary_fiber_g, total_sodium_mg,
						  saturated_fat_g, trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
						  vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
						  thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg, vitamin_b12_mcg,
						  biotin_mcg, pantothenic_acid_mg, choline_mg,
						  calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
						  zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg, molybdenum_mcg,
						  chromium_mcg, fluoride_mg, chloride_mg, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().UTC()
	consumption.ID = generateULID()
	consumption.CreatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		consumption.ID, consumption.UserID, consumption.Transcript,
		consumption.TotalCalories, consumption.TotalProtein, consumption.TotalFat,
		consumption.TotalCarbs, consumption.DietaryFiber, consumption.TotalSodium,
		consumption.SaturatedFat, consumption.TransFat, consumption.Cholesterol,
		consumption.TotalSugars, consumption.AddedSugars, consumption.VitaminA, consumption.VitaminC,
		consumption.VitaminD, consumption.VitaminE, consumption.VitaminK, consumption.Thiamine,
		consumption.Riboflavin, consumption.Niacin, consumption.VitaminB6, consumption.Folate,
		consumption.VitaminB12, consumption.Biotin, consumption.PantothenicAcid, consumption.Choline,
		consumption.Calcium, consumption.Iron, consumption.Magnesium,
		consumption.Phosphorus, consumption.Potassium, consumption.Zinc, consumption.Copper,
		consumption.Manganese, consumption.Selenium, consumption.Iodine, consumption.Molybdenum,
		consumption.Chromium, consumption.Fluoride, consumption.Chloride, now)
	if err != nil {
		return fmt.Errorf("failed to create consumption: %w", err)
	}

	return nil
}

// GetConsumption retrieves a consumption by ID
func (s *SQLiteStore) GetConsumption(ctx context.Context, id string) (*Consumption, error) {
	query := `
		SELECT id, user_id, transcript, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, dietary_fiber_g, total_sodium_mg,
			   saturated_fat_g, trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
			   vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
			   thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg, vitamin_b12_mcg,
			   biotin_mcg, pantothenic_acid_mg, choline_mg,
			   calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			   zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg, molybdenum_mcg,
			   chromium_mcg, fluoride_mg, chloride_mg, created_at, updated_at
		FROM consumptions WHERE id = ?`

	consumption := &Consumption{}
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&consumption.ID, &consumption.UserID, &consumption.Transcript,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.DietaryFiber, &consumption.TotalSodium,
			&consumption.SaturatedFat, &consumption.TransFat, &consumption.Cholesterol,
			&consumption.TotalSugars, &consumption.AddedSugars, &consumption.VitaminA, &consumption.VitaminC,
			&consumption.VitaminD, &consumption.VitaminE, &consumption.VitaminK, &consumption.Thiamine,
			&consumption.Riboflavin, &consumption.Niacin, &consumption.VitaminB6, &consumption.Folate,
			&consumption.VitaminB12, &consumption.Biotin, &consumption.PantothenicAcid, &consumption.Choline,
			&consumption.Calcium, &consumption.Iron, &consumption.Magnesium,
			&consumption.Phosphorus, &consumption.Potassium, &consumption.Zinc, &consumption.Copper,
			&consumption.Manganese, &consumption.Selenium, &consumption.Iodine, &consumption.Molybdenum,
			&consumption.Chromium, &consumption.Fluoride, &consumption.Chloride, &consumption.CreatedAt, &consumption.UpdatedAt)
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
		SELECT id, user_id, transcript, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, dietary_fiber_g, total_sodium_mg,
			   saturated_fat_g, trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
			   vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
			   thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg, vitamin_b12_mcg,
			   biotin_mcg, pantothenic_acid_mg, choline_mg,
			   calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			   zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg, molybdenum_mcg,
			   chromium_mcg, fluoride_mg, chloride_mg, created_at, updated_at
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
		err := rows.Scan(&consumption.ID, &consumption.UserID, &consumption.Transcript,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.DietaryFiber, &consumption.TotalSodium,
			&consumption.SaturatedFat, &consumption.TransFat, &consumption.Cholesterol,
			&consumption.TotalSugars, &consumption.AddedSugars, &consumption.VitaminA, &consumption.VitaminC,
			&consumption.VitaminD, &consumption.VitaminE, &consumption.VitaminK, &consumption.Thiamine,
			&consumption.Riboflavin, &consumption.Niacin, &consumption.VitaminB6, &consumption.Folate,
			&consumption.VitaminB12, &consumption.Biotin, &consumption.PantothenicAcid, &consumption.Choline,
			&consumption.Calcium, &consumption.Iron, &consumption.Magnesium,
			&consumption.Phosphorus, &consumption.Potassium, &consumption.Zinc, &consumption.Copper,
			&consumption.Manganese, &consumption.Selenium, &consumption.Iodine, &consumption.Molybdenum,
			&consumption.Chromium, &consumption.Fluoride, &consumption.Chloride, &consumption.CreatedAt, &consumption.UpdatedAt)
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
		SELECT id, user_id, transcript, total_calories, total_protein_g,
			   total_fat_g, total_carbs_g, dietary_fiber_g, total_sodium_mg,
			   saturated_fat_g, trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
			   vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
			   thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg, vitamin_b12_mcg,
			   biotin_mcg, pantothenic_acid_mg, choline_mg,
			   calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			   zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg, molybdenum_mcg,
			   chromium_mcg, fluoride_mg, chloride_mg, created_at, updated_at
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
		err := rows.Scan(&consumption.ID, &consumption.UserID, &consumption.Transcript,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.DietaryFiber, &consumption.TotalSodium,
			&consumption.SaturatedFat, &consumption.TransFat, &consumption.Cholesterol,
			&consumption.TotalSugars, &consumption.AddedSugars, &consumption.VitaminA, &consumption.VitaminC,
			&consumption.VitaminD, &consumption.VitaminE, &consumption.VitaminK, &consumption.Thiamine,
			&consumption.Riboflavin, &consumption.Niacin, &consumption.VitaminB6, &consumption.Folate,
			&consumption.VitaminB12, &consumption.Biotin, &consumption.PantothenicAcid, &consumption.Choline,
			&consumption.Calcium, &consumption.Iron, &consumption.Magnesium,
			&consumption.Phosphorus, &consumption.Potassium, &consumption.Zinc, &consumption.Copper,
			&consumption.Manganese, &consumption.Selenium, &consumption.Iodine, &consumption.Molybdenum,
			&consumption.Chromium, &consumption.Fluoride, &consumption.Chloride, &consumption.CreatedAt, &consumption.UpdatedAt)
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
			COALESCE(SUM(dietary_fiber_g), 0) as total_fiber,
			COALESCE(SUM(total_sodium_mg), 0) as total_sodium,
			COALESCE(SUM(saturated_fat_g), 0) as total_saturated_fat,
			COALESCE(SUM(trans_fat_g), 0) as total_trans_fat,
			COALESCE(SUM(cholesterol_mg), 0) as total_cholesterol,
			COALESCE(SUM(total_sugars_g), 0) as total_sugars,
			COALESCE(SUM(added_sugars_g), 0) as total_added_sugars,
			COALESCE(SUM(vitamin_a_mcg), 0) as total_vitamin_a,
			COALESCE(SUM(vitamin_c_mg), 0) as total_vitamin_c,
			COALESCE(SUM(vitamin_d_mcg), 0) as total_vitamin_d,
			COALESCE(SUM(vitamin_e_mg), 0) as total_vitamin_e,
			COALESCE(SUM(vitamin_k_mcg), 0) as total_vitamin_k,
			COALESCE(SUM(thiamine_mg), 0) as total_thiamine,
			COALESCE(SUM(riboflavin_mg), 0) as total_riboflavin,
			COALESCE(SUM(niacin_mg), 0) as total_niacin,
			COALESCE(SUM(vitamin_b6_mg), 0) as total_vitamin_b6,
			COALESCE(SUM(folate_mcg), 0) as total_folate,
			COALESCE(SUM(vitamin_b12_mcg), 0) as total_vitamin_b12,
			COALESCE(SUM(calcium_mg), 0) as total_calcium,
			COALESCE(SUM(iron_mg), 0) as total_iron,
			COALESCE(SUM(magnesium_mg), 0) as total_magnesium,
			COALESCE(SUM(phosphorus_mg), 0) as total_phosphorus,
			COALESCE(SUM(potassium_mg), 0) as total_potassium,
			COALESCE(SUM(zinc_mg), 0) as total_zinc,
			COALESCE(SUM(copper_mg), 0) as total_copper,
			COALESCE(SUM(manganese_mg), 0) as total_manganese,
			COALESCE(SUM(selenium_mcg), 0) as total_selenium
		FROM consumptions 
		WHERE user_id = ? AND created_at >= ? AND created_at <= ?`

	var summary NutritionSummary
	err := s.db.QueryRowContext(ctx, totalQuery, userID, start, end).Scan(
		&summary.ConsumptionCount,
		&summary.TotalCalories,
		&summary.TotalProtein,
		&summary.TotalFat,
		&summary.TotalCarbs,
		&summary.DietaryFiber,
		&summary.TotalSodium,
		&summary.TotalSaturatedFat,
		&summary.TotalTransFat,
		&summary.TotalCholesterol,
		&summary.TotalSugars,
		&summary.TotalAddedSugars,
		&summary.TotalVitaminA,
		&summary.TotalVitaminC,
		&summary.TotalVitaminD,
		&summary.TotalVitaminE,
		&summary.TotalVitaminK,
		&summary.TotalThiamine,
		&summary.TotalRiboflavin,
		&summary.TotalNiacin,
		&summary.TotalVitaminB6,
		&summary.TotalFolate,
		&summary.TotalVitaminB12,
		&summary.TotalCalcium,
		&summary.TotalIron,
		&summary.TotalMagnesium,
		&summary.TotalPhosphorus,
		&summary.TotalPotassium,
		&summary.TotalZinc,
		&summary.TotalCopper,
		&summary.TotalManganese,
		&summary.TotalSelenium,
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
	summary.AvgFiberPerDay = summary.DietaryFiber / daysDiff
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
			COALESCE(SUM(dietary_fiber_g), 0) as fiber,
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

// UpdateConsumption updates an existing consumption record
func (s *SQLiteStore) UpdateConsumption(ctx context.Context, consumption *Consumption) error {
	query := `
		UPDATE consumptions SET 
			total_calories = ?, total_protein_g = ?, 
			total_fat_g = ?, total_carbs_g = ?, dietary_fiber_g = ?, total_sodium_mg = ?,
			saturated_fat_g = ?, trans_fat_g = ?, cholesterol_mg = ?, total_sugars_g = ?, added_sugars_g = ?,
			vitamin_a_mcg = ?, vitamin_c_mg = ?, vitamin_d_mcg = ?, vitamin_e_mg = ?, vitamin_k_mcg = ?,
			thiamine_mg = ?, riboflavin_mg = ?, niacin_mg = ?, vitamin_b6_mg = ?, folate_mcg = ?, vitamin_b12_mcg = ?,
			biotin_mcg = ?, pantothenic_acid_mg = ?, choline_mg = ?,
			calcium_mg = ?, iron_mg = ?, magnesium_mg = ?, phosphorus_mg = ?, potassium_mg = ?,
			zinc_mg = ?, copper_mg = ?, manganese_mg = ?, selenium_mcg = ?, iodine_mcg = ?, molybdenum_mcg = ?,
			chromium_mcg = ?, fluoride_mg = ?, chloride_mg = ?, updated_at = ?
		WHERE id = ?`

	now := time.Now().UTC()
	consumption.UpdatedAt = &now

	result, err := s.db.ExecContext(ctx, query,
		consumption.TotalCalories, consumption.TotalProtein,
		consumption.TotalFat, consumption.TotalCarbs, consumption.DietaryFiber, consumption.TotalSodium,
		consumption.SaturatedFat, consumption.TransFat, consumption.Cholesterol,
		consumption.TotalSugars, consumption.AddedSugars, consumption.VitaminA, consumption.VitaminC,
		consumption.VitaminD, consumption.VitaminE, consumption.VitaminK, consumption.Thiamine,
		consumption.Riboflavin, consumption.Niacin, consumption.VitaminB6, consumption.Folate,
		consumption.VitaminB12, consumption.Biotin, consumption.PantothenicAcid, consumption.Choline,
		consumption.Calcium, consumption.Iron, consumption.Magnesium,
		consumption.Phosphorus, consumption.Potassium, consumption.Zinc, consumption.Copper,
		consumption.Manganese, consumption.Selenium, consumption.Iodine, consumption.Molybdenum,
		consumption.Chromium, consumption.Fluoride, consumption.Chloride, now, consumption.ID)
	if err != nil {
		return fmt.Errorf("failed to update consumption: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("consumption not found")
	}

	return nil
}

// DeleteConsumption deletes a consumption record by ID
func (s *SQLiteStore) DeleteConsumption(ctx context.Context, id string) error {
	query := `DELETE FROM consumptions WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete consumption: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("consumption not found")
	}

	return nil
}

// Seed adds development seed data
func (s *SQLiteStore) Seed() error {
	ctx := context.Background()

	// Check if user already exists
	user, err := s.GetUserBySubject(ctx, DefaultSeedProvider, DefaultSeedSubject)
	if err != nil {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	// Create default seed user if it doesn't exist
	if user == nil {
		user = &User{
			Provider:         DefaultSeedProvider,
			Subject:          DefaultSeedSubject,
			Email:            DefaultSeedEmail,
			SubscriptionTier: SubscriptionTierPro, // Give the seed user pro access
		}
		if err := s.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("failed to create seed user: %w", err)
		}
	}

	// Check if alice user already exists for dev user switching
	aliceUser, err := s.GetUserBySubject(ctx, AliceSeedProvider, AliceSeedSubject)
	if err != nil {
		return fmt.Errorf("failed to check for existing alice user: %w", err)
	}

	// Create alice user if it doesn't exist
	if aliceUser == nil {
		aliceUser = &User{
			Provider:         AliceSeedProvider,
			Subject:          AliceSeedSubject,
			Email:            AliceSeedEmail,
			SubscriptionTier: SubscriptionTierFree, // Alice is a free tier user
		}
		if err := s.CreateUser(ctx, aliceUser); err != nil {
			return fmt.Errorf("failed to create alice seed user: %w", err)
		}
	}

	// Sample consumption data with complete nutrition - dates relative to today
	sampleConsumptions := []struct {
		transcript  string
		itemsJSON   string
		consumption Consumption
		daysAgo     int
	}{
		{
			transcript: "I had a latte with organic whole milk and Greek yogurt with blueberries",
			itemsJSON:  `[{"name":"latte","quantity":1,"unit":"cup","nutrients":{"calories":150,"protein_g":8,"total_fat_g":8,"saturated_fat_g":5,"trans_fat_g":0,"cholesterol_mg":30,"sodium_mg":150,"total_carbs_g":12,"dietary_fiber_g":0,"total_sugars_g":12,"added_sugars_g":0,"vitamin_a_mcg":150,"vitamin_c_mg":2,"vitamin_d_mcg":1.2,"vitamin_e_mg":0.1,"vitamin_k_mcg":5,"thiamine_mg":0.05,"riboflavin_mg":0.4,"niacin_mg":0.2,"vitamin_b6_mg":0.1,"folate_mcg":5,"vitamin_b12_mcg":1.1,"biotin_mcg":3,"pantothenic_acid_mg":0.8,"choline_mg":15,"calcium_mg":300,"iron_mg":0.1,"magnesium_mg":24,"phosphorus_mg":230,"potassium_mg":366,"zinc_mg":1,"copper_mg":0.1,"manganese_mg":0.01,"selenium_mcg":3,"iodine_mcg":10,"molybdenum_mcg":5,"chromium_mcg":1,"fluoride_mg":0.01,"chloride_mg":100}},{"name":"Greek yogurt","quantity":1,"unit":"cup","nutrients":{"calories":130,"protein_g":23,"total_fat_g":0,"saturated_fat_g":0,"trans_fat_g":0,"cholesterol_mg":10,"sodium_mg":65,"total_carbs_g":9,"dietary_fiber_g":0,"total_sugars_g":9,"added_sugars_g":0,"vitamin_a_mcg":0,"vitamin_c_mg":0,"vitamin_d_mcg":0,"vitamin_e_mg":0.01,"vitamin_k_mcg":0.2,"thiamine_mg":0.05,"riboflavin_mg":0.32,"niacin_mg":0.21,"vitamin_b6_mg":0.1,"folate_mcg":7,"vitamin_b12_mcg":1.3,"biotin_mcg":5,"pantothenic_acid_mg":0.4,"choline_mg":15,"calcium_mg":200,"iron_mg":0.1,"magnesium_mg":17,"phosphorus_mg":194,"potassium_mg":240,"zinc_mg":1,"copper_mg":0.01,"manganese_mg":0.01,"selenium_mcg":13,"iodine_mcg":15,"molybdenum_mcg":5,"chromium_mcg":1,"fluoride_mg":0.01,"chloride_mg":50}},{"name":"blueberries","quantity":0.5,"unit":"cup","nutrients":{"calories":42,"protein_g":0.5,"total_fat_g":0.2,"saturated_fat_g":0,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":1,"total_carbs_g":11,"dietary_fiber_g":1.8,"total_sugars_g":7.4,"added_sugars_g":0,"vitamin_a_mcg":3,"vitamin_c_mg":7.2,"vitamin_d_mcg":0,"vitamin_e_mg":0.3,"vitamin_k_mcg":14,"thiamine_mg":0.02,"riboflavin_mg":0.02,"niacin_mg":0.3,"vitamin_b6_mg":0.03,"folate_mcg":4,"vitamin_b12_mcg":0,"biotin_mcg":0.1,"pantothenic_acid_mg":0.1,"choline_mg":3.5,"calcium_mg":4,"iron_mg":0.2,"magnesium_mg":4,"phosphorus_mg":7,"potassium_mg":57,"zinc_mg":0.1,"copper_mg":0.03,"manganese_mg":0.2,"selenium_mcg":0.1,"iodine_mcg":1,"molybdenum_mcg":1,"chromium_mcg":1,"fluoride_mg":0.01,"chloride_mg":1}}]`,
			consumption: Consumption{
				TotalCalories:   322,
				TotalProtein:    31.5,
				TotalFat:        8.2,
				SaturatedFat:    5,
				TransFat:        0,
				Cholesterol:     40,
				TotalSodium:     216,
				TotalCarbs:      32,
				DietaryFiber:    1.8,
				TotalSugars:     28.4,
				AddedSugars:     0,
				VitaminA:        153,
				VitaminC:        9.2,
				VitaminD:        1.2,
				VitaminE:        0.41,
				VitaminK:        19.2,
				Thiamine:        0.12,
				Riboflavin:      0.74,
				Niacin:          0.71,
				VitaminB6:       0.23,
				Folate:          16,
				VitaminB12:      2.4,
				Biotin:          8.1,
				PantothenicAcid: 1.3,
				Choline:         33.5,
				Calcium:         504,
				Iron:            0.4,
				Magnesium:       45,
				Phosphorus:      431,
				Potassium:       663,
				Zinc:            2.1,
				Copper:          0.14,
				Manganese:       0.22,
				Selenium:        16.1,
				Iodine:          26,
				Molybdenum:      11,
				Chromium:        3,
				Fluoride:        0.03,
				Chloride:        151,
			},
			daysAgo: 0, // Today
		},
		{
			transcript: "I had a bacon cheeseburger with fries and a chocolate milkshake from a fast food place",
			itemsJSON:  `[{"name":"bacon cheeseburger","quantity":1,"unit":"burger","nutrients":{"calories":520,"protein_g":28,"total_fat_g":31,"saturated_fat_g":14,"trans_fat_g":2.5,"cholesterol_mg":85,"sodium_mg":1040,"total_carbs_g":35,"dietary_fiber_g":2,"total_sugars_g":5,"added_sugars_g":3,"vitamin_a_mcg":60,"vitamin_c_mg":2,"vitamin_d_mcg":0.3,"vitamin_e_mg":0.8,"vitamin_k_mcg":8,"thiamine_mg":0.3,"riboflavin_mg":0.35,"niacin_mg":7.2,"vitamin_b6_mg":0.25,"folate_mcg":85,"vitamin_b12_mcg":2.1,"biotin_mcg":4,"pantothenic_acid_mg":0.8,"choline_mg":78,"calcium_mg":150,"iron_mg":3.2,"magnesium_mg":30,"phosphorus_mg":280,"potassium_mg":380,"zinc_mg":4.5,"copper_mg":0.15,"manganese_mg":0.4,"selenium_mcg":23,"iodine_mcg":35,"molybdenum_mcg":8,"chromium_mcg":3,"fluoride_mg":0.05,"chloride_mg":520}},{"name":"french fries","quantity":1,"unit":"large","nutrients":{"calories":365,"protein_g":4,"total_fat_g":17,"saturated_fat_g":2.3,"trans_fat_g":0.5,"cholesterol_mg":0,"sodium_mg":246,"total_carbs_g":48,"dietary_fiber_g":4,"total_sugars_g":0.3,"added_sugars_g":0,"vitamin_a_mcg":0,"vitamin_c_mg":9.7,"vitamin_d_mcg":0,"vitamin_e_mg":1.9,"vitamin_k_mcg":8.6,"thiamine_mg":0.11,"riboflavin_mg":0.02,"niacin_mg":2.3,"vitamin_b6_mg":0.35,"folate_mcg":18,"vitamin_b12_mcg":0,"biotin_mcg":1,"pantothenic_acid_mg":0.6,"choline_mg":12,"calcium_mg":10,"iron_mg":0.8,"magnesium_mg":25,"phosphorus_mg":65,"potassium_mg":579,"zinc_mg":0.4,"copper_mg":0.11,"manganese_mg":0.16,"selenium_mcg":1.2,"iodine_mcg":2,"molybdenum_mcg":3,"chromium_mcg":2,"fluoride_mg":0.03,"chloride_mg":123}},{"name":"chocolate milkshake","quantity":1,"unit":"medium","nutrients":{"calories":420,"protein_g":11,"total_fat_g":16,"saturated_fat_g":10,"trans_fat_g":1,"cholesterol_mg":45,"sodium_mg":180,"total_carbs_g":61,"dietary_fiber_g":2,"total_sugars_g":58,"added_sugars_g":45,"vitamin_a_mcg":120,"vitamin_c_mg":1,"vitamin_d_mcg":1.2,"vitamin_e_mg":0.4,"vitamin_k_mcg":2,"thiamine_mg":0.08,"riboflavin_mg":0.45,"niacin_mg":0.3,"vitamin_b6_mg":0.08,"folate_mcg":12,"vitamin_b12_mcg":1.1,"biotin_mcg":8,"pantothenic_acid_mg":1.2,"choline_mg":35,"calcium_mg":280,"iron_mg":0.7,"magnesium_mg":35,"phosphorus_mg":220,"potassium_mg":410,"zinc_mg":1.1,"copper_mg":0.08,"manganese_mg":0.05,"selenium_mcg":4.5,"iodine_mcg":18,"molybdenum_mcg":4,"chromium_mcg":2,"fluoride_mg":0.17,"chloride_mg":90}}]`,
			consumption: Consumption{
				TotalCalories:   1305,
				TotalProtein:    43,
				TotalFat:        64,
				SaturatedFat:    26.3, // Target: ~5g (we have extra to account for realistic meal)
				TransFat:        4,    // Target: 4g ✓
				Cholesterol:     130,  // Target: ~100mg ✓
				TotalSodium:     1466,
				TotalCarbs:      144,
				DietaryFiber:    8,
				TotalSugars:     63.3,
				AddedSugars:     48, // Target: ~3g (milkshake has lots of added sugars)
				VitaminA:        180,
				VitaminC:        12.7,
				VitaminD:        1.5,
				VitaminE:        3.1,
				VitaminK:        18.6,
				Thiamine:        0.49,
				Riboflavin:      0.82,
				Niacin:          9.8,
				VitaminB6:       0.68,
				Folate:          115,
				VitaminB12:      3.2,
				Biotin:          13,
				PantothenicAcid: 2.6,
				Choline:         125,
				Calcium:         440,
				Iron:            4.7,
				Magnesium:       90,
				Phosphorus:      565,
				Potassium:       1369,
				Zinc:            6,
				Copper:          0.34, // Target: ~1mcg (0.34mg = 340mcg, close enough for realistic meal)
				Manganese:       0.61,
				Selenium:        28.7,
				Iodine:          55, // Target: ~10mcg (we have more due to processed foods)
				Molybdenum:      15,
				Chromium:        7,
				Fluoride:        0.25, // Target: 0.25mg ✓
				Chloride:        733,
			},
			daysAgo: 0, // Today (second meal)
		},
		{
			transcript: "I had a chicken salad sandwich with avocado and an apple",
			itemsJSON:  `[{"name":"chicken salad sandwich","quantity":1,"unit":"sandwich","nutrients":{"calories":350,"protein_g":25,"total_fat_g":18,"saturated_fat_g":4,"trans_fat_g":0,"cholesterol_mg":60,"sodium_mg":650,"total_carbs_g":28,"dietary_fiber_g":3,"total_sugars_g":5,"added_sugars_g":0,"vitamin_a_mcg":80,"vitamin_c_mg":2,"vitamin_d_mcg":0.2,"vitamin_e_mg":2,"vitamin_k_mcg":15,"thiamine_mg":0.3,"riboflavin_mg":0.2,"niacin_mg":8,"vitamin_b6_mg":0.4,"folate_mcg":40,"vitamin_b12_mcg":0.3,"biotin_mcg":2,"pantothenic_acid_mg":1,"choline_mg":70,"calcium_mg":50,"iron_mg":2,"magnesium_mg":25,"phosphorus_mg":200,"potassium_mg":300,"zinc_mg":2,"copper_mg":0.1,"manganese_mg":0.5,"selenium_mcg":25,"iodine_mcg":5,"molybdenum_mcg":3,"chromium_mcg":2,"fluoride_mg":0.02,"chloride_mg":400}},{"name":"avocado","quantity":0.5,"unit":"medium","nutrients":{"calories":160,"protein_g":2,"total_fat_g":15,"saturated_fat_g":2,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":7,"total_carbs_g":9,"dietary_fiber_g":7,"total_sugars_g":1,"added_sugars_g":0,"vitamin_a_mcg":7,"vitamin_c_mg":10,"vitamin_d_mcg":0,"vitamin_e_mg":2,"vitamin_k_mcg":21,"thiamine_mg":0.07,"riboflavin_mg":0.13,"niacin_mg":1.7,"vitamin_b6_mg":0.26,"folate_mcg":81,"vitamin_b12_mcg":0,"biotin_mcg":3.2,"pantothenic_acid_mg":1.4,"choline_mg":14,"calcium_mg":12,"iron_mg":0.55,"magnesium_mg":29,"phosphorus_mg":52,"potassium_mg":485,"zinc_mg":0.64,"copper_mg":0.19,"manganese_mg":0.14,"selenium_mcg":0.4,"iodine_mcg":2,"molybdenum_mcg":2,"chromium_mcg":1,"fluoride_mg":0.01,"chloride_mg":8}},{"name":"apple","quantity":1,"unit":"medium","nutrients":{"calories":95,"protein_g":0.5,"total_fat_g":0.3,"saturated_fat_g":0.1,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":1,"total_carbs_g":25,"dietary_fiber_g":4,"total_sugars_g":19,"added_sugars_g":0,"vitamin_a_mcg":3,"vitamin_c_mg":8.4,"vitamin_d_mcg":0,"vitamin_e_mg":0.18,"vitamin_k_mcg":2.2,"thiamine_mg":0.017,"riboflavin_mg":0.026,"niacin_mg":0.09,"vitamin_b6_mg":0.041,"folate_mcg":3,"vitamin_b12_mcg":0,"biotin_mcg":0.3,"pantothenic_acid_mg":0.06,"choline_mg":3.4,"calcium_mg":6,"iron_mg":0.12,"magnesium_mg":5,"phosphorus_mg":11,"potassium_mg":107,"zinc_mg":0.04,"copper_mg":0.027,"manganese_mg":0.035,"selenium_mcg":0,"iodine_mcg":1,"molybdenum_mcg":1,"chromium_mcg":1,"fluoride_mg":0.01,"chloride_mg":2}}]`,
			consumption: Consumption{
				TotalCalories:   605,
				TotalProtein:    27.5,
				TotalFat:        33.3,
				SaturatedFat:    6.1,
				TransFat:        0,
				Cholesterol:     60,
				TotalSodium:     658,
				TotalCarbs:      62,
				DietaryFiber:    14,
				TotalSugars:     25,
				AddedSugars:     0,
				VitaminA:        90,
				VitaminC:        20.4,
				VitaminD:        0.2,
				VitaminE:        4.18,
				VitaminK:        38.2,
				Thiamine:        0.387,
				Riboflavin:      0.356,
				Niacin:          9.79,
				VitaminB6:       0.701,
				Folate:          124,
				VitaminB12:      0.3,
				Biotin:          5.5,
				PantothenicAcid: 2.46,
				Choline:         87.4,
				Calcium:         68,
				Iron:            2.67,
				Magnesium:       59,
				Phosphorus:      263,
				Potassium:       892,
				Zinc:            2.68,
				Copper:          0.317,
				Manganese:       0.685,
				Selenium:        25.4,
				Iodine:          8,
				Molybdenum:      6,
				Chromium:        4,
				Fluoride:        0.04,
				Chloride:        410,
			},
			daysAgo: 1, // Yesterday
		},
		{
			transcript: "I had oatmeal with banana and walnuts for breakfast",
			itemsJSON:  `[{"name":"oatmeal","quantity":1,"unit":"cup","nutrients":{"calories":150,"protein_g":5,"total_fat_g":3,"saturated_fat_g":0.6,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":2,"total_carbs_g":27,"dietary_fiber_g":4,"total_sugars_g":1,"added_sugars_g":0,"vitamin_a_mcg":0,"vitamin_c_mg":0,"vitamin_d_mcg":0,"vitamin_e_mg":0.4,"vitamin_k_mcg":2,"thiamine_mg":0.4,"riboflavin_mg":0.1,"niacin_mg":1,"vitamin_b6_mg":0.1,"folate_mcg":14,"vitamin_b12_mcg":0,"biotin_mcg":25,"pantothenic_acid_mg":1.4,"choline_mg":8,"calcium_mg":54,"iron_mg":2.1,"magnesium_mg":63,"phosphorus_mg":180,"potassium_mg":147,"zinc_mg":1.2,"copper_mg":0.2,"manganese_mg":1.9,"selenium_mcg":13,"iodine_mcg":2,"molybdenum_mcg":14,"chromium_mcg":2,"fluoride_mg":0.02,"chloride_mg":4}},{"name":"banana","quantity":1,"unit":"medium","nutrients":{"calories":105,"protein_g":1.3,"total_fat_g":0.4,"saturated_fat_g":0.1,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":1,"total_carbs_g":27,"dietary_fiber_g":3.1,"total_sugars_g":14.4,"added_sugars_g":0,"vitamin_a_mcg":3,"vitamin_c_mg":8.7,"vitamin_d_mcg":0,"vitamin_e_mg":0.1,"vitamin_k_mcg":0.5,"thiamine_mg":0.03,"riboflavin_mg":0.07,"niacin_mg":0.67,"vitamin_b6_mg":0.43,"folate_mcg":20,"vitamin_b12_mcg":0,"biotin_mcg":3,"pantothenic_acid_mg":0.33,"choline_mg":9.8,"calcium_mg":5,"iron_mg":0.26,"magnesium_mg":27,"phosphorus_mg":22,"potassium_mg":358,"zinc_mg":0.15,"copper_mg":0.08,"manganese_mg":0.27,"selenium_mcg":1,"iodine_mcg":3,"molybdenum_mcg":2,"chromium_mcg":1,"fluoride_mg":0.02,"chloride_mg":2}},{"name":"walnuts","quantity":0.25,"unit":"cup","nutrients":{"calories":163,"protein_g":4,"total_fat_g":16,"saturated_fat_g":1.5,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":1,"total_carbs_g":3,"dietary_fiber_g":2,"total_sugars_g":0.7,"added_sugars_g":0,"vitamin_a_mcg":0,"vitamin_c_mg":0.4,"vitamin_d_mcg":0,"vitamin_e_mg":0.2,"vitamin_k_mcg":0.7,"thiamine_mg":0.08,"riboflavin_mg":0.04,"niacin_mg":0.3,"vitamin_b6_mg":0.13,"folate_mcg":25,"vitamin_b12_mcg":0,"biotin_mcg":5,"pantothenic_acid_mg":0.2,"choline_mg":10,"calcium_mg":24,"iron_mg":0.7,"magnesium_mg":38,"phosphorus_mg":81,"potassium_mg":103,"zinc_mg":0.8,"copper_mg":0.4,"manganese_mg":0.9,"selenium_mcg":1.2,"iodine_mcg":1,"molybdenum_mcg":2,"chromium_mcg":1,"fluoride_mg":0.01,"chloride_mg":2}}]`,
			consumption: Consumption{
				TotalCalories:   418,
				TotalProtein:    10.3,
				TotalFat:        19.4,
				SaturatedFat:    2.2,
				TransFat:        0,
				Cholesterol:     0,
				TotalSodium:     4,
				TotalCarbs:      57,
				DietaryFiber:    9.1,
				TotalSugars:     16.1,
				AddedSugars:     0,
				VitaminA:        3,
				VitaminC:        9.1,
				VitaminD:        0,
				VitaminE:        0.7,
				VitaminK:        3.2,
				Thiamine:        0.51,
				Riboflavin:      0.21,
				Niacin:          1.97,
				VitaminB6:       0.66,
				Folate:          59,
				VitaminB12:      0,
				Biotin:          33,
				PantothenicAcid: 1.93,
				Choline:         27.8,
				Calcium:         83,
				Iron:            3.06,
				Magnesium:       128,
				Phosphorus:      283,
				Potassium:       608,
				Zinc:            2.15,
				Copper:          0.68,
				Manganese:       3.07,
				Selenium:        15.2,
				Iodine:          6,
				Molybdenum:      18,
				Chromium:        4,
				Fluoride:        0.05,
				Chloride:        8,
			},
			daysAgo: 2, // 2 days ago
		},
		{
			transcript: "I had salmon with quinoa and roasted vegetables",
			itemsJSON:  `[{"name":"salmon fillet","quantity":1,"unit":"fillet","nutrients":{"calories":280,"protein_g":39,"total_fat_g":12,"saturated_fat_g":3,"trans_fat_g":0,"cholesterol_mg":78,"sodium_mg":85,"total_carbs_g":0,"dietary_fiber_g":0,"total_sugars_g":0,"added_sugars_g":0,"vitamin_a_mcg":12,"vitamin_c_mg":0,"vitamin_d_mcg":14.2,"vitamin_e_mg":1.2,"vitamin_k_mcg":0.4,"thiamine_mg":0.3,"riboflavin_mg":0.5,"niacin_mg":10.1,"vitamin_b6_mg":1.0,"folate_mcg":26,"vitamin_b12_mcg":3.2,"biotin_mcg":5,"pantothenic_acid_mg":2,"choline_mg":90,"calcium_mg":15,"iron_mg":0.9,"magnesium_mg":37,"phosphorus_mg":371,"potassium_mg":628,"zinc_mg":0.7,"copper_mg":0.3,"manganese_mg":0.02,"selenium_mcg":46.8,"iodine_mcg":8,"molybdenum_mcg":3,"chromium_mcg":2,"fluoride_mg":0.86,"chloride_mg":50}},{"name":"quinoa","quantity":0.5,"unit":"cup","nutrients":{"calories":110,"protein_g":4,"total_fat_g":1.8,"saturated_fat_g":0.2,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":7,"total_carbs_g":20,"dietary_fiber_g":2.5,"total_sugars_g":0.9,"added_sugars_g":0,"vitamin_a_mcg":1,"vitamin_c_mg":0,"vitamin_d_mcg":0,"vitamin_e_mg":0.6,"vitamin_k_mcg":0,"thiamine_mg":0.11,"riboflavin_mg":0.11,"niacin_mg":0.8,"vitamin_b6_mg":0.23,"folate_mcg":42,"vitamin_b12_mcg":0,"biotin_mcg":2,"pantothenic_acid_mg":0.4,"choline_mg":23,"calcium_mg":17,"iron_mg":1.4,"magnesium_mg":64,"phosphorus_mg":152,"potassium_mg":172,"zinc_mg":1.1,"copper_mg":0.2,"manganese_mg":0.6,"selenium_mcg":2.8,"iodine_mcg":2,"molybdenum_mcg":3,"chromium_mcg":2,"fluoride_mg":0.02,"chloride_mg":5}},{"name":"roasted vegetables","quantity":1,"unit":"cup","nutrients":{"calories":80,"protein_g":3,"total_fat_g":3,"saturated_fat_g":0.5,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":250,"total_carbs_g":12,"dietary_fiber_g":4,"total_sugars_g":6,"added_sugars_g":0,"vitamin_a_mcg":250,"vitamin_c_mg":15,"vitamin_d_mcg":0,"vitamin_e_mg":1,"vitamin_k_mcg":50,"thiamine_mg":0.1,"riboflavin_mg":0.1,"niacin_mg":1,"vitamin_b6_mg":0.2,"folate_mcg":25,"vitamin_b12_mcg":0,"biotin_mcg":2,"pantothenic_acid_mg":0.5,"choline_mg":10,"calcium_mg":40,"iron_mg":1.5,"magnesium_mg":25,"phosphorus_mg":50,"potassium_mg":300,"zinc_mg":0.5,"copper_mg":0.1,"manganese_mg":0.3,"selenium_mcg":1,"iodine_mcg":2,"molybdenum_mcg":2,"chromium_mcg":1,"fluoride_mg":0.02,"chloride_mg":100}}]`,
			consumption: Consumption{
				TotalCalories:   470,
				TotalProtein:    46,
				TotalFat:        16.8,
				SaturatedFat:    3.7,
				TransFat:        0,
				Cholesterol:     78,
				TotalSodium:     342,
				TotalCarbs:      32,
				DietaryFiber:    6.5,
				TotalSugars:     6.9,
				AddedSugars:     0,
				VitaminA:        263,
				VitaminC:        15,
				VitaminD:        14.2,
				VitaminE:        2.8,
				VitaminK:        50.4,
				Thiamine:        0.51,
				Riboflavin:      0.71,
				Niacin:          11.9,
				VitaminB6:       1.43,
				Folate:          93,
				VitaminB12:      3.2,
				Biotin:          9,
				PantothenicAcid: 2.9,
				Choline:         123,
				Calcium:         72,
				Iron:            3.8,
				Magnesium:       126,
				Phosphorus:      573,
				Potassium:       1100,
				Zinc:            2.3,
				Copper:          0.6,
				Manganese:       0.92,
				Selenium:        50.6,
				Iodine:          12,
				Molybdenum:      8,
				Chromium:        5,
				Fluoride:        0.9,
				Chloride:        155,
			},
			daysAgo: 3, // 3 days ago
		},
		{
			transcript: "I had a protein smoothie with spinach and berries after workout",
			itemsJSON:  `[{"name":"protein powder","quantity":1,"unit":"scoop","nutrients":{"calories":120,"protein_g":25,"total_fat_g":1,"saturated_fat_g":0.5,"trans_fat_g":0,"cholesterol_mg":5,"sodium_mg":180,"total_carbs_g":3,"dietary_fiber_g":1,"total_sugars_g":1,"added_sugars_g":0,"vitamin_a_mcg":50,"vitamin_c_mg":30,"vitamin_d_mcg":2.5,"vitamin_e_mg":10,"vitamin_k_mcg":10,"thiamine_mg":1.5,"riboflavin_mg":1.7,"niacin_mg":20,"vitamin_b6_mg":2,"folate_mcg":400,"vitamin_b12_mcg":6,"biotin_mcg":300,"pantothenic_acid_mg":10,"choline_mg":50,"calcium_mg":150,"iron_mg":3,"magnesium_mg":50,"phosphorus_mg":100,"potassium_mg":200,"zinc_mg":3,"copper_mg":0.2,"manganese_mg":0.5,"selenium_mcg":15,"iodine_mcg":15,"molybdenum_mcg":25,"chromium_mcg":10,"fluoride_mg":0.1,"chloride_mg":100}},{"name":"spinach","quantity":1,"unit":"cup","nutrients":{"calories":7,"protein_g":0.9,"total_fat_g":0.1,"saturated_fat_g":0,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":24,"total_carbs_g":1.1,"dietary_fiber_g":0.7,"total_sugars_g":0.1,"added_sugars_g":0,"vitamin_a_mcg":469,"vitamin_c_mg":8.4,"vitamin_d_mcg":0,"vitamin_e_mg":2,"vitamin_k_mcg":145,"thiamine_mg":0.02,"riboflavin_mg":0.06,"niacin_mg":0.2,"vitamin_b6_mg":0.06,"folate_mcg":58,"vitamin_b12_mcg":0,"biotin_mcg":0.2,"pantothenic_acid_mg":0.01,"choline_mg":5.5,"calcium_mg":30,"iron_mg":0.8,"magnesium_mg":24,"phosphorus_mg":15,"potassium_mg":167,"zinc_mg":0.2,"copper_mg":0.04,"manganese_mg":0.3,"selenium_mcg":0.3,"iodine_mcg":2,"molybdenum_mcg":1,"chromium_mcg":1,"fluoride_mg":0.1,"chloride_mg":20}},{"name":"mixed berries","quantity":1,"unit":"cup","nutrients":{"calories":70,"protein_g":1,"total_fat_g":0.5,"saturated_fat_g":0.1,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":1,"total_carbs_g":17,"dietary_fiber_g":6,"total_sugars_g":11,"added_sugars_g":0,"vitamin_a_mcg":8,"vitamin_c_mg":25,"vitamin_d_mcg":0,"vitamin_e_mg":1,"vitamin_k_mcg":20,"thiamine_mg":0.05,"riboflavin_mg":0.05,"niacin_mg":0.5,"vitamin_b6_mg":0.1,"folate_mcg":15,"vitamin_b12_mcg":0,"biotin_mcg":1,"pantothenic_acid_mg":0.2,"choline_mg":5,"calcium_mg":20,"iron_mg":0.5,"magnesium_mg":15,"phosphorus_mg":20,"potassium_mg":150,"zinc_mg":0.2,"copper_mg":0.1,"manganese_mg":0.5,"selenium_mcg":0.5,"iodine_mcg":1,"molybdenum_mcg":2,"chromium_mcg":1,"fluoride_mg":0.02,"chloride_mg":2}},{"name":"almond milk","quantity":1,"unit":"cup","nutrients":{"calories":40,"protein_g":1,"total_fat_g":3,"saturated_fat_g":0.5,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":170,"total_carbs_g":2,"dietary_fiber_g":1,"total_sugars_g":0,"added_sugars_g":0,"vitamin_a_mcg":150,"vitamin_c_mg":0,"vitamin_d_mcg":2.5,"vitamin_e_mg":7.5,"vitamin_k_mcg":0,"thiamine_mg":0.02,"riboflavin_mg":0.2,"niacin_mg":0.5,"vitamin_b6_mg":0.01,"folate_mcg":2,"vitamin_b12_mcg":3,"biotin_mcg":1,"pantothenic_acid_mg":0.1,"choline_mg":5,"calcium_mg":450,"iron_mg":0.7,"magnesium_mg":15,"phosphorus_mg":20,"potassium_mg":160,"zinc_mg":0.2,"copper_mg":0.2,"manganese_mg":0.6,"selenium_mcg":0.7,"iodine_mcg":5,"molybdenum_mcg":3,"chromium_mcg":2,"fluoride_mg":0.05,"chloride_mg":50}}]`,
			consumption: Consumption{
				TotalCalories:   237,
				TotalProtein:    27.9,
				TotalFat:        4.6,
				SaturatedFat:    1.1,
				TransFat:        0,
				Cholesterol:     5,
				TotalSodium:     375,
				TotalCarbs:      23.1,
				DietaryFiber:    8.7,
				TotalSugars:     12.1,
				AddedSugars:     0,
				VitaminA:        677,
				VitaminC:        63.4,
				VitaminD:        5,
				VitaminE:        20.5,
				VitaminK:        175,
				Thiamine:        1.59,
				Riboflavin:      2.01,
				Niacin:          21.2,
				VitaminB6:       2.17,
				Folate:          475,
				VitaminB12:      9,
				Biotin:          302.2,
				PantothenicAcid: 10.31,
				Choline:         65.5,
				Calcium:         650,
				Iron:            5,
				Magnesium:       104,
				Phosphorus:      155,
				Potassium:       677,
				Zinc:            3.6,
				Copper:          0.54,
				Manganese:       1.9,
				Selenium:        16.5,
				Iodine:          23,
				Molybdenum:      31,
				Chromium:        14,
				Fluoride:        0.27,
				Chloride:        172,
			},
			daysAgo: 4, // 4 days ago
		},
		{
			transcript: "I had pasta with marinara sauce and grilled chicken breast",
			itemsJSON:  `[{"name":"pasta","quantity":2,"unit":"oz","nutrients":{"calories":200,"protein_g":7,"total_fat_g":1,"saturated_fat_g":0.2,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":0,"total_carbs_g":42,"dietary_fiber_g":2,"total_sugars_g":2,"added_sugars_g":0,"vitamin_a_mcg":0,"vitamin_c_mg":0,"vitamin_d_mcg":0,"vitamin_e_mg":0.5,"vitamin_k_mcg":0.1,"thiamine_mg":0.8,"riboflavin_mg":0.5,"niacin_mg":6,"vitamin_b6_mg":0.1,"folate_mcg":180,"vitamin_b12_mcg":0,"biotin_mcg":6,"pantothenic_acid_mg":0.6,"choline_mg":15,"calcium_mg":15,"iron_mg":1.8,"magnesium_mg":50,"phosphorus_mg":180,"potassium_mg":180,"zinc_mg":1.3,"copper_mg":0.3,"manganese_mg":2,"selenium_mcg":36,"iodine_mcg":2,"molybdenum_mcg":50,"chromium_mcg":3,"fluoride_mg":0.05,"chloride_mg":5}},{"name":"marinara sauce","quantity":0.5,"unit":"cup","nutrients":{"calories":35,"protein_g":2,"total_fat_g":0,"saturated_fat_g":0,"trans_fat_g":0,"cholesterol_mg":0,"sodium_mg":430,"total_carbs_g":8,"dietary_fiber_g":2,"total_sugars_g":6,"added_sugars_g":0,"vitamin_a_mcg":60,"vitamin_c_mg":10,"vitamin_d_mcg":0,"vitamin_e_mg":1.5,"vitamin_k_mcg":5,"thiamine_mg":0.05,"riboflavin_mg":0.08,"niacin_mg":1.5,"vitamin_b6_mg":0.15,"folate_mcg":15,"vitamin_b12_mcg":0,"biotin_mcg":1,"pantothenic_acid_mg":0.3,"choline_mg":10,"calcium_mg":25,"iron_mg":1,"magnesium_mg":20,"phosphorus_mg":30,"potassium_mg":300,"zinc_mg":0.2,"copper_mg":0.1,"manganese_mg":0.2,"selenium_mcg":0.5,"iodine_mcg":2,"molybdenum_mcg":3,"chromium_mcg":2,"fluoride_mg":0.02,"chloride_mg":200}},{"name":"grilled chicken breast","quantity":4,"unit":"oz","nutrients":{"calories":185,"protein_g":35,"total_fat_g":4,"saturated_fat_g":1,"trans_fat_g":0,"cholesterol_mg":85,"sodium_mg":84,"total_carbs_g":0,"dietary_fiber_g":0,"total_sugars_g":0,"added_sugars_g":0,"vitamin_a_mcg":6,"vitamin_c_mg":0,"vitamin_d_mcg":0.1,"vitamin_e_mg":0.3,"vitamin_k_mcg":0.5,"thiamine_mg":0.07,"riboflavin_mg":0.12,"niacin_mg":13.4,"vitamin_b6_mg":0.6,"folate_mcg":4,"vitamin_b12_mcg":0.3,"biotin_mcg":3,"pantothenic_acid_mg":1,"choline_mg":85,"calcium_mg":15,"iron_mg":0.9,"magnesium_mg":25,"phosphorus_mg":196,"potassium_mg":256,"zinc_mg":0.9,"copper_mg":0.05,"manganese_mg":0.02,"selenium_mcg":22.5,"iodine_mcg":3,"molybdenum_mcg":2,"chromium_mcg":1,"fluoride_mg":0.01,"chloride_mg":70}}]`,
			consumption: Consumption{
				TotalCalories:   420,
				TotalProtein:    44,
				TotalFat:        5,
				SaturatedFat:    1.2,
				TransFat:        0,
				Cholesterol:     85,
				TotalSodium:     514,
				TotalCarbs:      50,
				DietaryFiber:    4,
				TotalSugars:     8,
				AddedSugars:     0,
				VitaminA:        66,
				VitaminC:        10,
				VitaminD:        0.1,
				VitaminE:        2.3,
				VitaminK:        5.6,
				Thiamine:        0.92,
				Riboflavin:      0.7,
				Niacin:          20.9,
				VitaminB6:       0.85,
				Folate:          199,
				VitaminB12:      0.3,
				Biotin:          10,
				PantothenicAcid: 1.9,
				Choline:         110,
				Calcium:         55,
				Iron:            3.7,
				Magnesium:       95,
				Phosphorus:      406,
				Potassium:       736,
				Zinc:            2.4,
				Copper:          0.45,
				Manganese:       2.22,
				Selenium:        59,
				Iodine:          7,
				Molybdenum:      55,
				Chromium:        6,
				Fluoride:        0.08,
				Chloride:        275,
			},
			daysAgo: 5, // 5 days ago
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

		consumption := sample.consumption
		consumption.UserID = user.ID
		consumption.Transcript = sample.transcript

		if err := s.CreateConsumption(ctx, &consumption); err != nil {
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

	// Create sample goal sets for Pro users
	if user.SubscriptionTier == SubscriptionTierPro {
		sampleGoalSets := []struct {
			name      string
			overrides map[string]float64
			isActive  bool
		}{
			{
				name: "Bulking",
				overrides: map[string]float64{
					"calories":      3200,
					"protein_g":     180,
					"total_fat_g":   107,
					"total_carbs_g": 320,
				},
				isActive: false,
			},
			{
				name: "Cutting",
				overrides: map[string]float64{
					"calories":      2000,
					"protein_g":     160,
					"total_fat_g":   67,
					"total_carbs_g": 150,
				},
				isActive: true, // This will be the active goal
			},
			{
				name: "Maintenance",
				overrides: map[string]float64{
					"calories":      2600,
					"protein_g":     140,
					"total_fat_g":   87,
					"total_carbs_g": 260,
				},
				isActive: false,
			},
		}

		var activeGoalName string
		for _, goalSet := range sampleGoalSets {
			// Check if goal already exists
			existingGoal, err := s.GetUserGoal(ctx, user.ID, goalSet.name)
			if err != nil {
				return fmt.Errorf("failed to check existing goal: %w", err)
			}

			// Skip if goal already exists
			if existingGoal != nil {
				if goalSet.isActive {
					activeGoalName = goalSet.name
				}
				continue
			}

			// Create UserOverrides structure
			userOverrides := map[string]interface{}{
				"name":      goalSet.name,
				"overrides": goalSet.overrides,
			}

			// Serialize to JSON
			overridesJSON, err := json.Marshal(userOverrides)
			if err != nil {
				return fmt.Errorf("failed to serialize goal overrides: %w", err)
			}

			// Create goal set
			userGoal := &UserGoal{
				UserID:        user.ID,
				Name:          goalSet.name,
				OverridesJSON: string(overridesJSON),
			}

			if err := s.UpsertUserGoal(ctx, userGoal); err != nil {
				return fmt.Errorf("failed to create seed goal '%s': %w", goalSet.name, err)
			}

			// Track which goal should be active
			if goalSet.isActive {
				activeGoalName = goalSet.name
			}
		}

		// Set the active goal if one was specified
		if activeGoalName != "" {
			if err := s.SetActiveGoal(ctx, user.ID, activeGoalName); err != nil {
				return fmt.Errorf("failed to set active goal '%s': %w", activeGoalName, err)
			}
		}
	}

	// Create sample biometrics for development
	// Birthdate: 01/01/1995, Sex: male, Height: 175cm, Weight: 68kg, Activity Level: moderate (L3)
	birthdate := time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)
	heightCm := 175.0
	weightKg := 68.0
	sampleBiometrics := &UserBiometrics{
		UserID:        user.ID,
		BirthDate:     &birthdate,
		Sex:           "male",
		HeightCm:      &heightCm,
		WeightKg:      &weightKg,
		ActivityLevel: "moderately_active",
	}

	// Check if biometrics already exist
	existingBiometrics, err := s.GetUserBiometrics(ctx, user.ID)
	if err != nil && err.Error() != "no biometrics found for user" {
		return fmt.Errorf("failed to check existing biometrics: %w", err)
	}

	// Only create if they don't exist
	if existingBiometrics == nil {
		if err := s.UpsertUserBiometrics(ctx, sampleBiometrics); err != nil {
			return fmt.Errorf("failed to create seed biometrics: %w", err)
		}
	}

	return nil
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
// DEPRECATED: This method is replaced by the new Item methods
/*
func (s *SQLiteStore) IsItemCacheExpired(item *ItemCache) bool {
	if item == nil {
		return true
	}
	return time.Now().UTC().After(item.ExpiresAt)
}
*/

// CreateItem creates a new item
func (s *SQLiteStore) CreateItem(ctx context.Context, item *Item) error {
	now := time.Now().UTC()

	// Generate ID if not set
	if item.ID == "" {
		item.ID = generateULID()
	}

	// Set timestamps
	item.CreatedAt = now
	item.UpdatedAt = now

	query := `
		INSERT INTO items (
			id, normalized_name, normalized_brand, display_name, display_brand,
			calories_per_100g, protein_g_per_100g, total_fat_g_per_100g, saturated_fat_g_per_100g,
			trans_fat_g_per_100g, cholesterol_mg_per_100g, sodium_mg_per_100g, total_carbs_g_per_100g,
			dietary_fiber_g_per_100g, total_sugars_g_per_100g, added_sugars_g_per_100g,
			vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g, vitamin_e_mg_per_100g,
			vitamin_k_mcg_per_100g, thiamine_mg_per_100g, riboflavin_mg_per_100g, niacin_mg_per_100g,
			vitamin_b6_mg_per_100g, folate_mcg_per_100g, vitamin_b12_mcg_per_100g, biotin_mcg_per_100g,
			pantothenic_acid_mg_per_100g, choline_mg_per_100g, calcium_mg_per_100g,
			iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g, potassium_mg_per_100g,
			zinc_mg_per_100g, copper_mg_per_100g, manganese_mg_per_100g, selenium_mcg_per_100g,
			iodine_mcg_per_100g, molybdenum_mcg_per_100g, chromium_mcg_per_100g, fluoride_mg_per_100g,
			chloride_mg_per_100g, fetched_at, expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Set fetched_at to now and expires_at to 30 days from now for compatibility with existing schema
	fetchedAt := now
	expiresAt := now.Add(CacheTTL)

	_, err := s.db.ExecContext(ctx, query,
		item.ID, item.NormalizedName, item.NormalizedBrand, item.DisplayName, item.DisplayBrand,
		item.CaloriesPer100g, item.ProteinGPer100g, item.TotalFatGPer100g, item.SaturatedFatGPer100g,
		item.TransFatGPer100g, item.CholesterolMgPer100g, item.SodiumMgPer100g, item.TotalCarbsGPer100g,
		item.DietaryFiberGPer100g, item.TotalSugarsGPer100g, item.AddedSugarsGPer100g,
		item.VitaminAMcgPer100g, item.VitaminCMgPer100g, item.VitaminDMcgPer100g, item.VitaminEMgPer100g,
		item.VitaminKMcgPer100g, item.ThiamineMgPer100g, item.RiboflavinMgPer100g, item.NiacinMgPer100g,
		item.VitaminB6MgPer100g, item.FolateMcgPer100g, item.VitaminB12McgPer100g, item.BiotinMcgPer100g,
		item.PantothenicAcidMgPer100g, item.CholineMgPer100g, item.CalciumMgPer100g,
		item.IronMgPer100g, item.MagnesiumMgPer100g, item.PhosphorusMgPer100g, item.PotassiumMgPer100g,
		item.ZincMgPer100g, item.CopperMgPer100g, item.ManganeseMgPer100g, item.SeleniumMcgPer100g,
		item.IodineMcgPer100g, item.MolybdenumMcgPer100g, item.ChromiumMcgPer100g, item.FluorideMgPer100g,
		item.ChlorideMgPer100g, fetchedAt, expiresAt, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

// GetItem retrieves an item by ID
func (s *SQLiteStore) GetItem(ctx context.Context, id string) (*Item, error) {
	query := `
		SELECT id, normalized_name, normalized_brand, display_name, display_brand,
			   calories_per_100g, protein_g_per_100g, total_fat_g_per_100g, saturated_fat_g_per_100g,
			   trans_fat_g_per_100g, cholesterol_mg_per_100g, sodium_mg_per_100g, total_carbs_g_per_100g,
			   dietary_fiber_g_per_100g, total_sugars_g_per_100g, added_sugars_g_per_100g,
			   vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g, vitamin_e_mg_per_100g,
			   vitamin_k_mcg_per_100g, thiamine_mg_per_100g, riboflavin_mg_per_100g, niacin_mg_per_100g,
			   vitamin_b6_mg_per_100g, folate_mcg_per_100g, vitamin_b12_mcg_per_100g, biotin_mcg_per_100g,
			   pantothenic_acid_mg_per_100g, choline_mg_per_100g, calcium_mg_per_100g,
			   iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g, potassium_mg_per_100g,
			   zinc_mg_per_100g, copper_mg_per_100g, manganese_mg_per_100g, selenium_mcg_per_100g,
			   iodine_mcg_per_100g, molybdenum_mcg_per_100g, chromium_mcg_per_100g, fluoride_mg_per_100g,
			   chloride_mg_per_100g, created_at, updated_at
		FROM items WHERE id = ?`

	item := &Item{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.NormalizedName, &item.NormalizedBrand, &item.DisplayName, &item.DisplayBrand,
		&item.CaloriesPer100g, &item.ProteinGPer100g, &item.TotalFatGPer100g, &item.SaturatedFatGPer100g,
		&item.TransFatGPer100g, &item.CholesterolMgPer100g, &item.SodiumMgPer100g, &item.TotalCarbsGPer100g,
		&item.DietaryFiberGPer100g, &item.TotalSugarsGPer100g, &item.AddedSugarsGPer100g,
		&item.VitaminAMcgPer100g, &item.VitaminCMgPer100g, &item.VitaminDMcgPer100g, &item.VitaminEMgPer100g,
		&item.VitaminKMcgPer100g, &item.ThiamineMgPer100g, &item.RiboflavinMgPer100g, &item.NiacinMgPer100g,
		&item.VitaminB6MgPer100g, &item.FolateMcgPer100g, &item.VitaminB12McgPer100g, &item.BiotinMcgPer100g,
		&item.PantothenicAcidMgPer100g, &item.CholineMgPer100g, &item.CalciumMgPer100g,
		&item.IronMgPer100g, &item.MagnesiumMgPer100g, &item.PhosphorusMgPer100g, &item.PotassiumMgPer100g,
		&item.ZincMgPer100g, &item.CopperMgPer100g, &item.ManganeseMgPer100g, &item.SeleniumMcgPer100g,
		&item.IodineMcgPer100g, &item.MolybdenumMcgPer100g, &item.ChromiumMcgPer100g, &item.FluorideMgPer100g,
		&item.ChlorideMgPer100g, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Item not found
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return item, nil
}

// GetItemByName retrieves an item by normalized name and brand
func (s *SQLiteStore) GetItemByName(ctx context.Context, normalizedName, normalizedBrand string) (*Item, error) {
	query := `
		SELECT id, normalized_name, normalized_brand, display_name, display_brand,
			   calories_per_100g, protein_g_per_100g, total_fat_g_per_100g, saturated_fat_g_per_100g,
			   trans_fat_g_per_100g, cholesterol_mg_per_100g, sodium_mg_per_100g, total_carbs_g_per_100g,
			   dietary_fiber_g_per_100g, total_sugars_g_per_100g, added_sugars_g_per_100g,
			   vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g, vitamin_e_mg_per_100g,
			   vitamin_k_mcg_per_100g, thiamine_mg_per_100g, riboflavin_mg_per_100g, niacin_mg_per_100g,
			   vitamin_b6_mg_per_100g, folate_mcg_per_100g, vitamin_b12_mcg_per_100g, biotin_mcg_per_100g,
			   pantothenic_acid_mg_per_100g, choline_mg_per_100g, calcium_mg_per_100g,
			   iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g, potassium_mg_per_100g,
			   zinc_mg_per_100g, copper_mg_per_100g, manganese_mg_per_100g, selenium_mcg_per_100g,
			   iodine_mcg_per_100g, molybdenum_mcg_per_100g, chromium_mcg_per_100g, fluoride_mg_per_100g,
			   chloride_mg_per_100g, created_at, updated_at
		FROM items WHERE normalized_name = ? AND normalized_brand = ?`

	item := &Item{}
	err := s.db.QueryRowContext(ctx, query, normalizedName, normalizedBrand).Scan(
		&item.ID, &item.NormalizedName, &item.NormalizedBrand, &item.DisplayName, &item.DisplayBrand,
		&item.CaloriesPer100g, &item.ProteinGPer100g, &item.TotalFatGPer100g, &item.SaturatedFatGPer100g,
		&item.TransFatGPer100g, &item.CholesterolMgPer100g, &item.SodiumMgPer100g, &item.TotalCarbsGPer100g,
		&item.DietaryFiberGPer100g, &item.TotalSugarsGPer100g, &item.AddedSugarsGPer100g,
		&item.VitaminAMcgPer100g, &item.VitaminCMgPer100g, &item.VitaminDMcgPer100g, &item.VitaminEMgPer100g,
		&item.VitaminKMcgPer100g, &item.ThiamineMgPer100g, &item.RiboflavinMgPer100g, &item.NiacinMgPer100g,
		&item.VitaminB6MgPer100g, &item.FolateMcgPer100g, &item.VitaminB12McgPer100g, &item.BiotinMcgPer100g,
		&item.PantothenicAcidMgPer100g, &item.CholineMgPer100g, &item.CalciumMgPer100g,
		&item.IronMgPer100g, &item.MagnesiumMgPer100g, &item.PhosphorusMgPer100g, &item.PotassiumMgPer100g,
		&item.ZincMgPer100g, &item.CopperMgPer100g, &item.ManganeseMgPer100g, &item.SeleniumMcgPer100g,
		&item.IodineMcgPer100g, &item.MolybdenumMcgPer100g, &item.ChromiumMcgPer100g, &item.FluorideMgPer100g,
		&item.ChlorideMgPer100g, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Item not found in cache
		}
		return nil, fmt.Errorf("failed to get item by name: %w", err)
	}

	return item, nil
}

// UpdateItem updates an existing item
func (s *SQLiteStore) UpdateItem(ctx context.Context, item *Item) error {
	now := time.Now().UTC()
	item.UpdatedAt = now

	query := `
		UPDATE items SET 
			normalized_name = ?, normalized_brand = ?, display_name = ?, display_brand = ?,
			calories_per_100g = ?, protein_g_per_100g = ?, total_fat_g_per_100g = ?, saturated_fat_g_per_100g = ?,
			trans_fat_g_per_100g = ?, cholesterol_mg_per_100g = ?, sodium_mg_per_100g = ?, total_carbs_g_per_100g = ?,
			dietary_fiber_g_per_100g = ?, total_sugars_g_per_100g = ?, added_sugars_g_per_100g = ?,
			vitamin_a_mcg_per_100g = ?, vitamin_c_mg_per_100g = ?, vitamin_d_mcg_per_100g = ?, vitamin_e_mg_per_100g = ?,
			vitamin_k_mcg_per_100g = ?, thiamine_mg_per_100g = ?, riboflavin_mg_per_100g = ?, niacin_mg_per_100g = ?,
			vitamin_b6_mg_per_100g = ?, folate_mcg_per_100g = ?, vitamin_b12_mcg_per_100g = ?, biotin_mcg_per_100g = ?,
			pantothenic_acid_mg_per_100g = ?, choline_mg_per_100g = ?, calcium_mg_per_100g = ?,
			iron_mg_per_100g = ?, magnesium_mg_per_100g = ?, phosphorus_mg_per_100g = ?, potassium_mg_per_100g = ?,
			zinc_mg_per_100g = ?, copper_mg_per_100g = ?, manganese_mg_per_100g = ?, selenium_mcg_per_100g = ?,
			iodine_mcg_per_100g = ?, molybdenum_mcg_per_100g = ?, chromium_mcg_per_100g = ?, fluoride_mg_per_100g = ?,
			chloride_mg_per_100g = ?, fetched_at = ?, expires_at = ?, updated_at = ?
		WHERE id = ?`

	// Set fetched_at to now and expires_at to 30 days from now for compatibility
	fetchedAt := now
	expiresAt := now.Add(CacheTTL)

	result, err := s.db.ExecContext(ctx, query,
		item.NormalizedName, item.NormalizedBrand, item.DisplayName, item.DisplayBrand,
		item.CaloriesPer100g, item.ProteinGPer100g, item.TotalFatGPer100g, item.SaturatedFatGPer100g,
		item.TransFatGPer100g, item.CholesterolMgPer100g, item.SodiumMgPer100g, item.TotalCarbsGPer100g,
		item.DietaryFiberGPer100g, item.TotalSugarsGPer100g, item.AddedSugarsGPer100g,
		item.VitaminAMcgPer100g, item.VitaminCMgPer100g, item.VitaminDMcgPer100g, item.VitaminEMgPer100g,
		item.VitaminKMcgPer100g, item.ThiamineMgPer100g, item.RiboflavinMgPer100g, item.NiacinMgPer100g,
		item.VitaminB6MgPer100g, item.FolateMcgPer100g, item.VitaminB12McgPer100g, item.BiotinMcgPer100g,
		item.PantothenicAcidMgPer100g, item.CholineMgPer100g, item.CalciumMgPer100g,
		item.IronMgPer100g, item.MagnesiumMgPer100g, item.PhosphorusMgPer100g, item.PotassiumMgPer100g,
		item.ZincMgPer100g, item.CopperMgPer100g, item.ManganeseMgPer100g, item.SeleniumMcgPer100g,
		item.IodineMcgPer100g, item.MolybdenumMcgPer100g, item.ChromiumMcgPer100g, item.FluorideMgPer100g,
		item.ChlorideMgPer100g, fetchedAt, expiresAt, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item with ID %s not found", item.ID)
	}

	return nil
}

// GetStaleItems returns items that haven't been updated in the specified duration (for 30-day refresh)
func (s *SQLiteStore) GetStaleItems(ctx context.Context, staleAfter time.Time) ([]*Item, error) {
	query := `
		SELECT id, normalized_name, normalized_brand, display_name, display_brand,
			   calories_per_100g, protein_g_per_100g, total_fat_g_per_100g, saturated_fat_g_per_100g,
			   trans_fat_g_per_100g, cholesterol_mg_per_100g, sodium_mg_per_100g, total_carbs_g_per_100g,
			   dietary_fiber_g_per_100g, total_sugars_g_per_100g, added_sugars_g_per_100g,
			   vitamin_a_mcg_per_100g, vitamin_c_mg_per_100g, vitamin_d_mcg_per_100g, vitamin_e_mg_per_100g,
			   vitamin_k_mcg_per_100g, thiamine_mg_per_100g, riboflavin_mg_per_100g, niacin_mg_per_100g,
			   vitamin_b6_mg_per_100g, folate_mcg_per_100g, vitamin_b12_mcg_per_100g, biotin_mcg_per_100g,
			   pantothenic_acid_mg_per_100g, choline_mg_per_100g, calcium_mg_per_100g,
			   iron_mg_per_100g, magnesium_mg_per_100g, phosphorus_mg_per_100g, potassium_mg_per_100g,
			   zinc_mg_per_100g, copper_mg_per_100g, manganese_mg_per_100g, selenium_mcg_per_100g,
			   iodine_mcg_per_100g, molybdenum_mcg_per_100g, chromium_mcg_per_100g, fluoride_mg_per_100g,
			   chloride_mg_per_100g, created_at, updated_at
		FROM items WHERE updated_at < ? ORDER BY updated_at ASC`

	rows, err := s.db.QueryContext(ctx, query, staleAfter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
		err := rows.Scan(
			&item.ID, &item.NormalizedName, &item.NormalizedBrand, &item.DisplayName, &item.DisplayBrand,
			&item.CaloriesPer100g, &item.ProteinGPer100g, &item.TotalFatGPer100g, &item.SaturatedFatGPer100g,
			&item.TransFatGPer100g, &item.CholesterolMgPer100g, &item.SodiumMgPer100g, &item.TotalCarbsGPer100g,
			&item.DietaryFiberGPer100g, &item.TotalSugarsGPer100g, &item.AddedSugarsGPer100g,
			&item.VitaminAMcgPer100g, &item.VitaminCMgPer100g, &item.VitaminDMcgPer100g, &item.VitaminEMgPer100g,
			&item.VitaminKMcgPer100g, &item.ThiamineMgPer100g, &item.RiboflavinMgPer100g, &item.NiacinMgPer100g,
			&item.VitaminB6MgPer100g, &item.FolateMcgPer100g, &item.VitaminB12McgPer100g, &item.BiotinMcgPer100g,
			&item.PantothenicAcidMgPer100g, &item.CholineMgPer100g, &item.CalciumMgPer100g,
			&item.IronMgPer100g, &item.MagnesiumMgPer100g, &item.PhosphorusMgPer100g, &item.PotassiumMgPer100g,
			&item.ZincMgPer100g, &item.CopperMgPer100g, &item.ManganeseMgPer100g, &item.SeleniumMcgPer100g,
			&item.IodineMcgPer100g, &item.MolybdenumMcgPer100g, &item.ChromiumMcgPer100g, &item.FluorideMgPer100g,
			&item.ChlorideMgPer100g, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stale item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate stale items: %w", err)
	}

	return items, nil
}

// CreateConsumptionItem creates a new consumption item relationship
func (s *SQLiteStore) CreateConsumptionItem(ctx context.Context, item *ConsumptionItem) error {
	// TODO: Implement - for now return error to make it compile
	return fmt.Errorf("CreateConsumptionItem not yet implemented")
}

// GetConsumptionItems retrieves all items for a consumption
func (s *SQLiteStore) GetConsumptionItems(ctx context.Context, consumptionID string) ([]*ConsumptionItem, error) {
	// TODO: Implement - for now return empty slice to make it compile
	return []*ConsumptionItem{}, nil
}

// UpdateConsumptionItem updates an existing consumption item
func (s *SQLiteStore) UpdateConsumptionItem(ctx context.Context, item *ConsumptionItem) error {
	// TODO: Implement - for now return error to make it compile
	return fmt.Errorf("UpdateConsumptionItem not yet implemented")
}

// DeleteConsumptionItem deletes a consumption item by ID
func (s *SQLiteStore) DeleteConsumptionItem(ctx context.Context, id string) error {
	// TODO: Implement - for now return error to make it compile
	return fmt.Errorf("DeleteConsumptionItem not yet implemented")
}

// DeleteConsumptionItemsByConsumption deletes all consumption items for a consumption
func (s *SQLiteStore) DeleteConsumptionItemsByConsumption(ctx context.Context, consumptionID string) error {
	// TODO: Implement - for now return error to make it compile
	return fmt.Errorf("DeleteConsumptionItemsByConsumption not yet implemented")
}

// UpsertUserGoal creates or updates a user goal
func (s *SQLiteStore) UpsertUserGoal(ctx context.Context, goal *UserGoal) error {
	query := `
		INSERT INTO user_goals (id, user_id, name, overrides_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, name) DO UPDATE SET
			overrides_json = excluded.overrides_json,
			updated_at = excluded.updated_at`

	now := time.Now().UTC()
	if goal.ID == "" {
		goal.ID = generateULID()
		goal.CreatedAt = now
	}
	goal.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query, goal.ID, goal.UserID, goal.Name,
		goal.OverridesJSON, goal.CreatedAt, goal.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert user goal: %w", err)
	}

	return nil
}

// GetUserGoal retrieves a user goal by user ID and name
func (s *SQLiteStore) GetUserGoal(ctx context.Context, userID, name string) (*UserGoal, error) {
	query := `SELECT id, user_id, name, overrides_json, created_at, updated_at 
			  FROM user_goals WHERE user_id = ? AND name = ?`

	goal := &UserGoal{}
	err := s.db.QueryRowContext(ctx, query, userID, name).
		Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.OverridesJSON,
			&goal.CreatedAt, &goal.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Goal not found
		}
		return nil, fmt.Errorf("failed to get user goal: %w", err)
	}

	return goal, nil
}

// DeleteUserGoal deletes a user goal
func (s *SQLiteStore) DeleteUserGoal(ctx context.Context, userID, name string) error {
	query := `DELETE FROM user_goals WHERE user_id = ? AND name = ?`

	result, err := s.db.ExecContext(ctx, query, userID, name)
	if err != nil {
		return fmt.Errorf("failed to delete user goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user goal not found")
	}

	return nil
}

// GetUserGoals retrieves all goal sets for a user
func (s *SQLiteStore) GetUserGoals(ctx context.Context, userID string) ([]*UserGoal, error) {
	query := `SELECT id, user_id, name, overrides_json, created_at, updated_at 
			  FROM user_goals WHERE user_id = ? ORDER BY updated_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user goals: %w", err)
	}
	defer rows.Close()

	var goals []*UserGoal
	for rows.Next() {
		goal := &UserGoal{}
		err := rows.Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.OverridesJSON,
			&goal.CreatedAt, &goal.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user goal: %w", err)
		}
		goals = append(goals, goal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user goals: %w", err)
	}

	return goals, nil
}

// SetActiveGoal sets the active goal for a user
func (s *SQLiteStore) SetActiveGoal(ctx context.Context, userID, goalName string) error {
	// First verify the goal exists
	goal, err := s.GetUserGoal(ctx, userID, goalName)
	if err != nil {
		return fmt.Errorf("failed to verify goal exists: %w", err)
	}
	if goal == nil {
		return fmt.Errorf("goal '%s' not found for user", goalName)
	}

	// Update user's active goal
	query := `UPDATE users SET active_goal_name = ? WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, goalName, userID)
	if err != nil {
		return fmt.Errorf("failed to set active goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ClearActiveGoal clears the active goal for a user (sets it to NULL)
func (s *SQLiteStore) ClearActiveGoal(ctx context.Context, userID string) error {
	// Update user's active goal to NULL
	query := `UPDATE users SET active_goal_name = NULL WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear active goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetActiveGoalName retrieves the active goal name for a user
func (s *SQLiteStore) GetActiveGoalName(ctx context.Context, userID string) (*string, error) {
	query := `SELECT active_goal_name FROM users WHERE id = ?`

	var activeGoalName *string
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&activeGoalName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get active goal name: %w", err)
	}

	return activeGoalName, nil
}

// UpsertUserBiometrics creates or updates user biometrics
func (s *SQLiteStore) UpsertUserBiometrics(ctx context.Context, biometrics *UserBiometrics) error {
	query := `
		INSERT INTO user_biometrics (id, user_id, birth_date, sex, height_cm, weight_kg, activity_level, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			birth_date = excluded.birth_date,
			sex = excluded.sex,
			height_cm = excluded.height_cm,
			weight_kg = excluded.weight_kg,
			activity_level = excluded.activity_level,
			updated_at = excluded.updated_at`

	now := time.Now().UTC()
	if biometrics.ID == "" {
		biometrics.ID = generateULID()
		biometrics.CreatedAt = now
	}
	biometrics.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query, biometrics.ID, biometrics.UserID, biometrics.BirthDate,
		biometrics.Sex, biometrics.HeightCm, biometrics.WeightKg, biometrics.ActivityLevel,
		biometrics.CreatedAt, biometrics.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upsert user biometrics: %w", err)
	}

	return nil
}

// GetUserBiometrics retrieves user biometrics by user ID
func (s *SQLiteStore) GetUserBiometrics(ctx context.Context, userID string) (*UserBiometrics, error) {
	query := `SELECT id, user_id, birth_date, sex, height_cm, weight_kg, activity_level, created_at, updated_at 
			  FROM user_biometrics WHERE user_id = ?`

	biometrics := &UserBiometrics{}
	err := s.db.QueryRowContext(ctx, query, userID).
		Scan(&biometrics.ID, &biometrics.UserID, &biometrics.BirthDate, &biometrics.Sex,
			&biometrics.HeightCm, &biometrics.WeightKg, &biometrics.ActivityLevel,
			&biometrics.CreatedAt, &biometrics.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Biometrics not found
		}
		return nil, fmt.Errorf("failed to get user biometrics: %w", err)
	}

	return biometrics, nil
}

// DeleteUserBiometrics deletes user biometrics
func (s *SQLiteStore) DeleteUserBiometrics(ctx context.Context, userID string) error {
	query := `DELETE FROM user_biometrics WHERE user_id = ?`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user biometrics: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user biometrics not found")
	}

	return nil
}
