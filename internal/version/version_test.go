package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	info := Get()

	// Test that we get back the expected structure
	assert.Equal(t, "dev", info.Tag)
	assert.Equal(t, "unknown", info.Commit)
	assert.Equal(t, "unknown", info.BuildTime)
}

func TestGetCommitShort(t *testing.T) {
	tests := []struct {
		name           string
		originalCommit string
		expected       string
	}{
		{
			name:           "commit longer than 7 chars",
			originalCommit: "abcdef1234567890",
			expected:       "abcdef1",
		},
		{
			name:           "commit exactly 7 chars",
			originalCommit: "abcdef1",
			expected:       "abcdef1",
		},
		{
			name:           "commit shorter than 7 chars",
			originalCommit: "abc123",
			expected:       "abc123",
		},
		{
			name:           "empty commit",
			originalCommit: "",
			expected:       "",
		},
		{
			name:           "unknown commit",
			originalCommit: "unknown",
			expected:       "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original commit value
			originalCommit := commit

			// Set test value
			commit = tt.originalCommit

			// Test
			result := GetCommitShort()
			assert.Equal(t, tt.expected, result)

			// Restore original value
			commit = originalCommit
		})
	}
}
