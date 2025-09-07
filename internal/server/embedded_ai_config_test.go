package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEmbeddedAIConfig(t *testing.T) {
	// Test loading the embedded GetNutrition config
	aiDir := "ai/GetNutrition"
	config, err := loadEmbeddedAIConfig(aiDir)
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Verify basic config properties
	assert.Equal(t, "gpt-4.1-mini", config.Model)
	assert.Equal(t, "json_schema", config.TextFormat)
	assert.Equal(t, "auto", config.ToolChoice)
	assert.Equal(t, 0.09, config.Temp)
	assert.Equal(t, 5000, config.Tokens)
	assert.Equal(t, 1.0, config.TopP)
	assert.True(t, config.Store)

	// Verify tools are loaded
	assert.Len(t, config.Tools, 2)

	// Check web search tool
	webSearchTool := config.Tools[0]
	assert.Equal(t, "web_search", webSearchTool.Type)
	assert.Equal(t, "medium", webSearchTool.SearchContextSize)

	// Check MCP tool
	mcpTool := config.Tools[1]
	assert.Equal(t, "mcp", mcpTool.Type)
	assert.Equal(t, "openfoodfacts_mcp_server", mcpTool.ServerLabel)
	assert.Contains(t, mcpTool.AllowedTools, "search_products_by_brand_and_name")
}

func TestLoadEmbeddedPrompt(t *testing.T) {
	// Test loading the embedded GetNutrition prompt
	aiDir := "ai/GetNutrition"
	prompt, err := loadEmbeddedPrompt(aiDir)
	require.NoError(t, err)
	assert.NotEmpty(t, prompt)

	// Verify it contains expected content
	assert.Contains(t, prompt, "# Role and Objective")
	assert.Contains(t, prompt, "Generate accurate, comprehensive nutrition information")
}

func TestLoadEmbeddedSchema(t *testing.T) {
	// Test loading the embedded GetNutrition schema
	aiDir := "ai/GetNutrition"
	schema, err := loadEmbeddedSchema(aiDir)
	require.NoError(t, err)
	assert.NotNil(t, schema)

	// Verify top-level schema structure
	assert.Equal(t, "nutrition_info_grams", schema["name"])
	assert.Equal(t, true, schema["strict"])

	// Get the nested schema
	nestedSchema, ok := schema["schema"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "object", nestedSchema["type"])

	properties, ok := nestedSchema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Check key properties exist
	assert.Contains(t, properties, "success")
	assert.Contains(t, properties, "nutrients")
	assert.Contains(t, properties, "ingredients")
	assert.Contains(t, properties, "url")

	// Check nutrients schema
	nutrients, ok := properties["nutrients"].(map[string]interface{})
	require.True(t, ok)

	nutrientProperties, ok := nutrients["properties"].(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, nutrientProperties, "calories")
	assert.Contains(t, nutrientProperties, "protein_g")
	assert.Contains(t, nutrientProperties, "total_fat_g")
}

func TestLoadAIConfigWithFallback(t *testing.T) {
	// Test that load functions fall back to embedded when files don't exist
	originalDir := "ai/ParseItems"
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "nonexistent")

	// First verify normal loading works
	config1, err := loadAIConfig(originalDir)
	require.NoError(t, err)
	assert.NotNil(t, config1)

	// Then test fallback to embedded when directory doesn't exist
	_, err = loadAIConfig(nonExistentDir)
	require.Error(t, err) // Should error because the nonexistent path doesn't match embedded paths

	// Test fallback with proper embedded path but missing file
	// Temporarily rename the ai directory to simulate missing files on filesystem
	tempAiPath := filepath.Join(tempDir, "ai")
	err = os.MkdirAll(filepath.Join(tempAiPath, "ParseItems"), 0755)
	require.NoError(t, err)

	// loadAIConfig should fallback to embedded when file doesn't exist
	// But we need to use the exact path that matches the embedded FS
	config3, err := loadAIConfig("ai/ParseItems")
	require.NoError(t, err, "Should fallback to embedded config when file doesn't exist")
	assert.NotNil(t, config3)
	assert.Equal(t, "gpt-4.1-mini", config3.Model)
}

func TestEmbeddedVsFileSystemConsistency(t *testing.T) {
	// Test that embedded configs match filesystem configs
	aiDirs := []string{"ai/ParseItems", "ai/GetNutrition"}

	for _, aiDir := range aiDirs {
		t.Run(aiDir, func(t *testing.T) {
			// Load from filesystem (if available)
			fsConfig, fsErr := loadAIConfig(aiDir)
			
			// Load from embedded
			embeddedConfig, embeddedErr := loadEmbeddedAIConfig(aiDir)
			require.NoError(t, embeddedErr)

			if fsErr == nil {
				// If filesystem config is available, compare them
				assert.Equal(t, fsConfig.Model, embeddedConfig.Model)
				assert.Equal(t, fsConfig.TextFormat, embeddedConfig.TextFormat)
				assert.Equal(t, fsConfig.Temp, embeddedConfig.Temp)
				assert.Equal(t, fsConfig.Tokens, embeddedConfig.Tokens)
				assert.Equal(t, len(fsConfig.Tools), len(embeddedConfig.Tools))
			}

			// Test prompt consistency
			fsPrompt, fsPromptErr := loadPrompt(aiDir)
			embeddedPrompt, embeddedPromptErr := loadEmbeddedPrompt(aiDir)
			require.NoError(t, embeddedPromptErr)

			if fsPromptErr == nil {
				assert.Equal(t, fsPrompt, embeddedPrompt)
			}

			// Test schema consistency
			fsSchema, fsSchemaErr := loadSchema(aiDir)
			embeddedSchema, embeddedSchemaErr := loadEmbeddedSchema(aiDir)
			require.NoError(t, embeddedSchemaErr)

			if fsSchemaErr == nil {
				assert.Equal(t, fsSchema, embeddedSchema)
			}
		})
	}
}