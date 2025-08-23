package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	ProductName         string        `json:"product_name"`
	Brands              string        `json:"brands"`
	Nutriments          OFFNutriments `json:"nutriments"`
	ID                  string        `json:"id"`
	Code                string        `json:"code"`
	ServingQuantity     *FlexFloat    `json:"serving_quantity"`
	ServingQuantityUnit string        `json:"serving_quantity_unit"`
	ServingSize         string        `json:"serving_size"`
}

// OFFNutriments represents nutrition data from OFF
type OFFNutriments struct {
	// Energy and macros per 100g
	EnergyKcal100g    *FlexFloat `json:"energy-kcal_100g"`
	Proteins100g      *FlexFloat `json:"proteins_100g"`
	Fat100g           *FlexFloat `json:"fat_100g"`
	SaturatedFat100g  *FlexFloat `json:"saturated-fat_100g"`
	Carbohydrates100g *FlexFloat `json:"carbohydrates_100g"`
	Sugars100g        *FlexFloat `json:"sugars_100g"`
	Fiber100g         *FlexFloat `json:"fiber_100g"`
	Sodium100g        *FlexFloat `json:"sodium_100g"`

	// Vitamins per 100g (in mg unless specified)
	VitaminA100g *FlexFloat `json:"vitamin-a_100g"` // mcg
	VitaminC100g *FlexFloat `json:"vitamin-c_100g"` // mg
	VitaminD100g *FlexFloat `json:"vitamin-d_100g"` // mcg
	Calcium100g  *FlexFloat `json:"calcium_100g"`   // mg
	Iron100g     *FlexFloat `json:"iron_100g"`      // mg
}

// FlexFloat handles JSON values that can be either string or number
type FlexFloat float64

// UnmarshalJSON implements custom unmarshaling for FlexFloat
func (ff *FlexFloat) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as number first
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*ff = FlexFloat(f)
		return nil
	}

	// Try to unmarshal as string and convert to float
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Handle empty strings
	if s == "" {
		*ff = FlexFloat(0)
		return nil
	}

	// Convert string to float
	parsed, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("cannot parse '%s' as float: %w", s, err)
	}

	*ff = FlexFloat(parsed)
	return nil
}

// Float64 returns the float64 value, or nil if zero
func (ff *FlexFloat) Float64() *float64 {
	if ff == nil {
		return nil
	}
	f := float64(*ff)
	if f == 0 {
		return nil
	}
	return &f
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

// minInt returns the smaller of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// formatBrandTag formats a brand name according to OFF brands_tags conventions
func (c *OFFClient) formatBrandTag(brand string) string {
	// Start with lowercase
	tag := strings.ToLower(brand)

	// Handle apostrophes first (both straight and curly)
	tag = strings.ReplaceAll(tag, "'", "") // Straight apostrophe
	tag = strings.ReplaceAll(tag, "'", "") // Curly apostrophe

	// Replace common special characters according to OFF conventions
	replacements := map[string]string{
		"&":  "", // Ben & Jerry's → ben-jerrys (& gets dropped, spaces become -)
		"®":  "", // Remove registered trademark
		"™":  "", // Remove trademark
		"©":  "", // Remove copyright
		".":  "", // Remove dots
		",":  "", // Remove commas
		"!":  "", // Remove exclamation marks
		"?":  "", // Remove question marks
		"(":  "", // Remove parentheses
		")":  "",
		"[":  "", // Remove brackets
		"]":  "",
		"/":  "-", // Forward slash to hyphen
		"\\": "-", // Backslash to hyphen
		":":  "",  // Remove colons
		";":  "",  // Remove semicolons
		"é":  "e", // Nestlé → nestle
		"è":  "e",
		"ê":  "e",
		"ë":  "e",
		"à":  "a",
		"á":  "a",
		"â":  "a",
		"ä":  "a",
		"ñ":  "n",
		"ü":  "u",
		"ö":  "o",
		"ß":  "ss",
	}

	// Apply character replacements
	for old, new := range replacements {
		tag = strings.ReplaceAll(tag, old, new)
	}

	// Replace spaces with hyphens
	tag = strings.ReplaceAll(tag, " ", "-")

	// Clean up multiple consecutive hyphens
	for strings.Contains(tag, "--") {
		tag = strings.ReplaceAll(tag, "--", "-")
	}

	// Trim leading/trailing hyphens
	tag = strings.Trim(tag, "-")

	return tag
}

// SearchProduct searches for a product by name and optional brand
func (c *OFFClient) SearchProduct(ctx context.Context, name, brand string) (*OFFProduct, error) {
	if c == nil {
		return nil, fmt.Errorf("OFF client not initialized")
	}

	// Build search query - use just the product name for main query
	query := name

	// Search with required fields
	searchURL := fmt.Sprintf("%s/api/v2/search", c.baseURL)
	params := url.Values{}
	params.Set("q", query)
	params.Set("fields", "product_name,brands,nutriments,id,code,serving_quantity,serving_quantity_unit,serving_size")
	params.Set("page_size", "50") // Limit results

	// Use brands_tags for more precise brand filtering when available
	if brand != "" {
		brandTag := c.formatBrandTag(brand)
		params.Set("brands_tags", brandTag)
	}

	fullURL := searchURL + "?" + params.Encode()

	if brand != "" {
		LogDebug("Querying OFF API with brand filtering", "url", fullURL, "timeout", c.httpClient.Timeout, "brand_tag", params.Get("brands_tags"))
	} else {
		LogDebug("Querying OFF API", "url", fullURL, "timeout", c.httpClient.Timeout)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OFF request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	// Log request start time for timing analysis
	startTime := time.Now()
	LogDebug("OFF API request starting", "url", fullURL, "user_agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		LogDebug("OFF API request failed", "url", fullURL, "duration", duration, "error", err.Error())
		return nil, fmt.Errorf("OFF API request failed: %w", err)
	}
	defer resp.Body.Close()

	LogDebug("OFF API response received", "url", fullURL, "duration", duration, "status", resp.StatusCode, "content_length", resp.ContentLength)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		LogDebug("OFF API non-200 response", "status", resp.StatusCode, "body", string(body), "duration", duration)
		return nil, fmt.Errorf("OFF API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LogDebug("Failed to read OFF response body", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to read OFF response: %w", err)
	}

	LogDebug("OFF response body read", "body_size", len(body), "duration", duration)

	var searchResp OFFSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		LogDebug("Failed to parse OFF JSON response", "error", err.Error(), "body_preview", string(body[:minInt(500, len(body))]))
		return nil, fmt.Errorf("failed to parse OFF response: %w", err)
	}

	LogDebug("OFF search results parsed", "query", query, "count", len(searchResp.Products), "total_duration", duration)

	// Log the products we found for debugging
	for i, product := range searchResp.Products {
		LogDebug("OFF product candidate", "index", i, "name", product.ProductName, "brands", product.Brands)
	}

	// Find the best match
	bestProduct := c.findBestMatch(searchResp.Products, name, brand)
	if bestProduct == nil {
		return nil, fmt.Errorf("no suitable product found in OFF")
	}

	LogDebug("Selected OFF product", "name", bestProduct.ProductName, "brands", bestProduct.Brands)

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

		// Calculate match score based purely on name and brand similarity
		score := c.calculateMatchScore(product, name, brand)
		LogDebug("Product match score calculated", "name", product.ProductName, "brands", product.Brands, "score", score)

		if score > bestScore {
			bestScore = score
			bestProduct = product
		}
	}

	LogDebug("Best match analysis complete", "best_score", bestScore, "threshold", 0.3, "best_product", func() string {
		if bestProduct != nil {
			return bestProduct.ProductName
		}
		return "none"
	}())

	// Only return if we have a decent confidence score (lowered to 0.3 for name/brand focus)
	if bestScore < 0.3 {
		LogDebug("Rejecting best match due to low score", "best_score", bestScore, "threshold", 0.3)
		return nil
	}

	return bestProduct
}

