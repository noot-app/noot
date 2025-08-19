# noot 🍎

An AI-powered nutrition logging web app. Just say what you ate!

- Press and hold the mic button, speak your meal.
- Backend (Go) transcribes with OpenAI, parses items, fetches nutrients from USDA FDC.
- Returns a simple macro summary with calories, protein, carbs, fiber, etc.

## Quick Start

1) **Get API Keys:**

   - OpenAI API key from [platform.openai.com](https://platform.openai.com/api-keys)

2) **Configure:**

    ```ini
    # create .env file

    # Required
    OPENAI_API_KEY=<key>

    # Server Configuration
    PORT=3000
    ENV=development

    # Logging & Debugging (Enable detailed error reporting)
    LOG_LEVEL=DEBUG
    DEBUG=true

    # OpenAI Configuration
    OPENAI_TRANSCRIBE_MODEL=gpt-4o-mini-transcribe
    OPENAI_PARSE_MODEL=gpt-4o-mini

    TRANSCRIBE_LANGUAGE=en
    #OPENAI_TRANSCRIBE_RESPONSE_FORMAT=

    # Upload Configuration
    MAX_UPLOAD_BYTES=104857600  # Maximum upload size in bytes (default: 100MB)

    # Database Configuration
    DATABASE_PATH=./noot.db     # Path to SQLite database file (default: ./noot.db)
    DEV_DB_SEED=true           # Enable database seeding in development (optional)
    ```

3) **Run:** The local server can be run with `script/server` (use `--production` for production mode).

4) **Use:**
   - Visit <http://localhost:3000>
   - Press and hold the mic button
   - Say your meal (e.g., "I had a latte with organic whole milk and Greek yogurt with blueberries")
   - Release and see your nutrition summary

## API Endpoints

- `GET /` — Static frontend (embedded)
- `POST /api/ingest` — Multipart form with `audio` field (webm/opus/mp3/wav)
- `GET /api/meals` — View stored meals (development mode only)

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
- **Dev Server:** `script/server` (starts with development-friendly caching)

### Database Management

- **Reset:** `script/db reset` — Drop all tables and start fresh
- **Seed:** `script/db seed` — Add development data (monalisa user + sample meals)
- **Dump:** `script/db dump` — View database contents in human-readable format

## Technical Details

- Single binary web server with embedded frontend assets
- SQLite database with automatic migrations for meal storage
- OpenAI gpt-4o-mini-transcribe for speech-to-text
- OpenAI gpt-4o-mini for meal parsing with complete nutrition data
- Audio uploads are streamed to temporary files to avoid memory spikes
- Development seeding with realistic meal data for testing
