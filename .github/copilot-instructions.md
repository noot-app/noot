# GitHub Copilot Guidelines

This is a Go based repository that is a minimal web app to log meals by voice. The repository is structured to support a single binary web server with embedded assets using Go's `embed` package.

## Code Standards

### Development Flow

- Bootstrap: `script/bootstrap`
- Generate types: `script/generate-types`
- Test: `script/test`
- Lint: `script/lint`
- Build: `script/build`

> Note: `script/build --single-target` can be used when iterating on changes rapidly as it will only build the current target which is faster than building all targets.

## Repository Structure

- `cmd/*`: Main cli entry points and executables
- `internal/`: Logic related to the core functionality of the CLI
- `script/`: Scripts for building, testing, and releasing the CLI
- `.github/`: GitHub Actions workflows for CI/CD
- `vendor/`: Vendor directory for Go modules (committed to the repository for reproducibility)

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

## Helpful Tips

- Since you are an AI, you cannot run, view, and interact with the API/Webapp UI very well. For this reason, simply ensuring that tests pass and builds succeed is sufficient for you to verify that the code changes are correct.
