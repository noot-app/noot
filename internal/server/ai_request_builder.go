package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AIRequestBuilder handles building OpenAI API requests from configuration files
type AIRequestBuilder struct {
	config *AIConfig
	prompt string
	schema map[string]interface{}
}

// NewAIRequestBuilder creates a new AI request builder for the specified directory
// It loads and caches all configuration files at initialization time
// Panics if configuration files cannot be loaded
func NewAIRequestBuilder(aiDir string) *AIRequestBuilder {
	builder, err := NewAIRequestBuilderSafe(aiDir)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize AI request builder: %v", err))
	}
	return builder
}

// NewAIRequestBuilderSafe creates a new AI request builder for the specified directory
// It loads and caches all configuration files at initialization time
// Returns an error if configuration files cannot be loaded
func NewAIRequestBuilderSafe(aiDir string) (*AIRequestBuilder, error) {
	config, err := loadAIConfig(aiDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load AI config: %w", err)
	}

	prompt, err := loadPrompt(aiDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt: %w", err)
	}

	schema, err := loadSchema(aiDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return &AIRequestBuilder{
		config: config,
		prompt: prompt,
		schema: schema,
	}, nil
}

// AIRequestPayload represents the structure for OpenAI responses API requests
type AIRequestPayload struct {
	Model           string                   `json:"model"`
	Input           []map[string]interface{} `json:"input"`
	Text            map[string]interface{}   `json:"text"`
	Reasoning       AIReasoning              `json:"reasoning"`
	Tools           []AITool                 `json:"tools"`
	Temperature     float64                  `json:"temperature"`
	MaxOutputTokens int                      `json:"max_output_tokens"`
	TopP            float64                  `json:"top_p"`
	Store           bool                     `json:"store"`
	Include         []string                 `json:"include"`
}

// BuildRequestPayload constructs a complete OpenAI API request payload from cached config data
func (b *AIRequestBuilder) BuildRequestPayload(userInput string) (*AIRequestPayload, error) {
	// Extract schema components from the loaded schema file
	schemaName := "default_schema_name" // default fallback
	schemaStrict := true                // default fallback
	actualSchema := b.schema            // fallback to full schema if structure is unexpected

	// Check if the schema follows the new structure with name, strict, and nested schema
	if name, ok := b.schema["name"].(string); ok {
		schemaName = name
	}
	if strict, ok := b.schema["strict"].(bool); ok {
		schemaStrict = strict
	}
	if nestedSchema, ok := b.schema["schema"].(map[string]interface{}); ok {
		actualSchema = nestedSchema
	}

	// Build the request payload using the cached OpenAI responses API format data
	payload := &AIRequestPayload{
		Model: b.config.Model,
		Input: []map[string]interface{}{
			{
				"role": "system",
				"content": []map[string]interface{}{
					{
						"type": "input_text",
						"text": b.prompt,
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "input_text",
						"text": userInput,
					},
				},
			},
		},
		Text: map[string]interface{}{
			"format": map[string]interface{}{
				"type":   "json_schema",
				"name":   schemaName,
				"strict": schemaStrict,
				"schema": actualSchema,
			},
		},
		Reasoning:       b.config.Reasoning,
		Tools:           b.config.Tools,
		Temperature:     b.config.Temp,
		MaxOutputTokens: b.config.Tokens,
		TopP:            b.config.TopP,
		Store:           b.config.Store,
		Include:         b.config.Include,
	}

	return payload, nil
}

// loadAIConfig loads the AI configuration from files in the specified directory
// Falls back to embedded configs if files are not found
func loadAIConfig(aiDir string) (*AIConfig, error) {
	configPath := filepath.Join(aiDir, "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		// If file doesn't exist, try embedded config
		return loadEmbeddedAIConfig(aiDir)
	}

	var config AIConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Substitute environment variables in tools (only for sensitive fields)
	for i, tool := range config.Tools {
		if tool.Type == "mcp" {
			if tool.Authorization != "" {
				config.Tools[i].Authorization = os.ExpandEnv(tool.Authorization)
			}
		}
	}

	return &config, nil
}

// loadPrompt loads the prompt content from the specified directory
// Falls back to embedded prompt if file is not found
func loadPrompt(aiDir string) (string, error) {
	promptPath := filepath.Join(aiDir, "prompt.md")
	promptData, err := os.ReadFile(promptPath)
	if err != nil {
		// If file doesn't exist, try embedded prompt
		return loadEmbeddedPrompt(aiDir)
	}
	return string(promptData), nil
}

// loadSchema loads the JSON schema from the specified directory
// Falls back to embedded schema if file is not found
func loadSchema(aiDir string) (map[string]interface{}, error) {
	schemaPath := filepath.Join(aiDir, "schema.json")
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		// If file doesn't exist, try embedded schema
		return loadEmbeddedSchema(aiDir)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	return schema, nil
}
