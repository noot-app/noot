# noot

A minimal web app to log meals by voice:

- Press and hold the mic button, speak your meal.
- Backend (Go) transcribes with OpenAI, parses items, fetches nutrients from USDA FDC.
- Returns a simple macro summary.

## Run

1) Set environment variables (or copy `.env.example` to `.env`):
   - OPENAI_API_KEY
   - USDA_FDC_API_KEY
   - Optional: PORT, OPENAI_TRANSCRIBE_MODEL, OPENAI_PARSE_MODEL, TRANSCRIBE_PROMPT, TRANSCRIBE_LANGUAGE

2) Start the server:
   - go run ./cmd/go-template
   - Visit http://localhost:3000

## Endpoints

- GET / — embedded static frontend
- POST /api/ingest — multipart form with field `audio` (webm/opus/mp3/wav)

## Notes

- No CLI/Cobra. Single binary web server with embedded assets (Go `embed`).
- OpenAI endpoints: /v1/audio/transcriptions (gpt-4o-mini-transcribe), /v1/chat/completions (gpt-4o-mini).
- FDC endpoints: /v1/foods/search, /v1/food/{fdcId}.
