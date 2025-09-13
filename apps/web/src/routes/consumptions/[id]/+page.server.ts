import { createServerApiClientFromSession } from "$lib/api/server-client"

/**
 * Format API error for user display
 */
function formatApiError(error: unknown): string {
  if (typeof error === "string") {
    return error
  }
  
  if (error && typeof error === "object") {
    // Handle common API error formats
    if ("message" in error && typeof error.message === "string") {
      return error.message
    }
    if ("error" in error && typeof error.error === "string") {
      return error.error
    }
  }
  
  // Fallback message for consumption access errors
  return "This consumption doesn't exist or you don't have permission to view it."
}

export async function load({ params, locals }) {
  const id = params.id

  // Check if user is authenticated using session
  const session = await locals.getSession()
  const isUnauthenticated = !session

  // Create authenticated API client for server-side use
  const serverApiClient = createServerApiClientFromSession(session)

  // Always try to load the consumption first (with auth headers if available)
  const consumptionRes = await serverApiClient.GET("/consumption/{id}", { params: { path: { id } } })

  // If consumption failed to load, return error state
  if (consumptionRes.error || !consumptionRes.data) {
    return {
      consumption: null,
      goalsAuto: null,
      goalsDri: null,
      error: formatApiError(consumptionRes.error),
      isUnauthenticated: true
    }
  }

  // Only load goals if user is authenticated
  let goalsAuto = null
  let goalsDri = null

  if (!isUnauthenticated) {
    // User is authenticated - load their goals
    try {
      const [autoResult, driResult] = await Promise.allSettled([
        serverApiClient.GET("/goals", { params: { query: { source: "auto" } } }),
        serverApiClient.GET("/goals", { params: { query: { source: "dri" } } })
      ])

      goalsAuto = autoResult.status === 'fulfilled' && !autoResult.value.error ? 
        autoResult.value.data?.goals : null
      goalsDri = driResult.status === 'fulfilled' && !driResult.value.error ? 
        driResult.value.data?.goals : null
    } catch (err) {
      console.log("Goals loading failed for authenticated user:", err)
    }
  }

  return {
    consumption: consumptionRes.data,
    goalsAuto,
    goalsDri,
    isUnauthenticated
  }
}
