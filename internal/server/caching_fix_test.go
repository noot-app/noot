package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuantityBasedCachingFix(t *testing.T) {
	// Test that quantity extraction produces the correct values for caching
	tests := []struct {
		name                      string
		userInput                 string
		expectedMultiplier        float64
		expectedCleanName         string
		expectedReverseMultiplier float64
		description               string
	}{
		{
			name:                      "Half can should cache full can data",
			userInput:                 "half can cream soda",
			expectedMultiplier:        0.5,
			expectedCleanName:         "cream soda",
			expectedReverseMultiplier: 2.0,
			description:               "Half can should reverse scale to full can for caching",
		},
		{
			name:                      "Quarter cup should cache full cup data",
			userInput:                 "quarter cup rice",
			expectedMultiplier:        0.25,
			expectedCleanName:         "rice",
			expectedReverseMultiplier: 4.0,
			description:               "Quarter serving should reverse scale to full serving for caching",
		},
		{
			name:                      "Two cans should cache single can data",
			userInput:                 "2 cans soda",
			expectedMultiplier:        2.0,
			expectedCleanName:         "soda",
			expectedReverseMultiplier: 0.5,
			description:               "Multi servings should reverse scale to single serving for caching",
		},
		{
			name:                      "Regular item needs no scaling",
			userInput:                 "banana",
			expectedMultiplier:        1.0,
			expectedCleanName:         "banana",
			expectedReverseMultiplier: 1.0,
			description:               "Regular items should not be scaled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantityInfo := extractQuantityFromName(tt.userInput)

			// Verify quantity extraction
			assert.Equal(t, tt.expectedMultiplier, quantityInfo.Multiplier, tt.description)
			assert.Equal(t, tt.expectedCleanName, quantityInfo.CleanName, tt.description)

			// Verify reverse scaling calculation
			reverseMultiplier := 1.0 / quantityInfo.Multiplier
			assert.InDelta(t, tt.expectedReverseMultiplier, reverseMultiplier, 0.001,
				"Reverse multiplier should be %.3f for %s", tt.expectedReverseMultiplier, tt.userInput)

			t.Logf("✅ '%s': quantity=%.2fx, clean='%s', reverse=%.2fx",
				tt.userInput, quantityInfo.Multiplier, quantityInfo.CleanName, reverseMultiplier)
		})
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	// Test that cache keys are generated correctly for base items
	tests := []struct {
		name              string
		userInput         string
		userGrams         float64
		expectedBaseGrams float64
		expectedCleanName string
		description       string
	}{
		{
			name:              "Half can creates full can cache key",
			userInput:         "half can cream soda",
			userGrams:         177.5, // Half of 355g
			expectedBaseGrams: 355.0, // Full can
			expectedCleanName: "cream soda",
			description:       "Should cache under full can size, not half can",
		},
		{
			name:              "Quarter cup creates full cup cache key",
			userInput:         "quarter cup rice",
			userGrams:         46.0,  // Quarter of 184g
			expectedBaseGrams: 184.0, // Full cup
			expectedCleanName: "rice",
			description:       "Should cache under full cup size, not quarter cup",
		},
		{
			name:              "Regular item uses same grams",
			userInput:         "apple",
			userGrams:         182.0,
			expectedBaseGrams: 182.0, // Same
			expectedCleanName: "apple",
			description:       "Regular items should use same grams for caching",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quantityInfo := extractQuantityFromName(tt.userInput)

			var baseGrams float64
			if quantityInfo.Multiplier != 1.0 {
				baseGrams = tt.userGrams / quantityInfo.Multiplier
			} else {
				baseGrams = tt.userGrams
			}

			assert.Equal(t, tt.expectedCleanName, quantityInfo.CleanName, tt.description)
			assert.InDelta(t, tt.expectedBaseGrams, baseGrams, 0.1, tt.description)

			t.Logf("✅ '%s' (%.1fg) → base: '%.1fg' for cache key",
				tt.userInput, tt.userGrams, baseGrams)
		})
	}
}
