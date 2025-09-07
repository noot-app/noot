package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/grantbirki/noot/internal/storage"
	"github.com/microcosm-cc/bluemonday"
)

// OpenAIProvider implements the AIProvider interface using OpenAI's API
type OpenAIProvider struct {
	config     AIProviderConfig
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config AIProviderConfig) *OpenAIProvider {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	return &OpenAIProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// AIConfig represents the configuration for an AI prompt
type AIConfig struct {
	Model      string      `json:"model"`
	TextFormat string      `json:"text.format"`
	ToolChoice string      `json:"tool_choice"`
	Temp       float64     `json:"temp"`
	Tokens     int         `json:"tokens"`
	TopP       float64     `json:"top_p"`
	Tools      []AITool    `json:"tools"`
	Store      bool        `json:"store"`
	Include    []string    `json:"include"`
	Reasoning  AIReasoning `json:"reasoning"`
}

// AITool represents a tool configuration
type AITool struct {
	Type              string          `json:"type"`
	Filters           interface{}     `json:"filters,omitempty"`
	SearchContextSize string          `json:"search_context_size,omitempty"`
	UserLocation      *AIUserLocation `json:"user_location,omitempty"`
	ServerLabel       string          `json:"server_label,omitempty"`
	ServerURL         string          `json:"server_url,omitempty"`
	ServerDescription string          `json:"server_description,omitempty"`
	Authorization     string          `json:"authorization,omitempty"`
	AllowedTools      []string        `json:"allowed_tools,omitempty"`
	RequireApproval   string          `json:"require_approval,omitempty"`
}

// AIUserLocation represents user location for web search
type AIUserLocation struct {
	Type     string  `json:"type"`
	City     *string `json:"city"`
	Country  string  `json:"country"`
	Region   *string `json:"region"`
	Timezone *string `json:"timezone"`
}

// AIReasoning represents reasoning configuration
type AIReasoning struct{}

// AI configuration paths
const getNutritionAIDir = "ai/GetNutrition"
const parseItemsAIDir = "ai/ParseItems"

// Package-level AI request builder initialized lazily
var getNutritionBuilder *AIRequestBuilder
var parseItemsBuilder *AIRequestBuilder

// GetNutritionBuilder returns the nutrition V2 builder, initializing it if needed
func GetNutritionBuilder() (*AIRequestBuilder, error) {
	if getNutritionBuilder == nil {
		builder, err := NewAIRequestBuilderSafe(getNutritionAIDir)
		if err != nil {
			return nil, err
		}
		getNutritionBuilder = builder
	}
	return getNutritionBuilder, nil
}

// ParseItemsBuilder returns the parse items builder, initializing it if needed
func ParseItemsBuilder() (*AIRequestBuilder, error) {
	if parseItemsBuilder == nil {
		builder, err := NewAIRequestBuilderSafe(parseItemsAIDir)
		if err != nil {
			return nil, err
		}
		parseItemsBuilder = builder
	}
	return parseItemsBuilder, nil
}

// filterResponseForLogging removes verbose fields from OpenAI response for cleaner logging
func filterResponseForLogging(responseBody []byte) string {
	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		// If we can't parse it, return as-is (truncated if too long)
		if len(responseBody) > 1000 {
			return string(responseBody[:1000]) + "... [truncated]"
		}
		return string(responseBody)
	}

	// Remove the verbose instructions field
	delete(response, "instructions")

	// Re-marshal the filtered response
	filtered, err := json.Marshal(response)
	if err != nil {
		// Fallback to original if re-marshaling fails
		if len(responseBody) > 1000 {
			return string(responseBody[:1000]) + "... [truncated]"
		}
		return string(responseBody)
	}

	return string(filtered)
}

// Security constants for input validation
const (
	maxTranscriptLength = 10000 // Maximum allowed transcript length
	maxItemNameLength   = 200   // Maximum allowed item name length
	maxBrandLength      = 100   // Maximum allowed brand name length
)

