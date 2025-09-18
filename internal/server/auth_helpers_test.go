package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestGetAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		setupUser   func(*gin.Context)
		expectUser  bool
		description string
	}{
		{
			name: "user_set_in_context",
			setupUser: func(c *gin.Context) {
				user := &storage.User{
					ID:    "test-user-1",
					Email: "test@example.com",
				}
				c.Set("auth_user", user)
			},
			expectUser:  true,
			description: "Should return user when set in context",
		},
		{
			name: "no_user_in_context",
			setupUser: func(c *gin.Context) {
				// Don't set any user
			},
			expectUser:  false,
			description: "Should return nil when no user in context",
		},
		{
			name: "invalid_user_type_in_context",
			setupUser: func(c *gin.Context) {
				// Set something that's not a *storage.User
				c.Set("auth_user", "not-a-user")
			},
			expectUser:  false,
			description: "Should return nil when auth_user is not correct type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupUser(c)

			result := GetAuthenticatedUser(c)

			if tt.expectUser {
				assert.NotNil(t, result, tt.description)
				assert.Equal(t, "test-user-1", result.ID)
				assert.Equal(t, "test@example.com", result.Email)
			} else {
				assert.Nil(t, result, tt.description)
			}
		})
	}
}
