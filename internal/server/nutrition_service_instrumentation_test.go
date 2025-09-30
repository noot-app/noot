package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestHydrationTiming tests that the timing instrumentation works correctly
func TestHydrationTiming(t *testing.T) {
	// Initialize logger for the test
	InitLogger()

	// Create a mock AI provider that has controlled delays
	mockProvider := &MockAIProvider{
		delay: 100 * time.Millisecond, // Simulate 100ms AI call
	}

	// Create service with mock provider
	service := &NutritionService{
		aiProvider: mockProvider,
		converter:  NewUnitConverter(),
		store:      nil, // No cache for this test
	}

	// Create test items
	items := []Item{
		{Name: "apple", Grams: 100},
		{Name: "banana", Grams: 120},
		{Name: "orange", Grams: 150},
	}

	// Hydrate without cache (all items should hit AI)
	start := time.Now()
	result, err := service.HydrateNutritionWithoutCache(context.Background(), items, "I ate fruits")
	duration := time.Since(start)

	// Verify no errors
	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))

	// With 6 concurrent slots and 3 items, all should run in parallel
	// So total time should be ~100ms (one AI call) not 300ms (sequential)
	assert.Less(t, duration, 200*time.Millisecond, "Should complete in parallel, not sequentially")
	assert.Greater(t, duration, 100*time.Millisecond, "Should take at least as long as one AI call")
}

// MockAIProvider simulates AI provider with controlled delay
type MockAIProvider struct {
	delay time.Duration
}

func (m *MockAIProvider) TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error) {
	time.Sleep(m.delay)
	return "mock transcript", nil
}

func (m *MockAIProvider) ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	time.Sleep(m.delay)
	return ParsedItems{
		Items: []Item{
			{Name: "test", Grams: 100},
		},
	}, nil
}

func (m *MockAIProvider) GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error) {
	time.Sleep(m.delay)
	return CompleteNutrient{
		Calories: 100,
	}, nil
}

func (m *MockAIProvider) GetNutritionWithContext(ctx context.Context, item Item) (CompleteNutrient, error) {
	time.Sleep(m.delay)
	return CompleteNutrient{
		Calories: 100,
	}, nil
}

func (m *MockAIProvider) GetNutritionWithContextComplete(ctx context.Context, item Item) (NutritionResponse, error) {
	time.Sleep(m.delay)
	return NutritionResponse{
		Nutrients: CompleteNutrient{
			Calories: 100,
		},
	}, nil
}

func (m *MockAIProvider) GetNutritionWithTranscriptContext(ctx context.Context, item Item, transcript string) (CompleteNutrient, error) {
	time.Sleep(m.delay)
	return CompleteNutrient{
		Calories: 100,
	}, nil
}

func (m *MockAIProvider) GetNutritionWithTranscriptContextComplete(ctx context.Context, item Item, transcript string) (NutritionResponse, error) {
	time.Sleep(m.delay)
	return NutritionResponse{
		Nutrients: CompleteNutrient{
			Calories: 100,
		},
	}, nil
}
