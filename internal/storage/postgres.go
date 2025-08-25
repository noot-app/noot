package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
)

//go:embed migrations/postgres/*.sql
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

// Migrate applies all pending migrations
func (s *PostgreSQLStore) Migrate() error {
	// Get list of migration files
	entries, err := postgresMigrationFiles.ReadDir("migrations/postgres")
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
		content, err := postgresMigrationFiles.ReadFile("migrations/postgres/" + file)
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
func (s *PostgreSQLStore) Reset() error {
	// Drop tables in reverse dependency order
	tables := GetDropTableOrder()
	for _, table := range tables {
		if _, err := s.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Re-apply migrations
	return s.Migrate()
}

// CreateUser creates a new user
func (s *PostgreSQLStore) CreateUser(ctx context.Context, user *User) error {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = generateUUID()
	}

	// Validate required fields
	if user.Handle == "" {
		return fmt.Errorf("user handle is required")
	}

	// Set created_at if not provided
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	// Set default subscription tier if not provided
	if user.SubscriptionTier == "" {
		user.SubscriptionTier = SubscriptionTierFree
	}

	query := `
		INSERT INTO users (provider, subject, email, handle, full_name, subscription_tier, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id`

	var fullName *string
	if user.FullName != nil {
		fullName = user.FullName
	}

	err := s.db.QueryRow(query, user.Provider, user.Subject, user.Email, user.Handle, fullName, user.SubscriptionTier, user.CreatedAt).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *PostgreSQLStore) GetUser(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, handle, full_name, provider, subject, email, subscription_tier, active_goal_name, avatar_url, created_at
		FROM users WHERE id = $1`

	var user User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Handle, &user.FullName, &user.Provider, &user.Subject, &user.Email,
		&user.SubscriptionTier, &user.ActiveGoalName, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserBySubject retrieves a user by provider and subject
func (s *PostgreSQLStore) GetUserBySubject(ctx context.Context, provider, subject string) (*User, error) {
	query := `
		SELECT id, handle, full_name, provider, subject, email, subscription_tier, active_goal_name, avatar_url, created_at
		FROM users WHERE provider = $1 AND subject = $2`

	var user User
	err := s.db.QueryRowContext(ctx, query, provider, subject).Scan(
		&user.ID, &user.Handle, &user.FullName, &user.Provider, &user.Subject, &user.Email,
		&user.SubscriptionTier, &user.ActiveGoalName, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by subject: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email address
func (s *PostgreSQLStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, handle, full_name, provider, subject, email, subscription_tier, active_goal_name, avatar_url, created_at
		FROM users WHERE email = $1`

	var user User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Handle, &user.FullName, &user.Provider, &user.Subject, &user.Email,
		&user.SubscriptionTier, &user.ActiveGoalName, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// CreateConsumption creates a new consumption
func (s *PostgreSQLStore) CreateConsumption(ctx context.Context, consumption *Consumption) error {
	// Generate UUID if not provided
	if consumption.ID == "" {
		consumption.ID = generateUUID()
	}

	// Set created_at if not provided
	if consumption.CreatedAt.IsZero() {
		consumption.CreatedAt = time.Now()
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
	now := time.Now()
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

// Placeholder implementations for the remaining Store interface methods
// These would need to be implemented similarly to the SQLite versions

func (s *PostgreSQLStore) GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*NutritionSummary, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetNutritionSummary not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) CreateConsumptionItem(ctx context.Context, item *ConsumptionItem) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("CreateConsumptionItem not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetConsumptionItems(ctx context.Context, consumptionID string) ([]*ConsumptionItem, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetConsumptionItems not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) UpdateConsumptionItem(ctx context.Context, item *ConsumptionItem) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("UpdateConsumptionItem not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) DeleteConsumptionItem(ctx context.Context, id string) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("DeleteConsumptionItem not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) DeleteConsumptionItemsByConsumption(ctx context.Context, consumptionID string) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("DeleteConsumptionItemsByConsumption not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) CreateItem(ctx context.Context, item *Item) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("CreateItem not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetItem(ctx context.Context, id string) (*Item, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetItem not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetItemByName(ctx context.Context, normalizedName, normalizedBrand string) (*Item, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetItemByName not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) UpdateItem(ctx context.Context, item *Item) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("UpdateItem not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetStaleItems(ctx context.Context, staleAfter time.Time) ([]*Item, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetStaleItems not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) UpsertUserGoal(ctx context.Context, goal *UserGoal) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("UpsertUserGoal not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetUserGoal(ctx context.Context, userID, name string) (*UserGoal, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetUserGoal not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetUserGoals(ctx context.Context, userID string) ([]*UserGoal, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetUserGoals not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) DeleteUserGoal(ctx context.Context, userID, name string) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("DeleteUserGoal not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) SetActiveGoal(ctx context.Context, userID, goalName string) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("SetActiveGoal not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) ClearActiveGoal(ctx context.Context, userID string) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("ClearActiveGoal not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetActiveGoalName(ctx context.Context, userID string) (*string, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetActiveGoalName not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) UpsertUserBiometrics(ctx context.Context, biometrics *UserBiometrics) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("UpsertUserBiometrics not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetUserBiometrics(ctx context.Context, userID string) (*UserBiometrics, error) {
	// TODO: Implement PostgreSQL version
	return nil, fmt.Errorf("GetUserBiometrics not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) DeleteUserBiometrics(ctx context.Context, userID string) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("DeleteUserBiometrics not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) CreateItemAlias(ctx context.Context, alias *ItemAlias) error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("CreateItemAlias not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) GetCanonicalName(ctx context.Context, aliasName, aliasBrand string) (canonicalName, canonicalBrand string, err error) {
	// TODO: Implement PostgreSQL version
	return "", "", fmt.Errorf("GetCanonicalName not yet implemented for PostgreSQL")
}

func (s *PostgreSQLStore) Seed() error {
	// TODO: Implement PostgreSQL version
	return fmt.Errorf("Seed not yet implemented for PostgreSQL")
}
