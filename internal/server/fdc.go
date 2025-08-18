package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	fdcSearchURL = "https://api.nal.usda.gov/fdc/v1/foods/search"
	fdcFoodURL   = "https://api.nal.usda.gov/fdc/v1/food/"
)

type fdcSearchResp struct {
	Foods []struct {
		FdcID       int     `json:"fdcId"`
		Description string  `json:"description"`
		BrandOwner  *string `json:"brandOwner"`
		DataType    *string `json:"dataType"`
	} `json:"foods"`
}

type fdcSearchRespFood struct {
	FdcID       int     `json:"fdcId"`
	Description string  `json:"description"`
	BrandOwner  *string `json:"brandOwner"`
	DataType    *string `json:"dataType"`
}

func fdcSearchBestMatch(ctx context.Context, query string) (*fdcSearchRespFood, error) {
	apiKey := os.Getenv("USDA_FDC_API_KEY")
	if apiKey == "" {
		return nil, NewAppError("USDA_FDC_API_KEY not configured", http.StatusInternalServerError, nil)
	}

	LogDebug("Starting FDC search", "query", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fdcSearchURL, nil)
	if err != nil {
		return nil, NewAppError("Failed to create FDC search request", http.StatusInternalServerError, err)
	}
	q := req.URL.Query()
	q.Set("api_key", apiKey)
	q.Set("query", query)
	q.Set("dataType", "Branded,Foundation,SR Legacy")
	q.Set("pageSize", "5")
	q.Set("pageNumber", "1")
	q.Set("sortBy", "dataType.keyword")
	q.Set("sortOrder", "asc")
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, NewAppError("FDC search request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		LogError("FDC search error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body)))
		return nil, NewAppError("FDC search failed", http.StatusInternalServerError,
			fmt.Errorf("FDC API error: %d %s", resp.StatusCode, string(body)))
	}
	var out fdcSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, NewAppError("Failed to parse FDC search response", http.StatusInternalServerError, err)
	}
	if len(out.Foods) == 0 {
		LogDebug("No FDC results found", "query", query)
		return nil, nil
	}

	foods := make([]fdcSearchRespFood, 0, len(out.Foods))
	for _, f := range out.Foods {
		foods = append(foods, fdcSearchRespFood{
			FdcID:       f.FdcID,
			Description: f.Description,
			BrandOwner:  f.BrandOwner,
			DataType:    f.DataType,
		})
	}

	// Prefer branded if brand appears in query
	lq := strings.ToLower(query)
	for _, f := range foods {
		if f.DataType != nil && strings.Contains(*f.DataType, "Branded") && f.BrandOwner != nil {
			brand := strings.ToLower(*f.BrandOwner)
			if brand != "" && strings.Contains(lq, brand) {
				LogDebug("FDC branded match found", "fdcId", f.FdcID, "description", f.Description, "brand", *f.BrandOwner)
				return &f, nil
			}
		}
	}

	LogDebug("FDC match found", "fdcId", foods[0].FdcID, "description", foods[0].Description)
	return &foods[0], nil
}

type fdcFoodDetail struct {
	FoodNutrients []struct {
		Amount   *float64 `json:"amount"`
		Nutrient struct {
			Number         *string `json:"number"`
			Name           *string `json:"name"`
			NutrientNumber *string `json:"nutrientNumber"`
		} `json:"nutrient"`
	} `json:"foodNutrients"`
	LabelNutrients *struct {
		Calories *struct {
			Value *float64 `json:"value"`
		} `json:"calories"`
		Protein *struct {
			Value *float64 `json:"value"`
		} `json:"protein"`
		Fat *struct {
			Value *float64 `json:"value"`
		} `json:"fat"`
		Carbohydrates *struct {
			Value *float64 `json:"value"`
		} `json:"carbohydrates"`
		Fiber *struct {
			Value *float64 `json:"value"`
		} `json:"fiber"`
		Sugars *struct {
			Value *float64 `json:"value"`
		} `json:"sugars"`
	} `json:"labelNutrients"`
}

