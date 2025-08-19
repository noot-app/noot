# noot 🍎

An AI-powered nutrition logging web app. Just say what you ate!

- Press and hold the mic button, speak your consumption.
- Backend (Go) transcribes with OpenAI, parses items, fetches nutrients from an API based LLM.
- Returns a complete nutrient profile summary with calories, protein, carbs, fiber, vitamins, minerals, etc.

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
   - Say what you consumed (e.g., "I had a latte with organic whole milk and Greek yogurt with blueberries")
   - Release and see your nutrition summary

## API Endpoints

- `GET /` — Static frontend (embedded)
- `POST /api/consumption` — Multipart form with `audio` field (webm/opus/mp3/wav)
- `GET /api/consumptions` — View stored consumptions (development mode only)

## Development

- **Test:** `script/test`
- **Lint:** `script/lint`  
- **Build:** `script/build` (or `script/build --single-target` for faster iteration)
- **Dev Server:** `script/server` (optionally pass in `--clean` to reset and re-seed the local database)

### Database Management

- **Reset:** `script/db reset` — Drop all tables and start fresh
- **Seed:** `script/db seed` — Add development data (monalisa user + sample consumptions)
- **Dump:** `script/db dump` — View database contents in human-readable format

## Technical Details

- SQLite database with automatic migrations for consumption storage
- OpenAI gpt-4o-mini-transcribe for speech-to-text
- OpenAI gpt-4o-mini for consumption parsing with complete nutrition data
- Audio uploads are streamed to temporary files to avoid memory spikes
- Development seeding with realistic consumption data for testing
