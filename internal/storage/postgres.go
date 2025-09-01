package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
)

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
		INSERT INTO profiles (id, email, handle, full_name, subscription_tier, avatar_url, created_at) 
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
		FROM profiles WHERE id = $1`

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
		FROM profiles WHERE email = $1`

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
		UPDATE profiles 
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
			omega3_ala_g, omega3_epa_g, omega3_dha_g, omega6_g,
			creatine_mg, caffeine_mg, alcohol_g,
			polyunsaturated_fat_g, monounsaturated_fat_g,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44,
			$45, $46, $47, $48, $49, $50, $51, $52, $53
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
		consumption.Omega3Ala, consumption.Omega3Epa, consumption.Omega3Dha,
		consumption.Omega6, consumption.Creatine, consumption.Caffeine, consumption.Alcohol,
		consumption.PolyunsaturatedFat, consumption.MonounsaturatedFat,
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
			   omega3_ala_g, omega3_epa_g, omega3_dha_g, omega6_g,
			   creatine_mg, caffeine_mg, alcohol_g,
			   polyunsaturated_fat_g, monounsaturated_fat_g,
			   note, created_at, updated_at
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
		&consumption.Omega3Ala, &consumption.Omega3Epa, &consumption.Omega3Dha,
		&consumption.Omega6, &consumption.Creatine, &consumption.Caffeine, &consumption.Alcohol,
		&consumption.PolyunsaturatedFat, &consumption.MonounsaturatedFat,
		&consumption.Note, &consumption.CreatedAt, &consumption.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get consumption: %w", err)
	}

	// Load labels for the consumption
	labels, err := s.ListConsumptionLabels(ctx, consumption.UserID, consumption.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load consumption labels: %w", err)
	}
	consumption.Labels = labels

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
			   omega3_ala_g, omega3_epa_g, omega3_dha_g, omega6_g,
			   creatine_mg, caffeine_mg, alcohol_g,
			   polyunsaturated_fat_g, monounsaturated_fat_g,
			   note, created_at, updated_at
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
			&consumption.Omega3Ala, &consumption.Omega3Epa, &consumption.Omega3Dha,
			&consumption.Omega6, &consumption.Creatine, &consumption.Caffeine, &consumption.Alcohol,
			&consumption.PolyunsaturatedFat, &consumption.MonounsaturatedFat,
			&consumption.Note, &consumption.CreatedAt, &consumption.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption: %w", err)
		}
		consumptions = append(consumptions, &consumption)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Load labels for each consumption
	for _, consumption := range consumptions {
		labels, err := s.ListConsumptionLabels(ctx, userID, consumption.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load consumption labels: %w", err)
		}
		consumption.Labels = labels
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
			   omega3_ala_g, omega3_epa_g, omega3_dha_g, omega6_g,
			   creatine_mg, caffeine_mg, alcohol_g,
			   polyunsaturated_fat_g, monounsaturated_fat_g,
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
			&consumption.Omega3Ala, &consumption.Omega3Epa, &consumption.Omega3Dha,
			&consumption.Omega6, &consumption.Creatine, &consumption.Caffeine, &consumption.Alcohol,
			&consumption.PolyunsaturatedFat, &consumption.MonounsaturatedFat,
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

// GetConsumptionsByUserDateRange retrieves consumptions for a user within a date range with pagination
func (s *PostgreSQLStore) GetConsumptionsByUserDateRange(ctx context.Context, userID string, start, end time.Time, limit, offset int) ([]*Consumption, error) {
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
			   omega3_ala_g, omega3_epa_g, omega3_dha_g, omega6_g,
			   creatine_mg, caffeine_mg, alcohol_g,
			   polyunsaturated_fat_g, monounsaturated_fat_g,
			   created_at, updated_at
		FROM consumptions 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`

	rows, err := s.db.QueryContext(ctx, query, userID, start, end, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumptions by user date range: %w", err)
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
			&consumption.Omega3Ala, &consumption.Omega3Epa, &consumption.Omega3Dha,
			&consumption.Omega6, &consumption.Creatine, &consumption.Caffeine, &consumption.Alcohol,
			&consumption.PolyunsaturatedFat, &consumption.MonounsaturatedFat,
			&consumption.CreatedAt, &consumption.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption: %w", err)
		}
		consumptions = append(consumptions, &consumption)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Load labels for each consumption
	for _, consumption := range consumptions {
		labels, err := s.ListConsumptionLabels(ctx, consumption.UserID, consumption.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load consumption labels: %w", err)
		}
		consumption.Labels = labels
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
			COALESCE(SUM(selenium_mcg), 0) as total_selenium,
			COALESCE(SUM(iodine_mcg), 0) as total_iodine,
			COALESCE(SUM(molybdenum_mcg), 0) as total_molybdenum,
			COALESCE(SUM(chromium_mcg), 0) as total_chromium,
			COALESCE(SUM(fluoride_mg), 0) as total_fluoride,
			COALESCE(SUM(chloride_mg), 0) as total_chloride,
			COALESCE(SUM(biotin_mcg), 0) as total_biotin,
			COALESCE(SUM(pantothenic_acid_mg), 0) as total_pantothenic_acid,
			COALESCE(SUM(choline_mg), 0) as total_choline,
			COALESCE(SUM(omega3_ala_g), 0) as total_omega3_ala,
			COALESCE(SUM(omega3_epa_g), 0) as total_omega3_epa,
			COALESCE(SUM(omega3_dha_g), 0) as total_omega3_dha,
			COALESCE(SUM(omega6_g), 0) as total_omega6,
			COALESCE(SUM(creatine_mg), 0) as total_creatine,
			COALESCE(SUM(caffeine_mg), 0) as total_caffeine,
			COALESCE(SUM(alcohol_g), 0) as total_alcohol,
			COALESCE(SUM(polyunsaturated_fat_g), 0) as total_polyunsaturated_fat,
			COALESCE(SUM(monounsaturated_fat_g), 0) as total_monounsaturated_fat
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
		&summary.TotalIodine,
		&summary.TotalMolybdenum,
		&summary.TotalChromium,
		&summary.TotalFluoride,
		&summary.TotalChloride,
		&summary.TotalBiotin,
		&summary.TotalPantothenicAcid,
		&summary.TotalCholine,
		&summary.TotalOmega3Ala,
		&summary.TotalOmega3Epa,
		&summary.TotalOmega3Dha,
		&summary.TotalOmega6,
		&summary.TotalCreatine,
		&summary.TotalCaffeine,
		&summary.TotalAlcohol,
		&summary.TotalPolyunsaturatedFat,
		&summary.TotalMonounsaturatedFat,
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
	if item.UpdatedAt == nil {
		now := item.CreatedAt
		item.UpdatedAt = &now
	}

	query := `
		INSERT INTO consumption_items (
			id, consumption_id, item_id, name, brand, grams, user_quantity, user_unit, 
			note, calories, protein_g, total_fat_g, saturated_fat_g, trans_fat_g, 
			cholesterol_mg, sodium_mg, total_carbs_g, dietary_fiber_g, total_sugars_g, 
			added_sugars_g, vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, 
			vitamin_k_mcg, thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, 
			folate_mcg, vitamin_b12_mcg, biotin_mcg, pantothenic_acid_mg, choline_mg, 
			calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg, zinc_mg, 
			copper_mg, manganese_mg, selenium_mcg, iodine_mcg, molybdenum_mcg, 
			chromium_mcg, fluoride_mg, chloride_mg, omega3_ala_g, omega3_epa_g, 
			omega3_dha_g, omega6_g, creatine_mg, caffeine_mg, alcohol_g, 
			polyunsaturated_fat_g, monounsaturated_fat_g, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, 
			$30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, 
			$44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, 
			$58, $59
		)`

	_, err := s.db.ExecContext(ctx, query,
		item.ID, item.ConsumptionID, item.ItemID, item.Name, item.Brand,
		item.Grams, item.UserQuantity, item.UserUnit, item.Note,
		item.Calories, item.ProteinG, item.TotalFatG, item.SaturatedFatG, item.TransFatG,
		item.CholesterolMg, item.SodiumMg, item.TotalCarbsG, item.DietaryFiberG, item.TotalSugarsG,
		item.AddedSugarsG, item.VitaminAMcg, item.VitaminCMg, item.VitaminDMcg, item.VitaminEMg,
		item.VitaminKMcg, item.ThiamineMg, item.RiboflavinMg, item.NiacinMg, item.VitaminB6Mg,
		item.FolateMcg, item.VitaminB12Mcg, item.BiotinMcg, item.PantothenicAcidMg, item.CholineMg,
		item.CalciumMg, item.IronMg, item.MagnesiumMg, item.PhosphorusMg, item.PotassiumMg, item.ZincMg,
		item.CopperMg, item.ManganeseMg, item.SeleniumMcg, item.IodineMcg, item.MolybdenumMcg,
		item.ChromiumMcg, item.FluorideMg, item.ChlorideMg, item.Omega3AlaG, item.Omega3EpaG,
		item.Omega3DhaG, item.Omega6G, item.CreatineMg, item.CaffeineMg, item.AlcoholG,
		item.PolyunsaturatedFatG, item.MonounsaturatedFatG, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create consumption item: %w", err)
	}

	return nil
}

// GetConsumptionItems retrieves all items for a consumption
func (s *PostgreSQLStore) GetConsumptionItems(ctx context.Context, consumptionID string) ([]*ConsumptionItem, error) {
	query := `
		SELECT id, consumption_id, item_id, name, brand, grams, user_quantity, user_unit, 
			note, calories, protein_g, total_fat_g, saturated_fat_g, trans_fat_g, 
			cholesterol_mg, sodium_mg, total_carbs_g, dietary_fiber_g, total_sugars_g, 
			added_sugars_g, vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg, vitamin_e_mg, 
			vitamin_k_mcg, thiamine_mg, riboflavin_mg, niacin_mg, vitamin_b6_mg, 
			folate_mcg, vitamin_b12_mcg, biotin_mcg, pantothenic_acid_mg, choline_mg, 
			calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg, zinc_mg, 
			copper_mg, manganese_mg, selenium_mcg, iodine_mcg, molybdenum_mcg, 
			chromium_mcg, fluoride_mg, chloride_mg, omega3_ala_g, omega3_epa_g, 
			omega3_dha_g, omega6_g, creatine_mg, caffeine_mg, alcohol_g, 
			polyunsaturated_fat_g, monounsaturated_fat_g, created_at, updated_at
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
		err := rows.Scan(&item.ID, &item.ConsumptionID, &item.ItemID, &item.Name, &item.Brand,
			&item.Grams, &item.UserQuantity, &item.UserUnit, &item.Note,
			&item.Calories, &item.ProteinG, &item.TotalFatG, &item.SaturatedFatG, &item.TransFatG,
			&item.CholesterolMg, &item.SodiumMg, &item.TotalCarbsG, &item.DietaryFiberG, &item.TotalSugarsG,
			&item.AddedSugarsG, &item.VitaminAMcg, &item.VitaminCMg, &item.VitaminDMcg, &item.VitaminEMg,
			&item.VitaminKMcg, &item.ThiamineMg, &item.RiboflavinMg, &item.NiacinMg, &item.VitaminB6Mg,
			&item.FolateMcg, &item.VitaminB12Mcg, &item.BiotinMcg, &item.PantothenicAcidMg, &item.CholineMg,
			&item.CalciumMg, &item.IronMg, &item.MagnesiumMg, &item.PhosphorusMg, &item.PotassiumMg, &item.ZincMg,
			&item.CopperMg, &item.ManganeseMg, &item.SeleniumMcg, &item.IodineMcg, &item.MolybdenumMcg,
			&item.ChromiumMcg, &item.FluorideMg, &item.ChlorideMg, &item.Omega3AlaG, &item.Omega3EpaG,
			&item.Omega3DhaG, &item.Omega6G, &item.CreatineMg, &item.CaffeineMg, &item.AlcoholG,
			&item.PolyunsaturatedFatG, &item.MonounsaturatedFatG, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate consumption items: %w", err)
	}

	// Load labels for each consumption item
	for _, item := range items {
		// We need the user ID to load labels, get it from the consumption
		var userID string
		err := s.db.QueryRowContext(ctx, "SELECT user_id FROM consumptions WHERE id = $1", item.ConsumptionID).Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user ID for consumption: %w", err)
		}

		labels, err := s.ListConsumptionItemLabels(ctx, userID, item.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load consumption item labels: %w", err)
		}
		item.Labels = labels
	}

	return items, nil
}

