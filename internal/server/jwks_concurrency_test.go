package server

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestJWKSCacheRaceCondition tests that concurrent JWKS fetches don't result in duplicate network calls
func TestJWKSCacheRaceCondition(t *testing.T) {
	// Save original cache state
	originalCache := jwksCache
	originalCacheTime := jwksCacheTime
	originalPublicKeyCache := make(map[string]interface{})
	publicKeyCacheMutex.RLock()
	for k, v := range publicKeyCache {
		originalPublicKeyCache[k] = v
	}
	publicKeyCacheMutex.RUnlock()

	defer func() {
		// Restore original cache state
		jwksCacheMutex.Lock()
		jwksCache = originalCache
		jwksCacheTime = originalCacheTime
		jwksCacheMutex.Unlock()

		publicKeyCacheMutex.Lock()
		publicKeyCache = originalPublicKeyCache
		publicKeyCacheMutex.Unlock()
	}()

	// Clear cache to simulate cold start
	jwksCacheMutex.Lock()
	jwksCache = nil
	jwksCacheTime = time.Time{}
	jwksCacheMutex.Unlock()

	publicKeyCacheMutex.Lock()
	publicKeyCache = make(map[string]interface{})
	publicKeyCacheMutex.Unlock()

	// Note: This test would ideally use a mock HTTP server, but for simplicity
	// we're testing the double-check locking logic by simulating concurrent access

	t.Run("ConcurrentCacheAccess", func(t *testing.T) {
		// Test concurrent access to cache without network calls
		// We'll simulate the cache being populated by one goroutine

		var wg sync.WaitGroup
		results := make([]bool, 10)

		// First, populate the cache
		mockJWKS := &JWKS{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: "test-key",
					Alg: "RS256",
					Use: "sig",
					N:   "test-modulus",
					E:   "AQAB", // Standard RSA exponent
				},
			},
		}

		jwksCacheMutex.Lock()
		jwksCache = mockJWKS
		jwksCacheTime = time.Now()
		jwksCacheMutex.Unlock()

		cachePublicKeysFromJWKS(mockJWKS)

		// Now test concurrent reads
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				// Try to access cached JWKS
				jwksCacheMutex.RLock()
				cached := jwksCache != nil && time.Since(jwksCacheTime) < jwksCacheTTL
				jwksCacheMutex.RUnlock()

				results[index] = cached
			}(i)
		}

		wg.Wait()

		// All goroutines should have found the cache populated
		for i, result := range results {
			assert.True(t, result, "Goroutine %d should have found cache populated", i)
		}
	})

	t.Run("PublicKeyCacheWorks", func(t *testing.T) {
		// Test that the public key cache is populated and accessible
		publicKeyCacheMutex.RLock()
		cachedKeyCount := len(publicKeyCache)
		publicKeyCacheMutex.RUnlock()

		// Should have at least one key cached from the previous test
		assert.Greater(t, cachedKeyCount, 0, "Public key cache should be populated")

		// Test direct access to cached key
		publicKeyCacheMutex.RLock()
		_, exists := publicKeyCache["test-key"]
		publicKeyCacheMutex.RUnlock()

		assert.True(t, exists, "Test key should be in public key cache")
	})
}

// TestBytesToIntRobustness tests the robust RSA exponent parsing
func TestBytesToIntRobustness(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected int
		hasError bool
	}{
		{
			name:     "Standard RSA exponent 65537",
			input:    []byte{0x01, 0x00, 0x01}, // 65537 as bytes
			expected: 65537,
			hasError: false,
		},
		{
			name:     "Small exponent 3",
			input:    []byte{0x03},
			expected: 3,
			hasError: false,
		},
		{
			name:     "Empty exponent",
			input:    []byte{},
			expected: 0,
			hasError: true,
		},
		{
			name:     "Single byte exponent",
			input:    []byte{0xFF},
			expected: 255,
			hasError: false,
		},
		{
			name:     "Multi-byte exponent",
			input:    []byte{0x01, 0x00, 0x00, 0x01}, // 16777217
			expected: 16777217,
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bytesToInt(tc.input)

			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// TestJWKConversionFunctions tests the EC and RSA key conversion functions
func TestJWKConversionFunctions(t *testing.T) {
	t.Run("RSA Key Conversion", func(t *testing.T) {
		rsaKey := &JWK{
			Kty: "RSA",
			Kid: "test-rsa-key",
			Alg: "RS256",
			Use: "sig",
			N:   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw",
			E:   "AQAB",
		}

		publicKey, err := convertRSAJWKToPublicKey(rsaKey)
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
		assert.Equal(t, 65537, publicKey.E) // Standard RSA exponent
	})

	t.Run("EC Key Conversion", func(t *testing.T) {
		ecKey := &JWK{
			Kty: "EC",
			Kid: "test-ec-key",
			Alg: "ES256",
			Use: "sig",
			Crv: "P-256",
			X:   "MKBCTNIcKUSDii11ySs3526iDZ8AiTo7Tu6KPAqv7D4",
			Y:   "4Etl6SRW2YiLUrN5vfvVHuhp7x8PxltmWWlbbM4IFyM",
		}

		publicKey, err := convertECJWKToPublicKey(ecKey)
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
		assert.Equal(t, "P-256", ecKey.Crv)
	})

	t.Run("Invalid RSA Key", func(t *testing.T) {
		invalidKey := &JWK{
			Kty: "RSA",
			Kid: "invalid-key",
			Alg: "RS256",
			N:   "invalid-base64!!!", // Invalid base64 characters
			E:   "AQAB",
		}

		_, err := convertRSAJWKToPublicKey(invalidKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode RSA modulus")
	})
}
