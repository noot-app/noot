package server

import (
	"context"
	"os"
	"time"

	"github.com/grantbirki/noot/internal/storage"
)

// NutritionService handles nutrition data processing with caching and unit conversions
type NutritionService struct {
	aiProvider AIProvider
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

	LogDebug("Checking cache for item", "normalized_name", normalizedName, "normalized_brand", normalizedBrand)

	// Check cache first
	if s.store != nil {
		cached, err := s.store.GetItemFromCache(ctx, normalizedName, normalizedBrand)
		if err == nil && cached != nil {
			if !s.store.IsItemCacheExpired(cached) {
				LogDebug("Using cached nutrition data - scaling per-100g data to serving size",
					"name", item.Name, "quantity", item.Quantity, "unit", item.Unit)

				nutrition := s.convertCachedToNutrients(cached, item)
				item.Nutrients = &nutrition
				return item, nil
			} else {
				LogDebug("Cache expired for item", "name", item.Name, "expires_at", cached.ExpiresAt)
			}
		}
	}

	// Not in cache or expired/invalid - get from AI
	LogDebug("Fetching nutrition from AI provider", "name", item.Name)
	nutrition, err := s.aiProvider.GetNutrition(ctx, item)
	if err != nil {
		return item, err
	}

	// Cache the result if store is available
	if s.store != nil {
		cacheItem := s.convertNutrientsToCache(item, nutrition)
		if cached, _ := s.store.GetItemFromCache(ctx, normalizedName, normalizedBrand); cached != nil {
			// Refresh existing cache entry
			err = s.store.RefreshItemCache(ctx, normalizedName, normalizedBrand, cacheItem)
		} else {
			// Create new cache entry
			err = s.store.UpsertItemCache(ctx, cacheItem)
		}
		if err != nil {
			LogWarn("Failed to cache nutrition data", "name", item.Name, "error", err.Error())
			// Don't fail the request if caching fails
		}
	}

	item.Nutrients = &nutrition
	return item, nil
}

// convertCachedToNutrients converts cached per-100g data to actual serving size
func (s *NutritionService) convertCachedToNutrients(cached *storage.ItemCache, item Item) CompleteNutrient {
	// Estimate serving weight for scaling
	servingGrams, err := s.converter.EstimateServingWeight(item)
	if err != nil {
		LogWarn("Failed to estimate serving weight for cached item", "item", item.Name, "error", err.Error())
		servingGrams = 100.0 // Default to 100g
	}

	// Convert from per-100g cache data to actual serving size
	return CompleteNutrient{
		Calories:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CaloriesPer100g, servingGrams), 1),
		Protein:      RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ProteinGPer100g, servingGrams), 1),
		TotalFat:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TotalFatGPer100g, servingGrams), 1),
		SaturatedFat: RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.SaturatedFatGPer100g, servingGrams), 1),
		TransFat:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TransFatGPer100g, servingGrams), 1),
		Cholesterol:  RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CholesterolMgPer100g, servingGrams), 1),
		Sodium:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.SodiumMgPer100g, servingGrams), 1),
		TotalCarbs:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TotalCarbsGPer100g, servingGrams), 1),
		DietaryFiber: RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.DietaryFiberGPer100g, servingGrams), 1),
		TotalSugars:  RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TotalSugarsGPer100g, servingGrams), 1),
		AddedSugars:  RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.AddedSugarsGPer100g, servingGrams), 1),
		VitaminA:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminAMcgPer100g, servingGrams), 1),
		VitaminC:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminCMgPer100g, servingGrams), 1),
		VitaminD:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminDMcgPer100g, servingGrams), 1),
		VitaminE:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminEMgPer100g, servingGrams), 1),
		VitaminK:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminKMcgPer100g, servingGrams), 1),
		Thiamine:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ThiamineMgPer100g, servingGrams), 3),
		Riboflavin:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.RiboflavinMgPer100g, servingGrams), 3),
		Niacin:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.NiacinMgPer100g, servingGrams), 1),
		VitaminB6:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminB6MgPer100g, servingGrams), 3),
		Folate:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.FolateMcgPer100g, servingGrams), 1),
		VitaminB12:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminB12McgPer100g, servingGrams), 2),
		Calcium:      RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CalciumMgPer100g, servingGrams), 1),
		Iron:         RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.IronMgPer100g, servingGrams), 1),
		Magnesium:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.MagnesiumMgPer100g, servingGrams), 1),
		Phosphorus:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.PhosphorusMgPer100g, servingGrams), 1),
		Potassium:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.PotassiumMgPer100g, servingGrams), 1),
		Zinc:         RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ZincMgPer100g, servingGrams), 2),
		Copper:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CopperMgPer100g, servingGrams), 3),
		Manganese:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ManganeseMgPer100g, servingGrams), 3),
		Selenium:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.SeleniumMcgPer100g, servingGrams), 1),
	}
}

