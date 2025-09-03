package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAPIKey(t *testing.T) {
	t.Run("GeneratesValidKey", func(t *testing.T) {
		fullKey, prefix, hash, err := GenerateAPIKey()
		require.NoError(t, err)

		// Verify prefix format
		assert.True(t, strings.HasPrefix(prefix, APIKeyPrefix))
		assert.Len(t, prefix, len(APIKeyPrefix)+APIKeyPrefixLength)

		// Verify full key format  
		parts := strings.Split(fullKey, "_")
		assert.Len(t, parts, 3, "Full key should have format: noot_prefix_secret")
		assert.Equal(t, "noot", parts[0])

		// Verify hash is valid bcrypt
		assert.True(t, len(hash) > 50, "Bcrypt hash should be substantial length")
		assert.True(t, strings.HasPrefix(hash, "$2a$"), "Should be bcrypt hash")

		// Verify key can be verified
		assert.True(t, VerifyAPIKey(fullKey, hash))
	})

	t.Run("GeneratesUniqueKeys", func(t *testing.T) {
		fullKey1, prefix1, hash1, err1 := GenerateAPIKey()
		fullKey2, prefix2, hash2, err2 := GenerateAPIKey()

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Keys should be unique
		assert.NotEqual(t, fullKey1, fullKey2)
		assert.NotEqual(t, prefix1, prefix2)
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestVerifyAPIKey(t *testing.T) {
	fullKey, _, hash, err := GenerateAPIKey()
	require.NoError(t, err)

	t.Run("ValidKeyVerifies", func(t *testing.T) {
		assert.True(t, VerifyAPIKey(fullKey, hash))
	})

	t.Run("InvalidKeyFails", func(t *testing.T) {
		assert.False(t, VerifyAPIKey("invalid_key", hash))
		assert.False(t, VerifyAPIKey(fullKey, "invalid_hash"))
		assert.False(t, VerifyAPIKey("", hash))
		assert.False(t, VerifyAPIKey(fullKey, ""))
	})

	t.Run("ModifiedKeyFails", func(t *testing.T) {
		modifiedKey := fullKey + "x"
		assert.False(t, VerifyAPIKey(modifiedKey, hash))
	})
}

func TestExtractPrefix(t *testing.T) {
	fullKey, expectedPrefix, _, err := GenerateAPIKey()
	require.NoError(t, err)

	t.Run("ExtractsValidPrefix", func(t *testing.T) {
		prefix, err := ExtractPrefix(fullKey)
		require.NoError(t, err)
		assert.Equal(t, expectedPrefix, prefix)
	})

	t.Run("InvalidKeyFormats", func(t *testing.T) {
		testCases := []string{
			"invalid",
			"not_noot_key",
			"noot_",
			"noot_onlyprefix",
			"wrong_prefix_secret",
			"",
		}

		for _, testKey := range testCases {
			t.Run(testKey, func(t *testing.T) {
				_, err := ExtractPrefix(testKey)
				assert.Error(t, err)
			})
		}
	})
}

func TestAPIKeyIsActive(t *testing.T) {
	now := time.Now()

	t.Run("ActiveKey", func(t *testing.T) {
		key := &APIKey{
			RevokedAt: nil,
			ExpiresAt: nil,
		}
		assert.True(t, key.IsActive())
	})

	t.Run("RevokedKey", func(t *testing.T) {
		yesterday := now.Add(-24 * time.Hour)
		key := &APIKey{
			RevokedAt: &yesterday,
			ExpiresAt: nil,
		}
		assert.False(t, key.IsActive())
	})

	t.Run("ExpiredKey", func(t *testing.T) {
		yesterday := now.Add(-24 * time.Hour)
		key := &APIKey{
			RevokedAt: nil,
			ExpiresAt: &yesterday,
		}
		assert.False(t, key.IsActive())
	})

	t.Run("FutureExpirationActive", func(t *testing.T) {
		tomorrow := now.Add(24 * time.Hour)
		key := &APIKey{
			RevokedAt: nil,
			ExpiresAt: &tomorrow,
		}
		assert.True(t, key.IsActive())
	})

	t.Run("RevokedAndExpiredKey", func(t *testing.T) {
		yesterday := now.Add(-24 * time.Hour)
		key := &APIKey{
			RevokedAt: &yesterday,
			ExpiresAt: &yesterday,
		}
		assert.False(t, key.IsActive())
	})
}

func TestAPIKeyConstants(t *testing.T) {
	t.Run("ConstantsAreValid", func(t *testing.T) {
		assert.Equal(t, "noot_", APIKeyPrefix)
		assert.Equal(t, 20, APIKeySecretLength)
		assert.Equal(t, 8, APIKeyPrefixLength)
		assert.Equal(t, 12, bcryptCost)
	})
}

func BenchmarkGenerateAPIKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, err := GenerateAPIKey()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyAPIKey(b *testing.B) {
	fullKey, _, hash, err := GenerateAPIKey()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyAPIKey(fullKey, hash)
	}
}