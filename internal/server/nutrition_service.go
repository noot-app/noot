package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// NutritionService handles nutrition data processing with caching and unit conversions
type NutritionService struct {
	aiProvider AIProvider
	offClient  *OFFClient
	converter  *UnitConverter
	store      storage.Store
}

// NewNutritionService creates a new nutrition service
func NewNutritionService(store storage.Store) *NutritionService {
	// Create OpenAI provider with config from environment
	config := AIProviderConfig{
		APIKey:          os.Getenv("OPENAI_API_KEY"),
		TranscribeModel: getenv("OPENAI_TRANSCRIBE_MODEL", "gpt-4o-mini-transcribe"),
		ParseModel:      getenv("OPENAI_PARSE_MODEL", "gpt-4o-mini"),
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
			nutrition, err := s.aiProvider.GetNutrition(ctx, item)
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

// hydrateItemNutrition hydrates a single item with nutrition data from cache or AI
func (s *NutritionService) hydrateItemNutrition(ctx context.Context, item Item) (Item, error) {
	normalizedName := normalizeItemName(item.Name)
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))

	LogDebug("Checking cache for item", "normalized_name", normalizedName, "normalized_brand", normalizedBrand, "grams", item.Grams)

	// Check cache first - try to find exact serving size match
	if s.store != nil {
		normalizedGrams := s.getNormalizedGrams(item)
		exactKey := s.makeExactServingKey(normalizedName, normalizedBrand, normalizedGrams)
		if cached, err := s.store.GetItemByName(ctx, exactKey, ""); err == nil && cached != nil {
			// Check if cache is still fresh (30 days)
			if time.Since(cached.UpdatedAt) < 30*24*time.Hour {
				var nutrition CompleteNutrient

				// If item has BaseQuantity > 1, we need to scale the cached single-unit values
				if item.BaseQuantity != nil && *item.BaseQuantity > 1.0 {
					LogDebug("Using cached exact serving match with scaling for multi-unit quantity",
						"name", item.Name, "base_quantity", *item.BaseQuantity, "total_grams", item.Grams)

					// Get the single-unit nutrition values and scale by BaseQuantity
					singleUnitNutrition := s.convertExactCachedToNutrients(cached)
					scalingFactor := *item.BaseQuantity

					nutrition = CompleteNutrient{
						Calories:        singleUnitNutrition.Calories * scalingFactor,
						Protein:         singleUnitNutrition.Protein * scalingFactor,
						TotalFat:        singleUnitNutrition.TotalFat * scalingFactor,
						SaturatedFat:    singleUnitNutrition.SaturatedFat * scalingFactor,
						TransFat:        singleUnitNutrition.TransFat * scalingFactor,
						Cholesterol:     singleUnitNutrition.Cholesterol * scalingFactor,
						Sodium:          singleUnitNutrition.Sodium * scalingFactor,
						TotalCarbs:      singleUnitNutrition.TotalCarbs * scalingFactor,
						DietaryFiber:    singleUnitNutrition.DietaryFiber * scalingFactor,
						TotalSugars:     singleUnitNutrition.TotalSugars * scalingFactor,
						AddedSugars:     singleUnitNutrition.AddedSugars * scalingFactor,
						VitaminA:        singleUnitNutrition.VitaminA * scalingFactor,
						VitaminC:        singleUnitNutrition.VitaminC * scalingFactor,
						VitaminD:        singleUnitNutrition.VitaminD * scalingFactor,
						VitaminE:        singleUnitNutrition.VitaminE * scalingFactor,
						VitaminK:        singleUnitNutrition.VitaminK * scalingFactor,
						Thiamine:        singleUnitNutrition.Thiamine * scalingFactor,
						Riboflavin:      singleUnitNutrition.Riboflavin * scalingFactor,
						Niacin:          singleUnitNutrition.Niacin * scalingFactor,
						VitaminB6:       singleUnitNutrition.VitaminB6 * scalingFactor,
						Folate:          singleUnitNutrition.Folate * scalingFactor,
						VitaminB12:      singleUnitNutrition.VitaminB12 * scalingFactor,
						Biotin:          singleUnitNutrition.Biotin * scalingFactor,
						PantothenicAcid: singleUnitNutrition.PantothenicAcid * scalingFactor,
						Choline:         singleUnitNutrition.Choline * scalingFactor,
						Calcium:         singleUnitNutrition.Calcium * scalingFactor,
						Iron:            singleUnitNutrition.Iron * scalingFactor,
						Magnesium:       singleUnitNutrition.Magnesium * scalingFactor,
						Phosphorus:      singleUnitNutrition.Phosphorus * scalingFactor,
						Potassium:       singleUnitNutrition.Potassium * scalingFactor,
						Zinc:            singleUnitNutrition.Zinc * scalingFactor,
						Copper:          singleUnitNutrition.Copper * scalingFactor,
						Manganese:       singleUnitNutrition.Manganese * scalingFactor,
						Selenium:        singleUnitNutrition.Selenium * scalingFactor,
						Iodine:          singleUnitNutrition.Iodine * scalingFactor,
						Molybdenum:      singleUnitNutrition.Molybdenum * scalingFactor,
						Chromium:        singleUnitNutrition.Chromium * scalingFactor,
						Fluoride:        singleUnitNutrition.Fluoride * scalingFactor,
						Chloride:        singleUnitNutrition.Chloride * scalingFactor,
					}
				} else {
					LogDebug("Using cached exact serving match - returning original values without scaling",
						"name", item.Name, "grams", item.Grams)

					nutrition = s.convertExactCachedToNutrients(cached)
				}

				item.Nutrients = &nutrition
				return item, nil
			}
		}

		// Try to find any cached serving size for this item to scale from
		cachedServings := s.getCachedServingSizes(ctx, normalizedName, normalizedBrand)
		for _, cachedServing := range cachedServings {
			// Check if cache is still fresh (30 days)
			if time.Since(cachedServing.item.UpdatedAt) < 30*24*time.Hour {
				// Determine scaling method based on user input and cached data reliability
				if s.shouldUse100gScaling(item, cachedServing.item) {
					LogDebug("Using cached item with per-100g scaling (user provided grams or unreliable base units)",
						"name", item.Name, "requested_grams", item.Grams)

					nutrition := s.convertCachedToNutrients(cachedServing.item, item)
					item.Nutrients = &nutrition

					return item, nil
				} else {
					LogDebug("Using cached serving data - scaling from cached serving to requested serving",
						"name", item.Name, "cached_grams", cachedServing.servingGrams, "requested_grams", item.Grams)

					nutrition := s.scaleNutritionFromCachedServing(cachedServing.item, cachedServing.servingGrams, item.Grams)
					item.Nutrients = &nutrition

					return item, nil
				}
			}
		}
	}

	// Try Open Food Facts database before falling back to AI
	if s.offClient != nil {
		LogDebug("Checking OFF database for item", "name", item.Name, "brand", getBrandOrEmpty(item.Brand))
		
		offProduct, err := s.offClient.SearchProduct(ctx, item.Name, getBrandOrEmpty(item.Brand))
		if err == nil && offProduct != nil {
			LogDebug("Found item in OFF database", "name", item.Name, "product_name", offProduct.ProductName)
			
			// Convert OFF nutrition data to our format
			nutrition := s.offClient.ConvertToCompleteNutrient(offProduct, item.Grams)
			item.Nutrients = &nutrition
			
			// Cache the OFF result using the same caching logic as AI results
			if s.store != nil {
				normalizedGrams := s.getNormalizedGrams(item)
				exactKey := s.makeExactServingKey(normalizedName, normalizedBrand, normalizedGrams)
				exactCacheItem := s.convertNutrientsToExactCache(item, nutrition, exactKey)
				if exactCached, _ := s.store.GetItemByName(ctx, exactKey, ""); exactCached != nil {
					// Update existing exact cache entry
					exactCacheItem.ID = exactCached.ID
					err = s.store.UpdateItem(ctx, exactCacheItem)
				} else {
					// Create new exact cache entry
					err = s.store.CreateItem(ctx, exactCacheItem)
				}
				if err != nil {
					LogWarn("Failed to cache OFF nutrition data", "name", item.Name, "grams", item.Grams, "error", err.Error())
					// Don't fail the request if caching fails
				}
			}
			
			return item, nil
		}
		
		LogDebug("Item not found in OFF database, falling back to AI", "name", item.Name)
	}

	// Not in cache or OFF - get from AI
	LogDebug("Fetching nutrition from AI provider", "name", item.Name)
	nutrition, err := s.aiProvider.GetNutrition(ctx, item)
	if err != nil {
		return item, err
	}

	// Cache the exact serving size result
	if s.store != nil {
		normalizedGrams := s.getNormalizedGrams(item)
		exactKey := s.makeExactServingKey(normalizedName, normalizedBrand, normalizedGrams)
		exactCacheItem := s.convertNutrientsToExactCache(item, nutrition, exactKey)
		if exactCached, _ := s.store.GetItemByName(ctx, exactKey, ""); exactCached != nil {
			// Update existing exact cache entry
			exactCacheItem.ID = exactCached.ID
			err = s.store.UpdateItem(ctx, exactCacheItem)
		} else {
			// Create new exact cache entry
			err = s.store.CreateItem(ctx, exactCacheItem)
		}
		if err != nil {
			LogWarn("Failed to cache exact serving nutrition data", "name", item.Name, "grams", item.Grams, "error", err.Error())
			// Don't fail the request if caching fails
		}
	}

	item.Nutrients = &nutrition

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
		Calories:     convertAndRound(cached.CaloriesPer100g, 3), // More precision to avoid cumulative rounding errors
		Protein:      convertAndRound(cached.ProteinGPer100g, 2),
		TotalFat:     convertAndRound(cached.TotalFatGPer100g, 2),
		SaturatedFat: convertAndRound(cached.SaturatedFatGPer100g, 2),
		TransFat:     convertAndRound(cached.TransFatGPer100g, 1),
		Cholesterol:  convertAndRound(cached.CholesterolMgPer100g, 1),
		Sodium:       convertAndRound(cached.SodiumMgPer100g, 1),
		TotalCarbs:   convertAndRound(cached.TotalCarbsGPer100g, 1),
		DietaryFiber: convertAndRound(cached.DietaryFiberGPer100g, 1),
		TotalSugars:  convertAndRound(cached.TotalSugarsGPer100g, 1),
		AddedSugars:  convertAndRound(cached.AddedSugarsGPer100g, 1),
		VitaminA:     convertAndRound(cached.VitaminAMcgPer100g, 1),
		VitaminC:     convertAndRound(cached.VitaminCMgPer100g, 1),
		VitaminD:     convertAndRound(cached.VitaminDMcgPer100g, 1),
		VitaminE:     convertAndRound(cached.VitaminEMgPer100g, 1),
		VitaminK:     convertAndRound(cached.VitaminKMcgPer100g, 1),
		Thiamine:     convertAndRound(cached.ThiamineMgPer100g, 3),
		Riboflavin:   convertAndRound(cached.RiboflavinMgPer100g, 3),
		Niacin:       convertAndRound(cached.NiacinMgPer100g, 1),
		VitaminB6:    convertAndRound(cached.VitaminB6MgPer100g, 3),
		Folate:       convertAndRound(cached.FolateMcgPer100g, 1),
		VitaminB12:   convertAndRound(cached.VitaminB12McgPer100g, 2),
		Calcium:      convertAndRound(cached.CalciumMgPer100g, 1),
		Iron:         convertAndRound(cached.IronMgPer100g, 1),
		Magnesium:    convertAndRound(cached.MagnesiumMgPer100g, 1),
		Phosphorus:   convertAndRound(cached.PhosphorusMgPer100g, 1),
		Potassium:    convertAndRound(cached.PotassiumMgPer100g, 1),
		Zinc:         convertAndRound(cached.ZincMgPer100g, 2),
		Copper:       convertAndRound(cached.CopperMgPer100g, 3),
		Manganese:    convertAndRound(cached.ManganeseMgPer100g, 3),
		Selenium:     convertAndRound(cached.SeleniumMcgPer100g, 1),
	}
}

