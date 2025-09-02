package server

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// Constants for nutrition service configuration
const (
	// Cache TTL for nutrition data
	cacheTTLDays = 30
	cacheTTL     = cacheTTLDays * 24 * time.Hour

	// Common serving sizes for cache lookups (in grams)
	servingSizes = "100,355,250,200,500,150,300,400,50,75,125"

	// Decimal precision for different nutrient types
	caloriesPrecision     = 3
	proteinPrecision      = 2
	fatPrecision          = 2
	transFatPrecision     = 1
	cholesterolPrecision  = 1
	sodiumPrecision       = 1
	carbsPrecision        = 1
	fiberPrecision        = 1
	sugarsPrecision       = 1
	vitaminPrecision      = 1
	vitaminBPrecision     = 3 // For B vitamins that need higher precision
	mineralPrecision      = 1
	tracePrecision        = 3 // For trace elements like omega-3s
	zincPrecision         = 2
	vitaminB12Precision   = 2
	omegaPrecision        = 3
	omega6Precision       = 2
	alcoholPrecision      = 2
)

// getCommonServingSizes returns common serving sizes for cache lookups
func getCommonServingSizes() []float64 {
	return []float64{100, 355, 250, 200, 500, 150, 300, 400, 50, 75, 125}
}

// NutritionService handles nutrition data processing with caching and unit conversions
type NutritionService struct {
	aiProvider AIProvider
	offClient  *OFFClient
	converter  *UnitConverter
	store      storage.Store
}