// sanitizeTranscriptOutput validates and sanitizes AI transcript output
func sanitizeTranscriptOutput(text string) string {
	// Limit length to prevent potential issues
	if len(text) > maxTranscriptLength {
		LogWarn("Transcript length exceeded maximum, truncating", "length", len(text), "max", maxTranscriptLength)
		text = text[:maxTranscriptLength]
	}

	// Remove any control characters except common whitespace
	sanitized := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			return r
		}
		if unicode.IsControl(r) {
			return -1 // Remove control characters
		}
		return r
	}, text)

	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// validateParsedItems performs security validation on parsed items from AI
func validateParsedItems(items []Item) []Item {
	validItems := make([]Item, 0, len(items))

	for _, item := range items {
		// Validate and sanitize item name
		if len(item.Name) == 0 {
			LogWarn("Skipping item with empty name")
			continue
		}
		if len(item.Name) > maxItemNameLength {
			LogWarn("Item name too long, truncating", "original_length", len(item.Name), "max", maxItemNameLength)
			item.Name = item.Name[:maxItemNameLength]
		}

		// Sanitize item name
		item.Name = sanitizeText(item.Name)

		// Validate and sanitize brand if present
		if item.Brand != nil {
			if len(*item.Brand) > maxBrandLength {
				LogWarn("Brand name too long, truncating", "original_length", len(*item.Brand), "max", maxBrandLength)
				truncated := (*item.Brand)[:maxBrandLength]
				item.Brand = &truncated
			}
			sanitized := sanitizeText(*item.Brand)
			item.Brand = &sanitized
		}

		// Validate quantity values are reasonable
		if item.Grams < 0 || item.Grams > 10000 { // 10kg max per item
			LogWarn("Item grams value out of reasonable range", "name", item.Name, "grams", item.Grams)
			item.Grams = 0 // Reset to safe default
		}

		// This should maybe be revisited if the user is reporting consuming 1200mg of something
		if item.UserQuantity != nil && (*item.UserQuantity < 0 || *item.UserQuantity > 1000) {
			LogWarn("User quantity out of reasonable range", "name", item.Name, "quantity", *item.UserQuantity)
			item.UserQuantity = nil // Remove invalid value
		}

		validItems = append(validItems, item)
	}

	return validItems
}

// textSanitizer is a bluemonday policy for sanitizing text content
// It allows only plain text and removes all HTML tags and potentially dangerous content
var textSanitizer = bluemonday.StrictPolicy()

// sanitizeText removes potentially dangerous characters and content from text fields
// Uses bluemonday for robust, well-tested sanitization instead of manual regex patterns
func sanitizeText(text string) string {
	// Remove control characters first
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, text)

	// Use bluemonday to sanitize the text
	// StrictPolicy() removes all HTML tags and JavaScript, preventing XSS and other attacks
	sanitized := textSanitizer.Sanitize(text)

	// Log if sanitization changed the text (indicating potential attack)
	if sanitized != text {
		LogWarn("Potentially dangerous content detected and sanitized", "text_preview", truncateForLog(text))
	}

	return strings.TrimSpace(sanitized)
}

// truncateForLog safely truncates text for logging without exposing sensitive data
func truncateForLog(text string) string {
	const maxLogLength = 50
	if len(text) <= maxLogLength {
		return text
	}
	return text[:maxLogLength] + "..."
}

// makeOpenAIRequest creates and executes an OpenAI API request with common handling
func (p *OpenAIProvider) makeOpenAIRequest(ctx context.Context, builder *AIRequestBuilder, input string, endpoint string) ([]byte, error) {
	payloadStruct, err := builder.BuildRequestPayload(input)
	if err != nil {
		return nil, NewAppError("Failed to build request payload", http.StatusInternalServerError, err)
	}

	// Convert to map for HTTP request
	payload := map[string]any{
		"model":             payloadStruct.Model,
		"input":             payloadStruct.Input,
		"text":              payloadStruct.Text,
		"reasoning":         payloadStruct.Reasoning,
		"tools":             payloadStruct.Tools,
		"temperature":       payloadStruct.Temperature,
		"max_output_tokens": payloadStruct.MaxOutputTokens,
		"top_p":             payloadStruct.TopP,
		"store":             payloadStruct.Store,
		"include":           payloadStruct.Include,
	}

	b, _ := json.Marshal(payload)
	url := p.config.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, NewAppError("Failed to create request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, NewAppError("OpenAI request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI API error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body)))
		return nil, NewAppError("OpenAI API failed", http.StatusInternalServerError,
			fmt.Errorf("OpenAI API error: %d %s", resp.StatusCode, string(body)))
	}

	// Read the full response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAppError("Failed to read response body", http.StatusInternalServerError, err)
	}

	return responseBody, nil
}

