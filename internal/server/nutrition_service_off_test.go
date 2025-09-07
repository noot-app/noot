package server

import (
	"context"
	"errors"

	"github.com/grantbirki/noot/internal/storage"
)

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
	return m.GetNutritionWithContext(ctx, item)
}

func (m *mockAIProvider) GetNutritionWithContext(ctx context.Context, item Item) (CompleteNutrient, error) {
	if m.shouldError {
		return CompleteNutrient{}, errors.New("mock error")
	}
	return m.nutritionResponse, nil
}

func (m *mockAIProvider) GetNutritionWithContextComplete(ctx context.Context, item Item) (NutritionResponse, error) {
	if m.shouldError {
		return NutritionResponse{}, errors.New("mock error")
	}
	return NutritionResponse{
		Nutrients:   m.nutritionResponse,
		Ingredients: []storage.OFFIngredient{}, // Empty for tests
		URL:         nil,
	}, nil
}
