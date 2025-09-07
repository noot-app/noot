package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

// Embed the AI configuration files into the binary
//
//go:embed ai
var embeddedAIConfigFS embed.FS

// AIRequestBuilderInterface defines the common interface for both filesystem and embedded builders
type AIRequestBuilderInterface interface {
	BuildRequestPayload(userInput string) (*AIRequestPayload, error)
}

// EmbeddedAIRequestBuilder handles building OpenAI API requests from embedded configuration files
type EmbeddedAIRequestBuilder struct {
	config *AIConfig
	prompt string
	schema map[string]interface{}
}

// NewEmbeddedAIRequestBuilder creates a new AI request builder using embedded files
// It loads and caches all configuration files from the embedded filesystem
func NewEmbeddedAIRequestBuilder(aiType string) (*EmbeddedAIRequestBuilder, error) {
	// Construct the path within the embedded filesystem
	configDir := path.Join("ai", aiType)

	config, err := loadEmbeddedAIConfig(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded AI config: %w", err)
	}

	prompt, err := loadEmbeddedPrompt(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded prompt: %w", err)
	}

	schema, err := loadEmbeddedSchema(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded schema: %w", err)
	}

	return &EmbeddedAIRequestBuilder{
		config: config,
		prompt: prompt,
		schema: schema,
	}, nil
}

// BuildRequestPayload creates an AI request payload with the given user input
func (b *EmbeddedAIRequestBuilder) BuildRequestPayload(userInput string) (*AIRequestPayload, error) {
	// Extract schema details
	schemaName := "nutrition_response"
	schemaStrict := true
	actualSchema := b.schema

	if name, ok := b.schema["name"].(string); ok {
		schemaName = name
	}
	if strict, ok := b.schema["strict"].(bool); ok {
		schemaStrict = strict
	}
	if nestedSchema, ok := b.schema["schema"].(map[string]interface{}); ok {
		actualSchema = nestedSchema
	}

	// Build the request payload using the cached configuration data
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

// loadEmbeddedAIConfig loads the AI configuration from embedded files
func loadEmbeddedAIConfig(configDir string) (*AIConfig, error) {
	configPath := path.Join(configDir, "config.json")
	configData, err := embeddedAIConfigFS.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config file: %w", err)
	}

	var config AIConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse embedded config file: %w", err)
	}

	// Environment variable substitution for sensitive fields
	for i, tool := range config.Tools {
		if tool.Type == "mcp" {
			if tool.Authorization != "" {
				// Expand environment variables in the authorization field
				config.Tools[i].Authorization = expandEnvVars(tool.Authorization)
			}
		}
	}

	return &config, nil
}

// loadEmbeddedPrompt loads the prompt content from embedded files
func loadEmbeddedPrompt(configDir string) (string, error) {
	promptPath := path.Join(configDir, "prompt.md")
	promptData, err := embeddedAIConfigFS.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded prompt file: %w", err)
	}
	return string(promptData), nil
}

// loadEmbeddedSchema loads the JSON schema from embedded files
func loadEmbeddedSchema(configDir string) (map[string]interface{}, error) {
	schemaPath := path.Join(configDir, "schema.json")
	schemaData, err := embeddedAIConfigFS.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded schema file: %w", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse embedded schema file: %w", err)
	}

	return schema, nil
}

// expandEnvVars expands environment variables in the format ${VAR_NAME}
func expandEnvVars(text string) string {
	// Handle ${VAR_NAME} format
	if strings.HasPrefix(text, "${") && strings.HasSuffix(text, "}") {
		varName := text[2 : len(text)-1] // Remove ${ and }
		if envValue := os.Getenv(varName); envValue != "" {
			return envValue
		}
		// If environment variable is not set, return the original text
		// This maintains backward compatibility and prevents empty values
		return text
	}
	return text
}
