package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuantityExtractionWithRealExamples(t *testing.T) {
	realWorldExamples := []struct {
		input           string
		expectedClean   string
		expectedMult    float64
		contextualGrams float64 // What grams this would typically represent
	}{
		{
			input:           "half can cream soda Olipop",
			expectedClean:   "cream soda olipop",
			expectedMult:    0.5,
			contextualGrams: 177.5, // Half of typical 355ml can
		},
		{
			input:           "2 bananas",
			expectedClean:   "bananas",
			expectedMult:    2.0,
			contextualGrams: 236, // 2 × typical 118g banana
		},
		{
			input:           "large apple",
			expectedClean:   "apple",
			expectedMult:    1.3,
			contextualGrams: 236, // 1.3 × typical 182g apple
		},
		{
			input:           "small bowl yogurt",
			expectedClean:   "yogurt",
			expectedMult:    0.75,
			contextualGrams: 184, // 0.75 × typical 245g bowl
		},
		{
			input:           "quarter cup rice",
			expectedClean:   "rice",
			expectedMult:    0.25,
			contextualGrams: 46, // 0.25 × typical 185g cup
		},
		{
			input:           "three eggs",
			expectedClean:   "eggs",
			expectedMult:    3.0,
			contextualGrams: 150, // 3 × typical 50g egg
		},
	}

	for _, example := range realWorldExamples {
		t.Run(example.input, func(t *testing.T) {
			result := extractQuantityFromName(example.input)

			assert.Equal(t, example.expectedClean, result.CleanName,
				"Clean name should match expected for '%s'", example.input)

			assert.InDelta(t, example.expectedMult, result.Multiplier, 0.01,
				"Multiplier should match expected for '%s'", example.input)

			// This test documents the expected behavior - in real usage,
			// the contextual grams would come from the base item's typical serving size
			t.Logf("'%s' → clean: '%s', mult: %.2f, contextual grams: %.1fg",
				example.input, result.CleanName, result.Multiplier, example.contextualGrams)
		})
	}
}

func TestQuantityAwareNormalization(t *testing.T) {
	testCases := []struct {
		name               string
		input              string
		expectedNormalized string
		expectedMultiplier float64
		description        string
	}{
		{
			name:               "HalfCanOlipop",
			input:              "half can cream soda Olipop",
			expectedNormalized: "cream soda olipop",
			expectedMultiplier: 0.5,
			description:        "Should extract half multiplier and clean name",
		},
		{
			name:               "TwoCansSoda",
			input:              "2 cans cream soda Olipop",
			expectedNormalized: "cream soda olipop",
			expectedMultiplier: 2.0,
			description:        "Should extract numeric multiplier and clean name",
		},
		{
			name:               "LargeBanana",
			input:              "large banana",
			expectedNormalized: "banana",
			expectedMultiplier: 1.3,
			description:        "Should extract size multiplier and clean name",
		},
		{
			name:               "PlainItem",
			input:              "yogurt",
			expectedNormalized: "yogurt",
			expectedMultiplier: 1.0,
			description:        "Plain items should have 1.0 multiplier",
		},
		{
			name:               "ComplexBrandedItem",
			input:              "Greek Yogurt Chobani",
			expectedNormalized: "greek yogurt chobani",
			expectedMultiplier: 1.0,
			description:        "Complex branded items without quantities should normalize",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalizedName, quantityInfo := normalizeItemNameWithQuantity(tc.input)

			assert.Equal(t, tc.expectedNormalized, normalizedName,
				"Normalized name mismatch for input '%s': %s", tc.input, tc.description)

			assert.InDelta(t, tc.expectedMultiplier, quantityInfo.Multiplier, 0.01,
				"Multiplier mismatch for input '%s': %s", tc.input, tc.description)
		})
	}
}

func TestQuantityExtractionEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name               string
		input              string
		expectedClean      string
		expectedMultiplier float64
		description        string
	}{
		{
			name:               "MultipleQuantifiers",
			input:              "two large apples",
			expectedClean:      "large apples", // Should only extract the first quantifier
			expectedMultiplier: 2.0,
			description:        "Should extract first quantifier only",
		},
		{
			name:               "QuantifierAtEnd",
			input:              "apple large",
			expectedClean:      "apple large", // Size at end shouldn't match
			expectedMultiplier: 1.0,
			description:        "Size qualifiers at end should not be extracted",
		},
		{
			name:               "ContainerOnly",
			input:              "can",
			expectedClean:      "can", // Single container word should remain
			expectedMultiplier: 1.0,
			description:        "Single container words should remain as is",
		},
		{
			name:               "CaseInsensitive",
			input:              "HALF CAN SODA",
			expectedClean:      "soda",
			expectedMultiplier: 0.5,
			description:        "Should work case-insensitively",
		},
		{
			name:               "ExtraSpaces",
			input:              "  half   can   soda  ",
			expectedClean:      "soda",
			expectedMultiplier: 0.5,
			description:        "Should handle extra spaces gracefully",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractQuantityFromName(tc.input)

			assert.Equal(t, tc.expectedClean, result.CleanName,
				"Clean name mismatch for input '%s': %s", tc.input, tc.description)

			assert.InDelta(t, tc.expectedMultiplier, result.Multiplier, 0.01,
				"Multiplier mismatch for input '%s': %s", tc.input, tc.description)
		})
	}
}