// UpdateConsumptionItem updates an existing consumption item
func (s *PostgreSQLStore) UpdateConsumptionItem(ctx context.Context, item *ConsumptionItem) error {
	now := time.Now().UTC()
	item.UpdatedAt = &now

	query := `
		UPDATE consumption_items SET 
			item_id = $2, name = $3, brand = $4, grams = $5, user_quantity = $6, 
			user_unit = $7, note = $8, calories = $9, protein_g = $10, 
			total_fat_g = $11, saturated_fat_g = $12, trans_fat_g = $13, cholesterol_mg = $14, 
			sodium_mg = $15, total_carbs_g = $16, dietary_fiber_g = $17, total_sugars_g = $18, 
			added_sugars_g = $19, vitamin_a_mcg = $20, vitamin_c_mg = $21, vitamin_d_mcg = $22, 
			vitamin_e_mg = $23, vitamin_k_mcg = $24, thiamine_mg = $25, riboflavin_mg = $26, 
			niacin_mg = $27, vitamin_b6_mg = $28, folate_mcg = $29, vitamin_b12_mcg = $30, 
			biotin_mcg = $31, pantothenic_acid_mg = $32, choline_mg = $33, calcium_mg = $34, 
			iron_mg = $35, magnesium_mg = $36, phosphorus_mg = $37, potassium_mg = $38, 
			zinc_mg = $39, copper_mg = $40, manganese_mg = $41, selenium_mcg = $42, 
			iodine_mcg = $43, molybdenum_mcg = $44, chromium_mcg = $45, fluoride_mg = $46, 
			chloride_mg = $47, omega3_ala_g = $48, omega3_epa_g = $49, omega3_dha_g = $50, 
			omega6_g = $51, creatine_mg = $52, caffeine_mg = $53, alcohol_g = $54, 
			polyunsaturated_fat_g = $55, monounsaturated_fat_g = $56, updated_at = $57
		WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, item.ID, item.ItemID, item.Name, item.Brand,
		item.Grams, item.UserQuantity, item.UserUnit, item.Note,
		item.Calories, item.ProteinG, item.TotalFatG, item.SaturatedFatG, item.TransFatG,
		item.CholesterolMg, item.SodiumMg, item.TotalCarbsG, item.DietaryFiberG, item.TotalSugarsG,
		item.AddedSugarsG, item.VitaminAMcg, item.VitaminCMg, item.VitaminDMcg, item.VitaminEMg,
		item.VitaminKMcg, item.ThiamineMg, item.RiboflavinMg, item.NiacinMg, item.VitaminB6Mg,
		item.FolateMcg, item.VitaminB12Mcg, item.BiotinMcg, item.PantothenicAcidMg, item.CholineMg,
		item.CalciumMg, item.IronMg, item.MagnesiumMg, item.PhosphorusMg, item.PotassiumMg, item.ZincMg,
		item.CopperMg, item.ManganeseMg, item.SeleniumMcg, item.IodineMcg, item.MolybdenumMcg,
		item.ChromiumMcg, item.FluorideMg, item.ChlorideMg, item.Omega3AlaG, item.Omega3EpaG,
		item.Omega3DhaG, item.Omega6G, item.CreatineMg, item.CaffeineMg, item.AlcoholG,
		item.PolyunsaturatedFatG, item.MonounsaturatedFatG, item.UpdatedAt)
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
	query := `UPDATE profiles SET active_goal_name = $1 WHERE id = $2`
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
	query := `UPDATE profiles SET active_goal_name = NULL WHERE id = $1`
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
	query := `SELECT active_goal_name FROM profiles WHERE id = $1`

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

// Color normalization helper
func normalizeColor(color string) string {
	// Add # prefix if not present
	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}
	// Convert to uppercase for consistency
	return strings.ToUpper(color)
}

// CreateLabel creates a new label for a user
func (s *PostgreSQLStore) CreateLabel(ctx context.Context, label *Label) error {
	now := time.Now()
	label.ID = generateUUID()
	label.CreatedAt = now
	label.UpdatedAt = now
	label.Color = normalizeColor(label.Color)

	query := `INSERT INTO labels (id, user_id, name, description, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.ExecContext(ctx, query, label.ID, label.UserID, label.Name, label.Description,
		label.Color, label.CreatedAt, label.UpdatedAt)
	if err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "ux_labels_user_name") {
			return fmt.Errorf("label name already exists: %w", err)
		}
		// Check for limit violation
		if strings.Contains(err.Error(), "label_limit_exceeded") {
			return fmt.Errorf("label limit exceeded (100 labels per user): %w", err)
		}
		return fmt.Errorf("failed to create label: %w", err)
	}

	return nil
}

