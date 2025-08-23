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
		name     string
		config   OFFClientConfig
		wantNil  bool
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
				CompletedT:  0.8,
				Nutriments: OFFNutriments{
					EnergyKcal100g:    offFloatPtr(61),
					Proteins100g:      offFloatPtr(3.2),
					Fat100g:           offFloatPtr(3.5),
					Carbohydrates100g: offFloatPtr(4.8),
					Calcium100g:       offFloatPtr(113),
				},
			},
		},
		Count: 1,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/search", r.URL.Path)
		// Check if query contains the expected parameters (URL encoded)
		assert.Contains(t, r.URL.RawQuery, "q=clover")
		assert.Contains(t, r.URL.RawQuery, "fields=product_name")
		assert.Contains(t, r.URL.RawQuery, "page_size=5")
		
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
	assert.Equal(t, 0.8, product.CompletedT)
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

func TestOFFClient_ConvertToCompleteNutrient(t *testing.T) {
	client := NewOFFClient(OFFClientConfig{Enabled: true})

	product := &OFFProduct{
		ProductName: "Test Product",
		Nutriments: OFFNutriments{
			EnergyKcal100g:    offFloatPtr(100),
			Proteins100g:      offFloatPtr(10),
			Fat100g:           offFloatPtr(5),
			Carbohydrates100g: offFloatPtr(12),
			Sodium100g:        offFloatPtr(500), // mg
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
		name         string
		product      *OFFProduct
		searchName   string
		searchBrand  string
		expectMin    float64
		expectMax    float64
	}{
		{
			name: "exact match",
			product: &OFFProduct{
				ProductName: "whole milk",
				Brands:      "clover",
				CompletedT:  0.9,
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
				CompletedT:  0.7,
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
				CompletedT:  0.5,
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

// Helper function to create float pointers for OFF tests
func offFloatPtr(f float64) *float64 {
	return &f
}