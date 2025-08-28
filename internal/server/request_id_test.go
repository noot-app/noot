package server

import (
	"strings"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	t.Run("generates_unique_ids", func(t *testing.T) {
		id1 := generateRequestID()
		id2 := generateRequestID()

		// Should be different
		if id1 == id2 {
			t.Errorf("generateRequestID() should generate unique IDs, got same ID twice: %s", id1)
		}

		// Should be hex encoded (32 chars for 16 bytes)
		if len(id1) != 32 {
			t.Errorf("generateRequestID() should return 32 character hex string, got %d characters", len(id1))
		}

		// Should only contain hex characters
		for _, char := range id1 {
			if !strings.ContainsRune("0123456789abcdef", char) {
				t.Errorf("generateRequestID() should return hex string, found invalid character: %c", char)
			}
		}
	})

	t.Run("generates_multiple_unique_ids", func(t *testing.T) {
		ids := make(map[string]bool)

		// Generate 100 IDs and ensure they're all unique
		for i := 0; i < 100; i++ {
			id := generateRequestID()
			if ids[id] {
				t.Errorf("generateRequestID() generated duplicate ID: %s", id)
			}
			ids[id] = true
		}

		if len(ids) != 100 {
			t.Errorf("Expected 100 unique IDs, got %d", len(ids))
		}
	})
}
