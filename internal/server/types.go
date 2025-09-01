package server

import (
	"github.com/grantbirki/noot/internal/storage"
)

type ParsedItems struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name         string                  `json:"name"`
	Grams        float64                 `json:"grams"`         // standardized weight in grams
	UserQuantity *float64                `json:"user_quantity"` // original user input quantity for display
	UserUnit     *string                 `json:"user_unit"`     // original user input unit for display
	Brand        *string                 `json:"brand"`
	BaseQuantity *float64                `json:"base_quantity,omitempty"` // base quantity to normalize to (e.g., for "2 cans", this would be 2)
	Note         *string                 `json:"note,omitempty"`
	Nutrients    *CompleteNutrient       `json:"nutrients,omitempty"`
	Ingredients  []storage.OFFIngredient `json:"ingredients,omitempty"` // Ingredient list from OFF
	OFFUrl       *string                 `json:"off_url,omitempty"`     // Open Food Facts product URL
}

type ItemWithNutrition struct {
	Item Item   `json:"item"`
	Note string `json:"note,omitempty"`
}

type Summary struct {
	Totals          CompleteNutrient   `json:"totals"`
	PercentOfDaily  map[string]int     `json:"percent_of_daily"`
	DailyValuesUsed map[string]float64 `json:"daily_values"`
}
