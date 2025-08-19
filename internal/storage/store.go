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

	// Consumption operations
	CreateConsumption(ctx context.Context, consumption *Consumption) error
	GetConsumption(ctx context.Context, id string) (*Consumption, error)
	GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*Consumption, error)
	GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Consumption, error)
	GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*NutritionSummary, error)

	// User goal operations  
	UpsertUserGoal(ctx context.Context, goal *UserGoal) error
	GetUserGoal(ctx context.Context, userID, name string) (*UserGoal, error)
	DeleteUserGoal(ctx context.Context, userID, name string) error

	// Item cache operations (soft TTL)
	GetItemFromCache(ctx context.Context, normalizedName, normalizedBrand string) (*ItemCache, error)
	UpsertItemCache(ctx context.Context, item *ItemCache) error
	RefreshItemCache(ctx context.Context, normalizedName, normalizedBrand string, item *ItemCache) error
	IsItemCacheExpired(item *ItemCache) bool

	// Item alias operations
	CreateItemAlias(ctx context.Context, alias *ItemAlias) error
	GetCanonicalName(ctx context.Context, aliasName, aliasBrand string) (canonicalName, canonicalBrand string, err error)

	// Database lifecycle
	Close() error
	Migrate() error
	Seed() error
	Reset() error
}

// User represents a user in the system
type User struct {
	ID               string     `json:"id"`
	Provider         string     `json:"provider"`
	Subject          string     `json:"subject"`
	Email            string     `json:"email"`
	SubscriptionTier string     `json:"subscription_tier"` // "free", "pro"
	Sex              string     `json:"sex"`               // "male", "female", "unspecified"
	BirthDate        *time.Time `json:"birth_date"`        // nullable for age-based DRI calculation
	CreatedAt        time.Time  `json:"created_at"`
}

// Consumption represents a logged consumption with nutrition data
type Consumption struct {
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
	// Additional micronutrient totals
	SaturatedFat   float64   `json:"saturated_fat_g"`
	TransFat       float64   `json:"trans_fat_g"`
	Cholesterol    float64   `json:"cholesterol_mg"`
	TotalSugars    float64   `json:"total_sugars_g"`
	AddedSugars    float64   `json:"added_sugars_g"`
	VitaminA       float64   `json:"vitamin_a_mcg"`
	VitaminC       float64   `json:"vitamin_c_mg"`
	VitaminD       float64   `json:"vitamin_d_mcg"`
	VitaminE       float64   `json:"vitamin_e_mg"`
	VitaminK       float64   `json:"vitamin_k_mcg"`
	Thiamine       float64   `json:"thiamine_mg"`
	Riboflavin     float64   `json:"riboflavin_mg"`
	Niacin         float64   `json:"niacin_mg"`
	VitaminB6      float64   `json:"vitamin_b6_mg"`
	Folate         float64   `json:"folate_mcg"`
	VitaminB12     float64   `json:"vitamin_b12_mcg"`
	Calcium        float64   `json:"calcium_mg"`
	Iron           float64   `json:"iron_mg"`
	Magnesium      float64   `json:"magnesium_mg"`
	Phosphorus     float64   `json:"phosphorus_mg"`
	Potassium      float64   `json:"potassium_mg"`
	Zinc           float64   `json:"zinc_mg"`
	Copper         float64   `json:"copper_mg"`
	Manganese      float64   `json:"manganese_mg"`
	Selenium       float64   `json:"selenium_mcg"`
	CreatedAt      time.Time `json:"created_at"`
}

