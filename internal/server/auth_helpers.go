package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/storage"
)

// GetAuthenticatedUser extracts the authenticated user from Gin context
func GetAuthenticatedUser(c *gin.Context) *storage.User {
	// Check for JWT auth user
	if user, exists := c.Get("auth_user"); exists {
		if authUser, ok := user.(*storage.User); ok {
			return authUser
		}
	}

	return nil
}

// RequireAuth middleware ensures a user is authenticated
// Can be used on individual routes that need authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthenticatedUser(c)
		if user == nil {
			// More specific error if no Authorization header was provided
			if c.GetHeader("Authorization") == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireSubscriptionTiers middleware ensures user has one of the specified subscription tiers
func RequireSubscriptionTiers(allowedTiers ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthenticatedUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if user's tier is in the allowed list
		for _, tier := range allowedTiers {
			if user.SubscriptionTier == tier {
				c.Next()
				return
			}
		}

		// Build error message with allowed tiers
		tierNames := append([]string(nil), allowedTiers...)

		errorMsg := "Subscription required: " + strings.Join(tierNames, " or ")
		c.JSON(http.StatusForbidden, gin.H{"error": errorMsg})
		c.Abort()
	}
}

// RequireProSubscription middleware ensures user has pro subscription (backward compatibility)
func RequireProSubscription() gin.HandlerFunc {
	return RequireSubscriptionTiers(storage.SubscriptionTierPro)
}
