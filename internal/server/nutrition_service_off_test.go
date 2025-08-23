package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// flexFloatPtr creates a pointer to a FlexFloat for tests
func flexFloatPtr(f float64) *FlexFloat {
	ff := FlexFloat(f)
	return &ff
}

func TestNutritionService_OFF_Integration(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Mock OFF API response
	mockResponse := OFFSearchResponse{
		Products: []OFFProduct{
			{
				ProductName: "Whole Milk",
				Brands:      "Test Brand",
				Nutriments: OFFNutriments{
					EnergyKcal100g:    flexFloatPtr(60),
					Proteins100g:      flexFloatPtr(3.4),
					Fat100g:           flexFloatPtr(3.3),
					Carbohydrates100g: flexFloatPtr(4.7),
					Sodium100g:        flexFloatPtr(44),
				},
			},
		},
		Count: 1,
	} // Create mock OFF server
	offServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer offServer.Close()

	// Create test storage
	store, err := storage.NewSQLiteStore(":memory:")
	require.NoError(t, err)

	// Create mock AI provider that will receive OFF context and return enhanced nutrition
	mockAI := &mockAIProvider{
		nutritionResponse: CompleteNutrient{
			// AI should return enhanced nutrition data for 250g serving, potentially using OFF context
			Calories:   150,   // Expected final result
			Protein:    8.5,   // Expected final result
			TotalFat:   8.25,  // Expected final result
			TotalCarbs: 11.75, // Expected final result
			Sodium:     110,   // Expected final result
			VitaminB12: 2.4,   // AI can provide nutrients not available in OFF
		},
		contextReceived: nil,
	}

	// Create nutrition service with mock AI provider and OFF client
	service := &NutritionService{
		aiProvider: mockAI,
		offClient: NewOFFClient(OFFClientConfig{
			BaseURL:   offServer.URL,
			UserAgent: "test-agent",
			Timeout:   5 * time.Second,
			Enabled:   true,
		}),
		converter: NewUnitConverter(),
		store:     store,
	}

	// Test item
	item := Item{
		Name:  "whole milk",
		Grams: 250,
		Brand: stringPtr("Test Brand"),
	}

	// Hydrate nutrition
	ctx := context.Background()
	hydratedItem, err := service.hydrateItemNutrition(ctx, item)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, hydratedItem.Nutrients)

	// Verify AI was called with OFF context
	assert.NotNil(t, mockAI.contextReceived, "AI should have been called with context")

	// Verify the AI returned enhanced nutrition data (combining OFF insight with complete nutrition)
	assert.Equal(t, 150.0, hydratedItem.Nutrients.Calories)
	assert.Equal(t, 8.5, hydratedItem.Nutrients.Protein)
	assert.Equal(t, 8.25, hydratedItem.Nutrients.TotalFat)
	assert.Equal(t, 11.75, hydratedItem.Nutrients.TotalCarbs)
	assert.Equal(t, 110.0, hydratedItem.Nutrients.Sodium)
	// AI can provide additional nutrients not available in OFF
	assert.Equal(t, 2.4, hydratedItem.Nutrients.VitaminB12)
}

func TestNutritionService_OFF_Disabled(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Create test storage
	store, err := storage.NewSQLiteStore(":memory:")
	require.NoError(t, err)

	// Mock AI provider that returns test data
	mockAI := &mockAIProvider{
		nutritionResponse: CompleteNutrient{
			Calories: 100,
			Protein:  5,
		},
	}

	// Create nutrition service with OFF disabled
	service := &NutritionService{
		aiProvider: mockAI,
		offClient:  nil, // Disabled
		converter:  NewUnitConverter(),
		store:      store,
	}

	// Test item
	item := Item{
		Name:  "test item",
		Grams: 100,
	}

	// Hydrate nutrition
	ctx := context.Background()
	hydratedItem, err := service.hydrateItemNutrition(ctx, item)

	// Verify AI was used (not OFF)
	require.NoError(t, err)
	require.NotNil(t, hydratedItem.Nutrients)
	assert.Equal(t, 100.0, hydratedItem.Nutrients.Calories)
	assert.Equal(t, 5.0, hydratedItem.Nutrients.Protein)
}

