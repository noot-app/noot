package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

const (
	openAIBaseURL       = "https://api.openai.com/v1"
	openAITranscribeURL = openAIBaseURL + "/audio/transcriptions"
	openAIChatURL       = openAIBaseURL + "/chat/completions"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

func transcriptionPrompt() string {
	return `The audio is a short dictation of foods and drinks consumed. Preserve exact brand and product names (e.g., "Clover Organic", "Trader Joe's", "Siggi's", "Icelandic skyr", "LaCroix"), coffee drink terms (espresso, latte, macchiato), tea terms (matcha), and ingredient names (goji berries, blueberries, Greek yogurt, European style yogurt). Keep numbers and units (cups, grams, ounces, tbsp) and include standard punctuation. Do not add or infer items that were not spoken. If an item is given without a quantity, assume it is one standard serving size of that item which would make logical sense in the context of the consumption. For example, if a user says "I had a banana", assume it is one banana, not a bunch. If they say "I had some eggs", assume it is two eggs, not a dozen. If they say "I had some yogurt", assume it is one standard serving size of yogurt, not a gallon. If they say "I had some coffee", assume it is one standard cup of coffee, not a pot. If the user says "I had a latte" assume it contains two shots of espresso.`
}

func parseItemsSystemPrompt() string {
	return `You extract individual food and drink items from a freeform meal, snack, beverage, or consumption description. Return strict JSON with the following structure:

{
  "items": [
    {
      "name": string,
      "quantity": number | null,
      "unit": string | null,
      "brand": string | null
    }
  ]
}

IMPORTANT INSTRUCTIONS:
1. EXTRACT INDIVIDUAL ITEMS: Separate each distinct food or drink item mentioned.
2. INFER SERVING SIZES: If quantity is not specified, assume reasonable standard serving sizes based on context:
   - Yogurt: 1 cup (245g)
   - Banana: 1 medium (118g)  
   - Eggs: 2 large eggs (100g)
   - Coffee: 1 cup (240ml)
   - Latte: 10oz with 2 shots espresso using standard 20g shots
   - Apple: 1 medium (182g)
   - Bread slice: 1 slice (28g)
   - Chicken breast: 3.5oz (100g)
   - Rice: 1 cup cooked (158g)

3. PRESERVE BRANDS: Keep exact brand and product names (e.g., "Clover Sonoma", "Trader Joe's", "Siggi's", "KFC", "Starbucks").
4. NORMALIZE UNITS: Use standard units (g, mg, ml, cups, tbsp, etc.).
5. DO NOT ADD NUTRITION DATA: Only extract item identification, not nutrition information.`
}

func nutritionSystemPrompt() string {
	return `You provide complete nutrition information for a single food item. Return strict JSON with the following structure:

{
  "nutrients": {
    "calories": number,
    "protein_g": number,
    "total_fat_g": number,
    "saturated_fat_g": number,
    "trans_fat_g": number,
    "cholesterol_mg": number,
    "sodium_mg": number,
    "total_carbs_g": number,
    "dietary_fiber_g": number,
    "total_sugars_g": number,
    "added_sugars_g": number,
    "vitamin_a_mcg": number,
    "vitamin_c_mg": number,
    "vitamin_d_mcg": number,
    "vitamin_e_mg": number,
    "vitamin_k_mcg": number,
    "thiamine_mg": number,
    "riboflavin_mg": number,
    "niacin_mg": number,
    "vitamin_b6_mg": number,
    "folate_mcg": number,
    "vitamin_b12_mcg": number,
    "calcium_mg": number,
    "iron_mg": number,
    "magnesium_mg": number,
    "phosphorus_mg": number,
    "potassium_mg": number,
    "zinc_mg": number,
    "copper_mg": number,
    "manganese_mg": number,
    "selenium_mcg": number
  }
}

IMPORTANT INSTRUCTIONS:
1. NUTRITION DATA ACCURACY: Provide accurate nutrition data per serving for the specified quantity and unit. Use your knowledge of food composition databases, USDA data, and nutrition labels.
2. HANDLE COMPLEX ITEMS: For prepared foods, estimate based on typical recipes and ingredients. For restaurant items, use available nutrition information or estimate based on similar items.
3. ZERO VALUES: Use 0 for nutrients that are truly absent (like vitamin B12 in plants), but provide realistic non-zero values for nutrients that are typically present even in small amounts.
4. BRANDED VS GENERIC: Prioritize branded nutrition data when brand is specified, otherwise use generic USDA-style data for the food type.
5. QUANTITY-ADJUSTED: Provide nutrition values for the exact quantity/unit specified, not per 100g.`
}

func transcribeAudio(ctx context.Context, filePath, _ string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", NewAppError("OPENAI_API_KEY not configured", http.StatusInternalServerError, nil)
	}
	model := getenv("OPENAI_TRANSCRIBE_MODEL", "gpt-4o-mini-transcribe")

	LogDebug("Starting OpenAI transcription", "model", model, "file", filePath)

	prompt := transcriptionPrompt()

	lang := strings.TrimSpace(os.Getenv("TRANSCRIBE_LANGUAGE")) // e.g., "en"
	respFormat := strings.TrimSpace(os.Getenv("OPENAI_TRANSCRIBE_RESPONSE_FORMAT"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("model", model)
	if prompt != "" {
		_ = writer.WriteField("prompt", prompt)
	}
	if lang != "" {
		_ = writer.WriteField("language", lang)
	}
	if respFormat != "" {
		_ = writer.WriteField("response_format", respFormat)
	}

	fw, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", NewAppError("Failed to create form file", http.StatusInternalServerError, err)
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", NewAppError("Failed to open audio file", http.StatusInternalServerError, err)
	}
	defer f.Close()
	if _, err = io.Copy(fw, f); err != nil {
		return "", NewAppError("Failed to copy audio file", http.StatusInternalServerError, err)
	}
	if err := writer.Close(); err != nil {
		return "", NewAppError("Failed to close form writer", http.StatusInternalServerError, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAITranscribeURL, body)
	if err != nil {
		return "", NewAppError("Failed to create transcription request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	LogDebug("Sending transcription request to OpenAI")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", NewAppError("OpenAI transcription request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		LogError("OpenAI transcription error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(b)))
		return "", NewAppError("OpenAI transcription failed", http.StatusInternalServerError,
			fmt.Errorf("OpenAI API error: %d %s", resp.StatusCode, string(b)))
	}

	if strings.EqualFold(respFormat, "text") {
		b, _ := io.ReadAll(resp.Body)
		result := string(b)
		LogDebug("Transcription completed", "length", len(result))
		return result, nil
	}
	var out struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", NewAppError("Failed to parse transcription response", http.StatusInternalServerError, err)
	}

	LogDebug("Transcription completed", "length", len(out.Text))
	return out.Text, nil
}

func parseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return ParsedItems{}, NewAppError("OPENAI_API_KEY not configured", http.StatusInternalServerError, nil)
	}
	model := getenv("OPENAI_PARSE_MODEL", "gpt-4o-mini")

	LogDebug("Starting OpenAI item parsing (items only)", "model", model, "transcript_length", len(transcriptText))

	system := parseItemsSystemPrompt()
	user := "Meal: " + transcriptText

	payload := map[string]any{
		"model":       model,
		"temperature": 0.0, // deterministic for item parsing
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIChatURL, bytes.NewReader(b))
	if err != nil {
		return ParsedItems{}, NewAppError("Failed to create parsing request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	LogDebug("Sending item parsing request to OpenAI")
	resp, err := httpClient.Do(req)
	if err != nil {
		return ParsedItems{}, NewAppError("OpenAI parsing request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI parsing error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body)))
		return ParsedItems{}, NewAppError("OpenAI parsing failed", http.StatusInternalServerError,
			fmt.Errorf("OpenAI API error: %d %s", resp.StatusCode, string(body)))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ParsedItems{}, NewAppError("Failed to parse OpenAI response", http.StatusInternalServerError, err)
	}
	content := ""
	if len(out.Choices) > 0 {
		content = out.Choices[0].Message.Content
	}

	LogDebug("OpenAI item parsing response received", "content_length", len(content))
	LogDebug("OpenAI response content", "content", content)

	var parsed ParsedItems
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		LogWarn("Failed to parse OpenAI JSON response", "content", content, "error", err.Error())
		parsed = ParsedItems{Items: []Item{}}
	}

	// Normalize items (no nutrition data at this stage)
	clean := make([]Item, 0, len(parsed.Items))
	for _, i := range parsed.Items {
		name := strings.TrimSpace(i.Name)
		if name == "" {
			continue
		}
		clean = append(clean, Item{
			Name:      name,
			Quantity:  i.Quantity,
			Unit:      strPtrOrNil(i.Unit),
			Brand:     strPtrOrNil(i.Brand),
			Nutrients: nil, // No nutrition data in phase 1
		})
	}

	LogDebug("Item parsing completed", "items_found", len(clean))
	return ParsedItems{Items: clean}, nil
}

// hydrateNutrition takes parsed items and hydrates them with nutrition data
// Uses cache when available, parallel OpenAI calls when not cached
func hydrateNutrition(ctx context.Context, items []Item, store storage.Store) ([]Item, error) {
	LogDebug("Starting nutrition hydration", "item_count", len(items))

	if len(items) == 0 {
		return items, nil
	}

	// Use goroutines and channels for parallel processing
	type result struct {
		index int
		item  Item
		err   error
	}

	results := make(chan result, len(items))

	// Start goroutines for each item
	for i, item := range items {
		go func(index int, item Item) {
			hydratedItem, err := hydrateItemNutrition(ctx, item, store)
			results <- result{index: index, item: hydratedItem, err: err}
		}(i, item)
	}

	// Collect results
	hydratedItems := make([]Item, len(items))
	for i := 0; i < len(items); i++ {
		res := <-results
		if res.err != nil {
			LogWarn("Failed to hydrate item nutrition, using item without nutrition",
				"index", res.index, "name", items[res.index].Name, "error", res.err.Error())
			// Use original item without nutrition data rather than failing entire request
			hydratedItems[res.index] = items[res.index]
		} else {
			hydratedItems[res.index] = res.item
		}
	}

	LogDebug("Nutrition hydration completed", "hydrated_count", len(hydratedItems))
	return hydratedItems, nil
}

// hydrateNutritionWithoutCache hydrates items without using cache (direct OpenAI calls)
func hydrateNutritionWithoutCache(ctx context.Context, items []Item) ([]Item, error) {
	LogDebug("Starting nutrition hydration without cache", "item_count", len(items))

	if len(items) == 0 {
		return items, nil
	}

	// Use goroutines and channels for parallel processing
	type result struct {
		index int
		item  Item
		err   error
	}

	results := make(chan result, len(items))

	// Start goroutines for each item
	for i, item := range items {
		go func(index int, item Item) {
			nutrition, err := getNutritionFromOpenAI(ctx, item)
			if err != nil {
				results <- result{index: index, item: item, err: err}
				return
			}
			item.Nutrients = &nutrition
			results <- result{index: index, item: item, err: nil}
		}(i, item)
	}

	// Collect results
	hydratedItems := make([]Item, len(items))
	for i := 0; i < len(items); i++ {
		res := <-results
		if res.err != nil {
			LogWarn("Failed to get item nutrition, using item without nutrition",
				"index", res.index, "name", items[res.index].Name, "error", res.err.Error())
			// Use original item without nutrition data rather than failing entire request
			hydratedItems[res.index] = items[res.index]
		} else {
			hydratedItems[res.index] = res.item
		}
	}

	LogDebug("Nutrition hydration completed", "hydrated_count", len(hydratedItems))
	return hydratedItems, nil
}

// hydrateItemNutrition hydrates a single item with nutrition data from cache or OpenAI
func hydrateItemNutrition(ctx context.Context, item Item, store storage.Store) (Item, error) {
	normalizedName := normalizeItemName(item.Name)
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))

	LogDebug("Checking cache for item", "normalized_name", normalizedName, "normalized_brand", normalizedBrand)

	// Check cache first
	cached, err := store.GetItemFromCache(ctx, normalizedName, normalizedBrand)
	if err == nil && cached != nil {
		if !store.IsItemCacheExpired(cached) {
			LogDebug("Using cached nutrition data", "name", item.Name)
			nutrition := convertCachedToNutrients(cached, item.Quantity, item.Unit)
			item.Nutrients = &nutrition
			return item, nil
		}
		LogDebug("Cache expired for item", "name", item.Name, "expires_at", cached.ExpiresAt)
	}

	// Not in cache or expired - get from OpenAI
	LogDebug("Fetching nutrition from OpenAI", "name", item.Name)
	nutrition, err := getNutritionFromOpenAI(ctx, item)
	if err != nil {
		return item, err
	}

	// Cache the result
	cacheItem := convertNutrientsToCache(item, nutrition)
	if cached != nil {
		// Refresh existing cache entry
		err = store.RefreshItemCache(ctx, normalizedName, normalizedBrand, cacheItem)
	} else {
		// Create new cache entry
		err = store.UpsertItemCache(ctx, cacheItem)
	}
	if err != nil {
		LogWarn("Failed to cache nutrition data", "name", item.Name, "error", err.Error())
		// Don't fail the request if caching fails
	}

	item.Nutrients = &nutrition
	return item, nil
}

