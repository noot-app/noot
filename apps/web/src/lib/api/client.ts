import createClient from 'openapi-fetch';
import { env } from '$env/dynamic/public';
import type { paths } from './schema';

// Debug logging for API base URL (can remove after confirming it works)
console.log('API Base URL (runtime):', env.PUBLIC_API_BASE_URL);

// Create base client with runtime environment variable
const baseClient = createClient<paths>({ 
  baseUrl: env.PUBLIC_API_BASE_URL || 'https://api.nootapp.io/api/v1'
});

// Get access token from the auth provider (if available)
async function getAccessToken(): Promise<string | null> {
  if (typeof window === 'undefined') return null;
  
  try {
    // Dynamic import to avoid SSR issues
    const { authProvider } = await import('$lib/auth/store');
    
    if (authProvider && 'getAccessToken' in authProvider && typeof authProvider.getAccessToken === 'function') {
      return await authProvider.getAccessToken();
    }
  } catch (error) {
    console.warn('Failed to get access token:', error);
  }
  
  return null;
}

// Wrap client to add authentication headers
export const apiClient = new Proxy(baseClient, {
  get(target, prop) {
    const originalMethod = target[prop as keyof typeof target];
    
    if (typeof originalMethod === 'function' && (prop === 'GET' || prop === 'POST' || prop === 'PUT' || prop === 'DELETE' || prop === 'PATCH')) {
      return async function(url: string, init?: unknown) {
        init = init || {};
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const typedInit = init as Record<string, any>;
        typedInit.headers = typedInit.headers || {};

        // Add JWT authorization header
        const accessToken = await getAccessToken();
        if (accessToken) {
          typedInit.headers['Authorization'] = `Bearer ${accessToken}`;
        }
        
        // Note: CSRF protection not needed for Bearer token auth
        // Bearer tokens are not sent automatically by browsers, so CSRF attacks
        // cannot make the victim's browser include the Authorization header
        
        // Call the original method
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return (originalMethod as (...args: any[]) => any).call(target, url, typedInit);
      };
    }
    
    return originalMethod;
  }
});