// UpdateLabel updates an existing label
func (s *PostgreSQLStore) UpdateLabel(ctx context.Context, label *Label) error {
	label.UpdatedAt = time.Now()
	if label.Color != "" {
		label.Color = normalizeColor(label.Color)
	}

	query := `UPDATE labels SET name = $3, description = $4, color = $5, updated_at = $6
		WHERE id = $1 AND user_id = $2`

	result, err := s.db.ExecContext(ctx, query, label.ID, label.UserID, label.Name,
		label.Description, label.Color, label.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "ux_labels_user_name") {
			return fmt.Errorf("label name already exists: %w", err)
		}
		return fmt.Errorf("failed to update label: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("label not found or access denied")
	}

	return nil
}

// DeleteLabel deletes a label and all its assignments
func (s *PostgreSQLStore) DeleteLabel(ctx context.Context, userID, id string) error {
	query := `DELETE FROM labels WHERE id = $1 AND user_id = $2`

	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("label not found or access denied")
	}

	return nil
}

// GetLabel retrieves a label by ID for a user
func (s *PostgreSQLStore) GetLabel(ctx context.Context, userID, id string) (*Label, error) {
	query := `SELECT id, user_id, name, description, color, created_at, updated_at
		FROM labels WHERE id = $1 AND user_id = $2`

	label := &Label{}
	err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&label.ID, &label.UserID, &label.Name, &label.Description,
		&label.Color, &label.CreatedAt, &label.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("label not found")
		}
		return nil, fmt.Errorf("failed to get label: %w", err)
	}

	return label, nil
}

