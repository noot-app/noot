# GitHub Copilot Guidelines

This is a Go based repository that is a minimal web app to log meals by voice. The repository is structured as a mono repo that is a GoLang REST API backend and a SvelteKit with TypeScript frontend that is build and served out of the `apps/web` directory. 

## Code Standards

### Development Flow

- Bootstrap: `script/bootstrap`
- Generate types: `script/generate-types` - generates types for the frontend and the backend
- Test: `script/test` - frontend and backend tests
- Lint: `script/lint` - only lints the backend
- Build: `script/build` - only builds the backend
- Frontend tests: `cd apps/web && npm run test_run && cd -`
- Testing database schema: `script/db reset && script/db test`
- Checking svelte kit front end: `cd apps/web && npm run check` 

> Note: `script/build --single-target` can be used when iterating on changes rapidly as it will only build the current target which is faster than building all targets.

## Repository Structure

- `cmd/*`: Main entry points for the REST server
- `internal/`: Logic related to the core functionality of the REST API server
- `script/`: Scripts for building, testing, and releasing the REST API server
- `apps/web/`: The SvelteKit with TypeScript frontend
- `api/`: The openapi spec for the REST API backend
- `.github/`: GitHub Actions workflows for CI/CD
- `vendor/`: Vendor directory for Go modules (committed to the repository for reproducibility)
- `ai/`: Prompts, configs, and return response schemas for OpenAI

## Key Guidelines

1. Follow Go best practices and idiomatic patterns
2. Maintain existing code structure and organization
3. Use dependency injection patterns where appropriate
4. Write unit tests for new functionality. Use table-driven unit tests when possible. Use `testify` for assertions.
5. When responding to code refactoring suggestions, function suggestions, or other code changes, please keep your responses as concise as possible. We are capable engineers and can understand the code changes without excessive explanation. If you feel that a more detailed explanation is necessary, you can provide it, but keep it concise.
6. When suggesting code changes, always opt for the most maintainable approach. Try your best to keep the code clean and follow DRY principles. Avoid unnecessary complexity and always consider the long-term maintainability of the code.
7. When writing unit tests, always strive for 100% code coverage where it makes sense. Try to consider edge cases as well.
8. Always strive to keep the codebase clean and maintainable. Avoid unnecessary complexity and always consider the long-term maintainability of the code.
9. Always strive for the highest level of code coverage with unit tests where possible.
10. Keep your responses and summary messages short, and very concise. Don't over explain unless asked to in your "what I did" summaries.

## Helpful Tips

- Since you are an AI, you cannot run, view, and interact with the API/Webapp UI very well. For this reason, simply ensuring that tests pass and builds succeed is sufficient for you to verify that the code changes are correct.
- If any changes to the database sql/migration files are required, please edit exitising migrations directly rather than creating new ones. The reason that I ask for this is that nothing is deployed yet to production so there is no need for backwards compatibility or legacy support.
- If all the services are up and running (when interacting with me during local development and the right api keys are set) the following command can be run to create a new consumtpion: `$ curl -X POST -H "X-API-Key: noot_3eb35a4c_36cd5f4d802c8ab41d3d7e4f6b7302a09c61067c" -H "Content-Type: application/json" -d '{"text": "I ate one carrot"}' "http://localhost:3001/api/v1/consumption"`
- When giving summary responses, keep them very concise.
