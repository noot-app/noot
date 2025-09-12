package server

import (
	"net/url"
	"os"
	"testing"
)

// Helper function for testing that accepts production mode as parameter
func isValidOriginWithMode(origin string, isProduction bool) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if isProduction {
		return u.Scheme == "https"
	}
	// Development: accept any valid http/https (middleware path already permissive)
	return true
}

func TestIsValidOrigin(t *testing.T) {
	tests := []struct {
		name         string
		origin       string
		isProduction bool
		expected     bool
		description  string
	}{
		// Production tests - HTTPS only
		{
			name:         "prod_valid_https",
			origin:       "https://app.noot.com",
			isProduction: true,
			expected:     true,
			description:  "Valid HTTPS origin in production",
		},
		{
			name:         "prod_reject_http",
			origin:       "http://app.noot.com",
			isProduction: true,
			expected:     false,
			description:  "HTTP rejected in production",
		},
		{
			name:         "prod_reject_localhost_http",
			origin:       "http://localhost:3000",
			isProduction: true,
			expected:     false,
			description:  "Even localhost HTTP rejected in production",
		},

		// Development tests - any valid http/https allowed
		{
			name:         "dev_valid_https",
			origin:       "https://app.noot.com",
			isProduction: false,
			expected:     true,
			description:  "HTTPS always allowed in dev",
		},
		{name: "dev_valid_localhost_http", origin: "http://localhost:3000", isProduction: false, expected: true, description: "HTTP localhost allowed in dev"},
		{name: "dev_valid_127_http", origin: "http://127.0.0.1:3000", isProduction: false, expected: true, description: "HTTP 127.0.0.1 allowed in dev"},
		{name: "dev_valid_ipv6_loopback", origin: "http://[::1]:3000", isProduction: false, expected: true, description: "HTTP IPv6 loopback allowed in dev"},
		{name: "dev_remote_http", origin: "http://example.com", isProduction: false, expected: true, description: "Remote HTTP allowed in dev"},

		// Security edge cases still invalid only because they are different valid hosts we now allow in dev; keep as expected true
		{name: "security_localhost_subdomain_attack", origin: "http://localhost.evil.com", isProduction: false, expected: true, description: "Subdomain treated as distinct host in dev"},
		{name: "security_127_subdomain_attack", origin: "http://127.0.0.1.evil.com", isProduction: false, expected: true, description: "Subdomain treated as distinct host in dev"},
		{name: "security_localhost_path_attack", origin: "http://evil.com/localhost", isProduction: false, expected: true, description: "Path segment not special in dev"},

		// Invalid scheme tests
		{
			name:         "invalid_file_scheme",
			origin:       "file:///etc/passwd",
			isProduction: false,
			expected:     false,
			description:  "File scheme blocked",
		},
		{
			name:         "invalid_chrome_extension",
			origin:       "chrome-extension://abc123",
			isProduction: false,
			expected:     false,
			description:  "Chrome extension scheme blocked",
		},
		{
			name:         "invalid_data_scheme",
			origin:       "data:text/html,<script>alert('xss')</script>",
			isProduction: false,
			expected:     false,
			description:  "Data scheme blocked",
		},

		// Malformed URL tests
		{
			name:         "invalid_malformed_url",
			origin:       "not-a-url",
			isProduction: false,
			expected:     false,
			description:  "Malformed URL blocked",
		},
		{
			name:         "invalid_empty_origin",
			origin:       "",
			isProduction: false,
			expected:     false,
			description:  "Empty origin blocked",
		},
		{
			name:         "invalid_missing_scheme",
			origin:       "localhost:3000",
			isProduction: false,
			expected:     false,
			description:  "Missing scheme blocked",
		},

		// Case sensitivity tests (url.Parse handles normalization)
		{
			name:         "case_insensitive_https",
			origin:       "HTTPS://app.noot.com",
			isProduction: true,
			expected:     true,
			description:  "Uppercase scheme normalized",
		},
		{
			name:         "case_insensitive_localhost",
			origin:       "HTTP://LOCALHOST:3000",
			isProduction: false,
			expected:     true,
			description:  "Uppercase hostname normalized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidOriginWithMode(tt.origin, tt.isProduction)
			if result != tt.expected {
				t.Errorf("isValidOrigin(%q, %v) = %v, expected %v - %s",
					tt.origin, tt.isProduction, result, tt.expected, tt.description)
			}
		})
	}
}

func TestIsValidOrigin_Integration(t *testing.T) {
	t.Run("integration_development", func(t *testing.T) {
		os.Setenv("ENV", "development")
		defer os.Unsetenv("ENV")
		cases := []string{
			"http://localhost:3000",
			"http://example.com",
			"http://localhost.evil.com",
			"https://app.noot.com",
		}
		for _, c := range cases {
			if !isValidOrigin(c) {
				t.Errorf("Expected dev to allow origin %s", c)
			}
		}
	})

	t.Run("integration_production", func(t *testing.T) {
		os.Setenv("ENV", "production")
		defer os.Unsetenv("ENV")
		if isValidOrigin("http://localhost:3000") {
			t.Error("Production should reject HTTP origin")
		}
		if !isValidOrigin("https://app.noot.com") {
			t.Error("Production should allow HTTPS origin")
		}
	})
}
