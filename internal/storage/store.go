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
	UpdateConsumption(ctx context.Context, consumption *Consumption) error
	DeleteConsumption(ctx context.Context, id string) error
	GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*Consumption, error)
	GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Consumption, error)
	GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*NutritionSummary, error)

	// ConsumptionItem operations
	CreateConsumptionItem(ctx context.Context, item *ConsumptionItem) error
	GetConsumptionItems(ctx context.Context, consumptionID string) ([]*ConsumptionItem, error)
	UpdateConsumptionItem(ctx context.Context, item *ConsumptionItem) error
	DeleteConsumptionItem(ctx context.Context, id string) error
	DeleteConsumptionItemsByConsumption(ctx context.Context, consumptionID string) error

	// Item operations (evolved from ItemCache)
	CreateItem(ctx context.Context, item *Item) error
	GetItem(ctx context.Context, id string) (*Item, error)
	GetItemByName(ctx context.Context, normalizedName, normalizedBrand string) (*Item, error)
	UpdateItem(ctx context.Context, item *Item) error
	GetStaleItems(ctx context.Context, staleAfter time.Time) ([]*Item, error) // For 30-day refresh logic

	// User goal operations
	UpsertUserGoal(ctx context.Context, goal *UserGoal) error
	GetUserGoal(ctx context.Context, userID, name string) (*UserGoal, error)
	GetUserGoals(ctx context.Context, userID string) ([]*UserGoal, error)
	DeleteUserGoal(ctx context.Context, userID, name string) error
	SetActiveGoal(ctx context.Context, userID, goalName string) error
	GetActiveGoalName(ctx context.Context, userID string) (*string, error)

	// User biometrics operations
	UpsertUserBiometrics(ctx context.Context, biometrics *UserBiometrics) error
	GetUserBiometrics(ctx context.Context, userID string) (*UserBiometrics, error)
	DeleteUserBiometrics(ctx context.Context, userID string) error

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
	ID               string    `json:"id"`
	Provider         string    `json:"provider"`
	Subject          string    `json:"subject"`
	Email            string    `json:"email"`
	SubscriptionTier string    `json:"subscription_tier"` // "free", "pro"
	ActiveGoalName   *string   `json:"active_goal_name"`  // Name of the active goal set (Pro users only)
	CreatedAt        time.Time `json:"created_at"`
}

