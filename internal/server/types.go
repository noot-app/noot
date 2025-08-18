package server

type ParsedItems struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name     string   `json:"name"`
	Quantity *float64 `json:"quantity"` // null -> nil
	Unit     *string  `json:"unit"`
	Brand    *string  `json:"brand"`
}

type FDCRef struct {
	FDCID       int     `json:"fdcId"`
	Description string  `json:"description"`
	BrandOwner  *string `json:"brandOwner"`
	DataType    *string `json:"dataType"`
}

type Nutrients struct {
	EnergyKcal float64 `json:"energy_kcal"`
	ProteinG   float64 `json:"protein_g"`
	FatG       float64 `json:"fat_g"`
	CarbsG     float64 `json:"carbs_g"`
	FiberG     float64 `json:"fiber_g"`
	SugarG     float64 `json:"sugar_g"`
}

type ItemWithNutrition struct {
	Item  Item       `json:"item"`
	FDC   *FDCRef    `json:"fdc"`
	Nutri *Nutrients `json:"nutrients"`
	Note  string     `json:"note,omitempty"`
}

type Summary struct {
	Totals          Nutrients          `json:"totals"`
	PercentOfDaily  map[string]int     `json:"percent_of_daily"`
	DailyValuesUsed map[string]float64 `json:"daily_values"`
}