func fdcGetFoodDetail(ctx context.Context, fdcID int) (*fdcFoodDetail, error) {
	apiKey := os.Getenv("USDA_FDC_API_KEY")
	if apiKey == "" {
		return nil, NewAppError("USDA_FDC_API_KEY not configured", http.StatusInternalServerError, nil)
	}

	LogDebug("Getting FDC food detail", "fdcId", fdcID)

	url := fdcFoodURL + strconv.Itoa(fdcID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, NewAppError("Failed to create FDC detail request", http.StatusInternalServerError, err)
	}
	q := req.URL.Query()
	q.Set("api_key", apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, NewAppError("FDC detail request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		LogError("FDC detail error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body)))
		return nil, NewAppError("FDC food detail failed", http.StatusInternalServerError,
			fmt.Errorf("FDC API error: %d %s", resp.StatusCode, string(body)))
	}
	var out fdcFoodDetail
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, NewAppError("Failed to parse FDC detail response", http.StatusInternalServerError, err)
	}

	LogDebug("FDC food detail retrieved", "fdcId", fdcID, "nutrient_count", len(out.FoodNutrients))
	return &out, nil
}

func extractNutrients(food *fdcFoodDetail) *Nutrients {
	if food == nil {
		return nil
	}
	n := &Nutrients{}
	// Prefer detailed foodNutrients
	if len(food.FoodNutrients) > 0 {
		for _, fn := range food.FoodNutrients {
			if fn.Amount == nil {
				continue
			}
			amt := *fn.Amount
			if !isFinite(amt) {
				continue
			}
			num := ""
			if fn.Nutrient.Number != nil && *fn.Nutrient.Number != "" {
				num = *fn.Nutrient.Number
			} else if fn.Nutrient.NutrientNumber != nil {
				num = *fn.Nutrient.NutrientNumber
			}
			name := ""
			if fn.Nutrient.Name != nil {
				name = strings.ToLower(*fn.Nutrient.Name)
			}

			switch {
			case num == "1008" || strings.Contains(name, "energy"):
				if amt > n.EnergyKcal {
					n.EnergyKcal = amt
				}
			case num == "1003" || strings.Contains(name, "protein"):
				if amt > n.ProteinG {
					n.ProteinG = amt
				}
			case num == "1004" || strings.Contains(name, "total lipid") || strings.Contains(name, "fat"):
				if amt > n.FatG {
					n.FatG = amt
				}
			case num == "1005" || strings.Contains(name, "carbohydrate"):
				if amt > n.CarbsG {
					n.CarbsG = amt
				}
			case num == "1079" || strings.Contains(name, "fiber"):
				if amt > n.FiberG {
					n.FiberG = amt
				}
			case num == "2000" || strings.Contains(name, "sugars"):
				if amt > n.SugarG {
					n.SugarG = amt
				}
			}
		}
		return n
	}
	// Fall back to labelNutrients
	if ln := food.LabelNutrients; ln != nil {
		if ln.Calories != nil && ln.Calories.Value != nil {
			n.EnergyKcal = maxFloat(n.EnergyKcal, *ln.Calories.Value)
		}
		if ln.Protein != nil && ln.Protein.Value != nil {
			n.ProteinG = maxFloat(n.ProteinG, *ln.Protein.Value)
		}
		if ln.Fat != nil && ln.Fat.Value != nil {
			n.FatG = maxFloat(n.FatG, *ln.Fat.Value)
		}
		if ln.Carbohydrates != nil && ln.Carbohydrates.Value != nil {
			n.CarbsG = maxFloat(n.CarbsG, *ln.Carbohydrates.Value)
		}
		if ln.Fiber != nil && ln.Fiber.Value != nil {
			n.FiberG = maxFloat(n.FiberG, *ln.Fiber.Value)
		}
		if ln.Sugars != nil && ln.Sugars.Value != nil {
			n.SugarG = maxFloat(n.SugarG, *ln.Sugars.Value)
		}
		return n
	}
	return nil
}

