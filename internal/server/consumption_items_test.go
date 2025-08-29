package server

import (
	"testing"
	"time"

	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

func TestItemWithNutritionToConsumptionItem(t *testing.T) {
	// Create a test ItemWithNutrition with nutrients
	userQuantity := float32(1.0)
	userUnit := "cup"
	brand := "test brand"
	note := "test note"

	apiItem := api.ItemWithNutrition{
		Item: api.Item{
			Name:         "test item",
			Brand:        &brand,
			Grams:        100.0,
			UserQuantity: &userQuantity,
			UserUnit:     &userUnit,
			Nutrients: &api.CompleteNutrient{
				Calories:  200.0,
				ProteinG:  10.0,
				TotalFatG: 5.0,
			},
		},
		Note: &note,
	}

	consumptionID := "test-consumption-id"
	result := itemWithNutritionToConsumptionItem(consumptionID, apiItem)

	// Validate the conversion
	if result.ConsumptionID != consumptionID {
		t.Errorf("Expected ConsumptionID %s, got %s", consumptionID, result.ConsumptionID)
	}
	if result.Name != "test item" {
		t.Errorf("Expected Name 'test item', got %s", result.Name)
	}
	if result.Brand != brand {
		t.Errorf("Expected Brand '%s', got %s", brand, result.Brand)
	}
	if result.Grams != 100.0 {
		t.Errorf("Expected Grams 100.0, got %f", result.Grams)
	}
	if result.UserQuantity == nil || *result.UserQuantity != 1.0 {
		t.Errorf("Expected UserQuantity 1.0, got %v", result.UserQuantity)
	}
	if result.UserUnit == nil || *result.UserUnit != userUnit {
		t.Errorf("Expected UserUnit '%s', got %v", userUnit, result.UserUnit)
	}
	if result.Calories != 200.0 {
		t.Errorf("Expected Calories 200.0, got %f", result.Calories)
	}
	if result.ProteinG != 10.0 {
		t.Errorf("Expected ProteinG 10.0, got %f", result.ProteinG)
	}
	if result.TotalFatG != 5.0 {
		t.Errorf("Expected TotalFatG 5.0, got %f", result.TotalFatG)
	}
	if result.Note == nil || *result.Note != note {
		t.Errorf("Expected Note '%s', got %v", note, result.Note)
	}
}

func TestConvertConsumptionItemToAPIItemWithNutrition(t *testing.T) {
	userQuantity := 1.5
	userUnit := "slice"
	note := "test note"

	storageItem := &storage.ConsumptionItem{
		ID:            "test-id",
		ConsumptionID: "test-consumption-id",
		Name:          "pizza slice",
		Brand:         "test brand",
		Grams:         125.0,
		UserQuantity:  &userQuantity,
		UserUnit:      &userUnit,
		Note:          &note,
		Calories:      300.0,
		ProteinG:      15.0,
		TotalFatG:     12.0,
		CreatedAt:     time.Now(),
	}

	result := convertConsumptionItemToAPIItemWithNutrition(storageItem)

	// Validate the conversion
	if result.Item.Name != "pizza slice" {
		t.Errorf("Expected Name 'pizza slice', got %s", result.Item.Name)
	}
	if result.Item.Brand == nil || *result.Item.Brand != "test brand" {
		t.Errorf("Expected Brand 'test brand', got %v", result.Item.Brand)
	}
	if result.Item.Grams != 125.0 {
		t.Errorf("Expected Grams 125.0, got %f", result.Item.Grams)
	}
	if result.Item.UserQuantity == nil || *result.Item.UserQuantity != 1.5 {
		t.Errorf("Expected UserQuantity 1.5, got %v", result.Item.UserQuantity)
	}
	if result.Item.UserUnit == nil || *result.Item.UserUnit != userUnit {
		t.Errorf("Expected UserUnit '%s', got %v", userUnit, result.Item.UserUnit)
	}
	if result.Item.Nutrients == nil {
		t.Fatal("Expected Nutrients to be populated")
	}
	if result.Item.Nutrients.Calories != 300.0 {
		t.Errorf("Expected Calories 300.0, got %f", result.Item.Nutrients.Calories)
	}
	if result.Item.Nutrients.ProteinG != 15.0 {
		t.Errorf("Expected ProteinG 15.0, got %f", result.Item.Nutrients.ProteinG)
	}
	if result.Item.Nutrients.TotalFatG != 12.0 {
		t.Errorf("Expected TotalFatG 12.0, got %f", result.Item.Nutrients.TotalFatG)
	}
}

func TestConvertStorageConsumptionToAPI(t *testing.T) {
	userQuantity := 2.0
	userUnit := "cups"
	
	consumption := &storage.Consumption{
		ID:             "test-consumption-id",
		UserID:         "test-user-id", 
		Transcript:     "I had yogurt and berries",
		TotalCalories:  250.0,
		TotalProtein:   12.0,
		TotalFat:       8.0,
		CreatedAt:      time.Now(),
	}

	items := []*storage.ConsumptionItem{
		{
			ID:            "item-1",
			ConsumptionID: consumption.ID,
			Name:          "yogurt",
			Brand:         "greek",
			Grams:         150.0,
			UserQuantity:  &userQuantity,
			UserUnit:      &userUnit,
			Calories:      200.0,
			ProteinG:      10.0,
			TotalFatG:     6.0,
			CreatedAt:     time.Now(),
		},
		{
			ID:            "item-2", 
			ConsumptionID: consumption.ID,
			Name:          "berries",
			Brand:         "",
			Grams:         50.0,
			Calories:      50.0,
			ProteinG:      2.0,
			TotalFatG:     2.0,
			CreatedAt:     time.Now(),
		},
	}

	result := convertStorageConsumptionToAPI(consumption, items)

	// Validate the conversion
	if result.Id != consumption.ID {
		t.Errorf("Expected Id %s, got %s", consumption.ID, result.Id)
	}
	if result.Transcript != consumption.Transcript {
		t.Errorf("Expected Transcript '%s', got %s", consumption.Transcript, result.Transcript)
	}
	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Item.Name != "yogurt" {
		t.Errorf("Expected first item name 'yogurt', got %s", result.Items[0].Item.Name)
	}
	if result.Items[1].Item.Name != "berries" {
		t.Errorf("Expected second item name 'berries', got %s", result.Items[1].Item.Name)
	}
	if result.Summary.Totals.Calories != 250.0 {
		t.Errorf("Expected summary calories 250.0, got %f", result.Summary.Totals.Calories)
	}
	if result.Summary.Totals.ProteinG != 12.0 {
		t.Errorf("Expected summary protein 12.0, got %f", result.Summary.Totals.ProteinG)
	}
}