// calculateMatchScore calculates how well a product matches the search criteria
func (c *OFFClient) calculateMatchScore(product *OFFProduct, searchName, searchBrand string) float64 {
	score := 0.0

	// Name similarity (weighted 70%)
	nameScore := c.calculateSimilarity(strings.ToLower(product.ProductName), strings.ToLower(searchName))
	score += nameScore * 0.7

	// Brand similarity (weighted 30%)
	brandScore := 0.0
	if searchBrand != "" && product.Brands != "" {
		brandScore = c.calculateSimilarity(strings.ToLower(product.Brands), strings.ToLower(searchBrand))
		score += brandScore * 0.3
	} else if searchBrand == "" {
		// No brand penalty if user didn't specify brand
		brandScore = 1.0 // For logging purposes
		score += 0.3
	}

	LogDebug("Match score details",
		"product", product.ProductName,
		"search_name", searchName,
		"search_brand", searchBrand,
		"name_score", nameScore,
		"brand_score", brandScore,
		"total_score", score)

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

	var similarity float64
	if len(words1) > 0 && len(words2) > 0 {
		similarity = float64(commonWords) / float64(max(len(words1), len(words2)))
	} else {
		similarity = 0.0
	}

	return similarity
} // ConvertToCompleteNutrient converts OFF nutrition data to our CompleteNutrient format
func (c *OFFClient) ConvertToCompleteNutrient(product *OFFProduct, targetGrams float64) CompleteNutrient {
	if product == nil {
		return CompleteNutrient{}
	}

	// OFF data is per 100g, so we need to scale to target grams
	scaleFactor := targetGrams / 100.0

	// Helper function to safely convert and scale FlexFloat values
	scaleFlexValue := func(val *FlexFloat) float64 {
		if val == nil {
			return 0
		}
		return float64(*val) * scaleFactor
	}

	return CompleteNutrient{
		// Macronutrients
		Calories:     scaleFlexValue(product.Nutriments.EnergyKcal100g),
		Protein:      scaleFlexValue(product.Nutriments.Proteins100g),
		TotalFat:     scaleFlexValue(product.Nutriments.Fat100g),
		SaturatedFat: scaleFlexValue(product.Nutriments.SaturatedFat100g),
		TotalCarbs:   scaleFlexValue(product.Nutriments.Carbohydrates100g),
		TotalSugars:  scaleFlexValue(product.Nutriments.Sugars100g),
		DietaryFiber: scaleFlexValue(product.Nutriments.Fiber100g),
		Sodium:       scaleFlexValue(product.Nutriments.Sodium100g),

		// Vitamins and minerals (what's available from OFF)
		VitaminA: scaleFlexValue(product.Nutriments.VitaminA100g),
		VitaminC: scaleFlexValue(product.Nutriments.VitaminC100g),
		VitaminD: scaleFlexValue(product.Nutriments.VitaminD100g),
		Calcium:  scaleFlexValue(product.Nutriments.Calcium100g),
		Iron:     scaleFlexValue(product.Nutriments.Iron100g),

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