func resolveItemNutrition(ctx context.Context, item Item) (ItemWithNutrition, error) {
	// Build query for FDC search
	query := item.Name
	if item.Brand != nil && *item.Brand != "" {
		query = *item.Brand + " " + item.Name
	}

	LogDebug("Resolving item nutrition", "item", item.Name, "query", query)

	// Search for best match
	match, err := fdcSearchBestMatch(ctx, query)
	if err != nil {
		LogError("FDC search failed", err, "item", item.Name, "query", query)
		return ItemWithNutrition{Item: item}, err
	}
	if match == nil {
		LogWarn("No FDC match found", "item", item.Name, "query", query)
		return ItemWithNutrition{
			Item: item,
			Note: "No FDC match found",
		}, nil
	}

	LogDebug("FDC match found", "item", item.Name, "fdcId", match.FdcID, "description", match.Description)

	// Get detailed nutrition info
	detail, err := fdcGetFoodDetail(ctx, match.FdcID)
	if err != nil {
		LogError("Failed to get nutrition details", err, "item", item.Name, "fdcId", match.FdcID)
		return ItemWithNutrition{
			Item: item,
			FDC: &FDCRef{
				FDCID:       match.FdcID,
				Description: match.Description,
				BrandOwner:  match.BrandOwner,
				DataType:    match.DataType,
			},
			Note: "Failed to get nutrition details: " + err.Error(),
		}, nil
	}

	nutrients := extractNutrients(detail)
	if nutrients != nil {
		LogDebug("Nutrition extracted", "item", item.Name, "calories", nutrients.EnergyKcal, "protein", nutrients.ProteinG)
	} else {
		LogWarn("No nutrition data extracted", "item", item.Name, "fdcId", match.FdcID)
	}

	return ItemWithNutrition{
		Item: item,
		FDC: &FDCRef{
			FDCID:       match.FdcID,
			Description: match.Description,
			BrandOwner:  match.BrandOwner,
			DataType:    match.DataType,
		},
		Nutri: nutrients,
	}, nil
}

func summarize(items []ItemWithNutrition) Summary {
	totals := Nutrients{}

	// Sum up all nutrients
	for _, item := range items {
		if item.Nutri != nil {
			totals.EnergyKcal += item.Nutri.EnergyKcal
			totals.ProteinG += item.Nutri.ProteinG
			totals.FatG += item.Nutri.FatG
			totals.CarbsG += item.Nutri.CarbsG
			totals.FiberG += item.Nutri.FiberG
			totals.SugarG += item.Nutri.SugarG
		}
	}

	// Daily values for percentage calculation
	dailyValues := map[string]float64{
		"energy_kcal": 2000,
		"protein_g":   50,
		"fat_g":       78,
		"carbs_g":     275,
		"fiber_g":     28,
		"sugar_g":     50,
	}

	// Calculate percentages of daily values
	percentOfDaily := map[string]int{
		"energy.kcal": int((totals.EnergyKcal / dailyValues["energy_kcal"]) * 100),
		"protein.g":   int((totals.ProteinG / dailyValues["protein_g"]) * 100),
		"fat.g":       int((totals.FatG / dailyValues["fat_g"]) * 100),
		"carbs.g":     int((totals.CarbsG / dailyValues["carbs_g"]) * 100),
		"fiber.g":     int((totals.FiberG / dailyValues["fiber_g"]) * 100),
		"sugar.g":     int((totals.SugarG / dailyValues["sugar_g"]) * 100),
	}

	return Summary{
		Totals:          totals,
		PercentOfDaily:  percentOfDaily,
		DailyValuesUsed: dailyValues,
	}
}
