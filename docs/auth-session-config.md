# Authentication Session Configuration

## Session Persistence (30-Day Configuration)

This document outlines the configuration changes made to improve session persistence and eliminate authentication stuttering.

### Supabase Session Configuration

The application is configured for 30-day session persistence through the following settings:

#### Client Configuration (`apps/web/src/lib/supabase.ts`)
- **Client Type**: Uses `createBrowserClient` from `@supabase/ssr` for better SSR support
- **Storage**: Uses `localStorage` for client-side session persistence
- **Storage Key**: Custom key `noot-supabase-auth-token` for session storage
- **Auto Refresh**: Enabled to automatically refresh expired tokens
- **Session Detection**: Automatically detects sessions from URL parameters

#### Server Configuration (`apps/web/src/hooks.server.ts`)
- **Server Client**: Uses `createServerClient` from `@supabase/ssr`
- **Cookie Settings**: 
  - `secure: true` for HTTPS-only cookies
  - `httpOnly: false` to allow client-side access for Supabase
  - `sameSite: 'lax'` for cross-site compatibility
  - `path: '/'` for site-wide availability

### Session Duration in Supabase Dashboard

To ensure 30-day session persistence, verify these settings in your Supabase dashboard:

1. **Go to Authentication > Settings**
2. **JWT Settings**:
   - **JWT expiry**: Set to `3600` seconds (1 hour) - this is the access token expiry
   - **Refresh Token Rotation**: Should be enabled
   - **Refresh Token Reuse Interval**: Set to appropriate value (e.g., 10 seconds)

3. **Session Settings**:
   - **Session timeout**: Should be set to `2592000` seconds (30 days)
   - **Refresh token expiry**: Should be set to `2592000` seconds (30 days)

### How It Works

1. **Initial Load**: Server-side hooks check for existing session in cookies
2. **Client Hydration**: Session is passed from server to client, preventing flash of login screen
3. **Token Refresh**: Client automatically refreshes tokens before expiry
4. **Persistence**: Sessions are stored in localStorage for 30-day persistence
5. **Synchronization**: Server and client session states are kept in sync

### Authentication Flow Improvements

1. **No Login Screen Flash**: Server session is passed to client during hydration
2. **Faster Auth Checks**: AuthGuard component has improved timing and state management
3. **Graceful Fallbacks**: Multiple fallback mechanisms for session recovery
4. **Better Error Handling**: Improved error handling for auth state transitions

### Testing Session Persistence

To test 30-day session persistence:

1. Log into the application
2. Close browser completely
3. Reopen browser after a day/week
4. Navigate to protected route (e.g., `/profile`)
5. Should load without requiring re-authentication

### Troubleshooting

If users are still being logged out frequently:

1. **Check Supabase Dashboard Settings**: Ensure JWT and session timeouts are set correctly
2. **Browser Storage**: Verify localStorage is not being cleared by browser settings
3. **Network Issues**: Check for network connectivity during token refresh attempts
4. **Server Logs**: Monitor server logs for authentication errors during token refresh
