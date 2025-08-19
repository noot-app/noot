package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenAIProvider(t *testing.T) {
	config := AIProviderConfig{
		APIKey:          "test-key",
		TranscribeModel: "test-transcribe-model",
		ParseModel:      "test-parse-model",
		BaseURL:         "https://api.test.com/v1",
		Timeout:         30,
	}

	provider := NewOpenAIProvider(config)

	assert.NotNil(t, provider)
	assert.Equal(t, config, provider.config)
	assert.NotNil(t, provider.httpClient)

	// Check timeout was set correctly
	expectedTimeout := 30000000000 // 30 seconds in nanoseconds
	assert.Equal(t, expectedTimeout, int(provider.httpClient.Timeout))
}

func TestNewOpenAIProvider_DefaultTimeout(t *testing.T) {
	config := AIProviderConfig{
		APIKey:          "test-key",
		TranscribeModel: "test-transcribe-model",
		ParseModel:      "test-parse-model",
		BaseURL:         "https://api.test.com/v1",
		Timeout:         0, // Should default to 60 seconds
	}

	provider := NewOpenAIProvider(config)

	// Check default timeout was applied (60 seconds)
	expectedTimeout := 60000000000 // 60 seconds in nanoseconds
	assert.Equal(t, expectedTimeout, int(provider.httpClient.Timeout))
}

func TestOpenAIProvider_TranscriptionPrompt(t *testing.T) {
	config := AIProviderConfig{
		APIKey:          "test-key",
		TranscribeModel: "test-transcribe-model",
		ParseModel:      "test-parse-model",
		BaseURL:         "https://api.test.com/v1",
		Timeout:         30,
	}

	provider := NewOpenAIProvider(config)
	prompt := provider.transcriptionPrompt()

	// Check that the prompt contains expected keywords
	assert.Contains(t, prompt, "foods and drinks")
	assert.Contains(t, prompt, "brand")
	assert.Contains(t, prompt, "serving size")
	assert.Contains(t, prompt, "banana")
	assert.Contains(t, prompt, "latte")
}

func TestOpenAIProvider_ParseItemsSystemPrompt(t *testing.T) {
	config := AIProviderConfig{
		APIKey:          "test-key",
		TranscribeModel: "test-transcribe-model",
		ParseModel:      "test-parse-model",
		BaseURL:         "https://api.test.com/v1",
		Timeout:         30,
	}

	provider := NewOpenAIProvider(config)
	prompt := provider.parseItemsSystemPrompt()

	// Check that the prompt contains expected structure and instructions
	assert.Contains(t, prompt, "JSON")
	assert.Contains(t, prompt, "items")
	assert.Contains(t, prompt, "name")
	assert.Contains(t, prompt, "quantity")
	assert.Contains(t, prompt, "unit")
	assert.Contains(t, prompt, "brand")
	assert.Contains(t, prompt, "EXTRACT INDIVIDUAL ITEMS")
	assert.Contains(t, prompt, "INFER SERVING SIZES")
}

func TestOpenAIProvider_NutritionSystemPrompt(t *testing.T) {
	config := AIProviderConfig{
		APIKey:          "test-key",
		TranscribeModel: "test-transcribe-model",
		ParseModel:      "test-parse-model",
		BaseURL:         "https://api.test.com/v1",
		Timeout:         30,
	}

	provider := NewOpenAIProvider(config)
	prompt := provider.nutritionSystemPrompt()

	// Check that the prompt contains expected nutrition fields
	assert.Contains(t, prompt, "nutrients")
	assert.Contains(t, prompt, "calories")
	assert.Contains(t, prompt, "protein_g")
	assert.Contains(t, prompt, "total_fat_g")
	assert.Contains(t, prompt, "vitamin_a_mcg")
	assert.Contains(t, prompt, "calcium_mg")
	assert.Contains(t, prompt, "NUTRITION DATA ACCURACY")
	assert.Contains(t, prompt, "QUANTITY-ADJUSTED")
}
