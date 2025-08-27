package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var postgresMigrationFiles embed.FS

// PostgreSQLStore implements the Store interface using PostgreSQL
type PostgreSQLStore struct {
	db *sql.DB
}

// generateUUID generates a new UUID string
func generateUUID() string {
	return uuid.New().String()
}

// NewPostgreSQLStore creates a new PostgreSQL-based store
func NewPostgreSQLStore(connStr string) (*PostgreSQLStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQLStore{db: db}, nil
}

// Close closes the database connection
func (s *PostgreSQLStore) Close() error {
	return s.db.Close()
}

// Reset drops all tables and re-applies migrations
// CreateUser creates a new user
func (s *PostgreSQLStore) CreateUser(ctx context.Context, user *User) error {
	// For Supabase, the ID should already be provided as auth.users.id
	// For development, generate UUID if not provided
	if user.ID == "" {
		user.ID = generateUUID()
	}

	// Validate required fields
	if user.Handle == "" {
		return fmt.Errorf("user handle is required")
	}
	if user.Email == "" {
		return fmt.Errorf("user email is required")
	}

	// Set created_at if not provided
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	// Set default subscription tier if not provided
	if user.SubscriptionTier == "" {
		user.SubscriptionTier = SubscriptionTierFree
	}

	query := `
		INSERT INTO users (id, email, handle, full_name, subscription_tier, avatar_url, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Email, user.Handle, user.FullName, user.SubscriptionTier, user.AvatarURL, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *PostgreSQLStore) GetUser(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, handle, full_name, email, subscription_tier, active_goal_name, avatar_url, created_at
		FROM users WHERE id = $1`

	var user User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Handle, &user.FullName, &user.Email,
		&user.SubscriptionTier, &user.ActiveGoalName, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email address
func (s *PostgreSQLStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, handle, full_name, email, subscription_tier, active_goal_name, avatar_url, created_at
		FROM users WHERE email = $1`

	var user User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Handle, &user.FullName, &user.Email,
		&user.SubscriptionTier, &user.ActiveGoalName, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (s *PostgreSQLStore) UpdateUser(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	query := `
		UPDATE users 
		SET handle = $2, full_name = $3, email = $4, subscription_tier = $5, active_goal_name = $6, avatar_url = $7
		WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, user.ID, user.Handle, user.FullName, user.Email, user.SubscriptionTier, user.ActiveGoalName, user.AvatarURL)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// CreateConsumption creates a new consumption
func (s *PostgreSQLStore) CreateConsumption(ctx context.Context, consumption *Consumption) error {
	// Generate UUID if not provided
	if consumption.ID == "" {
		consumption.ID = generateUUID()
	}

	// Set created_at if not provided
	if consumption.CreatedAt.IsZero() {
		consumption.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO consumptions (
			id, user_id, transcript, total_calories, total_protein_g, total_fat_g, 
			total_carbs_g, dietary_fiber_g, total_sodium_mg, saturated_fat_g, 
			trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g, 
			vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
			thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg,
			vitamin_b12_mcg, biotin_mcg, pantothenic_acid_mg, choline_mg,
			calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg,
			molybdenum_mcg, chromium_mcg, fluoride_mg, chloride_mg,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44
		)`

	_, err := s.db.ExecContext(ctx, query,
		consumption.ID, consumption.UserID, consumption.Transcript,
		consumption.TotalCalories, consumption.TotalProtein, consumption.TotalFat,
		consumption.TotalCarbs, consumption.DietaryFiber, consumption.TotalSodium,
		consumption.SaturatedFat, consumption.TransFat, consumption.Cholesterol,
		consumption.TotalSugars, consumption.AddedSugars, consumption.VitaminA,
		consumption.VitaminC, consumption.VitaminD, consumption.VitaminE,
		consumption.VitaminK, consumption.Thiamine, consumption.Riboflavin,
		consumption.Niacin, consumption.VitaminB6, consumption.Folate,
		consumption.VitaminB12, consumption.Biotin, consumption.PantothenicAcid,
		consumption.Choline, consumption.Calcium, consumption.Iron,
		consumption.Magnesium, consumption.Phosphorus, consumption.Potassium,
		consumption.Zinc, consumption.Copper, consumption.Manganese,
		consumption.Selenium, consumption.Iodine, consumption.Molybdenum,
		consumption.Chromium, consumption.Fluoride, consumption.Chloride,
		consumption.CreatedAt, consumption.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create consumption: %w", err)
	}

	return nil
}

// GetConsumption retrieves a consumption by ID
func (s *PostgreSQLStore) GetConsumption(ctx context.Context, id string) (*Consumption, error) {
	query := `
		SELECT id, user_id, transcript, total_calories, total_protein_g, total_fat_g,
			   total_carbs_g, dietary_fiber_g, total_sodium_mg, saturated_fat_g,
			   trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
			   vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
			   thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg,
			   vitamin_b12_mcg, biotin_mcg, pantothenic_acid_mg, choline_mg,
			   calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			   zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg,
			   molybdenum_mcg, chromium_mcg, fluoride_mg, chloride_mg,
			   created_at, updated_at
		FROM consumptions WHERE id = $1`

	var consumption Consumption
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&consumption.ID, &consumption.UserID, &consumption.Transcript,
		&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
		&consumption.TotalCarbs, &consumption.DietaryFiber, &consumption.TotalSodium,
		&consumption.SaturatedFat, &consumption.TransFat, &consumption.Cholesterol,
		&consumption.TotalSugars, &consumption.AddedSugars, &consumption.VitaminA,
		&consumption.VitaminC, &consumption.VitaminD, &consumption.VitaminE,
		&consumption.VitaminK, &consumption.Thiamine, &consumption.Riboflavin,
		&consumption.Niacin, &consumption.VitaminB6, &consumption.Folate,
		&consumption.VitaminB12, &consumption.Biotin, &consumption.PantothenicAcid,
		&consumption.Choline, &consumption.Calcium, &consumption.Iron,
		&consumption.Magnesium, &consumption.Phosphorus, &consumption.Potassium,
		&consumption.Zinc, &consumption.Copper, &consumption.Manganese,
		&consumption.Selenium, &consumption.Iodine, &consumption.Molybdenum,
		&consumption.Chromium, &consumption.Fluoride, &consumption.Chloride,
		&consumption.CreatedAt, &consumption.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get consumption: %w", err)
	}

	return &consumption, nil
}

// UpdateConsumption updates an existing consumption
func (s *PostgreSQLStore) UpdateConsumption(ctx context.Context, consumption *Consumption) error {
	now := time.Now().UTC()
	consumption.UpdatedAt = &now

	query := `
		UPDATE consumptions SET 
			transcript = $2, total_calories = $3, total_protein_g = $4, total_fat_g = $5,
			total_carbs_g = $6, dietary_fiber_g = $7, total_sodium_mg = $8, 
			saturated_fat_g = $9, trans_fat_g = $10, cholesterol_mg = $11,
			total_sugars_g = $12, added_sugars_g = $13, vitamin_a_mcg = $14,
			vitamin_c_mg = $15, vitamin_d_mcg = $16, vitamin_e_mg = $17, vitamin_k_mcg = $18,
			thiamine_mg = $19, riboflavin_mg = $20, niacin_mg = $21, vitamin_b6_mg = $22,
			folate_mcg = $23, vitamin_b12_mcg = $24, biotin_mcg = $25, pantothenic_acid_mg = $26,
			choline_mg = $27, calcium_mg = $28, iron_mg = $29, magnesium_mg = $30,
			phosphorus_mg = $31, potassium_mg = $32, zinc_mg = $33, copper_mg = $34,
			manganese_mg = $35, selenium_mcg = $36, iodine_mcg = $37, molybdenum_mcg = $38,
			chromium_mcg = $39, fluoride_mg = $40, chloride_mg = $41, updated_at = $42
		WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, consumption.ID, consumption.Transcript,
		consumption.TotalCalories, consumption.TotalProtein, consumption.TotalFat,
		consumption.TotalCarbs, consumption.DietaryFiber, consumption.TotalSodium,
		consumption.SaturatedFat, consumption.TransFat, consumption.Cholesterol,
		consumption.TotalSugars, consumption.AddedSugars, consumption.VitaminA,
		consumption.VitaminC, consumption.VitaminD, consumption.VitaminE,
		consumption.VitaminK, consumption.Thiamine, consumption.Riboflavin,
		consumption.Niacin, consumption.VitaminB6, consumption.Folate,
		consumption.VitaminB12, consumption.Biotin, consumption.PantothenicAcid,
		consumption.Choline, consumption.Calcium, consumption.Iron,
		consumption.Magnesium, consumption.Phosphorus, consumption.Potassium,
		consumption.Zinc, consumption.Copper, consumption.Manganese,
		consumption.Selenium, consumption.Iodine, consumption.Molybdenum,
		consumption.Chromium, consumption.Fluoride, consumption.Chloride,
		consumption.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update consumption: %w", err)
	}

	return nil
}

// DeleteConsumption deletes a consumption by ID
func (s *PostgreSQLStore) DeleteConsumption(ctx context.Context, id string) error {
	query := `DELETE FROM consumptions WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete consumption: %w", err)
	}

	return nil
}

// GetConsumptionsByUser retrieves consumptions for a user with pagination
func (s *PostgreSQLStore) GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*Consumption, error) {
	query := `
		SELECT id, user_id, transcript, total_calories, total_protein_g, total_fat_g,
			   total_carbs_g, dietary_fiber_g, total_sodium_mg, saturated_fat_g,
			   trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
			   vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
			   thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg,
			   vitamin_b12_mcg, biotin_mcg, pantothenic_acid_mg, choline_mg,
			   calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			   zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg,
			   molybdenum_mcg, chromium_mcg, fluoride_mg, chloride_mg,
			   created_at, updated_at
		FROM consumptions 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumptions by user: %w", err)
	}
	defer rows.Close()

	var consumptions []*Consumption
	for rows.Next() {
		var consumption Consumption
		err := rows.Scan(
			&consumption.ID, &consumption.UserID, &consumption.Transcript,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.DietaryFiber, &consumption.TotalSodium,
			&consumption.SaturatedFat, &consumption.TransFat, &consumption.Cholesterol,
			&consumption.TotalSugars, &consumption.AddedSugars, &consumption.VitaminA,
			&consumption.VitaminC, &consumption.VitaminD, &consumption.VitaminE,
			&consumption.VitaminK, &consumption.Thiamine, &consumption.Riboflavin,
			&consumption.Niacin, &consumption.VitaminB6, &consumption.Folate,
			&consumption.VitaminB12, &consumption.Biotin, &consumption.PantothenicAcid,
			&consumption.Choline, &consumption.Calcium, &consumption.Iron,
			&consumption.Magnesium, &consumption.Phosphorus, &consumption.Potassium,
			&consumption.Zinc, &consumption.Copper, &consumption.Manganese,
			&consumption.Selenium, &consumption.Iodine, &consumption.Molybdenum,
			&consumption.Chromium, &consumption.Fluoride, &consumption.Chloride,
			&consumption.CreatedAt, &consumption.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption: %w", err)
		}
		consumptions = append(consumptions, &consumption)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return consumptions, nil
}

// GetConsumptionsByUserSince retrieves consumptions for a user since a specific time
func (s *PostgreSQLStore) GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Consumption, error) {
	query := `
		SELECT id, user_id, transcript, total_calories, total_protein_g, total_fat_g,
			   total_carbs_g, dietary_fiber_g, total_sodium_mg, saturated_fat_g,
			   trans_fat_g, cholesterol_mg, total_sugars_g, added_sugars_g,
			   vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, vitamin_k_mcg,
			   thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, folate_mcg,
			   vitamin_b12_mcg, biotin_mcg, pantothenic_acid_mg, choline_mg,
			   calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			   zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg,
			   molybdenum_mcg, chromium_mcg, fluoride_mg, chloride_mg,
			   created_at, updated_at
		FROM consumptions 
		WHERE user_id = $1 AND created_at >= $2
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumptions by user since: %w", err)
	}
	defer rows.Close()

	var consumptions []*Consumption
	for rows.Next() {
		var consumption Consumption
		err := rows.Scan(
			&consumption.ID, &consumption.UserID, &consumption.Transcript,
			&consumption.TotalCalories, &consumption.TotalProtein, &consumption.TotalFat,
			&consumption.TotalCarbs, &consumption.DietaryFiber, &consumption.TotalSodium,
			&consumption.SaturatedFat, &consumption.TransFat, &consumption.Cholesterol,
			&consumption.TotalSugars, &consumption.AddedSugars, &consumption.VitaminA,
			&consumption.VitaminC, &consumption.VitaminD, &consumption.VitaminE,
			&consumption.VitaminK, &consumption.Thiamine, &consumption.Riboflavin,
			&consumption.Niacin, &consumption.VitaminB6, &consumption.Folate,
			&consumption.VitaminB12, &consumption.Biotin, &consumption.PantothenicAcid,
			&consumption.Choline, &consumption.Calcium, &consumption.Iron,
			&consumption.Magnesium, &consumption.Phosphorus, &consumption.Potassium,
			&consumption.Zinc, &consumption.Copper, &consumption.Manganese,
			&consumption.Selenium, &consumption.Iodine, &consumption.Molybdenum,
			&consumption.Chromium, &consumption.Fluoride, &consumption.Chloride,
			&consumption.CreatedAt, &consumption.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption: %w", err)
		}
		consumptions = append(consumptions, &consumption)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return consumptions, nil
}

// GetNutritionSummary retrieves aggregated nutrition data for a user over a time period
func (s *PostgreSQLStore) GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*NutritionSummary, error) {
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
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`

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
			DATE(created_at)::text as date,
			COUNT(*) as consumption_count,
			COALESCE(SUM(total_calories), 0) as calories,
			COALESCE(SUM(total_protein_g), 0) as protein,
			COALESCE(SUM(total_fat_g), 0) as fat,
			COALESCE(SUM(total_carbs_g), 0) as carbs,
			COALESCE(SUM(dietary_fiber_g), 0) as fiber,
			COALESCE(SUM(total_sodium_mg), 0) as sodium
		FROM consumptions 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY DATE(created_at)
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

// CreateConsumptionItem creates a new consumption item relationship
func (s *PostgreSQLStore) CreateConsumptionItem(ctx context.Context, item *ConsumptionItem) error {
	if item.ID == "" {
		item.ID = generateUUID()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO consumption_items (id, consumption_id, item_id, grams, user_quantity, user_unit, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.ExecContext(ctx, query, item.ID, item.ConsumptionID, item.ItemID,
		item.Grams, item.UserQuantity, item.UserUnit, item.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create consumption item: %w", err)
	}

	return nil
}

// GetConsumptionItems retrieves all items for a consumption
func (s *PostgreSQLStore) GetConsumptionItems(ctx context.Context, consumptionID string) ([]*ConsumptionItem, error) {
	query := `
		SELECT id, consumption_id, item_id, grams, user_quantity, user_unit, created_at
		FROM consumption_items 
		WHERE consumption_id = $1 
		ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, consumptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query consumption items: %w", err)
	}
	defer rows.Close()

	var items []*ConsumptionItem
	for rows.Next() {
		item := &ConsumptionItem{}
		err := rows.Scan(&item.ID, &item.ConsumptionID, &item.ItemID, &item.Grams,
			&item.UserQuantity, &item.UserUnit, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate consumption items: %w", err)
	}

	return items, nil
}

// UpdateConsumptionItem updates an existing consumption item
func (s *PostgreSQLStore) UpdateConsumptionItem(ctx context.Context, item *ConsumptionItem) error {
	query := `
		UPDATE consumption_items SET 
			item_id = $2, grams = $3, user_quantity = $4, user_unit = $5
		WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, item.ID, item.ItemID, item.Grams,
		item.UserQuantity, item.UserUnit)
	if err != nil {
		return fmt.Errorf("failed to update consumption item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("consumption item not found")
	}

	return nil
}

// DeleteConsumptionItem deletes a consumption item by ID
func (s *PostgreSQLStore) DeleteConsumptionItem(ctx context.Context, id string) error {
	query := `DELETE FROM consumption_items WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete consumption item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("consumption item not found")
	}

	return nil
}

// DeleteConsumptionItemsByConsumption deletes all consumption items for a consumption
func (s *PostgreSQLStore) DeleteConsumptionItemsByConsumption(ctx context.Context, consumptionID string) error {
	query := `DELETE FROM consumption_items WHERE consumption_id = $1`

	_, err := s.db.ExecContext(ctx, query, consumptionID)
	if err != nil {
		return fmt.Errorf("failed to delete consumption items: %w", err)
	}

	return nil
}

// buildPostgreSQLInsertQuery builds the INSERT query for items table with PostgreSQL syntax
func buildPostgreSQLInsertQuery() string {
	columns := GetItemColumns()
	columnsList := strings.Join(columns, ", ")

	// Build placeholders for PostgreSQL ($1, $2, etc.)
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholdersList := strings.Join(placeholders, ", ")

	return fmt.Sprintf(`
		INSERT INTO items (%s) VALUES (%s)`, columnsList, placeholdersList)
}

// buildPostgreSQLSelectQuery builds the SELECT query for items table with PostgreSQL syntax
func buildPostgreSQLSelectQuery(where string) string {
	columns := GetItemColumns()
	columnsList := strings.Join(columns, ", ")

	query := fmt.Sprintf("SELECT %s FROM items", columnsList)
	if where != "" {
		query += " WHERE " + where
	}
	return query
}

// buildPostgreSQLUpdateQuery builds the UPDATE query for items table with PostgreSQL syntax
func buildPostgreSQLUpdateQuery() string {
	columns := GetItemColumns()
	// Skip id, created_at for updates
	updateColumns := make([]string, 0, len(columns)-2)
	paramIndex := 1
	for _, col := range columns {
		if col != "id" && col != "created_at" {
			updateColumns = append(updateColumns, fmt.Sprintf("%s = $%d", col, paramIndex))
			paramIndex++
		}
	}

	return fmt.Sprintf(`
		UPDATE items SET %s WHERE id = $%d`, strings.Join(updateColumns, ", "), paramIndex)
}

// CreateItem creates a new item
func (s *PostgreSQLStore) CreateItem(ctx context.Context, item *Item) error {
	now := time.Now().UTC()

	// Generate ID if not set
	if item.ID == "" {
		item.ID = generateUUID()
	}

	// Set timestamps
	item.CreatedAt = now
	item.UpdatedAt = now

	query := buildPostgreSQLInsertQuery()
	values := extractItemValues(item, true, true)

	_, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

// GetItem retrieves an item by ID
func (s *PostgreSQLStore) GetItem(ctx context.Context, id string) (*Item, error) {
	query := buildPostgreSQLSelectQuery("id = $1")

	item := &Item{}
	err := scanItemRow(s.db.QueryRowContext(ctx, query, id), item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Item not found
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return item, nil
}

// GetItemByName retrieves an item by normalized name and brand
func (s *PostgreSQLStore) GetItemByName(ctx context.Context, normalizedName, normalizedBrand string) (*Item, error) {
	query := buildPostgreSQLSelectQuery("normalized_name = $1 AND normalized_brand = $2")

	item := &Item{}
	err := scanItemRow(s.db.QueryRowContext(ctx, query, normalizedName, normalizedBrand), item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Item not found in cache
		}
		return nil, fmt.Errorf("failed to get item by name: %w", err)
	}

	return item, nil
}

// UpdateItem updates an existing item
func (s *PostgreSQLStore) UpdateItem(ctx context.Context, item *Item) error {
	now := time.Now().UTC()
	item.UpdatedAt = now

	query := buildPostgreSQLUpdateQuery()
	values := extractItemValues(item, false, false) // Don't include ID or created_at in values
	values = append(values, item.ID)                // Add ID at the end for WHERE clause

	result, err := s.db.ExecContext(ctx, query, values...)
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
func (s *PostgreSQLStore) GetStaleItems(ctx context.Context, staleAfter time.Time) ([]*Item, error) {
	query := buildPostgreSQLSelectQuery("updated_at < $1 ORDER BY updated_at ASC")

	rows, err := s.db.QueryContext(ctx, query, staleAfter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
		err := scanItemRow(rows, item)
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

// UpsertUserGoal creates or updates a user goal
func (s *PostgreSQLStore) UpsertUserGoal(ctx context.Context, goal *UserGoal) error {
	query := `
		INSERT INTO user_goals (id, user_id, name, overrides_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, name) DO UPDATE SET
			overrides_json = EXCLUDED.overrides_json,
			updated_at = EXCLUDED.updated_at`

	now := time.Now().UTC()
	if goal.ID == "" {
		goal.ID = generateUUID()
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
func (s *PostgreSQLStore) GetUserGoal(ctx context.Context, userID, name string) (*UserGoal, error) {
	query := `SELECT id, user_id, name, overrides_json, created_at, updated_at 
			  FROM user_goals WHERE user_id = $1 AND name = $2`

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

// GetUserGoals retrieves all goal sets for a user
func (s *PostgreSQLStore) GetUserGoals(ctx context.Context, userID string) ([]*UserGoal, error) {
	query := `SELECT id, user_id, name, overrides_json, created_at, updated_at 
			  FROM user_goals WHERE user_id = $1 ORDER BY updated_at DESC`

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

// DeleteUserGoal deletes a user goal
func (s *PostgreSQLStore) DeleteUserGoal(ctx context.Context, userID, name string) error {
	query := `DELETE FROM user_goals WHERE user_id = $1 AND name = $2`

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

// SetActiveGoal sets the active goal for a user
func (s *PostgreSQLStore) SetActiveGoal(ctx context.Context, userID, goalName string) error {
	// First verify the goal exists
	goal, err := s.GetUserGoal(ctx, userID, goalName)
	if err != nil {
		return fmt.Errorf("failed to verify goal exists: %w", err)
	}
	if goal == nil {
		return fmt.Errorf("goal '%s' not found for user", goalName)
	}

	// Update user's active goal
	query := `UPDATE users SET active_goal_name = $1 WHERE id = $2`
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
func (s *PostgreSQLStore) ClearActiveGoal(ctx context.Context, userID string) error {
	// Update user's active goal to NULL
	query := `UPDATE users SET active_goal_name = NULL WHERE id = $1`
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
func (s *PostgreSQLStore) GetActiveGoalName(ctx context.Context, userID string) (*string, error) {
	query := `SELECT active_goal_name FROM users WHERE id = $1`

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
func (s *PostgreSQLStore) UpsertUserBiometrics(ctx context.Context, biometrics *UserBiometrics) error {
	query := `
		INSERT INTO user_biometrics (id, user_id, birth_date, sex, height_cm, weight_kg, activity_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			birth_date = EXCLUDED.birth_date,
			sex = EXCLUDED.sex,
			height_cm = EXCLUDED.height_cm,
			weight_kg = EXCLUDED.weight_kg,
			activity_level = EXCLUDED.activity_level,
			updated_at = EXCLUDED.updated_at`

	now := time.Now().UTC()
	if biometrics.ID == "" {
		biometrics.ID = generateUUID()
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
func (s *PostgreSQLStore) GetUserBiometrics(ctx context.Context, userID string) (*UserBiometrics, error) {
	query := `SELECT id, user_id, birth_date, sex, height_cm, weight_kg, activity_level, created_at, updated_at 
			  FROM user_biometrics WHERE user_id = $1`

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
func (s *PostgreSQLStore) DeleteUserBiometrics(ctx context.Context, userID string) error {
	query := `DELETE FROM user_biometrics WHERE user_id = $1`

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

// CreateItemAlias creates a new item alias
func (s *PostgreSQLStore) CreateItemAlias(ctx context.Context, alias *ItemAlias) error {
	query := `
		INSERT INTO item_aliases (id, alias_name, alias_brand, canonical_name, canonical_brand, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	now := time.Now().UTC()
	alias.ID = generateUUID()
	alias.CreatedAt = now

	_, err := s.db.ExecContext(ctx, query, alias.ID, alias.AliasName, alias.AliasBrand,
		alias.CanonicalName, alias.CanonicalBrand, now)
	if err != nil {
		return fmt.Errorf("failed to create item alias: %w", err)
	}

	return nil
}

// GetCanonicalName retrieves canonical name and brand for an alias
func (s *PostgreSQLStore) GetCanonicalName(ctx context.Context, aliasName, aliasBrand string) (canonicalName, canonicalBrand string, err error) {
	query := `SELECT canonical_name, canonical_brand FROM item_aliases WHERE alias_name = $1 AND alias_brand = $2`

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
