package server

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockFile implements multipart.File for testing
type mockFile struct {
	*bytes.Reader
	content []byte
}

func (m *mockFile) Close() error {
	return nil
}

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[off:])
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

func newMockFile(content []byte) multipart.File {
	return &mockFile{
		Reader:  bytes.NewReader(content),
		content: content,
	}
}

func TestUploadSecurity(t *testing.T) {
	testCases := []struct {
		name          string
		filename      string
		contentType   string
		fileContent   []byte
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid webm audio file",
			filename:    "test.webm",
			contentType: "audio/webm",
			fileContent: append([]byte{0x1A, 0x45, 0xDF, 0xA3}, make([]byte, 100)...), // EBML header for webm
			expectError: false,
		},
		{
			name:          "Path traversal in filename",
			filename:      "../../../etc/passwd",
			contentType:   "audio/wav",
			fileContent:   append([]byte("RIFF"), append([]byte{0, 0, 0, 0}, append([]byte("WAVE"), make([]byte, 100)...)...)...), // Valid WAV header
			expectError:   true,
			errorContains: "path traversal",
		},
		{
			name:          "Windows path traversal",
			filename:      "..\\..\\windows\\system32\\cmd.exe",
			contentType:   "audio/wav",
			fileContent:   append([]byte("RIFF"), append([]byte{0, 0, 0, 0}, append([]byte("WAVE"), make([]byte, 100)...)...)...), // Valid WAV header
			expectError:   true,
			errorContains: "path traversal",
		},
		{
			name:          "Long filename",
			filename:      strings.Repeat("a", 300),
			contentType:   "audio/wav",
			fileContent:   append([]byte("RIFF"), append([]byte{0, 0, 0, 0}, append([]byte("WAVE"), make([]byte, 100)...)...)...), // Valid WAV header
			expectError:   true,
			errorContains: "filename too long",
		},
		{
			name:          "Unsupported content type",
			filename:      "test.exe",
			contentType:   "application/x-executable",
			fileContent:   []byte("MZ\x90\x00"),
			expectError:   true,
			errorContains: "unsupported content type",
		},
		{
			name:          "File too small",
			filename:      "tiny.wav",
			contentType:   "audio/wav",
			fileContent:   []byte("RIFF"),
			expectError:   true,
			errorContains: "file too small",
		},
		{
			name:          "Empty file",
			filename:      "empty.wav",
			contentType:   "audio/wav",
			fileContent:   []byte{},
			expectError:   true,
			errorContains: "file too small",
		},
		{
			name:          "Content type mismatch",
			filename:      "test.wav",
			contentType:   "audio/wav",
			fileContent:   []byte("<html><body>fake audio</body></html>"),
			expectError:   true,
			errorContains: "unsupported file type detected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock file
			file := newMockFile(tc.fileContent)

			// Create a mock header
			header := &multipart.FileHeader{
				Filename: tc.filename,
				Header:   make(map[string][]string),
				Size:     int64(len(tc.fileContent)),
			}
			if tc.contentType != "" {
				header.Header.Set("Content-Type", tc.contentType)
			}

			// Test the saveTempFile function
			tmpPath, mimeType, err := saveTempFile(file, header)

			if tc.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tc.name)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains,
						"Error should contain expected text")
				}
				assert.Empty(t, tmpPath, "Should not return temp path on error")
				assert.Empty(t, mimeType, "Should not return MIME type on error")
			} else {
				assert.NoError(t, err, "Should not error for valid file")
				assert.NotEmpty(t, tmpPath, "Should return temp path")
				assert.NotEmpty(t, mimeType, "Should return MIME type")

				// Clean up
				if tmpPath != "" {
					removeFile(tmpPath)
				}
			}
		})
	}
}

func TestContentTypeValidation(t *testing.T) {
	testCases := []struct {
		name          string
		contentType   string
		expectAllowed bool
	}{
		{"Valid webm", "audio/webm", true},
		{"Valid ogg", "audio/ogg", true},
		{"Valid mp3", "audio/mpeg", true},
		{"Valid wav", "audio/wav", true},
		{"Valid opus", "audio/opus", true},
		{"Video not allowed", "video/mp4", false},
		{"Text not allowed", "text/plain", false},
		{"Binary not allowed", "application/octet-stream", false},
		{"JavaScript not allowed", "application/javascript", false},
		{"HTML not allowed", "text/html", false},
		{"Executable not allowed", "application/x-executable", false},
		{"Empty string", "", false},
		{"Content type with params", "audio/webm; codecs=opus", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isAllowedAudioContentType(tc.contentType)
			assert.Equal(t, tc.expectAllowed, result,
				"Content type %s should be allowed: %v", tc.contentType, tc.expectAllowed)
		})
	}
}

func TestContentTypeCompatibility(t *testing.T) {
	testCases := []struct {
		name       string
		declared   string
		detected   string
		compatible bool
	}{
		{"Exact match", "audio/wav", "audio/wav", true},
		{"MP3 variants", "audio/mpeg", "audio/mp3", true},
		{"WAV variants", "audio/wav", "audio/wave", true},
		{"OGG variants", "audio/ogg", "application/ogg", true},
		{"Incompatible types", "audio/wav", "audio/mpeg", false},
		{"With parameters", "audio/webm; codecs=opus", "audio/webm", true},
		{"Case insensitive", "AUDIO/WAV", "audio/wav", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := areCompatibleContentTypes(tc.declared, tc.detected)
			assert.Equal(t, tc.compatible, result,
				"Types %s and %s should be compatible: %v", tc.declared, tc.detected, tc.compatible)
		})
	}
}
