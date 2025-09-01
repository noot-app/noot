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
	Ingredients         []interface{} `json:"ingredients"`
	Link                string        `json:"link"`
	Grade               string        `json:"grade"`
	IsBeverage          *int          `json:"is_beverage"`
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

	// Additional minerals and vitamins per 100g
	VitaminA100g        *FlexFloat `json:"vitamin-a_100g"`        // mcg
	VitaminC100g        *FlexFloat `json:"vitamin-c_100g"`        // mg
	VitaminD100g        *FlexFloat `json:"vitamin-d_100g"`        // mcg
	VitaminE100g        *FlexFloat `json:"vitamin-e_100g"`        // mg
	VitaminK100g        *FlexFloat `json:"vitamin-k_100g"`        // mcg
	Thiamine100g        *FlexFloat `json:"vitamin-b1_100g"`       // mg (thiamine)
	Riboflavin100g      *FlexFloat `json:"vitamin-b2_100g"`       // mg (riboflavin)
	Niacin100g          *FlexFloat `json:"vitamin-pp_100g"`       // mg (niacin/vitamin-pp)
	VitaminB6100g       *FlexFloat `json:"vitamin-b6_100g"`       // mg
	Folate100g          *FlexFloat `json:"vitamin-b9_100g"`       // mcg (folate)
	VitaminB12100g      *FlexFloat `json:"vitamin-b12_100g"`      // mcg
	PantothenicAcid100g *FlexFloat `json:"pantothenic-acid_100g"` // mg
	Calcium100g         *FlexFloat `json:"calcium_100g"`          // mg
	Iron100g            *FlexFloat `json:"iron_100g"`             // mg
	Magnesium100g       *FlexFloat `json:"magnesium_100g"`        // mg
	Phosphorus100g      *FlexFloat `json:"phosphorus_100g"`       // mg
	Potassium100g       *FlexFloat `json:"potassium_100g"`        // mg
	Zinc100g            *FlexFloat `json:"zinc_100g"`             // mg
	Copper100g          *FlexFloat `json:"copper_100g"`           // mg
	Manganese100g       *FlexFloat `json:"manganese_100g"`        // mg
	Selenium100g        *FlexFloat `json:"selenium_100g"`         // mcg
	Iodine100g          *FlexFloat `json:"iodine_100g"`           // mcg

	// ---- OPTIONAL / ADDITIONAL FIELDS (per 100g) ----
	TransFat100g           *FlexFloat `json:"trans-fat_100g"`             // g
	Cholesterol100g        *FlexFloat `json:"cholesterol_100g"`           // mg
	AddedSugars100g        *FlexFloat `json:"added-sugars_100g"`          // g
	Biotin100g             *FlexFloat `json:"biotin_100g"`                // mcg
	Choline100g            *FlexFloat `json:"choline_100g"`               // mg
	Molybdenum100g         *FlexFloat `json:"molybdenum_100g"`            // mcg
	Chromium100g           *FlexFloat `json:"chromium_100g"`              // mcg
	Fluoride100g           *FlexFloat `json:"fluoride_100g"`              // mg
	MonounsaturatedFat100g *FlexFloat `json:"monounsaturated-fat_100g"`   // g
	PolyunsaturatedFat100g *FlexFloat `json:"polyunsaturated-fat_100g"`   // g
	Omega3Ala100g          *FlexFloat `json:"alpha-linolenic-acid_100g"`  // g (ALA)
	Omega3Epa100g          *FlexFloat `json:"eicosapentaenoic-acid_100g"` // g (EPA)
	Omega3Dha100g          *FlexFloat `json:"docosahexaenoic-acid_100g"`  // g (DHA)
	Omega6100g             *FlexFloat `json:"omega-6-fat_100g"`           // g
	Alcohol100g            *FlexFloat `json:"alcohol_100g"`               // g (OFF may also use percent fields)
	Caffeine100g           *FlexFloat `json:"caffeine_100g"`              // mg
	Creatine100g           *FlexFloat `json:"creatine_100g"`              // mg (rare)
	Chloride100g           *FlexFloat `json:"chloride_100g"`              // mg

	// Per serving values (exact per serving, more accurate than scaling)
	EnergyKcalServing      *FlexFloat `json:"energy-kcal_serving"`
	ProteinsServing        *FlexFloat `json:"proteins_serving"`
	FatServing             *FlexFloat `json:"fat_serving"`
	SaturatedFatServing    *FlexFloat `json:"saturated-fat_serving"`
	CarbohydratesServing   *FlexFloat `json:"carbohydrates_serving"`
	SugarsServing          *FlexFloat `json:"sugars_serving"`
	FiberServing           *FlexFloat `json:"fiber_serving"`
	SodiumServing          *FlexFloat `json:"sodium_serving"`
	CalciumServing         *FlexFloat `json:"calcium_serving"`
	VitaminAServing        *FlexFloat `json:"vitamin-a_serving"`
	VitaminCServing        *FlexFloat `json:"vitamin-c_serving"`
	VitaminDServing        *FlexFloat `json:"vitamin-d_serving"`
	VitaminEServing        *FlexFloat `json:"vitamin-e_serving"`
	VitaminKServing        *FlexFloat `json:"vitamin-k_serving"`
	ThiamineServing        *FlexFloat `json:"vitamin-b1_serving"`
	RiboflavinServing      *FlexFloat `json:"vitamin-b2_serving"`
	NiacinServing          *FlexFloat `json:"vitamin-pp_serving"`
	VitaminB6Serving       *FlexFloat `json:"vitamin-b6_serving"`
	FolateServing          *FlexFloat `json:"vitamin-b9_serving"`
	VitaminB12Serving      *FlexFloat `json:"vitamin-b12_serving"`
	PantothenicAcidServing *FlexFloat `json:"pantothenic-acid_serving"`
	IronServing            *FlexFloat `json:"iron_serving"`
	MagnesiumServing       *FlexFloat `json:"magnesium_serving"`
	PhosphorusServing      *FlexFloat `json:"phosphorus_serving"`
	PotassiumServing       *FlexFloat `json:"potassium_serving"`
	ZincServing            *FlexFloat `json:"zinc_serving"`
	CopperServing          *FlexFloat `json:"copper_serving"`
	ManganeseServing       *FlexFloat `json:"manganese_serving"`
	SeleniumServing        *FlexFloat `json:"selenium_serving"`
	IodineServing          *FlexFloat `json:"iodine_serving"`

	// ---- OPTIONAL / ADDITIONAL FIELDS (per serving) ----
	TransFatServing           *FlexFloat `json:"trans-fat_serving"`
	CholesterolServing        *FlexFloat `json:"cholesterol_serving"`
	AddedSugarsServing        *FlexFloat `json:"added-sugars_serving"`
	BiotinServing             *FlexFloat `json:"biotin_serving"`
	CholineServing            *FlexFloat `json:"choline_serving"`
	MolybdenumServing         *FlexFloat `json:"molybdenum_serving"`
	ChromiumServing           *FlexFloat `json:"chromium_serving"`
	FluorideServing           *FlexFloat `json:"fluoride_serving"`
	MonounsaturatedFatServing *FlexFloat `json:"monounsaturated-fat_serving"`
	PolyunsaturatedFatServing *FlexFloat `json:"polyunsaturated-fat_serving"`
	Omega3AlaServing          *FlexFloat `json:"alpha-linolenic-acid_serving"`
	Omega3EpaServing          *FlexFloat `json:"eicosapentaenoic-acid_serving"`
	Omega3DhaServing          *FlexFloat `json:"docosahexaenoic-acid_serving"`
	Omega6Serving             *FlexFloat `json:"omega-6-fat_serving"`
	AlcoholServing            *FlexFloat `json:"alcohol_serving"`
	CaffeineServing           *FlexFloat `json:"caffeine_serving"`
	CreatineServing           *FlexFloat `json:"creatine_serving"`
	ChlorideServing           *FlexFloat `json:"chloride_serving"`
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

