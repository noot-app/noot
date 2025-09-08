package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

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
	result := nutrients
	result.Scale(factor)
	return result
}

// buildCacheKeys generates cache keys for the item
func (s *NutritionService) buildCacheKeys(item Item) (exactKey, fallbackKey, normalizedName, normalizedBrand string, normalizedGrams float64) {
	// Use brand-aware normalization for cache keys to prevent fragmentation
	normalizedNameForCache, _ := normalizeItemNameWithQuantityForCache(item.Name, item.Brand)
	normalizedBrand = normalizeItemName(getBrandOrEmpty(item.Brand))
	fallbackNormalizedName, _ := normalizeItemNameWithQuantity(item.Name)
	normalizedGrams = s.getNormalizedGrams(item)

	exactKey = s.makeExactServingKey(normalizedNameForCache, normalizedBrand, normalizedGrams)
	fallbackKey = s.makeExactServingKey(fallbackNormalizedName, normalizedBrand, normalizedGrams)

	return exactKey, fallbackKey, normalizedNameForCache, normalizedBrand, normalizedGrams
}

// tryExactKey attempts to find a fresh exact cache match
func (s *NutritionService) tryExactKey(ctx context.Context, key, normalizedBrand string) (*CompleteNutrient, bool, error) {
	if cached, err := s.store.GetItemByName(ctx, key, normalizedBrand); err == nil && cached != nil {
		if isFresh(cached.UpdatedAt, cacheTTL) {
			logHydrationDecision("exact_cache_hit", "key", key, "age_days", int(time.Since(cached.UpdatedAt).Hours()/24))
			nutrition := s.convertExactCachedToNutrients(cached)
			return &nutrition, true, nil
		}
	}
	return nil, false, nil
}

// tryFallbackKey attempts to find a fresh fallback cache match
func (s *NutritionService) tryFallbackKey(ctx context.Context, key, normalizedBrand string) (*CompleteNutrient, bool, error) {
	if cached, err := s.store.GetItemByName(ctx, key, normalizedBrand); err == nil && cached != nil {
		if isFresh(cached.UpdatedAt, cacheTTL) {
			logHydrationDecision("fallback_cache_hit", "key", key, "age_days", int(time.Since(cached.UpdatedAt).Hours()/24))
			nutrition := s.convertExactCachedToNutrients(cached)
			return &nutrition, true, nil
		}
	}
	return nil, false, nil
}

// findScalableServings attempts to find cached servings that can be scaled
func (s *NutritionService) findScalableServings(ctx context.Context, item Item, normalizedName, normalizedBrand string) (*CompleteNutrient, bool, error) {
	logHydrationDecision("checking_scalable_servings", "normalized_name", normalizedName, "normalized_brand", normalizedBrand)
	cachedServings := s.getCachedServingSizes(ctx, normalizedName, normalizedBrand)
	logHydrationDecision("found_cached_servings", "count", len(cachedServings))

	for _, cachedServing := range cachedServings {
		if isFresh(cachedServing.item.UpdatedAt, cacheTTL) {
			logHydrationDecision("scalable_cache_hit", "cached_grams", cachedServing.servingGrams, "age_days", int(time.Since(cachedServing.item.UpdatedAt).Hours()/24))

			// Determine scaling method based on user input and cached data reliability
			if s.shouldUse100gScaling(item, cachedServing.item) {
				logHydrationDecision("using_per100g_scaling", "name", item.Name, "requested_grams", item.Grams)
				nutrition := s.convertCachedToNutrients(cachedServing.item, item)
				return &nutrition, true, nil
			} else {
				logHydrationDecision("using_serving_scaling", "name", item.Name, "cached_grams", cachedServing.servingGrams, "requested_grams", item.Grams)
				nutrition := s.scaleNutritionFromCachedServing(cachedServing.item, cachedServing.servingGrams, item.Grams)
				return &nutrition, true, nil
			}
		}
	}
	return nil, false, nil
}

