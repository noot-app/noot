package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// GetFavorites implements ServerInterface.GetFavorites
func (s *APIServer) GetFavorites(c *gin.Context) {
	if s.store == nil {
		handleStorageUnavailableError(c)
		return
	}

	requestID := c.GetString("request_id")

	// Get current user
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	ctx := c.Request.Context()

	// Get user's favorites
	favorites, err := s.store.GetUserFavorites(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to get favorites", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API format
	apiFavorites := make([]api.FavoriteWithConsumption, len(favorites))
	for i, fav := range favorites {
		// Convert consumption to API format
		apiConsumption := consumptionToAPI(fav.Consumption)

		apiFavorites[i] = api.FavoriteWithConsumption{
			Id:            fav.ID,
			UserId:        fav.UserID,
			ConsumptionId: fav.ConsumptionID,
			CreatedAt:     fav.CreatedAt,
			Consumption:   apiConsumption,
		}
	}

	response := api.FavoritesResponse{
		Favorites: apiFavorites,
	}

	LogInfo("Retrieved user favorites", "user_id", user.ID, "count", len(favorites), "request_id", requestID)
	c.JSON(http.StatusOK, response)
}

// AddToFavorites implements ServerInterface.AddToFavorites
func (s *APIServer) AddToFavorites(c *gin.Context) {
	if s.store == nil {
		handleStorageUnavailableError(c)
		return
	}

	requestID := c.GetString("request_id")

	// Get current user
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Parse request body
	var req api.AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	ctx := c.Request.Context()

	// Verify the consumption exists and belongs to the user
	consumption, err := s.store.GetConsumptionForUser(ctx, user.ID, req.ConsumptionId)
	if err != nil {
		appErr := NewAppError("Failed to verify consumption", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if consumption == nil {
		appErr := NewAppError("Consumption not found or not owned by user", http.StatusNotFound, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Check if already favorited
	isFavorited, err := s.store.IsFavorited(ctx, user.ID, req.ConsumptionId)
	if err != nil {
		appErr := NewAppError("Failed to check favorite status", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}
	if isFavorited {
		appErr := NewAppError("Consumption is already favorited", http.StatusConflict, nil)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Create the favorite
	favorite, err := s.store.CreateFavorite(ctx, user.ID, req.ConsumptionId)
	if err != nil {
		appErr := NewAppError("Failed to add to favorites", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API format
	apiFavorite := api.Favorite{
		Id:            favorite.ID,
		UserId:        favorite.UserID,
		ConsumptionId: favorite.ConsumptionID,
		CreatedAt:     favorite.CreatedAt,
	}

	LogInfo("Added consumption to favorites", "user_id", user.ID, "consumption_id", req.ConsumptionId, "request_id", requestID)
	c.JSON(http.StatusCreated, apiFavorite)
}

// RemoveFromFavorites implements ServerInterface.RemoveFromFavorites
func (s *APIServer) RemoveFromFavorites(c *gin.Context, consumptionID string) {
	if s.store == nil {
		handleStorageUnavailableError(c)
		return
	}

	requestID := c.GetString("request_id")

	// Get current user
	user, err := getCurrentUser(c)
	if err != nil {
		appErr := NewAppError("Authentication required", http.StatusUnauthorized, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	ctx := c.Request.Context()

	// Remove the favorite
	err = s.store.DeleteFavorite(ctx, user.ID, consumptionID)
	if err != nil {
		if err.Error() == "favorite not found" {
			appErr := NewAppError("Favorite not found", http.StatusNotFound, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
		appErr := NewAppError("Failed to remove from favorites", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Create success response
	resp := api.DeleteResponse{
		Message: "Consumption removed from favorites successfully",
		Id:      consumptionID,
	}

	LogInfo("Removed consumption from favorites", "user_id", user.ID, "consumption_id", consumptionID, "request_id", requestID)
	c.JSON(http.StatusOK, resp)
}

// Helper function to convert storage.Consumption to api.Consumption
func consumptionToAPI(consumption *storage.Consumption) api.Consumption {
	// Create the complete nutrient totals
	totals := api.CompleteNutrient{
		Calories:            float32(consumption.TotalCalories),
		ProteinG:            float32(consumption.TotalProtein),
		TotalFatG:           float32(consumption.TotalFat),
		TotalCarbsG:         float32(consumption.TotalCarbs),
		DietaryFiberG:       float32(consumption.DietaryFiber),
		SodiumMg:            float32(consumption.TotalSodium),
		SaturatedFatG:       float32(consumption.SaturatedFat),
		TransFatG:           float32(consumption.TransFat),
		CholesterolMg:       float32(consumption.Cholesterol),
		TotalSugarsG:        float32(consumption.TotalSugars),
		AddedSugarsG:        float32(consumption.AddedSugars),
		VitaminAMcg:         float32(consumption.VitaminA),
		VitaminCMg:          float32(consumption.VitaminC),
		VitaminDMcg:         float32(consumption.VitaminD),
		VitaminEMg:          float32(consumption.VitaminE),
		VitaminKMcg:         float32(consumption.VitaminK),
		ThiamineMg:          float32(consumption.Thiamine),
		RiboflavinMg:        float32(consumption.Riboflavin),
		NiacinMg:            float32(consumption.Niacin),
		VitaminB6Mg:         float32(consumption.VitaminB6),
		FolateMcg:           float32(consumption.Folate),
		VitaminB12Mcg:       float32(consumption.VitaminB12),
		BiotinMcg:           float32(consumption.Biotin),
		PantothenicAcidMg:   float32(consumption.PantothenicAcid),
		CholineMg:           float32(consumption.Choline),
		CalciumMg:           float32(consumption.Calcium),
		IronMg:              float32(consumption.Iron),
		MagnesiumMg:         float32(consumption.Magnesium),
		PhosphorusMg:        float32(consumption.Phosphorus),
		PotassiumMg:         float32(consumption.Potassium),
		ZincMg:              float32(consumption.Zinc),
		CopperMg:            float32(consumption.Copper),
		ManganeseMg:         float32(consumption.Manganese),
		SeleniumMcg:         float32(consumption.Selenium),
		IodineMcg:           float32(consumption.Iodine),
		MolybdenumMcg:       float32(consumption.Molybdenum),
		ChromiumMcg:         float32(consumption.Chromium),
		FluorideMg:          float32(consumption.Fluoride),
		ChlorideMg:          float32(consumption.Chloride),
		Omega3AlaG:          float32(consumption.Omega3Ala),
		Omega3EpaG:          float32(consumption.Omega3Epa),
		Omega3DhaG:          float32(consumption.Omega3Dha),
		Omega6G:             float32(consumption.Omega6),
		CreatineMg:          float32(consumption.Creatine),
		CaffeineMg:          float32(consumption.Caffeine),
		AlcoholG:            float32(consumption.Alcohol),
		PolyunsaturatedFatG: float32(consumption.PolyunsaturatedFat),
		MonounsaturatedFatG: float32(consumption.MonounsaturatedFat),
	}

	// Create the summary with totals and empty daily values/percentages for now
	summary := api.Summary{
		Totals:         totals,
		DailyValues:    make(map[string]float32),
		PercentOfDaily: make(map[string]int),
	}

	apiConsumption := api.Consumption{
		Id:         consumption.ID,
		UserId:     consumption.UserID,
		Transcript: consumption.Transcript,
		Note:       consumption.Note,
		Summary:    summary,
		CreatedAt:  consumption.CreatedAt,
		Items:      []api.ItemWithNutrition{}, // Empty items for now
	}

	// Convert labels if present
	if consumption.Labels != nil {
		apiLabels := make([]api.Label, len(consumption.Labels))
		for i, label := range consumption.Labels {
			apiLabels[i] = api.Label{
				Id:          label.ID,
				Name:        label.Name,
				Description: label.Description,
				Color:       label.Color,
				CreatedAt:   label.CreatedAt,
			}
		}
		apiConsumption.Labels = &apiLabels
	}

	return apiConsumption
}
