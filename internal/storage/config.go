package storage

import (
	"context"
	"fmt"
)

// Config represents database configuration that can be used for different database types
type Config struct {
	Type     string // "sqlite" or "postgres"
	Host     string // For PostgreSQL
	Port     int    // For PostgreSQL
	Database string // Database name or file path for SQLite
	Username string // For PostgreSQL
	Password string // For PostgreSQL
	SSLMode  string // For PostgreSQL
}

// NewStore creates a new store based on the configuration
func NewStore(config *Config) (Store, error) {
	switch config.Type {
	case "sqlite", "":
		return NewSQLiteStore(config.Database)
	case "postgres":
		// Future implementation for PostgreSQL
		return nil, fmt.Errorf("postgres support not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// NewStoreFromEnv creates a store from environment variables (future enhancement)
func NewStoreFromEnv() (Store, error) {
	// This would read from environment variables like:
	// DB_TYPE, DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASS, etc.
	// For now, default to SQLite
	return NewSQLiteStore("./noot.db")
}

// StoreManager provides additional store management functionality
type StoreManager interface {
	Store
	// Additional methods for store management
	Backup(ctx context.Context, destination string) error
	Restore(ctx context.Context, source string) error
	GetStats(ctx context.Context) (*StoreStats, error)
}

// StoreStats provides database statistics
type StoreStats struct {
	UserCount        int64   `json:"user_count"`
	ConsumptionCount int64   `json:"consumption_count"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	DatabaseSize     int64   `json:"database_size_bytes"`
}
