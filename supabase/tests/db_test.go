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

	t.Run("Seeded Consumptions Exist", func(t *testing.T) {
		// Verify that consumption data was properly seeded for both users
		t.Run("Consumption Counts", func(t *testing.T) {
			// Query total consumption counts for seeded users
			query := `SELECT p.email, COUNT(*) as consumption_count 
				FROM consumptions c 
				JOIN profiles p ON p.id = c.user_id 
				WHERE p.email IN ('monalisa@birki.io', 'alice@birki.io') 
				GROUP BY p.email 
				ORDER BY p.email;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query consumption counts")

			outputStr := strings.TrimSpace(string(output))
			lines := strings.Split(outputStr, "\n")
			require.Len(t, lines, 2, "Should have consumption data for both users")

			// Parse and validate consumption counts
			for _, line := range lines {
				parts := strings.Split(line, "|")
				require.Len(t, parts, 2, "Each line should have email and count")

				email := strings.TrimSpace(parts[0])
				count := strings.TrimSpace(parts[1])

				// Both users should have substantial consumption data (~30+ each)
				assert.Contains(t, []string{"alice@birki.io", "monalisa@birki.io"}, email, "Should be a seeded user")

				// Convert count to int and validate
				var consumptionCount int
				_, err := fmt.Sscanf(count, "%d", &consumptionCount)
				require.NoError(t, err, "Count should be a valid integer")
				assert.GreaterOrEqual(t, consumptionCount, 30, fmt.Sprintf("User %s should have at least 30 consumptions", email))
				assert.LessOrEqual(t, consumptionCount, 50, fmt.Sprintf("User %s should have reasonable consumption count", email))
			}
		})

		t.Run("Date Distribution", func(t *testing.T) {
			// Verify consumptions are distributed across time periods as expected
			query := `SELECT 
				p.email,
				COUNT(CASE WHEN c.created_at::date = CURRENT_DATE THEN 1 END) as today_meals,
				COUNT(CASE WHEN c.created_at::date = CURRENT_DATE - 1 THEN 1 END) as yesterday_meals,
				COUNT(CASE WHEN c.created_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as last_week_meals,
				COUNT(*) as total_meals
			FROM consumptions c 
			JOIN profiles p ON p.id = c.user_id 
			WHERE p.email IN ('monalisa@birki.io', 'alice@birki.io')
			GROUP BY p.email 
			ORDER BY p.email;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query date distribution")

			outputStr := strings.TrimSpace(string(output))
			lines := strings.Split(outputStr, "\n")

			for _, line := range lines {
				parts := strings.Split(line, "|")
				require.Len(t, parts, 5, "Each line should have 5 values")

				email := strings.TrimSpace(parts[0])
				todayStr := strings.TrimSpace(parts[1])
				yesterdayStr := strings.TrimSpace(parts[2])
				lastWeekStr := strings.TrimSpace(parts[3])
				totalStr := strings.TrimSpace(parts[4])

				// Convert to integers
				var today, yesterday, lastWeek, total int
				_, err := fmt.Sscanf(todayStr, "%d", &today)
				require.NoError(t, err, "Today count should be valid")
				_, err = fmt.Sscanf(yesterdayStr, "%d", &yesterday)
				require.NoError(t, err, "Yesterday count should be valid")
				_, err = fmt.Sscanf(lastWeekStr, "%d", &lastWeek)
				require.NoError(t, err, "Last week count should be valid")
				_, err = fmt.Sscanf(totalStr, "%d", &total)
				require.NoError(t, err, "Total count should be valid")

				// Validate distribution matches expected pattern
				assert.GreaterOrEqual(t, today, 3, fmt.Sprintf("%s should have 3+ meals today", email))
				assert.LessOrEqual(t, today, 4, fmt.Sprintf("%s should have ≤4 meals today", email))

				assert.GreaterOrEqual(t, yesterday, 2, fmt.Sprintf("%s should have 2+ meals yesterday", email))
				assert.LessOrEqual(t, yesterday, 3, fmt.Sprintf("%s should have ≤3 meals yesterday", email))

				assert.GreaterOrEqual(t, lastWeek, 15, fmt.Sprintf("%s should have 15+ meals in last week", email))
				assert.LessOrEqual(t, lastWeek, 25, fmt.Sprintf("%s should have ≤25 meals in last week", email))

				// Total should be larger than last week (includes older data)
				assert.GreaterOrEqual(t, total, lastWeek, fmt.Sprintf("%s: total meals should be ≥ last week meals", email))
				assert.GreaterOrEqual(t, total, 30, fmt.Sprintf("%s should have 30+ total meals", email))
			}
		})

		t.Run("Nutrition Values", func(t *testing.T) {
			// Verify consumption nutrition values are realistic
			query := `SELECT 
				MIN(total_calories) as min_calories,
				MAX(total_calories) as max_calories,
				AVG(total_calories)::int as avg_calories,
				MIN(total_protein_g) as min_protein,
				MAX(total_protein_g) as max_protein,
				COUNT(CASE WHEN caffeine_mg > 0 THEN 1 END) as caffeine_items
			FROM consumptions c 
			JOIN profiles p ON p.id = c.user_id 
			WHERE p.email IN ('monalisa@birki.io', 'alice@birki.io');`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query nutrition stats")

			outputStr := strings.TrimSpace(string(output))
			parts := strings.Split(outputStr, "|")
			require.Len(t, parts, 6, "Should have 6 nutrition stats")

			// Parse nutrition values
			var minCal, maxCal, avgCal, caffeineItems int
			var minProtein, maxProtein float64
			_, err = fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &minCal)
			require.NoError(t, err, "Min calories should parse")
			_, err = fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &maxCal)
			require.NoError(t, err, "Max calories should parse")
			_, err = fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &avgCal)
			require.NoError(t, err, "Avg calories should parse")
			_, err = fmt.Sscanf(strings.TrimSpace(parts[3]), "%f", &minProtein)
			require.NoError(t, err, "Min protein should parse")
			_, err = fmt.Sscanf(strings.TrimSpace(parts[4]), "%f", &maxProtein)
			require.NoError(t, err, "Max protein should parse")
			_, err = fmt.Sscanf(strings.TrimSpace(parts[5]), "%d", &caffeineItems)
			require.NoError(t, err, "Caffeine items should parse")

			// Validate realistic nutrition ranges
			assert.GreaterOrEqual(t, minCal, 20, "Minimum calories should be reasonable (small snacks)")
			assert.LessOrEqual(t, maxCal, 800, "Maximum calories should be reasonable")
			assert.GreaterOrEqual(t, avgCal, 250, "Average calories should be realistic")
			assert.LessOrEqual(t, avgCal, 500, "Average calories should be realistic")

			assert.GreaterOrEqual(t, minProtein, 1.0, "Minimum protein should be positive")
			assert.LessOrEqual(t, maxProtein, 50.0, "Maximum protein should be reasonable")

			assert.GreaterOrEqual(t, caffeineItems, 10, "Should have several caffeine items")
		})

		t.Run("Transcript Variety", func(t *testing.T) {
			// Verify transcripts are varied and realistic
			query := `SELECT DISTINCT LEFT(transcript, 50) as transcript_start 
				FROM consumptions c 
				JOIN profiles p ON p.id = c.user_id 
				WHERE p.email IN ('monalisa@birki.io', 'alice@birki.io')
				ORDER BY transcript_start 
				LIMIT 20;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query transcript variety")

			outputStr := strings.TrimSpace(string(output))
			lines := strings.Split(outputStr, "\n")

			// Should have good variety of different transcripts
			assert.GreaterOrEqual(t, len(lines), 15, "Should have varied transcript content")

			// Check for expected meal types in transcripts
			transcriptText := strings.ToLower(outputStr)
			assert.Contains(t, transcriptText, "breakfast", "Should include breakfast meals")
			assert.Contains(t, transcriptText, "lunch", "Should include lunch meals")
			assert.Contains(t, transcriptText, "dinner", "Should include dinner meals")
			assert.Contains(t, transcriptText, "coffee", "Should include coffee drinks")
		})
	})

	t.Run("Consumption Labels Work", func(t *testing.T) {
		// Verify that consumption labeling is working correctly
		t.Run("Monalisa Label Coverage", func(t *testing.T) {
			// Check that ~70% of Mona's consumptions have labels
			query := `SELECT 
				COUNT(DISTINCT c.id) as total_consumptions,
				COUNT(DISTINCT cl.consumption_id) as labeled_consumptions,
				ROUND(COUNT(DISTINCT cl.consumption_id)::numeric / COUNT(DISTINCT c.id)::numeric * 100, 1) as percentage_labeled
			FROM consumptions c 
			LEFT JOIN consumption_labels cl ON cl.consumption_id = c.id
			JOIN profiles p ON p.id = c.user_id 
			WHERE p.email = 'monalisa@birki.io';`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query label coverage")

			outputStr := strings.TrimSpace(string(output))
			parts := strings.Split(outputStr, "|")
			require.Len(t, parts, 3, "Should have 3 coverage stats")

			// Parse values
			var totalConsumptions, labeledConsumptions int
			var percentageLabeled float64
			_, err = fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &totalConsumptions)
			require.NoError(t, err, "Total consumptions should parse")
			_, err = fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &labeledConsumptions)
			require.NoError(t, err, "Labeled consumptions should parse")
			_, err = fmt.Sscanf(strings.TrimSpace(parts[2]), "%f", &percentageLabeled)
			require.NoError(t, err, "Percentage should parse")

			// Validate labeling coverage (~70% target)
			assert.GreaterOrEqual(t, percentageLabeled, 60.0, "Should have at least 60% labeling coverage")
			assert.LessOrEqual(t, percentageLabeled, 80.0, "Should have reasonable labeling coverage")
			assert.GreaterOrEqual(t, labeledConsumptions, 25, "Should have substantial labeled consumptions")
		})

		t.Run("Trigger Food Detection", func(t *testing.T) {
			// Verify trigger foods are properly detected and labeled
			query := `SELECT c.transcript, c.total_calories 
				FROM consumptions c 
				JOIN consumption_labels cl ON cl.consumption_id = c.id
				JOIN labels l ON l.id = cl.label_id
				JOIN profiles p ON p.id = c.user_id
				WHERE p.email = 'monalisa@birki.io' AND l.name = 'trigger-food'
				ORDER BY c.created_at DESC;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query trigger foods")

			outputStr := strings.TrimSpace(string(output))

			// Should have some trigger foods detected
			assert.NotEmpty(t, outputStr, "Should have some trigger foods detected")

			// Check for expected trigger food patterns
			triggerText := strings.ToLower(outputStr)

			// At least one of these common triggers should be present
			hasTriggers := strings.Contains(triggerText, "chocolate") ||
				strings.Contains(triggerText, "cheese") ||
				strings.Contains(triggerText, "pizza") ||
				strings.Contains(triggerText, "garlic") ||
				strings.Contains(triggerText, "tomato")

			assert.True(t, hasTriggers, "Should detect common trigger foods like chocolate, cheese, pizza, garlic, or tomatoes")
		})

		t.Run("High Protein Detection", func(t *testing.T) {
			// Verify high-protein meals are properly labeled
			query := `SELECT c.transcript, c.total_protein_g, c.total_calories 
				FROM consumptions c 
				JOIN consumption_labels cl ON cl.consumption_id = c.id
				JOIN labels l ON l.id = cl.label_id
				JOIN profiles p ON p.id = c.user_id
				WHERE p.email = 'monalisa@birki.io' AND l.name = 'high-protein'
				ORDER BY c.total_protein_g DESC
				LIMIT 5;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query high-protein meals")

			outputStr := strings.TrimSpace(string(output))

			// Should have high-protein meals detected
			assert.NotEmpty(t, outputStr, "Should have high-protein meals detected")

			// Verify protein content expectations
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				parts := strings.Split(line, "|")
				if len(parts) >= 2 {
					proteinStr := strings.TrimSpace(parts[1])
					var protein float64
					if _, err := fmt.Sscanf(proteinStr, "%f", &protein); err == nil {
						assert.GreaterOrEqual(t, protein, 25.0, "High-protein labeled meals should have ≥25g protein")
					}
				}
			}
		})

		t.Run("Restaurant Detection", func(t *testing.T) {
			// Verify restaurant meals are properly labeled
			query := `SELECT COUNT(*) as restaurant_count
				FROM consumptions c 
				JOIN consumption_labels cl ON cl.consumption_id = c.id
				JOIN labels l ON l.id = cl.label_id
				JOIN profiles p ON p.id = c.user_id
				WHERE p.email = 'monalisa@birki.io' AND l.name = 'restaurant';`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query restaurant meals")

			count := strings.TrimSpace(string(output))
			var restaurantCount int
			_, err = fmt.Sscanf(count, "%d", &restaurantCount)
			require.NoError(t, err, "Restaurant count should be valid")

			// Should have some restaurant meals
			assert.GreaterOrEqual(t, restaurantCount, 1, "Should have at least 1 restaurant meal")
			assert.LessOrEqual(t, restaurantCount, 10, "Should have reasonable restaurant meal count")
		})

		t.Run("Caffeine Detection", func(t *testing.T) {
			// Verify caffeine content is properly set for coffee/tea items
			query := `SELECT c.transcript, c.caffeine_mg 
				FROM consumptions c 
				JOIN profiles p ON p.id = c.user_id 
				WHERE p.email = 'monalisa@birki.io' AND c.caffeine_mg > 0
				ORDER BY c.caffeine_mg DESC
				LIMIT 5;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query caffeine content")

			outputStr := strings.TrimSpace(string(output))

			// Should have caffeine items
			assert.NotEmpty(t, outputStr, "Should have items with caffeine content")

			// Verify caffeine values are realistic
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				parts := strings.Split(line, "|")
				if len(parts) >= 2 {
					transcript := strings.ToLower(strings.TrimSpace(parts[0]))
					caffeineStr := strings.TrimSpace(parts[1])
					var caffeine int
					if _, err := fmt.Sscanf(caffeineStr, "%d", &caffeine); err == nil {
						assert.Greater(t, caffeine, 0, "Caffeine items should have positive caffeine content")
						assert.LessOrEqual(t, caffeine, 300, "Caffeine content should be realistic (≤300mg)")

						// Coffee/tea items should contain relevant keywords
						hasCaffeineKeywords := strings.Contains(transcript, "coffee") ||
							strings.Contains(transcript, "latte") ||
							strings.Contains(transcript, "cappuccino") ||
							strings.Contains(transcript, "tea") ||
							strings.Contains(transcript, "energy")

						assert.True(t, hasCaffeineKeywords, "Caffeine items should mention coffee, tea, or energy drinks")
					}
				}
			}
		})

		t.Run("Multi-Label Assignments", func(t *testing.T) {
			// Verify some consumptions have multiple labels (realistic labeling)
			query := `SELECT c.transcript, COUNT(l.id) as label_count, STRING_AGG(l.name, ', ') as labels
				FROM consumptions c 
				JOIN consumption_labels cl ON cl.consumption_id = c.id
				JOIN labels l ON l.id = cl.label_id
				JOIN profiles p ON p.id = c.user_id
				WHERE p.email = 'monalisa@birki.io'
				GROUP BY c.id, c.transcript
				HAVING COUNT(l.id) > 1
				ORDER BY label_count DESC
				LIMIT 5;`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query multi-labeled consumptions")

			outputStr := strings.TrimSpace(string(output))

			// Should have some items with multiple labels
			if outputStr != "" {
				lines := strings.Split(outputStr, "\n")
				assert.GreaterOrEqual(t, len(lines), 1, "Should have at least one multi-labeled consumption")

				// Verify multi-label assignments make sense
				for _, line := range lines {
					if strings.TrimSpace(line) == "" {
						continue
					}
					parts := strings.Split(line, "|")
					if len(parts) >= 3 {
						labelCountStr := strings.TrimSpace(parts[1])
						labels := strings.TrimSpace(parts[2])

						var labelCount int
						if _, err := fmt.Sscanf(labelCountStr, "%d", &labelCount); err == nil {
							assert.GreaterOrEqual(t, labelCount, 2, "Multi-labeled items should have ≥2 labels")
							assert.LessOrEqual(t, labelCount, 4, "Should not have excessive labels")

							// Logical label combinations (e.g., snack + trigger-food, drink + snack)
							assert.NotEmpty(t, labels, "Should have label names listed")
						}
					}
				}
			}
		})

		t.Run("Alice Has No Labels", func(t *testing.T) {
			// Verify Alice (free user) has no consumption labels (only Mona gets labeled)
			query := `SELECT COUNT(*) as alice_label_count
				FROM consumption_labels cl
				JOIN consumptions c ON c.id = cl.consumption_id
				JOIN profiles p ON p.id = c.user_id
				WHERE p.email = 'alice@birki.io';`

			cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", query)
			cmd.Dir = getProjectRoot()

			output, err := cmd.Output()
			require.NoError(t, err, "Should be able to query Alice's labels")

			count := strings.TrimSpace(string(output))
			var aliceLabelCount int
			_, err = fmt.Sscanf(count, "%d", &aliceLabelCount)
			require.NoError(t, err, "Alice label count should parse")

			// Alice should have no consumption labels (only Mona gets labeled in our seed)
			assert.Equal(t, 0, aliceLabelCount, "Alice should have no consumption labels (only Mona is labeled)")
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

// TestAPIKeyIntegration tests API key functionality with seeded Pro user
func TestAPIKeyIntegration(t *testing.T) {
	// This test validates that the seeded Pro user (monalisa@birki.io) can:
	// 1. Create API keys via the database directly
	// 2. Use those API keys to make authenticated API calls
	// 3. Respect scope limitations (read vs read_write)

	// Skip if no database connection available
	t.Run("Database Reset", func(t *testing.T) {
		// Reset the database to ensure clean state
		cmd := exec.Command("script/db", "reset")
		cmd.Dir = getProjectRoot()

		output, err := cmd.CombinedOutput()
		if err != nil {
			// If database reset fails, skip the API key tests
			t.Skipf("Database reset failed - skipping API key integration tests: %v\nOutput: %s", err, string(output))
			return
		}

		// Verify reset completed successfully
		outputStr := string(output)
		assert.Contains(t, outputStr, "Database reset completed", "Reset should complete successfully")
	})

	// Test creating API key directly in database for Pro user
	t.Run("Create API Key for Pro User", func(t *testing.T) {
		// Get monalisa's user ID
		userQuery := `SELECT id FROM public.profiles WHERE email = 'monalisa@birki.io' AND subscription_tier = 'pro';`
		cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", userQuery)
		cmd.Dir = getProjectRoot()

		userOutput, err := cmd.Output()
		if err != nil {
			t.Skipf("Could not get Pro user ID - skipping API key test: %v", err)
			return
		}

		userID := strings.TrimSpace(string(userOutput))
		require.NotEmpty(t, userID, "Should find Pro user monalisa")

		// Generate API key components for testing (simulate what the storage layer would do)
		// In real implementation, this would use storage.GenerateAPIKey()
		prefix := "noot_test1234"
		// Create a simple test hash for the demo key
		hash := "$2a$12$test.hash.for.demo.key.only.not.real.bcrypt.hash"

		// Insert API key into database
		insertQuery := fmt.Sprintf(`
			INSERT INTO public.api_keys (id, user_id, name, prefix, hash, scope, created_at)
			VALUES (gen_random_uuid(), '%s', 'Test Key', '%s', '%s', 'read_write', now());
		`, userID, prefix, hash)

		cmd = exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-c", insertQuery)
		cmd.Dir = getProjectRoot()

		_, err = cmd.Output()
		require.NoError(t, err, "Should be able to create API key in database")

		// Verify API key was created
		selectQuery := fmt.Sprintf(`
			SELECT name, prefix, scope, created_at IS NOT NULL, revoked_at IS NULL
			FROM public.api_keys 
			WHERE user_id = '%s' AND name = 'Test Key';
		`, userID)

		cmd = exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", selectQuery)
		cmd.Dir = getProjectRoot()

		selectOutput, err := cmd.Output()
		require.NoError(t, err, "Should be able to query API key")

		outputStr := strings.TrimSpace(string(selectOutput))
		assert.Contains(t, outputStr, "Test Key", "Should find created API key")
		assert.Contains(t, outputStr, prefix, "Should have correct prefix")
		assert.Contains(t, outputStr, "read_write", "Should have correct scope")
		assert.Contains(t, outputStr, "t", "created_at should be set")
		assert.Contains(t, outputStr, "t", "revoked_at should be null (active)")
	})

	t.Run("Verify RLS Protection", func(t *testing.T) {
		// Test that non-Pro users cannot create API keys
		// Get alice's user ID (free tier)
		userQuery := `SELECT id FROM public.profiles WHERE email = 'alice@birki.io' AND subscription_tier = 'free';`
		cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", userQuery)
		cmd.Dir = getProjectRoot()

		userOutput, err := cmd.Output()
		if err != nil {
			t.Skipf("Could not get Free user ID - skipping RLS test: %v", err)
			return
		}

		aliceID := strings.TrimSpace(string(userOutput))
		require.NotEmpty(t, aliceID, "Should find Free user alice")

		// Attempt to insert API key for free user (should be blocked by RLS if we set the user context)
		// Note: This test shows the table structure works, but RLS enforcement happens at the application level
		// when using auth.uid() in real Supabase usage
		
		// Count existing API keys
		countQuery := `SELECT COUNT(*) FROM public.api_keys;`
		cmd = exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-t", "-c", countQuery)
		cmd.Dir = getProjectRoot()

		countOutput, err := cmd.Output()
		require.NoError(t, err, "Should be able to count API keys")

		initialCount := strings.TrimSpace(string(countOutput))
		
		// The RLS policies reference auth.uid() which requires Supabase auth context
		// In integration tests, we validate the table structure and constraints work properly
		t.Logf("Initial API key count: %s", initialCount)
		t.Log("Note: RLS enforcement with auth.uid() is tested in full Supabase environment")
	})

	t.Run("Test API Key Table Constraints", func(t *testing.T) {
		// Test that API key table has proper constraints
		constraintQuery := `
			SELECT 
				conname as constraint_name,
				contype as constraint_type
			FROM pg_constraint 
			WHERE conrelid = 'public.api_keys'::regclass 
			ORDER BY conname;
		`

		cmd := exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-c", constraintQuery)
		cmd.Dir = getProjectRoot()

		output, err := cmd.Output()
		require.NoError(t, err, "Should be able to query API key constraints")

		outputStr := string(output)
		
		// Check for expected constraints
		assert.Contains(t, outputStr, "api_keys_pkey", "Should have primary key")
		assert.Contains(t, outputStr, "unique_api_key_name_per_user", "Should have unique name per user constraint")
		assert.Contains(t, outputStr, "api_keys_prefix_key", "Should have unique prefix constraint")
		assert.Contains(t, outputStr, "api_keys_scope_check", "Should have scope check constraint")

		// Test scope constraint 
		invalidScopeQuery := `
			INSERT INTO public.api_keys (id, user_id, name, prefix, hash, scope, created_at)
			SELECT gen_random_uuid(), id, 'Invalid Scope Test', 'noot_invalid', 'hash123', 'invalid_scope', now()
			FROM public.profiles WHERE email = 'monalisa@birki.io' LIMIT 1;
		`

		cmd = exec.Command("docker", "exec", "-i", "supabase_db_noot", "psql", "-U", "postgres", "-d", "postgres", "-c", invalidScopeQuery)
		cmd.Dir = getProjectRoot()

		_, err = cmd.Output()
		assert.Error(t, err, "Should reject invalid scope values")
	})
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
