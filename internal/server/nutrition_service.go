package server

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// Constants for nutrition service configuration
const (
	// Cache TTL for nutrition data
	cacheTTLDays = 30
	cacheTTL     = cacheTTLDays * 24 * time.Hour
)

// getCommonServingSizes returns common serving sizes for cache lookups
func getCommonServingSizes() []float64 {
	return []float64{100, 355, 250, 200, 500, 150, 300, 400, 50, 75, 125}
}

// NutritionService handles nutrition data processing with caching and unit conversions
type NutritionService struct {
	aiProvider AIProvider
	converter  *UnitConverter
	store      storage.Store
}

// scaleNutritionData scales nutrition values by the given factor
func scaleNutritionData(nutrients CompleteNutrient, factor float64) CompleteNutrient {
	var result CompleteNutrient

	// Use reflection to scale all fields with their appropriate precision
	nutrientsVal := reflect.ValueOf(nutrients)
	resultVal := reflect.ValueOf(&result).Elem()

	for i := 0; i < nutrientsVal.NumField(); i++ {
		field := nutrientsVal.Field(i)
		if field.Kind() == reflect.Float64 {
			fieldName := nutrientsVal.Type().Field(i).Name
			scaled := field.Float() * factor

			// Apply precision rounding using the generated map
			precision, exists := NutrientPrecision[fieldName]
			if !exists {
				precision = 2 // default precision
			}

			rounded := convertAndRound(scaled, precision)
			resultVal.Field(i).SetFloat(rounded)
		}
	}

	return result
}

// fetchNutritionFromCache attempts to fetch nutrition data from cache
// Returns nutrition data if found, nil if not found or cache is stale
func (s *NutritionService) fetchNutritionFromCache(ctx context.Context, item Item) (*CompleteNutrient, error) {
	if s.store == nil {
		return nil, nil // No cache available
	}

	// Use brand-aware normalization for cache keys to prevent fragmentation
	normalizedNameForCache, _ := normalizeItemNameWithQuantityForCache(item.Name, item.Brand)
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))
	normalizedName, _ := normalizeItemNameWithQuantity(item.Name)
	normalizedGrams := s.getNormalizedGrams(item)

	LogDebug("Checking cache for item", "original_name", item.Name, "normalized_name_cache", normalizedNameForCache,
		"normalized_name_fallback", normalizedName, "normalized_brand", normalizedBrand, "grams", item.Grams)

	// Try brand-aware cache key first (new approach)
	exactKey := s.makeExactServingKey(normalizedNameForCache, normalizedBrand, normalizedGrams)
	LogDebug("Checking exact serving cache", "exact_key", exactKey)
	if cached, err := s.store.GetItemByName(ctx, exactKey, normalizedBrand); err == nil && cached != nil {
		// Check if cache is still fresh
		if time.Since(cached.UpdatedAt) < cacheTTL {
			LogDebug("Found fresh exact serving cache match", "key", exactKey, "age_days", int(time.Since(cached.UpdatedAt).Hours()/24))

			var nutrition CompleteNutrient
			// If item has BaseQuantity > 1, we need to scale the cached single-unit values
			if item.BaseQuantity != nil && *item.BaseQuantity > 1.0 {
				LogDebug("Using cached exact serving match with scaling for multi-unit quantity (brand-aware cache)",
					"name", item.Name, "base_quantity", *item.BaseQuantity, "total_grams", item.Grams)

				// Get the single-unit nutrition values and scale by BaseQuantity
				singleUnitNutrition := s.convertExactCachedToNutrients(cached)
				scalingFactor := *item.BaseQuantity
				nutrition = scaleNutritionData(singleUnitNutrition, scalingFactor)
			} else {
				LogDebug("Using cached exact serving match - returning original values without scaling",
					"name", item.Name, "grams", item.Grams)
				nutrition = s.convertExactCachedToNutrients(cached)
			}

			return &nutrition, nil
		}
	}

	// Fallback: try traditional cache key for backward compatibility
	fallbackExactKey := s.makeExactServingKey(normalizedName, normalizedBrand, normalizedGrams)
	LogDebug("Checking fallback exact serving cache", "fallback_key", fallbackExactKey)
	if fallbackExactKey != exactKey { // Only check if different from brand-aware key
		if cached, err := s.store.GetItemByName(ctx, fallbackExactKey, normalizedBrand); err == nil && cached != nil {
			// Check if cache is still fresh
			if time.Since(cached.UpdatedAt) < cacheTTL {
				LogDebug("Found fresh fallback exact serving cache match", "key", fallbackExactKey, "age_days", int(time.Since(cached.UpdatedAt).Hours()/24))
				LogDebug("Using cached exact serving match from fallback key - returning original values without scaling (backward compatibility)",
					"name", item.Name, "grams", item.Grams, "fallback_key", fallbackExactKey)

				nutrition := s.convertExactCachedToNutrients(cached)
				return &nutrition, nil
			}
		}
	}

	// Try to find any cached serving size for this item to scale from
	LogDebug("Checking scalable serving cache", "normalized_name", normalizedNameForCache, "normalized_brand", normalizedBrand)
	cachedServings := s.getCachedServingSizes(ctx, normalizedNameForCache, normalizedBrand)
	LogDebug("Found cached servings for scaling", "count", len(cachedServings))
	for _, cachedServing := range cachedServings {
		// Check if cache is still fresh
		if time.Since(cachedServing.item.UpdatedAt) < cacheTTL {
			LogDebug("Found fresh scalable serving cache match", "cached_grams", cachedServing.servingGrams, "age_days", int(time.Since(cachedServing.item.UpdatedAt).Hours()/24))
			// Determine scaling method based on user input and cached data reliability
			if s.shouldUse100gScaling(item, cachedServing.item) {
				LogDebug("Using cached item with per-100g scaling (user provided grams or unreliable base units)",
					"name", item.Name, "requested_grams", item.Grams)

				nutrition := s.convertCachedToNutrients(cachedServing.item, item)
				return &nutrition, nil
			} else {
				LogDebug("Using cached serving data - scaling from cached serving to requested serving",
					"name", item.Name, "cached_grams", cachedServing.servingGrams, "requested_grams", item.Grams)

				nutrition := s.scaleNutritionFromCachedServing(cachedServing.item, cachedServing.servingGrams, item.Grams)
				return &nutrition, nil
			}
		}
	}

	LogDebug("No cache matches found", "name", item.Name, "brand", getBrandOrEmpty(item.Brand))
	return nil, nil // No cache hit
}