// ItemCache represents cached nutrition data for a food item with soft TTL
type ItemCache struct {
	ID              string `json:"id"`
	NormalizedName  string `json:"normalized_name"`
	NormalizedBrand string `json:"normalized_brand"`
	DisplayName     string `json:"display_name"`
	DisplayBrand    string `json:"display_brand"`
	// Nutrition data per 100g in canonical units
	CaloriesPer100g      float64 `json:"calories_per_100g"`
	ProteinGPer100g      float64 `json:"protein_g_per_100g"`
	TotalFatGPer100g     float64 `json:"total_fat_g_per_100g"`
	SaturatedFatGPer100g float64 `json:"saturated_fat_g_per_100g"`
	TransFatGPer100g     float64 `json:"trans_fat_g_per_100g"`
	CholesterolMgPer100g float64 `json:"cholesterol_mg_per_100g"`
	SodiumMgPer100g      float64 `json:"sodium_mg_per_100g"`
	TotalCarbsGPer100g   float64 `json:"total_carbs_g_per_100g"`
	DietaryFiberGPer100g float64 `json:"dietary_fiber_g_per_100g"`
	TotalSugarsGPer100g  float64 `json:"total_sugars_g_per_100g"`
	AddedSugarsGPer100g  float64 `json:"added_sugars_g_per_100g"`
	VitaminAMcgPer100g   float64 `json:"vitamin_a_mcg_per_100g"`
	VitaminCMgPer100g    float64 `json:"vitamin_c_mg_per_100g"`
	VitaminDMcgPer100g   float64 `json:"vitamin_d_mcg_per_100g"`
	VitaminEMgPer100g    float64 `json:"vitamin_e_mg_per_100g"`
	VitaminKMcgPer100g   float64 `json:"vitamin_k_mcg_per_100g"`
	ThiamineMgPer100g    float64 `json:"thiamine_mg_per_100g"`
	RiboflavinMgPer100g  float64 `json:"riboflavin_mg_per_100g"`
	NiacinMgPer100g      float64 `json:"niacin_mg_per_100g"`
	VitaminB6MgPer100g   float64 `json:"vitamin_b6_mg_per_100g"`
	FolateMcgPer100g     float64 `json:"folate_mcg_per_100g"`
	VitaminB12McgPer100g float64 `json:"vitamin_b12_mcg_per_100g"`
	CalciumMgPer100g     float64 `json:"calcium_mg_per_100g"`
	IronMgPer100g        float64 `json:"iron_mg_per_100g"`
	MagnesiumMgPer100g   float64 `json:"magnesium_mg_per_100g"`
	PhosphorusMgPer100g  float64 `json:"phosphorus_mg_per_100g"`
	PotassiumMgPer100g   float64 `json:"potassium_mg_per_100g"`
	ZincMgPer100g        float64 `json:"zinc_mg_per_100g"`
	CopperMgPer100g      float64 `json:"copper_mg_per_100g"`
	ManganeseMgPer100g   float64 `json:"manganese_mg_per_100g"`
	SeleniumMcgPer100g   float64 `json:"selenium_mcg_per_100g"`
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

// NutritionSummary represents aggregated nutrition data over a time period
type NutritionSummary struct {
	UserID           string    `json:"user_id"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	ConsumptionCount int       `json:"consumption_count"`

	// Totals for the time period
	TotalCalories float64 `json:"total_calories"`
	TotalProtein  float64 `json:"total_protein_g"`
	TotalFat      float64 `json:"total_fat_g"`
	TotalCarbs    float64 `json:"total_carbs_g"`
	TotalFiber    float64 `json:"total_fiber_g"`
	TotalSodium   float64 `json:"total_sodium_mg"`

	// Averages per day
	AvgCaloriesPerDay float64 `json:"avg_calories_per_day"`
	AvgProteinPerDay  float64 `json:"avg_protein_per_day"`
	AvgFatPerDay      float64 `json:"avg_fat_per_day"`
	AvgCarbsPerDay    float64 `json:"avg_carbs_per_day"`
	AvgFiberPerDay    float64 `json:"avg_fiber_per_day"`
	AvgSodiumPerDay   float64 `json:"avg_sodium_per_day"`

	// Daily breakdown for charts
	DailyBreakdown []DailySummary `json:"daily_breakdown"`
}

// DailySummary represents nutrition data for a single day
type DailySummary struct {
	Date             time.Time `json:"date"`
	ConsumptionCount int       `json:"consumption_count"`
	Calories         float64   `json:"calories"`
	Protein          float64   `json:"protein_g"`
	Fat              float64   `json:"total_fat_g"`
	Carbs            float64   `json:"total_carbs_g"`
	Fiber            float64   `json:"fiber_g"`
	Sodium           float64   `json:"sodium_mg"`
}

// UserGoal represents custom nutrition goals for Pro users
type UserGoal struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`          // typically "custom", allows for future goal presets
	OverridesJSON string    `json:"overrides_json"` // JSON map of nutrient_key -> target value
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
