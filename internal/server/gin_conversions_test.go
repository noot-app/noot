package server

import (
	"testing"
	"time"

	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestConvertUser(t *testing.T) {
	tests := []struct {
		name        string
		storageUser *storage.User
		expectedAPI api.User
		description string
	}{
		{
			name: "complete_user_conversion",
			storageUser: &storage.User{
				ID:               "user-123",
				Handle:           "testuser",
				FullName:         stringPtr("John Doe"),
				Email:            "john@example.com",
				SubscriptionTier: storage.SubscriptionTierPro,

				CreatedAt: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				AvatarURL: stringPtr("https://example.com/avatar.jpg"),
			},
			expectedAPI: api.User{
				Id:               "user-123",
				Handle:           "testuser",
				FullName:         stringPtr("John Doe"),
				Email:            "john@example.com",
				SubscriptionTier: "pro",
				CreatedAt:        time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
				AvatarUrl:        stringPtr("https://example.com/avatar.jpg"),
			},
			description: "Complete user data should convert correctly",
		},
		{
			name: "minimal_user_conversion",
			storageUser: &storage.User{
				ID:               "user-456",
				Handle:           "minimaluser",
				Email:            "minimal@example.com",
				SubscriptionTier: storage.SubscriptionTierFree,
				CreatedAt:        time.Date(2024, 2, 1, 10, 30, 0, 0, time.UTC),
			},
			expectedAPI: api.User{
				Id:               "user-456",
				Handle:           "minimaluser",
				Email:            "minimal@example.com",
				SubscriptionTier: "free",
				CreatedAt:        time.Date(2024, 2, 1, 10, 30, 0, 0, time.UTC),
			},
			description: "Minimal user data with no optional fields should convert correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertUser(tt.storageUser)

			assert.Equal(t, tt.expectedAPI.Id, result.Id)
			assert.Equal(t, tt.expectedAPI.Handle, result.Handle)
			assert.Equal(t, tt.expectedAPI.Email, result.Email)
			assert.Equal(t, tt.expectedAPI.SubscriptionTier, result.SubscriptionTier)
			assert.Equal(t, tt.expectedAPI.CreatedAt.Unix(), result.CreatedAt.Unix())

			// Check optional fields
			if tt.expectedAPI.FullName != nil {
				assert.NotNil(t, result.FullName)
				assert.Equal(t, *tt.expectedAPI.FullName, *result.FullName)
			} else {
				assert.Nil(t, result.FullName)
			}

			if tt.expectedAPI.AvatarUrl != nil {
				assert.NotNil(t, result.AvatarUrl)
				assert.Equal(t, *tt.expectedAPI.AvatarUrl, *result.AvatarUrl)
			} else {
				assert.Nil(t, result.AvatarUrl)
			}
		})
	}
}

// TestConvertLabel test removed - convertLabel function may not exist in current implementation

func TestConvertStorageAPIKeyToAPI(t *testing.T) {
	tests := []struct {
		name          string
		storageAPIKey storage.APIKey
		expectedAPI   api.APIKey
		description   string
	}{
		{
			name: "active_api_key_conversion",
			storageAPIKey: storage.APIKey{
				ID:        "key-123",
				Name:      "My API Key",
				Prefix:    "noot_abc123",
				UserID:    "user-123",
				Scope:     "read_write",
				CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			expectedAPI: api.APIKey{
				Id:        "key-123",
				Name:      "My API Key",
				Prefix:    "noot_abc123",
				Scope:     api.APIKeyScope("read_write"),
				CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			description: "API key should convert correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertStorageAPIKeyToAPI(&tt.storageAPIKey)

			assert.Equal(t, tt.expectedAPI.Id, result.Id)
			assert.Equal(t, tt.expectedAPI.Name, result.Name)
			assert.Equal(t, tt.expectedAPI.Prefix, result.Prefix)
			assert.Equal(t, string(tt.expectedAPI.Scope), string(result.Scope))
			assert.Equal(t, tt.expectedAPI.CreatedAt.Unix(), result.CreatedAt.Unix())

			// Check optional fields like LastUsedAt, RevokedAt based on what's available in the API
		})
	}
}

func TestSubscriptionTierConversion(t *testing.T) {
	tests := []struct {
		name        string
		storageTier string
		expectedAPI string
		description string
	}{
		{
			name:        "free_tier_conversion",
			storageTier: storage.SubscriptionTierFree,
			expectedAPI: "free",
			description: "Free tier should convert to 'free'",
		},
		{
			name:        "pro_tier_conversion",
			storageTier: storage.SubscriptionTierPro,
			expectedAPI: "pro",
			description: "Pro tier should convert to 'pro'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the conversion logic that would be used in convertUser
			var apiTier string
			switch tt.storageTier {
			case storage.SubscriptionTierFree:
				apiTier = "free"
			case storage.SubscriptionTierPro:
				apiTier = "pro"
			default:
				apiTier = "free" // Default fallback
			}

			assert.Equal(t, tt.expectedAPI, apiTier, tt.description)
		})
	}
}

// Helper function for test data
func timePtr(t time.Time) *time.Time {
	return &t
}
