import { createServerClient } from '@supabase/ssr';
import { redirect, type Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';
import { dev } from '$app/environment';
import type { Session } from '@supabase/supabase-js';
import { getValidatedSession } from '$lib/utils.js';

export const handle: Handle = async ({ event, resolve }) => {
  // Ensure environment variables are available
  const supabaseUrl = env.PUBLIC_SUPABASE_URL;
  const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY;
  
  if (!supabaseUrl || !supabaseAnonKey) {
    console.warn('Supabase environment variables not configured');
    return resolve(event);
  }

  event.locals.supabase = createServerClient(
    supabaseUrl,
    supabaseAnonKey,
    {
      cookies: {
        getAll: () => event.cookies.getAll(),
        setAll: (cookiesToSet: Array<{ name: string; value: string; options: Record<string, unknown> }>) => {
          cookiesToSet.forEach(({ name, value, options }) => {
            // Enforce security settings for auth cookies in production
            const secureOptions = {
              ...options,
              path: '/',
              // Security hardening for production
              httpOnly: true,
              secure: !dev, // Only secure in production (HTTPS required)
              sameSite: 'lax' as const, // Prevent CSRF while allowing normal navigation
            };
            
            event.cookies.set(name, value, secureOptions);
          });
        },
      },
    }
  ) as unknown as App.Locals['supabase'];

  /**
   * We use getSession, as a function, rather than a static object
   * like `session`, in order to make reactivity work for some
   * features of our pages. For example, if this wasn't a function,
   * things like auth state changes wouldn't correctly update data on pages.
   */
  event.locals.getSession = async (): Promise<Session | null> => {
    return await getValidatedSession(event.locals.supabase);
  }

  const session = await event.locals.getSession();

  /**
   * Only authenticated users can access these paths and their sub-paths.
   * We protect /summary, /profile, and /record routes.
   */
  const protectedPaths = ['/summary', '/profile', '/record'];
  const isProtectedPath = protectedPaths.some(path => 
    event.url.pathname === path || event.url.pathname.startsWith(path + '/')
  );
  
  if (!session && isProtectedPath) {
    throw redirect(303, '/login?returnUrl=' + encodeURIComponent(event.url.pathname + event.url.search));
  }

  return resolve(event, {
    filterSerializedResponseHeaders(name) {
      return name === 'content-range' || name === 'x-supabase-api-version'
    },
  });
};