// getNutritionFromOpenAI gets nutrition data for a single item from OpenAI
func getNutritionFromOpenAI(ctx context.Context, item Item) (CompleteNutrient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return CompleteNutrient{}, NewAppError("OPENAI_API_KEY not configured", http.StatusInternalServerError, nil)
	}
	model := getenv("OPENAI_PARSE_MODEL", "gpt-4o-mini")

	system := nutritionSystemPrompt()

	// Build user message with item details
	var userMsg strings.Builder
	userMsg.WriteString("Item: ")
	userMsg.WriteString(item.Name)

	if item.Quantity != nil {
		userMsg.WriteString(fmt.Sprintf("\nQuantity: %v", *item.Quantity))
	}
	if item.Unit != nil {
		userMsg.WriteString(fmt.Sprintf("\nUnit: %s", *item.Unit))
	}
	if item.Brand != nil {
		userMsg.WriteString(fmt.Sprintf("\nBrand: %s", *item.Brand))
	}

	payload := map[string]any{
		"model":       model,
		"temperature": 0.1, // slightly higher for nutrition estimates
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": userMsg.String()},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIChatURL, bytes.NewReader(b))
	if err != nil {
		return CompleteNutrient{}, NewAppError("Failed to create nutrition request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return CompleteNutrient{}, NewAppError("OpenAI nutrition request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI nutrition error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body)))
		return CompleteNutrient{}, NewAppError("OpenAI nutrition failed", http.StatusInternalServerError,
			fmt.Errorf("OpenAI API error: %d %s", resp.StatusCode, string(body)))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return CompleteNutrient{}, NewAppError("Failed to parse nutrition response", http.StatusInternalServerError, err)
	}

	content := ""
	if len(out.Choices) > 0 {
		content = out.Choices[0].Message.Content
	}

	var result struct {
		Nutrients CompleteNutrient `json:"nutrients"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		LogWarn("Failed to parse nutrition JSON", "content", content, "error", err.Error())
		return CompleteNutrient{}, NewAppError("Failed to parse nutrition data", http.StatusInternalServerError, err)
	}

	return result.Nutrients, nil
}

// Helper functions for cache conversion and normalization

func normalizeItemName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func getBrandOrEmpty(brand *string) string {
	if brand == nil {
		return ""
	}
	return *brand
}

func convertCachedToNutrients(cached *storage.ItemCache, quantity *float64, unit *string) CompleteNutrient {
	// Create unit converter
	converter := NewUnitConverter()

	// Create a temporary item to estimate serving weight
	item := Item{
		Name:     cached.DisplayName,
		Quantity: quantity,
		Unit:     unit,
	}

	// Estimate serving weight for scaling
	servingGrams, err := converter.EstimateServingWeight(item)
	if err != nil {
		servingGrams = 100.0 // Default to 100g
	}

	// Convert from per-100g cache data to actual serving size
	return CompleteNutrient{
		Calories:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.CaloriesPer100g, servingGrams), 1),
		Protein:      RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.ProteinGPer100g, servingGrams), 1),
		TotalFat:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.TotalFatGPer100g, servingGrams), 1),
		SaturatedFat: RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.SaturatedFatGPer100g, servingGrams), 1),
		TransFat:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.TransFatGPer100g, servingGrams), 1),
		Cholesterol:  RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.CholesterolMgPer100g, servingGrams), 1),
		Sodium:       RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.SodiumMgPer100g, servingGrams), 1),
		TotalCarbs:   RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.TotalCarbsGPer100g, servingGrams), 1),
		DietaryFiber: RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.DietaryFiberGPer100g, servingGrams), 1),
		TotalSugars:  RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.TotalSugarsGPer100g, servingGrams), 1),
		AddedSugars:  RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.AddedSugarsGPer100g, servingGrams), 1),
		VitaminA:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.VitaminAMcgPer100g, servingGrams), 1),
		VitaminC:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.VitaminCMgPer100g, servingGrams), 1),
		VitaminD:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.VitaminDMcgPer100g, servingGrams), 1),
		VitaminE:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.VitaminEMgPer100g, servingGrams), 1),
		VitaminK:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.VitaminKMcgPer100g, servingGrams), 1),
		Thiamine:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.ThiamineMgPer100g, servingGrams), 3),
		Riboflavin:   RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.RiboflavinMgPer100g, servingGrams), 3),
		Niacin:       RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.NiacinMgPer100g, servingGrams), 1),
		VitaminB6:    RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.VitaminB6MgPer100g, servingGrams), 3),
		Folate:       RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.FolateMcgPer100g, servingGrams), 1),
		VitaminB12:   RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.VitaminB12McgPer100g, servingGrams), 2),
		Calcium:      RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.CalciumMgPer100g, servingGrams), 1),
		Iron:         RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.IronMgPer100g, servingGrams), 1),
		Magnesium:    RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.MagnesiumMgPer100g, servingGrams), 1),
		Phosphorus:   RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.PhosphorusMgPer100g, servingGrams), 1),
		Potassium:    RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.PotassiumMgPer100g, servingGrams), 1),
		Zinc:         RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.ZincMgPer100g, servingGrams), 2),
		Copper:       RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.CopperMgPer100g, servingGrams), 3),
		Manganese:    RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.ManganeseMgPer100g, servingGrams), 3),
		Selenium:     RoundToDecimalPlaces(converter.ConvertFromPer100gToServing(cached.SeleniumMcgPer100g, servingGrams), 1),
	}
}

func convertNutrientsToCache(item Item, nutrients CompleteNutrient) *storage.ItemCache {
	normalizedName := normalizeItemName(item.Name)
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))

	// Create unit converter and estimate serving weight
	converter := NewUnitConverter()

	// Since the LLM provided nutrition for the exact serving size,
	// we need to estimate what that serving weighs to convert to per-100g for caching
	servingGrams, err := converter.EstimateServingWeight(item)
	if err != nil {
		// If we can't estimate the weight, default to 100g
		servingGrams = 100.0
	}

	// Convert from serving size to per-100g for cache storage
	return &storage.ItemCache{
		NormalizedName:       normalizedName,
		NormalizedBrand:      normalizedBrand,
		DisplayName:          item.Name,
		DisplayBrand:         getBrandOrEmpty(item.Brand),
		CaloriesPer100g:      RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Calories, servingGrams), 1),
		ProteinGPer100g:      RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Protein, servingGrams), 1),
		TotalFatGPer100g:     RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.TotalFat, servingGrams), 1),
		SaturatedFatGPer100g: RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.SaturatedFat, servingGrams), 1),
		TransFatGPer100g:     RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.TransFat, servingGrams), 1),
		CholesterolMgPer100g: RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Cholesterol, servingGrams), 1),
		SodiumMgPer100g:      RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Sodium, servingGrams), 1),
		TotalCarbsGPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.TotalCarbs, servingGrams), 1),
		DietaryFiberGPer100g: RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.DietaryFiber, servingGrams), 1),
		TotalSugarsGPer100g:  RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.TotalSugars, servingGrams), 1),
		AddedSugarsGPer100g:  RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.AddedSugars, servingGrams), 1),
		VitaminAMcgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.VitaminA, servingGrams), 1),
		VitaminCMgPer100g:    RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.VitaminC, servingGrams), 1),
		VitaminDMcgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.VitaminD, servingGrams), 1),
		VitaminEMgPer100g:    RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.VitaminE, servingGrams), 1),
		VitaminKMcgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.VitaminK, servingGrams), 1),
		ThiamineMgPer100g:    RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Thiamine, servingGrams), 3),
		RiboflavinMgPer100g:  RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Riboflavin, servingGrams), 3),
		NiacinMgPer100g:      RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Niacin, servingGrams), 1),
		VitaminB6MgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.VitaminB6, servingGrams), 3),
		FolateMcgPer100g:     RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Folate, servingGrams), 1),
		VitaminB12McgPer100g: RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.VitaminB12, servingGrams), 2),
		CalciumMgPer100g:     RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Calcium, servingGrams), 1),
		IronMgPer100g:        RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Iron, servingGrams), 1),
		MagnesiumMgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Magnesium, servingGrams), 1),
		PhosphorusMgPer100g:  RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Phosphorus, servingGrams), 1),
		PotassiumMgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Potassium, servingGrams), 1),
		ZincMgPer100g:        RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Zinc, servingGrams), 2),
		CopperMgPer100g:      RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Copper, servingGrams), 3),
		ManganeseMgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Manganese, servingGrams), 3),
		SeleniumMcgPer100g:   RoundToDecimalPlaces(converter.ConvertFromServingToPer100g(nutrients.Selenium, servingGrams), 1),
		FetchedAt:            time.Now().UTC(),
		ExpiresAt:            time.Now().UTC().Add(30 * 24 * time.Hour), // 30 day TTL
	}
}

func strPtrOrNil(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
