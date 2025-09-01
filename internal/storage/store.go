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
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error

	// Consumption operations
	CreateConsumption(ctx context.Context, consumption *Consumption) error
	GetConsumption(ctx context.Context, id string) (*Consumption, error)
	GetConsumptionForUser(ctx context.Context, userID, id string) (*Consumption, error)
	GetPublicConsumption(ctx context.Context, id string) (*Consumption, error)
	UpdateConsumption(ctx context.Context, consumption *Consumption) error
	DeleteConsumption(ctx context.Context, id string) error
	GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*Consumption, error)
	GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*Consumption, error)
	GetConsumptionsByUserDateRange(ctx context.Context, userID string, start, end time.Time, limit, offset int) ([]*Consumption, error)
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
	ClearActiveGoal(ctx context.Context, userID string) error
	GetActiveGoalName(ctx context.Context, userID string) (*string, error)

	// User biometrics operations
	UpsertUserBiometrics(ctx context.Context, biometrics *UserBiometrics) error
	GetUserBiometrics(ctx context.Context, userID string) (*UserBiometrics, error)
	DeleteUserBiometrics(ctx context.Context, userID string) error

	// Item alias operations
	CreateItemAlias(ctx context.Context, alias *ItemAlias) error
	GetCanonicalName(ctx context.Context, aliasName, aliasBrand string) (canonicalName, canonicalBrand string, err error)

	// Label operations
	CreateLabel(ctx context.Context, label *Label) error
	UpdateLabel(ctx context.Context, label *Label) error
	DeleteLabel(ctx context.Context, userID, id string) error
	GetLabel(ctx context.Context, userID, id string) (*Label, error)
	ListLabels(ctx context.Context, userID string) ([]*LabelWithUsage, error)

	// Label assignment operations
	ListConsumptionLabels(ctx context.Context, userID, consumptionID string) ([]*Label, error)
	AssignConsumptionLabels(ctx context.Context, userID, consumptionID string, labelIDs []string) error
	UnassignConsumptionLabel(ctx context.Context, userID, consumptionID, labelID string) error
	ListConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string) ([]*Label, error)
	AssignConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string, labelIDs []string) error
	UnassignConsumptionItemLabel(ctx context.Context, userID, consumptionItemID, labelID string) error

	// Filtering operations
	GetConsumptionsByLabels(ctx context.Context, userID string, labelNames []string, matchAll bool, limit, offset int) ([]*Consumption, error)

	// Event operations
	CreateEvent(ctx context.Context, event *Event) error
	UpdateEvent(ctx context.Context, event *Event) error
	DeleteEvent(ctx context.Context, userID, id string) error
	GetEvent(ctx context.Context, userID, id string) (*Event, error)
	ListEvents(ctx context.Context, userID string, options EventListOptions) ([]*Event, error)

	// Event label assignment operations
	ListEventLabels(ctx context.Context, userID, eventID string) ([]*Label, error)
	AssignEventLabels(ctx context.Context, userID, eventID string, labelIDs []string) error
	UnassignEventLabel(ctx context.Context, userID, eventID, labelID string) error

	// Event link operations
	CreateEventLink(ctx context.Context, link *EventLink) error
	DeleteEventLink(ctx context.Context, userID, linkID string) error
	ListEventLinks(ctx context.Context, userID, eventID string) ([]*EventLink, error)

	// Database lifecycle
	Close() error
}

