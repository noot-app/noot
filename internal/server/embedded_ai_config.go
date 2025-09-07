package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Embed all AI configuration files into the binary
// Using symlinked ai directory to make paths work
//go:embed ai/ParseItems/config.json ai/ParseItems/prompt.md ai/ParseItems/schema.json
//go:embed ai/GetNutrition/config.json ai/GetNutrition/prompt.md ai/GetNutrition/schema.json
var embeddedAIFiles embed.FS

// loadEmbeddedAIConfig loads AI configuration from embedded files
func loadEmbeddedAIConfig(aiDir string) (*AIConfig, error) {
	configPath := filepath.Join(aiDir, "config.json")
	configData, err := embeddedAIFiles.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config file: %w", err)
	}

	var config AIConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse embedded config file: %w", err)
	}

	// Environment variable substitution still needed for embedded configs
	for i, tool := range config.Tools {
		if tool.Type == "mcp" {
			if tool.Authorization != "" {
				config.Tools[i].Authorization = os.ExpandEnv(tool.Authorization)
			}
		}
	}

	return &config, nil
}

// loadEmbeddedPrompt loads prompt content from embedded files
func loadEmbeddedPrompt(aiDir string) (string, error) {
	promptPath := filepath.Join(aiDir, "prompt.md")
	promptData, err := embeddedAIFiles.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded prompt file: %w", err)
	}
	return string(promptData), nil
}

// loadEmbeddedSchema loads JSON schema from embedded files
func loadEmbeddedSchema(aiDir string) (map[string]interface{}, error) {
	schemaPath := filepath.Join(aiDir, "schema.json")
	schemaData, err := embeddedAIFiles.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded schema file: %w", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse embedded schema file: %w", err)
	}

	return schema, nil
}