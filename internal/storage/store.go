package storage

import (
	"context"
	"time"
)

// Store defines the interface for data persistence
type Store interface {
	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserBySubject(ctx context.Context, provider, subject string) (*User, error)
	DeleteUserCascade(ctx context.Context, id string) error

	// Meal operations
	CreateMeal(ctx context.Context, meal *Meal) error
	GetMeal(ctx context.Context, id string) (*Meal, error)
	GetMealsByUser(ctx context.Context, userID string, limit, offset int) ([]*Meal, error)
	GetMealsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Meal, error)

	// Items cache operations
	GetCachedItem(ctx context.Context, normalizedName, normalizedBrand string) (*CachedItem, error)
	UpsertCachedItem(ctx context.Context, item *CachedItem) error

	// Database lifecycle
	Close() error
	Migrate() error
	Seed() error
	Reset() error
}

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Subject   string    `json:"subject"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Meal represents a logged meal with nutrition data
type Meal struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Transcript    string    `json:"transcript"`
	ItemsJSON     string    `json:"items_json"` // JSON serialized items array
	TotalCalories float64   `json:"total_calories"`
	TotalProtein  float64   `json:"total_protein_g"`
	TotalFat      float64   `json:"total_fat_g"`
	TotalCarbs    float64   `json:"total_carbs_g"`
	TotalFiber    float64   `json:"total_fiber_g"`
	TotalSodium   float64   `json:"total_sodium_mg"`
	CreatedAt     time.Time `json:"created_at"`
}

// CachedItem represents a cached nutrition item with soft TTL
type CachedItem struct {
	ID               string    `json:"id"`
	NormalizedName   string    `json:"normalized_name"`
	NormalizedBrand  string    `json:"normalized_brand"`
	OriginalName     string    `json:"original_name"`
	OriginalBrand    string    `json:"original_brand"`
	NutrientDataJSON string    `json:"nutrient_data_json"` // JSON serialized nutrition data
	FetchedAt        time.Time `json:"fetched_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// ItemAlias represents alternative names for cached items
type ItemAlias struct {
	ID              string `json:"id"`
	CachedItemID    string `json:"cached_item_id"`
	AliasName       string `json:"alias_name"`
	NormalizedAlias string `json:"normalized_alias"`
}