// parseOpenAIResponse extracts content from OpenAI's /responses endpoint structure
func parseOpenAIResponse(responseBody []byte) (string, error) {
	// Parse the /responses endpoint structure
	var response struct {
		Output []struct {
			Type    string `json:"type"`
			Status  string `json:"status"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", NewAppError("Failed to parse response", http.StatusInternalServerError, err)
	}

	// Extract the text content from the message output
	var content string
	var statusIssues []string

	for _, output := range response.Output {
		if output.Type == "message" {
			if output.Status != "completed" {
				statusIssues = append(statusIssues, fmt.Sprintf("message status: %s", output.Status))
				LogWarn("OpenAI message output not completed", "type", output.Type, "status", output.Status)
				continue
			}

			for _, contentItem := range output.Content {
				if contentItem.Type == "output_text" {
					content = contentItem.Text
					break
				}
			}
			if content != "" {
				break
			}
		}
	}

	// Check if we found any non-completed statuses
	if len(statusIssues) > 0 && content == "" {
		errorMsg := fmt.Sprintf("OpenAI request failed with status issues: %s", strings.Join(statusIssues, ", "))
		LogError("OpenAI request failed", errors.New(errorMsg))
		return "", NewAppError("OpenAI request failed", http.StatusInternalServerError, errors.New(errorMsg))
	}

	if content == "" {
		LogWarn("No content found in OpenAI response")
		return "", NewAppError("No content found in OpenAI response", http.StatusInternalServerError, fmt.Errorf("empty content"))
	}

	return strings.TrimSpace(content), nil
}

func (p *OpenAIProvider) transcriptionPrompt() string {
	return `The audio is a short dictation of foods and drinks consumed. Preserve exact brand and product names (e.g., "Clover Organic", "Trader Joe's", "Siggi's", "Icelandic skyr", "LaCroix"), coffee drink terms (espresso, latte, macchiato), tea terms (matcha), and ingredient names (goji berries, blueberries, Greek yogurt, European style yogurt). Keep numbers and units (cups, grams, ounces, tbsp, tsp, slices, pieces) and include standard punctuation.

Do not add or infer items that were not spoken. If an item is given without a quantity, assume one standard serving size that would make logical sense in the context of the consumption:

- Fruits: 1 medium (banana ~118g, apple ~182g, orange ~154g)
- Eggs: 2 large eggs (~100g)
- Yogurt: 1 container (~170g)
- Coffee: 1 cup (~240g)
- Bread: 1-2 slices (~28-56g)
- Meat/fish: 1 serving portion (~100g)
- Beverages: 1 can/bottle (~355g for canned, ~500g for bottled)

For coffee drinks, assume standard sizes: latte contains 2 shots espresso, cappuccino 1-2 shots. Preserve preparation methods when mentioned (grilled, baked, raw, steamed).`
}

// TranscribeAudio implements AIProvider.TranscribeAudio
func (p *OpenAIProvider) TranscribeAudio(ctx context.Context, filePath, mimeType string) (string, error) {
	LogDebug("Starting OpenAI transcription", "model", p.config.TranscribeModel, "file", filePath)

	prompt := p.transcriptionPrompt()
	lang := strings.TrimSpace(os.Getenv("TRANSCRIBE_LANGUAGE")) // e.g., "en"
	respFormat := strings.TrimSpace(os.Getenv("OPENAI_TRANSCRIBE_RESPONSE_FORMAT"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("model", p.config.TranscribeModel)
	if prompt != "" {
		_ = writer.WriteField("prompt", prompt)
	}
	if lang != "" {
		_ = writer.WriteField("language", lang)
	}
	if respFormat != "" {
		_ = writer.WriteField("response_format", respFormat)
	}

	fw, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", NewAppError("Failed to create form file", http.StatusInternalServerError, err)
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", NewAppError("Failed to open audio file", http.StatusInternalServerError, err)
	}
	defer f.Close()
	if _, err = io.Copy(fw, f); err != nil {
		return "", NewAppError("Failed to copy audio file", http.StatusInternalServerError, err)
	}
	if err := writer.Close(); err != nil {
		return "", NewAppError("Failed to close form writer", http.StatusInternalServerError, err)
	}

	url := p.config.BaseURL + "/audio/transcriptions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", NewAppError("Failed to create transcription request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	LogDebug("Sending transcription request to OpenAI")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", NewAppError("OpenAI transcription request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		LogError("OpenAI transcription error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(b)))
		return "", NewAppError("OpenAI transcription failed", http.StatusInternalServerError,
			fmt.Errorf("OpenAI API error: %d %s", resp.StatusCode, string(b)))
	}

	if strings.EqualFold(respFormat, "text") {
		b, _ := io.ReadAll(resp.Body)
		result := sanitizeTranscriptOutput(string(b))
		LogDebug("Transcription completed", "length", len(result))
		return result, nil
	}

	var out struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", NewAppError("Failed to parse transcription response", http.StatusInternalServerError, err)
	}

	// Security: Validate and sanitize transcript output
	sanitized := sanitizeTranscriptOutput(out.Text)
	LogDebug("Transcription completed", "length", len(sanitized))
	return sanitized, nil
}

// ParseItems implements AIProvider.ParseItems
func (p *OpenAIProvider) ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	// Security: Validate input transcript length and content
	if len(transcriptText) > maxTranscriptLength {
		return ParsedItems{}, NewAppError("Transcript too long for processing", http.StatusBadRequest,
			fmt.Errorf("transcript length %d exceeds maximum %d", len(transcriptText), maxTranscriptLength))
	}

	if len(strings.TrimSpace(transcriptText)) == 0 {
		return ParsedItems{}, NewAppError("Empty transcript provided", http.StatusBadRequest,
			fmt.Errorf("transcript cannot be empty"))
	}

	// Log with truncated transcript for security
	LogDebug("Starting OpenAI ParseItems request", "transcript_preview", truncateForLog(transcriptText))

	// Build input message as JSON
	inputObj := map[string]string{
		"transcript_text": transcriptText,
	}
	inputBytes, _ := json.Marshal(inputObj)
	input := string(inputBytes)

	// Security: Log truncated input to avoid exposing full transcript in logs
	LogDebug("Sending OpenAI ParseItems request", "input_preview", truncateForLog(input))

	// Build request payload using pre-initialized AIRequestBuilder
	builder, err := ParseItemsBuilder()
	if err != nil {
		return ParsedItems{}, NewAppError("Failed to initialize AI request builder", http.StatusInternalServerError, err)
	}

	LogDebug("Sending item parsing request to OpenAI")
	responseBody, err := p.makeOpenAIRequest(ctx, builder, input, "/responses")
	if err != nil {
		return ParsedItems{}, err
	}

	LogDebug("OpenAI parsing response received", "full_response", filterResponseForLogging(responseBody))

	content, err := parseOpenAIResponse(responseBody)
	if err != nil {
		return ParsedItems{}, err
	}

	LogDebug("Extracted content from OpenAI response", "content_length", len(content), "content_preview", func() string {
		if len(content) > 100 {
			return content[:100] + "..."
		}
		return content
	}())

	// Temporary struct for parsing OpenAI response with new schema structure
	var parsed struct {
		Success bool    `json:"success"`
		Message *string `json:"message"`
		Items   []struct {
			Name         string   `json:"name"`
			Grams        *float64 `json:"grams"`
			UserQuantity *float64 `json:"user_quantity"`
			UserUnit     *string  `json:"user_unit"`
			Brand        *string  `json:"brand"`
			Context      *string  `json:"context"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		LogWarn("Failed to parse OpenAI JSON response", "content", content, "error", err.Error(), "full_response", filterResponseForLogging(responseBody))
		return ParsedItems{}, NewAppError("Failed to parse OpenAI response", http.StatusInternalServerError, err)
	}

	// Check if parsing was successful according to the schema
	if !parsed.Success {
		message := "Unknown parsing failure"
		if parsed.Message != nil {
			message = *parsed.Message
		}
		LogWarn("OpenAI parsing reported failure", "message", message)
		return ParsedItems{}, NewAppError("Failed to parse items: "+message, http.StatusBadRequest, fmt.Errorf("parsing failed: %s", message))
	}

	// Normalize items (no nutrition data at this stage)
	clean := make([]Item, 0, len(parsed.Items))
	for _, i := range parsed.Items {
		name := strings.TrimSpace(i.Name)
		if name == "" {
			continue
		}
		// Handle grams being potentially null in the new schema
		if i.Grams == nil || *i.Grams <= 0 {
			LogWarn("Invalid or missing grams value for item", "name", name, "grams", i.Grams)
			continue
		}
		// Detect multi-unit quantities and set base quantity
		var baseQuantity *float64
		if i.UserQuantity != nil && *i.UserQuantity > 1.0 {
			// If user specified more than 1 unit, normalize to single unit
			baseQuantity = i.UserQuantity
		}

		clean = append(clean, Item{
			Name:         name,
			Grams:        *i.Grams,
			UserQuantity: i.UserQuantity,
			UserUnit:     strPtrOrNil(i.UserUnit),
			Brand:        strPtrOrNil(i.Brand),
			BaseQuantity: baseQuantity,
			Context:      strPtrOrNil(i.Context),
			Nutrients:    nil, // No nutrition data in phase 1
		})
	}

	// Security: Validate and sanitize parsed items before returning
	clean = validateParsedItems(clean)

	LogDebug("Item parsing completed", "items_found", len(clean))
	return ParsedItems{Items: clean}, nil
}

// GetNutrition implements AIProvider.GetNutrition
func (p *OpenAIProvider) GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error) {
	return p.GetNutritionWithContext(ctx, item)
}

