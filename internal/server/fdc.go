package server

import (
	"context"
	"encoding/json"
	"errors"
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
		return nil, errors.New("USDA_FDC_API_KEY not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fdcSearchURL, nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FDC search error: %s", string(body))
	}
	var out fdcSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Foods) == 0 {
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
				return &f, nil
			}
		}
	}
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
		return nil, errors.New("USDA_FDC_API_KEY not configured")
	}
	url := fdcFoodURL + strconv.Itoa(fdcID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("api_key", apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FDC food error: %s", string(body))
	}
	var out fdcFoodDetail
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
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
