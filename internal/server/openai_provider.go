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

func (p *OpenAIProvider) transcriptionPrompt() string {
	return `The audio is a short dictation of foods and drinks consumed. Preserve exact brand and product names (e.g., "Clover Organic", "Trader Joe's", "Siggi's", "Icelandic skyr", "LaCroix"), coffee drink terms (espresso, latte, macchiato), tea terms (matcha), and ingredient names (goji berries, blueberries, Greek yogurt, European style yogurt). Keep numbers and units (cups, grams, ounces, tbsp, tsp, slices, pieces) and include standard punctuation.

Do not add or infer items that were not spoken. If an item is given without a quantity, assume one standard serving size that would make logical sense in the context of the consumption:

- Fruits: 1 medium (banana, apple, orange)
- Eggs: 2 large eggs (not a dozen)
- Yogurt: 1 cup or 1 container (not a gallon)
- Coffee: 1 cup (not a pot)
- Bread: 1-2 slices (not a loaf)
- Meat/fish: 1 serving portion (3-4 oz)
- Beverages: 1 glass/cup (8-12 oz)

For coffee drinks, assume standard sizes: latte contains 2 shots espresso, cappuccino 1-2 shots. Preserve preparation methods when mentioned (grilled, baked, raw, steamed).`
}

func (p *OpenAIProvider) parseItemsSystemPrompt() string {
	return `You extract individual food and drink items from a freeform meal, snack, beverage, or consumption description. Convert all quantities to grams for internal processing while preserving the original user input for display. Return strict JSON with the following structure:

{
  "items": [
    {
      "name": string,
      "grams": number,
      "user_quantity": number | null,
      "user_unit": string | null,
      "brand": string | null
    }
  ]
}

IMPORTANT INSTRUCTIONS:
1. EXTRACT INDIVIDUAL ITEMS: Separate each distinct food or drink item mentioned.
2. CONVERT TO GRAMS: Always provide the "grams" field with the equivalent weight in grams. Use standard food weights:
   - Butter: 1 stick = 113g, 1 tbsp = 14.2g, 1 cup = 227g
   - Yogurt: 1 cup = 245g, 1 container (typical) = 170g
   - Banana: 1 medium = 118g, 1 large = 136g
   - Eggs: 1 large egg = 50g, 2 large eggs = 100g
   - Coffee/liquids: 1 cup = 240ml = 240g, 1 tbsp = 15ml = 15g
   - Apple: 1 medium = 182g, 1 large = 223g
   - Bread: 1 slice = 28g, 1 thick slice = 35g
   - Chicken breast: 3.5oz = 100g, 1 breast (typical) = 140g
   - Rice: 1 cup cooked = 158g, 1 cup uncooked = 185g

3. PRESERVE USER INPUT: Store the original quantity and unit in "user_quantity" and "user_unit" for display purposes.
4. INFER SERVING SIZES: If quantity is not specified, assume reasonable standard serving sizes and convert to grams.
5. PRESERVE BRANDS: Keep exact brand and product names (e.g., "Clover Sonoma", "Trader Joe's", "Siggi's", "KFC", "Starbucks").
6. DO NOT ADD NUTRITION DATA: Only extract item identification and weight conversion, not nutrition information.`
}

