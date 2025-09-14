package server

import (
	"testing"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestAPIServerCreation(t *testing.T) {
	tests := []struct {
		name        string
		store       storage.Store
		expectError bool
	}{
		{
			name:        "valid_store_creates_server",
			store:       &NullStore{},
			expectError: false,
		},
		{
			name:        "nil_store_handled_gracefully",
			store:       nil,
			expectError: false, // NewAPIServer should accept nil store
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewAPIServer(tt.store)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, server)
			}
		})
	}
}

// Use NullStore from test_helpers.go for full storage interface implementation