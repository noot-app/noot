package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractQuantityFromName(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedClean string
		expectedMult  float64
		description   string
	}{
		// Fractional quantities
		{
			name:          "HalfCan",
			input:         "half can cream soda Olipop",
			expectedClean: "cream soda olipop",
			expectedMult:  0.5,
			description:   "Half can should extract 0.5 multiplier",
		},
		{
			name:          "HalfBottle",
			input:         "half bottle of water",
			expectedClean: "water",
			expectedMult:  0.5,
			description:   "Half bottle with 'of' should work",
		},
		{
			name:          "Quarter",
			input:         "quarter cup rice",
			expectedClean: "rice",
			expectedMult:  0.25,
			description:   "Quarter should extract 0.25 multiplier",
		},
		{
			name:          "TwoThirds",
			input:         "two thirds banana",
			expectedClean: "banana",
			expectedMult:  0.67,
			description:   "Two thirds should work",
		},
		{
			name:          "FractionalSymbols",
			input:         "1/2 apple",
			expectedClean: "apple",
			expectedMult:  0.5,
			description:   "Fractional symbols should work",
		},

		// Numeric quantities
		{
			name:          "TwoCans",
			input:         "2 cans cream soda",
			expectedClean: "cream soda",
			expectedMult:  2.0,
			description:   "Numeric prefix should work",
		},
		{
			name:          "ThreeApples",
			input:         "3 apples",
			expectedClean: "apples",
			expectedMult:  3.0,
			description:   "Numeric quantities should work",
		},
		{
			name:          "TwoWord",
			input:         "two bananas",
			expectedClean: "bananas",
			expectedMult:  2.0,
			description:   "Written numbers should work",
		},
		{
			name:          "FiveCans",
			input:         "five cans of soda",
			expectedClean: "soda",
			expectedMult:  5.0,
			description:   "Written numbers with containers should work",
		},

		// Size descriptors
		{
			name:          "SmallApple",
			input:         "small apple",
			expectedClean: "apple",
			expectedMult:  0.75,
			description:   "Small size should use 0.75 multiplier",
		},
		{
			name:          "LargeBanana",
			input:         "large banana",
			expectedClean: "banana",
			expectedMult:  1.3,
			description:   "Large size should use 1.3 multiplier",
		},
		{
			name:          "JumboShrimp",
			input:         "jumbo shrimp",
			expectedClean: "shrimp",
			expectedMult:  1.8,
			description:   "Jumbo should use 1.8 multiplier",
		},
		{
			name:          "MiniCan",
			input:         "mini can soda",
			expectedClean: "soda",
			expectedMult:  0.5,
			description:   "Mini should use 0.5 multiplier",
		},
		{
			name:          "HandfulStrawberries",
			input:         "handful of strawberries",
			expectedClean: "strawberries",
			expectedMult:  0.6,
			description:   "Handful should use 0.6 multiplier",
		},
		{
			name:          "PinchSalt",
			input:         "pinch of salt",
			expectedClean: "salt",
			expectedMult:  0.05,
			description:   "Pinch should use 0.05 multiplier",
		},

		// Container removal
		{
			name:          "CanOf",
			input:         "can of beans",
			expectedClean: "beans",
			expectedMult:  1.0,
			description:   "Container words should be removed",
		},
		{
			name:          "BottleWater",
			input:         "bottle water",
			expectedClean: "water",
			expectedMult:  1.0,
			description:   "Trailing container words should be removed",
		},
		{
			name:          "CupOfCoffee",
			input:         "cup of coffee",
			expectedClean: "coffee",
			expectedMult:  1.0,
			description:   "Cup of should be handled",
		},

		// No quantity (baseline)
		{
			name:          "PlainBanana",
			input:         "banana",
			expectedClean: "banana",
			expectedMult:  1.0,
			description:   "Plain items should have 1.0 multiplier",
		},
		{
			name:          "CreamSodaOlipop",
			input:         "cream soda Olipop",
			expectedClean: "cream soda olipop",
			expectedMult:  1.0,
			description:   "Complex names without quantities should work",
		},

		// Edge cases
		{
			name:          "EmptyString",
			input:         "",
			expectedClean: "",
			expectedMult:  1.0,
			description:   "Empty string should be handled gracefully",
		},
		{
			name:          "OnlySpaces",
			input:         "   ",
			expectedClean: "",
			expectedMult:  1.0,
			description:   "Only spaces should be handled gracefully",
		},
		{
			name:          "CaseInsensitive",
			input:         "HALF CAN SODA",
			expectedClean: "soda",
			expectedMult:  0.5,
			description:   "Should work case-insensitively",
		},

		// Complex real-world examples
		{
			name:          "HalfCanOlipop",
			input:         "half can cream soda Olipop",
			expectedClean: "cream soda olipop",
			expectedMult:  0.5,
			description:   "Real example from user request",
		},
		{
			name:          "TwoCansSoda",
			input:         "2 cans cream soda Olipop",
			expectedClean: "cream soda olipop",
			expectedMult:  2.0,
			description:   "Multiple cans example",
		},
		{
			name:          "SmallBowlYogurt",
			input:         "small bowl yogurt",
			expectedClean: "yogurt",
			expectedMult:  0.75,
			description:   "Size with container",
		},
		{
			name:          "LargeCupCoffee",
			input:         "large cup of coffee",
			expectedClean: "coffee",
			expectedMult:  1.3,
			description:   "Size with container and 'of'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractQuantityFromName(tc.input)

			assert.Equal(t, tc.expectedClean, result.CleanName,
				"Clean name mismatch for input '%s': %s", tc.input, tc.description)

			assert.InDelta(t, tc.expectedMult, result.Multiplier, 0.01,
				"Multiplier mismatch for input '%s': %s", tc.input, tc.description)
		})
	}
}

func TestNormalizeItemNameWithQuantity(t *testing.T) {
	testCases := []struct {
		name               string
		input              string
		expectedNormalized string
		expectedMultiplier float64
	}{
		{
			name:               "HalfCanSoda",
			input:              "Half Can Cream Soda Olipop",
			expectedNormalized: "cream soda olipop",
			expectedMultiplier: 0.5,
		},
		{
			name:               "TwoBananas",
			input:              "Two BANANAS",
			expectedNormalized: "bananas",
			expectedMultiplier: 2.0,
		},
		{
			name:               "PlainItem",
			input:              "Greek Yogurt",
			expectedNormalized: "greek yogurt",
			expectedMultiplier: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalizedName, quantityInfo := normalizeItemNameWithQuantity(tc.input)

			assert.Equal(t, tc.expectedNormalized, normalizedName,
				"Normalized name mismatch for input '%s'", tc.input)

			assert.InDelta(t, tc.expectedMultiplier, quantityInfo.Multiplier, 0.01,
				"Multiplier mismatch for input '%s'", tc.input)
		})
	}
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkExtractQuantityFromName(b *testing.B) {
	testInputs := []string{
		"banana",
		"half can cream soda Olipop",
		"2 cans soda",
		"large apple",
		"small bowl of yogurt",
		"quarter cup rice",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := testInputs[i%len(testInputs)]
		extractQuantityFromName(input)
	}
}
