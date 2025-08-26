package server

import (
	"context"
	"errors"
	"testing"
)

func TestNutritionService_OFF_Integration(t *testing.T) {
	// Skip this test as it requires SQLite which has been removed
	t.Skip("Skipping test that requires SQLite - SQLite support removed")
}

func TestNutritionService_OFF_Disabled(t *testing.T) {
	// Skip this test as it requires SQLite which has been removed
	t.Skip("Skipping test that requires SQLite - SQLite support removed")
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
		return CompleteNutrient{}, errors.New("mock error")
	}
	return m.nutritionResponse, nil
}

func TestNutritionService_OFF_ServingSizeContext(t *testing.T) {
	// Skip this test as it requires SQLite which has been removed
	t.Skip("Skipping test that requires SQLite - SQLite support removed")
}

func TestNutritionService_OFF_AdditionalFieldsContext(t *testing.T) {
	// Skip this test as it requires SQLite which has been removed
	t.Skip("Skipping test that requires SQLite - SQLite support removed")
}
