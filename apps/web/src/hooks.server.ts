import { createServerClient } from "@supabase/ssr"
import { redirect, type Handle } from "@sveltejs/kit"
import { env } from "$env/dynamic/public"
import type { Session } from "@supabase/supabase-js"
import { getValidatedSession } from "$lib/utils.js"
import { DEFAULT_REDIRECT_PATH, getRedirectParam } from "$lib/utils/redirect"

export const handle: Handle = async ({ event, resolve }) => {
  // Ensure environment variables are available
  const supabaseUrl = env.PUBLIC_SUPABASE_URL
  const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY
  // Default to production unless explicitly set to development
  const nodeEnv = (env.PUBLIC_NODE_ENV || "production").toLowerCase()
  const isDevelopment = nodeEnv === "development"

  if (!supabaseUrl || !supabaseAnonKey) {
    console.warn("Supabase environment variables not configured")
    return resolve(event)
  }

  event.locals.supabase = createServerClient(supabaseUrl, supabaseAnonKey, {
    cookies: {
      getAll: () => event.cookies.getAll(),
      setAll: (
        cookiesToSet: Array<{
          name: string
          value: string
          options: Record<string, unknown>
        }>,
      ) => {
        cookiesToSet.forEach(({ name, value, options }) => {
          event.cookies.set(name, value, {
            ...options,
            path: "/",
          })
        })
      },
    },
  }) as unknown as App.Locals["supabase"]

  /**
   * We use getSession, as a function, rather than a static object
   * like `session`, in order to make reactivity work for some
   * features of our pages. For example, if this wasn't a function,
   * things like auth state changes wouldn't correctly update data on pages.
   */
  event.locals.getSession = async (): Promise<Session | null> => {
    return await getValidatedSession(event.locals.supabase)
  }

  const session = await event.locals.getSession()

  /**
   * Only authenticated users can access these paths and their sub-paths.
   * We protect /summary, /profile, and /record routes.
   */
  const protectedPaths = ["/summary", "/profile", "/record", "/labels", "/events", "/api-keys", "/log", "/quick", "/dashboard"]
  const isProtectedPath = protectedPaths.some(
    (path) =>
      event.url.pathname === path || event.url.pathname.startsWith(path + "/"),
  )

  if (!session && isProtectedPath) {
    throw redirect(
      303,
      "/login?redirect=" +
        encodeURIComponent(event.url.pathname + event.url.search),
    )
  }

  // If authenticated, avoid staying on auth pages. Respect the redirect query param when present.
  if (
    session &&
    (event.url.pathname === "/login" || event.url.pathname === "/signup")
  ) {
    const target = getRedirectParam(event.url, DEFAULT_REDIRECT_PATH)
    throw redirect(303, target)
  }

  return resolve(event, {
    filterSerializedResponseHeaders(name) {
      return name === "content-range" || name === "x-supabase-api-version"
    },
    transformPageChunk({ html, done }) {
      // Add security headers on final response (only in production)
      if (done && !isDevelopment) {
        // Set security headers (CSP is now handled by SvelteKit config)
        event.setHeaders({
          // Legacy frame protection
          'X-Frame-Options': 'DENY',
          // MIME type sniffing protection
          'X-Content-Type-Options': 'nosniff',
          // Referrer policy for privacy
          'Referrer-Policy': 'strict-origin-when-cross-origin',
          // Force HTTPS everywhere (adjust max-age as needed)
          'Strict-Transport-Security': 'max-age=63072000; includeSubDomains; preload',
          // Disable potentially dangerous browser features (allow microphone for voice logging and maybe future camera usage for labels/barcodes)
          'Permissions-Policy': 'geolocation=(), camera=(self), microphone=(self)'
        })
      }
      return html
    },
  })
}
