package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	// APIKeyPrefix is the prefix for all API keys
	APIKeyPrefix = "noot_"
	// APIKeySecretLength is the length of the secret part of the key (in bytes before hex encoding)
	APIKeySecretLength = 20  // Reduced to keep hex length reasonable 
	// APIKeyPrefixLength is the length of the visible prefix (not including noot_)
	APIKeyPrefixLength = 8
	// bcrypt cost for hashing API keys
	bcryptCost = 12
)

// GenerateAPIKey generates a new API key with prefix and secret
func GenerateAPIKey() (fullKey, prefix, hash string, err error) {
	// Generate random prefix (for fast lookup)
	prefixBytes := make([]byte, APIKeyPrefixLength)
	if _, err := rand.Read(prefixBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate prefix: %w", err)
	}

	// Generate random secret
	secretBytes := make([]byte, APIKeySecretLength)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate secret: %w", err)
	}

	// Encode to hex for safe handling (no special characters)
	prefixPart := strings.ToLower(fmt.Sprintf("%x", prefixBytes))[:APIKeyPrefixLength]
	secretPart := fmt.Sprintf("%x", secretBytes)

	// Build full key and visible prefix
	prefix = APIKeyPrefix + prefixPart
	fullKey = prefix + "_" + secretPart

	// Hash the full key for storage
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcryptCost)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to hash key: %w", err)
	}
	hash = string(hashBytes)

	return fullKey, prefix, hash, nil
}

// VerifyAPIKey verifies if the provided key matches the stored hash
func VerifyAPIKey(providedKey, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedKey))
	return err == nil
}

// ExtractPrefix extracts the prefix from a full API key for lookup
func ExtractPrefix(fullKey string) (string, error) {
	parts := strings.SplitN(fullKey, "_", 3)
	if len(parts) != 3 || !strings.HasPrefix(fullKey, APIKeyPrefix) {
		return "", fmt.Errorf("invalid API key format")
	}
	return parts[0] + "_" + parts[1], nil
}

// IsActive checks if an API key is currently active (not revoked or expired)
func (ak *APIKey) IsActive() bool {
	now := time.Now()

	// Check if revoked
	if ak.RevokedAt != nil {
		return false
	}

	// Check if expired
	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(now) {
		return false
	}

	return true
}

// CreateAPIKey creates a new API key for a user
func (s *PostgreSQLStore) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	if apiKey.ID == "" {
		apiKey.ID = generateUUID()
	}

	query := `
		INSERT INTO ` + TableNames.APIKeys + ` 
		(id, user_id, name, prefix, hash, scope, created_at, last_used_at, expires_at, revoked_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := s.db.ExecContext(ctx, query,
		apiKey.ID,
		apiKey.UserID,
		apiKey.Name,
		apiKey.Prefix,
		apiKey.Hash,
		apiKey.Scope,
		apiKey.CreatedAt,
		apiKey.LastUsedAt,
		apiKey.ExpiresAt,
		apiKey.RevokedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetAPIKey retrieves an API key by ID for a specific user
func (s *PostgreSQLStore) GetAPIKey(ctx context.Context, userID, id string) (*APIKey, error) {
	query := `
		SELECT id, user_id, name, prefix, hash, scope, created_at, last_used_at, expires_at, revoked_at
		FROM ` + TableNames.APIKeys + `
		WHERE id = $1 AND user_id = $2`

	var apiKey APIKey
	err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.Prefix,
		&apiKey.Hash,
		&apiKey.Scope,
		&apiKey.CreatedAt,
		&apiKey.LastUsedAt,
		&apiKey.ExpiresAt,
		&apiKey.RevokedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// GetAPIKeyByPrefix retrieves an API key by its prefix for authentication
func (s *PostgreSQLStore) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*APIKey, error) {
	query := `
		SELECT id, user_id, name, prefix, hash, scope, created_at, last_used_at, expires_at, revoked_at
		FROM ` + TableNames.APIKeys + `
		WHERE prefix = $1`

	var apiKey APIKey
	err := s.db.QueryRowContext(ctx, query, prefix).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.Prefix,
		&apiKey.Hash,
		&apiKey.Scope,
		&apiKey.CreatedAt,
		&apiKey.LastUsedAt,
		&apiKey.ExpiresAt,
		&apiKey.RevokedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by prefix: %w", err)
	}

	return &apiKey, nil
}

// ListAPIKeys lists all API keys for a user
func (s *PostgreSQLStore) ListAPIKeys(ctx context.Context, userID string) ([]*APIKey, error) {
	query := `
		SELECT id, user_id, name, prefix, hash, scope, created_at, last_used_at, expires_at, revoked_at
		FROM ` + TableNames.APIKeys + `
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*APIKey
	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Name,
			&apiKey.Prefix,
			&apiKey.Hash,
			&apiKey.Scope,
			&apiKey.CreatedAt,
			&apiKey.LastUsedAt,
			&apiKey.ExpiresAt,
			&apiKey.RevokedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, &apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over API keys: %w", err)
	}

	return apiKeys, nil
}

// UpdateAPIKey updates an existing API key
func (s *PostgreSQLStore) UpdateAPIKey(ctx context.Context, apiKey *APIKey) error {
	query := `
		UPDATE ` + TableNames.APIKeys + `
		SET name = $3, scope = $4, expires_at = $5, revoked_at = $6
		WHERE id = $1 AND user_id = $2`

	result, err := s.db.ExecContext(ctx, query,
		apiKey.ID,
		apiKey.UserID,
		apiKey.Name,
		apiKey.Scope,
		apiKey.ExpiresAt,
		apiKey.RevokedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found or not owned by user")
	}

	return nil
}

// RevokeAPIKey revokes (soft deletes) an API key
func (s *PostgreSQLStore) RevokeAPIKey(ctx context.Context, userID, id string) error {
	now := time.Now()
	query := `
		UPDATE ` + TableNames.APIKeys + `
		SET revoked_at = $3
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`

	result, err := s.db.ExecContext(ctx, query, id, userID, now)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found, not owned by user, or already revoked")
	}

	return nil
}

// UpdateAPIKeyLastUsed updates the last used timestamp for an API key
func (s *PostgreSQLStore) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	query := `
		UPDATE ` + TableNames.APIKeys + `
		SET last_used_at = $2
		WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, id, lastUsed)
	if err != nil {
		return fmt.Errorf("failed to update API key last used: %w", err)
	}

	return nil
}
