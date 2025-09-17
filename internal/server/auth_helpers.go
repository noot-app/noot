package server

import (
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
