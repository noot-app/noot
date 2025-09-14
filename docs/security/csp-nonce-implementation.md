# CSP Nonce Implementation

This document describes the Content Security Policy (CSP) nonce implementation for the SvelteKit frontend, which eliminates the need for `'unsafe-inline'` directives and provides strong protection against XSS attacks.

## Overview

The nonce-based CSP system generates unique cryptographic nonces for each HTTP request and automatically injects them into all inline scripts and styles. This allows the browser to execute only authorized inline content while blocking any malicious injected scripts or styles.

## Implementation Details

### Core Components

1. **Nonce Generation** (`hooks.server.ts`)
   - Generates a unique 16-byte base64-encoded nonce per request
   - Uses Node.js `crypto.randomBytes()` for cryptographic security
   - Stored in `event.locals.cspNonce` for access throughout the request lifecycle

2. **HTML Transformation** (`hooks.server.ts`)
   - Automatically injects nonces into `<script>` and `<style>` tags during SSR
   - Handles SvelteKit's `%sveltekit.csp.nonce%` placeholder replacement
   - Uses intelligent regex patterns to avoid duplicating existing nonces
   - Only active in production mode to maintain development convenience

3. **CSP Header Generation** (`hooks.server.ts`)
   - Sets strict CSP headers with nonce-based directives
   - Removes `'unsafe-inline'` from both `script-src` and `style-src`
   - Includes necessary domains for Supabase and application functionality

4. **Client-Side Utilities** (`lib/utils/csp.ts`)
   - Provides safe functions for creating dynamic scripts and styles
   - Includes utilities for CSP debugging in development
   - Validates nonce format and handles edge cases

### Security Benefits

**Before (Vulnerable):**
```
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.nootapp.io
style-src 'self' 'unsafe-inline' https://*.nootapp.io
```

**After (Secure):**
```
script-src 'self' 'nonce-<random>' 'unsafe-eval' https://*.nootapp.io
style-src 'self' 'nonce-<random>' https://*.nootapp.io
```

This change prevents execution of any injected malicious scripts or styles, as they won't have the correct nonce value.

## Usage Guide

### Accessing CSP Nonce in Components

The CSP nonce is available through SvelteKit's layout data:

```typescript
// In a +page.svelte or component
<script lang="ts">
  import { page } from '$app/stores'
  
  // Access nonce from layout data
  $: cspNonce = $page.data.cspNonce
</script>
```

### Creating Dynamic Scripts/Styles Safely

Use the provided utilities for dynamic content:

```typescript
import { createNoncedScript, createNoncedStyle, setSafeInnerHTML } from '$lib/utils/csp'

// Create a dynamic script
const script = createNoncedScript('console.log("Hello")', cspNonce)
document.body.appendChild(script)

// Create a dynamic style
const style = createNoncedStyle('body { color: red; }', cspNonce)
document.head.appendChild(style)

// Safely set innerHTML with automatic nonce injection
setSafeInnerHTML(container, htmlString, cspNonce)
```

### Development Debugging

CSP debugging is automatically enabled in development mode and provides:
- Real-time CSP violation logging
- Helpful suggestions for fixing violations
- Detailed violation information in the console

## Template Integration

### Critical CSS in app.html

The main app template uses SvelteKit's nonce placeholder:

```html
<style nonce="%sveltekit.csp.nonce%">
  /* Critical CSS styles */
</style>
```

This placeholder is automatically replaced with the actual nonce during rendering.

### Component Inline Styles

For Svelte components with `<style>` blocks, nonces are automatically injected during the build process. No manual intervention required.

## Development vs Production

- **Development Mode**: CSP enforcement is disabled for convenience and faster iteration
- **Production Mode**: Full CSP enforcement with nonce-based directives

This is controlled by the `PUBLIC_NODE_ENV` environment variable.

## Testing

The implementation includes comprehensive unit tests covering:

- Unique nonce generation per request
- HTML transformation with nonce injection
- SvelteKit placeholder replacement
- CSP utility functions
- Nonce validation and security
- CSP debugging functionality

Run tests with:
```bash
cd apps/web && npm run test
```

## Browser Support

Nonce-based CSP is supported by all modern browsers:
- Chrome 40+
- Firefox 31+
- Safari 10+
- Edge 12+

## Troubleshooting

### Common CSP Violations

1. **Inline Scripts Without Nonces**
   ```
   Error: Refused to execute inline script because it violates CSP
   ```
   **Solution**: Use `createNoncedScript()` or ensure the script has a nonce attribute

2. **Dynamic HTML Injection**
   ```
   Error: Refused to execute inline event handler because it violates CSP
   ```
   **Solution**: Use `setSafeInnerHTML()` instead of direct `innerHTML` assignment

3. **Third-Party Scripts**
   ```
   Error: Refused to load script from 'https://example.com/script.js'
   ```
   **Solution**: Add the domain to the CSP policy in `hooks.server.ts`

### Debugging Tools

- Enable CSP debugging in development mode
- Use browser DevTools Security tab to inspect CSP policy
- Check Network tab for CSP violation reports
- Validate nonces with `isValidNonce()` utility

## Security Considerations

- Nonces are regenerated for each request, preventing replay attacks
- The implementation follows OWASP CSP best practices
- Regular security audits should validate the CSP policy effectiveness
- Keep third-party domains in CSP policy minimal and reviewed
- Monitor CSP violation reports in production

## Files Modified

- `apps/web/src/hooks.server.ts` - Main CSP implementation
- `apps/web/src/routes/+layout.server.ts` - Nonce propagation to client
- `apps/web/src/routes/+layout.svelte` - CSP debugging integration
- `apps/web/src/app.html` - Template nonce placeholder
- `apps/web/src/app.d.ts` - TypeScript definitions
- `apps/web/src/lib/utils/csp.ts` - Client-side utilities
- `apps/web/src/hooks.server.test.ts` - Unit tests
- `apps/web/src/lib/utils/csp.test.ts` - Utility tests

## Future Enhancements

- **CSP Reporting**: Implement CSP violation reporting endpoint
- **Strict Dynamic**: Consider upgrading to `'strict-dynamic'` CSP
- **Hash-Based CSP**: Alternative to nonces for static content
- **Automated Testing**: CSP policy validation in CI/CD pipeline