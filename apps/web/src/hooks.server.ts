import { createServerClient } from '@supabase/ssr';
import { redirect, type Handle } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';
import { env } from '$env/dynamic/public';
import type { User } from '$lib/auth/provider';
import type { Session } from '@supabase/supabase-js';
import type { Database } from './DatabaseDefinitions';

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
        setAll: (cookiesToSet: Array<{ name: string; value: string; options: any }>) => {
          cookiesToSet.forEach(({ name, value, options }) => {
            event.cookies.set(name, value, { ...options, path: '/' });
          });
        },
      },
    }
  ) as any;

  event.locals.safeGetSession = async () => {
    const { data: { session } } = await event.locals.supabase.auth.getSession();
    if (!session) return { session: null, user: null, amr: null };
    
    const { data: { user }, error } = await event.locals.supabase.auth.getUser();
    if (error) return { session: null, user: null, amr: null };
    
    return { session, user, amr: null };
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
        .select('id')
        .eq('id', session.user.id)
        .single();
      
      if (error || !data) {
        // Fallback to basic user data from session
        appUser = {
          id: session.user.id,
          email: session.user.email || '',
          subscriptionTier: 'free',
          provider: 'supabase',
          subject: session.user.id
        };
      } else {
        appUser = {
          id: data.id,
          email: session.user.email || '',
          subscriptionTier: 'free', // Default to free since column doesn't exist
          provider: 'supabase',
          subject: session.user.id
        };
      }
    } catch (error) {
      console.warn('Failed to fetch user profile data:', error);
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
  if (!session && event.url.pathname.match(/^\/(summary|profile|record)/)) {
    throw redirect(303, '/login?returnUrl=' + encodeURIComponent(event.url.pathname + event.url.search));
  }

  return resolve(event);
};

export const handle = sequence(supabase, authGuard);
