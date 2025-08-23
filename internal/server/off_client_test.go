package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	InitLogger()
	m.Run()
}

func TestNewOFFClient(t *testing.T) {
	tests := []struct {
		name    string
		config  OFFClientConfig
		wantNil bool
	}{
		{
			name: "enabled client",
			config: OFFClientConfig{
				BaseURL:   "https://world.openfoodfacts.org",
				UserAgent: "test-agent",
				Timeout:   5 * time.Second,
				Enabled:   true,
			},
			wantNil: false,
		},
		{
			name: "disabled client",
			config: OFFClientConfig{
				Enabled: false,
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOFFClient(tt.config)
			if tt.wantNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
				assert.Equal(t, tt.config.BaseURL, client.baseURL)
				assert.Equal(t, tt.config.UserAgent, client.userAgent)
			}
		})
	}
}

func TestOFFClient_SearchProduct(t *testing.T) {
	// Mock OFF API response
	mockResponse := OFFSearchResponse{
		Products: []OFFProduct{
			{
				ProductName: "Whole Milk",
				Brands:      "Clover",
				Nutriments: OFFNutriments{
					EnergyKcal100g:    flexFloatPtr(61),
					Proteins100g:      flexFloatPtr(3.2),
					Fat100g:           flexFloatPtr(3.5),
					Carbohydrates100g: flexFloatPtr(4.8),
					Calcium100g:       flexFloatPtr(113),
				},
			},
		},
		Count: 1,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/search", r.URL.Path)
		// Check if query contains the expected parameters (URL encoded)
		// With new brand filtering logic: q should be "whole milk" and brands_tags should be "clover"
		assert.Contains(t, r.URL.RawQuery, "q=whole+milk")
		assert.Contains(t, r.URL.RawQuery, "brands_tags=clover")
		assert.Contains(t, r.URL.RawQuery, "fields=product_name%2Cbrands%2Cnutriments%2Cid%2Ccode%2Cserving_quantity%2Cserving_quantity_unit%2Cserving_size%2Cingredients%2Clink%2Cgrade%2Cis_beverage")
		assert.Contains(t, r.URL.RawQuery, "page_size=50")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := NewOFFClient(OFFClientConfig{
		BaseURL:   server.URL,
		UserAgent: "test-agent",
		Timeout:   5 * time.Second,
		Enabled:   true,
	})

	// Test search
	ctx := context.Background()
	product, err := client.SearchProduct(ctx, "whole milk", "clover")

	require.NoError(t, err)
	require.NotNil(t, product)
	assert.Equal(t, "Whole Milk", product.ProductName)
	assert.Equal(t, "Clover", product.Brands)
}

func TestOFFClient_SearchProduct_WithServingSize(t *testing.T) {
	// Mock OFF API response with serving size information
	servingQuantity := FlexFloat(355)
	mockResponse := OFFSearchResponse{
		Products: []OFFProduct{
			{
				ProductName:         "Cream Soda",
				Brands:              "Olipop",
				ServingQuantity:     &servingQuantity,
				ServingQuantityUnit: "ml",
				ServingSize:         "1 can (355 ml)",
				Nutriments: OFFNutriments{
					EnergyKcal100g:    flexFloatPtr(11.3),
					Carbohydrates100g: flexFloatPtr(4.79),
					Sugars100g:        flexFloatPtr(0.563),
					Fiber100g:         flexFloatPtr(2.54),
				},
			},
		},
		Count: 1,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/search", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "q=cream+soda")
		assert.Contains(t, r.URL.RawQuery, "brands_tags=olipop")
		assert.Contains(t, r.URL.RawQuery, "serving_quantity")
		assert.Contains(t, r.URL.RawQuery, "serving_size")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := NewOFFClient(OFFClientConfig{
		BaseURL:   server.URL,
		UserAgent: "test-agent",
		Timeout:   5 * time.Second,
		Enabled:   true,
	})

	// Test search
	ctx := context.Background()
	product, err := client.SearchProduct(ctx, "cream soda", "olipop")

	require.NoError(t, err)
	require.NotNil(t, product)
	assert.Equal(t, "Cream Soda", product.ProductName)
	assert.Equal(t, "Olipop", product.Brands)

	// Verify serving size information
	require.NotNil(t, product.ServingQuantity)
	assert.Equal(t, FlexFloat(355), *product.ServingQuantity)
	assert.Equal(t, "ml", product.ServingQuantityUnit)
	assert.Equal(t, "1 can (355 ml)", product.ServingSize)
}

func TestOFFClient_SearchProduct_WithAdditionalFields(t *testing.T) {
	// Mock OFF API response with additional fields
	servingQuantity := FlexFloat(355)
	beverageFlag := 1
	mockResponse := OFFSearchResponse{
		Products: []OFFProduct{
			{
				ProductName:         "Craft Soda",
				Brands:              "Artisan Brand",
				ServingQuantity:     &servingQuantity,
				ServingQuantityUnit: "ml",
				ServingSize:         "1 bottle (355 ml)",
				Ingredients:         []interface{}{"water", "cane sugar", "natural flavors", "citric acid"},
				Link:                "https://world.openfoodfacts.org/product/1234567890",
				Grade:               "b",
				IsBeverage:          &beverageFlag,
				Nutriments: OFFNutriments{
					EnergyKcal100g:    flexFloatPtr(42),
					Carbohydrates100g: flexFloatPtr(10.5),
					Sugars100g:        flexFloatPtr(10.0),
				},
			},
		},
		Count: 1,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/search", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "q=craft+soda")
		assert.Contains(t, r.URL.RawQuery, "brands_tags=artisan-brand")
		assert.Contains(t, r.URL.RawQuery, "ingredients")
		assert.Contains(t, r.URL.RawQuery, "link")
		assert.Contains(t, r.URL.RawQuery, "grade")
		assert.Contains(t, r.URL.RawQuery, "is_beverage")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := NewOFFClient(OFFClientConfig{
		BaseURL:   server.URL,
		UserAgent: "test-agent",
		Timeout:   5 * time.Second,
		Enabled:   true,
	})

	// Test search
	ctx := context.Background()
	product, err := client.SearchProduct(ctx, "craft soda", "artisan brand")

	require.NoError(t, err)
	require.NotNil(t, product)
	assert.Equal(t, "Craft Soda", product.ProductName)
	assert.Equal(t, "Artisan Brand", product.Brands)

	// Verify serving size information
	require.NotNil(t, product.ServingQuantity)
	assert.Equal(t, FlexFloat(355), *product.ServingQuantity)
	assert.Equal(t, "ml", product.ServingQuantityUnit)
	assert.Equal(t, "1 bottle (355 ml)", product.ServingSize)

	// Verify additional fields
	require.Len(t, product.Ingredients, 4)
	assert.Equal(t, "water", product.Ingredients[0])
	assert.Equal(t, "cane sugar", product.Ingredients[1])
	assert.Equal(t, "natural flavors", product.Ingredients[2])
	assert.Equal(t, "citric acid", product.Ingredients[3])

	assert.Equal(t, "https://world.openfoodfacts.org/product/1234567890", product.Link)
	assert.Equal(t, "b", product.Grade)
	require.NotNil(t, product.IsBeverage)
	assert.Equal(t, 1, *product.IsBeverage)
}

func TestOFFClient_SearchProduct_NoResults(t *testing.T) {
	// Mock empty response
	mockResponse := OFFSearchResponse{
		Products: []OFFProduct{},
		Count:    0,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewOFFClient(OFFClientConfig{
		BaseURL:   server.URL,
		UserAgent: "test-agent",
		Timeout:   5 * time.Second,
		Enabled:   true,
	})

	ctx := context.Background()
	product, err := client.SearchProduct(ctx, "nonexistent item", "")

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "no suitable product found in OFF")
}

func TestOFFClient_ConvertToCompleteNutrient_WithPerServingData(t *testing.T) {
	client := NewOFFClient(OFFClientConfig{Enabled: true})

	// Test exact serving match with per-serving nutrition data
	servingQuantity := FlexFloat(355)
	product := &OFFProduct{
		ProductName:         "Cream Soda",
		Brands:              "Olipop",
		ServingQuantity:     &servingQuantity,
		ServingQuantityUnit: "ml",
		Nutriments: OFFNutriments{
			// Per-100g values would give 40.115 calories when scaled
			EnergyKcal100g: flexFloatPtr(11.3),

			// Per-serving values (exact from OFF database)
			EnergyKcalServing: flexFloatPtr(35), // Should use this instead
		},
	}

	// Test exact serving size match - should use per-serving data
	nutrition := client.ConvertToCompleteNutrient(product, 355)

	// Should use exact per-serving value (35) not scaled 100g value (40.115)
	assert.Equal(t, 35.0, nutrition.Calories, "Should use exact per-serving calories")

	// Test non-exact serving size - should fall back to 100g scaling
	nutrition = client.ConvertToCompleteNutrient(product, 200)
	expected := (11.3 / 100.0) * 200 // = 22.6
	assert.Equal(t, expected, nutrition.Calories, "Should scale from 100g when not exact serving match")
}

func TestOFFClient_ConvertToCompleteNutrient(t *testing.T) {
	client := NewOFFClient(OFFClientConfig{Enabled: true})

	product := &OFFProduct{
		ProductName: "Test Product",
		Nutriments: OFFNutriments{
			EnergyKcal100g:    flexFloatPtr(100),
			Proteins100g:      flexFloatPtr(10),
			Fat100g:           flexFloatPtr(5),
			Carbohydrates100g: flexFloatPtr(12),
			Sodium100g:        flexFloatPtr(500), // mg
		},
	}

	// Test conversion for 200g serving
	nutrients := client.ConvertToCompleteNutrient(product, 200)

	// Values should be scaled by 2 (200g / 100g)
	assert.Equal(t, 200.0, nutrients.Calories)
	assert.Equal(t, 20.0, nutrients.Protein)
	assert.Equal(t, 10.0, nutrients.TotalFat)
	assert.Equal(t, 24.0, nutrients.TotalCarbs)
	assert.Equal(t, 1000.0, nutrients.Sodium)

	// Fields not in OFF should be 0
	assert.Equal(t, 0.0, nutrients.TransFat)
	assert.Equal(t, 0.0, nutrients.Cholesterol)
	assert.Equal(t, 0.0, nutrients.VitaminB12)
}

func TestOFFClient_CalculateMatchScore(t *testing.T) {
	client := NewOFFClient(OFFClientConfig{Enabled: true})

	tests := []struct {
		name        string
		product     *OFFProduct
		searchName  string
		searchBrand string
		expectMin   float64
		expectMax   float64
	}{
		{
			name: "exact match",
			product: &OFFProduct{
				ProductName: "whole milk",
				Brands:      "clover",
			},
			searchName:  "whole milk",
			searchBrand: "clover",
			expectMin:   0.8,
			expectMax:   1.0,
		},
		{
			name: "partial match",
			product: &OFFProduct{
				ProductName: "organic whole milk",
				Brands:      "clover organic",
			},
			searchName:  "whole milk",
			searchBrand: "clover",
			expectMin:   0.4,
			expectMax:   0.8,
		},
		{
			name: "poor match",
			product: &OFFProduct{
				ProductName: "skim milk",
				Brands:      "different brand",
			},
			searchName:  "whole milk",
			searchBrand: "clover",
			expectMin:   0.0,
			expectMax:   0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := client.calculateMatchScore(tt.product, tt.searchName, tt.searchBrand)
			assert.GreaterOrEqual(t, score, tt.expectMin)
			assert.LessOrEqual(t, score, tt.expectMax)
		})
	}
}

func TestOFFClient_NilClient(t *testing.T) {
	var client *OFFClient = nil

	ctx := context.Background()
	product, err := client.SearchProduct(ctx, "test", "")

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "OFF client not initialized")
}