func (p *OpenAIProvider) nutritionSystemPrompt() string {
	return `You provide complete nutrition information for a single food item based on its weight in grams. Return strict JSON with the following structure:

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
    "biotin_mcg": number,
    "pantothenic_acid_mg": number,
    "choline_mg": number,
    "calcium_mg": number,
    "iron_mg": number,
    "magnesium_mg": number,
    "phosphorus_mg": number,
    "potassium_mg": number,
    "zinc_mg": number,
    "copper_mg": number,
    "manganese_mg": number,
    "selenium_mcg": number,
    "iodine_mcg": number,
    "molybdenum_mcg": number,
    "chromium_mcg": number,
    "fluoride_mg": number,
    "chloride_mg": number,
    "dha_mg": number
  }
}

IMPORTANT INSTRUCTIONS:
1. WEIGHT-BASED NUTRITION: Provide accurate nutrition data for the exact gram weight specified. Use your knowledge of food composition databases and USDA data.
2. BRANDED VS GENERIC: Prioritize branded nutrition data when brand is specified, otherwise use generic USDA-style data for the food type.
3. ZERO VALUES: Use 0 for nutrients that are truly absent, but provide realistic non-zero values for nutrients that are typically present.
4. ACCURACY REFERENCE: Use these as accuracy checkpoints for common foods per 100g:
   - Butter (salted): ~717 calories, ~81 g fat, ~0.9 g protein
   - Whole eggs (raw, whole): ~143 calories, ~9.5 g fat, ~12.6 g protein
   - Banana (raw): ~89 calories, ~0.3 g fat, ~1.1 g protein
   - White bread (commercial): ~265 calories, ~3.2 g fat, ~9 g protein
   - Whole milk (3.25 percent fat): ~61 calories, ~3.3 g fat, ~3.2 g protein
   - Rice, white (cooked): ~130 calories, ~0.3 g fat, ~2.7 g protein
   - Skim milk (~0.5 percent fat or less): ~34 calories, ~0.1 g fat, ~3.4 g protein
   - Chicken breast (cooked, skinless): ~165 calories, ~3.6 g fat, ~31 g protein
   - Potato (raw, white): ~77 calories, ~0.1 g fat, ~2 g protein
   - Onion (raw): ~40 calories, ~0.1 g fat, ~1.1 g protein
   - Blueberries (raw): ~57 calories, ~0.3 g fat, ~0.7 g protein
   - Apple (raw): ~52 calories, ~0.2 g fat, ~0.3 g protein
   - Beef, ground (85% lean, cooked): ~250 calories, ~17 g fat, ~26 g protein
   - Tomato (raw): ~18 calories, ~0.2 g fat, ~0.9 g protein
   - Cheddar cheese (aged): ~403 calories, ~33.1 g fat, ~24.9 g protein
   - Carrot (raw): ~41 calories, ~0.2 g fat, ~0.9 g protein

DHA CONTENT REFERENCE (per 100g):
   - Salmon, farmed: ~1800mg DHA
   - Salmon, wild: ~1400mg DHA  
   - Sardines, canned: ~1400mg DHA
   - Mackerel: ~1600mg DHA
   - Tuna, bluefin: ~1200mg DHA
   - Fish oil supplements: Very high DHA (check product specs)
   - Most plant foods: 0mg DHA (may contain ALA omega-3 instead)

5. CALCULATE FROM WEIGHT: Scale nutrition values proportionally based on the gram weight provided.`
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

	LogDebug("OpenAI item parsing response received", "content_length", len(content), "content", content)

	// Temporary struct for parsing OpenAI response
	var parsed struct {
		Items []struct {
			Name         string   `json:"name"`
			Grams        float64  `json:"grams"`
			UserQuantity *float64 `json:"user_quantity"`
			UserUnit     *string  `json:"user_unit"`
			Brand        *string  `json:"brand"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		LogWarn("Failed to parse OpenAI JSON response", "content", content, "error", err.Error())
		parsed = struct {
			Items []struct {
				Name         string   `json:"name"`
				Grams        float64  `json:"grams"`
				UserQuantity *float64 `json:"user_quantity"`
				UserUnit     *string  `json:"user_unit"`
				Brand        *string  `json:"brand"`
			} `json:"items"`
		}{Items: []struct {
			Name         string   `json:"name"`
			Grams        float64  `json:"grams"`
			UserQuantity *float64 `json:"user_quantity"`
			UserUnit     *string  `json:"user_unit"`
			Brand        *string  `json:"brand"`
		}{}}
	}

	// Normalize items (no nutrition data at this stage)
	clean := make([]Item, 0, len(parsed.Items))
	for _, i := range parsed.Items {
		name := strings.TrimSpace(i.Name)
		if name == "" {
			continue
		}
		// Ensure grams is positive
		if i.Grams <= 0 {
			LogWarn("Invalid grams value for item", "name", name, "grams", i.Grams)
			continue
		}
		clean = append(clean, Item{
			Name:         name,
			Grams:        i.Grams,
			UserQuantity: i.UserQuantity,
			UserUnit:     strPtrOrNil(i.UserUnit),
			Brand:        strPtrOrNil(i.Brand),
			Nutrients:    nil, // No nutrition data in phase 1
		})
	}

	LogDebug("Item parsing completed", "items_found", len(clean), "items", clean)
	return ParsedItems{Items: clean}, nil
}

// GetNutrition implements AIProvider.GetNutrition
func (p *OpenAIProvider) GetNutrition(ctx context.Context, item Item) (CompleteNutrient, error) {
	system := p.nutritionSystemPrompt()

	LogDebug("Starting OpenAI GetNutrition request", "item", item.Name)

	// Build user message with item details in grams
	var userMsg strings.Builder
	userMsg.WriteString("Item: ")
	userMsg.WriteString(item.Name)
	userMsg.WriteString(fmt.Sprintf("\nWeight: %.1fg", item.Grams))

	if item.Brand != nil {
		userMsg.WriteString(fmt.Sprintf("\nBrand: %s", *item.Brand))
	}

	LogDebug("Starting OpenAI GetNutrition request", "item", item.Name, "user_message", userMsg.String())

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

	LogDebug("CompleteNutrient resolved", "item", item.Name, "nutrients", result.Nutrients)

	return result.Nutrients, nil
}
