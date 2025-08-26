package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDateRangeParams(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedDays  int
		expectError   bool
		expectedStart bool // whether to check start time
		expectedEnd   bool // whether to check end time
	}{
		{
			name:          "direct date range",
			queryParams:   map[string]string{"start": "2024-01-01T00:00:00Z", "end": "2024-01-03T23:59:59Z"},
			expectedDays:  3,
			expectError:   false,
			expectedStart: true,
			expectedEnd:   true,
		},
		{
			name:         "days parameter only",
			queryParams:  map[string]string{"days": "5"},
			expectedDays: 5,
			expectError:  false,
		},
		{
			name:         "no parameters - default",
			queryParams:  map[string]string{},
			expectedDays: DefaultDays,
			expectError:  false,
		},
		{
			name:        "invalid start date",
			queryParams: map[string]string{"start": "invalid", "end": "2024-01-01T00:00:00Z"},
			expectError: true,
		},
		{
			name:        "invalid end date",
			queryParams: map[string]string{"start": "2024-01-01T00:00:00Z", "end": "invalid"},
			expectError: true,
		},
		{
			name:         "invalid days parameter",
			queryParams:  map[string]string{"days": "invalid"},
			expectedDays: DefaultDays,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Set(k, v)
			}
			req.URL.RawQuery = q.Encode()

			result, err := parseDateRangeParams(req)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedDays, result.Days)

			if tt.expectedStart {
				expectedStart, _ := time.Parse(time.RFC3339, tt.queryParams["start"])
				assert.Equal(t, expectedStart, result.StartTime)
			}

			if tt.expectedEnd {
				expectedEnd, _ := time.Parse(time.RFC3339, tt.queryParams["end"])
				assert.Equal(t, expectedEnd, result.EndTime)
			}

			if !tt.expectedStart && !tt.expectedEnd {
				// Check that times are set to reasonable values
				assert.False(t, result.StartTime.IsZero())
				assert.False(t, result.EndTime.IsZero())
				assert.True(t, result.EndTime.After(result.StartTime))
			}
		})
	}
}

func TestValidateSubscriptionAccess(t *testing.T) {
	tests := []struct {
		name        string
		user        *storage.User
		days        int
		expectError bool
	}{
		{
			name:        "free user single day - allowed",
			user:        &storage.User{SubscriptionTier: SubscriptionTierFree},
			days:        1,
			expectError: false,
		},
		{
			name:        "free user multiple days - denied",
			user:        &storage.User{SubscriptionTier: SubscriptionTierFree},
			days:        7,
			expectError: true,
		},
		{
			name:        "pro user multiple days - allowed",
			user:        &storage.User{SubscriptionTier: SubscriptionTierPro},
			days:        7,
			expectError: false,
		},
		{
			name:        "pro user single day - allowed",
			user:        &storage.User{SubscriptionTier: SubscriptionTierPro},
			days:        1,
			expectError: false,
		},
		{
			name:        "case insensitive pro user",
			user:        &storage.User{SubscriptionTier: "PRO"},
			days:        7,
			expectError: false, // Should pass because validation is case-insensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSubscriptionAccess(tt.user, tt.days)

			if tt.expectError {
				assert.Error(t, err)
				if appErr, ok := err.(*AppError); ok {
					assert.Equal(t, http.StatusForbidden, appErr.StatusCode)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplyDaysLimit(t *testing.T) {
	tests := []struct {
		name     string
		days     int
		expected int
	}{
		{
			name:     "within limit",
			days:     5,
			expected: 5,
		},
		{
			name:     "at limit",
			days:     MaxDaysAllowed,
			expected: MaxDaysAllowed,
		},
		{
			name:     "exceeds limit",
			days:     10,
			expected: MaxDaysAllowed,
		},
		{
			name:     "zero days",
			days:     0,
			expected: 0,
		},
		{
			name:     "negative days",
			days:     -1,
			expected: -1, // Function doesn't handle negative validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyDaysLimit(tt.days)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultUser(t *testing.T) {
	// Skip this test as it requires SQLite which has been removed
	t.Skip("Skipping test that requires SQLite - SQLite support removed")
}
