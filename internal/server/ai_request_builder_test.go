package server

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAIRequestBuilder(t *testing.T) {
	builder := NewAIRequestBuilder(filepath.Join("../../ai", "GetNutrition"))
	assert.NotNil(t, builder)

	// Verify that the config data was loaded successfully
	assert.NotNil(t, builder.config)
	assert.NotEmpty(t, builder.prompt)
	assert.NotNil(t, builder.schema)

	// Verify some expected config values
	assert.Equal(t, "gpt-4.1-mini", builder.config.Model)
}

func TestAIRequestBuilder_BuildRequestPayload(t *testing.T) {
	// Test with the actual GetNutrition configuration using relative path from repository root
	builder := NewAIRequestBuilder(filepath.Join("../../ai", "GetNutrition"))

	userInput := `{"name": "apple", "grams": 100, "brand": null, "context": null, "nutrition_context": null}`

	payload, err := builder.BuildRequestPayload(userInput)
	require.NoError(t, err)
	assert.NotNil(t, payload)

	// Verify basic structure
	assert.Equal(t, "gpt-4.1-mini", payload.Model)
	assert.Equal(t, 0.09, payload.Temperature)
	assert.Equal(t, 5000, payload.MaxOutputTokens)
	assert.Equal(t, 1.0, payload.TopP)
	assert.True(t, payload.Store)

	// Verify input structure
	require.Len(t, payload.Input, 2)

	// Check system message
	systemMsg := payload.Input[0]
	assert.Equal(t, "system", systemMsg["role"])
	content, ok := systemMsg["content"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, content, 1)
	assert.Equal(t, "input_text", content[0]["type"])
	assert.Contains(t, content[0]["text"], "Role and Objective")

	// Check user message
	userMsg := payload.Input[1]
	assert.Equal(t, "user", userMsg["role"])
	userContent, ok := userMsg["content"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, userContent, 1)
	assert.Equal(t, "input_text", userContent[0]["type"])
	assert.Equal(t, userInput, userContent[0]["text"])

	// Verify text format structure
	textFormat, ok := payload.Text["format"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "json_schema", textFormat["type"])
	assert.Equal(t, "nutrition_info_grams", textFormat["name"])
	assert.True(t, textFormat["strict"].(bool))

	schema, ok := textFormat["schema"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "object", schema["type"])

	// Verify tools are loaded
	assert.Len(t, payload.Tools, 2)

	// Check web search tool
	webSearchTool := payload.Tools[0]
	assert.Equal(t, "web_search", webSearchTool.Type)
	assert.Equal(t, "medium", webSearchTool.SearchContextSize)

	// Check MCP tool
	mcpTool := payload.Tools[1]
	assert.Equal(t, "mcp", mcpTool.Type)
	assert.Equal(t, "openfoodfacts_mcp_server", mcpTool.ServerLabel)
	assert.Contains(t, mcpTool.AllowedTools, "search_products_by_brand_and_name")

	// Verify include array
	assert.Contains(t, payload.Include, "web_search_call.action.sources")
}

func TestAIRequestBuilder_BuildRequestPayload_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		aiDir     string
		expectErr bool
	}{
		{
			name:      "invalid directory",
			aiDir:     "nonexistent/directory",
			expectErr: true,
		},
		{
			name:      "valid directory",
			aiDir:     filepath.Join("../../ai", "GetNutrition"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				// Use safe version for error testing
				builder, err := NewAIRequestBuilderSafe(tt.aiDir)
				assert.Error(t, err)
				assert.Nil(t, builder)
			} else {
				// Use regular version for success testing
				builder := NewAIRequestBuilder(tt.aiDir)
				payload, err := builder.BuildRequestPayload("test input")
				assert.NoError(t, err)
				assert.NotNil(t, payload)
			}
		})
	}
}

func TestAIRequestPayload_Structure(t *testing.T) {
	// Test that the payload structure is correct for JSON marshaling
	builder := NewAIRequestBuilder(filepath.Join("../../ai", "GetNutrition"))
	payload, err := builder.BuildRequestPayload("test")
	require.NoError(t, err)

	// Verify all required fields are present
	assert.NotEmpty(t, payload.Model)
	assert.NotNil(t, payload.Input)
	assert.NotNil(t, payload.Text)
	assert.NotNil(t, payload.Tools)
	assert.Greater(t, payload.Temperature, 0.0)
	assert.Greater(t, payload.MaxOutputTokens, 0)
	assert.Greater(t, payload.TopP, 0.0)
}
