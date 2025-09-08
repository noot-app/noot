package server

import (
	"context"
	"testing"
	"time"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStore implements a simple in-memory store for testing deduplication
type MockStore struct {
	items map[string]*storage.Item
}

func NewMockStore() *MockStore {
	return &MockStore{
		items: make(map[string]*storage.Item),
	}
}

func (m *MockStore) CreateItem(ctx context.Context, item *storage.Item) error {
	key := item.NormalizedName + "|" + item.NormalizedBrand
	m.items[key] = item
	return nil
}

func (m *MockStore) GetItemByName(ctx context.Context, normalizedName, normalizedBrand string) (*storage.Item, error) {
	key := normalizedName + "|" + normalizedBrand
	if item, exists := m.items[key]; exists {
		return item, nil
	}
	return nil, nil
}

func (m *MockStore) GetItem(ctx context.Context, id string) (*storage.Item, error) {
	// Simple implementation for testing - in practice this would lookup by ID
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (m *MockStore) UpdateItem(ctx context.Context, item *storage.Item) error {
	key := item.NormalizedName + "|" + item.NormalizedBrand
	m.items[key] = item
	return nil
}

// Implement other required Store methods as no-ops for testing
func (m *MockStore) CreateUser(ctx context.Context, user *storage.User) error      { return nil }
func (m *MockStore) GetUser(ctx context.Context, id string) (*storage.User, error) { return nil, nil }
func (m *MockStore) GetUserByEmail(ctx context.Context, email string) (*storage.User, error) {
	return nil, nil
}
func (m *MockStore) UpdateUser(ctx context.Context, user *storage.User) error { return nil }
func (m *MockStore) CreateConsumption(ctx context.Context, consumption *storage.Consumption) error {
	return nil
}
func (m *MockStore) GetConsumption(ctx context.Context, id string) (*storage.Consumption, error) {
	return nil, nil
}
func (m *MockStore) GetConsumptionForUser(ctx context.Context, userID, id string) (*storage.Consumption, error) {
	return nil, nil
}
func (m *MockStore) GetPublicConsumption(ctx context.Context, id string) (*storage.Consumption, error) {
	return nil, nil
}
func (m *MockStore) UpdateConsumption(ctx context.Context, consumption *storage.Consumption) error {
	return nil
}
func (m *MockStore) DeleteConsumption(ctx context.Context, id string) error { return nil }
func (m *MockStore) GetConsumptionsByUser(ctx context.Context, userID string, limit, offset int) ([]*storage.Consumption, error) {
	return nil, nil
}
func (m *MockStore) GetConsumptionsByUserSince(ctx context.Context, userID string, since time.Time) ([]*storage.Consumption, error) {
	return nil, nil
}
func (m *MockStore) GetConsumptionsByUserDateRange(ctx context.Context, userID string, start, end time.Time, limit, offset int) ([]*storage.Consumption, error) {
	return nil, nil
}
func (m *MockStore) GetNutritionSummary(ctx context.Context, userID string, start, end time.Time) (*storage.NutritionSummary, error) {
	return nil, nil
}
func (m *MockStore) CreateConsumptionItem(ctx context.Context, item *storage.ConsumptionItem) error {
	return nil
}
func (m *MockStore) GetConsumptionItems(ctx context.Context, consumptionID string) ([]*storage.ConsumptionItem, error) {
	return nil, nil
}
func (m *MockStore) UpdateConsumptionItem(ctx context.Context, item *storage.ConsumptionItem) error {
	return nil
}
func (m *MockStore) DeleteConsumptionItem(ctx context.Context, id string) error { return nil }
func (m *MockStore) DeleteConsumptionItemsByConsumption(ctx context.Context, consumptionID string) error {
	return nil
}
func (m *MockStore) GetStaleItems(ctx context.Context, staleAfter time.Time) ([]*storage.Item, error) {
	return nil, nil
}
func (m *MockStore) UpsertUserGoal(ctx context.Context, goal *storage.UserGoal) error { return nil }
func (m *MockStore) GetUserGoal(ctx context.Context, userID, name string) (*storage.UserGoal, error) {
	return nil, nil
}
func (m *MockStore) GetUserGoals(ctx context.Context, userID string) ([]*storage.UserGoal, error) {
	return nil, nil
}
func (m *MockStore) DeleteUserGoal(ctx context.Context, userID, name string) error    { return nil }
func (m *MockStore) SetActiveGoal(ctx context.Context, userID, goalName string) error { return nil }
func (m *MockStore) ClearActiveGoal(ctx context.Context, userID string) error         { return nil }
func (m *MockStore) GetActiveGoalName(ctx context.Context, userID string) (*string, error) {
	return nil, nil
}
func (m *MockStore) UpsertUserBiometrics(ctx context.Context, biometrics *storage.UserBiometrics) error {
	return nil
}
func (m *MockStore) GetUserBiometrics(ctx context.Context, userID string) (*storage.UserBiometrics, error) {
	return nil, nil
}
func (m *MockStore) DeleteUserBiometrics(ctx context.Context, userID string) error { return nil }
func (m *MockStore) CreateLabel(ctx context.Context, label *storage.Label) error   { return nil }
func (m *MockStore) UpdateLabel(ctx context.Context, label *storage.Label) error   { return nil }
func (m *MockStore) DeleteLabel(ctx context.Context, userID, id string) error      { return nil }
func (m *MockStore) GetLabel(ctx context.Context, userID, id string) (*storage.Label, error) {
	return nil, nil
}
func (m *MockStore) ListLabels(ctx context.Context, userID string) ([]*storage.LabelWithUsage, error) {
	return nil, nil
}
func (m *MockStore) ListConsumptionLabels(ctx context.Context, userID, consumptionID string) ([]*storage.Label, error) {
	return nil, nil
}
func (m *MockStore) AssignConsumptionLabels(ctx context.Context, userID, consumptionID string, labelIDs []string) error {
	return nil
}
func (m *MockStore) UnassignConsumptionLabel(ctx context.Context, userID, consumptionID, labelID string) error {
	return nil
}
func (m *MockStore) ListConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string) ([]*storage.Label, error) {
	return nil, nil
}
func (m *MockStore) AssignConsumptionItemLabels(ctx context.Context, userID, consumptionItemID string, labelIDs []string) error {
	return nil
}
func (m *MockStore) UnassignConsumptionItemLabel(ctx context.Context, userID, consumptionItemID, labelID string) error {
	return nil
}
func (m *MockStore) GetConsumptionsByLabels(ctx context.Context, userID string, labelNames []string, matchAll bool, limit, offset int) ([]*storage.Consumption, error) {
	return nil, nil
}
func (m *MockStore) CreateEvent(ctx context.Context, event *storage.Event) error { return nil }
func (m *MockStore) UpdateEvent(ctx context.Context, event *storage.Event) error { return nil }
func (m *MockStore) DeleteEvent(ctx context.Context, userID, id string) error    { return nil }
func (m *MockStore) GetEvent(ctx context.Context, userID, id string) (*storage.Event, error) {
	return nil, nil
}
func (m *MockStore) ListEvents(ctx context.Context, userID string, options storage.EventListOptions) ([]*storage.Event, error) {
	return nil, nil
}
func (m *MockStore) CreateEventType(ctx context.Context, eventType *storage.EventType) error {
	return nil
}
func (m *MockStore) UpdateEventType(ctx context.Context, eventType *storage.EventType) error {
	return nil
}
func (m *MockStore) DeleteEventType(ctx context.Context, userID, id string) error { return nil }
func (m *MockStore) GetEventType(ctx context.Context, userID, id string) (*storage.EventType, error) {
	return nil, nil
}
func (m *MockStore) ListEventTypes(ctx context.Context, userID string) ([]*storage.EventType, error) {
	return nil, nil
}
func (m *MockStore) GetEventTypeCounts(ctx context.Context, userID string) (map[string]int, error) {
	return nil, nil
}
func (m *MockStore) ListEventLabels(ctx context.Context, userID, eventID string) ([]*storage.Label, error) {
	return nil, nil
}
func (m *MockStore) AssignEventLabels(ctx context.Context, userID, eventID string, labelIDs []string) error {
	return nil
}
func (m *MockStore) UnassignEventLabel(ctx context.Context, userID, eventID, labelID string) error {
	return nil
}
func (m *MockStore) CreateEventLink(ctx context.Context, link *storage.EventLink) error { return nil }
func (m *MockStore) DeleteEventLink(ctx context.Context, userID, linkID string) error   { return nil }
func (m *MockStore) ListEventLinks(ctx context.Context, userID, eventID string) ([]*storage.EventLink, error) {
	return nil, nil
}
func (m *MockStore) CreateAPIKey(ctx context.Context, apiKey *storage.APIKey) error { return nil }
func (m *MockStore) GetAPIKey(ctx context.Context, userID, id string) (*storage.APIKey, error) {
	return nil, nil
}
func (m *MockStore) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*storage.APIKey, error) {
	return nil, nil
}
func (m *MockStore) ListAPIKeys(ctx context.Context, userID string) ([]*storage.APIKey, error) {
	return nil, nil
}
func (m *MockStore) UpdateAPIKey(ctx context.Context, apiKey *storage.APIKey) error { return nil }
func (m *MockStore) RevokeAPIKey(ctx context.Context, userID, id string) error      { return nil }
func (m *MockStore) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	return nil
}
func (m *MockStore) Close() error { return nil }

