package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OFFClient handles Open Food Facts API integration
type OFFClient struct {
	baseURL    string
	userAgent  string
	httpClient *http.Client
	timeout    time.Duration
}

// OFFProduct represents a product from Open Food Facts
type OFFProduct struct {
	ProductName string                 `json:"product_name"`
	Brands      string                 `json:"brands"`
	Nutriments  OFFNutriments         `json:"nutriments"`
	ID          string                 `json:"id"`
	Code        string                 `json:"code"`
	CompletedT  float64               `json:"completed_t"` // Completeness score (0-1)
}

// OFFNutriments represents nutrition data from OFF
type OFFNutriments struct {
	// Energy and macros per 100g
	EnergyKcal100g      *float64 `json:"energy-kcal_100g"`
	Proteins100g        *float64 `json:"proteins_100g"`
	Fat100g             *float64 `json:"fat_100g"`
	SaturatedFat100g    *float64 `json:"saturated-fat_100g"`
	Carbohydrates100g   *float64 `json:"carbohydrates_100g"`
	Sugars100g          *float64 `json:"sugars_100g"`
	Fiber100g           *float64 `json:"fiber_100g"`
	Sodium100g          *float64 `json:"sodium_100g"`
	
	// Vitamins per 100g (in mg unless specified)
	VitaminA100g        *float64 `json:"vitamin-a_100g"`     // mcg
	VitaminC100g        *float64 `json:"vitamin-c_100g"`     // mg  
	VitaminD100g        *float64 `json:"vitamin-d_100g"`     // mcg
	Calcium100g         *float64 `json:"calcium_100g"`       // mg
	Iron100g            *float64 `json:"iron_100g"`          // mg
}

