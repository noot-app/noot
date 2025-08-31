import { redirect } from "@sveltejs/kit"
import type { RequestHandler } from "./$types"
import { DEFAULT_REDIRECT_PATH, getRedirectParam } from "$lib/utils/redirect"

export const GET: RequestHandler = async ({ url, locals }) => {
  const code = url.searchParams.get("code")
  const redirectParam = getRedirectParam(url, DEFAULT_REDIRECT_PATH)

  // Validate that code parameter exists; if not, prefer forwarding provider error codes
  if (!code) {
    const providerErrorCode =
      url.searchParams.get("error_code") || url.searchParams.get("error")
    const errorParam = providerErrorCode ?? "missing_oauth_code"
    throw redirect(303, `/login?error=${encodeURIComponent(errorParam)}`)
  }

  // Sanitize redirect to only allow internal paths (security measure)
  const sanitizedRedirect = redirectParam

  try {
    // Exchange the OAuth code for a session using Supabase SSR
    const { error } = await locals.supabase.auth.exchangeCodeForSession(code)

    if (error) {
      console.error("OAuth callback error:", error)
      throw redirect(303, "/login?error=oauth_exchange_failed")
    }

    // Successful OAuth - redirect to target location
    throw redirect(303, sanitizedRedirect)
  } catch (err) {
    console.error("OAuth callback exception:", err)
    throw redirect(303, "/login?error=oauth_callback_failed")
  }
}
