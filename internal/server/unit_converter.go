package server

import (
	"math"
	"strings"
)

// UnitConverter handles unit conversions for nutrition calculations
type UnitConverter struct{}

// NewUnitConverter creates a new unit converter
func NewUnitConverter() *UnitConverter {
	return &UnitConverter{}
}

// ConvertToGrams converts standard measurement units to grams
// Note: This only handles true measurement units, not food-specific units like "pieces"
func (c *UnitConverter) ConvertToGrams(quantity float64, unit string) (float64, error) {
	if unit == "" {
		return quantity, nil
	}

	unit = strings.ToLower(strings.TrimSpace(unit))

	switch unit {
	case "g", "gram", "grams":
		return quantity, nil
	case "kg", "kilogram", "kilograms":
		return quantity * 1000, nil
	case "oz", "ounce", "ounces":
		return quantity * 28.3495, nil
	case "lb", "pound", "pounds":
		return quantity * 453.592, nil
	case "cup", "cups":
		// Generic cup conversion (varies by food type, using water as default)
		return quantity * 240, nil
	case "tbsp", "tablespoon", "tablespoons":
		return quantity * 15, nil
	case "tsp", "teaspoon", "teaspoons":
		return quantity * 5, nil
	case "ml", "milliliter", "milliliters":
		// Assuming density of water (1ml = 1g)
		return quantity, nil
	case "l", "liter", "liters":
		return quantity * 1000, nil
	default:
		// For non-standard units like "pieces", "items", "medium", etc.
		// we let the LLM handle the conversion since it has better context
		// Return as-is
		return quantity, nil
	}
}

// ConvertFromPer100gToServing converts nutrition values from per-100g to actual serving size
func (c *UnitConverter) ConvertFromPer100gToServing(per100gValue float64, servingGrams float64) float64 {
	if servingGrams <= 0 {
		return 0
	}
	return per100gValue * (servingGrams / 100.0)
}

// ConvertFromServingToPer100g converts nutrition values from serving size to per-100g
func (c *UnitConverter) ConvertFromServingToPer100g(servingValue float64, servingGrams float64) float64 {
	if servingGrams <= 0 {
		return 0
	}
	return servingValue * (100.0 / servingGrams)
}

// EstimateServingWeight attempts to estimate weight for standard measurement units only
// For food-specific units like "pieces" or "medium", the LLM handles the conversion
func (c *UnitConverter) EstimateServingWeight(item Item) (float64, error) {
	if item.Quantity == nil {
		// No quantity specified - use a default weight for per-100g calculations
		return 100.0, nil
	}

	quantity := *item.Quantity
	unit := ""
	if item.Unit != nil {
		unit = *item.Unit
	}

	// If no unit specified, assume grams for numeric quantities
	if unit == "" {
		return quantity, nil
	}

	// Try to convert standard measurement units to grams
	// For food-specific units, just return the quantity since LLM handles the conversion
	if grams, err := c.ConvertToGrams(quantity, unit); err == nil {
		return grams, nil
	}

	// For non-standard units, return the original quantity
	// The LLM already converted this to proper serving nutrition values
	return quantity, nil
}

// RoundToDecimalPlaces rounds a float64 to the specified number of decimal places
func RoundToDecimalPlaces(value float64, places int) float64 {
	multiplier := math.Pow(10, float64(places))
	return math.Round(value*multiplier) / multiplier
}
