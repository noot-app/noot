package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenAIProvider(t *testing.T) {
	config := AIProviderConfig{
		APIKey:          "test-key",
		TranscribeModel: "test-transcribe-model",
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
