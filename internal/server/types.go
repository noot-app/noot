package server

import (
	"github.com/grantbirki/noot/internal/storage"
)

type ParsedItems struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name          string                  `json:"name"`
	CanonicalName string                  `json:"canonical_name"` // LLM-provided canonical name for deduplication
	Grams         float64                 `json:"grams"`          // standardized weight in grams
	UserQuantity  *float64                `json:"user_quantity"`  // original user input quantity for display
	UserUnit      *string                 `json:"user_unit"`      // original user input unit for display
	Brand         *string                 `json:"brand"`
	BaseQuantity  *float64                `json:"base_quantity,omitempty"` // base quantity to normalize to (e.g., for "2 cans", this would be 2)
	Note          *string                 `json:"note,omitempty"`
	Nutrients     *CompleteNutrient       `json:"nutrients,omitempty"`
	Ingredients   []storage.OFFIngredient `json:"ingredients,omitempty"` // Ingredient list from OFF
	Url           *string                 `json:"url,omitempty"`         // Open Food Facts product URL
	Context       *string                 `json:"context"`               // Contextual information, e.g., "for a salad"
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

// ConsumptionInput represents the normalized input for consumption processing
type ConsumptionInput struct {
	Text      string // The text to process (from audio transcription or direct input)
	Source    string // "audio" or "text" to track the input source
	RequestID string // For logging purposes
}
