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

	// Meal operations
	CreateMeal(ctx context.Context, meal *Meal) error
	GetMeal(ctx context.Context, id string) (*Meal, error)
	GetMealsByUser(ctx context.Context, userID string, limit, offset int) ([]*Meal, error)
	GetMealsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Meal, error)

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
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
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

// ItemCache represents cached nutrition data for a food item with soft TTL
type ItemCache struct {
	ID              string    `json:"id"`
	NormalizedName  string    `json:"normalized_name"`
	NormalizedBrand string    `json:"normalized_brand"`
	DisplayName     string    `json:"display_name"`
	DisplayBrand    string    `json:"display_brand"`
	// Nutrition data per 100g in canonical units
	CaloriesPer100g       float64 `json:"calories_per_100g"`
	ProteinGPer100g       float64 `json:"protein_g_per_100g"`
	TotalFatGPer100g      float64 `json:"total_fat_g_per_100g"`
	SaturatedFatGPer100g  float64 `json:"saturated_fat_g_per_100g"`
	TransFatGPer100g      float64 `json:"trans_fat_g_per_100g"`
	CholesterolMgPer100g  float64 `json:"cholesterol_mg_per_100g"`
	SodiumMgPer100g       float64 `json:"sodium_mg_per_100g"`
	TotalCarbsGPer100g    float64 `json:"total_carbs_g_per_100g"`
	DietaryFiberGPer100g  float64 `json:"dietary_fiber_g_per_100g"`
	TotalSugarsGPer100g   float64 `json:"total_sugars_g_per_100g"`
	AddedSugarsGPer100g   float64 `json:"added_sugars_g_per_100g"`
	VitaminAMcgPer100g    float64 `json:"vitamin_a_mcg_per_100g"`
	VitaminCMgPer100g     float64 `json:"vitamin_c_mg_per_100g"`
	VitaminDMcgPer100g    float64 `json:"vitamin_d_mcg_per_100g"`
	VitaminEMgPer100g     float64 `json:"vitamin_e_mg_per_100g"`
	VitaminKMcgPer100g    float64 `json:"vitamin_k_mcg_per_100g"`
	ThiamineMgPer100g     float64 `json:"thiamine_mg_per_100g"`
	RiboflavinMgPer100g   float64 `json:"riboflavin_mg_per_100g"`
	NiacinMgPer100g       float64 `json:"niacin_mg_per_100g"`
	VitaminB6MgPer100g    float64 `json:"vitamin_b6_mg_per_100g"`
	FolateMcgPer100g      float64 `json:"folate_mcg_per_100g"`
	VitaminB12McgPer100g  float64 `json:"vitamin_b12_mcg_per_100g"`
	CalciumMgPer100g      float64 `json:"calcium_mg_per_100g"`
	IronMgPer100g         float64 `json:"iron_mg_per_100g"`
	MagnesiumMgPer100g    float64 `json:"magnesium_mg_per_100g"`
	PhosphorusMgPer100g   float64 `json:"phosphorus_mg_per_100g"`
	PotassiumMgPer100g    float64 `json:"potassium_mg_per_100g"`
	ZincMgPer100g         float64 `json:"zinc_mg_per_100g"`
	CopperMgPer100g       float64 `json:"copper_mg_per_100g"`
	ManganeseMgPer100g    float64 `json:"manganese_mg_per_100g"`
	SeleniumMcgPer100g    float64 `json:"selenium_mcg_per_100g"`
	// Soft TTL fields
	FetchedAt time.Time `json:"fetched_at"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ItemAlias represents alternative names/spellings for food items
type ItemAlias struct {
	ID             string    `json:"id"`
	AliasName      string    `json:"alias_name"`
	AliasBrand     string    `json:"alias_brand"`
	CanonicalName  string    `json:"canonical_name"`
	CanonicalBrand string    `json:"canonical_brand"`
	CreatedAt      time.Time `json:"created_at"`
}