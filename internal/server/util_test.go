package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMultipartFile implements multipart.File for testing
type mockMultipartFile struct {
	*bytes.Reader
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func TestSaveTempFile_StreamsAndDetectsMIME(t *testing.T) {
	// Create a test file with a WebM header followed by some payload
	// WebM signature: 0x1A 0x45 0xDF 0xA3
	webmHeader := []byte{0x1A, 0x45, 0xDF, 0xA3}
	payload := bytes.Repeat([]byte("test data "), 100) // Some additional payload

	// Combine header with payload
	testData := append(webmHeader, payload...)

	// Create mock file
	mockFile := &mockMultipartFile{
		Reader: bytes.NewReader(testData),
	}

	header := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(testData)),
	}

	// Call saveTempFile
	tmpPath, mimeType, err := saveTempFile(mockFile, header)

	// Verify no error
	require.NoError(t, err)

	// Clean up the temp file when done
	defer os.Remove(tmpPath)

	// Verify file exists
	assert.FileExists(t, tmpPath)

	// Verify file size matches what was written
	fileInfo, err := os.Stat(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, int64(len(testData)), fileInfo.Size())

	// Verify MIME type detection
	// WebM files are detected as "video/webm" by http.DetectContentType when they have proper headers
	assert.Equal(t, "video/webm", mimeType)

	// Verify file content
	fileContent, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, testData, fileContent)
}

func TestSaveTempFile_EnforcesMaxBytes(t *testing.T) {
	// Set a small upload limit for testing
	originalEnv := os.Getenv("MAX_UPLOAD_BYTES")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("MAX_UPLOAD_BYTES")
		} else {
			os.Setenv("MAX_UPLOAD_BYTES", originalEnv)
		}
	}()

	os.Setenv("MAX_UPLOAD_BYTES", "1024") // 1KB limit

	// Create test data larger than the limit
	largeData := bytes.Repeat([]byte("a"), 2048) // 2KB of data

	mockFile := &mockMultipartFile{
		Reader: bytes.NewReader(largeData),
	}

	header := &multipart.FileHeader{
		Filename: "large.webm",
		Size:     int64(len(largeData)),
	}

	// Call saveTempFile - should fail
	tmpPath, mimeType, err := saveTempFile(mockFile, header)

	// Verify error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum upload size")
	assert.Empty(t, tmpPath)
	assert.Empty(t, mimeType)

	// Verify no temp file was left behind
	if tmpPath != "" {
		assert.NoFileExists(t, tmpPath)
	}
}

func TestSaveTempFile_HandlesShortFiles(t *testing.T) {
	// Create a very short file (less than 512 bytes)
	shortData := []byte("short file content")

	mockFile := &mockMultipartFile{
		Reader: bytes.NewReader(shortData),
	}

	header := &multipart.FileHeader{
		Filename: "short.txt",
		Size:     int64(len(shortData)),
	}

	// Call saveTempFile
	tmpPath, mimeType, err := saveTempFile(mockFile, header)

	// Verify no error
	require.NoError(t, err)

	// Clean up
	defer os.Remove(tmpPath)

	// Verify file exists and has correct size
	fileInfo, err := os.Stat(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, int64(len(shortData)), fileInfo.Size())

	// Verify MIME type was detected (text files should be detected as text/plain)
	assert.Equal(t, "text/plain; charset=utf-8", mimeType)

	// Verify file content
	fileContent, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, shortData, fileContent)
}

func TestSaveTempFile_DefaultMaxBytes(t *testing.T) {
	// Ensure no MAX_UPLOAD_BYTES is set
	originalEnv := os.Getenv("MAX_UPLOAD_BYTES")
	os.Unsetenv("MAX_UPLOAD_BYTES")
	defer func() {
		if originalEnv != "" {
			os.Setenv("MAX_UPLOAD_BYTES", originalEnv)
		}
	}()

	// Create test data that's reasonable (1MB)
	testData := bytes.Repeat([]byte("test"), 256*1024) // 1MB

	mockFile := &mockMultipartFile{
		Reader: bytes.NewReader(testData),
	}

	header := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(testData)),
	}

	// Call saveTempFile - should succeed with default 100MB limit
	tmpPath, mimeType, err := saveTempFile(mockFile, header)

	// Verify no error
	require.NoError(t, err)

	// Clean up
	defer os.Remove(tmpPath)

	// Verify file exists
	assert.FileExists(t, tmpPath)
	assert.NotEmpty(t, mimeType)
}

