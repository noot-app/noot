package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

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
