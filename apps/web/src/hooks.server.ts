import { createServerClient } from '@supabase/ssr';
import { redirect, type Handle } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';
import { env } from '$env/dynamic/public';
import type { User } from '$lib/auth/provider';

const supabase: Handle = async ({ event, resolve }) => {
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
            event.cookies.set(name, value, { ...options, path: '/', secure: true });
          });
        },
      },
    }
  ) as unknown as App.Locals['supabase'];

  event.locals.safeGetSession = async () => {
    const { data: { session } } = await event.locals.supabase.auth.getSession();
    return { session, user: session?.user ?? null, amr: null };
  };

  return resolve(event);
};

const authGuard: Handle = async ({ event, resolve }) => {
  const { session } = await event.locals.safeGetSession();
  event.locals.session = session;
  
  // Map Supabase user to our User interface if session exists
  if (session && session.user) {
    // Try to get additional user data from our profiles table
    let appUser: User;
    try {
      const { data, error } = await event.locals.supabase
        .from('profiles')
        .select('id, subscription_tier')
        .eq('id', session.user.id)
        .single();
      
      if (error || !data) {
        // Fallback to basic user data from session
        appUser = {
          id: session.user.id,
          email: session.user.email || '',
          subscriptionTier: 'free', // Default fallback
          provider: 'supabase',
          subject: session.user.id
        };
      } else {
        appUser = {
          id: data.id,
          email: session.user.email || '',
          subscriptionTier: (data.subscription_tier as 'free' | 'pro') || 'free',
          provider: 'supabase',
          subject: session.user.id
        };
      }
    } catch (error: unknown) {
      const supabaseError = error as { message?: string; code?: string; details?: string };
      console.warn('Failed to fetch user profile data:', {
        message: supabaseError.message,
        code: supabaseError.code,
        details: supabaseError.details
      });
      // Fallback to basic user data
      appUser = {
        id: session.user.id,
        email: session.user.email || '',
        subscriptionTier: 'free',
        provider: 'supabase',
        subject: session.user.id
      };
    }
    
    event.locals.user = appUser;
  } else {
    event.locals.user = null;
  }

  // Protect routes - redirect to login if not authenticated
  if (!session && /^(\/summary|\/profile|\/record)(\/|$)/.test(event.url.pathname)) {
    throw redirect(303, '/login?returnUrl=' + encodeURIComponent(event.url.pathname + event.url.search));
  }

  return resolve(event);
};

export const handle = sequence(supabase, authGuard);
