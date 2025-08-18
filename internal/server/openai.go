package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	openAIBaseURL       = "https://api.openai.com/v1"
	openAITranscribeURL = openAIBaseURL + "/audio/transcriptions"
	openAIChatURL       = openAIBaseURL + "/chat/completions"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

func transcriptionPrompt() string {
	return `The audio is a short dictation of foods and drinks consumed. Preserve exact brand and product names (e.g., "Clover Organic", "Trader Joe's", "Siggi's", "Icelandic skyr", "LaCroix"), coffee drink terms (espresso, latte, macchiato), tea terms (matcha), and ingredient names (goji berries, blueberries, Greek yogurt, European style yogurt). Keep numbers and units (cups, grams, ounces, tbsp) and include standard punctuation. Do not add or infer items that were not spoken. If an item is given without a quantity, assume it is one standard serving size of that item which would make logical sense in the context of the meal. For example, if a user says "I had a banana", assume it is one banana, not a bunch. If they say "I had some eggs", assume it is two eggs, not a dozen. If they say "I had some yogurt", assume it is one standard serving size of yogurt, not a gallon. If they say "I had some coffee", assume it is one standard cup of coffee, not a pot. If the user sayd "I had a lattle" assume it contains two shots of espresso.`
}

func parseSystemPrompt() string {
	return `You extract foods and drinks from a freeform meal, snack, or drink description. Return strict JSON with { \"items\": [ { \"name\": string, \"quantity\": number | null, \"unit\": string | null, \"brand\": string | null } ] }.`
}

func transcribeAudio(ctx context.Context, filePath, _ string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", NewAppError("OPENAI_API_KEY not configured", http.StatusInternalServerError, nil)
	}
	model := getenv("OPENAI_TRANSCRIBE_MODEL", "gpt-4o-mini-transcribe")

	LogDebug("Starting OpenAI transcription", "model", model, "file", filePath)

	prompt := transcriptionPrompt()

	lang := strings.TrimSpace(os.Getenv("TRANSCRIBE_LANGUAGE")) // e.g., "en"
	respFormat := strings.TrimSpace(os.Getenv("OPENAI_TRANSCRIBE_RESPONSE_FORMAT"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("model", model)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAITranscribeURL, body)
	if err != nil {
		return "", NewAppError("Failed to create transcription request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	LogDebug("Sending transcription request to OpenAI")
	resp, err := httpClient.Do(req)
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
		result := string(b)
		LogDebug("Transcription completed", "length", len(result))
		return result, nil
	}
	var out struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", NewAppError("Failed to parse transcription response", http.StatusInternalServerError, err)
	}

	LogDebug("Transcription completed", "length", len(out.Text))
	return out.Text, nil
}

func parseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return ParsedItems{}, NewAppError("OPENAI_API_KEY not configured", http.StatusInternalServerError, nil)
	}
	model := getenv("OPENAI_PARSE_MODEL", "gpt-4o-mini")

	LogDebug("Starting OpenAI item parsing", "model", model, "transcript_length", len(transcriptText))

	system := parseSystemPrompt()
	user := "Meal: " + transcriptText

	payload := map[string]any{
		"model":       model,
		"temperature": 0.0, // maybe set to 0.1 or 0.2 if items are being dropped from the json: 0.0 is most deterministic
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIChatURL, bytes.NewReader(b))
	if err != nil {
		return ParsedItems{}, NewAppError("Failed to create parsing request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	LogDebug("Sending parsing request to OpenAI")
	resp, err := httpClient.Do(req)
	if err != nil {
		return ParsedItems{}, NewAppError("OpenAI parsing request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI parsing error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body)))
		return ParsedItems{}, NewAppError("OpenAI parsing failed", http.StatusInternalServerError,
			fmt.Errorf("OpenAI API error: %d %s", resp.StatusCode, string(body)))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ParsedItems{}, NewAppError("Failed to parse OpenAI response", http.StatusInternalServerError, err)
	}
	content := ""
	if len(out.Choices) > 0 {
		content = out.Choices[0].Message.Content
	}

	LogDebug("OpenAI parsing response received", "content_length", len(content))

	var parsed ParsedItems
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		LogWarn("Failed to parse OpenAI JSON response", "content", content, "error", err.Error())
		parsed = ParsedItems{Items: []Item{}}
	}
	// Normalize
	clean := make([]Item, 0, len(parsed.Items))
	for _, i := range parsed.Items {
		name := strings.TrimSpace(i.Name)
		if name == "" {
			continue
		}
		clean = append(clean, Item{
			Name:     name,
			Quantity: i.Quantity,
			Unit:     strPtrOrNil(i.Unit),
			Brand:    strPtrOrNil(i.Brand),
		})
	}

	LogDebug("Item parsing completed", "items_found", len(clean))
	return ParsedItems{Items: clean}, nil
}

func strPtrOrNil(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
