package server

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitConverter_ConvertToGrams(t *testing.T) {
	converter := NewUnitConverter()

	tests := []struct {
		name      string
		quantity  float64
		unit      string
		expected  float64
		expectErr bool
	}{
		{
			name:     "grams to grams",
			quantity: 100,
			unit:     "g",
			expected: 100,
		},
		{
			name:     "kilograms to grams",
			quantity: 1,
			unit:     "kg",
			expected: 1000,
		},
		{
			name:     "ounces to grams",
			quantity: 1,
			unit:     "oz",
			expected: 28.3495,
		},
		{
			name:     "pounds to grams",
			quantity: 1,
			unit:     "lb",
			expected: 453.592,
		},
		{
			name:     "cups to grams",
			quantity: 1,
			unit:     "cup",
			expected: 240,
		},
		{
			name:     "tablespoons to grams",
			quantity: 1,
			unit:     "tbsp",
			expected: 15,
		},
		{
			name:     "teaspoons to grams",
			quantity: 1,
			unit:     "tsp",
			expected: 5,
		},
		{
			name:     "milliliters to grams",
			quantity: 100,
			unit:     "ml",
			expected: 100,
		},
		{
			name:     "liters to grams",
			quantity: 1,
			unit:     "l",
			expected: 1000,
		},
		{
			name:     "empty unit",
			quantity: 50,
			unit:     "",
			expected: 50,
		},
		{
			name:     "case insensitive",
			quantity: 2,
			unit:     "KG",
			expected: 2000,
		},
		{
			name:     "plural forms",
			quantity: 3,
			unit:     "cups",
			expected: 720,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.ConvertToGrams(tt.quantity, tt.unit)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.InDelta(t, tt.expected, result, 0.001, "Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestUnitConverter_ConvertFromPer100gToServing(t *testing.T) {
	converter := NewUnitConverter()

	tests := []struct {
		name         string
		per100gValue float64
		servingGrams float64
		expected     float64
	}{
		{
			name:         "50g serving",
			per100gValue: 100,
			servingGrams: 50,
			expected:     50,
		},
		{
			name:         "200g serving",
			per100gValue: 100,
			servingGrams: 200,
			expected:     200,
		},
		{
			name:         "100g serving (same as per 100g)",
			per100gValue: 150,
			servingGrams: 100,
			expected:     150,
		},
		{
			name:         "zero serving weight",
			per100gValue: 100,
			servingGrams: 0,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertFromPer100gToServing(tt.per100gValue, tt.servingGrams)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnitConverter_ConvertFromServingToPer100g(t *testing.T) {
	converter := NewUnitConverter()

	tests := []struct {
		name         string
		servingValue float64
		servingGrams float64
		expected     float64
	}{
		{
			name:         "50g serving",
			servingValue: 50,
			servingGrams: 50,
			expected:     100,
		},
		{
			name:         "200g serving",
			servingValue: 200,
			servingGrams: 200,
			expected:     100,
		},
		{
			name:         "100g serving (same as per 100g)",
			servingValue: 150,
			servingGrams: 100,
			expected:     150,
		},
		{
			name:         "zero serving weight",
			servingValue: 100,
			servingGrams: 0,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertFromServingToPer100g(tt.servingValue, tt.servingGrams)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnitConverter_EstimateServingWeight(t *testing.T) {
	converter := NewUnitConverter()

	tests := []struct {
		name     string
		item     Item
		expected float64
	}{
		{
			name: "item with quantity and standard unit (grams)",
			item: Item{
				Name:     "Rice",
				Quantity: float64Ptr(150),
				Unit:     stringPtr("g"),
			},
			expected: 150,
		},
		{
			name: "item with quantity and standard unit (cups)",
			item: Item{
				Name:     "Milk",
				Quantity: float64Ptr(1),
				Unit:     stringPtr("cup"),
			},
			expected: 240,
		},
		{
			name: "item with quantity but no unit (assumes grams)",
			item: Item{
				Name:     "Rice",
				Quantity: float64Ptr(150),
			},
			expected: 150,
		},
		{
			name: "item with no quantity (uses default 100g)",
			item: Item{
				Name: "Unknown food",
			},
			expected: 100,
		},
		{
			name: "item with non-standard unit (pieces - returns as-is since LLM handles it)",
			item: Item{
				Name:     "Apple",
				Quantity: float64Ptr(2),
				Unit:     stringPtr("pieces"),
			},
			expected: 2, // Non-standard units return the quantity as-is
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.EstimateServingWeight(tt.item)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundToDecimalPlaces(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		places   int
		expected float64
	}{
		{
			name:     "round to 1 decimal place",
			value:    3.14159,
			places:   1,
			expected: 3.1,
		},
		{
			name:     "round to 2 decimal places",
			value:    3.14159,
			places:   2,
			expected: 3.14,
		},
		{
			name:     "round to 0 decimal places",
			value:    3.7,
			places:   0,
			expected: 4,
		},
		{
			name:     "round up",
			value:    3.95,
			places:   1,
			expected: 4.0,
		},
		{
			name:     "round down",
			value:    3.94,
			places:   1,
			expected: 3.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToDecimalPlaces(tt.value, tt.places)

			// Use small epsilon for floating point comparison
			assert.True(t, math.Abs(result-tt.expected) < 0.0001,
				"Expected %v, got %v", tt.expected, result)
		})
	}
}