// OFFSearchResponse represents the API response structure
type OFFSearchResponse struct {
	Products []OFFProduct `json:"products"`
	Count    int          `json:"count"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// OFFClientConfig holds configuration for OFF client
type OFFClientConfig struct {
	BaseURL   string
	UserAgent string
	Timeout   time.Duration
	Enabled   bool
}

// NewOFFClient creates a new Open Food Facts client
func NewOFFClient(config OFFClientConfig) *OFFClient {
	if !config.Enabled {
		return nil
	}
	
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &OFFClient{
		baseURL:   config.BaseURL,
		userAgent: config.UserAgent,
		timeout:   timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// SearchProduct searches for a product by name and optional brand
func (c *OFFClient) SearchProduct(ctx context.Context, name, brand string) (*OFFProduct, error) {
	if c == nil {
		return nil, fmt.Errorf("OFF client not initialized")
	}

	// Build search query
	query := name
	if brand != "" {
		query = fmt.Sprintf("%s %s", brand, name)
	}

	// Search with required fields
	searchURL := fmt.Sprintf("%s/api/v2/search", c.baseURL)
	params := url.Values{}
	params.Set("q", query)
	params.Set("fields", "product_name,brands,nutriments,id,code,completed_t")
	params.Set("page_size", "5") // Limit results
	
	fullURL := searchURL + "?" + params.Encode()

	LogDebug("Querying OFF API", "url", fullURL)

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OFF request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OFF API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OFF API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OFF response: %w", err)
	}

	var searchResp OFFSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse OFF response: %w", err)
	}

	LogDebug("OFF search results", "query", query, "count", len(searchResp.Products))

	// Find the best match
	bestProduct := c.findBestMatch(searchResp.Products, name, brand)
	if bestProduct == nil {
		return nil, fmt.Errorf("no suitable product found in OFF")
	}

	LogDebug("Selected OFF product", "name", bestProduct.ProductName, "brands", bestProduct.Brands, "completeness", bestProduct.CompletedT)

	return bestProduct, nil
}

// findBestMatch finds the best product match based on name and brand similarity
func (c *OFFClient) findBestMatch(products []OFFProduct, name, brand string) *OFFProduct {
	if len(products) == 0 {
		return nil
	}

	var bestProduct *OFFProduct
	var bestScore float64

	for i := range products {
		product := &products[i]
		
		// Skip products with very low completeness scores
		if product.CompletedT < 0.5 {
			continue
		}

		// Calculate match score based on name similarity and brand match
		score := c.calculateMatchScore(product, name, brand)
		
		if score > bestScore {
			bestScore = score
			bestProduct = product
		}
	}

	// Only return if we have a decent confidence score
	if bestScore < 0.6 {
		return nil
	}

	return bestProduct
}

// calculateMatchScore calculates how well a product matches the search criteria
func (c *OFFClient) calculateMatchScore(product *OFFProduct, searchName, searchBrand string) float64 {
	score := 0.0
	
	// Name similarity (weighted 60%)
	nameScore := c.calculateSimilarity(strings.ToLower(product.ProductName), strings.ToLower(searchName))
	score += nameScore * 0.6
	
	// Brand similarity (weighted 30%) 
	if searchBrand != "" && product.Brands != "" {
		brandScore := c.calculateSimilarity(strings.ToLower(product.Brands), strings.ToLower(searchBrand))
		score += brandScore * 0.3
	} else if searchBrand == "" {
		// No brand penalty if user didn't specify brand
		score += 0.3
	}
	
	// Completeness bonus (weighted 10%)
	score += product.CompletedT * 0.1
	
	return score
}

// calculateSimilarity calculates basic string similarity (simple containment check)
func (c *OFFClient) calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		return 0.8
	}
	
	// Check for common words
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)
	commonWords := 0
	
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 && len(w1) > 2 { // Skip short words
				commonWords++
				break
			}
		}
	}
	
	if len(words1) > 0 && len(words2) > 0 {
		return float64(commonWords) / float64(max(len(words1), len(words2)))
	}
	
	return 0.0
}

// ConvertToCompleteNutrient converts OFF nutrition data to our CompleteNutrient format
func (c *OFFClient) ConvertToCompleteNutrient(product *OFFProduct, targetGrams float64) CompleteNutrient {
	if product == nil {
		return CompleteNutrient{}
	}

	// OFF data is per 100g, so we need to scale to target grams
	scaleFactor := targetGrams / 100.0
	
	// Helper function to safely convert and scale values
	scaleValue := func(val *float64) float64 {
		if val == nil {
			return 0
		}
		return *val * scaleFactor
	}

	return CompleteNutrient{
		// Macronutrients
		Calories:     scaleValue(product.Nutriments.EnergyKcal100g),
		Protein:      scaleValue(product.Nutriments.Proteins100g),
		TotalFat:     scaleValue(product.Nutriments.Fat100g),
		SaturatedFat: scaleValue(product.Nutriments.SaturatedFat100g),
		TotalCarbs:   scaleValue(product.Nutriments.Carbohydrates100g),
		TotalSugars:  scaleValue(product.Nutriments.Sugars100g),
		DietaryFiber: scaleValue(product.Nutriments.Fiber100g),
		Sodium:       scaleValue(product.Nutriments.Sodium100g),
		
		// Vitamins and minerals (what's available from OFF)
		VitaminA:     scaleValue(product.Nutriments.VitaminA100g),
		VitaminC:     scaleValue(product.Nutriments.VitaminC100g),
		VitaminD:     scaleValue(product.Nutriments.VitaminD100g),
		Calcium:      scaleValue(product.Nutriments.Calcium100g),
		Iron:         scaleValue(product.Nutriments.Iron100g),
		
		// Zero out fields not typically available in OFF
		// These could be filled by LLM in a future enhancement
		TransFat:        0,
		Cholesterol:     0,
		AddedSugars:     0,
		VitaminE:        0,
		VitaminK:        0,
		Thiamine:        0,
		Riboflavin:      0,
		Niacin:          0,
		VitaminB6:       0,
		Folate:          0,
		VitaminB12:      0,
		Biotin:          0,
		PantothenicAcid: 0,
		Choline:         0,
		Magnesium:       0,
		Phosphorus:      0,
		Potassium:       0,
		Zinc:            0,
		Copper:          0,
		Manganese:       0,
		Selenium:        0,
		Iodine:          0,
		Molybdenum:      0,
		Chromium:        0,
		Fluoride:        0,
		Chloride:        0,
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}