func TestSaveTempFile_CustomTempDir(t *testing.T) {
	// Create a custom temp directory for testing
	customTempDir := t.TempDir()
	
	// Set TEMP_DIR environment variable
	originalEnv := os.Getenv("TEMP_DIR")
	os.Setenv("TEMP_DIR", customTempDir)
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("TEMP_DIR")
		} else {
			os.Setenv("TEMP_DIR", originalEnv)
		}
	}()

	// Create test data
	testData := []byte("test audio data")

	mockFile := &mockMultipartFile{
		Reader: bytes.NewReader(testData),
	}

	header := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(testData)),
	}

	// Call saveTempFile
	tmpPath, mimeType, err := saveTempFile(mockFile, header)

	// Verify no error
	require.NoError(t, err)

	// Clean up
	defer os.Remove(tmpPath)

	// Verify file was created in custom temp directory
	assert.Contains(t, tmpPath, customTempDir)
	assert.FileExists(t, tmpPath)
	assert.NotEmpty(t, mimeType)

	// Verify file content
	fileContent, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, testData, fileContent)
}

func TestSaveTempFile_TempDirCreation(t *testing.T) {
	// Create a non-existent directory path
	baseTempDir := t.TempDir()
	customTempDir := filepath.Join(baseTempDir, "non-existent", "temp")
	
	// Set TEMP_DIR environment variable to non-existent path
	originalEnv := os.Getenv("TEMP_DIR")
	os.Setenv("TEMP_DIR", customTempDir)
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("TEMP_DIR")
		} else {
			os.Setenv("TEMP_DIR", originalEnv)
		}
	}()

	// Verify the directory doesn't exist initially
	_, err := os.Stat(customTempDir)
	assert.True(t, os.IsNotExist(err))

	// Create test data
	testData := []byte("test audio data")

	mockFile := &mockMultipartFile{
		Reader: bytes.NewReader(testData),
	}

	header := &multipart.FileHeader{
		Filename: "test.webm",
		Size:     int64(len(testData)),
	}

	// Call saveTempFile - should create the directory and succeed
	tmpPath, mimeType, err := saveTempFile(mockFile, header)

	// Verify no error
	require.NoError(t, err)

	// Clean up
	defer os.Remove(tmpPath)

	// Verify directory was created and file exists
	assert.DirExists(t, customTempDir)
	assert.Contains(t, tmpPath, customTempDir)
	assert.FileExists(t, tmpPath)
	assert.NotEmpty(t, mimeType)
}

