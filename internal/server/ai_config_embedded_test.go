package server

import (
	"testing"
)

func TestEmbeddedAIRequestBuilder(t *testing.T) {
	tests := []struct {
		name   string
		aiType string
	}{
		{
			name:   "GetNutrition embedded config",
			aiType: "GetNutrition",
		},
		{
			name:   "ParseItems embedded config",
			aiType: "ParseItems",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewEmbeddedAIRequestBuilder(tt.aiType)
			if err != nil {
				t.Fatalf("Failed to create embedded AI request builder: %v", err)
			}

			if builder == nil {
				t.Fatal("Builder should not be nil")
			}

			// Test building a request
			payload, err := builder.BuildRequestPayload("test input")
			if err != nil {
				t.Fatalf("Failed to build request payload: %v", err)
			}

			if payload == nil {
				t.Fatal("Payload should not be nil")
			}

			if payload.Model == "" {
				t.Error("Model should not be empty")
			}

			if len(payload.Input) == 0 {
				t.Error("Input should not be empty")
			}
		})
	}
}

func TestEmbeddedConfigFallback(t *testing.T) {
	// Test that the builders can initialize with embedded configs
	nutritionBuilder, err := GetNutritionBuilder()
	if err != nil {
		t.Fatalf("Failed to get nutrition builder: %v", err)
	}
	if nutritionBuilder == nil {
		t.Fatal("Nutrition builder should not be nil")
	}

	parseItemsBuilder, err := ParseItemsBuilder()
	if err != nil {
		t.Fatalf("Failed to get parse items builder: %v", err)
	}
	if parseItemsBuilder == nil {
		t.Fatal("Parse items builder should not be nil")
	}
}
