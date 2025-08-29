/**
 * Documentation for the Supabase Authentication Implementation
 * 
 * This document outlines the security, performance, and architectural decisions
 * made in the authentication system for the SvelteKit + Cloudflare SSR + Supabase stack.
 */

## Architecture Overview

The authentication system is designed around several key principles:

1. **Security First**: All auth operations use `getUser()` instead of `getSession()` for server-side validation
2. **Performance Optimized**: Caching, rate limiting, and smart client/server coordination
3. **SSR Compatible**: Seamless handoff between server-side and client-side auth states
4. **Production Safe**: Environment-aware logging and error sanitization

## Key Components

### Core Files

- **`hooks.server.ts`**: Server-side auth middleware with route protection
- **`supabase-auth.ts`**: Main auth provider with comprehensive error handling
- **`store.ts`**: Svelte stores for reactive auth state management
- **`load_helpers.ts`**: Utilities for consistent SSR/client session handling

### Utility Modules

- **`logger.ts`**: Production-safe logging with environment-based levels
- **`errors.ts`**: Error sanitization and custom error classes
- **`user-cache.ts`**: User profile caching with automatic expiration
- **`rate-limit.ts`**: Rate limiting and retry logic for API calls

## Security Features

### Authentication Flow
1. Server validates tokens using `getUser()` (not `getSession()`)
2. User profiles fetched with RLS (Row Level Security) protection
3. Client-side auth state synchronized with server state
4. Invalid sessions immediately rejected

### Error Handling
- Production errors are sanitized to prevent information leakage
- Development errors include full details for debugging
- Specific error codes mapped to user-friendly messages
- Rate limiting prevents brute force attacks

### Session Management
- JWT tokens cached with proper expiration handling
- Automatic token refresh with exponential backoff
- Session cleanup on logout
- Cross-tab session synchronization

## Performance Optimizations

### Caching Strategy
- **User Profiles**: 5-minute cache with automatic cleanup
- **Access Tokens**: Cached until 1 minute before expiry
- **Auth State**: Deduplication prevents redundant updates

### API Call Reduction
- Client reuses server session when available
- Rate limiting prevents excessive token refresh calls
- Batch user data fetching where possible
- Smart cache invalidation

### SSR/Client Coordination
- Server populates initial auth state
- Client only verifies if server state unavailable
- Consistent session object structure
- Optimized hydration process

## Usage Examples

### Basic Authentication Check
```typescript
import { currentUser } from '$lib/auth/store';

// Reactive auth state
$: user = $currentUser;
$: isAuthenticated = !!user;
$: isPro = user?.subscriptionTier === 'pro';
```

### Programmatic Auth Operations
```typescript
import { signIn, signOut } from '$lib/auth/store';

// Sign in
const { user, error } = await signIn('user@example.com', 'password');

// Sign out  
const { error: signOutError } = await signOut();
```

### Route Protection (Server-side)
Routes under `/summary`, `/profile`, and `/record` are automatically protected by the `authGuard` hook in `hooks.server.ts`.

### API Requests with Auth
```typescript
import { apiClient } from '$lib/api/client';

// JWT token automatically included
const response = await apiClient.GET('/api/user/profile');
```

## Environment Configuration

### Required Variables
```env
PUBLIC_SUPABASE_URL=https://your-project.supabase.co
PUBLIC_SUPABASE_ANON_KEY=your-anon-key
```

### Optional Debug Variables
```env
VITE_AUTH_DEBUG=true     # Enable auth debug logging in production
VITE_API_DEBUG=true      # Enable API debug logging in production  
VITE_LOG_LEVEL=debug     # Set minimum log level (debug|info|warn|error)
```

## Best Practices

### For Developers
1. Always handle auth errors gracefully with user-friendly messages
2. Use the provided error classes (`AuthError`, `ApiError`) for consistent error handling
3. Leverage caching utilities to avoid redundant database calls
4. Test auth flows with network failures and edge cases

### For Production
1. Never set `VITE_AUTH_DEBUG=true` in production unless debugging
2. Monitor rate limiting metrics to detect abuse
3. Set up proper error logging and alerting
4. Regularly review auth logs for security issues

## Troubleshooting

### Common Issues
- **"Auth session missing"**: Normal when user not logged in, not an error
- **Rate limit exceeded**: Indicates too many token refresh attempts
- **Cache misses**: May indicate memory pressure or session instability

### Debug Mode
Enable debug logging with `VITE_AUTH_DEBUG=true` to see detailed auth flow information.

### Performance Monitoring
Monitor these metrics:
- Cache hit/miss ratios
- Token refresh frequency  
- Database query counts
- API response times