// UserBiometrics represents user physical and demographic data
type UserBiometrics struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	BirthDate     *time.Time `json:"birth_date"`
	Sex           string     `json:"sex"` // "male", "female", "other", "prefer_not_to_say"
	HeightCm      *float64   `json:"height_cm"`
	WeightKg      *float64   `json:"weight_kg"`
	ActivityLevel string     `json:"activity_level"` // "sedentary", "lightly_active", etc.
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Consumption represents a logged consumption with nutrition data
type Consumption struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Transcript    string    `json:"transcript"`
	TotalCalories float64   `json:"total_calories"`
	TotalProtein  float64  `json:"total_protein_g"`
	TotalFat      float64  `json:"total_fat_g"`
	TotalCarbs    float64  `json:"total_carbs_g"`
	DietaryFiber  float64  `json:"dietary_fiber_g"`
	TotalSodium   float64  `json:"total_sodium_mg"`
	// Additional micronutrient totals
	SaturatedFat    float64    `json:"saturated_fat_g"`
	TransFat        float64    `json:"trans_fat_g"`
	Cholesterol     float64    `json:"cholesterol_mg"`
	TotalSugars     float64    `json:"total_sugars_g"`
	AddedSugars     float64    `json:"added_sugars_g"`
	VitaminA        float64    `json:"vitamin_a_mcg"`
	VitaminC        float64    `json:"vitamin_c_mg"`
	VitaminD        float64    `json:"vitamin_d_mcg"`
	VitaminE        float64    `json:"vitamin_e_mg"`
	VitaminK        float64    `json:"vitamin_k_mcg"`
	Thiamine        float64    `json:"thiamine_mg"`
	Riboflavin      float64    `json:"riboflavin_mg"`
	Niacin          float64    `json:"niacin_mg"`
	VitaminB6       float64    `json:"vitamin_b6_mg"`
	Folate          float64    `json:"folate_mcg"`
	VitaminB12      float64    `json:"vitamin_b12_mcg"`
	Biotin          float64    `json:"biotin_mcg"`
	PantothenicAcid float64    `json:"pantothenic_acid_mg"`
	Choline         float64    `json:"choline_mg"`
	Calcium         float64    `json:"calcium_mg"`
	Iron            float64    `json:"iron_mg"`
	Magnesium       float64    `json:"magnesium_mg"`
	Phosphorus      float64    `json:"phosphorus_mg"`
	Potassium       float64    `json:"potassium_mg"`
	Zinc            float64    `json:"zinc_mg"`
	Copper          float64    `json:"copper_mg"`
	Manganese       float64    `json:"manganese_mg"`
	Selenium        float64    `json:"selenium_mcg"`
	Iodine          float64    `json:"iodine_mcg"`
	Molybdenum      float64    `json:"molybdenum_mcg"`
	Chromium        float64    `json:"chromium_mcg"`
	Fluoride        float64    `json:"fluoride_mg"`
	Chloride        float64    `json:"chloride_mg"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

// Item represents permanent nutrition data for a food item (evolved from ItemCache)
type Item struct {
	ID              string `json:"id"`
	NormalizedName  string `json:"normalized_name"`
	NormalizedBrand string `json:"normalized_brand"`
	DisplayName     string `json:"display_name"`
	DisplayBrand    string `json:"display_brand"`
	// Nutrition data per 100g
	CaloriesPer100g              float64 `json:"calories_per_100g"`
	ProteinGPer100g              float64 `json:"protein_g_per_100g"`
	TotalFatGPer100g             float64 `json:"total_fat_g_per_100g"`
	SaturatedFatGPer100g         float64 `json:"saturated_fat_g_per_100g"`
	TransFatGPer100g             float64 `json:"trans_fat_g_per_100g"`
	CholesterolMgPer100g         float64 `json:"cholesterol_mg_per_100g"`
	SodiumMgPer100g              float64 `json:"sodium_mg_per_100g"`
	TotalCarbsGPer100g           float64 `json:"total_carbs_g_per_100g"`
	DietaryFiberGPer100g         float64 `json:"dietary_fiber_g_per_100g"`
	TotalSugarsGPer100g          float64 `json:"total_sugars_g_per_100g"`
	AddedSugarsGPer100g          float64 `json:"added_sugars_g_per_100g"`
	VitaminAMcgPer100g           float64 `json:"vitamin_a_mcg_per_100g"`
	VitaminCMgPer100g            float64 `json:"vitamin_c_mg_per_100g"`
	VitaminDMcgPer100g           float64 `json:"vitamin_d_mcg_per_100g"`
	VitaminEMgPer100g            float64 `json:"vitamin_e_mg_per_100g"`
	VitaminKMcgPer100g           float64 `json:"vitamin_k_mcg_per_100g"`
	ThiamineMgPer100g            float64 `json:"thiamine_mg_per_100g"`
	RiboflavinMgPer100g          float64 `json:"riboflavin_mg_per_100g"`
	NiacinMgPer100g              float64 `json:"niacin_mg_per_100g"`
	VitaminB6MgPer100g           float64 `json:"vitamin_b6_mg_per_100g"`
	FolateMcgPer100g             float64 `json:"folate_mcg_per_100g"`
	VitaminB12McgPer100g         float64 `json:"vitamin_b12_mcg_per_100g"`
	BiotinMcgPer100g             float64 `json:"biotin_mcg_per_100g"`
	PantothenicAcidMgPer100g     float64 `json:"pantothenic_acid_mg_per_100g"`
	CholineMgPer100g             float64 `json:"choline_mg_per_100g"`
	CalciumMgPer100g             float64 `json:"calcium_mg_per_100g"`
	IronMgPer100g                float64 `json:"iron_mg_per_100g"`
	MagnesiumMgPer100g           float64 `json:"magnesium_mg_per_100g"`
	PhosphorusMgPer100g          float64 `json:"phosphorus_mg_per_100g"`
	PotassiumMgPer100g           float64 `json:"potassium_mg_per_100g"`
	ZincMgPer100g                float64 `json:"zinc_mg_per_100g"`
	CopperMgPer100g              float64 `json:"copper_mg_per_100g"`
	ManganeseMgPer100g           float64 `json:"manganese_mg_per_100g"`
	SeleniumMcgPer100g           float64 `json:"selenium_mcg_per_100g"`
	IodineMcgPer100g             float64 `json:"iodine_mcg_per_100g"`
	MolybdenumMcgPer100g         float64 `json:"molybdenum_mcg_per_100g"`
	ChromiumMcgPer100g           float64 `json:"chromium_mcg_per_100g"`
	FluorideMgPer100g            float64 `json:"fluoride_mg_per_100g"`
	ChlorideMgPer100g            float64 `json:"chloride_mg_per_100g"`
	// Timestamps for 30-day refresh logic
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConsumptionItem represents the relationship between a consumption and an item
type ConsumptionItem struct {
	ID             string     `json:"id"`
	ConsumptionID  string     `json:"consumption_id"`
	ItemID         string     `json:"item_id"`
	Grams          float64    `json:"grams"`          // Actual grams consumed (normalized internally)
	UserQuantity   *float64   `json:"user_quantity"`  // Original user input quantity for display
	UserUnit       *string    `json:"user_unit"`      // Original user input unit for display
	CreatedAt      time.Time  `json:"created_at"`
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

	// Totals for the time period - basic macronutrients
	TotalCalories float64 `json:"total_calories"`
	TotalProtein  float64 `json:"total_protein_g"`
	TotalFat      float64 `json:"total_fat_g"`
	TotalCarbs    float64 `json:"total_carbs_g"`
	DietaryFiber  float64 `json:"total_fiber_g"`
	TotalSodium   float64 `json:"total_sodium_mg"`

	// Additional macronutrients
	TotalSaturatedFat float64 `json:"total_saturated_fat_g"`
	TotalTransFat     float64 `json:"total_trans_fat_g"`
	TotalCholesterol  float64 `json:"total_cholesterol_mg"`
	TotalSugars       float64 `json:"total_sugars_g"`
	TotalAddedSugars  float64 `json:"total_added_sugars_g"`

	// Vitamins
	TotalVitaminA        float64 `json:"total_vitamin_a_mcg"`
	TotalVitaminC        float64 `json:"total_vitamin_c_mg"`
	TotalVitaminD        float64 `json:"total_vitamin_d_mcg"`
	TotalVitaminE        float64 `json:"total_vitamin_e_mg"`
	TotalVitaminK        float64 `json:"total_vitamin_k_mcg"`
	TotalThiamine        float64 `json:"total_thiamine_mg"`
	TotalRiboflavin      float64 `json:"total_riboflavin_mg"`
	TotalNiacin          float64 `json:"total_niacin_mg"`
	TotalVitaminB6       float64 `json:"total_vitamin_b6_mg"`
	TotalFolate          float64 `json:"total_folate_mcg"`
	TotalVitaminB12      float64 `json:"total_vitamin_b12_mcg"`
	TotalBiotin          float64 `json:"total_biotin_mcg"`
	TotalPantothenicAcid float64 `json:"total_pantothenic_acid_mg"`
	TotalCholine         float64 `json:"total_choline_mg"`

	// Minerals
	TotalCalcium    float64 `json:"total_calcium_mg"`
	TotalIron       float64 `json:"total_iron_mg"`
	TotalMagnesium  float64 `json:"total_magnesium_mg"`
	TotalPhosphorus float64 `json:"total_phosphorus_mg"`
	TotalPotassium  float64 `json:"total_potassium_mg"`
	TotalZinc       float64 `json:"total_zinc_mg"`
	TotalCopper     float64 `json:"total_copper_mg"`
	TotalManganese  float64 `json:"total_manganese_mg"`
	TotalSelenium   float64 `json:"total_selenium_mcg"`
	TotalIodine     float64 `json:"total_iodine_mcg"`
	TotalMolybdenum float64 `json:"total_molybdenum_mcg"`
	TotalChromium   float64 `json:"total_chromium_mcg"`
	TotalFluoride   float64 `json:"total_fluoride_mg"`
	TotalChloride   float64 `json:"total_chloride_mg"`

	// Averages per day - basic macronutrients
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
	Name          string    `json:"name"`           // typically "custom", allows for future goal presets
	OverridesJSON string    `json:"overrides_json"` // JSON map of nutrient_key -> target value
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
