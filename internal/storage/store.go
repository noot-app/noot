package storage

import (
	"context"
	"time"
)

// Store defines the interface for data persistence
type Store interface {
	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id int64) (*User, error)
	GetUserBySubject(ctx context.Context, provider, subject string) (*User, error)

	// Meal operations
	CreateMeal(ctx context.Context, meal *Meal) error
	GetMeal(ctx context.Context, id int64) (*Meal, error)
	GetMealsByUser(ctx context.Context, userID int64, limit, offset int) ([]*Meal, error)
	GetMealsByUserSince(ctx context.Context, userID int64, since time.Time) ([]*Meal, error)

	// Database lifecycle
	Close() error
	Migrate() error
	Seed() error
	Reset() error
}

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	Provider  string    `json:"provider"`
	Subject   string    `json:"subject"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Meal represents a logged meal with nutrition data
type Meal struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Transcript      string    `json:"transcript"`
	ItemsJSON       string    `json:"items_json"` // JSON serialized items array
	TotalCalories   float64   `json:"total_calories"`
	TotalProtein    float64   `json:"total_protein_g"`
	TotalFat        float64   `json:"total_fat_g"`
	TotalCarbs      float64   `json:"total_carbs_g"`
	TotalFiber      float64   `json:"total_fiber_g"`
	TotalSodium     float64   `json:"total_sodium_mg"`
	CreatedAt       time.Time `json:"created_at"`
}