// GetNutritionWithContext implements AIProvider.GetNutritionWithContext
func (p *OpenAIProvider) GetNutritionWithContext(ctx context.Context, item Item) (CompleteNutrient, error) {
	// Security: Validate item input
	if len(strings.TrimSpace(item.Name)) == 0 {
		return CompleteNutrient{}, NewAppError("Item name cannot be empty", http.StatusBadRequest,
			fmt.Errorf("item name is required"))
	}

	if len(item.Name) > maxItemNameLength {
		return CompleteNutrient{}, NewAppError("Item name too long", http.StatusBadRequest,
			fmt.Errorf("item name length %d exceeds maximum %d", len(item.Name), maxItemNameLength))
	}

	LogDebug("Starting OpenAI GetNutrition request", "item_name", truncateForLog(item.Name))

	// Build input message as JSON
	inputObj := map[string]any{
		"name":    item.Name,
		"grams":   item.Grams,
		"brand":   item.Brand,
		"context": item.Context,
	}
	inputBytes, _ := json.Marshal(inputObj)
	input := string(inputBytes)

	LogDebug("Starting OpenAI GetNutrition request", "item", item.Name, "input_message", input)

	// Build request payload using pre-initialized AIRequestBuilder
	builder, err := GetNutritionBuilder()
	if err != nil {
		return CompleteNutrient{}, NewAppError("Failed to initialize AI request builder", http.StatusInternalServerError, err)
	}

	responseBody, err := p.makeOpenAIRequest(ctx, builder, input, "/responses")
	if err != nil {
		return CompleteNutrient{}, err
	}

	LogDebug("OpenAI nutrition response received", "full_response", filterResponseForLogging(responseBody))

	content, err := parseOpenAIResponse(responseBody)
	if err != nil {
		return CompleteNutrient{}, err
	}

	LogDebug("Extracted content from OpenAI response", "content_length", len(content), "content_preview", func() string {
		if len(content) > 100 {
			return content[:100] + "..."
		}
		return content
	}())

	// Parse the new schema structure with success and message fields
	var result struct {
		Success     bool                    `json:"success"`
		Message     *string                 `json:"message"`
		Nutrients   CompleteNutrient        `json:"nutrients"`
		Ingredients []storage.OFFIngredient `json:"ingredients,omitempty"`
		URL         *string                 `json:"url,omitempty"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		LogWarn("Failed to parse nutrition JSON", "content", content, "error", err.Error(), "full_response", filterResponseForLogging(responseBody))
		return CompleteNutrient{}, NewAppError("Failed to parse nutrition data", http.StatusInternalServerError, err)
	}

	// Check if nutrition fetching was successful according to the schema
	if !result.Success {
		message := "Unknown nutrition fetching failure"
		if result.Message != nil {
			message = *result.Message
		}
		LogWarn("OpenAI nutrition fetching reported failure", "message", message)
		return CompleteNutrient{}, NewAppError("Failed to fetch nutrition: "+message, http.StatusBadRequest, fmt.Errorf("nutrition fetching failed: %s", message))
	}

	LogDebug("CompleteNutrient resolved", "item", item.Name, "nutrients", result.Nutrients)

	return result.Nutrients, nil
}

// GetNutritionWithContextComplete implements AIProvider.GetNutritionWithContextComplete
func (p *OpenAIProvider) GetNutritionWithContextComplete(ctx context.Context, item Item) (NutritionResponse, error) {
	// Security: Validate item input
	if len(strings.TrimSpace(item.Name)) == 0 {
		return NutritionResponse{}, NewAppError("Item name cannot be empty", http.StatusBadRequest,
			fmt.Errorf("item name is required"))
	}

	if len(item.Name) > maxItemNameLength {
		return NutritionResponse{}, NewAppError("Item name too long", http.StatusBadRequest,
			fmt.Errorf("item name length %d exceeds maximum %d", len(item.Name), maxItemNameLength))
	}

	LogDebug("Starting OpenAI GetNutritionComplete request", "item_name", truncateForLog(item.Name))

	// Build input message as JSON
	inputObj := map[string]any{
		"name":    item.Name,
		"grams":   item.Grams,
		"brand":   item.Brand,
		"context": item.Context,
	}
	inputBytes, _ := json.Marshal(inputObj)
	input := string(inputBytes)

	LogDebug("Starting OpenAI GetNutritionComplete request", "item", item.Name, "input_message", input)

	// Build request payload using pre-initialized AIRequestBuilder
	builder, err := GetNutritionBuilder()
	if err != nil {
		return NutritionResponse{}, NewAppError("Failed to initialize AI request builder", http.StatusInternalServerError, err)
	}

	responseBody, err := p.makeOpenAIRequest(ctx, builder, input, "/responses")
	if err != nil {
		return NutritionResponse{}, err
	}

	LogDebug("OpenAI nutrition response received", "full_response", filterResponseForLogging(responseBody))

	content, err := parseOpenAIResponse(responseBody)
	if err != nil {
		return NutritionResponse{}, err
	}

	LogDebug("Extracted content from OpenAI response", "content_length", len(content), "content_preview", func() string {
		if len(content) > 100 {
			return content[:100] + "..."
		}
		return content
	}())

	// Parse the complete response structure including ingredients and URL
	var result struct {
		Success     bool                    `json:"success"`
		Message     *string                 `json:"message"`
		Nutrients   CompleteNutrient        `json:"nutrients"`
		Ingredients []storage.OFFIngredient `json:"ingredients,omitempty"`
		URL         *string                 `json:"url,omitempty"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		LogWarn("Failed to parse nutrition JSON", "content", content, "error", err.Error(), "full_response", filterResponseForLogging(responseBody))
		return NutritionResponse{}, NewAppError("Failed to parse nutrition data", http.StatusInternalServerError, err)
	}

	// Check if nutrition fetching was successful according to the schema
	if !result.Success {
		message := "Unknown nutrition fetching failure"
		if result.Message != nil {
			message = *result.Message
		}
		LogWarn("OpenAI nutrition fetching reported failure", "message", message)
		return NutritionResponse{}, NewAppError("Failed to fetch nutrition: "+message, http.StatusBadRequest, fmt.Errorf("nutrition fetching failed: %s", message))
	}

	// Log ingredients if found
	if len(result.Ingredients) > 0 {
		LogDebug("Extracted ingredients from AI", "item", item.Name, "ingredient_count", len(result.Ingredients))
	}

	// Log URL if found
	if result.URL != nil && *result.URL != "" {
		LogDebug("Extracted URL from AI", "item", item.Name, "url", *result.URL)
	}

	LogDebug("Complete nutrition resolved", "item", item.Name, "nutrients", result.Nutrients, "has_ingredients", len(result.Ingredients) > 0, "has_url", result.URL != nil)

	return NutritionResponse{
		Nutrients:   result.Nutrients,
		Ingredients: result.Ingredients,
		URL:         result.URL,
	}, nil
}