// fetchNutritionFromAI fetches nutrition data from AI provider
func (s *NutritionService) fetchNutritionFromAI(ctx context.Context, item *Item, _ interface{}) (*CompleteNutrient, error) {
	LogDebug("Fetching nutrition from AI provider", "name", item.Name)

	// For generic items without brand, use complete AI response to get ingredients
	isGenericItem := (item.Brand == nil || (item.Brand != nil && *item.Brand == ""))
	LogDebug("Checking if item is generic", "item", item.Name, "brand_nil", item.Brand == nil,
		"brand_empty", item.Brand != nil && *item.Brand == "", "is_generic", isGenericItem)

	var nutrition CompleteNutrient
	var err error

	if isGenericItem {
		// No brand - use complete AI response for ingredients
		aiResponse, aiErr := s.aiProvider.GetNutritionWithContextComplete(ctx, *item)
		if aiErr != nil {
			return nil, aiErr
		}
		nutrition = aiResponse.Nutrients

		// Extract ingredients from AI response for generic items
		if len(aiResponse.Ingredients) > 0 {
			item.Ingredients = aiResponse.Ingredients
			LogDebug("Extracted ingredients from AI", "item", item.Name, "ingredient_count", len(item.Ingredients))
		}

		// Extract URL from AI response if available
		if aiResponse.URL != nil && *aiResponse.URL != "" {
			item.Url = aiResponse.URL
			LogDebug("Saved AI URL", "item", item.Name, "url", *aiResponse.URL)
		}

		LogDebug("Using complete AI nutrition", "item", item.Name, "calories", nutrition.Calories)
	} else {
		// Branded items - use standard nutrition only
		nutrition, err = s.aiProvider.GetNutritionWithContext(ctx, *item)
		if err != nil {
			return nil, err
		}
		LogDebug("Using AI nutrition", "item", item.Name, "calories", nutrition.Calories)
	}

	return &nutrition, nil
}

// cacheNutritionData stores nutrition data in the cache with proper scaling
func (s *NutritionService) cacheNutritionData(ctx context.Context, item Item, nutrition CompleteNutrient) error {
	if s.store == nil {
		return nil // No cache available
	}

	LogDebug("Attempting to cache item nutrition data", "name", item.Name, "brand", getBrandOrEmpty(item.Brand), "grams", item.Grams)

	// Extract original quantity info to reverse-scale if needed
	quantityInfo := extractQuantityFromName(item.Name)

	// Calculate the nutrition data for the BASE item (full serving)
	var baseNutrition CompleteNutrient
	var baseGrams float64
	var baseName string

	if quantityInfo.Multiplier != 1.0 {
		// This was a fractional quantity - reverse-scale to get base nutrition
		reverseMultiplier := 1.0 / quantityInfo.Multiplier
		baseGrams = item.Grams * reverseMultiplier
		baseName = quantityInfo.CleanName

		baseNutrition = scaleNutritionData(nutrition, reverseMultiplier)

		LogDebug("Reverse-scaling nutrition for base cache storage", "original_multiplier", quantityInfo.Multiplier,
			"reverse_multiplier", reverseMultiplier, "base_grams", baseGrams, "original_grams", item.Grams)
	} else {
		// This is already a base serving - use as-is
		baseNutrition = nutrition
		baseGrams = item.Grams
		baseName = item.Name
	}

	// Create a base item for caching (using clean name and base grams)
	baseItem := Item{
		Name:        baseName,
		Brand:       item.Brand,
		Grams:       baseGrams,
		Ingredients: item.Ingredients, // Include ingredients from AI
		Url:         item.Url,         // Include AI URL
	}

	// Use brand-aware normalization for consistent cache keys
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))
	baseNormalizedName := normalizeItemNameForCache(baseName, item.Brand)

	// Cache under the base serving size
	exactKey := s.makeExactServingKey(baseNormalizedName, normalizedBrand, baseGrams)
	exactCacheItem := s.convertNutrientsToExactCache(baseItem, baseNutrition, exactKey)

	if exactCached, _ := s.store.GetItemByName(ctx, exactKey, normalizedBrand); exactCached != nil {
		// Update existing exact cache entry
		exactCacheItem.ID = exactCached.ID
		err := s.store.UpdateItem(ctx, exactCacheItem)
		if err != nil {
			LogWarn("Failed to update cached nutrition data", "base_name", baseName,
				"base_grams", baseGrams, "error", err.Error())
			return err
		}
	} else {
		// Create new exact cache entry
		err := s.store.CreateItem(ctx, exactCacheItem)
		if err != nil {
			LogWarn("Failed to cache nutrition data", "base_name", baseName,
				"base_grams", baseGrams, "error", err.Error())
			return err
		}
	}

	LogInfo("Successfully cached nutrition data", "base_name", baseName,
		"base_grams", baseGrams, "cache_key", exactKey)
	return nil
}

