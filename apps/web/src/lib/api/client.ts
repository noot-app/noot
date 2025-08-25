import createClient from 'openapi-fetch';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { dev } from '$app/environment';
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
      return async function(url: string, init?: any) {
        init = init || {};
        init.headers = init.headers || {};

        // Add dev user header in development mode
        if (dev && typeof localStorage !== 'undefined') {
          const selectedUser = localStorage.getItem('dev-selected-user');
          if (selectedUser) {
            init.headers['X-Dev-User-ID'] = selectedUser;
          }
        }

        // Add JWT authorization header in production (or when Supabase is enabled in dev)
        const accessToken = await getAccessToken();
        if (accessToken) {
          init.headers['Authorization'] = `Bearer ${accessToken}`;
        }
        
        return originalMethod.call(target, url, init);
      };
    }
    
    return originalMethod;
  }
});