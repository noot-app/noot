# CSP Nonce Implementation

## Overview

This document describes the implementation of nonce-based Content Security Policy (CSP) in the SvelteKit frontend to eliminate the use of 'unsafe-inline' directives, thereby improving the application's security posture against XSS attacks.

## What Changed

### Before (Insecure)
```
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.nootapp.io https://uygqcgnmlzmuwkpsmixs.supabase.co
style-src 'self' 'unsafe-inline' https://*.nootapp.io
```

### After (Hardened)
```
script-src 'self' 'nonce-<random-value>' 'unsafe-eval' https://*.nootapp.io https://uygqcgnmlzmuwkpsmixs.supabase.co
style-src 'self' 'nonce-<random-value>' https://*.nootapp.io
```

## Implementation Details

### Nonce Generation
- **Location**: `hooks.server.ts`
- **Method**: `generateCSPNonce()` function using `crypto.randomBytes(16).toString('base64')`
- **Uniqueness**: Each HTTP request gets a unique nonce
- **Storage**: Nonce is stored in `event.locals.cspNonce` for access during page generation

### Nonce Injection
- **Process**: During `transformPageChunk` processing, all `<script>` and `<style>` tags are automatically injected with the nonce attribute
- **Regex Pattern**: Case-insensitive matching to handle both lowercase and uppercase HTML tags
- **Target Elements**: 
  - SvelteKit's hydration scripts
  - Inline styles in `app.html`
  - Any other dynamically generated scripts/styles

### CSP Header Update
- **Scope**: Only applied in production (when `PUBLIC_NODE_ENV !== "development"`)
- **Headers**: Both `script-src` and `style-src` directives use nonce-based approach
- **Note**: `'unsafe-eval'` is retained for Supabase compatibility, but `'unsafe-inline'` is completely removed

## Security Benefits

1. **XSS Protection**: Only scripts and styles with matching nonces can execute
2. **No Inline Execution**: Prevents injection of malicious inline scripts/styles
3. **Per-Request Uniqueness**: Each nonce is cryptographically random and unique per request
4. **Standards Compliance**: Follows OWASP CSP best practices

## Development vs Production

- **Development**: CSP nonce injection is disabled to maintain development convenience
- **Production**: Full nonce-based CSP enforcement with headers and HTML transformation

## Testing

The implementation includes comprehensive unit tests:
- Nonce generation uniqueness
- HTML transformation verification
- CSP header validation
- Security regex pattern testing

## Compatibility

- **SvelteKit**: Compatible with SvelteKit 2.x SSR and hydration
- **Cloudflare**: Works with Cloudflare adapter
- **Supabase**: Maintains compatibility with Supabase authentication flows
- **Third-party**: External scripts from allowed domains continue to work

## Files Modified

- `apps/web/src/hooks.server.ts` - Main implementation
- `apps/web/src/app.d.ts` - TypeScript definitions
- `apps/web/src/hooks.server.test.ts` - Unit tests

## Future Considerations

- Monitor for any third-party scripts that may need nonce attributes
- Consider implementing CSP reporting for monitoring violations
- Evaluate removing `'unsafe-eval'` if Supabase alternatives become available