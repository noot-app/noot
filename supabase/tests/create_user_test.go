package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	authURL    = "http://localhost:54321/auth/v1/signup"
	anonAPIKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0"
)

type SignupRequest struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Data     map[string]interface{} `json:"data"`
}

type SignupResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type User struct {
	ID           string                 `json:"id"`
	Aud          string                 `json:"aud"`
	Role         string                 `json:"role"`
	Email        string                 `json:"email"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

type ErrorResponse struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"error_code"`
	Msg       string `json:"msg"`
	ErrorID   string `json:"error_id"`
}

type ProfileRow struct {
	ID               string `json:"id"`
	Handle           string `json:"handle"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	SubscriptionTier string `json:"subscription_tier"`
}

func TestUserCreationFlow(t *testing.T) {
	// Test data
	testEmail := "test@birki.io"
	testPassword := "testpassword123"
	testHandle := "test-user"
	testFullName := "Test User"

	t.Run("Database Reset", func(t *testing.T) {
		// Reset the database to ensure clean state
		cmd := exec.Command("script/db", "reset")
		cmd.Dir = getProjectRoot()

		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Database reset should succeed")

		// Verify reset completed successfully
		outputStr := string(output)
		assert.Contains(t, outputStr, "Database reset completed", "Reset should complete successfully")
		assert.Contains(t, outputStr, "Applying migration 20250825000001_profiles.sql", "User migration should be applied")

		// Wait a moment for containers to fully restart
		time.Sleep(2 * time.Second)
	})

	t.Run("User Signup Success", func(t *testing.T) {
		signupReq := SignupRequest{
			Email:    testEmail,
			Password: testPassword,
			Data: map[string]interface{}{
				"handle":    testHandle,
				"full_name": testFullName,
			},
		}

		// Make signup request
		resp, err := makeSignupRequest(signupReq)
		require.NoError(t, err, "Signup request should not fail")
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read response body")

		// Assert successful response
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Signup should return 200 OK")

		// Parse successful response
		var signupResp SignupResponse
		err = json.Unmarshal(body, &signupResp)
		require.NoError(t, err, "Should be able to parse signup response")

		// Validate response structure
		assert.NotEmpty(t, signupResp.AccessToken, "Access token should be provided")
		assert.Equal(t, "bearer", signupResp.TokenType, "Token type should be bearer")
		assert.Greater(t, signupResp.ExpiresIn, 0, "Expires in should be positive")
		assert.NotEmpty(t, signupResp.RefreshToken, "Refresh token should be provided")

		// Validate user data
		assert.NotEmpty(t, signupResp.User.ID, "User ID should be provided")
		assert.Equal(t, "authenticated", signupResp.User.Role, "User role should be authenticated")
		assert.Equal(t, testEmail, signupResp.User.Email, "User email should match")

		// Validate user metadata
		assert.Equal(t, testHandle, signupResp.User.UserMetadata["handle"], "Handle should match")
		assert.Equal(t, testFullName, signupResp.User.UserMetadata["full_name"], "Full name should match")
	})

	t.Run("Profile Created in Database", func(t *testing.T) {
		// Query the profiles table to verify trigger worked
		query := fmt.Sprintf("SELECT id, handle, full_name, email, subscription_tier FROM public.profiles WHERE email = '%s';", testEmail)
		cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
		cmd.Dir = getProjectRoot()

		output, err := cmd.Output()
		require.NoError(t, err, "Database query should succeed")

		outputStr := strings.TrimSpace(string(output))
		assert.NotEmpty(t, outputStr, "Profile should exist in database")

		// Parse the psql output (pipe-separated values)
		parts := strings.Split(outputStr, "|")
		require.Len(t, parts, 5, "Should have 5 columns in result")

		// Trim whitespace from each part
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		// Validate profile data
		assert.NotEmpty(t, parts[0], "Profile ID should exist")
		assert.Equal(t, testHandle, parts[1], "Handle should match")
		assert.Equal(t, testFullName, parts[2], "Full name should match")
		assert.Equal(t, testEmail, parts[3], "Email should match")
		assert.Equal(t, "free", parts[4], "Subscription tier should default to free")
	})

	t.Run("Auth Logs Show Success", func(t *testing.T) {
		// Check auth container logs for successful signup
		cmd := exec.Command("docker", "logs", "supabase_auth_noot")
		cmd.Dir = getProjectRoot()

		output, err := cmd.CombinedOutput() // Use CombinedOutput to get both stdout and stderr
		require.NoError(t, err, "Should be able to read auth logs")

		outputStr := string(output)

		// Look for successful completion in recent logs
		assert.Contains(t, outputStr, `"status":200`, "Auth logs should show successful request")
		assert.Contains(t, outputStr, `"action":"login"`, "Auth logs should show login event")
		assert.Contains(t, outputStr, `"msg":"request completed"`, "Auth logs should show completed request")
		assert.Contains(t, outputStr, `"path":"/signup"`, "Auth logs should show signup path")
	})

	t.Run("Duplicate Signup Fails", func(t *testing.T) {
		// Try to signup with the same email again
		signupReq := SignupRequest{
			Email:    testEmail,
			Password: testPassword,
			Data: map[string]interface{}{
				"handle":    "duplicate-user",
				"full_name": "Duplicate User",
			},
		}

		// Make duplicate signup request
		resp, err := makeSignupRequest(signupReq)
		require.NoError(t, err, "Duplicate signup request should not fail at HTTP level")
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "Should be able to read response body")

		// Assert error response
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "Duplicate signup should return 422")

		// Parse error response
		var errorResp ErrorResponse
		err = json.Unmarshal(body, &errorResp)
		require.NoError(t, err, "Should be able to parse error response")

		// Validate error details
		assert.Equal(t, 422, errorResp.Code, "Error code should be 422")
		assert.Contains(t, strings.ToLower(errorResp.Msg), "already", "Error message should mention user already exists")
	})

	t.Run("Profile Count Remains One", func(t *testing.T) {
		// Verify only one profile exists for this email
		query := fmt.Sprintf("SELECT COUNT(*) FROM public.profiles WHERE email = '%s';", testEmail)
		cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
		cmd.Dir = getProjectRoot()

		output, err := cmd.Output()
		require.NoError(t, err, "Database count query should succeed")

		count := strings.TrimSpace(string(output))
		assert.Equal(t, "1", count, "Should have exactly one profile for the email")
	})

	t.Run("Seeded Labels Exist", func(t *testing.T) {
		// Check that labels were properly seeded for test users
		t.Run("Monalisa Labels", func(t *testing.T) {
			// Query labels for monalisa@birki.io
			query := `SELECT name, description, color FROM public.labels l 
				JOIN public.profiles p ON l.user_id = p.id 
				WHERE p.email = 'monalisa@birki.io' 
				ORDER BY name;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query Monalisa's labels")

			outputStr := strings.TrimSpace(string(output))
			lines := strings.Split(outputStr, "\n")

			// Should have 9 default labels for Monalisa
			assert.Len(t, lines, 9, "Monalisa should have 9 default seeded labels")

			// Parse each line and check expected default labels
			expectedLabels := map[string]struct {
				description string
				color       string
			}{
				"breakfast":    {"", "FFD700"},
				"lunch":        {"", "74B986"},
				"dinner":       {"", "1E90FF"},
				"snack":        {"A small snack or light bite", "9B59B6"},
				"drink":        {"A beverage", "9CA3AF"},
				"trigger-food": {"The consumption contained a known trigger food", "FF0000"},
				"high-protein": {"High protein foods", "2DD4BF"},
				"meal-prep":    {"Pre-prepared meals", "A8E6CF"},
				"restaurant":   {"Restaurant or takeout meal", "F59E0B"},
			}

			foundLabels := make(map[string]bool)
			for _, line := range lines {
				parts := strings.Split(line, "|")
				require.Len(t, parts, 3, "Each label line should have 3 parts")

				name := strings.TrimSpace(parts[0])
				description := strings.TrimSpace(parts[1])
				color := strings.TrimSpace(parts[2])

				if expected, exists := expectedLabels[name]; exists {
					foundLabels[name] = true
					assert.Equal(t, expected.description, description, fmt.Sprintf("Description should match for label %s", name))
					assert.Equal(t, expected.color, color, fmt.Sprintf("Color should match for label %s", name))
				}
			}

			// Verify all expected labels were found
			for labelName := range expectedLabels {
				assert.True(t, foundLabels[labelName], fmt.Sprintf("Label %s should be found", labelName))
			}
		})

		t.Run("Alice Labels", func(t *testing.T) {
			// Query labels for alice@birki.io
			query := `SELECT name, description, color FROM public.labels l 
				JOIN public.profiles p ON l.user_id = p.id 
				WHERE p.email = 'alice@birki.io' 
				ORDER BY name;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query Alice's labels")

			outputStr := strings.TrimSpace(string(output))
			lines := strings.Split(outputStr, "\n")

			// Should have 9 default labels for Alice
			assert.Len(t, lines, 9, "Alice should have 9 default seeded labels")

			// Parse each line and check expected default labels (same as Monalisa)
			expectedLabels := map[string]struct {
				description string
				color       string
			}{
				"breakfast":    {"", "FFD700"},
				"lunch":        {"", "74B986"},
				"dinner":       {"", "1E90FF"},
				"snack":        {"A small snack or light bite", "9B59B6"},
				"drink":        {"A beverage", "9CA3AF"},
				"trigger-food": {"The consumption contained a known trigger food", "FF0000"},
				"high-protein": {"High protein foods", "2DD4BF"},
				"meal-prep":    {"Pre-prepared meals", "A8E6CF"},
				"restaurant":   {"Restaurant or takeout meal", "F59E0B"},
			}

			foundLabels := make(map[string]bool)
			for _, line := range lines {
				parts := strings.Split(line, "|")
				require.Len(t, parts, 3, "Each label line should have 3 parts")

				name := strings.TrimSpace(parts[0])
				description := strings.TrimSpace(parts[1])
				color := strings.TrimSpace(parts[2])

				if expected, exists := expectedLabels[name]; exists {
					foundLabels[name] = true
					assert.Equal(t, expected.description, description, fmt.Sprintf("Description should match for label %s", name))
					assert.Equal(t, expected.color, color, fmt.Sprintf("Color should match for label %s", name))
				}
			}

			// Verify all expected labels were found
			for labelName := range expectedLabels {
				assert.True(t, foundLabels[labelName], fmt.Sprintf("Label %s should be found", labelName))
			}
		})

		t.Run("Label Constraints", func(t *testing.T) {
			// Test that labels table has proper constraints and indexes
			constraintQuery := `SELECT conname, contype FROM pg_constraint 
				WHERE conrelid = 'public.labels'::regclass 
				ORDER BY conname;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", constraintQuery)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query label constraints")

			outputStr := strings.TrimSpace(string(output))

			// Check for key constraints
			assert.Contains(t, outputStr, "labels_color_hex", "Should have color hex constraint")
			assert.Contains(t, outputStr, "labels_description_len", "Should have description length constraint")
			assert.Contains(t, outputStr, "labels_pkey", "Should have primary key constraint")
		})
	})
}

// makeSignupRequest makes an HTTP POST request to the Supabase signup endpoint
func makeSignupRequest(req SignupRequest) (*http.Response, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", anonAPIKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return client.Do(httpReq)
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Try to find the project root by looking for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(fmt.Sprintf("%s/go.mod", dir)); err == nil {
			return dir
		}

		parent := fmt.Sprintf("%s/..", dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "."
}
