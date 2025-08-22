package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/storage"
)

// DateRangeParams represents parsed date range parameters from a request
type DateRangeParams struct {
	StartTime time.Time
	EndTime   time.Time
	Days      int
}

// getDefaultUser retrieves the default user for development/demo purposes
// TODO: When implementing Supabase auth, replace this with proper user resolution
// from JWT tokens or session management
func getDefaultUser(ctx context.Context, store storage.Store) (*storage.User, error) {
	return store.GetUserBySubject(ctx, DefaultUserProvider, DefaultUserSubject)
}

// getCurrentUser retrieves the current user from context (dev user if set, otherwise default)
// TODO: When implementing Supabase auth, this should extract user from JWT context
func getCurrentUser(c *gin.Context, store storage.Store) (*storage.User, error) {
	// Check if dev user is set in context (development mode only)
	if devUser, exists := c.Get("dev_user"); exists {
		if user, ok := devUser.(*storage.User); ok {
			return user, nil
		}
	}

	// Fall back to default user
	return getDefaultUser(c.Request.Context(), store)
}

// parseDateRangeParams parses date range parameters from HTTP request
// Supports both direct date range (start/end) and simplified timeWindow approach
func parseDateRangeParams(r *http.Request) (*DateRangeParams, error) {
	startParam := r.URL.Query().Get("start")
	endParam := r.URL.Query().Get("end")

	var startTime, endTime time.Time
	var days int

	if startParam != "" && endParam != "" {
		// Direct date range provided by client (already in UTC)
		var err error
		startTime, err = time.Parse(time.RFC3339, startParam)
		if err != nil {
			return nil, NewAppError("Invalid start date format", http.StatusBadRequest, err)
		}

		endTime, err = time.Parse(time.RFC3339, endParam)
		if err != nil {
			return nil, NewAppError("Invalid end date format", http.StatusBadRequest, err)
		}

		// Calculate days for subscription validation
		days = int(endTime.Sub(startTime).Hours()/24) + 1
	} else {
		// Use simplified approach - check for days parameter
		days = DefaultDays // default to week view

		dayStr := r.URL.Query().Get("days")
		if dayStr != "" {
			if parsedDays, err := strconv.Atoi(dayStr); err == nil && parsedDays > 0 {
				days = parsedDays
			}
		}

		// Calculate date range using server UTC time
		endTime = time.Now().UTC()
		startTime = endTime.AddDate(0, 0, -days+1)

		// Set times to beginning/end of day for proper date range queries
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)
	}

	return &DateRangeParams{
		StartTime: startTime,
		EndTime:   endTime,
		Days:      days,
	}, nil
}

// validateSubscriptionAccess checks if user has access to the requested number of days
func validateSubscriptionAccess(user *storage.User, days int) error {
	if days > 1 && strings.ToLower(user.SubscriptionTier) != SubscriptionTierPro {
		return NewAppError("Week view requires pro subscription", http.StatusForbidden, nil)
	}
	return nil
}

// applyDaysLimit applies the maximum days limit for performance
func applyDaysLimit(days int) int {
	if days > MaxDaysAllowed {
		return MaxDaysAllowed
	}
	return days
}

// normalizeItemName normalizes item names for consistent matching
func normalizeItemName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// getBrandOrEmpty returns the brand string or empty string if nil
func getBrandOrEmpty(brand *string) string {
	if brand == nil {
		return ""
	}
	return *brand
}
