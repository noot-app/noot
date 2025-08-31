import { dev } from "$app/environment"
import { env } from "$env/dynamic/public"

/** @type {import('./$types').LayoutServerLoad} */
export async function load({ locals }) {
  // Get session using the new getSession function (if available)
  let session = null
  try {
    if (locals.getSession) {
      session = await locals.getSession()
    }
  } catch (error) {
    console.warn("Session loading failed:", error)
  }

  return {
    // Pass environment info to client
    isDevMode: dev,
    supabaseEnabled: !!(
      env.PUBLIC_SUPABASE_URL &&
      env.PUBLIC_SUPABASE_ANON_KEY &&
      !env.PUBLIC_SUPABASE_URL.includes("REPLACE_ME") &&
      !env.PUBLIC_SUPABASE_ANON_KEY.includes("REPLACE_ME")
    ),

    // Pass SSR session data to client
    session,
  }
}