// Mock AI provider for testing
type mockAIProvider struct {
	nutritionResponse CompleteNutrient
	shouldError       bool
	contextReceived   interface{}
}

func (m *mockAIProvider) TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error) {
	return "mock transcription", nil
}

func (m *mockAIProvider) ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	return ParsedItems{}, nil
}

func (m *mockAIProvider) GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error) {
	return m.GetNutritionWithContext(ctx, item, nil)
}

func (m *mockAIProvider) GetNutritionWithContext(ctx context.Context, item Item, nutritionContext interface{}) (CompleteNutrient, error) {
	// Store the context for verification in tests
	m.contextReceived = nutritionContext

	if m.shouldError {
		return CompleteNutrient{}, assert.AnError
	}
	return m.nutritionResponse, nil
}

func TestNutritionService_OFF_ServingSizeContext(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Mock OFF API response with serving size information
	servingQuantity := FlexFloat(355)
	mockResponse := OFFSearchResponse{
		Products: []OFFProduct{
			{
				ProductName:         "Cream Soda",
				Brands:              "Olipop",
				ServingQuantity:     &servingQuantity,
				ServingQuantityUnit: "ml",
				ServingSize:         "1 can (355 ml)",
				Nutriments: OFFNutriments{
					EnergyKcal100g:    flexFloatPtr(11.3),
					Carbohydrates100g: flexFloatPtr(4.79),
					Sugars100g:        flexFloatPtr(0.563),
					Fiber100g:         flexFloatPtr(2.54),
				},
			},
		},
		Count: 1,
	}

	// Create mock OFF server
	offServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer offServer.Close()

	// Create test storage
	store, err := storage.NewSQLiteStore(":memory:")
	require.NoError(t, err)
	require.NoError(t, store.Migrate())

	// Create OFF client
	offClient := NewOFFClient(OFFClientConfig{
		BaseURL:   offServer.URL,
		UserAgent: "test-agent",
		Timeout:   5 * time.Second,
		Enabled:   true,
	})

	// Create mock AI provider
	mockAI := &mockAIProvider{
		nutritionResponse: CompleteNutrient{
			Calories: 40,
			Protein:  0,
		},
	}

	// Create nutrition service
	service := &NutritionService{
		aiProvider: mockAI,
		offClient:  offClient,
		converter:  NewUnitConverter(),
		store:      store,
	}

	// Test item with brand (to trigger OFF lookup)
	brand := "Olipop"
	items := []Item{
		{
			Name:  "Cream soda Olipop",
			Grams: 355,
			Brand: &brand,
		},
	}

	// Hydrate nutrition
	ctx := context.Background()
	result, err := service.HydrateNutritionWithoutCache(ctx, items)

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.NotNil(t, result[0].Nutrients)

	// Verify that the context received by the AI provider includes serving size information
	require.NotNil(t, mockAI.contextReceived)
	contextMap, ok := mockAI.contextReceived.(map[string]interface{})
	require.True(t, ok, "Context should be a map")

	source, ok := contextMap["source"].(string)
	require.True(t, ok)
	assert.Equal(t, "open_food_facts", source)

	products, ok := contextMap["products"].([]interface{})
	require.True(t, ok)
	require.Len(t, products, 1)

	product, ok := products[0].(map[string]interface{})
	require.True(t, ok)

	// Verify basic product information
	assert.Equal(t, "Cream Soda", product["product_name"])
	assert.Equal(t, "Olipop", product["brands"])

	// Verify serving size information is included
	servingQuantityStr, ok := product["serving_quantity"].(string)
	require.True(t, ok, "serving_quantity should be present as string")
	assert.Equal(t, "355", servingQuantityStr)

	servingUnit, ok := product["serving_quantity_unit"].(string)
	require.True(t, ok, "serving_quantity_unit should be present")
	assert.Equal(t, "ml", servingUnit)

	servingSize, ok := product["serving_size"].(string)
	require.True(t, ok, "serving_size should be present")
	assert.Equal(t, "1 can (355 ml)", servingSize)
}
