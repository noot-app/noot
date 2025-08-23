package server

import (
	"context"
)

// AIProvider defines the interface for AI-based nutrition services
type AIProvider interface {
	// TranscribeAudio transcribes an audio file to text
	TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error)

	// ParseItems extracts individual food items from a meal description
	ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error)

	// GetNutrition gets nutrition data for a single food item
	GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error)

	// GetNutritionWithContext gets nutrition data for a single food item with optional context
	GetNutritionWithContext(ctx context.Context, item Item, nutritionContext interface{}) (CompleteNutrient, error)
}

// AIProviderConfig holds configuration for AI providers
type AIProviderConfig struct {
	APIKey          string
	TranscribeModel string
	ParseModel      string
	BaseURL         string
	Timeout         int // seconds
}