// scaleNutritionData scales nutrition values by the given factor
func scaleNutritionData(nutrients CompleteNutrient, factor float64) CompleteNutrient {
	return CompleteNutrient{
		Calories:           nutrients.Calories * factor,
		Protein:            nutrients.Protein * factor,
		TotalFat:           nutrients.TotalFat * factor,
		SaturatedFat:       nutrients.SaturatedFat * factor,
		TransFat:           nutrients.TransFat * factor,
		Cholesterol:        nutrients.Cholesterol * factor,
		Sodium:             nutrients.Sodium * factor,
		TotalCarbs:         nutrients.TotalCarbs * factor,
		DietaryFiber:       nutrients.DietaryFiber * factor,
		TotalSugars:        nutrients.TotalSugars * factor,
		AddedSugars:        nutrients.AddedSugars * factor,
		VitaminA:           nutrients.VitaminA * factor,
		VitaminC:           nutrients.VitaminC * factor,
		VitaminD:           nutrients.VitaminD * factor,
		VitaminE:           nutrients.VitaminE * factor,
		VitaminK:           nutrients.VitaminK * factor,
		Thiamine:           nutrients.Thiamine * factor,
		Riboflavin:         nutrients.Riboflavin * factor,
		Niacin:             nutrients.Niacin * factor,
		VitaminB6:          nutrients.VitaminB6 * factor,
		Folate:             nutrients.Folate * factor,
		VitaminB12:         nutrients.VitaminB12 * factor,
		Biotin:             nutrients.Biotin * factor,
		PantothenicAcid:    nutrients.PantothenicAcid * factor,
		Choline:            nutrients.Choline * factor,
		Calcium:            nutrients.Calcium * factor,
		Iron:               nutrients.Iron * factor,
		Magnesium:          nutrients.Magnesium * factor,
		Phosphorus:         nutrients.Phosphorus * factor,
		Potassium:          nutrients.Potassium * factor,
		Zinc:               nutrients.Zinc * factor,
		Copper:             nutrients.Copper * factor,
		Manganese:          nutrients.Manganese * factor,
		Selenium:           nutrients.Selenium * factor,
		Iodine:             nutrients.Iodine * factor,
		Molybdenum:         nutrients.Molybdenum * factor,
		Chromium:           nutrients.Chromium * factor,
		Fluoride:           nutrients.Fluoride * factor,
		Chloride:           nutrients.Chloride * factor,
		Omega3Ala:          nutrients.Omega3Ala * factor,
		Omega3Epa:          nutrients.Omega3Epa * factor,
		Omega3Dha:          nutrients.Omega3Dha * factor,
		Omega6:             nutrients.Omega6 * factor,
		Creatine:           nutrients.Creatine * factor,
		Caffeine:           nutrients.Caffeine * factor,
		Alcohol:            nutrients.Alcohol * factor,
		PolyunsaturatedFat: nutrients.PolyunsaturatedFat * factor,
		MonounsaturatedFat: nutrients.MonounsaturatedFat * factor,
	}
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
	if cached, err := s.store.GetItemByName(ctx, exactKey, ""); err == nil && cached != nil {
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
		if cached, err := s.store.GetItemByName(ctx, fallbackExactKey, ""); err == nil && cached != nil {
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

// fetchNutritionContext fetches nutrition context from Open Food Facts
// Returns context data and ingredients/URL information to be added to the item
func (s *NutritionService) fetchNutritionContext(ctx context.Context, item *Item) (interface{}, error) {
	// Only query OFF if we have a brand (OFF is only good for branded items)
	brand := getBrandOrEmpty(item.Brand)
	if s.offClient == nil || brand == "" || strings.TrimSpace(brand) == "" {
		LogDebug("Skipping OFF database query - no brand available", "name", item.Name, "brand", brand)
		return nil, nil
	}

	LogDebug("Checking OFF database for item context", "name", item.Name, "brand", brand)

	offProduct, err := s.offClient.SearchProduct(ctx, item.Name, brand)
	if err != nil || offProduct == nil {
		LogDebug("Item not found in OFF database", "name", item.Name, "error", err)
		return nil, nil
	}

	LogDebug("Found item in OFF database for context", "name", item.Name, "product_name", offProduct.ProductName)

	// Extract ingredients and OFF URL BEFORE creating context so they get saved
	if len(offProduct.Ingredients) > 0 {
		item.Ingredients = parseOFFIngredients(offProduct.Ingredients)
		LogDebug("Extracted ingredients from OFF", "item", item.Name, "ingredient_count", len(item.Ingredients))
	}

	if offProduct.Link != "" {
		item.Url = &offProduct.Link
		LogDebug("Saved OFF URL", "item", item.Name, "url", offProduct.Link)
	}

	// Create nutrition context for AI
	nutritionContext := map[string]interface{}{
		"source": "open_food_facts",
		"products": []interface{}{
			map[string]interface{}{
				"product_name":          offProduct.ProductName,
				"brands":                offProduct.Brands,
				"nutrients":             offProduct.Nutriments,
				"serving_quantity":      offProduct.ServingQuantity,
				"serving_quantity_unit": offProduct.ServingQuantityUnit,
				"serving_size":          offProduct.ServingSize,
				"ingredients":           parseOFFIngredients(offProduct.Ingredients),
				"link":                  offProduct.Link,
				"grade":                 offProduct.Grade,
				"is_beverage":           offProduct.IsBeverage,
			},
		},
		"note": "This context provides real product data from Open Food Facts that may help inform nutrition estimates. Use this data as reference but provide complete nutrition data including nutrients not available in the context.",
	}

	return nutritionContext, nil
}

// fetchNutritionFromAI fetches nutrition data from AI provider with optional context
func (s *NutritionService) fetchNutritionFromAI(ctx context.Context, item *Item, nutritionContext interface{}) (*CompleteNutrient, error) {
	LogDebug("Fetching nutrition from AI provider", "name", item.Name, "has_context", nutritionContext != nil)

	// For generic items without OFF data, use complete AI response to get ingredients
	isGenericItem := (item.Brand == nil || (item.Brand != nil && *item.Brand == "")) && nutritionContext == nil
	LogDebug("Checking if item is generic", "item", item.Name, "brand_nil", item.Brand == nil, 
		"brand_empty", item.Brand != nil && *item.Brand == "", "has_context", nutritionContext != nil, "is_generic", isGenericItem)

	var nutrition CompleteNutrient
	var err error

	if isGenericItem {
		// No OFF data and no brand - use complete AI response for ingredients
		aiResponse, aiErr := s.aiProvider.GetNutritionWithContextComplete(ctx, *item, nutritionContext)
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
		// Branded items or items with OFF context - use standard nutrition only
		nutrition, err = s.aiProvider.GetNutritionWithContext(ctx, *item, nutritionContext)
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
		Ingredients: item.Ingredients, // Include ingredients from OFF/AI
		Url:         item.Url,         // Include OFF/AI URL
	}

	// Use brand-aware normalization for consistent cache keys
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))
	baseNormalizedName := normalizeItemNameForCache(baseName, item.Brand)

	// Cache under the base serving size
	exactKey := s.makeExactServingKey(baseNormalizedName, normalizedBrand, baseGrams)
	exactCacheItem := s.convertNutrientsToExactCache(baseItem, baseNutrition, exactKey)
	
	if exactCached, _ := s.store.GetItemByName(ctx, exactKey, ""); exactCached != nil {
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

	// Create OFF client with config from environment
	offConfig := OFFClientConfig{
		BaseURL:   getenv("OFF_API_URL", "https://world.openfoodfacts.org"),
		UserAgent: getenv("OFF_USER_AGENT", "noot/0.1 (https://github.com/GrantBirki/noot)"),
		Timeout:   time.Duration(getenvInt("OFF_TIMEOUT", 5)) * time.Second,
		Enabled:   getenvBool("OFF_ENABLED", true),
	}

	aiProvider := NewOpenAIProvider(config)
	offClient := NewOFFClient(offConfig)
	converter := NewUnitConverter()

	return &NutritionService{
		aiProvider: aiProvider,
		offClient:  offClient,
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
			// Try to get OFF context for this item
			var nutritionContext interface{}
			var directNutrition *CompleteNutrient
			var offProduct *OFFProduct // Declare here so we can use it later for ingredients

			// Only query OFF if we have a brand (OFF is only good for branded items)
			brand := getBrandOrEmpty(item.Brand)
			if s.offClient != nil && brand != "" && strings.TrimSpace(brand) != "" {
				var err error
				offProduct, err = s.offClient.SearchProduct(ctx, item.Name, brand)
				if err == nil && offProduct != nil {
					// Check if we have an exact serving size match - use direct OFF data
					if offProduct.ServingQuantity != nil &&
						float64(*offProduct.ServingQuantity) == item.Grams {
						LogDebug("Exact serving size match found - using direct OFF nutrition",
							"item", item.Name,
							"serving_quantity", float64(*offProduct.ServingQuantity),
							"item_grams", item.Grams)

						// Use precise OFF nutrition directly
						nutrition := s.offClient.ConvertToCompleteNutrient(offProduct, item.Grams)
						directNutrition = &nutrition
					} else {
						// Fall back to AI with OFF context
						LogDebug("No exact serving match - using OFF as AI context",
							"item", item.Name,
							"serving_quantity", offProduct.ServingQuantity,
							"item_grams", item.Grams)
					}

					productInfo := map[string]interface{}{
						"product_name": offProduct.ProductName,
						"brands":       offProduct.Brands,
						"nutrients":    offProduct.Nutriments,
					}

					// Add serving size information if available
					if offProduct.ServingQuantity != nil {
						productInfo["serving_quantity"] = fmt.Sprintf("%.0f", float64(*offProduct.ServingQuantity))
					}
					if offProduct.ServingQuantityUnit != "" {
						productInfo["serving_quantity_unit"] = offProduct.ServingQuantityUnit
					}
					if offProduct.ServingSize != "" {
						productInfo["serving_size"] = offProduct.ServingSize
					}

					// Add additional product information if available
					if len(offProduct.Ingredients) > 0 {
						productInfo["ingredients"] = parseOFFIngredients(offProduct.Ingredients)
					}
					if offProduct.Link != "" {
						productInfo["link"] = offProduct.Link
					}
					if offProduct.Grade != "" {
						productInfo["grade"] = offProduct.Grade
					}
					if offProduct.IsBeverage != nil {
						productInfo["is_beverage"] = *offProduct.IsBeverage
					}

					nutritionContext = map[string]interface{}{
						"source": "open_food_facts",
						"products": []interface{}{
							productInfo,
						},
						"note": "This context provides real product data from Open Food Facts that may help inform nutrition estimates. Use this data as a reference but provide complete nutrition data including nutrients not available in the context. This data could be a closely related product, the exact product, or an entirely incorrect product. Please inspect it carefully and use your best judgement. If ingredients/link are provided and seem to match the user's input, you may optionally include them in your response.",
					}
				} else {
					LogDebug("Item not found in OFF database", "name", item.Name, "brand", brand, "error", err)
				}
			} else {
				LogDebug("Skipping OFF database query - no brand available", "name", item.Name, "brand", brand)
			}

			// Extract ingredients and OFF URL from OFF product data
			if offProduct != nil {
				// Extract and convert ingredients from OFF format to our format
				if len(offProduct.Ingredients) > 0 {
					item.Ingredients = parseOFFIngredients(offProduct.Ingredients)
					LogDebug("Extracted ingredients from OFF", "item", item.Name, "ingredient_count", len(item.Ingredients))
				}

				// Save OFF URL for historical reference
				if offProduct.Link != "" {
					item.Url = &offProduct.Link
					LogDebug("Saved OFF URL", "item", item.Name, "url", offProduct.Link)
				}
			}

			// Use direct OFF nutrition if available, otherwise use AI
			var nutrition CompleteNutrient
			var err error
			if directNutrition != nil {
				nutrition = *directNutrition
				LogDebug("Using direct OFF nutrition", "item", item.Name, "calories", nutrition.Calories)
			} else {
				// For generic items without OFF data, use complete AI response to get ingredients
				isGenericItem := offProduct == nil && (item.Brand == nil || (item.Brand != nil && *item.Brand == ""))
				LogDebug("Checking if item is generic", "item", item.Name, "offProduct_nil", offProduct == nil, "brand_nil", item.Brand == nil, "brand_empty", item.Brand != nil && *item.Brand == "", "is_generic", isGenericItem)

				if isGenericItem {
					// No OFF data and no brand - use complete AI response for ingredients
					aiResponse, aiErr := s.aiProvider.GetNutritionWithContextComplete(ctx, item, nutritionContext)
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
					// Branded items or items with OFF context - use standard nutrition only
					nutrition, err = s.aiProvider.GetNutritionWithContext(ctx, item, nutritionContext)
					if err != nil {
						results <- result{index: index, item: item, err: err}
						return
					}
					LogDebug("Using AI nutrition", "item", item.Name, "calories", nutrition.Calories)
				}
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

	// Step 2: Fetch nutrition context from OFF and extract ingredients/URL
	nutritionContext, err := s.fetchNutritionContext(ctx, &item)
	if err != nil {
		LogWarn("Error fetching OFF context", "error", err.Error())
	}

	// Step 3: Fetch nutrition from AI with context
	nutrition, err := s.fetchNutritionFromAI(ctx, &item, nutritionContext)
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

	// Helper function to convert and round in one step
	convertAndRound := func(per100gValue float64, decimalPlaces int) float64 {
		return RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(per100gValue, actualGrams), decimalPlaces)
	}

	// Convert from per-100g cache data to actual weight
	return CompleteNutrient{
		Calories:           convertAndRound(cached.CaloriesPer100g, caloriesPrecision),
		Protein:            convertAndRound(cached.ProteinGPer100g, proteinPrecision),
		TotalFat:           convertAndRound(cached.TotalFatGPer100g, fatPrecision),
		SaturatedFat:       convertAndRound(cached.SaturatedFatGPer100g, fatPrecision),
		TransFat:           convertAndRound(cached.TransFatGPer100g, transFatPrecision),
		Cholesterol:        convertAndRound(cached.CholesterolMgPer100g, cholesterolPrecision),
		Sodium:             convertAndRound(cached.SodiumMgPer100g, sodiumPrecision),
		TotalCarbs:         convertAndRound(cached.TotalCarbsGPer100g, carbsPrecision),
		DietaryFiber:       convertAndRound(cached.DietaryFiberGPer100g, fiberPrecision),
		TotalSugars:        convertAndRound(cached.TotalSugarsGPer100g, sugarsPrecision),
		AddedSugars:        convertAndRound(cached.AddedSugarsGPer100g, sugarsPrecision),
		VitaminA:           convertAndRound(cached.VitaminAMcgPer100g, vitaminPrecision),
		VitaminC:           convertAndRound(cached.VitaminCMgPer100g, vitaminPrecision),
		VitaminD:           convertAndRound(cached.VitaminDMcgPer100g, vitaminPrecision),
		VitaminE:           convertAndRound(cached.VitaminEMgPer100g, vitaminPrecision),
		VitaminK:           convertAndRound(cached.VitaminKMcgPer100g, vitaminPrecision),
		Thiamine:           convertAndRound(cached.ThiamineMgPer100g, vitaminBPrecision),
		Riboflavin:         convertAndRound(cached.RiboflavinMgPer100g, vitaminBPrecision),
		Niacin:             convertAndRound(cached.NiacinMgPer100g, vitaminPrecision),
		VitaminB6:          convertAndRound(cached.VitaminB6MgPer100g, vitaminBPrecision),
		Folate:             convertAndRound(cached.FolateMcgPer100g, vitaminPrecision),
		VitaminB12:         convertAndRound(cached.VitaminB12McgPer100g, vitaminB12Precision),
		Calcium:            convertAndRound(cached.CalciumMgPer100g, mineralPrecision),
		Iron:               convertAndRound(cached.IronMgPer100g, mineralPrecision),
		Magnesium:          convertAndRound(cached.MagnesiumMgPer100g, mineralPrecision),
		Phosphorus:         convertAndRound(cached.PhosphorusMgPer100g, mineralPrecision),
		Potassium:          convertAndRound(cached.PotassiumMgPer100g, mineralPrecision),
		Zinc:               convertAndRound(cached.ZincMgPer100g, zincPrecision),
		Copper:             convertAndRound(cached.CopperMgPer100g, tracePrecision),
		Manganese:          convertAndRound(cached.ManganeseMgPer100g, tracePrecision),
		Selenium:           convertAndRound(cached.SeleniumMcgPer100g, mineralPrecision),
		Iodine:             convertAndRound(cached.IodineMcgPer100g, mineralPrecision),
		Molybdenum:         convertAndRound(cached.MolybdenumMcgPer100g, mineralPrecision),
		Chromium:           convertAndRound(cached.ChromiumMcgPer100g, mineralPrecision),
		Fluoride:           convertAndRound(cached.FluorideMgPer100g, mineralPrecision),
		Chloride:           convertAndRound(cached.ChlorideMgPer100g, mineralPrecision),
		Omega3Ala:          convertAndRound(cached.Omega3AlaGPer100g, omegaPrecision),
		Omega3Epa:          convertAndRound(cached.Omega3EpaGPer100g, omegaPrecision),
		Omega3Dha:          convertAndRound(cached.Omega3DhaGPer100g, omegaPrecision),
		Omega6:             convertAndRound(cached.Omega6GPer100g, omega6Precision),
		Creatine:           convertAndRound(cached.CreatineMgPer100g, mineralPrecision),
		Caffeine:           convertAndRound(cached.CaffeineMgPer100g, mineralPrecision),
		Alcohol:            convertAndRound(cached.AlcoholGPer100g, alcoholPrecision),
		PolyunsaturatedFat: convertAndRound(cached.PolyunsaturatedFatGPer100g, fatPrecision),
		MonounsaturatedFat: convertAndRound(cached.MonounsaturatedFatGPer100g, fatPrecision),
	}
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
		if cached, err := s.store.GetItemByName(ctx, exactKey, ""); err == nil && cached != nil {
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
			if cached, err := s.store.GetItemByName(ctx, exactKey, ""); err == nil && cached != nil {
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

	// Helper function to safely dereference pointers and scale
	scaleValue := func(ptr *float64) float64 {
		if ptr == nil {
			return 0.0
		}
		return *ptr * scalingFactor
	}

	return CompleteNutrient{
		Calories:           float64(RoundCaloriesUp(scaleValue(cachedItem.OriginalCalories))),
		Protein:            scaleValue(cachedItem.OriginalProteinG),
		TotalFat:           scaleValue(cachedItem.OriginalTotalFatG),
		SaturatedFat:       scaleValue(cachedItem.OriginalSaturatedFatG),
		TransFat:           scaleValue(cachedItem.OriginalTransFatG),
		Cholesterol:        scaleValue(cachedItem.OriginalCholesterolMg),
		Sodium:             scaleValue(cachedItem.OriginalSodiumMg),
		TotalCarbs:         scaleValue(cachedItem.OriginalTotalCarbsG),
		DietaryFiber:       scaleValue(cachedItem.OriginalDietaryFiberG),
		TotalSugars:        scaleValue(cachedItem.OriginalTotalSugarsG),
		AddedSugars:        scaleValue(cachedItem.OriginalAddedSugarsG),
		VitaminA:           scaleValue(cachedItem.OriginalVitaminAMcg),
		VitaminC:           scaleValue(cachedItem.OriginalVitaminCMg),
		VitaminD:           scaleValue(cachedItem.OriginalVitaminDMcg),
		VitaminE:           scaleValue(cachedItem.OriginalVitaminEMg),
		VitaminK:           scaleValue(cachedItem.OriginalVitaminKMcg),
		Thiamine:           scaleValue(cachedItem.OriginalThiamineMg),
		Riboflavin:         scaleValue(cachedItem.OriginalRiboflavinMg),
		Niacin:             scaleValue(cachedItem.OriginalNiacinMg),
		VitaminB6:          scaleValue(cachedItem.OriginalVitaminB6Mg),
		Folate:             scaleValue(cachedItem.OriginalFolateMcg),
		VitaminB12:         scaleValue(cachedItem.OriginalVitaminB12Mcg),
		Biotin:             scaleValue(cachedItem.OriginalBiotinMcg),
		PantothenicAcid:    scaleValue(cachedItem.OriginalPantothenicAcidMg),
		Choline:            scaleValue(cachedItem.OriginalCholineMg),
		Calcium:            scaleValue(cachedItem.OriginalCalciumMg),
		Iron:               scaleValue(cachedItem.OriginalIronMg),
		Magnesium:          scaleValue(cachedItem.OriginalMagnesiumMg),
		Phosphorus:         scaleValue(cachedItem.OriginalPhosphorusMg),
		Potassium:          scaleValue(cachedItem.OriginalPotassiumMg),
		Zinc:               scaleValue(cachedItem.OriginalZincMg),
		Copper:             scaleValue(cachedItem.OriginalCopperMg),
		Manganese:          scaleValue(cachedItem.OriginalManganeseMg),
		Selenium:           scaleValue(cachedItem.OriginalSeleniumMcg),
		Iodine:             scaleValue(cachedItem.OriginalIodineMcg),
		Molybdenum:         scaleValue(cachedItem.OriginalMolybdenumMcg),
		Chromium:           scaleValue(cachedItem.OriginalChromiumMcg),
		Fluoride:           scaleValue(cachedItem.OriginalFluorideMg),
		Chloride:           scaleValue(cachedItem.OriginalChlorideMg),
		Omega3Ala:          scaleValue(cachedItem.OriginalOmega3AlaG),
		Omega3Epa:          scaleValue(cachedItem.OriginalOmega3EpaG),
		Omega3Dha:          scaleValue(cachedItem.OriginalOmega3DhaG),
		Omega6:             scaleValue(cachedItem.OriginalOmega6G),
		Creatine:           scaleValue(cachedItem.OriginalCreatineMg),
		Caffeine:           scaleValue(cachedItem.OriginalCaffeineMg),
		Alcohol:            scaleValue(cachedItem.OriginalAlcoholG),
		PolyunsaturatedFat: scaleValue(cachedItem.OriginalPolyunsaturatedFatG),
		MonounsaturatedFat: scaleValue(cachedItem.OriginalMonounsaturatedFatG),
	}
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
		return CompleteNutrient{
			Calories:           floatValue(cached.OriginalCalories),
			Protein:            floatValue(cached.OriginalProteinG),
			TotalFat:           floatValue(cached.OriginalTotalFatG),
			SaturatedFat:       floatValue(cached.OriginalSaturatedFatG),
			TransFat:           floatValue(cached.OriginalTransFatG),
			Cholesterol:        floatValue(cached.OriginalCholesterolMg),
			Sodium:             floatValue(cached.OriginalSodiumMg),
			TotalCarbs:         floatValue(cached.OriginalTotalCarbsG),
			DietaryFiber:       floatValue(cached.OriginalDietaryFiberG),
			TotalSugars:        floatValue(cached.OriginalTotalSugarsG),
			AddedSugars:        floatValue(cached.OriginalAddedSugarsG),
			VitaminA:           floatValue(cached.OriginalVitaminAMcg),
			VitaminC:           floatValue(cached.OriginalVitaminCMg),
			VitaminD:           floatValue(cached.OriginalVitaminDMcg),
			VitaminE:           floatValue(cached.OriginalVitaminEMg),
			VitaminK:           floatValue(cached.OriginalVitaminKMcg),
			Thiamine:           floatValue(cached.OriginalThiamineMg),
			Riboflavin:         floatValue(cached.OriginalRiboflavinMg),
			Niacin:             floatValue(cached.OriginalNiacinMg),
			VitaminB6:          floatValue(cached.OriginalVitaminB6Mg),
			Folate:             floatValue(cached.OriginalFolateMcg),
			VitaminB12:         floatValue(cached.OriginalVitaminB12Mcg),
			Biotin:             floatValue(cached.OriginalBiotinMcg),
			PantothenicAcid:    floatValue(cached.OriginalPantothenicAcidMg),
			Choline:            floatValue(cached.OriginalCholineMg),
			Calcium:            floatValue(cached.OriginalCalciumMg),
			Iron:               floatValue(cached.OriginalIronMg),
			Magnesium:          floatValue(cached.OriginalMagnesiumMg),
			Phosphorus:         floatValue(cached.OriginalPhosphorusMg),
			Potassium:          floatValue(cached.OriginalPotassiumMg),
			Zinc:               floatValue(cached.OriginalZincMg),
			Copper:             floatValue(cached.OriginalCopperMg),
			Manganese:          floatValue(cached.OriginalManganeseMg),
			Selenium:           floatValue(cached.OriginalSeleniumMcg),
			Iodine:             floatValue(cached.OriginalIodineMcg),
			Molybdenum:         floatValue(cached.OriginalMolybdenumMcg),
			Chromium:           floatValue(cached.OriginalChromiumMcg),
			Fluoride:           floatValue(cached.OriginalFluorideMg),
			Chloride:           floatValue(cached.OriginalChlorideMg),
			Omega3Ala:          floatValue(cached.OriginalOmega3AlaG),
			Omega3Epa:          floatValue(cached.OriginalOmega3EpaG),
			Omega3Dha:          floatValue(cached.OriginalOmega3DhaG),
			Omega6:             floatValue(cached.OriginalOmega6G),
			Creatine:           floatValue(cached.OriginalCreatineMg),
			Caffeine:           floatValue(cached.OriginalCaffeineMg),
			Alcohol:            floatValue(cached.OriginalAlcoholG),
			PolyunsaturatedFat: floatValue(cached.OriginalPolyunsaturatedFatG),
			MonounsaturatedFat: floatValue(cached.OriginalMonounsaturatedFatG),
		}
	}

	// Fallback to per-100g data if original data is not available (shouldn't happen with new schema)
	LogWarn("Original serving data not found, falling back to per-100g conversion",
		"normalized_name", cached.NormalizedName)
	return CompleteNutrient{
		Calories:           cached.CaloriesPer100g,
		Protein:            cached.ProteinGPer100g,
		TotalFat:           cached.TotalFatGPer100g,
		SaturatedFat:       cached.SaturatedFatGPer100g,
		TransFat:           cached.TransFatGPer100g,
		Cholesterol:        cached.CholesterolMgPer100g,
		Sodium:             cached.SodiumMgPer100g,
		TotalCarbs:         cached.TotalCarbsGPer100g,
		DietaryFiber:       cached.DietaryFiberGPer100g,
		TotalSugars:        cached.TotalSugarsGPer100g,
		AddedSugars:        cached.AddedSugarsGPer100g,
		VitaminA:           cached.VitaminAMcgPer100g,
		VitaminC:           cached.VitaminCMgPer100g,
		VitaminD:           cached.VitaminDMcgPer100g,
		VitaminE:           cached.VitaminEMgPer100g,
		VitaminK:           cached.VitaminKMcgPer100g,
		Thiamine:           cached.ThiamineMgPer100g,
		Riboflavin:         cached.RiboflavinMgPer100g,
		Niacin:             cached.NiacinMgPer100g,
		VitaminB6:          cached.VitaminB6MgPer100g,
		Folate:             cached.FolateMcgPer100g,
		VitaminB12:         cached.VitaminB12McgPer100g,
		Biotin:             cached.BiotinMcgPer100g,
		PantothenicAcid:    cached.PantothenicAcidMgPer100g,
		Choline:            cached.CholineMgPer100g,
		Calcium:            cached.CalciumMgPer100g,
		Iron:               cached.IronMgPer100g,
		Magnesium:          cached.MagnesiumMgPer100g,
		Phosphorus:         cached.PhosphorusMgPer100g,
		Potassium:          cached.PotassiumMgPer100g,
		Zinc:               cached.ZincMgPer100g,
		Copper:             cached.CopperMgPer100g,
		Manganese:          cached.ManganeseMgPer100g,
		Selenium:           cached.SeleniumMcgPer100g,
		Iodine:             cached.IodineMcgPer100g,
		Molybdenum:         cached.MolybdenumMcgPer100g,
		Chromium:           cached.ChromiumMcgPer100g,
		Fluoride:           cached.FluorideMgPer100g,
		Chloride:           cached.ChlorideMgPer100g,
		Omega3Ala:          cached.Omega3AlaGPer100g,
		Omega3Epa:          cached.Omega3EpaGPer100g,
		Omega3Dha:          cached.Omega3DhaGPer100g,
		Omega6:             cached.Omega6GPer100g,
		Creatine:           cached.CreatineMgPer100g,
		Caffeine:           cached.CaffeineMgPer100g,
		Alcohol:            cached.AlcoholGPer100g,
		PolyunsaturatedFat: cached.PolyunsaturatedFatGPer100g,
		MonounsaturatedFat: cached.MonounsaturatedFatGPer100g,
	}
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
	convertAndRound := func(servingValue float64, decimalPlaces int) float64 {
		return RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(servingValue, normalizedGrams), decimalPlaces)
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
		CaloriesPer100g:            convertAndRound(normalizedNutrients.Calories, 4), // Higher precision for calories
		ProteinGPer100g:            convertAndRound(normalizedNutrients.Protein, tracePrecision),
		TotalFatGPer100g:           convertAndRound(normalizedNutrients.TotalFat, tracePrecision),
		SaturatedFatGPer100g:       convertAndRound(normalizedNutrients.SaturatedFat, tracePrecision),
		TransFatGPer100g:           convertAndRound(normalizedNutrients.TransFat, tracePrecision),
		CholesterolMgPer100g:       convertAndRound(normalizedNutrients.Cholesterol, fatPrecision),
		SodiumMgPer100g:            convertAndRound(normalizedNutrients.Sodium, mineralPrecision),
		TotalCarbsGPer100g:         convertAndRound(normalizedNutrients.TotalCarbs, carbsPrecision),
		DietaryFiberGPer100g:       convertAndRound(normalizedNutrients.DietaryFiber, fiberPrecision),
		TotalSugarsGPer100g:        convertAndRound(normalizedNutrients.TotalSugars, sugarsPrecision),
		AddedSugarsGPer100g:        convertAndRound(normalizedNutrients.AddedSugars, sugarsPrecision),
		VitaminAMcgPer100g:         convertAndRound(normalizedNutrients.VitaminA, vitaminPrecision),
		VitaminCMgPer100g:          convertAndRound(normalizedNutrients.VitaminC, vitaminPrecision),
		VitaminDMcgPer100g:         convertAndRound(normalizedNutrients.VitaminD, vitaminPrecision),
		VitaminEMgPer100g:          convertAndRound(normalizedNutrients.VitaminE, vitaminPrecision),
		VitaminKMcgPer100g:         convertAndRound(normalizedNutrients.VitaminK, vitaminPrecision),
		ThiamineMgPer100g:          convertAndRound(normalizedNutrients.Thiamine, vitaminBPrecision),
		RiboflavinMgPer100g:        convertAndRound(normalizedNutrients.Riboflavin, vitaminBPrecision),
		NiacinMgPer100g:            convertAndRound(normalizedNutrients.Niacin, vitaminPrecision),
		VitaminB6MgPer100g:         convertAndRound(normalizedNutrients.VitaminB6, vitaminBPrecision),
		FolateMcgPer100g:           convertAndRound(normalizedNutrients.Folate, vitaminPrecision),
		VitaminB12McgPer100g:       convertAndRound(normalizedNutrients.VitaminB12, vitaminB12Precision),
		BiotinMcgPer100g:           convertAndRound(normalizedNutrients.Biotin, vitaminPrecision),
		PantothenicAcidMgPer100g:   convertAndRound(normalizedNutrients.PantothenicAcid, vitaminPrecision),
		CholineMgPer100g:           convertAndRound(normalizedNutrients.Choline, vitaminPrecision),
		CalciumMgPer100g:           convertAndRound(normalizedNutrients.Calcium, mineralPrecision),
		IronMgPer100g:              convertAndRound(normalizedNutrients.Iron, mineralPrecision),
		MagnesiumMgPer100g:         convertAndRound(normalizedNutrients.Magnesium, mineralPrecision),
		PhosphorusMgPer100g:        convertAndRound(normalizedNutrients.Phosphorus, mineralPrecision),
		PotassiumMgPer100g:         convertAndRound(normalizedNutrients.Potassium, mineralPrecision),
		ZincMgPer100g:              convertAndRound(normalizedNutrients.Zinc, zincPrecision),
		CopperMgPer100g:            convertAndRound(normalizedNutrients.Copper, tracePrecision),
		ManganeseMgPer100g:         convertAndRound(normalizedNutrients.Manganese, tracePrecision),
		SeleniumMcgPer100g:         convertAndRound(normalizedNutrients.Selenium, mineralPrecision),
		IodineMcgPer100g:           convertAndRound(normalizedNutrients.Iodine, mineralPrecision),
		MolybdenumMcgPer100g:       convertAndRound(normalizedNutrients.Molybdenum, mineralPrecision),
		ChromiumMcgPer100g:         convertAndRound(normalizedNutrients.Chromium, mineralPrecision),
		FluorideMgPer100g:          convertAndRound(normalizedNutrients.Fluoride, mineralPrecision),
		ChlorideMgPer100g:          convertAndRound(normalizedNutrients.Chloride, mineralPrecision),
		Omega3AlaGPer100g:          convertAndRound(normalizedNutrients.Omega3Ala, omegaPrecision),
		Omega3EpaGPer100g:          convertAndRound(normalizedNutrients.Omega3Epa, omegaPrecision),
		Omega3DhaGPer100g:          convertAndRound(normalizedNutrients.Omega3Dha, omegaPrecision),
		Omega6GPer100g:             convertAndRound(normalizedNutrients.Omega6, omega6Precision),
		CreatineMgPer100g:          convertAndRound(normalizedNutrients.Creatine, mineralPrecision),
		CaffeineMgPer100g:          convertAndRound(normalizedNutrients.Caffeine, mineralPrecision),
		AlcoholGPer100g:            convertAndRound(normalizedNutrients.Alcohol, alcoholPrecision),
		PolyunsaturatedFatGPer100g: convertAndRound(normalizedNutrients.PolyunsaturatedFat, fatPrecision),
		MonounsaturatedFatGPer100g: convertAndRound(normalizedNutrients.MonounsaturatedFat, fatPrecision),
		Note:                       item.Note,

		// Include ingredients and OFF URL when caching items
		Ingredients: item.Ingredients, // Copy ingredients from the item
		Url:         item.Url,         // Copy OFF URL from the item

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

// parseOFFIngredients converts raw OFF ingredient data to our OFFIngredient format
func parseOFFIngredients(rawIngredients []interface{}) []storage.OFFIngredient {
	if len(rawIngredients) == 0 {
		return []storage.OFFIngredient{}
	}

	var ingredients []storage.OFFIngredient

	for _, rawIngredient := range rawIngredients {
		// OFF ingredients come as map[string]interface{}
		ingredientMap, ok := rawIngredient.(map[string]interface{})
		if !ok {
			continue
		}

		ingredient := storage.OFFIngredient{}

		// Extract ID
		if id, ok := ingredientMap["id"].(string); ok {
			ingredient.ID = id
		}

		// Extract text/display name
		if text, ok := ingredientMap["text"].(string); ok {
			ingredient.Text = text
		}

		// Extract percentage values (they might be numbers or strings)
		if percentEstimate := extractFloatFromInterface(ingredientMap["percent_estimate"]); percentEstimate != nil {
			ingredient.PercentEstimate = percentEstimate
		}
		if percentMax := extractFloatFromInterface(ingredientMap["percent_max"]); percentMax != nil {
			ingredient.PercentMax = percentMax
		}
		if percentMin := extractFloatFromInterface(ingredientMap["percent_min"]); percentMin != nil {
			ingredient.PercentMin = percentMin
		}

		// Only add ingredient if it has meaningful data
		if ingredient.ID != "" || ingredient.Text != "" {
			ingredients = append(ingredients, ingredient)
		}
	}

	return ingredients
}

// extractFloatFromInterface safely extracts a float64 from an interface{} value
func extractFloatFromInterface(value interface{}) *float64 {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case float64:
		return &v
	case float32:
		f := float64(v)
		return &f
	case int:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case string:
		// Try to parse string as float
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return &f
		}
	}

	return nil
}
