import createClient from "openapi-fetch"
import { env } from "$env/dynamic/public"
import type { paths } from "./schema"

// Create base client with runtime environment variable
const apiBaseUrl = env.PUBLIC_API_BASE_URL || "https://api.nootapp.io/api/v1"

const baseClient = createClient<paths>({
  baseUrl: apiBaseUrl,
})

// Get access token from the new auth system
async function getAccessToken(): Promise<string | null> {
  if (typeof window === "undefined") return null

  try {
    // Dynamic import to avoid SSR issues
    const { getAccessToken } = await import("$lib/auth/store")
    return await getAccessToken()
  } catch (error) {
    console.warn("Failed to get access token:", error)
  }

  return null
}

// Simple in-flight request map to coalesce duplicate concurrent requests
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const inFlight: Map<string, Promise<any>> = new Map()

// Wrap client to add authentication headers and coalesce requests
export const apiClient = new Proxy(baseClient, {
  get(target, prop) {
    const originalMethod = target[prop as keyof typeof target]

    if (
      typeof originalMethod === "function" &&
      (prop === "GET" ||
        prop === "POST" ||
        prop === "PUT" ||
        prop === "DELETE" ||
        prop === "PATCH")
    ) {
      return async function (url: string, init?: unknown) {
        init = init || {}
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const typedInit = init as Record<string, any>
        typedInit.headers = typedInit.headers || {}

        // Add JWT authorization header
        const accessToken = await getAccessToken()
        if (accessToken) {
          typedInit.headers["Authorization"] = `Bearer ${accessToken}`
        }

        // Build a key to detect duplicate concurrent requests
        const keyParts = [String(prop), url]
        if (typedInit.params) keyParts.push(JSON.stringify(typedInit.params))
        if (typedInit.body) keyParts.push(JSON.stringify(typedInit.body))
        const cacheKey = keyParts.join("::")

        // Return existing promise if same request is in-flight
        if (inFlight.has(cacheKey)) return inFlight.get(cacheKey)!

        // Note: CSRF protection not needed for Bearer token auth
        // Bearer tokens are not sent automatically by browsers, so CSRF attacks
        // cannot make the victim's browser include the Authorization header

        // Call the original method
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const p = (originalMethod as (...args: any[]) => any)
          .call(target, url, typedInit)
          .finally(() => inFlight.delete(cacheKey))
        inFlight.set(cacheKey, p)
        return p
      }
    }

    return originalMethod
  },
})