// ListLabels retrieves all labels for a user with usage counts
func (s *PostgreSQLStore) ListLabels(ctx context.Context, userID string) ([]*LabelWithUsage, error) {
	query := `
		SELECT l.id, l.user_id, l.name, l.description, l.color, l.created_at, l.updated_at,
			COALESCE(c.consumption_count, 0) as consumption_count,
			COALESCE(i.item_count, 0) as item_count
		FROM labels l
		LEFT JOIN (
			SELECT label_id, COUNT(*) as consumption_count
			FROM consumption_labels cl
			JOIN consumptions c ON c.id = cl.consumption_id
			WHERE c.user_id = $1
			GROUP BY label_id
		) c ON c.label_id = l.id
		LEFT JOIN (
			SELECT label_id, COUNT(*) as item_count
			FROM consumption_item_labels cil
			JOIN consumption_items ci ON ci.id = cil.consumption_item_id
			JOIN consumptions cons ON cons.id = ci.consumption_id
			WHERE cons.user_id = $1
			GROUP BY label_id
		) i ON i.label_id = l.id
		WHERE l.user_id = $1
		ORDER BY l.name`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}
	defer rows.Close()

	var labels []*LabelWithUsage
	for rows.Next() {
		label := &LabelWithUsage{}
		err := rows.Scan(
			&label.ID, &label.UserID, &label.Name, &label.Description,
			&label.Color, &label.CreatedAt, &label.UpdatedAt,
			&label.ConsumptionCount, &label.ItemCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over labels: %w", err)
	}

	return labels, nil
}

