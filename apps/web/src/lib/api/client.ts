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

// CSRF token management
let currentCSRFToken: string | null = null;

// Get CSRF token from localStorage or request a new one
async function getCSRFToken(): Promise<string | null> {
  if (typeof window === 'undefined') return null;
  
  // Try to get token from memory first
  if (currentCSRFToken) {
    return currentCSRFToken;
  }
  
  // Try to get a new token by making a GET request to any authenticated endpoint
  try {
    const response = await fetch(`${PUBLIC_API_BASE_URL}/health`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${await getAccessToken() || ''}`
      }
    });
    
    if (response.ok) {
      const token = response.headers.get('X-CSRF-Token');
      if (token) {
        currentCSRFToken = token;
        return token;
      }
    }
  } catch (error) {
    console.warn('Failed to get CSRF token:', error);
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
        
        // Add CSRF token for state-changing operations
        if (prop === 'POST' || prop === 'PUT' || prop === 'DELETE' || prop === 'PATCH') {
          const csrfToken = await getCSRFToken();
          if (csrfToken) {
            typedInit.headers['X-CSRF-Token'] = csrfToken;
          }
        }
        
        // Call the original method
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const response = await (originalMethod as (...args: any[]) => any).call(target, url, typedInit);
        
        // Update CSRF token from response headers if available
        if (response && response.response && response.response.headers) {
          const newCSRFToken = response.response.headers.get('X-CSRF-Token');
          if (newCSRFToken) {
            currentCSRFToken = newCSRFToken;
          }
        }
        
        return response;
      };
    }
    
    return originalMethod;
  }
});
