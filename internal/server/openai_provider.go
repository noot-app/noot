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

// ParseItems implements AIProvider.ParseItems
func (p *OpenAIProvider) ParseItems(ctx context.Context, transcriptText string) (ParsedItems, error) {
	LogDebug("Starting OpenAI item parsing", "model", p.config.ParseModel, "transcript_length", len(transcriptText))

	system := p.parseItemsSystemPrompt()
	user := "Meal: " + transcriptText

	payload := map[string]any{
		"model":       p.config.ParseModel,
		"temperature": 0.0, // deterministic for item parsing
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	b, _ := json.Marshal(payload)
	url := p.config.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return ParsedItems{}, NewAppError("Failed to create parsing request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	LogDebug("Sending item parsing request to OpenAI")
	resp, err := p.httpClient.Do(req)
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

	LogDebug("OpenAI item parsing response received", "content_length", len(content))
	LogDebug("OpenAI response content", "content", content)

	var parsed ParsedItems
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		LogWarn("Failed to parse OpenAI JSON response", "content", content, "error", err.Error())
		parsed = ParsedItems{Items: []Item{}}
	}

	// Normalize items (no nutrition data at this stage)
	clean := make([]Item, 0, len(parsed.Items))
	for _, i := range parsed.Items {
		name := strings.TrimSpace(i.Name)
		if name == "" {
			continue
		}
		clean = append(clean, Item{
			Name:      name,
			Quantity:  i.Quantity,
			Unit:      strPtrOrNil(i.Unit),
			Brand:     strPtrOrNil(i.Brand),
			Nutrients: nil, // No nutrition data in phase 1
		})
	}

	LogDebug("Item parsing completed", "items_found", len(clean))
	return ParsedItems{Items: clean}, nil
}

// GetNutrition implements AIProvider.GetNutrition
func (p *OpenAIProvider) GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error) {
	system := p.nutritionSystemPrompt()

	// Build user message with item details
	var userMsg strings.Builder
	userMsg.WriteString("Item: ")
	userMsg.WriteString(item.Name)

	if item.Quantity != nil {
		userMsg.WriteString(fmt.Sprintf("\nQuantity: %v", *item.Quantity))
	}
	if item.Unit != nil {
		userMsg.WriteString(fmt.Sprintf("\nUnit: %s", *item.Unit))
	}
	if item.Brand != nil {
		userMsg.WriteString(fmt.Sprintf("\nBrand: %s", *item.Brand))
	}

	payload := map[string]any{
		"model":       p.config.ParseModel,
		"temperature": 0.1, // slightly higher for nutrition estimates
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": userMsg.String()},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	b, _ := json.Marshal(payload)
	url := p.config.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return CompleteNutrient{}, NewAppError("Failed to create nutrition request", http.StatusInternalServerError, err)
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return CompleteNutrient{}, NewAppError("OpenAI nutrition request failed", http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI nutrition error", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(body)))
		return CompleteNutrient{}, NewAppError("OpenAI nutrition failed", http.StatusInternalServerError,
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
		return CompleteNutrient{}, NewAppError("Failed to parse nutrition response", http.StatusInternalServerError, err)
	}

	content := ""
	if len(out.Choices) > 0 {
		content = out.Choices[0].Message.Content
	}

	var result struct {
		Nutrients CompleteNutrient `json:"nutrients"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		LogWarn("Failed to parse nutrition JSON", "content", content, "error", err.Error())
		return CompleteNutrient{}, NewAppError("Failed to parse nutrition data", http.StatusInternalServerError, err)
	}

	return result.Nutrients, nil
}

// Private helper methods for prompts

func (p *OpenAIProvider) transcriptionPrompt() string {
	return `The audio is a short dictation of foods and drinks consumed. Preserve exact brand and product names (e.g., "Clover Organic", "Trader Joe's", "Siggi's", "Icelandic skyr", "LaCroix"), coffee drink terms (espresso, latte, macchiato), tea terms (matcha), and ingredient names (goji berries, blueberries, Greek yogurt, European style yogurt). Keep numbers and units (cups, grams, ounces, tbsp) and include standard punctuation. Do not add or infer items that were not spoken. If an item is given without a quantity, assume it is one standard serving size of that item which would make logical sense in the context of the consumption. For example, if a user says "I had a banana", assume it is one banana, not a bunch. If they say "I had some eggs", assume it is two eggs, not a dozen. If they say "I had some yogurt", assume it is one standard serving size of yogurt, not a gallon. If they say "I had some coffee", assume it is one standard cup of coffee, not a pot. If the user says "I had a latte" assume it contains two shots of espresso.`
}

func (p *OpenAIProvider) parseItemsSystemPrompt() string {
	return `You extract individual food and drink items from a freeform meal, snack, beverage, or consumption description. Return strict JSON with the following structure:

{
  "items": [
    {
      "name": string,
      "quantity": number | null,
      "unit": string | null,
      "brand": string | null
    }
  ]
}

IMPORTANT INSTRUCTIONS:
1. EXTRACT INDIVIDUAL ITEMS: Separate each distinct food or drink item mentioned.
2. INFER SERVING SIZES: If quantity is not specified, assume reasonable standard serving sizes based on context:
   - Yogurt: 1 cup (245g)
   - Banana: 1 medium (118g)  
   - Eggs: 2 large eggs (100g)
   - Coffee: 1 cup (240ml)
   - Latte: 10oz with 2 shots espresso using standard 20g shots
   - Apple: 1 medium (182g)
   - Bread slice: 1 slice (28g)
   - Chicken breast: 3.5oz (100g)
   - Rice: 1 cup cooked (158g)

3. PRESERVE BRANDS: Keep exact brand and product names (e.g., "Clover Sonoma", "Trader Joe's", "Siggi's", "KFC", "Starbucks").
4. NORMALIZE UNITS: Use standard units (g, mg, ml, cups, tbsp, etc.).
5. DO NOT ADD NUTRITION DATA: Only extract item identification, not nutrition information.`
}

func (p *OpenAIProvider) nutritionSystemPrompt() string {
	return `You provide complete nutrition information for a single food item. Return strict JSON with the following structure:

{
  "nutrients": {
    "calories": number,
    "protein_g": number,
    "total_fat_g": number,
    "saturated_fat_g": number,
    "trans_fat_g": number,
    "cholesterol_mg": number,
    "sodium_mg": number,
    "total_carbs_g": number,
    "dietary_fiber_g": number,
    "total_sugars_g": number,
    "added_sugars_g": number,
    "vitamin_a_mcg": number,
    "vitamin_c_mg": number,
    "vitamin_d_mcg": number,
    "vitamin_e_mg": number,
    "vitamin_k_mcg": number,
    "thiamine_mg": number,
    "riboflavin_mg": number,
    "niacin_mg": number,
    "vitamin_b6_mg": number,
    "folate_mcg": number,
    "vitamin_b12_mcg": number,
    "calcium_mg": number,
    "iron_mg": number,
    "magnesium_mg": number,
    "phosphorus_mg": number,
    "potassium_mg": number,
    "zinc_mg": number,
    "copper_mg": number,
    "manganese_mg": number,
    "selenium_mcg": number
  }
}

IMPORTANT INSTRUCTIONS:
1. NUTRITION DATA ACCURACY: Provide accurate nutrition data per serving for the specified quantity and unit. Use your knowledge of food composition databases, USDA data, and nutrition labels.
2. HANDLE COMPLEX ITEMS: For prepared foods, estimate based on typical recipes and ingredients. For restaurant items, use available nutrition information or estimate based on similar items.
3. ZERO VALUES: Use 0 for nutrients that are truly absent (like vitamin B12 in plants), but provide realistic non-zero values for nutrients that are typically present even in small amounts.
4. BRANDED VS GENERIC: Prioritize branded nutrition data when brand is specified, otherwise use generic USDA-style data for the food type.
5. QUANTITY-ADJUSTED: Provide nutrition values for the exact quantity/unit specified, not per 100g.`
}