// ListConsumptionLabels retrieves all labels assigned to a consumption
func (s *PostgreSQLStore) ListConsumptionLabels(ctx context.Context, userID, consumptionID string) ([]*Label, error) {
	query := `
		SELECT l.id, l.user_id, l.name, l.description, l.color, l.created_at, l.updated_at
		FROM labels l
		JOIN consumption_labels cl ON l.id = cl.label_id
		JOIN consumptions c ON c.id = cl.consumption_id
		WHERE cl.consumption_id = $1 AND c.user_id = $2 AND l.user_id = $2
		ORDER BY l.name`

	rows, err := s.db.QueryContext(ctx, query, consumptionID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumption labels: %w", err)
	}
	defer rows.Close()

	var labels []*Label
	for rows.Next() {
		label := &Label{}
		err := rows.Scan(
			&label.ID, &label.UserID, &label.Name, &label.Description,
			&label.Color, &label.CreatedAt, &label.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over consumption labels: %w", err)
	}

	return labels, nil
}

// AssignConsumptionLabels assigns labels to a consumption
func (s *PostgreSQLStore) AssignConsumptionLabels(ctx context.Context, userID, consumptionID string, labelIDs []string) error {
	if len(labelIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify consumption ownership
	var ownerID string
	err = tx.QueryRowContext(ctx, "SELECT user_id FROM consumptions WHERE id = $1", consumptionID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("consumption not found")
		}
		return fmt.Errorf("failed to verify consumption ownership: %w", err)
	}
	if ownerID != userID {
		return fmt.Errorf("access denied")
	}

	// Clear existing assignments
	_, err = tx.ExecContext(ctx, "DELETE FROM consumption_labels WHERE consumption_id = $1", consumptionID)
	if err != nil {
		return fmt.Errorf("failed to clear existing assignments: %w", err)
	}

	// Add new assignments
	now := time.Now()
	for _, labelID := range labelIDs {
		// Verify label ownership
		var labelOwnerID string
		err = tx.QueryRowContext(ctx, "SELECT user_id FROM labels WHERE id = $1", labelID).Scan(&labelOwnerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("label not found: %s", labelID)
			}
			return fmt.Errorf("failed to verify label ownership: %w", err)
		}
		if labelOwnerID != userID {
			return fmt.Errorf("label access denied: %s", labelID)
		}

		// Insert assignment (ignore duplicates with ON CONFLICT)
		_, err = tx.ExecContext(ctx,
			"INSERT INTO consumption_labels (consumption_id, label_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			consumptionID, labelID, now)
		if err != nil {
			return fmt.Errorf("failed to assign label %s: %w", labelID, err)
		}
	}

	return tx.Commit()
}

// UnassignConsumptionLabel removes a label assignment from a consumption
func (s *PostgreSQLStore) UnassignConsumptionLabel(ctx context.Context, userID, consumptionID, labelID string) error {
	query := `
		DELETE FROM consumption_labels 
		WHERE consumption_id = $1 AND label_id = $2
		AND EXISTS (SELECT 1 FROM consumptions WHERE id = $1 AND user_id = $3)
		AND EXISTS (SELECT 1 FROM labels WHERE id = $2 AND user_id = $3)`

	result, err := s.db.ExecContext(ctx, query, consumptionID, labelID, userID)
	if err != nil {
		return fmt.Errorf("failed to unassign consumption label: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("assignment not found or access denied")
	}

	return nil
}

// ListConsumptionItemLabels retrieves all labels assigned to a consumption item
func (s *PostgreSQLStore) ListConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string) ([]*Label, error) {
	query := `
		SELECT l.id, l.user_id, l.name, l.description, l.color, l.created_at, l.updated_at
		FROM labels l
		JOIN consumption_item_labels cil ON l.id = cil.label_id
		JOIN consumption_items ci ON ci.id = cil.consumption_item_id
		JOIN consumptions c ON c.id = ci.consumption_id
		WHERE cil.consumption_item_id = $1 AND c.user_id = $2 AND l.user_id = $2
		ORDER BY l.name`

	rows, err := s.db.QueryContext(ctx, query, consumptionItemID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumption item labels: %w", err)
	}
	defer rows.Close()

	var labels []*Label
	for rows.Next() {
		label := &Label{}
		err := rows.Scan(
			&label.ID, &label.UserID, &label.Name, &label.Description,
			&label.Color, &label.CreatedAt, &label.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over consumption item labels: %w", err)
	}

	return labels, nil
}

// AssignConsumptionItemLabels assigns labels to a consumption item
func (s *PostgreSQLStore) AssignConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string, labelIDs []string) error {
	if len(labelIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify consumption item ownership
	var ownerID string
	err = tx.QueryRowContext(ctx, `
		SELECT c.user_id 
		FROM consumption_items ci 
		JOIN consumptions c ON c.id = ci.consumption_id 
		WHERE ci.id = $1`, consumptionItemID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("consumption item not found")
		}
		return fmt.Errorf("failed to verify consumption item ownership: %w", err)
	}
	if ownerID != userID {
		return fmt.Errorf("access denied")
	}

	// Clear existing assignments
	_, err = tx.ExecContext(ctx, "DELETE FROM consumption_item_labels WHERE consumption_item_id = $1", consumptionItemID)
	if err != nil {
		return fmt.Errorf("failed to clear existing assignments: %w", err)
	}

	// Add new assignments
	now := time.Now()
	for _, labelID := range labelIDs {
		// Verify label ownership
		var labelOwnerID string
		err = tx.QueryRowContext(ctx, "SELECT user_id FROM labels WHERE id = $1", labelID).Scan(&labelOwnerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("label not found: %s", labelID)
			}
			return fmt.Errorf("failed to verify label ownership: %w", err)
		}
		if labelOwnerID != userID {
			return fmt.Errorf("label access denied: %s", labelID)
		}

		// Insert assignment (ignore duplicates with ON CONFLICT)
		_, err = tx.ExecContext(ctx,
			"INSERT INTO consumption_item_labels (consumption_item_id, label_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			consumptionItemID, labelID, now)
		if err != nil {
			return fmt.Errorf("failed to assign label %s: %w", labelID, err)
		}
	}

	return tx.Commit()
}

// UnassignConsumptionItemLabel removes a label assignment from a consumption item
func (s *PostgreSQLStore) UnassignConsumptionItemLabel(ctx context.Context, userID, consumptionItemID, labelID string) error {
	query := `
		DELETE FROM consumption_item_labels 
		WHERE consumption_item_id = $1 AND label_id = $2
		AND EXISTS (
			SELECT 1 FROM consumption_items ci 
			JOIN consumptions c ON c.id = ci.consumption_id 
			WHERE ci.id = $1 AND c.user_id = $3
		)
		AND EXISTS (SELECT 1 FROM labels WHERE id = $2 AND user_id = $3)`

	result, err := s.db.ExecContext(ctx, query, consumptionItemID, labelID, userID)
	if err != nil {
		return fmt.Errorf("failed to unassign consumption item label: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("assignment not found or access denied")
	}

	return nil
}

// GetConsumptionsByLabels retrieves consumptions filtered by labels
func (s *PostgreSQLStore) GetConsumptionsByLabels(ctx context.Context, userID string, labelNames []string, matchAll bool, limit, offset int) ([]*Consumption, error) {
	if len(labelNames) == 0 {
		// If no labels specified, use the regular method
		return s.GetConsumptionsByUser(ctx, userID, limit, offset)
	}

	// Build placeholders for label names
	placeholders := make([]string, len(labelNames))
	args := []interface{}{userID}
	for i, name := range labelNames {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, strings.ToLower(name))
	}

	var havingClause string
	if matchAll {
		havingClause = fmt.Sprintf("HAVING COUNT(DISTINCT l.id) = %d", len(labelNames))
	} else {
		havingClause = "HAVING COUNT(DISTINCT l.id) > 0"
	}

	// Query to get consumption IDs that match label criteria
	labelFilter := fmt.Sprintf(`
		SELECT c.id
		FROM consumptions c
		JOIN consumption_labels cl ON c.id = cl.consumption_id
		JOIN labels l ON l.id = cl.label_id
		WHERE c.user_id = $1 AND LOWER(l.name) IN (%s)
		GROUP BY c.id
		%s`, strings.Join(placeholders, ","), havingClause)

	// Main query with limit and offset
	query := fmt.Sprintf(`
		SELECT id, user_id, transcript, total_calories, total_protein_g, total_fat_g, total_carbs_g,
			dietary_fiber_g, total_sodium_mg, saturated_fat_g, trans_fat_g, cholesterol_mg,
			total_sugars_g, added_sugars_g, vitamin_a_mcg, vitamin_c_mg, vitamin_d_mcg,
			vitamin_e_mg, vitamin_k_mcg, thiamine_mg, riboflavin_mg, niacin_mg,
			vitamin_b6_mg, folate_mcg, vitamin_b12_mcg, biotin_mcg, pantothenic_acid_mg,
			choline_mg, calcium_mg, iron_mg, magnesium_mg, phosphorus_mg, potassium_mg,
			zinc_mg, copper_mg, manganese_mg, selenium_mcg, iodine_mcg, molybdenum_mcg,
			chromium_mcg, fluoride_mg, chloride_mg, omega3_ala_g, omega3_epa_g,
			omega3_dha_g, omega6_g, creatine_mg, caffeine_mg, alcohol_g,
			polyunsaturated_fat_g, monounsaturated_fat_g, note, created_at, updated_at
		FROM consumptions
		WHERE user_id = $1 AND id IN (%s)
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		labelFilter, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query consumptions by labels: %w", err)
	}
	defer rows.Close()

	var consumptions []*Consumption
	for rows.Next() {
		consumption := &Consumption{}
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
			&consumption.Omega3Ala, &consumption.Omega3Epa, &consumption.Omega3Dha,
			&consumption.Omega6, &consumption.Creatine, &consumption.Caffeine,
			&consumption.Alcohol, &consumption.PolyunsaturatedFat, &consumption.MonounsaturatedFat,
			&consumption.Note, &consumption.CreatedAt, &consumption.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan consumption: %w", err)
		}
		consumptions = append(consumptions, consumption)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over consumptions: %w", err)
	}

	// Load labels for each consumption
	for _, consumption := range consumptions {
		labels, err := s.ListConsumptionLabels(ctx, userID, consumption.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load consumption labels: %w", err)
		}
		consumption.Labels = labels
	}

	return consumptions, nil
}
// Event operations

// CreateEvent creates a new event for a user
func (s *PostgreSQLStore) CreateEvent(ctx context.Context, event *Event) error {
	now := time.Now()
	event.ID = generateUUID()
	event.CreatedAt = now
	event.UpdatedAt = now
	if event.Color != nil && *event.Color != "" {
		normalized := normalizeColor(*event.Color)
		event.Color = &normalized
	}

	query := `
		INSERT INTO events (id, user_id, name, category, started_at, ended_at, level, note, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := s.db.ExecContext(ctx, query,
		event.ID, event.UserID, event.Name, event.Category, event.StartedAt, event.EndedAt,
		event.Level, event.Note, event.Color, event.CreatedAt, event.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// UpdateEvent updates an existing event
func (s *PostgreSQLStore) UpdateEvent(ctx context.Context, event *Event) error {
	event.UpdatedAt = time.Now()
	if event.Color != nil && *event.Color != "" {
		normalized := normalizeColor(*event.Color)
		event.Color = &normalized
	}

	query := `
		UPDATE events 
		SET name = $3, category = $4, started_at = $5, ended_at = $6, 
		    level = $7, note = $8, color = $9, updated_at = $10
		WHERE id = $1 AND user_id = $2`

	result, err := s.db.ExecContext(ctx, query,
		event.ID, event.UserID, event.Name, event.Category, event.StartedAt, event.EndedAt,
		event.Level, event.Note, event.Color, event.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("event not found or access denied")
	}

	return nil
}

// DeleteEvent deletes an event and all its associations
func (s *PostgreSQLStore) DeleteEvent(ctx context.Context, userID, id string) error {
	query := `DELETE FROM events WHERE id = $1 AND user_id = $2`

	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("event not found or access denied")
	}

	return nil
}

// GetEvent retrieves a specific event by ID
func (s *PostgreSQLStore) GetEvent(ctx context.Context, userID, id string) (*Event, error) {
	query := `
		SELECT id, user_id, name, category, started_at, ended_at, level, note, color, created_at, updated_at
		FROM events 
		WHERE id = $1 AND user_id = $2`

	event := &Event{}
	err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&event.ID, &event.UserID, &event.Name, &event.Category, &event.StartedAt, &event.EndedAt,
		&event.Level, &event.Note, &event.Color, &event.CreatedAt, &event.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}

// ListEvents retrieves events for a user with filtering options
func (s *PostgreSQLStore) ListEvents(ctx context.Context, userID string, options EventListOptions) ([]*Event, error) {
	baseQuery := `
		SELECT DISTINCT e.id, e.user_id, e.name, e.category, e.started_at, e.ended_at, 
		       e.level, e.note, e.color, e.created_at, e.updated_at
		FROM events e`

	var args []interface{}
	var conditions []string
	argIndex := 1

	// Base condition: user owns the events
	conditions = append(conditions, fmt.Sprintf("e.user_id = $%d", argIndex))
	args = append(args, userID)
	argIndex++

	// Date filtering
	if options.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("e.started_at >= $%d", argIndex))
		args = append(args, *options.StartDate)
		argIndex++
	}
	if options.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("e.started_at <= $%d", argIndex))
		args = append(args, *options.EndDate)
		argIndex++
	}

	// Category filtering
	if options.Category != nil {
		conditions = append(conditions, fmt.Sprintf("e.category = $%d", argIndex))
		args = append(args, *options.Category)
		argIndex++
	}

	// Level filtering
	if options.LevelMin != nil {
		conditions = append(conditions, fmt.Sprintf("e.level >= $%d", argIndex))
		args = append(args, *options.LevelMin)
		argIndex++
	}
	if options.LevelMax != nil {
		conditions = append(conditions, fmt.Sprintf("e.level <= $%d", argIndex))
		args = append(args, *options.LevelMax)
		argIndex++
	}

	// Label filtering
	if len(options.Labels) > 0 {
		if options.MatchAll {
			// Match ALL labels (event must have all specified labels)
			baseQuery += ` 
				INNER JOIN event_labels el ON e.id = el.event_id 
				INNER JOIN labels l ON el.label_id = l.id`
			
			placeholders := make([]string, len(options.Labels))
			for i, label := range options.Labels {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, strings.ToLower(label))
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf("LOWER(l.name) IN (%s)", strings.Join(placeholders, ",")))
			
			// Group by event and ensure it has all labels
			baseQuery += fmt.Sprintf(" WHERE %s GROUP BY e.id, e.user_id, e.name, e.category, e.started_at, e.ended_at, e.level, e.note, e.color, e.created_at, e.updated_at HAVING COUNT(DISTINCT l.id) = %d", strings.Join(conditions, " AND "), len(options.Labels))
		} else {
			// Match ANY labels (event has at least one of the specified labels)
			baseQuery += ` 
				INNER JOIN event_labels el ON e.id = el.event_id 
				INNER JOIN labels l ON el.label_id = l.id`
			
			placeholders := make([]string, len(options.Labels))
			for i, label := range options.Labels {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, strings.ToLower(label))
				argIndex++
			}
			conditions = append(conditions, fmt.Sprintf("LOWER(l.name) IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	// Add WHERE clause if we have non-label conditions or no label filtering
	if len(options.Labels) == 0 || !options.MatchAll {
		if len(conditions) > 0 {
			baseQuery += " WHERE " + strings.Join(conditions, " AND ")
		}
	}

	// Add ordering and pagination
	baseQuery += " ORDER BY e.started_at DESC"
	if options.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, options.Limit)
		argIndex++
	}
	if options.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, options.Offset)
		argIndex++
	}

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		err := rows.Scan(
			&event.ID, &event.UserID, &event.Name, &event.Category, &event.StartedAt, &event.EndedAt,
			&event.Level, &event.Note, &event.Color, &event.CreatedAt, &event.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over events: %w", err)
	}

	return events, nil
}

