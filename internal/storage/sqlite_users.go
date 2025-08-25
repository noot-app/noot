package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

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