// convertNutrientsToCache converts actual weight nutrition data to per-100g for cache storage
// The LLM provides nutrition for the exact weight in grams, so we need to normalize to per-100g
func (s *NutritionService) convertNutrientsToCache(item Item, nutrients CompleteNutrient) *storage.Item {
	normalizedName := normalizeItemName(item.Name)
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))

	// Use the actual grams from the item since LLM already provided nutrition for that exact weight
	actualGrams := item.Grams

	// Helper function to convert and round in one step
	convertAndRound := func(servingValue float64, decimalPlaces int) float64 {
		return RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(servingValue, actualGrams), decimalPlaces)
	}

	// Convert from actual weight nutrition data to per-100g for consistent cache storage
	return &storage.Item{
		NormalizedName:           normalizedName,
		NormalizedBrand:          normalizedBrand,
		DisplayName:              item.Name,
		DisplayBrand:             getBrandOrEmpty(item.Brand),
		CaloriesPer100g:          convertAndRound(nutrients.Calories, 4), // Increased precision to avoid rounding errors
		ProteinGPer100g:          convertAndRound(nutrients.Protein, 3),
		TotalFatGPer100g:         convertAndRound(nutrients.TotalFat, 3),
		SaturatedFatGPer100g:     convertAndRound(nutrients.SaturatedFat, 3),
		TransFatGPer100g:         convertAndRound(nutrients.TransFat, 3),
		CholesterolMgPer100g:     convertAndRound(nutrients.Cholesterol, 2),
		SodiumMgPer100g:          convertAndRound(nutrients.Sodium, 1),
		TotalCarbsGPer100g:       convertAndRound(nutrients.TotalCarbs, 1),
		DietaryFiberGPer100g:     convertAndRound(nutrients.DietaryFiber, 1),
		TotalSugarsGPer100g:      convertAndRound(nutrients.TotalSugars, 1),
		AddedSugarsGPer100g:      convertAndRound(nutrients.AddedSugars, 1),
		VitaminAMcgPer100g:       convertAndRound(nutrients.VitaminA, 1),
		VitaminCMgPer100g:        convertAndRound(nutrients.VitaminC, 1),
		VitaminDMcgPer100g:       convertAndRound(nutrients.VitaminD, 1),
		VitaminEMgPer100g:        convertAndRound(nutrients.VitaminE, 1),
		VitaminKMcgPer100g:       convertAndRound(nutrients.VitaminK, 1),
		ThiamineMgPer100g:        convertAndRound(nutrients.Thiamine, 3),
		RiboflavinMgPer100g:      convertAndRound(nutrients.Riboflavin, 3),
		NiacinMgPer100g:          convertAndRound(nutrients.Niacin, 1),
		VitaminB6MgPer100g:       convertAndRound(nutrients.VitaminB6, 3),
		FolateMcgPer100g:         convertAndRound(nutrients.Folate, 1),
		VitaminB12McgPer100g:     convertAndRound(nutrients.VitaminB12, 2),
		BiotinMcgPer100g:         convertAndRound(nutrients.Biotin, 1),
		PantothenicAcidMgPer100g: convertAndRound(nutrients.PantothenicAcid, 1),
		CholineMgPer100g:         convertAndRound(nutrients.Choline, 1),
		CalciumMgPer100g:         convertAndRound(nutrients.Calcium, 1),
		IronMgPer100g:            convertAndRound(nutrients.Iron, 1),
		MagnesiumMgPer100g:       convertAndRound(nutrients.Magnesium, 1),
		PhosphorusMgPer100g:      convertAndRound(nutrients.Phosphorus, 1),
		PotassiumMgPer100g:       convertAndRound(nutrients.Potassium, 1),
		ZincMgPer100g:            convertAndRound(nutrients.Zinc, 2),
		CopperMgPer100g:          convertAndRound(nutrients.Copper, 3),
		ManganeseMgPer100g:       convertAndRound(nutrients.Manganese, 3),
		SeleniumMcgPer100g:       convertAndRound(nutrients.Selenium, 1),
		IodineMcgPer100g:         convertAndRound(nutrients.Iodine, 1),
		MolybdenumMcgPer100g:     convertAndRound(nutrients.Molybdenum, 1),
		ChromiumMcgPer100g:       convertAndRound(nutrients.Chromium, 1),
		FluorideMgPer100g:        convertAndRound(nutrients.Fluoride, 1),
		ChlorideMgPer100g:        convertAndRound(nutrients.Chloride, 1),
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
	}
}

