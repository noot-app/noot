import createClient from "openapi-fetch"
import { env } from "$env/dynamic/public"
import type { paths } from "./schema"

/**
 * Creates an authenticated API client for server-side use.
 * This is separate from the main apiClient which is designed for browser use.
 * 
 * @param accessToken - JWT access token from Supabase session
 * @returns Configured API client with authentication headers
 */
export function createServerApiClient(accessToken?: string) {
  const apiBaseUrl = env.PUBLIC_API_BASE_URL || "https://api.nootapp.io/api/v1"
  
  return createClient<paths>({
    baseUrl: apiBaseUrl,
    headers: accessToken ? {
      'Authorization': `Bearer ${accessToken}`
    } : {}
  })
}

/**
 * Convenience function to create an authenticated server API client from a session.
 * 
 * @param session - Supabase session object with access_token property
 * @returns Configured API client with authentication headers
 */
export function createServerApiClientFromSession(session: { access_token?: string } | null) {
  return createServerApiClient(session?.access_token)
}
