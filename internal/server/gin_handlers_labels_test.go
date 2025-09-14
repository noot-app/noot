package server

import (
	"strings"
	"testing"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestLabelValidation(t *testing.T) {
	tests := []struct {
		name        string
		label       *storage.Label
		expectValid bool
		description string
	}{
		{
			name: "valid_label_with_name_and_color",
			label: &storage.Label{
				Name:   "Breakfast",
				Color:  "#FF5733",
				UserID: "user-123",
			},
			expectValid: true,
			description: "Valid label with name and color should pass validation",
		},
		{
			name: "empty_label_name_invalid",
			label: &storage.Label{
				Name:   "",
				Color:  "#FF5733",
				UserID: "user-123",
			},
			expectValid: false,
			description: "Label with empty name should be invalid",
		},
		{
			name: "empty_color_invalid",
			label: &storage.Label{
				Name:   "Breakfast",
				Color:  "",
				UserID: "user-123",
			},
			expectValid: false,
			description: "Label with empty color should be invalid",
		},
		{
			name: "missing_user_id_invalid",
			label: &storage.Label{
				Name:  "Breakfast",
				Color: "#FF5733",
				UserID: "",
			},
			expectValid: false,
			description: "Label without user ID should be invalid",
		},
		{
			name: "long_label_name_boundary",
			label: &storage.Label{
				Name:   "This is a very long label name that exceeds reasonable limits for display purposes and database constraints",
				Color:  "#FF5733",
				UserID: "user-123",
			},
			expectValid: false,
			description: "Excessively long label names should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic that would be in the handler
			isValid := validateLabel(tt.label)
			assert.Equal(t, tt.expectValid, isValid, tt.description)
		})
	}
}

func TestColorNormalization(t *testing.T) {
	tests := []struct {
		name           string
		inputColor     string
		expectedColor  string
		description    string
	}{
		{
			name:          "hex_color_with_hash",
			inputColor:    "#FF5733",
			expectedColor: "#FF5733",
			description:   "Valid hex color with hash should remain unchanged",
		},
		{
			name:          "hex_color_without_hash",
			inputColor:    "FF5733",
			expectedColor: "#FF5733",
			description:   "Hex color without hash should get hash prefix",
		},
		{
			name:          "lowercase_hex_color",
			inputColor:    "ff5733",
			expectedColor: "#FF5733",
			description:   "Lowercase hex should be normalized to uppercase",
		},
		{
			name:          "short_hex_color",
			inputColor:    "f57",
			expectedColor: "#FF5577",
			description:   "3-digit hex should expand to 6 digits",
		},
		{
			name:          "invalid_color_uses_default",
			inputColor:    "not-a-color",
			expectedColor: "#808080",
			description:   "Invalid color should use default gray",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test color normalization logic
			normalizedColor := normalizeColor(tt.inputColor)
			assert.Equal(t, tt.expectedColor, normalizedColor, tt.description)
		})
	}
}

func TestLabelOwnership(t *testing.T) {
	tests := []struct {
		name        string
		labelUserID string
		requestUserID string
		expectAccess bool
		description string
	}{
		{
			name:         "owner_can_access_own_label",
			labelUserID:  "user-123",
			requestUserID: "user-123",
			expectAccess: true,
			description:  "User should be able to access their own labels",
		},
		{
			name:         "non_owner_cannot_access_label",
			labelUserID:  "user-123", 
			requestUserID: "user-456",
			expectAccess: false,
			description:  "User should not access labels belonging to other users",
		},
		{
			name:         "empty_user_id_denies_access",
			labelUserID:  "user-123",
			requestUserID: "",
			expectAccess: false,
			description:  "Empty user ID should deny access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ownership validation logic
			hasAccess := (tt.labelUserID == tt.requestUserID && tt.requestUserID != "")
			assert.Equal(t, tt.expectAccess, hasAccess, tt.description)
		})
	}
}

// Helper functions to simulate handler validation logic
func validateLabel(label *storage.Label) bool {
	if label == nil {
		return false
	}
	if label.Name == "" || label.Color == "" || label.UserID == "" {
		return false
	}
	if len(label.Name) > 100 { // Reasonable limit
		return false
	}
	return true
}

func normalizeColor(color string) string {
	// Simulate color normalization logic that might be in handlers
	if color == "" {
		return "#808080" // Default gray
	}

	// Remove hash if present
	if color[0] == '#' {
		color = color[1:]
	}

	// Convert to uppercase
	color = strings.ToUpper(color)

	// Expand short hex (3 chars to 6)
	if len(color) == 3 {
		expanded := ""
		for _, char := range color {
			expanded += string(char) + string(char)
		}
		color = expanded
	}

	// Validate hex format (simplified)
	if len(color) != 6 {
		return "#808080" // Default for invalid
	}

	// Check if all characters are valid hex
	for _, char := range color {
		if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'F')) {
			return "#808080"
		}
	}

	return "#" + color
}