package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grantbirki/noot/internal/storage"
)

// MockAIProviderWithTranscript is a mock that tracks if transcript context is used
type MockAIProviderWithTranscript struct {
	transcriptUsed         bool
	capturedTranscript     string
	parseItemsCallCount    int
	transcriptContextCalls int
	expectedItems          []Item
	expectedNutrients      CompleteNutrient
	expectedIngredients    []storage.OFFIngredient
	expectedURL            *string
}

func (m *MockAIProviderWithTranscript) TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error) {
	return "mocked transcript", nil
}

func (m *MockAIProviderWithTranscript) ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	m.parseItemsCallCount++
	// Return mock items - transcript will be added by NutritionService.ParseItems
	return ParsedItems{
		Items: m.expectedItems,
		// Note: Transcript will be set by caller
	}, nil
}

func (m *MockAIProviderWithTranscript) GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error) {
	return m.expectedNutrients, nil
}

func (m *MockAIProviderWithTranscript) GetNutritionWithContext(ctx context.Context, item Item) (CompleteNutrient, error) {
	return m.expectedNutrients, nil
}

func (m *MockAIProviderWithTranscript) GetNutritionWithContextComplete(ctx context.Context, item Item) (NutritionResponse, error) {
	return NutritionResponse{
		Nutrients:   m.expectedNutrients,
		Ingredients: m.expectedIngredients,
		URL:         m.expectedURL,
	}, nil
}

func (m *MockAIProviderWithTranscript) GetNutritionWithTranscriptContext(ctx context.Context, item Item, transcript string) (CompleteNutrient, error) {
	m.transcriptUsed = true
	m.capturedTranscript = transcript
	m.transcriptContextCalls++
	return m.expectedNutrients, nil
}

func (m *MockAIProviderWithTranscript) GetNutritionWithTranscriptContextComplete(ctx context.Context, item Item, transcript string) (NutritionResponse, error) {
	m.transcriptUsed = true
	m.capturedTranscript = transcript
	m.transcriptContextCalls++
	return NutritionResponse{
		Nutrients:   m.expectedNutrients,
		Ingredients: m.expectedIngredients,
		URL:         m.expectedURL,
	}, nil
}

func TestTranscriptWiredThroughParseToNutrition(t *testing.T) {
	// Initialize logger for the test
	InitLogger()

	ctx := context.Background()
	transcript := "I had lunch: a grilled chicken salad with vegetables and ranch dressing"

	// Create mock nutrition data
	expectedNutrients := CompleteNutrient{
		Calories:   450.0,
		Protein:    35.0,
		TotalCarbs: 15.0,
		TotalFat:   25.0,
	}

	expectedItems := []Item{
		{
			Name:          "grilled chicken salad",
			Grams:         200.0,
			UserQuantity:  nil,
			UserUnit:      nil,
			Brand:         nil,
			Context:       stringPtr("for lunch"),
			CanonicalName: "chicken salad",
		},
	}

	// Create mock AI provider
	mockProvider := &MockAIProviderWithTranscript{
		expectedItems:     expectedItems,
		expectedNutrients: expectedNutrients,
		expectedIngredients: []storage.OFFIngredient{
			{ID: "en:chicken", Text: "chicken", PercentEstimate: func() *float64 { v := 60.0; return &v }()},
			{ID: "en:lettuce", Text: "lettuce", PercentEstimate: func() *float64 { v := 30.0; return &v }()},
		},
		expectedURL: stringPtr("https://example.com/nutrition"),
	}

	// Create nutrition service with mock store
	store := NewMockStore()
	service := &NutritionService{
		aiProvider: mockProvider,
		converter:  NewUnitConverter(),
		store:      store,
	}

	t.Run("ParseItems stores transcript", func(t *testing.T) {
		// Test that ParseItems captures and stores the transcript
		parsed, err := service.ParseItems(ctx, transcript)
		require.NoError(t, err)

		// Verify transcript is stored in ParsedItems
		assert.Equal(t, transcript, parsed.Transcript)
		assert.Len(t, parsed.Items, 1)
		assert.Equal(t, "grilled chicken salad", parsed.Items[0].Name)
		assert.Equal(t, 1, mockProvider.parseItemsCallCount)
	})

	t.Run("HydrateNutrition uses transcript context", func(t *testing.T) {
		// Parse items first to get transcript-enabled ParsedItems
		parsed, err := service.ParseItems(ctx, transcript)
		require.NoError(t, err)

		// Reset call counters
		mockProvider.transcriptUsed = false
		mockProvider.capturedTranscript = ""
		mockProvider.transcriptContextCalls = 0

		// Test hydration with transcript context
		hydratedItems, err := service.HydrateNutrition(ctx, parsed.Items, parsed.Transcript)
		require.NoError(t, err)

		// Verify transcript was used
		assert.True(t, mockProvider.transcriptUsed, "Transcript should be used in nutrition calls")
		assert.Equal(t, transcript, mockProvider.capturedTranscript, "Captured transcript should match original")
		assert.Greater(t, mockProvider.transcriptContextCalls, 0, "Transcript context calls should be made")

		// Verify items were hydrated correctly
		assert.Len(t, hydratedItems, 1)
		assert.NotNil(t, hydratedItems[0].Nutrients)
		assert.Equal(t, expectedNutrients.Calories, hydratedItems[0].Nutrients.Calories)
		assert.Len(t, hydratedItems[0].Ingredients, 2)
		assert.NotNil(t, hydratedItems[0].Url)
		assert.Equal(t, "https://example.com/nutrition", *hydratedItems[0].Url)
	})

	t.Run("HydrateNutritionWithoutCache uses transcript context", func(t *testing.T) {
		// Parse items first to get transcript-enabled ParsedItems
		parsed, err := service.ParseItems(ctx, transcript)
		require.NoError(t, err)

		// Reset call counters
		mockProvider.transcriptUsed = false
		mockProvider.capturedTranscript = ""
		mockProvider.transcriptContextCalls = 0

		// Test hydration without cache (should also use transcript)
		hydratedItems, err := service.HydrateNutritionWithoutCache(ctx, parsed.Items, parsed.Transcript)
		require.NoError(t, err)

		// Verify transcript was used
		assert.True(t, mockProvider.transcriptUsed, "Transcript should be used in nutrition calls")
		assert.Equal(t, transcript, mockProvider.capturedTranscript, "Captured transcript should match original")
		assert.Greater(t, mockProvider.transcriptContextCalls, 0, "Transcript context calls should be made")

		// Verify items were hydrated correctly
		assert.Len(t, hydratedItems, 1)
		assert.NotNil(t, hydratedItems[0].Nutrients)
		assert.Equal(t, expectedNutrients.Calories, hydratedItems[0].Nutrients.Calories)
	})
}
