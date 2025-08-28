package server

import (
	"net/url"
	"os"
	"strings"
	"testing"
)

// Helper function for testing that accepts production mode as parameter
func isValidOriginWithMode(origin string, isProduction bool) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Must be http or https scheme only
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// In production, require HTTPS only
	if isProduction {
		return u.Scheme == "https"
	}

	// In development, allow HTTP only for localhost/loopback addresses
	if u.Scheme == "http" {
		hostname := strings.ToLower(u.Hostname()) // Case-insensitive hostname check
		// Allow localhost, 127.0.0.1, and IPv6 loopback
		return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
	}

	// HTTPS is always allowed in development
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

		// Development tests - HTTP allowed for localhost/loopback
		{
			name:         "dev_valid_https",
			origin:       "https://app.noot.com",
			isProduction: false,
			expected:     true,
			description:  "HTTPS always allowed in dev",
		},
		{
			name:         "dev_valid_localhost_http",
			origin:       "http://localhost:3000",
			isProduction: false,
			expected:     true,
			description:  "HTTP localhost allowed in dev",
		},
		{
			name:         "dev_valid_127_http",
			origin:       "http://127.0.0.1:3000",
			isProduction: false,
			expected:     true,
			description:  "HTTP 127.0.0.1 allowed in dev",
		},
		{
			name:         "dev_valid_ipv6_loopback",
			origin:       "http://[::1]:3000",
			isProduction: false,
			expected:     true,
			description:  "HTTP IPv6 loopback allowed in dev",
		},
		{
			name:         "dev_reject_remote_http",
			origin:       "http://example.com",
			isProduction: false,
			expected:     false,
			description:  "HTTP to remote hosts rejected in dev",
		},

		// Security vulnerability tests (these would pass with old substring method)
		{
			name:         "security_localhost_subdomain_attack",
			origin:       "http://localhost.evil.com",
			isProduction: false,
			expected:     false,
			description:  "Subdomain attack with localhost blocked",
		},
		{
			name:         "security_127_subdomain_attack",
			origin:       "http://127.0.0.1.evil.com",
			isProduction: false,
			expected:     false,
			description:  "Subdomain attack with 127.0.0.1 blocked",
		},
		{
			name:         "security_localhost_path_attack",
			origin:       "http://evil.com/localhost",
			isProduction: false,
			expected:     false,
			description:  "Path-based attack with localhost blocked",
		},

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
	// Test the actual function by setting environment variables
	t.Run("integration_development", func(t *testing.T) {
		os.Setenv("ENV", "development")
		defer os.Unsetenv("ENV")

		// Should allow localhost HTTP in development
		if !isValidOrigin("http://localhost:3000") {
			t.Error("Development should allow localhost HTTP")
		}

		// Should block subdomain attacks in development
		if isValidOrigin("http://localhost.evil.com") {
			t.Error("Development should block subdomain attacks")
		}
	})

	t.Run("integration_production", func(t *testing.T) {
		os.Setenv("ENV", "production")
		defer os.Unsetenv("ENV")

		// Should reject all HTTP in production
		if isValidOrigin("http://localhost:3000") {
			t.Error("Production should reject all HTTP")
		}

		// Should allow HTTPS in production
		if !isValidOrigin("https://app.noot.com") {
			t.Error("Production should allow HTTPS")
		}
	})
}
