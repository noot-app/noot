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
		cached, err := s.store.GetItemByName(ctx, normalizedName, normalizedBrand)
		if err == nil && cached != nil {
			// Check if cache is still fresh (30 days)
			if time.Since(cached.UpdatedAt) < 30*24*time.Hour {
				LogDebug("Using cached nutrition data - scaling per-100g data to actual weight",
					"name", item.Name, "grams", item.Grams)

				nutrition := s.convertCachedToNutrients(cached, item)
				item.Nutrients = &nutrition

				return item, nil
			} else {
				LogDebug("Cache expired for item", "name", item.Name, "updated_at", cached.UpdatedAt)
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
		if cached, _ := s.store.GetItemByName(ctx, normalizedName, normalizedBrand); cached != nil {
			// Update existing cache entry
			cacheItem.ID = cached.ID
			err = s.store.UpdateItem(ctx, cacheItem)
		} else {
			// Create new cache entry
			err = s.store.CreateItem(ctx, cacheItem)
		}
		if err != nil {
			LogWarn("Failed to cache nutrition data", "name", item.Name, "error", err.Error())
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

	// Convert from per-100g cache data to actual weight
	return CompleteNutrient{
		Calories:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CaloriesPer100g, actualGrams), 1),
		Protein:      RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ProteinGPer100g, actualGrams), 1),
		TotalFat:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TotalFatGPer100g, actualGrams), 1),
		SaturatedFat: RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.SaturatedFatGPer100g, actualGrams), 1),
		TransFat:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TransFatGPer100g, actualGrams), 1),
		Cholesterol:  RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CholesterolMgPer100g, actualGrams), 1),
		Sodium:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.SodiumMgPer100g, actualGrams), 1),
		TotalCarbs:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TotalCarbsGPer100g, actualGrams), 1),
		DietaryFiber: RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.DietaryFiberGPer100g, actualGrams), 1),
		TotalSugars:  RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.TotalSugarsGPer100g, actualGrams), 1),
		AddedSugars:  RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.AddedSugarsGPer100g, actualGrams), 1),
		VitaminA:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminAMcgPer100g, actualGrams), 1),
		VitaminC:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminCMgPer100g, actualGrams), 1),
		VitaminD:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminDMcgPer100g, actualGrams), 1),
		VitaminE:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminEMgPer100g, actualGrams), 1),
		VitaminK:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminKMcgPer100g, actualGrams), 1),
		Thiamine:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ThiamineMgPer100g, actualGrams), 3),
		Riboflavin:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.RiboflavinMgPer100g, actualGrams), 3),
		Niacin:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.NiacinMgPer100g, actualGrams), 1),
		VitaminB6:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminB6MgPer100g, actualGrams), 3),
		Folate:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.FolateMcgPer100g, actualGrams), 1),
		VitaminB12:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.VitaminB12McgPer100g, actualGrams), 2),
		Calcium:      RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CalciumMgPer100g, actualGrams), 1),
		Iron:         RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.IronMgPer100g, actualGrams), 1),
		Magnesium:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.MagnesiumMgPer100g, actualGrams), 1),
		Phosphorus:   RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.PhosphorusMgPer100g, actualGrams), 1),
		Potassium:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.PotassiumMgPer100g, actualGrams), 1),
		Zinc:         RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ZincMgPer100g, actualGrams), 2),
		Copper:       RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.CopperMgPer100g, actualGrams), 3),
		Manganese:    RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.ManganeseMgPer100g, actualGrams), 3),
		Selenium:     RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(cached.SeleniumMcgPer100g, actualGrams), 1),
	}
}

// convertNutrientsToCache converts actual weight nutrition data to per-100g for cache storage
// The LLM provides nutrition for the exact weight in grams, so we need to normalize to per-100g
func (s *NutritionService) convertNutrientsToCache(item Item, nutrients CompleteNutrient) *storage.Item {
	normalizedName := normalizeItemName(item.Name)
	normalizedBrand := normalizeItemName(getBrandOrEmpty(item.Brand))

	// Use the actual grams from the item since LLM already provided nutrition for that exact weight
	actualGrams := item.Grams

	// Convert from actual weight nutrition data to per-100g for consistent cache storage
	return &storage.Item{
		NormalizedName:           normalizedName,
		NormalizedBrand:          normalizedBrand,
		DisplayName:              item.Name,
		DisplayBrand:             getBrandOrEmpty(item.Brand),
		CaloriesPer100g:          RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Calories, actualGrams), 1),
		ProteinGPer100g:          RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Protein, actualGrams), 1),
		TotalFatGPer100g:         RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TotalFat, actualGrams), 1),
		SaturatedFatGPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.SaturatedFat, actualGrams), 1),
		TransFatGPer100g:         RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TransFat, actualGrams), 1),
		CholesterolMgPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Cholesterol, actualGrams), 1),
		SodiumMgPer100g:          RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Sodium, actualGrams), 1),
		TotalCarbsGPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TotalCarbs, actualGrams), 1),
		DietaryFiberGPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.DietaryFiber, actualGrams), 1),
		TotalSugarsGPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.TotalSugars, actualGrams), 1),
		AddedSugarsGPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.AddedSugars, actualGrams), 1),
		VitaminAMcgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminA, actualGrams), 1),
		VitaminCMgPer100g:        RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminC, actualGrams), 1),
		VitaminDMcgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminD, actualGrams), 1),
		VitaminEMgPer100g:        RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminE, actualGrams), 1),
		VitaminKMcgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminK, actualGrams), 1),
		ThiamineMgPer100g:        RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Thiamine, actualGrams), 3),
		RiboflavinMgPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Riboflavin, actualGrams), 3),
		NiacinMgPer100g:          RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Niacin, actualGrams), 1),
		VitaminB6MgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminB6, actualGrams), 3),
		FolateMcgPer100g:         RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Folate, actualGrams), 1),
		VitaminB12McgPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.VitaminB12, actualGrams), 2),
		BiotinMcgPer100g:         RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add biotin to CompleteNutrient
		PantothenicAcidMgPer100g: RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add pantothenic acid to CompleteNutrient
		CholineMgPer100g:         RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add choline to CompleteNutrient
		CalciumMgPer100g:         RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Calcium, actualGrams), 1),
		IronMgPer100g:            RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Iron, actualGrams), 1),
		MagnesiumMgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Magnesium, actualGrams), 1),
		PhosphorusMgPer100g:      RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Phosphorus, actualGrams), 1),
		PotassiumMgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Potassium, actualGrams), 1),
		ZincMgPer100g:            RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Zinc, actualGrams), 2),
		CopperMgPer100g:          RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Copper, actualGrams), 3),
		ManganeseMgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Manganese, actualGrams), 3),
		SeleniumMcgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(nutrients.Selenium, actualGrams), 1),
		IodineMcgPer100g:         RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add iodine to CompleteNutrient
		MolybdenumMcgPer100g:     RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add molybdenum to CompleteNutrient
		ChromiumMcgPer100g:       RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add chromium to CompleteNutrient
		FluorideMgPer100g:        RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add fluoride to CompleteNutrient
		ChlorideMgPer100g:        RoundToDecimalPlaces(s.converter.ConvertFromServingToPer100g(0, actualGrams), 1), // TODO: Add chloride to CompleteNutrient
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
	}
}