// fetchNutritionFromCache attempts to fetch nutrition data from cache
// Returns nutrition data if found, nil if not found or cache is stale
func (s *NutritionService) fetchNutritionFromCache(ctx context.Context, item Item) (*CompleteNutrient, error) {
	if s.store == nil {
		return nil, nil // No cache available
	}

	// Build cache keys
	exactKey, fallbackKey, normalizedName, normalizedBrand, _ := s.buildCacheKeys(item)

	logHydrationDecision("cache_lookup_start", "original_name", item.Name, "normalized_name", normalizedName,
		"normalized_brand", normalizedBrand, "grams", item.Grams)

	// Try exact key first (brand-aware)
	logHydrationDecision("trying_exact_key", "exact_key", exactKey)
	if nutrition, found, err := s.tryExactKey(ctx, exactKey, normalizedBrand); err != nil {
		return nil, err
	} else if found {
		// Handle BaseQuantity scaling if needed
		if item.BaseQuantity != nil && *item.BaseQuantity > 1.0 {
			logHydrationDecision("scaling_for_base_quantity", "name", item.Name, "base_quantity", *item.BaseQuantity)
			singleUnitNutrition := *nutrition
			scalingFactor := *item.BaseQuantity
			singleUnitNutrition.Scale(scalingFactor)
			return &singleUnitNutrition, nil
		}
		logHydrationDecision("returning_exact_match", "name", item.Name, "grams", item.Grams)
		return nutrition, nil
	}

	// Try fallback key (backward compatibility)
	if fallbackKey != exactKey {
		logHydrationDecision("trying_fallback_key", "fallback_key", fallbackKey)
		if nutrition, found, err := s.tryFallbackKey(ctx, fallbackKey, normalizedBrand); err != nil {
			return nil, err
		} else if found {
			logHydrationDecision("returning_fallback_match", "name", item.Name, "grams", item.Grams)
			return nutrition, nil
		}
	}

	// Try scalable servings
	if nutrition, found, err := s.findScalableServings(ctx, item, normalizedName, normalizedBrand); err != nil {
		return nil, err
	} else if found {
		return nutrition, nil
	}

	logHydrationDecision("cache_miss", "name", item.Name, "brand", getBrandOrEmpty(item.Brand))
	return nil, nil // No cache hit
}

// getAINutrition handles AI nutrition lookup for both generic and branded items
func (s *NutritionService) getAINutrition(ctx context.Context, item Item, provider AIProvider) (CompleteNutrient, []storage.OFFIngredient, *string, error) {
	// Determine if this is a generic item (no brand)
	isGenericItem := (item.Brand == nil || (item.Brand != nil && *item.Brand == ""))
	logHydrationDecision("ai_lookup", "item", item.Name, "is_generic", isGenericItem)

	if isGenericItem {
		// No brand - use complete AI response for ingredients
		aiResponse, err := provider.GetNutritionWithContextComplete(ctx, item)
		if err != nil {
			return CompleteNutrient{}, nil, nil, err
		}

		logHydrationDecision("ai_complete_response", "item", item.Name, "calories", aiResponse.Nutrients.Calories)

		// Extract ingredients and URL from AI response for generic items
		var ingredients []storage.OFFIngredient
		var url *string

		if len(aiResponse.Ingredients) > 0 {
			ingredients = aiResponse.Ingredients
			logHydrationDecision("extracted_ingredients", "item", item.Name, "ingredient_count", len(ingredients))
		}

		if aiResponse.URL != nil && *aiResponse.URL != "" {
			url = aiResponse.URL
			logHydrationDecision("extracted_url", "item", item.Name, "url", *url)
		}

		return aiResponse.Nutrients, ingredients, url, nil
	} else {
		// Branded items - use standard nutrition only
		nutrition, err := provider.GetNutritionWithContext(ctx, item)
		if err != nil {
			return CompleteNutrient{}, nil, nil, err
		}
		logHydrationDecision("ai_standard_response", "item", item.Name, "calories", nutrition.Calories)
		return nutrition, nil, nil, nil
	}
}

