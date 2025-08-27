import createClient from 'openapi-fetch';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { paths } from './schema';

// Create base client
const baseClient = createClient<paths>({ 
  baseUrl: PUBLIC_API_BASE_URL 
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
        
        // Call the original method
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return (originalMethod as (...args: any[]) => any).call(target, url, typedInit);
      };
    }
    
    return originalMethod;
  }
});
