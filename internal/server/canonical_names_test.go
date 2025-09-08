package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCanonicalFoodName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		brand    *string
		expected string
		description string
	}{
		{
			name:        "BasicCarrots",
			input:       "carrots",
			brand:       nil,
			expected:    "carrot",
			description: "Should normalize plural to singular",
		},
		{
			name:        "HandfulOfCarrots", 
			input:       "a handful of carrots",
			brand:       nil,
			expected:    "carrot",
			description: "Should strip quantity and normalize to singular",
		},
		{
			name:        "FewSlicesOfCarrots",
			input:       "a few slices of carrots",
			brand:       nil,
			expected:    "carrot",
			description: "Should strip descriptors and normalize to singular",
		},
		{
			name:        "TwoApples",
			input:       "2 apples",
			brand:       nil,
			expected:    "apple",
			description: "Should strip numeric quantity and normalize to singular",
		},
		{
			name:        "LargeBanana",
			input:       "large banana",
			brand:       nil,
			expected:    "banana",
			description: "Should strip size descriptors",
		},
		{
			name:        "FreshOrgaincApples",
			input:       "fresh organic apples",
			brand:       nil,
			expected:    "apple",
			description: "Should strip multiple descriptors and normalize plural",
		},
		{
			name:        "HalfCanSoda",
			input:       "half can of soda",
			brand:       nil,
			expected:    "soda",
			description: "Should strip quantity and container words",
		},
		{
			name:        "BrandedItem",
			input:       "Coca Cola",
			brand:       stringPtr("Coca Cola"),
			expected:    "",
			description: "Should remove brand completely when input is just brand",
		},
		{
			name:        "BrandedItemWithFood",
			input:       "Ben Jerry vanilla ice cream",
			brand:       stringPtr("Ben Jerry"),
			expected:    "vanilla ice cream",
			description: "Should remove brand but keep food description",
		},
		{
			name:        "ChoppedOnions",
			input:       "chopped onions",
			brand:       nil,
			expected:    "onion",
			description: "Should strip preparation method and normalize plural",
		},
		{
			name:        "GratedCheese",
			input:       "grated cheese",
			brand:       nil,
			expected:    "cheese",
			description: "Should strip preparation method",
		},
		{
			name:        "FrozenBlueberries",
			input:       "frozen blueberries",
			brand:       nil,
			expected:    "blueberry",
			description: "Should strip state descriptor and normalize plural",
		},
		{
			name:        "RedApples",
			input:       "red apples",
			brand:       nil,
			expected:    "apple",
			description: "Should strip color descriptor and normalize plural",
		},
		{
			name:        "ComplexDescriptor",
			input:       "fresh sliced red tomatoes",
			brand:       nil,
			expected:    "tomato",
			description: "Should strip multiple descriptors and normalize",
		},
		{
			name:        "ContainerWithQuantity",
			input:       "2 bottles of water",
			brand:       nil,
			expected:    "water",
			description: "Should strip quantity and container",
		},
		{
			name:        "EdgeCaseEmpty",
			input:       "",
			brand:       nil,
			expected:    "",
			description: "Should handle empty input",
		},
		{
			name:        "SingleWord",
			input:       "milk",
			brand:       nil,
			expected:    "milk",
			description: "Should handle simple single words",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateCanonicalFoodName(tc.input, tc.brand)
			assert.Equal(t, tc.expected, result, 
				"Input: '%s', Brand: %v - %s", tc.input, tc.brand, tc.description)
		})
	}
}

func TestNormalizeCanonicalName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		description string
	}{
		{
			name:        "PluralCarrots",
			input:       "carrots",
			expected:    "carrot",
			description: "Should normalize carrots to carrot",
		},
		{
			name:        "FreshDescriptor",
			input:       "fresh apples",
			expected:    "apple",
			description: "Should strip fresh descriptor and normalize plural",
		},
		{
			name:        "OrganicDescriptor",
			input:       "organic spinach",
			expected:    "spinach",
			description: "Should strip organic descriptor",
		},
		{
			name:        "MultipleDescriptors",
			input:       "fresh organic raw carrots",
			expected:    "carrot",
			description: "Should strip first descriptor and normalize plural",
		},
		{
			name:        "ColorDescriptor",
			input:       "red bell peppers",
			expected:    "bell peppers",
			description: "Should strip color descriptor",
		},
		{
			name:        "PreparationMethod",
			input:       "grilled chicken",
			expected:    "chicken",
			description: "Should strip preparation method",
		},
		{
			name:        "StateDescriptor",
			input:       "frozen strawberries",
			expected:    "strawberry",
			description: "Should strip state and normalize plural",
		},
		{
			name:        "NoChange",
			input:       "quinoa",
			expected:    "quinoa",
			description: "Should not change words without descriptors",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeCanonicalName(tc.input)
			assert.Equal(t, tc.expected, result, 
				"Input: '%s' - %s", tc.input, tc.description)
		})
	}
}

