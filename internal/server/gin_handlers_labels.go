package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grantbirki/noot/internal/api"
	"github.com/grantbirki/noot/internal/storage"
)

// GetLabels retrieves all labels for the current user with usage counts
func (s *APIServer) GetLabels(c *gin.Context) {
	user := requireAuthentication(c)
	if user == nil {
		return // Error already handled by requireAuthentication
	}

	ctx := c.Request.Context()
	labels, err := s.store.ListLabels(ctx, user.ID)
	if err != nil {
		appErr := NewAppError("Failed to retrieve labels", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, getRequestID(c))
		return
	}

	// Convert to API format
	apiLabels := make([]api.LabelWithUsage, len(labels))
	for i, label := range labels {
		apiLabels[i] = api.LabelWithUsage{
			Id:               label.ID,
			Name:             label.Name,
			Description:      label.Description,
			Color:            label.Color,
			CreatedAt:        label.CreatedAt,
			UpdatedAt:        label.UpdatedAt,
			ConsumptionCount: label.ConsumptionCount,
			ItemCount:        label.ItemCount,
		}
	}

	response := api.LabelsResponse{
		Labels: apiLabels,
	}

	c.JSON(http.StatusOK, response)
}

// CreateLabel creates a new label for the current user
func (s *APIServer) CreateLabel(c *gin.Context) {
	user := requireAuthentication(c)
	if user == nil {
		return // Error already handled by requireAuthentication
	}

	var req api.LabelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, getRequestID(c))
		return
	}

	// Validate required fields
	if req.Name == "" {
		appErr := NewAppError("Label name is required", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, getRequestID(c))
		return
	}

	if req.Color == "" {
		appErr := NewAppError("Label color is required", http.StatusBadRequest, nil)
		s.handleAppError(c, appErr, getRequestID(c))
		return
	}

	// Create storage label
	label := &storage.Label{
		UserID:      user.ID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}

	ctx := c.Request.Context()
	err := s.store.CreateLabel(ctx, label)
	if err != nil {
		requestID := getRequestID(c)
		if err.Error() == "label name already exists" {
			appErr := NewAppError("Label name already exists", http.StatusConflict, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
		if err.Error() == "label limit exceeded (100 labels per user)" {
			appErr := NewAppError("Label limit exceeded (100 labels per user)", http.StatusForbidden, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
		appErr := NewAppError("Failed to create label", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API format
	apiLabel := api.Label{
		Id:          label.ID,
		Name:        label.Name,
		Description: label.Description,
		Color:       label.Color,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
	}

	c.JSON(http.StatusCreated, apiLabel)
}

// UpdateLabel updates an existing label
func (s *APIServer) UpdateLabel(c *gin.Context, id string) {
	user := requireAuthentication(c)
	if user == nil {
		return // Error already handled by requireAuthentication
	}

	var req api.LabelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := NewAppError("Invalid request body", http.StatusBadRequest, err)
		s.handleAppError(c, appErr, getRequestID(c))
		return
	}

	ctx := c.Request.Context()

	// Get existing label to ensure it exists and user owns it
	existingLabel, err := s.store.GetLabel(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Label not found", http.StatusNotFound, err)
		s.handleAppError(c, appErr, getRequestID(c))
		return
	}

	// Update fields if provided
	if req.Name != nil {
		existingLabel.Name = *req.Name
	}
	if req.Description != nil {
		existingLabel.Description = req.Description
	}
	if req.Color != nil {
		existingLabel.Color = *req.Color
	}

	err = s.store.UpdateLabel(ctx, existingLabel)
	if err != nil {
		requestID := getRequestID(c)
		if err.Error() == "label name already exists" {
			appErr := NewAppError("Label name already exists", http.StatusConflict, err)
			s.handleAppError(c, appErr, requestID)
			return
		}
		appErr := NewAppError("Failed to update label", http.StatusInternalServerError, err)
		s.handleAppError(c, appErr, requestID)
		return
	}

	// Convert to API format
	apiLabel := api.Label{
		Id:          existingLabel.ID,
		Name:        existingLabel.Name,
		Description: existingLabel.Description,
		Color:       existingLabel.Color,
		CreatedAt:   existingLabel.CreatedAt,
		UpdatedAt:   existingLabel.UpdatedAt,
	}

	c.JSON(http.StatusOK, apiLabel)
}

// DeleteLabel deletes a label and all its assignments
func (s *APIServer) DeleteLabel(c *gin.Context, id string) {
	user := requireAuthentication(c)
	if user == nil {
		return // Error already handled by requireAuthentication
	}

	ctx := c.Request.Context()
	err := s.store.DeleteLabel(ctx, user.ID, id)
	if err != nil {
		appErr := NewAppError("Label not found or access denied", http.StatusNotFound, err)
		s.handleAppError(c, appErr, getRequestID(c))
		return
	}

	c.Status(http.StatusNoContent)
}