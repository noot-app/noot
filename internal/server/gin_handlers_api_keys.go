package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// ListAPIKeys implements ServerInterface.ListAPIKeys
func (s *APIServer) ListAPIKeys(c *gin.Context) {
	if !checkStorageAvailable(c, s.store) {
		return
	}

	requestID := getRequestID(c)
	ctx := c.Request.Context()

	// Get authenticated user
	user, err := getCurrentUser(c)
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			handleInternalServerError(c, "Failed to get user", err)
		}
		return
	}

	// Verify user is Pro
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pro subscription required"})
		return
	}

	// Get API keys for user
	apiKeys, err := s.store.ListAPIKeys(ctx, user.ID)
	if err != nil {
		handleInternalServerError(c, "Failed to list API keys", err)
		return
	}

	// Convert to API response format
	responseKeys := make([]api.APIKey, 0, len(apiKeys))
	for _, key := range apiKeys {
		responseKeys = append(responseKeys, convertStorageAPIKeyToAPI(key))
	}

	response := api.APIKeysResponse{
		ApiKeys: responseKeys,
	}

	c.JSON(http.StatusOK, response)
}

// CreateAPIKey implements ServerInterface.CreateAPIKey
// Creates a new API key for the authenticated Pro user
func (s *APIServer) CreateAPIKey(c *gin.Context) {
	if !checkStorageAvailable(c, s.store) {
		return
	}

	requestID := getRequestID(c)
	ctx := c.Request.Context()

	// Get authenticated user
	user, err := getCurrentUser(c)
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			handleInternalServerError(c, "Failed to get user", err)
		}
		return
	}

	// Verify user is Pro
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pro subscription required"})
		return
	}

	// Parse request body
	var request api.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if request.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if request.Scope != api.CreateAPIKeyRequestScopeRead && request.Scope != api.CreateAPIKeyRequestScopeReadWrite {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scope. Must be 'read' or 'read_write'"})
		return
	}

	// Generate new API key
	fullKey, prefix, hash, err := storage.GenerateAPIKey()
	if err != nil {
		handleInternalServerError(c, "Failed to generate API key", err)
		return
	}

	// Create API key record
	apiKey := &storage.APIKey{
		UserID:    user.ID,
		Name:      request.Name,
		Prefix:    prefix,
		Hash:      hash,
		Scope:     string(request.Scope),
		CreatedAt: time.Now(),
		ExpiresAt: request.ExpiresAt,
	}

	if err := s.store.CreateAPIKey(ctx, apiKey); err != nil {
		// Check for duplicate name error
		if strings.Contains(err.Error(), "unique_api_key_name_per_user") ||
			strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API key name already exists"})
			return
		}
		handleInternalServerError(c, "Failed to create API key", err)
		return
	}

	// Return response with secret (only time it's returned)
	response := api.CreateAPIKeyResponse{
		ApiKey: convertStorageAPIKeyToAPI(apiKey),
		Secret: fullKey,
	}

	c.JSON(http.StatusCreated, response)
}

// RevokeAPIKey implements ServerInterface.RevokeAPIKey
// Revokes (soft deletes) an API key for the authenticated Pro user
func (s *APIServer) RevokeAPIKey(c *gin.Context, id string) {
	if !checkStorageAvailable(c, s.store) {
		return
	}

	requestID := getRequestID(c)
	ctx := c.Request.Context()

	// Get authenticated user
	user, err := getCurrentUser(c)
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			handleInternalServerError(c, "Failed to get user", err)
		}
		return
	}

	// Verify user is Pro
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pro subscription required"})
		return
	}

	// Revoke the API key
	if err := s.store.RevokeAPIKey(ctx, user.ID, id); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not owned") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		handleInternalServerError(c, "Failed to revoke API key", err)
		return
	}

	response := api.DeleteResponse{
		Message: "API key revoked successfully",
		Id:      id,
	}

	c.JSON(http.StatusOK, response)
}

// RotateAPIKey implements ServerInterface.RotateAPIKey
// Rotates an API key by generating a new secret and revoking the old one
func (s *APIServer) RotateAPIKey(c *gin.Context, id string) {
	if !checkStorageAvailable(c, s.store) {
		return
	}

	requestID := getRequestID(c)
	ctx := c.Request.Context()

	// Get authenticated user
	user, err := getCurrentUser(c)
	if err != nil {
		if appErr, ok := err.(*AppError); ok {
			s.handleAppError(c, appErr, requestID)
		} else {
			handleInternalServerError(c, "Failed to get user", err)
		}
		return
	}

	// Verify user is Pro
	if user.SubscriptionTier != storage.SubscriptionTierPro {
		c.JSON(http.StatusForbidden, gin.H{"error": "Pro subscription required"})
		return
	}

	// Get existing API key to verify ownership
	existingKey, err := s.store.GetAPIKey(ctx, user.ID, id)
	if err != nil {
		handleInternalServerError(c, "Failed to get API key", err)
		return
	}

	if existingKey == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	// Check if key is already revoked
	if existingKey.RevokedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot rotate revoked API key"})
		return
	}

	// Generate new API key secret
	fullKey, prefix, hash, err := storage.GenerateAPIKey()
	if err != nil {
		handleInternalServerError(c, "Failed to generate new API key", err)
		return
	}

	// Update the key with new credentials
	existingKey.Prefix = prefix
	existingKey.Hash = hash
	existingKey.LastUsedAt = nil // Reset last used

	if err := s.store.UpdateAPIKey(ctx, existingKey); err != nil {
		handleInternalServerError(c, "Failed to update API key", err)
		return
	}

	// Return response with new secret (only time it's returned)
	response := api.CreateAPIKeyResponse{
		ApiKey: convertStorageAPIKeyToAPI(existingKey),
		Secret: fullKey,
	}

	c.JSON(http.StatusOK, response)
}

// convertStorageAPIKeyToAPI converts a storage APIKey to an API APIKey
func convertStorageAPIKeyToAPI(key *storage.APIKey) api.APIKey {
	return api.APIKey{
		Id:         key.ID,
		UserId:     key.UserID,
		Name:       key.Name,
		Prefix:     key.Prefix,
		Scope:      api.APIKeyScope(key.Scope),
		CreatedAt:  key.CreatedAt,
		LastUsedAt: key.LastUsedAt,
		ExpiresAt:  key.ExpiresAt,
		RevokedAt:  key.RevokedAt,
	}
}