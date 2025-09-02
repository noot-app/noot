package server

import (
	"testing"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestParseOFFIngredients(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []storage.OFFIngredient
	}{
		{
			name:     "empty input",
			input:    []interface{}{},
			expected: []storage.OFFIngredient{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: []storage.OFFIngredient{},
		},
		{
			name: "single ingredient with all fields",
			input: []interface{}{
				map[string]interface{}{
					"id":               "en:filtered-water",
					"text":             "Purified water",
					"percent_estimate": 54.1666666666667,
					"percent_max":      100.0,
					"percent_min":      8.33333333333333,
				},
			},
			expected: []storage.OFFIngredient{
				{
					ID:              "en:filtered-water",
					Text:            "Purified water",
					PercentEstimate: &[]float64{54.1666666666667}[0],
					PercentMax:      &[]float64{100.0}[0],
					PercentMin:      &[]float64{8.33333333333333}[0],
				},
			},
		},
		{
			name: "ingredient with missing fields",
			input: []interface{}{
				map[string]interface{}{
					"id":   "en:sugar",
					"text": "Sugar",
				},
			},
			expected: []storage.OFFIngredient{
				{
					ID:   "en:sugar",
					Text: "Sugar",
				},
			},
		},
		{
			name: "multiple ingredients",
			input: []interface{}{
				map[string]interface{}{
					"id":   "en:water",
					"text": "Water",
				},
				map[string]interface{}{
					"id":               "en:sugar",
					"text":             "Sugar",
					"percent_estimate": 10.5,
				},
			},
			expected: []storage.OFFIngredient{
				{
					ID:   "en:water",
					Text: "Water",
				},
				{
					ID:              "en:sugar",
					Text:            "Sugar",
					PercentEstimate: &[]float64{10.5}[0],
				},
			},
		},
		{
			name: "invalid ingredient type",
			input: []interface{}{
				"not a map",
				map[string]interface{}{
					"id":   "en:valid",
					"text": "Valid ingredient",
				},
			},
			expected: []storage.OFFIngredient{
				{
					ID:   "en:valid",
					Text: "Valid ingredient",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOFFIngredients(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFloatFromInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected *float64
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "float64 input",
			input:    42.5,
			expected: &[]float64{42.5}[0],
		},
		{
			name:     "float32 input",
			input:    float32(42.5),
			expected: &[]float64{42.5}[0],
		},
		{
			name:     "int input",
			input:    42,
			expected: &[]float64{42.0}[0],
		},
		{
			name:     "string number input",
			input:    "42.5",
			expected: &[]float64{42.5}[0],
		},
		{
			name:     "invalid string input",
			input:    "not a number",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFloatFromInterface(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}