// Event label assignment operations

// ListEventLabels retrieves all labels assigned to a specific event
func (s *PostgreSQLStore) ListEventLabels(ctx context.Context, userID, eventID string) ([]*Label, error) {
	query := `
		SELECT l.id, l.user_id, l.name, l.description, l.color, l.created_at, l.updated_at
		FROM labels l
		INNER JOIN event_labels el ON l.id = el.label_id
		INNER JOIN events e ON el.event_id = e.id
		WHERE el.event_id = $1 AND e.user_id = $2 AND l.user_id = $2
		ORDER BY l.name`

	rows, err := s.db.QueryContext(ctx, query, eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query event labels: %w", err)
	}
	defer rows.Close()

	var labels []*Label
	for rows.Next() {
		label := &Label{}
		err := rows.Scan(&label.ID, &label.UserID, &label.Name, &label.Description, &label.Color, &label.CreatedAt, &label.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over event labels: %w", err)
	}

	return labels, nil
}

// AssignEventLabels assigns multiple labels to an event
func (s *PostgreSQLStore) AssignEventLabels(ctx context.Context, userID, eventID string, labelIDs []string) error {
	if len(labelIDs) == 0 {
		return nil
	}

	// First verify the event belongs to the user
	_, err := s.GetEvent(ctx, userID, eventID)
	if err != nil {
		return fmt.Errorf("failed to verify event ownership: %w", err)
	}

	// Build bulk insert query
	placeholders := make([]string, len(labelIDs))
	args := []interface{}{eventID}
	argIndex := 2
	
	for i, labelID := range labelIDs {
		placeholders[i] = fmt.Sprintf("($1, $%d, NOW())", argIndex)
		args = append(args, labelID)
		argIndex++
	}

	query := fmt.Sprintf(`
		INSERT INTO event_labels (event_id, label_id, created_at) 
		VALUES %s 
		ON CONFLICT (event_id, label_id) DO NOTHING`, strings.Join(placeholders, ", "))

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to assign labels to event: %w", err)
	}

	return nil
}

