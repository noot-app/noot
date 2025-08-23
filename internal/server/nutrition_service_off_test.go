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
