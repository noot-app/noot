import createClient from 'openapi-fetch';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { dev } from '$app/environment';
import type { paths } from './schema';

// Create base client
const baseClient = createClient<paths>({ 
  baseUrl: PUBLIC_API_BASE_URL 
});

// Wrap client to add dev user headers when in development mode
// TODO: When implementing Supabase auth, replace X-Dev-User-ID header with Authorization header
export const apiClient = new Proxy(baseClient, {
  get(target, prop) {
    const originalMethod = target[prop as keyof typeof target];
    
    if (typeof originalMethod === 'function' && (prop === 'GET' || prop === 'POST' || prop === 'PUT' || prop === 'DELETE' || prop === 'PATCH')) {
      return function(url: string, init?: any) {
        // Add dev user header in development mode
        if (dev && typeof localStorage !== 'undefined') {
          const selectedUser = localStorage.getItem('dev-selected-user');
          if (selectedUser) {
            init = init || {};
            init.headers = {
              ...init.headers,
              'X-Dev-User-ID': selectedUser
            };
          }
        }
        
        return originalMethod.call(target, url, init);
      };
    }
    
    return originalMethod;
  }
});