package server

import (
	"context"

	"github.com/grantbirki/noot/internal/storage"
)

// NutritionResponse represents a complete nutrition response with ingredients and URL
type NutritionResponse struct {
	Nutrients   CompleteNutrient        `json:"nutrients"`
	Ingredients []storage.OFFIngredient `json:"ingredients,omitempty"`
	URL         *string                 `json:"url,omitempty"`
}

// AIProvider defines the interface for AI-based nutrition services
type AIProvider interface {
	// TranscribeAudio transcribes an audio file to text
	TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error)

	// ParseItems extracts individual food items from a meal description
	ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error)

	// GetNutrition gets nutrition data for a single food item
	GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error)

	// GetNutritionWithContext gets nutrition data for a single food item with context
	GetNutritionWithContext(ctx context.Context, item Item) (CompleteNutrient, error)

	// GetNutritionWithContextComplete gets complete nutrition data including ingredients and URL
	GetNutritionWithContextComplete(ctx context.Context, item Item) (NutritionResponse, error)

	// GetNutritionWithTranscriptContext gets nutrition data with both item context and full transcript
	GetNutritionWithTranscriptContext(ctx context.Context, item Item, transcript string) (CompleteNutrient, error)

	// GetNutritionWithTranscriptContextComplete gets complete nutrition data with transcript context
	GetNutritionWithTranscriptContextComplete(ctx context.Context, item Item, transcript string) (NutritionResponse, error)
}

// AIProviderConfig holds configuration for AI providers
type AIProviderConfig struct {
	APIKey          string
	TranscribeModel string
	BaseURL         string
	Timeout         int // seconds
}