// convertNutrientsToCache converts serving size nutrition data to per-100g for cache storage
// The LLM provides nutrition for the exact serving, so we need to normalize to per-100g
func (s *NutritionService) convertNutrientsToCache(item Item, nutrients CompleteNutrient) *storage.ItemCache {
	normalizedName := normalizeItemName(item.Name)
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))

	// The key insight: since the LLM already provided nutrition for the exact serving size,
	// we need to estimate what that serving weighs to convert to per-100g for caching
	servingGrams, err := s.converter.EstimateServingWeight(item)
	if err != nil {
		LogWarn("Failed to estimate serving weight for cache storage", "item", item.Name, "error", err.Error())
		// If we can't estimate the weight, we'll store as-is and assume 100g equivalent
		// This isn't perfect but prevents cache failures
		servingGrams = 100.0
	}

	// Convert from serving size nutrition data to per-100g for consistent cache storage
	return &storage.ItemCache{
		NormalizedName:       normalizedName,
		NormalizedBrand:      normalizedBrand,
		DisplayName:          item.Name,
		DisplayBrand:         getBrandOrEmpty(item.Brand),
		CaloriesPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Calories, servingGrams), 1),
		ProteinGPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Protein, servingGrams), 1),
		TotalFatGPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TotalFat, servingGrams), 1),
		SaturatedFatGPer100g: RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.SaturatedFat, servingGrams), 1),
		TransFatGPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TransFat, servingGrams), 1),
		CholesterolMgPer100g: RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Cholesterol, servingGrams), 1),
		SodiumMgPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Sodium, servingGrams), 1),
		TotalCarbsGPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TotalCarbs, servingGrams), 1),
		DietaryFiberGPer100g: RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.DietaryFiber, servingGrams), 1),
		TotalSugarsGPer100g:  RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TotalSugars, servingGrams), 1),
		AddedSugarsGPer100g:  RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.AddedSugars, servingGrams), 1),
		VitaminAMcgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminA, servingGrams), 1),
		VitaminCMgPer100g:    RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminC, servingGrams), 1),
		VitaminDMcgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminD, servingGrams), 1),
		VitaminEMgPer100g:    RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminE, servingGrams), 1),
		VitaminKMcgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminK, servingGrams), 1),
		ThiamineMgPer100g:    RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Thiamine, servingGrams), 3),
		RiboflavinMgPer100g:  RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Riboflavin, servingGrams), 3),
		NiacinMgPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Niacin, servingGrams), 1),
		VitaminB6MgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminB6, servingGrams), 3),
		FolateMcgPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Folate, servingGrams), 1),
		VitaminB12McgPer100g: RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminB12, servingGrams), 2),
		CalciumMgPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Calcium, servingGrams), 1),
		IronMgPer100g:        RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Iron, servingGrams), 1),
		MagnesiumMgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Magnesium, servingGrams), 1),
		PhosphorusMgPer100g:  RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Phosphorus, servingGrams), 1),
		PotassiumMgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Potassium, servingGrams), 1),
		ZincMgPer100g:        RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Zinc, servingGrams), 2),
		CopperMgPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Copper, servingGrams), 3),
		ManganeseMgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Manganese, servingGrams), 3),
		SeleniumMcgPer100g:   RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Selenium, servingGrams), 1),
		FetchedAt:            time.Now().UTC(),
		ExpiresAt:            time.Now().UTC().Add(30 * 24 * time.Hour), // 30 day TTL
	}
}