// SearchProduct searches for a product by name and brand
func (c *OFFClient) SearchProduct(ctx context.Context, name, brand string) (*OFFProduct, error) {
	if c == nil {
		return nil, fmt.Errorf("OFF client not initialized")
	}

	// For branded items, require both name and brand for better accuracy
	if brand == "" || strings.TrimSpace(brand) == "" {
		LogDebug("Skipping OFF search - no brand specified", "name", name)
		return nil, fmt.Errorf("no brand specified for OFF search: %s", name)
	}

	return c.searchBrandedProduct(ctx, name, brand)
}

// searchBrandedProduct performs the actual OFF API search for branded products
func (c *OFFClient) searchBrandedProduct(ctx context.Context, name, brand string) (*OFFProduct, error) {
	// Build search query - use just the product name for main query
	query := name

	// Search with required fields
	searchURL := fmt.Sprintf("%s/api/v2/search", c.baseURL)
	params := url.Values{}
	params.Set("q", query)
	params.Set("fields", "product_name,brands,nutriments,id,code,serving_quantity,serving_quantity_unit,serving_size,ingredients,link,grade,is_beverage")
	params.Set("page_size", "50") // Limit results

	// Always use brands_tags for branded product searches
	brandTag := c.formatBrandTag(brand)
	params.Set("brands_tags", brandTag)

	fullURL := searchURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OFF request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	// Log request start time for timing analysis
	startTime := time.Now()
	LogDebug("Starting OFF API request", "url", fullURL, "timeout", c.httpClient.Timeout, "user_agent", c.userAgent, "brand_filtered", brand != "", "brand_tag", params.Get("brands_tags"))

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

	LogDebug("Best match analysis complete", "best_score", bestScore, "threshold", 0.5, "best_product", func() string {
		if bestProduct != nil {
			return bestProduct.ProductName
		}
		return "none"
	}())

	// Higher threshold for branded products - require stronger name/brand similarity
	if bestScore < 0.5 {
		LogDebug("Rejecting best match due to low score", "best_score", bestScore, "threshold", 0.5)
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

	// Brand similarity (weighted 30%) - now required since we only search branded items
	brandScore := 0.0
	if searchBrand != "" && product.Brands != "" {
		brandScore = c.calculateSimilarity(strings.ToLower(product.Brands), strings.ToLower(searchBrand))
		score += brandScore * 0.3
	}
	// No fallback score for missing brands since we require brands now

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

	// Check if we have exact serving size match and per-serving nutrition data
	var isExactServing bool

	if product.ServingQuantity != nil && float64(*product.ServingQuantity) == targetGrams {
		isExactServing = true
		LogDebug("Using exact serving size with per-serving nutrition data", "serving_quantity", float64(*product.ServingQuantity), "target_grams", targetGrams)
	}

	// Helper function to get the best nutrition value (per-serving preferred, then scaled from 100g)
	getBestValue := func(servingVal *FlexFloat, per100gVal *FlexFloat) float64 {
		if isExactServing && servingVal != nil {
			// Use exact per-serving value - most accurate
			return float64(*servingVal)
		}
		if per100gVal != nil {
			// Scale from per-100g value
			return float64(*per100gVal) * (targetGrams / 100.0)
		}
		return 0
	}

	return CompleteNutrient{
		// Macronutrients - use per-serving values when available for exact servings
		Calories:     getBestValue(product.Nutriments.EnergyKcalServing, product.Nutriments.EnergyKcal100g),
		Protein:      getBestValue(product.Nutriments.ProteinsServing, product.Nutriments.Proteins100g),
		TotalFat:     getBestValue(product.Nutriments.FatServing, product.Nutriments.Fat100g),
		SaturatedFat: getBestValue(product.Nutriments.SaturatedFatServing, product.Nutriments.SaturatedFat100g),
		TotalCarbs:   getBestValue(product.Nutriments.CarbohydratesServing, product.Nutriments.Carbohydrates100g),
		TotalSugars:  getBestValue(product.Nutriments.SugarsServing, product.Nutriments.Sugars100g),
		DietaryFiber: getBestValue(product.Nutriments.FiberServing, product.Nutriments.Fiber100g),
		Sodium:       getBestValue(product.Nutriments.SodiumServing, product.Nutriments.Sodium100g),

		// Vitamins and minerals - use per-serving values when available, nil if not present
		VitaminA:        getBestValue(product.Nutriments.VitaminAServing, product.Nutriments.VitaminA100g),
		VitaminC:        getBestValue(product.Nutriments.VitaminCServing, product.Nutriments.VitaminC100g),
		VitaminD:        getBestValue(product.Nutriments.VitaminDServing, product.Nutriments.VitaminD100g),
		VitaminE:        getBestValue(product.Nutriments.VitaminEServing, product.Nutriments.VitaminE100g),
		VitaminK:        getBestValue(product.Nutriments.VitaminKServing, product.Nutriments.VitaminK100g),
		Thiamine:        getBestValue(product.Nutriments.ThiamineServing, product.Nutriments.Thiamine100g),
		Riboflavin:      getBestValue(product.Nutriments.RiboflavinServing, product.Nutriments.Riboflavin100g),
		Niacin:          getBestValue(product.Nutriments.NiacinServing, product.Nutriments.Niacin100g),
		VitaminB6:       getBestValue(product.Nutriments.VitaminB6Serving, product.Nutriments.VitaminB6100g),
		Folate:          getBestValue(product.Nutriments.FolateServing, product.Nutriments.Folate100g),
		VitaminB12:      getBestValue(product.Nutriments.VitaminB12Serving, product.Nutriments.VitaminB12100g),
		PantothenicAcid: getBestValue(product.Nutriments.PantothenicAcidServing, product.Nutriments.PantothenicAcid100g),
		Calcium:         getBestValue(product.Nutriments.CalciumServing, product.Nutriments.Calcium100g),
		Iron:            getBestValue(product.Nutriments.IronServing, product.Nutriments.Iron100g),
		Magnesium:       getBestValue(product.Nutriments.MagnesiumServing, product.Nutriments.Magnesium100g),
		Phosphorus:      getBestValue(product.Nutriments.PhosphorusServing, product.Nutriments.Phosphorus100g),
		Potassium:       getBestValue(product.Nutriments.PotassiumServing, product.Nutriments.Potassium100g),
		Zinc:            getBestValue(product.Nutriments.ZincServing, product.Nutriments.Zinc100g),
		Copper:          getBestValue(product.Nutriments.CopperServing, product.Nutriments.Copper100g),
		Manganese:       getBestValue(product.Nutriments.ManganeseServing, product.Nutriments.Manganese100g),
		Selenium:        getBestValue(product.Nutriments.SeleniumServing, product.Nutriments.Selenium100g),
		Iodine:          getBestValue(product.Nutriments.IodineServing, product.Nutriments.Iodine100g),

		// Additional nutrients - use OFF data when available, otherwise zero
		TransFat:    getBestValue(product.Nutriments.TransFatServing, product.Nutriments.TransFat100g),
		Cholesterol: getBestValue(product.Nutriments.CholesterolServing, product.Nutriments.Cholesterol100g),
		AddedSugars: getBestValue(product.Nutriments.AddedSugarsServing, product.Nutriments.AddedSugars100g),
		Biotin:      getBestValue(product.Nutriments.BiotinServing, product.Nutriments.Biotin100g),
		Choline:     getBestValue(product.Nutriments.CholineServing, product.Nutriments.Choline100g),
		Molybdenum:  getBestValue(product.Nutriments.MolybdenumServing, product.Nutriments.Molybdenum100g),
		Chromium:    getBestValue(product.Nutriments.ChromiumServing, product.Nutriments.Chromium100g),
		Fluoride:    getBestValue(product.Nutriments.FluorideServing, product.Nutriments.Fluoride100g),

		// New fatty acids and functional compounds - use OFF data when available
		MonounsaturatedFat: getBestValue(product.Nutriments.MonounsaturatedFatServing, product.Nutriments.MonounsaturatedFat100g),
		PolyunsaturatedFat: getBestValue(product.Nutriments.PolyunsaturatedFatServing, product.Nutriments.PolyunsaturatedFat100g),
		Omega3Ala:          getBestValue(product.Nutriments.Omega3AlaServing, product.Nutriments.Omega3Ala100g),
		Omega3Epa:          getBestValue(product.Nutriments.Omega3EpaServing, product.Nutriments.Omega3Epa100g),
		Omega3Dha:          getBestValue(product.Nutriments.Omega3DhaServing, product.Nutriments.Omega3Dha100g),
		Omega6:             getBestValue(product.Nutriments.Omega6Serving, product.Nutriments.Omega6100g),
		Alcohol:            getBestValue(product.Nutriments.AlcoholServing, product.Nutriments.Alcohol100g),
		Caffeine:           getBestValue(product.Nutriments.CaffeineServing, product.Nutriments.Caffeine100g),
		Creatine:           getBestValue(product.Nutriments.CreatineServing, product.Nutriments.Creatine100g),
		Chloride:           getBestValue(product.Nutriments.ChlorideServing, product.Nutriments.Chloride100g),
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
