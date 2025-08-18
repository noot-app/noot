# noot 🍎

An AI-powered nutrition logging web app. Just say what you ate!

- Press and hold the mic button, speak your meal.
- Backend (Go) transcribes with OpenAI, parses items, fetches nutrients from USDA FDC.
- Returns a simple macro summary with calories, protein, carbs, fiber, etc.

## Quick Start

1) **Get API Keys:**
   - OpenAI API key from [platform.openai.com](https://platform.openai.com/api-keys)
   - USDA FDC API key from [fdc.nal.usda.gov/api-key-signup.html](https://fdc.nal.usda.gov/api-key-signup.html)

2) **Configure:**

   ```bash
   cp .env.example .env
   # Edit .env and add your API keys
   ```

3) **Run:**

   ```bash
   # Build and run
   script/build --single-target
   ./bin/noot
   
   # Or run directly
   go run ./cmd/noot
   ```

4) **Use:**
   - Visit <http://localhost:3000>
   - Press and hold the mic button
   - Say your meal (e.g., "I had a latte with organic whole milk and Greek yogurt with blueberries")
   - Release and see your nutrition summary

## API Endpoints

- `GET /` — Static frontend (embedded)
- `POST /api/ingest` — Multipart form with `audio` field (webm/opus/mp3/wav)

## Example Response

```json
{
  "transcript": "I had a latte with organic whole milk and Greek yogurt with blueberries",
  "parsedItems": [
    {"name": "latte", "quantity": 1, "unit": "cup", "brand": null},
    {"name": "organic whole milk", "quantity": null, "unit": null, "brand": null},
    {"name": "Greek yogurt", "quantity": null, "unit": null, "brand": null},
    {"name": "blueberries", "quantity": null, "unit": null, "brand": null}
  ],
  "items": [
    {
      "item": {"name": "latte", "quantity": 1, "unit": "cup", "brand": null},
      "fdc": {"fdcId": 123456, "description": "Coffee, brewed", "brandOwner": null, "dataType": "SR Legacy"},
      "nutrients": {"energy_kcal": 5, "protein_g": 0.3, "fat_g": 0.1, "carbs_g": 0.9, "fiber_g": 0, "sugar_g": 0}
    }
  ],
  "summary": {
    "totals": {"energy_kcal": 320, "protein_g": 18, "fat_g": 12, "carbs_g": 35, "fiber_g": 4, "sugar_g": 28},
    "percent_of_daily": {"energy.kcal": 16, "protein.g": 36, "fat.g": 15, "carbs.g": 13, "fiber.g": 14, "sugar.g": 56},
    "daily_values": {"energy_kcal": 2000, "protein_g": 50, "fat_g": 78, "carbs_g": 275, "fiber_g": 28, "sugar_g": 50}
  }
}
```

## Development

- **Test:** `script/test`
- **Lint:** `script/lint`  
- **Build:** `script/build` (or `script/build --single-target` for faster iteration)

## Technical Details

- Single binary web server with embedded frontend assets
- OpenAI gpt-4o-mini-transcribe for speech-to-text
- OpenAI gpt-4o-mini for meal parsing
- USDA FoodData Central API for nutrition data
- Go stdlib only (no external dependencies)
- Hermetic CI/CD with vendored modules
