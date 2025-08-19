package server

import (
	"bytes"
	"mime/multipart"
	"os"
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