// UnassignEventLabel removes a label assignment from an event
func (s *PostgreSQLStore) UnassignEventLabel(ctx context.Context, userID, eventID, labelID string) error {
	query := `
		DELETE FROM event_labels el
		USING events e, labels l
		WHERE el.event_id = e.id AND el.label_id = l.id 
		AND el.event_id = $1 AND el.label_id = $2 
		AND e.user_id = $3 AND l.user_id = $3`

	result, err := s.db.ExecContext(ctx, query, eventID, labelID, userID)
	if err != nil {
		return fmt.Errorf("failed to unassign event label: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check unassign result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("event label assignment not found or access denied")
	}

	return nil
}

// Event link operations

// CreateEventLink creates a manual link between an event and a consumption/item
func (s *PostgreSQLStore) CreateEventLink(ctx context.Context, link *EventLink) error {
	link.ID = generateUUID()
	link.CreatedAt = time.Now()

	query := `
		INSERT INTO event_links (id, event_id, consumption_id, consumption_item_id, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.ExecContext(ctx, query, link.ID, link.EventID, link.ConsumptionID, link.ConsumptionItemID, link.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "ux_event_links_") {
			return fmt.Errorf("link already exists")
		}
		return fmt.Errorf("failed to create event link: %w", err)
	}

	return nil
}

// DeleteEventLink removes a link between an event and a consumption/item
func (s *PostgreSQLStore) DeleteEventLink(ctx context.Context, userID, linkID string) error {
	query := `
		DELETE FROM event_links el
		USING events e
		WHERE el.event_id = e.id AND el.id = $1 AND e.user_id = $2`

	result, err := s.db.ExecContext(ctx, query, linkID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete event link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("event link not found or access denied")
	}

	return nil
}

// ListEventLinks retrieves all consumption/item links for an event
func (s *PostgreSQLStore) ListEventLinks(ctx context.Context, userID, eventID string) ([]*EventLink, error) {
	query := `
		SELECT el.id, el.event_id, el.consumption_id, el.consumption_item_id, el.created_at
		FROM event_links el
		INNER JOIN events e ON el.event_id = e.id
		WHERE el.event_id = $1 AND e.user_id = $2
		ORDER BY el.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query event links: %w", err)
	}
	defer rows.Close()

	var links []*EventLink
	for rows.Next() {
		link := &EventLink{}
		err := rows.Scan(&link.ID, &link.EventID, &link.ConsumptionID, &link.ConsumptionItemID, &link.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event link: %w", err)
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over event links: %w", err)
	}

	return links, nil
}