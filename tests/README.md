# Acceptance Tests

This directory contains acceptance tests for the Noot API, specifically focusing on authentication and API security after the backend code audit that removed unused authentication middleware functions.

## Overview

These tests verify that the authentication system works correctly after removing the unused middleware functions:

- `RequireAuth`
- `RequireSubscriptionTiers`
- `RequireProSubscription`

The tests ensure that the current `DualAuthMiddleware` (which supports both JWT and API key authentication) continues to work properly.

## Test Categories

### 1. Public Endpoints (`TestPublicEndpoints`)

- Verifies that public endpoints like `/api/v1/health` work without authentication
- Tests that OPTIONS requests bypass authentication (CORS support)

### 2. API Key Authentication (`TestAPIKeyAuthentication`)

- Tests valid API key authentication for both read and write operations
- Tests invalid and malformed API key rejection
- Uses the development API key that works during local testing

### 3. Missing Authentication (`TestMissingAuthentication`)

- Verifies that protected endpoints return 401 when no authentication is provided
- Tests various protected endpoints (consumptions, API keys, biometrics, etc.)

### 4. Consumption Endpoint (`TestConsumptionEndpointWithAPIKey`)

- Tests the main consumption creation endpoint with proper API key
- Tests various payload scenarios (valid text, empty text, missing payload)
- Verifies successful consumption creation and proper error handling

### 5. HTTP Methods (`TestDifferentHTTPMethods`)

- Tests authentication across different HTTP methods (GET, POST, PUT, DELETE)
- Verifies that authentication works consistently across all methods
- Tests both authenticated and unauthenticated requests

### 6. Authentication Bypass Prevention (`TestAuthenticationBypass`)

- Tests that authentication cannot be bypassed with incorrect headers
- Tests various malformed authentication attempts
- Verifies header name sensitivity and proper validation

### 7. End-to-End Flow (`TestEndToEndFlow`)

- Complete workflow test from health check to consumption creation
- Verifies the entire authentication and API flow works together

## Running the Tests

### Prerequisites

- The API server must be running at `http://localhost:3001`
- Start the server with: `script/server`

### Run Tests

```bash
script/acceptance
```

The script will:

1. Check if the server is running
2. Run all acceptance tests using vendored dependencies
3. Report results

## API Key for Testing

The tests use a hardcoded API key that works during local development:

```text
noot_3eb35a4c_36cd5f4d802c8ab41d3d7e4f6b7302a09c61067c
```

This key is safe to use in tests as it only works in development mode.

## Test Design

- **Timeout Protection**: All HTTP requests have a 30-second timeout
- **Comprehensive Coverage**: Tests cover positive and negative cases
- **Real API Calls**: Tests make actual HTTP requests to the running server
- **Detailed Assertions**: Tests verify both status codes and response content
- **Error Validation**: Tests check error response structure and messages

## Security Validation

These tests specifically validate that:

- Authentication is properly enforced on protected endpoints
- Public endpoints remain accessible without authentication
- API keys work correctly for authorized operations
- Invalid authentication attempts are properly rejected
- The authentication system cannot be bypassed

This provides confidence that the removal of unused authentication middleware did not introduce security vulnerabilities.
