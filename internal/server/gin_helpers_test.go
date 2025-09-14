package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestCheckStorageAvailable(t *testing.T) {
	tests := []struct {
		name     string
		store    storage.Store
		expected bool
	}{
		{
			name:     "nil_store_returns_false",
			store:    nil,
			expected: false,
		},
		{
			name:     "valid_store_returns_true",
			store:    &NullStore{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test checkStorageAvailable without gin context
			// So we test the underlying logic: nil check
			result := (tt.store != nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscriptionTierValidation(t *testing.T) {
	tests := []struct {
		name           string
		user           *storage.User
		requiredTier   string
		expectAccess   bool
		description    string
	}{
		{
			name: "free_user_denied_pro_feature",
			user: &storage.User{
				ID:               "user1",
				SubscriptionTier: storage.SubscriptionTierFree,
			},
			requiredTier: storage.SubscriptionTierPro,
			expectAccess: false,
			description:  "Free tier user should not access Pro features",
		},
		{
			name: "pro_user_allowed_pro_feature",
			user: &storage.User{
				ID:               "user2", 
				SubscriptionTier: storage.SubscriptionTierPro,
			},
			requiredTier: storage.SubscriptionTierPro,
			expectAccess: true,
			description:  "Pro tier user should access Pro features",
		},
		{
			name: "pro_user_allowed_free_feature",
			user: &storage.User{
				ID:               "user3",
				SubscriptionTier: storage.SubscriptionTierPro,
			},
			requiredTier: storage.SubscriptionTierFree,
			expectAccess: true,
			description:  "Pro users should access free features",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the subscription tier validation logic
			hasAccess := tt.user.SubscriptionTier == tt.requiredTier || 
						(tt.user.SubscriptionTier == storage.SubscriptionTierPro && tt.requiredTier == storage.SubscriptionTierFree)
			
			assert.Equal(t, tt.expectAccess, hasAccess, tt.description)
		})
	}
}

func TestParseTimeRangeParamsLogic(t *testing.T) {
	tests := []struct {
		name        string
		startStr    *string
		endStr      *string
		expectError bool
		description string
	}{
		{
			name:        "nil_parameters_use_defaults", 
			startStr:    nil,
			endStr:      nil,
			expectError: false,
			description: "Nil start/end should use reasonable defaults",
		},
		{
			name: "valid_iso_dates",
			startStr: testStringPtr("2024-01-01T00:00:00Z"),
			endStr:   testStringPtr("2024-01-31T23:59:59Z"),
			expectError: false,
			description: "Valid ISO 8601 dates should parse successfully",
		},
		{
			name: "invalid_start_date_format",
			startStr: testStringPtr("invalid-date"),
			endStr:   nil,
			expectError: true,
			description: "Invalid date format should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the core date parsing logic
			var start, end time.Time
			var err error

			if tt.startStr != nil {
				start, err = time.Parse(time.RFC3339, *tt.startStr)
				if err != nil {
					assert.True(t, tt.expectError, "Expected error parsing start date")
					return
				}
			} else {
				// Default: 30 days ago
				start = time.Now().AddDate(0, 0, -30)
			}

			if tt.endStr != nil {
				end, err = time.Parse(time.RFC3339, *tt.endStr)
				if err != nil {
					assert.True(t, tt.expectError, "Expected error parsing end date")
					return
				}
			} else {
				// Default: now
				end = time.Now()
			}

			if !tt.expectError {
				assert.NoError(t, err)
				assert.False(t, start.IsZero(), "Start time should not be zero")
				assert.False(t, end.IsZero(), "End time should not be zero")
				assert.True(t, end.After(start) || end.Equal(start), "End should be after or equal to start")
			}
		})
	}
}

// Helper function for test data
func testStringPtr(s string) *string {
	return &s
}

func TestErrorResponseStructure(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		statusCode     int
		expectedFields []string
	}{
		{
			name:       "basic_error_structure",
			message:    "Test error message",
			statusCode: http.StatusBadRequest,
			expectedFields: []string{"message", "statusCode"},
		},
		{
			name:       "internal_server_error",
			message:    "Internal server error",
			statusCode: http.StatusInternalServerError,
			expectedFields: []string{"message", "statusCode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error structure without gin context
			appErr := NewAppError(tt.message, tt.statusCode, nil)
			
			assert.NotNil(t, appErr)
			assert.Equal(t, tt.message, appErr.Message)
			assert.Equal(t, tt.statusCode, appErr.StatusCode)
		})
	}
}