// fetchNutritionFromAI fetches nutrition data from AI provider
func (s *NutritionService) fetchNutritionFromAI(ctx context.Context, item *Item, _ interface{}) (*CompleteNutrient, error) {
	logHydrationDecision("ai_fetch_start", "name", item.Name)

	nutrition, ingredients, url, err := s.getAINutrition(ctx, *item, s.aiProvider)
	if err != nil {
		return nil, err
	}

	// Update item with extracted data
	if ingredients != nil {
		item.Ingredients = ingredients
	}
	if url != nil {
		item.Url = url
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
	logHydrationDecision("hydration_start", "item_count", len(items))

	if len(items) == 0 {
		return items, nil
	}

	// Create errgroup for concurrency control and context cancellation
	g, ctx := errgroup.WithContext(ctx)

	// Limit concurrent AI calls to 6 (reasonable for API rate limits)
	const maxConcurrentAICalls = 6
	semaphore := make(chan struct{}, maxConcurrentAICalls)

	// Results channel to collect hydrated items
	hydratedItems := make([]Item, len(items))

	// Process each item with concurrency limiting
	for i, item := range items {
		i, item := i, item // Capture loop variables
		g.Go(func() error {
			// Acquire semaphore for AI call limiting
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return ctx.Err()
			}

			hydratedItem, err := s.hydrateItemNutrition(ctx, item)
			if err != nil {
				logHydrationDecision("hydration_failed", "index", i, "name", item.Name, "error", err.Error())
				// Use original item without nutrition data rather than failing entire request
				hydratedItems[i] = item
				return nil // Don't fail the entire batch for individual failures
			}

			hydratedItems[i] = hydratedItem
			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("nutrition hydration failed: %w", err)
	}

	logHydrationDecision("hydration_completed", "hydrated_count", len(hydratedItems))
	return hydratedItems, nil
}

// HydrateNutritionWithoutCache hydrates items without using cache (direct AI calls)
func (s *NutritionService) HydrateNutritionWithoutCache(ctx context.Context, items []Item) ([]Item, error) {
	logHydrationDecision("hydration_without_cache_start", "item_count", len(items))

	if len(items) == 0 {
		return items, nil
	}

	// Create errgroup for concurrency control and context cancellation
	g, ctx := errgroup.WithContext(ctx)

	// Limit concurrent AI calls to 6 (reasonable for API rate limits)
	const maxConcurrentAICalls = 6
	semaphore := make(chan struct{}, maxConcurrentAICalls)

	// Results to collect hydrated items
	hydratedItems := make([]Item, len(items))

	// Process each item with concurrency limiting
	for i, item := range items {
		i, item := i, item // Capture loop variables
		g.Go(func() error {
			// Acquire semaphore for AI call limiting
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return ctx.Err()
			}

			// Get nutrition from AI (using extracted helper)
			nutrition, ingredients, url, err := s.getAINutrition(ctx, item, s.aiProvider)
			if err != nil {
				logHydrationDecision("ai_call_failed", "index", i, "name", item.Name, "error", err.Error())
				// Use original item without nutrition data rather than failing entire request
				hydratedItems[i] = item
				return nil // Don't fail the entire batch for individual failures
			}

			// Update item with nutrition and extracted data
			item.Nutrients = &nutrition
			if ingredients != nil {
				item.Ingredients = ingredients
			}
			if url != nil {
				item.Url = url
			}

			hydratedItems[i] = item
			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("nutrition hydration without cache failed: %w", err)
	}

	logHydrationDecision("hydration_without_cache_completed", "hydrated_count", len(hydratedItems))
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
	return ConvertPer100gToServing(cached, item.Grams, s.converter)
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
	return ScaleNutritionFromCachedServing(cachedItem, fromGrams, toGrams)
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
	return ConvertExactCachedToNutrients(cached)
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

	// Create the cache item with metadata
	cacheItem := &storage.Item{
		NormalizedName:  exactKey, // Use the special key as the normalized name
		NormalizedBrand: normalizeItemName(getBrandOrEmpty(item.Brand)),
		DisplayName:     item.Name,
		DisplayBrand:    getBrandOrEmpty(item.Brand),
		Note:            item.Note,

		// Include ingredients and AI URL when caching items
		Ingredients: item.Ingredients, // Copy ingredients from the item
		Url:         item.Url,         // Copy AI URL from the item

		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Use generated helpers to populate nutrition fields
	ConvertNutrientsToExactCacheFields(cacheItem, normalizedNutrients, normalizedGrams)
	ConvertNutrientsToPer100gCacheFields(cacheItem, normalizedNutrients, normalizedGrams, s.converter)

	return cacheItem
}

// floatValue safely dereferences a float64 pointer, returning 0 if nil
func floatValue(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// isFresh checks if a cache entry is still within its TTL
func isFresh(updatedAt time.Time, ttl time.Duration) bool {
	return time.Since(updatedAt) < ttl
}

// isGramUnit checks if a unit string represents grams
func isGramUnit(u *string) bool {
	if u == nil {
		return false
	}
	unit := *u
	return unit == "g" || unit == "grams" || unit == "gram"
}

// logHydrationDecision logs hydration-related decisions with consistent formatting
func logHydrationDecision(stage string, kv ...any) {
	args := []any{"stage", stage}
	args = append(args, kv...)
	LogDebug("Nutrition hydration decision", args...)
}
