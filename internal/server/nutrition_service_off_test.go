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

func TestNutritionService_OFF_Integration(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Mock OFF API response
	mockResponse := OFFSearchResponse{
		Products: []OFFProduct{
			{
				ProductName: "Whole Milk",
				Brands:      "Test Brand",
				CompletedT:  0.9,
				Nutriments: OFFNutriments{
					EnergyKcal100g:    offFloatPtr(60),
					Proteins100g:      offFloatPtr(3.4),
					Fat100g:           offFloatPtr(3.3),
					Carbohydrates100g: offFloatPtr(4.7),
					Sodium100g:        offFloatPtr(44),
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

	// Create nutrition service with mock OFF client
	service := &NutritionService{
		aiProvider: &mockAIProvider{}, // Fallback should not be used
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

	// Verify nutrition values are scaled correctly (250g serving)
	// OFF data is per 100g, so 250g should be 2.5x the values
	expected := CompleteNutrient{
		Calories:     150,  // 60 * 2.5
		Protein:      8.5,  // 3.4 * 2.5
		TotalFat:     8.25, // 3.3 * 2.5
		TotalCarbs:   11.75, // 4.7 * 2.5
		Sodium:       110,  // 44 * 2.5
		// Other fields should be 0 since not in OFF
		TransFat:     0,
		Cholesterol:  0,
		VitaminB12:   0,
	}

	assert.Equal(t, expected.Calories, hydratedItem.Nutrients.Calories)
	assert.Equal(t, expected.Protein, hydratedItem.Nutrients.Protein)
	assert.Equal(t, expected.TotalFat, hydratedItem.Nutrients.TotalFat)
	assert.Equal(t, expected.TotalCarbs, hydratedItem.Nutrients.TotalCarbs)
	assert.Equal(t, expected.Sodium, hydratedItem.Nutrients.Sodium)
	assert.Equal(t, expected.TransFat, hydratedItem.Nutrients.TransFat)
	assert.Equal(t, expected.VitaminB12, hydratedItem.Nutrients.VitaminB12)
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
}

func (m *mockAIProvider) TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error) {
	return "mock transcription", nil
}

func (m *mockAIProvider) ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	return ParsedItems{}, nil
}

func (m *mockAIProvider) GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error) {
	if m.shouldError {
		return CompleteNutrient{}, assert.AnError
	}
	return m.nutritionResponse, nil
}