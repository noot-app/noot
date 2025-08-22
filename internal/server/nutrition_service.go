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

	// Helper function to convert and round in one step
	convertAndRound := func(per100gValue float64, decimalPlaces int) float64 {
		return RoundToDecimalPlaces(s.converter.ConvertFromPer100gToServing(per100gValue, actualGrams), decimalPlaces)
	}

	// Convert from per-100g cache data to actual weight
	return CompleteNutrient{
		Calories:     convertAndRound(cached.CaloriesPer100g, 1),
		Protein:      convertAndRound(cached.ProteinGPer100g, 1),
		TotalFat:     convertAndRound(cached.TotalFatGPer100g, 1),
		SaturatedFat: convertAndRound(cached.SaturatedFatGPer100g, 1),
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
		CaloriesPer100g:          convertAndRound(nutrients.Calories, 1),
		ProteinGPer100g:          convertAndRound(nutrients.Protein, 1),
		TotalFatGPer100g:         convertAndRound(nutrients.TotalFat, 1),
		SaturatedFatGPer100g:     convertAndRound(nutrients.SaturatedFat, 1),
		TransFatGPer100g:         convertAndRound(nutrients.TransFat, 1),
		CholesterolMgPer100g:     convertAndRound(nutrients.Cholesterol, 1),
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
		BiotinMcgPer100g:         convertAndRound(0, 1), // TODO: Add biotin to CompleteNutrient
		PantothenicAcidMgPer100g: convertAndRound(0, 1), // TODO: Add pantothenic acid to CompleteNutrient
		CholineMgPer100g:         convertAndRound(0, 1), // TODO: Add choline to CompleteNutrient
		CalciumMgPer100g:         convertAndRound(nutrients.Calcium, 1),
		IronMgPer100g:            convertAndRound(nutrients.Iron, 1),
		MagnesiumMgPer100g:       convertAndRound(nutrients.Magnesium, 1),
		PhosphorusMgPer100g:      convertAndRound(nutrients.Phosphorus, 1),
		PotassiumMgPer100g:       convertAndRound(nutrients.Potassium, 1),
		ZincMgPer100g:            convertAndRound(nutrients.Zinc, 2),
		CopperMgPer100g:          convertAndRound(nutrients.Copper, 3),
		ManganeseMgPer100g:       convertAndRound(nutrients.Manganese, 3),
		SeleniumMcgPer100g:       convertAndRound(nutrients.Selenium, 1),
		IodineMcgPer100g:         convertAndRound(0, 1), // TODO: Add iodine to CompleteNutrient
		MolybdenumMcgPer100g:     convertAndRound(0, 1), // TODO: Add molybdenum to CompleteNutrient
		ChromiumMcgPer100g:       convertAndRound(0, 1), // TODO: Add chromium to CompleteNutrient
		FluorideMgPer100g:        convertAndRound(0, 1), // TODO: Add fluoride to CompleteNutrient
		ChlorideMgPer100g:        convertAndRound(0, 1), // TODO: Add chloride to CompleteNutrient
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
	}
}