// cachedServingData holds info about a cached serving size for scaling
type cachedServingData struct {
	item         *storage.Item
	servingGrams float64
}

// getCachedServingSizes finds all cached serving sizes for an item
func (s *NutritionService) getCachedServingSizes(ctx context.Context, normalizedName, normalizedBrand string) []cachedServingData {
	// For now, we'll implement a simple approach that searches for exact serving keys
	// This could be optimized with a database query in the future
	var results []cachedServingData

	// Try some common serving sizes to see if we have them cached
	commonGrams := []float64{100, 355, 250, 200, 500, 150, 300, 400, 50, 75, 125}

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
		Calories:        float64(RoundCaloriesUp(scaleValue(cachedItem.OriginalCalories))),
		Protein:         scaleValue(cachedItem.OriginalProteinG),
		TotalFat:        scaleValue(cachedItem.OriginalTotalFatG),
		SaturatedFat:    scaleValue(cachedItem.OriginalSaturatedFatG),
		TransFat:        scaleValue(cachedItem.OriginalTransFatG),
		Cholesterol:     scaleValue(cachedItem.OriginalCholesterolMg),
		Sodium:          scaleValue(cachedItem.OriginalSodiumMg),
		TotalCarbs:      scaleValue(cachedItem.OriginalTotalCarbsG),
		DietaryFiber:    scaleValue(cachedItem.OriginalDietaryFiberG),
		TotalSugars:     scaleValue(cachedItem.OriginalTotalSugarsG),
		AddedSugars:     scaleValue(cachedItem.OriginalAddedSugarsG),
		VitaminA:        scaleValue(cachedItem.OriginalVitaminAMcg),
		VitaminC:        scaleValue(cachedItem.OriginalVitaminCMg),
		VitaminD:        scaleValue(cachedItem.OriginalVitaminDMcg),
		VitaminE:        scaleValue(cachedItem.OriginalVitaminEMg),
		VitaminK:        scaleValue(cachedItem.OriginalVitaminKMcg),
		Thiamine:        scaleValue(cachedItem.OriginalThiamineMg),
		Riboflavin:      scaleValue(cachedItem.OriginalRiboflavinMg),
		Niacin:          scaleValue(cachedItem.OriginalNiacinMg),
		VitaminB6:       scaleValue(cachedItem.OriginalVitaminB6Mg),
		Folate:          scaleValue(cachedItem.OriginalFolateMcg),
		VitaminB12:      scaleValue(cachedItem.OriginalVitaminB12Mcg),
		Biotin:          scaleValue(cachedItem.OriginalBiotinMcg),
		PantothenicAcid: scaleValue(cachedItem.OriginalPantothenicAcidMg),
		Choline:         scaleValue(cachedItem.OriginalCholineMg),
		Calcium:         scaleValue(cachedItem.OriginalCalciumMg),
		Iron:            scaleValue(cachedItem.OriginalIronMg),
		Magnesium:       scaleValue(cachedItem.OriginalMagnesiumMg),
		Phosphorus:      scaleValue(cachedItem.OriginalPhosphorusMg),
		Potassium:       scaleValue(cachedItem.OriginalPotassiumMg),
		Zinc:            scaleValue(cachedItem.OriginalZincMg),
		Copper:          scaleValue(cachedItem.OriginalCopperMg),
		Manganese:       scaleValue(cachedItem.OriginalManganeseMg),
		Selenium:        scaleValue(cachedItem.OriginalSeleniumMcg),
		Iodine:          scaleValue(cachedItem.OriginalIodineMcg),
		Molybdenum:      scaleValue(cachedItem.OriginalMolybdenumMcg),
		Chromium:        scaleValue(cachedItem.OriginalChromiumMcg),
		Fluoride:        scaleValue(cachedItem.OriginalFluorideMg),
		Chloride:        scaleValue(cachedItem.OriginalChlorideMg),
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
			Calories:        floatValue(cached.OriginalCalories),
			Protein:         floatValue(cached.OriginalProteinG),
			TotalFat:        floatValue(cached.OriginalTotalFatG),
			SaturatedFat:    floatValue(cached.OriginalSaturatedFatG),
			TransFat:        floatValue(cached.OriginalTransFatG),
			Cholesterol:     floatValue(cached.OriginalCholesterolMg),
			Sodium:          floatValue(cached.OriginalSodiumMg),
			TotalCarbs:      floatValue(cached.OriginalTotalCarbsG),
			DietaryFiber:    floatValue(cached.OriginalDietaryFiberG),
			TotalSugars:     floatValue(cached.OriginalTotalSugarsG),
			AddedSugars:     floatValue(cached.OriginalAddedSugarsG),
			VitaminA:        floatValue(cached.OriginalVitaminAMcg),
			VitaminC:        floatValue(cached.OriginalVitaminCMg),
			VitaminD:        floatValue(cached.OriginalVitaminDMcg),
			VitaminE:        floatValue(cached.OriginalVitaminEMg),
			VitaminK:        floatValue(cached.OriginalVitaminKMcg),
			Thiamine:        floatValue(cached.OriginalThiamineMg),
			Riboflavin:      floatValue(cached.OriginalRiboflavinMg),
			Niacin:          floatValue(cached.OriginalNiacinMg),
			VitaminB6:       floatValue(cached.OriginalVitaminB6Mg),
			Folate:          floatValue(cached.OriginalFolateMcg),
			VitaminB12:      floatValue(cached.OriginalVitaminB12Mcg),
			Biotin:          floatValue(cached.OriginalBiotinMcg),
			PantothenicAcid: floatValue(cached.OriginalPantothenicAcidMg),
			Choline:         floatValue(cached.OriginalCholineMg),
			Calcium:         floatValue(cached.OriginalCalciumMg),
			Iron:            floatValue(cached.OriginalIronMg),
			Magnesium:       floatValue(cached.OriginalMagnesiumMg),
			Phosphorus:      floatValue(cached.OriginalPhosphorusMg),
			Potassium:       floatValue(cached.OriginalPotassiumMg),
			Zinc:            floatValue(cached.OriginalZincMg),
			Copper:          floatValue(cached.OriginalCopperMg),
			Manganese:       floatValue(cached.OriginalManganeseMg),
			Selenium:        floatValue(cached.OriginalSeleniumMcg),
			Iodine:          floatValue(cached.OriginalIodineMcg),
			Molybdenum:      floatValue(cached.OriginalMolybdenumMcg),
			Chromium:        floatValue(cached.OriginalChromiumMcg),
			Fluoride:        floatValue(cached.OriginalFluorideMg),
			Chloride:        floatValue(cached.OriginalChlorideMg),
		}
	}

	// Fallback to per-100g data if original data is not available (shouldn't happen with new schema)
	LogWarn("Original serving data not found, falling back to per-100g conversion",
		"normalized_name", cached.NormalizedName)
	return CompleteNutrient{
		Calories:        cached.CaloriesPer100g,
		Protein:         cached.ProteinGPer100g,
		TotalFat:        cached.TotalFatGPer100g,
		SaturatedFat:    cached.SaturatedFatGPer100g,
		TransFat:        cached.TransFatGPer100g,
		Cholesterol:     cached.CholesterolMgPer100g,
		Sodium:          cached.SodiumMgPer100g,
		TotalCarbs:      cached.TotalCarbsGPer100g,
		DietaryFiber:    cached.DietaryFiberGPer100g,
		TotalSugars:     cached.TotalSugarsGPer100g,
		AddedSugars:     cached.AddedSugarsGPer100g,
		VitaminA:        cached.VitaminAMcgPer100g,
		VitaminC:        cached.VitaminCMgPer100g,
		VitaminD:        cached.VitaminDMcgPer100g,
		VitaminE:        cached.VitaminEMgPer100g,
		VitaminK:        cached.VitaminKMcgPer100g,
		Thiamine:        cached.ThiamineMgPer100g,
		Riboflavin:      cached.RiboflavinMgPer100g,
		Niacin:          cached.NiacinMgPer100g,
		VitaminB6:       cached.VitaminB6MgPer100g,
		Folate:          cached.FolateMcgPer100g,
		VitaminB12:      cached.VitaminB12McgPer100g,
		Biotin:          cached.BiotinMcgPer100g,
		PantothenicAcid: cached.PantothenicAcidMgPer100g,
		Choline:         cached.CholineMgPer100g,
		Calcium:         cached.CalciumMgPer100g,
		Iron:            cached.IronMgPer100g,
		Magnesium:       cached.MagnesiumMgPer100g,
		Phosphorus:      cached.PhosphorusMgPer100g,
		Potassium:       cached.PotassiumMgPer100g,
		Zinc:            cached.ZincMgPer100g,
		Copper:          cached.CopperMgPer100g,
		Manganese:       cached.ManganeseMgPer100g,
		Selenium:        cached.SeleniumMcgPer100g,
		Iodine:          cached.IodineMcgPer100g,
		Molybdenum:      cached.MolybdenumMcgPer100g,
		Chromium:        cached.ChromiumMcgPer100g,
		Fluoride:        cached.FluorideMgPer100g,
		Chloride:        cached.ChlorideMgPer100g,
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

		normalizedNutrients = CompleteNutrient{
			Calories:        nutrients.Calories / divider,
			Protein:         nutrients.Protein / divider,
			TotalFat:        nutrients.TotalFat / divider,
			SaturatedFat:    nutrients.SaturatedFat / divider,
			TransFat:        nutrients.TransFat / divider,
			Cholesterol:     nutrients.Cholesterol / divider,
			Sodium:          nutrients.Sodium / divider,
			TotalCarbs:      nutrients.TotalCarbs / divider,
			DietaryFiber:    nutrients.DietaryFiber / divider,
			TotalSugars:     nutrients.TotalSugars / divider,
			AddedSugars:     nutrients.AddedSugars / divider,
			VitaminA:        nutrients.VitaminA / divider,
			VitaminC:        nutrients.VitaminC / divider,
			VitaminD:        nutrients.VitaminD / divider,
			VitaminE:        nutrients.VitaminE / divider,
			VitaminK:        nutrients.VitaminK / divider,
			Thiamine:        nutrients.Thiamine / divider,
			Riboflavin:      nutrients.Riboflavin / divider,
			Niacin:          nutrients.Niacin / divider,
			VitaminB6:       nutrients.VitaminB6 / divider,
			Folate:          nutrients.Folate / divider,
			VitaminB12:      nutrients.VitaminB12 / divider,
			Biotin:          nutrients.Biotin / divider,
			PantothenicAcid: nutrients.PantothenicAcid / divider,
			Choline:         nutrients.Choline / divider,
			Calcium:         nutrients.Calcium / divider,
			Iron:            nutrients.Iron / divider,
			Magnesium:       nutrients.Magnesium / divider,
			Phosphorus:      nutrients.Phosphorus / divider,
			Potassium:       nutrients.Potassium / divider,
			Zinc:            nutrients.Zinc / divider,
			Copper:          nutrients.Copper / divider,
			Manganese:       nutrients.Manganese / divider,
			Selenium:        nutrients.Selenium / divider,
			Iodine:          nutrients.Iodine / divider,
			Molybdenum:      nutrients.Molybdenum / divider,
			Chromium:        nutrients.Chromium / divider,
			Fluoride:        nutrients.Fluoride / divider,
			Chloride:        nutrients.Chloride / divider,
		}
	}

	// Helper function to convert and round in one step for per-100g values
	convertAndRound := func(servingValue float64, decimalPlaces int) float64 {
		return RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(servingValue, normalizedGrams), decimalPlaces)
	}

	// Helper function to create pointer to float64
	float64Ptr := func(val float64) *float64 { return &val }

	return &storage.Item{
		NormalizedName:  exactKey, // Use the special key as the normalized name
		NormalizedBrand: "",       // Empty brand for exact matches
		DisplayName:     item.Name,
		DisplayBrand:    getBrandOrEmpty(item.Brand),

		// Store normalized single-unit serving data
		OriginalServingGrams:      float64Ptr(normalizedGrams),
		OriginalCalories:          float64Ptr(normalizedNutrients.Calories),
		OriginalProteinG:          float64Ptr(normalizedNutrients.Protein),
		OriginalTotalFatG:         float64Ptr(normalizedNutrients.TotalFat),
		OriginalSaturatedFatG:     float64Ptr(normalizedNutrients.SaturatedFat),
		OriginalTransFatG:         float64Ptr(normalizedNutrients.TransFat),
		OriginalCholesterolMg:     float64Ptr(normalizedNutrients.Cholesterol),
		OriginalSodiumMg:          float64Ptr(normalizedNutrients.Sodium),
		OriginalTotalCarbsG:       float64Ptr(normalizedNutrients.TotalCarbs),
		OriginalDietaryFiberG:     float64Ptr(normalizedNutrients.DietaryFiber),
		OriginalTotalSugarsG:      float64Ptr(normalizedNutrients.TotalSugars),
		OriginalAddedSugarsG:      float64Ptr(normalizedNutrients.AddedSugars),
		OriginalVitaminAMcg:       float64Ptr(normalizedNutrients.VitaminA),
		OriginalVitaminCMg:        float64Ptr(normalizedNutrients.VitaminC),
		OriginalVitaminDMcg:       float64Ptr(normalizedNutrients.VitaminD),
		OriginalVitaminEMg:        float64Ptr(normalizedNutrients.VitaminE),
		OriginalVitaminKMcg:       float64Ptr(normalizedNutrients.VitaminK),
		OriginalThiamineMg:        float64Ptr(normalizedNutrients.Thiamine),
		OriginalRiboflavinMg:      float64Ptr(normalizedNutrients.Riboflavin),
		OriginalNiacinMg:          float64Ptr(normalizedNutrients.Niacin),
		OriginalVitaminB6Mg:       float64Ptr(normalizedNutrients.VitaminB6),
		OriginalFolateMcg:         float64Ptr(normalizedNutrients.Folate),
		OriginalVitaminB12Mcg:     float64Ptr(normalizedNutrients.VitaminB12),
		OriginalBiotinMcg:         float64Ptr(normalizedNutrients.Biotin),
		OriginalPantothenicAcidMg: float64Ptr(normalizedNutrients.PantothenicAcid),
		OriginalCholineMg:         float64Ptr(normalizedNutrients.Choline),
		OriginalCalciumMg:         float64Ptr(normalizedNutrients.Calcium),
		OriginalIronMg:            float64Ptr(normalizedNutrients.Iron),
		OriginalMagnesiumMg:       float64Ptr(normalizedNutrients.Magnesium),
		OriginalPhosphorusMg:      float64Ptr(normalizedNutrients.Phosphorus),
		OriginalPotassiumMg:       float64Ptr(normalizedNutrients.Potassium),
		OriginalZincMg:            float64Ptr(normalizedNutrients.Zinc),
		OriginalCopperMg:          float64Ptr(normalizedNutrients.Copper),
		OriginalManganeseMg:       float64Ptr(normalizedNutrients.Manganese),
		OriginalSeleniumMcg:       float64Ptr(normalizedNutrients.Selenium),
		OriginalIodineMcg:         float64Ptr(normalizedNutrients.Iodine),
		OriginalMolybdenumMcg:     float64Ptr(normalizedNutrients.Molybdenum),
		OriginalChromiumMcg:       float64Ptr(normalizedNutrients.Chromium),
		OriginalFluorideMg:        float64Ptr(normalizedNutrients.Fluoride),
		OriginalChlorideMg:        float64Ptr(normalizedNutrients.Chloride),

		// Also store per-100g data for scaling to other serving sizes
		CaloriesPer100g:          convertAndRound(normalizedNutrients.Calories, 4),
		ProteinGPer100g:          convertAndRound(normalizedNutrients.Protein, 3),
		TotalFatGPer100g:         convertAndRound(normalizedNutrients.TotalFat, 3),
		SaturatedFatGPer100g:     convertAndRound(normalizedNutrients.SaturatedFat, 3),
		TransFatGPer100g:         convertAndRound(normalizedNutrients.TransFat, 3),
		CholesterolMgPer100g:     convertAndRound(normalizedNutrients.Cholesterol, 2),
		SodiumMgPer100g:          convertAndRound(normalizedNutrients.Sodium, 1),
		TotalCarbsGPer100g:       convertAndRound(normalizedNutrients.TotalCarbs, 1),
		DietaryFiberGPer100g:     convertAndRound(normalizedNutrients.DietaryFiber, 1),
		TotalSugarsGPer100g:      convertAndRound(normalizedNutrients.TotalSugars, 1),
		AddedSugarsGPer100g:      convertAndRound(normalizedNutrients.AddedSugars, 1),
		VitaminAMcgPer100g:       convertAndRound(normalizedNutrients.VitaminA, 1),
		VitaminCMgPer100g:        convertAndRound(normalizedNutrients.VitaminC, 1),
		VitaminDMcgPer100g:       convertAndRound(normalizedNutrients.VitaminD, 1),
		VitaminEMgPer100g:        convertAndRound(normalizedNutrients.VitaminE, 1),
		VitaminKMcgPer100g:       convertAndRound(normalizedNutrients.VitaminK, 1),
		ThiamineMgPer100g:        convertAndRound(normalizedNutrients.Thiamine, 3),
		RiboflavinMgPer100g:      convertAndRound(normalizedNutrients.Riboflavin, 3),
		NiacinMgPer100g:          convertAndRound(normalizedNutrients.Niacin, 1),
		VitaminB6MgPer100g:       convertAndRound(normalizedNutrients.VitaminB6, 3),
		FolateMcgPer100g:         convertAndRound(normalizedNutrients.Folate, 1),
		VitaminB12McgPer100g:     convertAndRound(normalizedNutrients.VitaminB12, 2),
		BiotinMcgPer100g:         convertAndRound(normalizedNutrients.Biotin, 1),
		PantothenicAcidMgPer100g: convertAndRound(normalizedNutrients.PantothenicAcid, 1),
		CholineMgPer100g:         convertAndRound(normalizedNutrients.Choline, 1),
		CalciumMgPer100g:         convertAndRound(normalizedNutrients.Calcium, 1),
		IronMgPer100g:            convertAndRound(normalizedNutrients.Iron, 1),
		MagnesiumMgPer100g:       convertAndRound(normalizedNutrients.Magnesium, 1),
		PhosphorusMgPer100g:      convertAndRound(normalizedNutrients.Phosphorus, 1),
		PotassiumMgPer100g:       convertAndRound(normalizedNutrients.Potassium, 1),
		ZincMgPer100g:            convertAndRound(normalizedNutrients.Zinc, 2),
		CopperMgPer100g:          convertAndRound(normalizedNutrients.Copper, 3),
		ManganeseMgPer100g:       convertAndRound(normalizedNutrients.Manganese, 3),
		SeleniumMcgPer100g:       convertAndRound(normalizedNutrients.Selenium, 1),
		IodineMcgPer100g:         convertAndRound(normalizedNutrients.Iodine, 1),
		MolybdenumMcgPer100g:     convertAndRound(normalizedNutrients.Molybdenum, 1),
		ChromiumMcgPer100g:       convertAndRound(normalizedNutrients.Chromium, 1),
		FluorideMgPer100g:        convertAndRound(normalizedNutrients.Fluoride, 1),
		ChlorideMgPer100g:        convertAndRound(normalizedNutrients.Chloride, 1),

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
