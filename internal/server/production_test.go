package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionScenario(t *testing.T) {
	// Test that simulates production environment where only embedded configs are available
	// This test temporarily removes the original ai directory to verify fallback

	originalAIDir := "../../ai"
	tempDir := t.TempDir()
	backupAIDir := filepath.Join(tempDir, "ai_backup")

	// Backup the original ai directory
	err := os.Rename(originalAIDir, backupAIDir)
	require.NoError(t, err)

	// Ensure we restore it after the test
	defer func() {
		os.Rename(backupAIDir, originalAIDir)
	}()

	// Now test that the ParseItemsBuilder and GetNutritionBuilder can work with only embedded configs
	parseBuilder, err := ParseItemsBuilder()
	require.NoError(t, err, "ParseItemsBuilder should work with embedded configs when files are missing")
	assert.NotNil(t, parseBuilder)

	nutritionBuilder, err := GetNutritionBuilder()
	require.NoError(t, err, "GetNutritionBuilder should work with embedded configs when files are missing")
	assert.NotNil(t, nutritionBuilder)

	// Test that we can build request payloads
	parsePayload, err := parseBuilder.BuildRequestPayload(`{"transcript_text": "I ate an apple"}`)
	require.NoError(t, err)
	assert.NotNil(t, parsePayload)
	assert.Equal(t, "gpt-4.1-mini", parsePayload.Model)

	nutritionPayload, err := nutritionBuilder.BuildRequestPayload(`{"name": "apple", "grams": 100}`)
	require.NoError(t, err)
	assert.NotNil(t, nutritionPayload)
	assert.Equal(t, "gpt-4.1-mini", nutritionPayload.Model)
}