// src/hooks.server.ts
import type { Handle } from "@sveltejs/kit"
import { createServerClient } from "@supabase/ssr"
import { env } from "$env/dynamic/public"

export const handle: Handle = async ({ event, resolve }) => {
  // Only create Supabase client if properly configured
  if (env.PUBLIC_SUPABASE_URL && env.PUBLIC_SUPABASE_ANON_KEY && 
      !env.PUBLIC_SUPABASE_URL.includes('REPLACE_ME') && 
      !env.PUBLIC_SUPABASE_ANON_KEY.includes('REPLACE_ME')) {
    
    event.locals.supabase = createServerClient(
      env.PUBLIC_SUPABASE_URL,
      env.PUBLIC_SUPABASE_ANON_KEY,
      {
        cookies: {
          get: (key) => event.cookies.get(key),
          set: (key, value, options) => {
            event.cookies.set(key, value, {
              ...options,
              // Ensure secure cookie settings for production
              secure: true,
              httpOnly: false, // Allow client access for Supabase
              sameSite: 'lax',
              path: '/'
            })
          },
          remove: (key, options) => {
            event.cookies.delete(key, { path: '/', ...options })
          }
        }
      }
    )

    // Get session from server-side cookies
    try {
      const { data: { session } } = await event.locals.supabase.auth.getSession()
      event.locals.session = session
    } catch (error) {
      console.warn('Failed to get server session:', error)
      event.locals.session = null
    }
  } else {
    event.locals.supabase = null
    event.locals.session = null
  }

  return resolve(event, {
    filterSerializedResponseHeaders(name) {
      return name === 'content-range'
    }
  })
}
