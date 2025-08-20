import createClient from 'openapi-fetch';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { paths } from './schema';

export const apiClient = createClient<paths>({ 
  baseUrl: PUBLIC_API_BASE_URL 
});