func TestFoodDeduplication(t *testing.T) {
	// Initialize logger for the test
	InitLogger()

	ctx := context.Background()
	store := NewMockStore()

	service := &NutritionService{
		store:     store,
		converter: NewUnitConverter(),
	}

	// Test data - different ways to refer to the same food
	// Carrots have about 41 calories per 100g in reality
	baseCaloriesPer100g := 41.0

	testCases := []struct {
		name        string
		itemName    string
		grams       float64
		brand       *string
		expectedKey string
		description string
	}{
		{
			name:        "BaseCarrots",
			itemName:    "carrots",
			grams:       100.0,
			brand:       nil,
			expectedKey: "carrot|",
			description: "Base carrot entry",
		},
		{
			name:        "HandfulOfCarrots",
			itemName:    "a handful of carrots",
			grams:       60.0,
			brand:       nil,
			expectedKey: "carrot|",
			description: "Should deduplicate to same carrot entry",
		},
		{
			name:        "SlicesOfCarrots",
			itemName:    "a few slices of carrots",
			grams:       80.0,
			brand:       nil,
			expectedKey: "carrot|",
			description: "Should deduplicate to same carrot entry",
		},
		{
			name:        "FreshCarrots",
			itemName:    "fresh organic carrots",
			grams:       150.0,
			brand:       nil,
			expectedKey: "carrot|",
			description: "Should deduplicate to same carrot entry",
		},
		{
			name:        "TwoCarrots",
			itemName:    "2 carrots",
			grams:       200.0,
			brand:       nil,
			expectedKey: "carrot|",
			description: "Should deduplicate to same carrot entry",
		},
	}

	// Cache all the different carrot variations
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate nutrition based on actual grams (proportional to base per-100g values)
			actualCalories := baseCaloriesPer100g * (tc.grams / 100.0)
			carrotNutrition := CompleteNutrient{
				Calories:   actualCalories,
				Protein:    0.9 * (tc.grams / 100.0), // 0.9g protein per 100g carrots
				TotalCarbs: 9.6 * (tc.grams / 100.0), // 9.6g carbs per 100g carrots
				TotalFat:   0.2 * (tc.grams / 100.0), // 0.2g fat per 100g carrots
			}

			item := Item{
				Name:          tc.itemName,
				CanonicalName: "carrot", // LLM would provide this canonical name
				Grams:         tc.grams,
				Brand:         tc.brand,
			}

			err := service.cacheNutritionData(ctx, item, carrotNutrition)
			require.NoError(t, err, "Failed to cache %s", tc.description)

			// Verify the canonical key was generated correctly
			actualKey := service.makeCanonicalFoodKey(item)
			assert.Equal(t, tc.expectedKey, actualKey, "Canonical key mismatch for %s", tc.description)
		})
	}

	// Verify only one item was stored (deduplication worked)
	assert.Equal(t, 1, len(store.items), "Should have only one deduplicated carrot item")

	// Verify we can retrieve and scale nutrition for all variations
	for _, tc := range testCases {
		t.Run("Retrieve_"+tc.name, func(t *testing.T) {
			item := Item{
				Name:          tc.itemName,
				CanonicalName: "carrot", // LLM would provide this canonical name
				Grams:         tc.grams,
				Brand:         tc.brand,
			}

			nutrition, err := service.fetchNutritionFromCache(ctx, item)
			require.NoError(t, err, "Failed to fetch nutrition for %s", tc.description)
			require.NotNil(t, nutrition, "Should find cached nutrition for %s", tc.description)

			// Verify nutrition is properly scaled
			// All items should return the same per-100g equivalent since they're the same food
			expectedCalories := baseCaloriesPer100g * (tc.grams / 100.0)
			assert.InDelta(t, expectedCalories, nutrition.Calories, 0.5,
				"Calories should be properly scaled for %s (expected %.1f, got %.1f)",
				tc.description, expectedCalories, nutrition.Calories)

			// Verify it scales correctly - the per-100g rate should be consistent
			actualPer100g := nutrition.Calories * (100.0 / tc.grams)
			assert.InDelta(t, baseCaloriesPer100g, actualPer100g, 1.0,
				"Per-100g calories should be consistent for %s (expected %.1f, got %.1f)",
				tc.description, baseCaloriesPer100g, actualPer100g)
		})
	}
}