func TestGuessExtension(t *testing.T) {
	tests := []struct {
		mime     string
		expected string
	}{
		{"audio/webm", ".webm"},
		{"audio/ogg", ".ogg"},
		{"audio/mp3", ".mp3"},
		{"audio/mpeg", ".mp3"},
		{"audio/wav", ".wav"},
		{"video/webm", ".webm"},
		{"application/octet-stream", ".webm"}, // default
		{"text/plain", ".webm"},               // default
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			result := guessExtension(tt.mime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetenv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		def      string
		envValue string
		expected string
	}{
		{
			name:     "env var exists",
			key:      "TEST_VAR",
			def:      "default",
			envValue: "actual_value",
			expected: "actual_value",
		},
		{
			name:     "env var empty",
			key:      "TEST_VAR_EMPTY",
			def:      "default",
			envValue: "",
			expected: "default",
		},
		{
			name:     "env var not set",
			key:      "TEST_VAR_NOT_SET",
			def:      "default",
			envValue: "",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing env var
			originalValue := os.Getenv(tt.key)
			defer func() {
				if originalValue == "" {
					os.Unsetenv(tt.key)
				} else {
					os.Setenv(tt.key, originalValue)
				}
			}()

			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getenv(tt.key, tt.def)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHttpError(t *testing.T) {
	// Initialize logger for test
	InitLogger()

	w := httptest.NewRecorder()

	httpError(w, http.StatusBadRequest, "test error")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test error", response["error"])
	assert.Equal(t, float64(400), response["code"])
	assert.NotNil(t, response["timestamp"])
}

func TestHttpErrorWithDetails(t *testing.T) {
	// Initialize logger for test
	InitLogger()

	// Set debug mode to ensure stack trace is included
	originalDebug := os.Getenv("DEBUG")
	os.Setenv("DEBUG", "true")
	defer func() {
		if originalDebug == "" {
			os.Unsetenv("DEBUG")
		} else {
			os.Setenv("DEBUG", originalDebug)
		}
	}()

	w := httptest.NewRecorder()
	stack := []string{"line1", "line2"}
	traceID := "trace123"

	httpErrorWithDetails(w, http.StatusInternalServerError, "detailed error", stack, traceID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "detailed error", response["error"])
	assert.Equal(t, float64(500), response["code"])
	assert.Equal(t, "trace123", response["trace_id"])
	assert.NotNil(t, response["timestamp"])

	// Check if stack is included
	if stack, exists := response["stack"]; exists && stack != nil {
		stackSlice := stack.([]interface{})
		assert.Len(t, stackSlice, 2)
		assert.Equal(t, "line1", stackSlice[0])
		assert.Equal(t, "line2", stackSlice[1])
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	writeJSON(w, http.StatusCreated, data)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test", response["message"])
}

func TestWriteJSON_InvalidData(t *testing.T) {
	w := httptest.NewRecorder()

	// Create data that will fail JSON marshaling (channel can't be marshaled)
	data := map[string]interface{}{
		"invalid": make(chan int),
	}

	writeJSON(w, http.StatusOK, data)

	// Should still write the status code
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestRemoveFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	// Verify file exists
	_, err = os.Stat(tmpPath)
	assert.NoError(t, err)

	// Remove file
	removeFile(tmpPath)

	// Verify file is gone
	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveFile_NonExistentFile(t *testing.T) {
	// Should not panic when removing non-existent file
	assert.NotPanics(t, func() {
		removeFile("/non/existent/file")
	})
}

func TestLoadDotEnv(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	envContent := `# This is a comment
TEST_KEY1=value1
TEST_KEY2="quoted value"
TEST_KEY3='single quoted'
	TEST_KEY4   =   spaced value   
EMPTY_KEY=

# Another comment
OVERRIDE_KEY=override_value`

	err := os.WriteFile(envFile, []byte(envContent), 0644)
	require.NoError(t, err)

	// Save current directory and change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Clear test env vars
	testKeys := []string{"TEST_KEY1", "TEST_KEY2", "TEST_KEY3", "TEST_KEY4", "EMPTY_KEY", "OVERRIDE_KEY"}
	for _, key := range testKeys {
		os.Unsetenv(key)
	}

	// Set one key that should not be overridden
	os.Setenv("OVERRIDE_KEY", "original_value")

	// Load .env file
	LoadDotEnv()

	// Test loaded values
	assert.Equal(t, "value1", os.Getenv("TEST_KEY1"))
	assert.Equal(t, "quoted value", os.Getenv("TEST_KEY2"))
	assert.Equal(t, "single quoted", os.Getenv("TEST_KEY3"))
	assert.Equal(t, "spaced value", os.Getenv("TEST_KEY4"))
	assert.Equal(t, "", os.Getenv("EMPTY_KEY"))
	assert.Equal(t, "original_value", os.Getenv("OVERRIDE_KEY")) // Should not be overridden

	// Clean up
	for _, key := range testKeys {
		os.Unsetenv(key)
	}
}

func TestLoadDotEnv_NoFile(t *testing.T) {
	// Should not panic when no .env file exists
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		LoadDotEnv()
	})
}

func TestSummarize(t *testing.T) {
	items := []ItemWithNutrition{
		{
			Item: Item{
				Name: "Apple",
				Nutrients: &CompleteNutrient{
					Calories:        100,
					Protein:         1,
					TotalFat:        0.5,
					SaturatedFat:    0.1,
					TransFat:        0,
					Cholesterol:     0,
					Sodium:          2,
					TotalCarbs:      25,
					DietaryFiber:    4,
					TotalSugars:     19,
					AddedSugars:     0,
					VitaminA:        5,
					VitaminC:        8.4,
					VitaminD:        0,
					VitaminE:        0.3,
					VitaminK:        2.2,
					Thiamine:        0.02,
					Riboflavin:      0.03,
					Niacin:          0.1,
					VitaminB6:       0.04,
					Folate:          3,
					VitaminB12:      0,
					Calcium:         6,
					Iron:            0.12,
					Magnesium:       5,
					Phosphorus:      11,
					Potassium:       107,
					Zinc:            0.04,
					Copper:          0.03,
					Manganese:       0.04,
					Selenium:        0,
					Biotin:          0,
					PantothenicAcid: 0.1,
					Choline:         0,
					Iodine:          0,
					Molybdenum:      0,
					Chromium:        0,
					Fluoride:        0,
					Chloride:        0,
					// New nutrients
					Omega3Ala:          0.01,
					Omega3Epa:          0,
					Omega3Dha:          0,
					Omega6:             0.05,
					Creatine:           0,
					Caffeine:           0,
					Alcohol:            0,
					PolyunsaturatedFat: 0.1,
					MonounsaturatedFat: 0.02,
				},
			},
		},
		{
			Item: Item{
				Name: "Banana",
				Nutrients: &CompleteNutrient{
					Calories:        105,
					Protein:         1.3,
					TotalFat:        0.4,
					SaturatedFat:    0.1,
					TransFat:        0,
					Cholesterol:     0,
					Sodium:          1,
					TotalCarbs:      27,
					DietaryFiber:    3.1,
					TotalSugars:     14.4,
					AddedSugars:     0,
					VitaminA:        3.1,
					VitaminC:        8.7,
					VitaminD:        0,
					VitaminE:        0.1,
					VitaminK:        0.5,
					Thiamine:        0.03,
					Riboflavin:      0.07,
					Niacin:          0.67,
					VitaminB6:       0.37,
					Folate:          20,
					VitaminB12:      0,
					Calcium:         5,
					Iron:            0.26,
					Magnesium:       27,
					Phosphorus:      22,
					Potassium:       358,
					Zinc:            0.15,
					Copper:          0.08,
					Manganese:       0.27,
					Selenium:        1,
					Biotin:          0,
					PantothenicAcid: 0.33,
					Choline:         9.8,
					Iodine:          0,
					Molybdenum:      0,
					Chromium:        0,
					Fluoride:        0,
					Chloride:        0,
					// New nutrients
					Omega3Ala:          0.027,
					Omega3Epa:          0,
					Omega3Dha:          0,
					Omega6:             0.046,
					Creatine:           0,
					Caffeine:           0,
					Alcohol:            0,
					PolyunsaturatedFat: 0.073,
					MonounsaturatedFat: 0.032,
				},
			},
		},
	}

	summary := summarize(items)

	// Test totals
	assert.Equal(t, 205.0, summary.Totals.Calories)
	assert.Equal(t, 2.3, summary.Totals.Protein)
	assert.Equal(t, 0.9, summary.Totals.TotalFat)
	assert.Equal(t, 52.0, summary.Totals.TotalCarbs)
	assert.Equal(t, 465.0, summary.Totals.Potassium)

	// Test daily value percentages (using standard daily values)
	assert.NotNil(t, summary.PercentOfDaily)
	assert.NotNil(t, summary.DailyValuesUsed)

	// Check that daily values are set (using actual values from function)
	assert.Equal(t, 2000.0, summary.DailyValuesUsed["calories"])
	assert.Equal(t, 50.0, summary.DailyValuesUsed["protein_g"])
	assert.Equal(t, 78.0, summary.DailyValuesUsed["total_fat_g"])

	// Check percentage calculations (using actual daily values)
	caloriePercent := 205.0 / 2000.0 * 100.0
	expectedCaloriePercent := int(caloriePercent)
	assert.Equal(t, expectedCaloriePercent, summary.PercentOfDaily["calories"])
}

func TestSummarize_EmptyItems(t *testing.T) {
	items := []ItemWithNutrition{}

	summary := summarize(items)

	assert.Equal(t, 0.0, summary.Totals.Calories)
	assert.Equal(t, 0.0, summary.Totals.Protein)
	assert.NotNil(t, summary.PercentOfDaily)
	assert.NotNil(t, summary.DailyValuesUsed)
}

func TestSummarize_NilNutrients(t *testing.T) {
	items := []ItemWithNutrition{
		{
			Item: Item{
				Name:      "Test Item",
				Nutrients: nil, // No nutrition data
			},
		},
	}

	summary := summarize(items)

	assert.Equal(t, 0.0, summary.Totals.Calories)
	assert.Equal(t, 0.0, summary.Totals.Protein)
}