// User represents a user in the system
type User struct {
	ID               string    `json:"id"`                // Auth provider's user ID (e.g., Supabase auth.users.id)
	Handle           string    `json:"handle"`            // Unique username/handle
	FullName         *string   `json:"full_name"`         // Optional display name
	Email            string    `json:"email"`             // Email address
	SubscriptionTier string    `json:"subscription_tier"` // "free", "pro"
	ActiveGoalName   *string   `json:"active_goal_name"`  // Name of the active goal set (Pro users only)
	CreatedAt        time.Time `json:"created_at"`
	AvatarURL        *string   `json:"avatar_url"` // Optional profile image URL
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
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	Transcript    string  `json:"transcript"`
	TotalCalories float64 `json:"total_calories"`
	TotalProtein  float64 `json:"total_protein_g"`
	TotalFat      float64 `json:"total_fat_g"`
	TotalCarbs    float64 `json:"total_carbs_g"`
	DietaryFiber  float64 `json:"dietary_fiber_g"`
	TotalSodium   float64 `json:"total_sodium_mg"`
	// Additional micronutrient totals
	SaturatedFat       float64    `json:"saturated_fat_g"`
	TransFat           float64    `json:"trans_fat_g"`
	Cholesterol        float64    `json:"cholesterol_mg"`
	TotalSugars        float64    `json:"total_sugars_g"`
	AddedSugars        float64    `json:"added_sugars_g"`
	VitaminA           float64    `json:"vitamin_a_mcg"`
	VitaminC           float64    `json:"vitamin_c_mg"`
	VitaminD           float64    `json:"vitamin_d_mcg"`
	VitaminE           float64    `json:"vitamin_e_mg"`
	VitaminK           float64    `json:"vitamin_k_mcg"`
	Thiamine           float64    `json:"thiamine_mg"`
	Riboflavin         float64    `json:"riboflavin_mg"`
	Niacin             float64    `json:"niacin_mg"`
	VitaminB6          float64    `json:"vitamin_b6_mg"`
	Folate             float64    `json:"folate_mcg"`
	VitaminB12         float64    `json:"vitamin_b12_mcg"`
	Biotin             float64    `json:"biotin_mcg"`
	PantothenicAcid    float64    `json:"pantothenic_acid_mg"`
	Choline            float64    `json:"choline_mg"`
	Calcium            float64    `json:"calcium_mg"`
	Iron               float64    `json:"iron_mg"`
	Magnesium          float64    `json:"magnesium_mg"`
	Phosphorus         float64    `json:"phosphorus_mg"`
	Potassium          float64    `json:"potassium_mg"`
	Zinc               float64    `json:"zinc_mg"`
	Copper             float64    `json:"copper_mg"`
	Manganese          float64    `json:"manganese_mg"`
	Selenium           float64    `json:"selenium_mcg"`
	Iodine             float64    `json:"iodine_mcg"`
	Molybdenum         float64    `json:"molybdenum_mcg"`
	Chromium           float64    `json:"chromium_mcg"`
	Fluoride           float64    `json:"fluoride_mg"`
	Chloride           float64    `json:"chloride_mg"`
	Omega3Ala          float64    `json:"omega3_ala_g"`
	Omega3Epa          float64    `json:"omega3_epa_g"`
	Omega3Dha          float64    `json:"omega3_dha_g"`
	Omega6             float64    `json:"omega6_g"`
	Creatine           float64    `json:"creatine_mg"`
	Caffeine           float64    `json:"caffeine_mg"`
	Alcohol            float64    `json:"alcohol_g"`
	PolyunsaturatedFat float64    `json:"polyunsaturated_fat_g"`
	MonounsaturatedFat float64    `json:"monounsaturated_fat_g"`
	Note               *string    `json:"note,omitempty"`
	Labels             []*Label   `json:"labels,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
}

// Item represents permanent nutrition data for a food item (evolved from ItemCache)
type Item struct {
	ID              string `json:"id"`
	NormalizedName  string `json:"normalized_name"`
	NormalizedBrand string `json:"normalized_brand"`
	DisplayName     string `json:"display_name"`
	DisplayBrand    string `json:"display_brand"`

	// Original serving data (exact values for one serving of the item - ex: one can of soda, one burger, one carrot, one handful of blueberries, one plate of pasta with pesto sauce, etc)
	OriginalServingGrams        *float64 `json:"original_serving_grams,omitempty"`
	OriginalCalories            *float64 `json:"original_calories,omitempty"`
	OriginalProteinG            *float64 `json:"original_protein_g,omitempty"`
	OriginalTotalFatG           *float64 `json:"original_total_fat_g,omitempty"`
	OriginalSaturatedFatG       *float64 `json:"original_saturated_fat_g,omitempty"`
	OriginalTransFatG           *float64 `json:"original_trans_fat_g,omitempty"`
	OriginalCholesterolMg       *float64 `json:"original_cholesterol_mg,omitempty"`
	OriginalSodiumMg            *float64 `json:"original_sodium_mg,omitempty"`
	OriginalTotalCarbsG         *float64 `json:"original_total_carbs_g,omitempty"`
	OriginalDietaryFiberG       *float64 `json:"original_dietary_fiber_g,omitempty"`
	OriginalTotalSugarsG        *float64 `json:"original_total_sugars_g,omitempty"`
	OriginalAddedSugarsG        *float64 `json:"original_added_sugars_g,omitempty"`
	OriginalVitaminAMcg         *float64 `json:"original_vitamin_a_mcg,omitempty"`
	OriginalVitaminCMg          *float64 `json:"original_vitamin_c_mg,omitempty"`
	OriginalVitaminDMcg         *float64 `json:"original_vitamin_d_mcg,omitempty"`
	OriginalVitaminEMg          *float64 `json:"original_vitamin_e_mg,omitempty"`
	OriginalVitaminKMcg         *float64 `json:"original_vitamin_k_mcg,omitempty"`
	OriginalThiamineMg          *float64 `json:"original_thiamine_mg,omitempty"`
	OriginalRiboflavinMg        *float64 `json:"original_riboflavin_mg,omitempty"`
	OriginalNiacinMg            *float64 `json:"original_niacin_mg,omitempty"`
	OriginalVitaminB6Mg         *float64 `json:"original_vitamin_b6_mg,omitempty"`
	OriginalFolateMcg           *float64 `json:"original_folate_mcg,omitempty"`
	OriginalVitaminB12Mcg       *float64 `json:"original_vitamin_b12_mcg,omitempty"`
	OriginalBiotinMcg           *float64 `json:"original_biotin_mcg,omitempty"`
	OriginalPantothenicAcidMg   *float64 `json:"original_pantothenic_acid_mg,omitempty"`
	OriginalCholineMg           *float64 `json:"original_choline_mg,omitempty"`
	OriginalCalciumMg           *float64 `json:"original_calcium_mg,omitempty"`
	OriginalIronMg              *float64 `json:"original_iron_mg,omitempty"`
	OriginalMagnesiumMg         *float64 `json:"original_magnesium_mg,omitempty"`
	OriginalPhosphorusMg        *float64 `json:"original_phosphorus_mg,omitempty"`
	OriginalPotassiumMg         *float64 `json:"original_potassium_mg,omitempty"`
	OriginalZincMg              *float64 `json:"original_zinc_mg,omitempty"`
	OriginalCopperMg            *float64 `json:"original_copper_mg,omitempty"`
	OriginalManganeseMg         *float64 `json:"original_manganese_mg,omitempty"`
	OriginalSeleniumMcg         *float64 `json:"original_selenium_mcg,omitempty"`
	OriginalIodineMcg           *float64 `json:"original_iodine_mcg,omitempty"`
	OriginalMolybdenumMcg       *float64 `json:"original_molybdenum_mcg,omitempty"`
	OriginalChromiumMcg         *float64 `json:"original_chromium_mcg,omitempty"`
	OriginalFluorideMg          *float64 `json:"original_fluoride_mg,omitempty"`
	OriginalChlorideMg          *float64 `json:"original_chloride_mg,omitempty"`
	OriginalOmega3AlaG          *float64 `json:"original_omega3_ala_g,omitempty"`
	OriginalOmega3EpaG          *float64 `json:"original_omega3_epa_g,omitempty"`
	OriginalOmega3DhaG          *float64 `json:"original_omega3_dha_g,omitempty"`
	OriginalOmega6G             *float64 `json:"original_omega6_g,omitempty"`
	OriginalCreatineMg          *float64 `json:"original_creatine_mg,omitempty"`
	OriginalCaffeineMg          *float64 `json:"original_caffeine_mg,omitempty"`
	OriginalAlcoholG            *float64 `json:"original_alcohol_g,omitempty"`
	OriginalPolyunsaturatedFatG *float64 `json:"original_polyunsaturated_fat_g,omitempty"`
	OriginalMonounsaturatedFatG *float64 `json:"original_monounsaturated_fat_g,omitempty"`

	// Normalized nutrition data per 100g (for scaling)
	CaloriesPer100g            float64 `json:"calories_per_100g"`
	ProteinGPer100g            float64 `json:"protein_g_per_100g"`
	TotalFatGPer100g           float64 `json:"total_fat_g_per_100g"`
	SaturatedFatGPer100g       float64 `json:"saturated_fat_g_per_100g"`
	TransFatGPer100g           float64 `json:"trans_fat_g_per_100g"`
	CholesterolMgPer100g       float64 `json:"cholesterol_mg_per_100g"`
	SodiumMgPer100g            float64 `json:"sodium_mg_per_100g"`
	TotalCarbsGPer100g         float64 `json:"total_carbs_g_per_100g"`
	DietaryFiberGPer100g       float64 `json:"dietary_fiber_g_per_100g"`
	TotalSugarsGPer100g        float64 `json:"total_sugars_g_per_100g"`
	AddedSugarsGPer100g        float64 `json:"added_sugars_g_per_100g"`
	VitaminAMcgPer100g         float64 `json:"vitamin_a_mcg_per_100g"`
	VitaminCMgPer100g          float64 `json:"vitamin_c_mg_per_100g"`
	VitaminDMcgPer100g         float64 `json:"vitamin_d_mcg_per_100g"`
	VitaminEMgPer100g          float64 `json:"vitamin_e_mg_per_100g"`
	VitaminKMcgPer100g         float64 `json:"vitamin_k_mcg_per_100g"`
	ThiamineMgPer100g          float64 `json:"thiamine_mg_per_100g"`
	RiboflavinMgPer100g        float64 `json:"riboflavin_mg_per_100g"`
	NiacinMgPer100g            float64 `json:"niacin_mg_per_100g"`
	VitaminB6MgPer100g         float64 `json:"vitamin_b6_mg_per_100g"`
	FolateMcgPer100g           float64 `json:"folate_mcg_per_100g"`
	VitaminB12McgPer100g       float64 `json:"vitamin_b12_mcg_per_100g"`
	BiotinMcgPer100g           float64 `json:"biotin_mcg_per_100g"`
	PantothenicAcidMgPer100g   float64 `json:"pantothenic_acid_mg_per_100g"`
	CholineMgPer100g           float64 `json:"choline_mg_per_100g"`
	CalciumMgPer100g           float64 `json:"calcium_mg_per_100g"`
	IronMgPer100g              float64 `json:"iron_mg_per_100g"`
	MagnesiumMgPer100g         float64 `json:"magnesium_mg_per_100g"`
	PhosphorusMgPer100g        float64 `json:"phosphorus_mg_per_100g"`
	PotassiumMgPer100g         float64 `json:"potassium_mg_per_100g"`
	ZincMgPer100g              float64 `json:"zinc_mg_per_100g"`
	CopperMgPer100g            float64 `json:"copper_mg_per_100g"`
	ManganeseMgPer100g         float64 `json:"manganese_mg_per_100g"`
	SeleniumMcgPer100g         float64 `json:"selenium_mcg_per_100g"`
	IodineMcgPer100g           float64 `json:"iodine_mcg_per_100g"`
	MolybdenumMcgPer100g       float64 `json:"molybdenum_mcg_per_100g"`
	ChromiumMcgPer100g         float64 `json:"chromium_mcg_per_100g"`
	FluorideMgPer100g          float64 `json:"fluoride_mg_per_100g"`
	ChlorideMgPer100g          float64 `json:"chloride_mg_per_100g"`
	Omega3AlaGPer100g          float64 `json:"omega3_ala_g_per_100g"`
	Omega3EpaGPer100g          float64 `json:"omega3_epa_g_per_100g"`
	Omega3DhaGPer100g          float64 `json:"omega3_dha_g_per_100g"`
	Omega6GPer100g             float64 `json:"omega6_g_per_100g"`
	CreatineMgPer100g          float64 `json:"creatine_mg_per_100g"`
	CaffeineMgPer100g          float64 `json:"caffeine_mg_per_100g"`
	AlcoholGPer100g            float64 `json:"alcohol_g_per_100g"`
	PolyunsaturatedFatGPer100g float64 `json:"polyunsaturated_fat_g_per_100g"`
	MonounsaturatedFatGPer100g float64 `json:"monounsaturated_fat_g_per_100g"`
	Note                       *string `json:"note,omitempty"`
	// Timestamps for 30-day refresh logic
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConsumptionItem represents the relationship between a consumption and an item
// with nutrition snapshots preserved at time of logging
type ConsumptionItem struct {
	ID            string   `json:"id"`
	ConsumptionID string   `json:"consumption_id"`
	ItemID        *string  `json:"item_id"`        // Optional reference to global items cache
	Name          string   `json:"name"`           // Display name snapshot
	Brand         string   `json:"brand"`          // Brand snapshot
	Grams         float64  `json:"grams"`          // Actual grams consumed (normalized internally)
	UserQuantity  *float64 `json:"user_quantity"`  // Original user input quantity for display
	UserUnit      *string  `json:"user_unit"`      // Original user input unit for display
	Note          *string  `json:"note,omitempty"` // Additional note about the item

	// Nutrition snapshot for THIS SERVING (not per-100g)
	Calories            float64 `json:"calories"`
	ProteinG            float64 `json:"protein_g"`
	TotalFatG           float64 `json:"total_fat_g"`
	SaturatedFatG       float64 `json:"saturated_fat_g"`
	TransFatG           float64 `json:"trans_fat_g"`
	CholesterolMg       float64 `json:"cholesterol_mg"`
	SodiumMg            float64 `json:"sodium_mg"`
	TotalCarbsG         float64 `json:"total_carbs_g"`
	DietaryFiberG       float64 `json:"dietary_fiber_g"`
	TotalSugarsG        float64 `json:"total_sugars_g"`
	AddedSugarsG        float64 `json:"added_sugars_g"`
	VitaminAMcg         float64 `json:"vitamin_a_mcg"`
	VitaminCMg          float64 `json:"vitamin_c_mg"`
	VitaminDMcg         float64 `json:"vitamin_d_mcg"`
	VitaminEMg          float64 `json:"vitamin_e_mg"`
	VitaminKMcg         float64 `json:"vitamin_k_mcg"`
	ThiamineMg          float64 `json:"thiamine_mg"`
	RiboflavinMg        float64 `json:"riboflavin_mg"`
	NiacinMg            float64 `json:"niacin_mg"`
	VitaminB6Mg         float64 `json:"vitamin_b6_mg"`
	FolateMcg           float64 `json:"folate_mcg"`
	VitaminB12Mcg       float64 `json:"vitamin_b12_mcg"`
	BiotinMcg           float64 `json:"biotin_mcg"`
	PantothenicAcidMg   float64 `json:"pantothenic_acid_mg"`
	CholineMg           float64 `json:"choline_mg"`
	CalciumMg           float64 `json:"calcium_mg"`
	IronMg              float64 `json:"iron_mg"`
	MagnesiumMg         float64 `json:"magnesium_mg"`
	PhosphorusMg        float64 `json:"phosphorus_mg"`
	PotassiumMg         float64 `json:"potassium_mg"`
	ZincMg              float64 `json:"zinc_mg"`
	CopperMg            float64 `json:"copper_mg"`
	ManganeseMg         float64 `json:"manganese_mg"`
	SeleniumMcg         float64 `json:"selenium_mcg"`
	IodineMcg           float64 `json:"iodine_mcg"`
	MolybdenumMcg       float64 `json:"molybdenum_mcg"`
	ChromiumMcg         float64 `json:"chromium_mcg"`
	FluorideMg          float64 `json:"fluoride_mg"`
	ChlorideMg          float64 `json:"chloride_mg"`
	Omega3AlaG          float64 `json:"omega3_ala_g"`
	Omega3EpaG          float64 `json:"omega3_epa_g"`
	Omega3DhaG          float64 `json:"omega3_dha_g"`
	Omega6G             float64 `json:"omega6_g"`
	CreatineMg          float64 `json:"creatine_mg"`
	CaffeineMg          float64 `json:"caffeine_mg"`
	AlcoholG            float64 `json:"alcohol_g"`
	PolyunsaturatedFatG float64 `json:"polyunsaturated_fat_g"`
	MonounsaturatedFatG float64 `json:"monounsaturated_fat_g"`

	Labels    []*Label   `json:"labels,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
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
	TotalCalcium            float64 `json:"total_calcium_mg"`
	TotalIron               float64 `json:"total_iron_mg"`
	TotalMagnesium          float64 `json:"total_magnesium_mg"`
	TotalPhosphorus         float64 `json:"total_phosphorus_mg"`
	TotalPotassium          float64 `json:"total_potassium_mg"`
	TotalZinc               float64 `json:"total_zinc_mg"`
	TotalCopper             float64 `json:"total_copper_mg"`
	TotalManganese          float64 `json:"total_manganese_mg"`
	TotalSelenium           float64 `json:"total_selenium_mcg"`
	TotalIodine             float64 `json:"total_iodine_mcg"`
	TotalMolybdenum         float64 `json:"total_molybdenum_mcg"`
	TotalChromium           float64 `json:"total_chromium_mcg"`
	TotalFluoride           float64 `json:"total_fluoride_mg"`
	TotalChloride           float64 `json:"total_chloride_mg"`
	TotalOmega3Ala          float64 `json:"total_omega3_ala_g"`
	TotalOmega3Epa          float64 `json:"total_omega3_epa_g"`
	TotalOmega3Dha          float64 `json:"total_omega3_dha_g"`
	TotalOmega6             float64 `json:"total_omega6_g"`
	TotalCreatine           float64 `json:"total_creatine_mg"`
	TotalCaffeine           float64 `json:"total_caffeine_mg"`
	TotalAlcohol            float64 `json:"total_alcohol_g"`
	TotalPolyunsaturatedFat float64 `json:"total_polyunsaturated_fat_g"`
	TotalMonounsaturatedFat float64 `json:"total_monounsaturated_fat_g"`

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

// Label represents a user-owned label with GitHub-style properties
type Label struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LabelWithUsage extends Label with usage count information
type LabelWithUsage struct {
	Label
	ConsumptionCount int `json:"consumption_count"`
	ItemCount        int `json:"item_count"`
}

// Event represents a user-owned event for tracking symptoms, activities, etc.
type Event struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      string     `json:"name"`
	Category  *string    `json:"category,omitempty"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Level     *int       `json:"level,omitempty"` // 0-10 scale
	Note      *string    `json:"note,omitempty"`  // Up to 1000 chars
	Color     *string    `json:"color,omitempty"` // Hex color code
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// EventLink represents a manual link between an event and a consumption/item
type EventLink struct {
	ID                string    `json:"id"`
	EventID           string    `json:"event_id"`
	ConsumptionID     *string   `json:"consumption_id,omitempty"`
	ConsumptionItemID *string   `json:"consumption_item_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// EventListOptions contains filtering options for event queries
type EventListOptions struct {
	Limit     int
	Offset    int
	StartDate *time.Time
	EndDate   *time.Time
	Category  *string
	LevelMin  *int
	LevelMax  *int
	Labels    []string
	MatchAll  bool // If true, match ALL labels; if false, match ANY label
}
