package server

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAIConfig(t *testing.T) {
	// Test loading the GetNutrition config
	aiDir := filepath.Join("../../ai", "GetNutrition")
	config, err := loadAIConfig(aiDir)
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

func TestLoadPrompt(t *testing.T) {
	// Test loading the GetNutrition prompt
	aiDir := filepath.Join("../../ai", "GetNutrition")
	prompt, err := loadPrompt(aiDir)
	require.NoError(t, err)
	assert.NotEmpty(t, prompt)

	// Verify it contains expected content
	assert.Contains(t, prompt, "# Role and Objective")
	assert.Contains(t, prompt, "Generate accurate, comprehensive nutrition information")
}

func TestLoadSchema(t *testing.T) {
	// Test loading the GetNutrition schema
	aiDir := filepath.Join("../../ai", "GetNutrition")
	schema, err := loadSchema(aiDir)
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
