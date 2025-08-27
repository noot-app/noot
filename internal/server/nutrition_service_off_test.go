package server

import (
	"context"
	"errors"
)

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
		return CompleteNutrient{}, errors.New("mock error")
	}
	return m.nutritionResponse, nil
}