// NewNutritionService creates a new nutrition service
func NewNutritionService(store storage.Store) *NutritionService {
	// Create OpenAI provider with config from environment
	config := AIProviderConfig{
		APIKey:          os.Getenv("OPENAI_API_KEY"),
		TranscribeModel: getenv("OPENAI_TRANSCRIBE_MODEL", "gpt-4o-mini-transcribe"),
		BaseURL:         getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		Timeout:         60,
	}

	aiProvider := NewOpenAIProvider(config)
	converter := NewUnitConverter()

	return &NutritionService{
		aiProvider: aiProvider,
		converter:  converter,
		store:      store,
	}
}

// TranscribeAudio transcribes audio to text
func (s *NutritionService) TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error) {
	return s.aiProvider.TranscribeAudio(ctx, filePath, mimeType)
}

// ParseItems extracts food items from transcript
func (s *NutritionService) ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	return s.aiProvider.ParseItems(ctx, transcriptText)
}

// HydrateNutrition hydrates parsed items with nutrition data using cache when available
func (s *NutritionService) HydrateNutrition(ctx context.Context, items []Item) ([]Item, error) {
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
			hydratedItem, err := s.hydrateItemNutrition(ctx, item)
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

// HydrateNutritionWithoutCache hydrates items without using cache (direct AI calls)
func (s *NutritionService) HydrateNutritionWithoutCache(ctx context.Context, items []Item) ([]Item, error) {
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
			// Determine if this is a generic item (no brand)
			isGenericItem := (item.Brand == nil || (item.Brand != nil && *item.Brand == ""))
			LogDebug("Processing item for nutrition", "item", item.Name, "is_generic", isGenericItem)

			var nutrition CompleteNutrient
			var err error

			if isGenericItem {
				// No brand - use complete AI response for ingredients
				aiResponse, aiErr := s.aiProvider.GetNutritionWithContextComplete(ctx, item)
				if aiErr != nil {
					results <- result{index: index, item: item, err: aiErr}
					return
				}
				nutrition = aiResponse.Nutrients

				// Extract ingredients from AI response for generic items
				if len(aiResponse.Ingredients) > 0 {
					item.Ingredients = aiResponse.Ingredients
					LogDebug("Extracted ingredients from AI", "item", item.Name, "ingredient_count", len(item.Ingredients))
				}

				// Extract URL from AI response if available
				if aiResponse.URL != nil && *aiResponse.URL != "" {
					item.Url = aiResponse.URL
					LogDebug("Saved AI URL", "item", item.Name, "url", *aiResponse.URL)
				}

				LogDebug("Using complete AI nutrition", "item", item.Name, "calories", nutrition.Calories)
			} else {
				// Branded items - use standard nutrition only
				nutrition, err = s.aiProvider.GetNutritionWithContext(ctx, item)
				if err != nil {
					results <- result{index: index, item: item, err: err}
					return
				}
				LogDebug("Using AI nutrition", "item", item.Name, "calories", nutrition.Calories)
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

// hydrateItemNutrition hydrates a single item with nutrition data from cache or AI
func (s *NutritionService) hydrateItemNutrition(ctx context.Context, item Item) (Item, error) {
	LogDebug("Starting nutrition hydration for item", "name", item.Name, "brand", getBrandOrEmpty(item.Brand), "grams", item.Grams)

	// Step 1: Try to fetch from cache
	if cachedNutrition, err := s.fetchNutritionFromCache(ctx, item); err != nil {
		LogWarn("Error fetching from cache", "error", err.Error())
	} else if cachedNutrition != nil {
		LogDebug("Using cached nutrition data", "name", item.Name)
		item.Nutrients = cachedNutrition
		return item, nil
	}

	// Step 2: Fetch nutrition from AI
	nutrition, err := s.fetchNutritionFromAI(ctx, &item, nil)
	if err != nil {
		return item, err
	}

	// Step 4: Cache the nutrition data
	if err := s.cacheNutritionData(ctx, item, *nutrition); err != nil {
		LogWarn("Failed to cache nutrition data", "error", err.Error())
		// Don't fail the request if caching fails
	}

	// Step 5: Set nutrition on item and return
	item.Nutrients = nutrition
	return item, nil
}

// convertCachedToNutrients converts cached per-100g data to actual weight
func (s *NutritionService) convertCachedToNutrients(cached *storage.Item, item Item) CompleteNutrient {
	// Use the actual grams from the item
	actualGrams := item.Grams

	var result CompleteNutrient

	// Use reflection to convert all fields dynamically
	cachedVal := reflect.ValueOf(cached).Elem()
	resultVal := reflect.ValueOf(&result).Elem()

	// Map field names for conversion
	fieldMap := map[string]string{
		"Calories":           "CaloriesPer100g",
		"Protein":            "ProteinGPer100g",
		"TotalFat":           "TotalFatGPer100g",
		"SaturatedFat":       "SaturatedFatGPer100g",
		"TransFat":           "TransFatGPer100g",
		"MonounsaturatedFat": "MonounsaturatedFatGPer100g",
		"PolyunsaturatedFat": "PolyunsaturatedFatGPer100g",
		"Cholesterol":        "CholesterolMgPer100g",
		"Sodium":             "SodiumMgPer100g",
		"TotalCarbs":         "TotalCarbsGPer100g",
		"DietaryFiber":       "DietaryFiberGPer100g",
		"TotalSugars":        "TotalSugarsGPer100g",
		"AddedSugars":        "AddedSugarsGPer100g",
		"VitaminA":           "VitaminAMcgPer100g",
		"VitaminC":           "VitaminCMgPer100g",
		"VitaminD":           "VitaminDMcgPer100g",
		"VitaminE":           "VitaminEMgPer100g",
		"VitaminK":           "VitaminKMcgPer100g",
		"Thiamine":           "ThiamineMgPer100g",
		"Riboflavin":         "RiboflavinMgPer100g",
		"Niacin":             "NiacinMgPer100g",
		"VitaminB6":          "VitaminB6MgPer100g",
		"Folate":             "FolateMcgPer100g",
		"VitaminB12":         "VitaminB12McgPer100g",
		"Biotin":             "BiotinMcgPer100g",
		"PantothenicAcid":    "PantothenicAcidMgPer100g",
		"Choline":            "CholineMgPer100g",
		"Calcium":            "CalciumMgPer100g",
		"Iron":               "IronMgPer100g",
		"Magnesium":          "MagnesiumMgPer100g",
		"Phosphorus":         "PhosphorusMgPer100g",
		"Potassium":          "PotassiumMgPer100g",
		"Zinc":               "ZincMgPer100g",
		"Copper":             "CopperMgPer100g",
		"Manganese":          "ManganeseMgPer100g",
		"Selenium":           "SeleniumMcgPer100g",
		"Iodine":             "IodineMcgPer100g",
		"Molybdenum":         "MolybdenumMcgPer100g",
		"Chromium":           "ChromiumMcgPer100g",
		"Fluoride":           "FluorideMgPer100g",
		"Chloride":           "ChlorideMgPer100g",
		"Omega3Ala":          "Omega3AlaGPer100g",
		"Omega3Epa":          "Omega3EpaGPer100g",
		"Omega3Dha":          "Omega3DhaGPer100g",
		"Omega6":             "Omega6GPer100g",
		"Alcohol":            "AlcoholGPer100g",
		"Caffeine":           "CaffeineMgPer100g",
		"Creatine":           "CreatineMgPer100g",
	}

	for resultFieldName, cachedFieldName := range fieldMap {
		cachedField := cachedVal.FieldByName(cachedFieldName)
		resultField := resultVal.FieldByName(resultFieldName)

		if cachedField.IsValid() && resultField.IsValid() && cachedField.Kind() == reflect.Float64 {
			per100gValue := cachedField.Float()
			servingValue := s.converter.ConvertFromPer100gToServing(per100gValue, actualGrams)

			// Apply precision rounding using the generated map
			precision, exists := NutrientPrecision[resultFieldName]
			if !exists {
				precision = 2 // default precision
			}

			rounded := RoundToDecimalPlaces(servingValue, precision)
			resultField.SetFloat(rounded)
		}
	}

	return result
}

// cachedServingData holds info about a cached serving size for scaling
type cachedServingData struct {
	item         *storage.Item
	servingGrams float64
}

// getCachedServingSizes finds all cached serving sizes for an item
func (s *NutritionService) getCachedServingSizes(ctx context.Context, normalizedName, normalizedBrand string) []cachedServingData {
	return s.getCachedServingSizesWithVariations(ctx, normalizedName, normalizedBrand)
}

// getCachedServingSizesWithVariations finds cached serving sizes trying multiple name variations
func (s *NutritionService) getCachedServingSizesWithVariations(ctx context.Context, normalizedName, normalizedBrand string) []cachedServingData {
	var results []cachedServingData

	// Try exact match first (current behavior)
	results = s.tryExactServingMatch(ctx, normalizedName, normalizedBrand)
	if len(results) > 0 {
		return results
	}

	// Phase 2: Try brand-aware variations for better matching
	results = s.tryBrandAwareServingMatch(ctx, normalizedName, normalizedBrand)
	if len(results) > 0 {
		return results
	}

	// Try quantity-normalized variations if exact match failed
	// Extract quantity info from the original normalized name
	quantityInfo := extractQuantityFromName(normalizedName)
	if quantityInfo.Multiplier != 1.0 {
		// Try the cleaned name (without quantity expressions)
		cleanNormalized := normalizeItemName(quantityInfo.CleanName)
		results = s.tryExactServingMatch(ctx, cleanNormalized, normalizedBrand)

		// If we found a match, we need to adjust the serving sizes based on the multiplier
		for i := range results {
			results[i].servingGrams = results[i].servingGrams * quantityInfo.Multiplier
		}
	}

	return results
}

// tryExactServingMatch tries to find cached items using common serving sizes
func (s *NutritionService) tryExactServingMatch(ctx context.Context, normalizedName, normalizedBrand string) []cachedServingData {
	var results []cachedServingData

	// Try some common serving sizes to see if we have them cached
	commonGrams := getCommonServingSizes()

	for _, grams := range commonGrams {
		exactKey := s.makeExactServingKey(normalizedName, normalizedBrand, grams)
		if cached, err := s.store.GetItemByName(ctx, exactKey, normalizedBrand); err == nil && cached != nil {
			results = append(results, cachedServingData{
				item:         cached,
				servingGrams: grams,
			})
			// Return the first match to keep it simple
			break
		}
	}

	return results
}

// tryBrandAwareServingMatch tries to find cached items using brand-aware name variations
func (s *NutritionService) tryBrandAwareServingMatch(ctx context.Context, normalizedName, normalizedBrand string) []cachedServingData {
	var results []cachedServingData

	// Only do brand-aware matching if we have a brand
	if normalizedBrand == "" {
		return results
	}

	// Convert normalizedBrand back to original form for variation generation
	var brandPtr *string = &normalizedBrand

	// Generate brand-aware variations using the LLM-parsed brand
	// The normalizedName is the item name and normalizedBrand is the LLM-extracted brand
	variations := generateBrandAwareVariations(normalizedName, brandPtr)

	// Try each variation with common serving sizes
	commonGrams := getCommonServingSizes()

	for _, variation := range variations {
		// Skip the original exact match since we already tried that
		if variation == normalizedName {
			continue
		}

		for _, grams := range commonGrams {
			// Maintain brand safety - only match within the same brand
			exactKey := s.makeExactServingKey(variation, normalizedBrand, grams)
			if cached, err := s.store.GetItemByName(ctx, exactKey, normalizedBrand); err == nil && cached != nil {
				LogDebug("Found brand-aware cache match", "variation", variation,
					"original_name", normalizedName, "brand", normalizedBrand, "grams", grams)

				results = append(results, cachedServingData{
					item:         cached,
					servingGrams: grams,
				})
				// Return first match to keep performance good
				return results
			}
		}
	}

	return results
}

// scaleNutritionFromCachedServing scales nutrition data from one serving size to another
func (s *NutritionService) scaleNutritionFromCachedServing(cachedItem *storage.Item, fromGrams, toGrams float64) CompleteNutrient {
	// Extract the original serving values from the exact cache and scale them
	scalingFactor := toGrams / fromGrams

	var result CompleteNutrient

	// Use reflection to scale all fields dynamically
	cachedVal := reflect.ValueOf(cachedItem).Elem()
	resultVal := reflect.ValueOf(&result).Elem()

	// Map field names for conversion from original cached values
	fieldMap := map[string]string{
		"Calories":           "OriginalCalories",
		"Protein":            "OriginalProteinG",
		"TotalFat":           "OriginalTotalFatG",
		"SaturatedFat":       "OriginalSaturatedFatG",
		"TransFat":           "OriginalTransFatG",
		"MonounsaturatedFat": "OriginalMonounsaturatedFatG",
		"PolyunsaturatedFat": "OriginalPolyunsaturatedFatG",
		"Cholesterol":        "OriginalCholesterolMg",
		"Sodium":             "OriginalSodiumMg",
		"TotalCarbs":         "OriginalTotalCarbsG",
		"DietaryFiber":       "OriginalDietaryFiberG",
		"TotalSugars":        "OriginalTotalSugarsG",
		"AddedSugars":        "OriginalAddedSugarsG",
		"VitaminA":           "OriginalVitaminAMcg",
		"VitaminC":           "OriginalVitaminCMg",
		"VitaminD":           "OriginalVitaminDMcg",
		"VitaminE":           "OriginalVitaminEMg",
		"VitaminK":           "OriginalVitaminKMcg",
		"Thiamine":           "OriginalThiamineMg",
		"Riboflavin":         "OriginalRiboflavinMg",
		"Niacin":             "OriginalNiacinMg",
		"VitaminB6":          "OriginalVitaminB6Mg",
		"Folate":             "OriginalFolateMcg",
		"VitaminB12":         "OriginalVitaminB12Mcg",
		"Biotin":             "OriginalBiotinMcg",
		"PantothenicAcid":    "OriginalPantothenicAcidMg",
		"Choline":            "OriginalCholineMg",
		"Calcium":            "OriginalCalciumMg",
		"Iron":               "OriginalIronMg",
		"Magnesium":          "OriginalMagnesiumMg",
		"Phosphorus":         "OriginalPhosphorusMg",
		"Potassium":          "OriginalPotassiumMg",
		"Zinc":               "OriginalZincMg",
		"Copper":             "OriginalCopperMg",
		"Manganese":          "OriginalManganeseMg",
		"Selenium":           "OriginalSeleniumMcg",
		"Iodine":             "OriginalIodineMcg",
		"Molybdenum":         "OriginalMolybdenumMcg",
		"Chromium":           "OriginalChromiumMcg",
		"Fluoride":           "OriginalFluorideMg",
		"Chloride":           "OriginalChlorideMg",
		"Omega3Ala":          "OriginalOmega3AlaG",
		"Omega3Epa":          "OriginalOmega3EpaG",
		"Omega3Dha":          "OriginalOmega3DhaG",
		"Omega6":             "OriginalOmega6G",
		"Alcohol":            "OriginalAlcoholG",
		"Caffeine":           "OriginalCaffeineMg",
		"Creatine":           "OriginalCreatineMg",
	}

	for resultFieldName, cachedFieldName := range fieldMap {
		cachedField := cachedVal.FieldByName(cachedFieldName)
		resultField := resultVal.FieldByName(resultFieldName)

		if cachedField.IsValid() && resultField.IsValid() {
			// Handle pointer fields (most original fields are pointers)
			var scaledValue float64
			if cachedField.Kind() == reflect.Ptr && !cachedField.IsNil() {
				scaledValue = cachedField.Elem().Float() * scalingFactor
			} else if cachedField.Kind() == reflect.Float64 {
				scaledValue = cachedField.Float() * scalingFactor
			}

			// Special handling for calories (round up)
			if resultFieldName == "Calories" {
				scaledValue = float64(RoundCaloriesUp(scaledValue))
			}

			resultField.SetFloat(scaledValue)
		}
	}

	return result
}
func (s *NutritionService) makeExactServingKey(normalizedName, normalizedBrand string, grams float64) string {
	return fmt.Sprintf("%s|%s|%.1fg", normalizedName, normalizedBrand, grams)
}

// getNormalizedGrams returns the normalized grams for cache key generation
func (s *NutritionService) getNormalizedGrams(item Item) float64 {
	if item.BaseQuantity != nil && *item.BaseQuantity > 1.0 {
		// Normalize to single unit by dividing by base quantity
		return item.Grams / *item.BaseQuantity
	}
	return item.Grams
}

// shouldUse100gScaling determines if we should use per-100g scaling instead of base unit scaling
func (s *NutritionService) shouldUse100gScaling(item Item, cachedItem *storage.Item) bool {
	// If user provided consumption in grams (no logical units), prefer 100g scaling
	if item.UserUnit == nil || *item.UserUnit == "g" || *item.UserUnit == "grams" || *item.UserUnit == "gram" {
		return true
	}

	// If cached item doesn't have reliable original serving data, use 100g scaling
	if cachedItem.OriginalServingGrams == nil || *cachedItem.OriginalServingGrams <= 0 {
		return true
	}

	// Otherwise, use base unit scaling (more accurate for logical units like "cans", "slices", etc.)
	return false
}

// convertExactCachedToNutrients converts cached data from exact serving match (uses original values)
func (s *NutritionService) convertExactCachedToNutrients(cached *storage.Item) CompleteNutrient {
	// If we have original serving data, use it directly (no conversion needed)
	if cached.OriginalServingGrams != nil {
		var result CompleteNutrient

		// Use reflection to convert all fields dynamically
		cachedVal := reflect.ValueOf(cached).Elem()
		resultVal := reflect.ValueOf(&result).Elem()

		// Map field names for conversion from original cached values
		fieldMap := map[string]string{
			"Calories":           "OriginalCalories",
			"Protein":            "OriginalProteinG",
			"TotalFat":           "OriginalTotalFatG",
			"SaturatedFat":       "OriginalSaturatedFatG",
			"TransFat":           "OriginalTransFatG",
			"MonounsaturatedFat": "OriginalMonounsaturatedFatG",
			"PolyunsaturatedFat": "OriginalPolyunsaturatedFatG",
			"Cholesterol":        "OriginalCholesterolMg",
			"Sodium":             "OriginalSodiumMg",
			"TotalCarbs":         "OriginalTotalCarbsG",
			"DietaryFiber":       "OriginalDietaryFiberG",
			"TotalSugars":        "OriginalTotalSugarsG",
			"AddedSugars":        "OriginalAddedSugarsG",
			"VitaminA":           "OriginalVitaminAMcg",
			"VitaminC":           "OriginalVitaminCMg",
			"VitaminD":           "OriginalVitaminDMcg",
			"VitaminE":           "OriginalVitaminEMg",
			"VitaminK":           "OriginalVitaminKMcg",
			"Thiamine":           "OriginalThiamineMg",
			"Riboflavin":         "OriginalRiboflavinMg",
			"Niacin":             "OriginalNiacinMg",
			"VitaminB6":          "OriginalVitaminB6Mg",
			"Folate":             "OriginalFolateMcg",
			"VitaminB12":         "OriginalVitaminB12Mcg",
			"Biotin":             "OriginalBiotinMcg",
			"PantothenicAcid":    "OriginalPantothenicAcidMg",
			"Choline":            "OriginalCholineMg",
			"Calcium":            "OriginalCalciumMg",
			"Iron":               "OriginalIronMg",
			"Magnesium":          "OriginalMagnesiumMg",
			"Phosphorus":         "OriginalPhosphorusMg",
			"Potassium":          "OriginalPotassiumMg",
			"Zinc":               "OriginalZincMg",
			"Copper":             "OriginalCopperMg",
			"Manganese":          "OriginalManganeseMg",
			"Selenium":           "OriginalSeleniumMcg",
			"Iodine":             "OriginalIodineMcg",
			"Molybdenum":         "OriginalMolybdenumMcg",
			"Chromium":           "OriginalChromiumMcg",
			"Fluoride":           "OriginalFluorideMg",
			"Chloride":           "OriginalChlorideMg",
			"Omega3Ala":          "OriginalOmega3AlaG",
			"Omega3Epa":          "OriginalOmega3EpaG",
			"Omega3Dha":          "OriginalOmega3DhaG",
			"Omega6":             "OriginalOmega6G",
			"Alcohol":            "OriginalAlcoholG",
			"Caffeine":           "OriginalCaffeineMg",
			"Creatine":           "OriginalCreatineMg",
		}

		for resultFieldName, cachedFieldName := range fieldMap {
			cachedField := cachedVal.FieldByName(cachedFieldName)
			resultField := resultVal.FieldByName(resultFieldName)

			if cachedField.IsValid() && resultField.IsValid() {
				// Handle pointer fields (original fields are pointers)
				if cachedField.Kind() == reflect.Ptr && !cachedField.IsNil() {
					resultField.SetFloat(cachedField.Elem().Float())
				} else if cachedField.Kind() == reflect.Float64 {
					resultField.SetFloat(cachedField.Float())
				}
			}
		}

		return result
	}

	// Fallback to per-100g data if original data is not available (shouldn't happen with new schema)
	LogWarn("Original serving data not found, falling back to per-100g conversion",
		"normalized_name", cached.NormalizedName)

	// Use the generated function for per-100g conversion
	return ConvertCachedToNutrientsGenerated(cached)
}

// convertNutrientsToExactCache stores nutrition data with hybrid approach
func (s *NutritionService) convertNutrientsToExactCache(item Item, nutrients CompleteNutrient, exactKey string) *storage.Item {
	actualGrams := item.Grams

	// Normalize nutrition values to single unit if BaseQuantity is set
	normalizedNutrients := nutrients
	normalizedGrams := actualGrams

	if item.BaseQuantity != nil && *item.BaseQuantity > 1.0 {
		// Normalize to single unit by dividing by base quantity
		divider := *item.BaseQuantity
		normalizedGrams = actualGrams / divider

		normalizedNutrients = scaleNutritionData(nutrients, 1.0/divider)
	}

	// Helper function to convert and round in one step for per-100g values
	convertAndRound := func(servingValue float64, fieldName string) float64 {
		precision, exists := NutrientPrecision[fieldName]
		if !exists {
			precision = 2 // default precision
		}
		return RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(servingValue, normalizedGrams), precision)
	}

	// Helper function to create pointer to float64
	float64Ptr := func(val float64) *float64 { return &val }

	return &storage.Item{
		NormalizedName:  exactKey, // Use the special key as the normalized name
		NormalizedBrand: normalizeItemName(getBrandOrEmpty(item.Brand)),
		DisplayName:     item.Name,
		DisplayBrand:    getBrandOrEmpty(item.Brand),

		// Store normalized single-unit serving data
		OriginalServingGrams:        float64Ptr(normalizedGrams),
		OriginalCalories:            float64Ptr(normalizedNutrients.Calories),
		OriginalProteinG:            float64Ptr(normalizedNutrients.Protein),
		OriginalTotalFatG:           float64Ptr(normalizedNutrients.TotalFat),
		OriginalSaturatedFatG:       float64Ptr(normalizedNutrients.SaturatedFat),
		OriginalTransFatG:           float64Ptr(normalizedNutrients.TransFat),
		OriginalCholesterolMg:       float64Ptr(normalizedNutrients.Cholesterol),
		OriginalSodiumMg:            float64Ptr(normalizedNutrients.Sodium),
		OriginalTotalCarbsG:         float64Ptr(normalizedNutrients.TotalCarbs),
		OriginalDietaryFiberG:       float64Ptr(normalizedNutrients.DietaryFiber),
		OriginalTotalSugarsG:        float64Ptr(normalizedNutrients.TotalSugars),
		OriginalAddedSugarsG:        float64Ptr(normalizedNutrients.AddedSugars),
		OriginalVitaminAMcg:         float64Ptr(normalizedNutrients.VitaminA),
		OriginalVitaminCMg:          float64Ptr(normalizedNutrients.VitaminC),
		OriginalVitaminDMcg:         float64Ptr(normalizedNutrients.VitaminD),
		OriginalVitaminEMg:          float64Ptr(normalizedNutrients.VitaminE),
		OriginalVitaminKMcg:         float64Ptr(normalizedNutrients.VitaminK),
		OriginalThiamineMg:          float64Ptr(normalizedNutrients.Thiamine),
		OriginalRiboflavinMg:        float64Ptr(normalizedNutrients.Riboflavin),
		OriginalNiacinMg:            float64Ptr(normalizedNutrients.Niacin),
		OriginalVitaminB6Mg:         float64Ptr(normalizedNutrients.VitaminB6),
		OriginalFolateMcg:           float64Ptr(normalizedNutrients.Folate),
		OriginalVitaminB12Mcg:       float64Ptr(normalizedNutrients.VitaminB12),
		OriginalBiotinMcg:           float64Ptr(normalizedNutrients.Biotin),
		OriginalPantothenicAcidMg:   float64Ptr(normalizedNutrients.PantothenicAcid),
		OriginalCholineMg:           float64Ptr(normalizedNutrients.Choline),
		OriginalCalciumMg:           float64Ptr(normalizedNutrients.Calcium),
		OriginalIronMg:              float64Ptr(normalizedNutrients.Iron),
		OriginalMagnesiumMg:         float64Ptr(normalizedNutrients.Magnesium),
		OriginalPhosphorusMg:        float64Ptr(normalizedNutrients.Phosphorus),
		OriginalPotassiumMg:         float64Ptr(normalizedNutrients.Potassium),
		OriginalZincMg:              float64Ptr(normalizedNutrients.Zinc),
		OriginalCopperMg:            float64Ptr(normalizedNutrients.Copper),
		OriginalManganeseMg:         float64Ptr(normalizedNutrients.Manganese),
		OriginalSeleniumMcg:         float64Ptr(normalizedNutrients.Selenium),
		OriginalIodineMcg:           float64Ptr(normalizedNutrients.Iodine),
		OriginalMolybdenumMcg:       float64Ptr(normalizedNutrients.Molybdenum),
		OriginalChromiumMcg:         float64Ptr(normalizedNutrients.Chromium),
		OriginalFluorideMg:          float64Ptr(normalizedNutrients.Fluoride),
		OriginalChlorideMg:          float64Ptr(normalizedNutrients.Chloride),
		OriginalOmega3AlaG:          float64Ptr(normalizedNutrients.Omega3Ala),
		OriginalOmega3EpaG:          float64Ptr(normalizedNutrients.Omega3Epa),
		OriginalOmega3DhaG:          float64Ptr(normalizedNutrients.Omega3Dha),
		OriginalOmega6G:             float64Ptr(normalizedNutrients.Omega6),
		OriginalCreatineMg:          float64Ptr(normalizedNutrients.Creatine),
		OriginalCaffeineMg:          float64Ptr(normalizedNutrients.Caffeine),
		OriginalAlcoholG:            float64Ptr(normalizedNutrients.Alcohol),
		OriginalPolyunsaturatedFatG: float64Ptr(normalizedNutrients.PolyunsaturatedFat),
		OriginalMonounsaturatedFatG: float64Ptr(normalizedNutrients.MonounsaturatedFat),

		// Also store per-100g data for scaling to other serving sizes
		CaloriesPer100g:            convertAndRound(normalizedNutrients.Calories, "Calories"),
		ProteinGPer100g:            convertAndRound(normalizedNutrients.Protein, "Protein"),
		TotalFatGPer100g:           convertAndRound(normalizedNutrients.TotalFat, "TotalFat"),
		SaturatedFatGPer100g:       convertAndRound(normalizedNutrients.SaturatedFat, "SaturatedFat"),
		TransFatGPer100g:           convertAndRound(normalizedNutrients.TransFat, "TransFat"),
		CholesterolMgPer100g:       convertAndRound(normalizedNutrients.Cholesterol, "Cholesterol"),
		SodiumMgPer100g:            convertAndRound(normalizedNutrients.Sodium, "Sodium"),
		TotalCarbsGPer100g:         convertAndRound(normalizedNutrients.TotalCarbs, "TotalCarbs"),
		DietaryFiberGPer100g:       convertAndRound(normalizedNutrients.DietaryFiber, "DietaryFiber"),
		TotalSugarsGPer100g:        convertAndRound(normalizedNutrients.TotalSugars, "TotalSugars"),
		AddedSugarsGPer100g:        convertAndRound(normalizedNutrients.AddedSugars, "AddedSugars"),
		VitaminAMcgPer100g:         convertAndRound(normalizedNutrients.VitaminA, "VitaminA"),
		VitaminCMgPer100g:          convertAndRound(normalizedNutrients.VitaminC, "VitaminC"),
		VitaminDMcgPer100g:         convertAndRound(normalizedNutrients.VitaminD, "VitaminD"),
		VitaminEMgPer100g:          convertAndRound(normalizedNutrients.VitaminE, "VitaminE"),
		VitaminKMcgPer100g:         convertAndRound(normalizedNutrients.VitaminK, "VitaminK"),
		ThiamineMgPer100g:          convertAndRound(normalizedNutrients.Thiamine, "Thiamine"),
		RiboflavinMgPer100g:        convertAndRound(normalizedNutrients.Riboflavin, "Riboflavin"),
		NiacinMgPer100g:            convertAndRound(normalizedNutrients.Niacin, "Niacin"),
		VitaminB6MgPer100g:         convertAndRound(normalizedNutrients.VitaminB6, "VitaminB6"),
		FolateMcgPer100g:           convertAndRound(normalizedNutrients.Folate, "Folate"),
		VitaminB12McgPer100g:       convertAndRound(normalizedNutrients.VitaminB12, "VitaminB12"),
		BiotinMcgPer100g:           convertAndRound(normalizedNutrients.Biotin, "Biotin"),
		PantothenicAcidMgPer100g:   convertAndRound(normalizedNutrients.PantothenicAcid, "PantothenicAcid"),
		CholineMgPer100g:           convertAndRound(normalizedNutrients.Choline, "Choline"),
		CalciumMgPer100g:           convertAndRound(normalizedNutrients.Calcium, "Calcium"),
		IronMgPer100g:              convertAndRound(normalizedNutrients.Iron, "Iron"),
		MagnesiumMgPer100g:         convertAndRound(normalizedNutrients.Magnesium, "Magnesium"),
		PhosphorusMgPer100g:        convertAndRound(normalizedNutrients.Phosphorus, "Phosphorus"),
		PotassiumMgPer100g:         convertAndRound(normalizedNutrients.Potassium, "Potassium"),
		ZincMgPer100g:              convertAndRound(normalizedNutrients.Zinc, "Zinc"),
		CopperMgPer100g:            convertAndRound(normalizedNutrients.Copper, "Copper"),
		ManganeseMgPer100g:         convertAndRound(normalizedNutrients.Manganese, "Manganese"),
		SeleniumMcgPer100g:         convertAndRound(normalizedNutrients.Selenium, "Selenium"),
		IodineMcgPer100g:           convertAndRound(normalizedNutrients.Iodine, "Iodine"),
		MolybdenumMcgPer100g:       convertAndRound(normalizedNutrients.Molybdenum, "Molybdenum"),
		ChromiumMcgPer100g:         convertAndRound(normalizedNutrients.Chromium, "Chromium"),
		FluorideMgPer100g:          convertAndRound(normalizedNutrients.Fluoride, "Fluoride"),
		ChlorideMgPer100g:          convertAndRound(normalizedNutrients.Chloride, "Chloride"),
		Omega3AlaGPer100g:          convertAndRound(normalizedNutrients.Omega3Ala, "Omega3Ala"),
		Omega3EpaGPer100g:          convertAndRound(normalizedNutrients.Omega3Epa, "Omega3Epa"),
		Omega3DhaGPer100g:          convertAndRound(normalizedNutrients.Omega3Dha, "Omega3Dha"),
		Omega6GPer100g:             convertAndRound(normalizedNutrients.Omega6, "Omega6"),
		CreatineMgPer100g:          convertAndRound(normalizedNutrients.Creatine, "Creatine"),
		CaffeineMgPer100g:          convertAndRound(normalizedNutrients.Caffeine, "Caffeine"),
		AlcoholGPer100g:            convertAndRound(normalizedNutrients.Alcohol, "Alcohol"),
		PolyunsaturatedFatGPer100g: convertAndRound(normalizedNutrients.PolyunsaturatedFat, "PolyunsaturatedFat"),
		MonounsaturatedFatGPer100g: convertAndRound(normalizedNutrients.MonounsaturatedFat, "MonounsaturatedFat"),
		Note:                       item.Note,

		// Include ingredients and AI URL when caching items
		Ingredients: item.Ingredients, // Copy ingredients from the item
		Url:         item.Url,         // Copy AI URL from the item

		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// floatValue safely dereferences a float64 pointer, returning 0 if nil
func floatValue(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
