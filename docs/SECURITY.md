# Security Guidelines for SvelteKit + Supabase Auth Integration

This document outlines the security measures implemented in the Noot application to protect against common web application vulnerabilities.

## Summary of Security Fixes

### 1. URL Sanitization for Open Redirect Prevention

**Issue**: Login and signup pages accepted unsanitized `returnUrl` parameters that could redirect users to malicious external sites.

**Fix**: Implemented comprehensive URL sanitization in `src/lib/security.ts`:

```typescript
import { sanitizeReturnUrl } from '$lib/security';

// Before (vulnerable):
const returnUrl = $page.url.searchParams.get('returnUrl') || '/';

// After (secure):
const returnUrl = sanitizeReturnUrl($page.url.searchParams.get('returnUrl'));
```

**Protection Against**:
- External domain redirects (`https://evil.com/phishing`)
- Path traversal attacks (`../../../admin`)
- JavaScript/data URL schemes (`javascript:alert(1)`)
- Protocol-relative URLs (`//malicious.com`)

**Files Updated**:
- `src/routes/login/+page.svelte`
- `src/routes/signup/+page.svelte`

### 2. Secure Cookie Configuration

**Issue**: Authentication cookies were set without proper security attributes, making them vulnerable to XSS and CSRF attacks.

**Fix**: Enhanced cookie security in `src/hooks.server.ts`:

```typescript
// Security hardening for production
const secureOptions = {
  ...options,
  path: '/',
  httpOnly: true,        // Prevents client-side JS access
  secure: !dev,          // HTTPS-only in production
  sameSite: 'lax' as const, // CSRF protection
};
```

**Protection Against**:
- Cross-site scripting (XSS) cookie theft
- Cross-site request forgery (CSRF)
- Man-in-the-middle attacks (HTTPS enforcement)

### 3. Authorization Header Origin Restriction

**Issue**: The API client attached JWT tokens to all HTTP requests, potentially leaking credentials to third-party services.

**Fix**: Implemented origin validation in `src/lib/api/client.ts`:

```typescript
import { shouldAttachAuthHeader } from '$lib/security';

// Only add auth headers for trusted origins
if (shouldAttachAuthHeader(url, apiBaseUrl)) {
  const accessToken = await getAccessToken();
  if (accessToken) {
    typedInit.headers['Authorization'] = `Bearer ${accessToken}`;
  }
}
```

**Protection Against**:
- Token leakage to external APIs
- Credential theft via malicious redirects
- CSRF attacks using leaked tokens

### 4. Production Debug Information Removal

**Issue**: Debug logging exposed internal API configuration in production builds.

**Fix**: Removed production console logging that revealed:
- API base URLs
- Environment configuration
- Internal routing information

## Security Utility Functions

### `sanitizeReturnUrl(url, defaultUrl = '/')`

Validates and sanitizes redirect URLs to prevent open redirect attacks.

**Parameters**:
- `url`: The URL to sanitize (from query parameters)
- `defaultUrl`: Fallback URL if input is invalid (default: '/')

**Returns**: A safe, validated relative path

**Examples**:
```typescript
sanitizeReturnUrl('/dashboard');           // ✅ '/dashboard'
sanitizeReturnUrl('https://evil.com');     // ❌ '/' (external URL)
sanitizeReturnUrl('javascript:alert(1)');  // ❌ '/' (dangerous scheme)
sanitizeReturnUrl('/../admin');            // ❌ '/' (path traversal)
```

### `shouldAttachAuthHeader(requestUrl, apiBaseUrl)`

Determines if Authorization headers should be attached to requests.

**Parameters**:
- `requestUrl`: The URL being requested
- `apiBaseUrl`: The trusted API base URL

**Returns**: `true` if auth headers should be attached

**Examples**:
```typescript
shouldAttachAuthHeader('/api/users', 'https://api.noot.app');           // ✅ true (relative)
shouldAttachAuthHeader('https://api.noot.app/v1', 'https://api.noot.app'); // ✅ true (same origin)
shouldAttachAuthHeader('https://evil.com/steal', 'https://api.noot.app');   // ❌ false (external)
```

### `isSafeRedirectUrl(url)`

Quick check for redirect URL safety.

**Parameters**:
- `url`: The URL to validate

**Returns**: `true` if the URL is safe for redirects

## Testing Security Measures

The security implementations include comprehensive test coverage:

- **23 security-specific tests** in `src/lib/security.test.ts`
- **Attack vector validation** (XSS, CSRF, open redirects)
- **Edge case handling** (malformed URLs, encoding issues)
- **Integration tests** for API client and hooks

Run security tests:
```bash
npm run test -- src/lib/security.test.ts
```

## Environment Considerations

### Development vs Production

- **Development**: Cookies use `secure: false` to work with HTTP
- **Production**: Cookies use `secure: true` to enforce HTTPS

### Configuration Validation

Ensure these environment variables are properly configured:
- `PUBLIC_SUPABASE_URL`: Your Supabase project URL
- `PUBLIC_SUPABASE_ANON_KEY`: Your Supabase anonymous key
- `PUBLIC_API_BASE_URL`: Your trusted API base URL

## Future Considerations

### Content Security Policy (CSP)

Consider implementing CSP headers for additional XSS protection:
```typescript
// In hooks.server.ts
event.setHeaders({
  'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-inline'"
});
```

### Rate Limiting

Implement rate limiting for authentication endpoints to prevent brute force attacks.

### Security Headers

The Go backend already implements security headers. Consider adding:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Referrer-Policy: strict-origin-when-cross-origin`

### Regular Security Audits

1. Review authentication flows regularly
2. Update dependencies to patch vulnerabilities
3. Monitor for new attack vectors
4. Test security measures with penetration testing tools

## Incident Response

If a security vulnerability is discovered:

1. **Assess impact**: Determine affected users and data
2. **Patch immediately**: Deploy fixes to production
3. **Notify users**: If credentials may be compromised
4. **Update documentation**: Record lessons learned
5. **Review related code**: Look for similar vulnerabilities

## Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [SvelteKit Security Best Practices](https://kit.svelte.dev/docs/security)
- [Supabase Auth Security](https://supabase.com/docs/guides/auth/auth-deep-dive/auth-deep-dive-jwts)
- [Web Security Guidelines](https://infosec.mozilla.org/guidelines/web_security)