func TestCanonicalKeyGeneration(t *testing.T) {
	service := &NutritionService{
		converter: NewUnitConverter(),
	}

	testCases := []struct {
		name        string
		item        Item
		expectedKey string
	}{
		{
			name: "SimpleFood",
			item: Item{
				Name:          "apple",
				CanonicalName: "apple", // LLM canonical name
				Brand:         nil,
			},
			expectedKey: "apple|",
		},
		{
			name: "PluralFood",
			item: Item{
				Name:          "apples",
				CanonicalName: "apple", // LLM normalizes plural to singular
				Brand:         nil,
			},
			expectedKey: "apple|",
		},
		{
			name: "QuantityFood",
			item: Item{
				Name:          "2 apples",
				CanonicalName: "apple", // LLM strips quantity and normalizes
				Brand:         nil,
			},
			expectedKey: "apple|",
		},
		{
			name: "ComplexQuantity",
			item: Item{
				Name:          "a few slices of fresh organic apples",
				CanonicalName: "apple", // LLM strips all descriptors
				Brand:         nil,
			},
			expectedKey: "apple|",
		},
		{
			name: "BrandedItem",
			item: Item{
				Name:          "Ben Jerry vanilla ice cream",
				CanonicalName: "vanilla ice cream", // LLM provides canonical name without brand
				Brand:         stringPtr("Ben Jerry"),
			},
			expectedKey: "vanilla ice cream|ben jerry",
		},
		{
			name: "EmptyCanonicalFallback",
			item: Item{
				Name:          "Coca Cola",
				CanonicalName: "", // Empty canonical name, falls back to normalized name
				Brand:         stringPtr("Coca Cola"),
			},
			expectedKey: "coca cola|coca cola",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := service.makeCanonicalFoodKey(tc.item)
			assert.Equal(t, tc.expectedKey, key, "Canonical key mismatch for %s", tc.item.Name)
		})
	}
}
