# AI Configuration Embedding

The noot application uses AI configuration files located in the `ai/` directory for OpenAI API requests. To ensure these configurations are available in production deployments (especially in minimal Docker containers), they are embedded directly into the binary at build time.

## How it works

1. **Source configurations**: AI configs are maintained in `ai/ParseItems/` and `ai/GetNutrition/` directories
2. **Build-time embedding**: During build, configs are copied to `internal/server/ai/` and embedded using Go's `embed` package
3. **Runtime fallback**: The application first tries to load configs from the filesystem, then falls back to embedded configs

## Files embedded

For each AI directory (`ParseItems` and `GetNutrition`):
- `config.json` - OpenAI API configuration
- `prompt.md` - System prompt
- `schema.json` - JSON schema for responses

## Development vs Production

- **Development**: Edit files in `ai/` directory normally - changes are picked up immediately
- **Production**: Embedded configs are used automatically when filesystem files are not available

## Build process

The build process automatically:
1. Runs `script/prepare-ai-embed` to copy AI configs for embedding
2. Embeds the configs into the binary using Go's `embed` package
3. Creates a self-contained binary that doesn't require external config files

## File locations

- **Source**: `ai/ParseItems/`, `ai/GetNutrition/`
- **Embedded copy**: `internal/server/ai/` (excluded from git via `.gitignore`)
- **Access**: `loadAIConfig()`, `loadPrompt()`, `loadSchema()` with automatic fallback