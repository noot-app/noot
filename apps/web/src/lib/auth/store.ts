import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { invalidateAll } from '$app/navigation';
import { createBrowserClient } from '@supabase/ssr';
import { env } from '$env/dynamic/public';
import type { Session, AuthError, User as SupabaseUser } from '@supabase/supabase-js';

/**
 * Simple auth store based on session from SSR data and browser client for auth actions.
 * This replaces the complex provider system with a simpler approach following
 * the j4w8n/sveltekit-supabase-ssr pattern.
 */

// Create browser client for auth actions (sign in, sign out, etc.)
const getBrowserClient = () => {
  if (!browser) return null;
  
  const supabaseUrl = env.PUBLIC_SUPABASE_URL;
  const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY;
  
  if (!supabaseUrl || !supabaseAnonKey) {
    console.warn('Supabase environment variables not configured');
    return null;
  }
  
  return createBrowserClient(supabaseUrl, supabaseAnonKey);
};

/**
 * Session store - initialized from SSR data in layout and updated on auth changes
 */
export const session = writable<Session | null>(null);

/**
 * Derived user store from session
 */
export const user = derived(session, ($session) => $session?.user ?? null);

/**
 * Derived store to check if user is authenticated
 */
export const isAuthenticated = derived(session, ($session) => !!$session);

/**
 * User profile data store (from our profiles table)
 */
export const userProfile = writable<{
  id: string;
  email: string;
  subscription_tier: 'free' | 'pro';
} | null>(null);

/**
 * Derived store to check if user has pro subscription
 */
export const isPro = derived(userProfile, ($userProfile) => 
  $userProfile?.subscription_tier === 'pro'
);

/**
 * Initialize session from SSR data and set up auth state change listener
 */
export function initAuth(initialSession: Session | null = null) {
  if (!browser) return;
  
  // Set initial session from SSR
  if (initialSession) {
    session.set(initialSession);
  }
  
  const supabase = getBrowserClient();
  if (!supabase) return;
  
  // Set up auth state change listener
  const { data: { subscription } } = supabase.auth.onAuthStateChange(async (event, newSession) => {
    console.debug('Auth state changed:', event);
    
    // Update session store
    session.set(newSession);
    
    // Invalidate all data to refetch with new auth state
    if (event === 'SIGNED_IN' || event === 'SIGNED_OUT') {
      await invalidateAll();
    }
  });

  // Cleanup subscription on page unload
  if (typeof window !== 'undefined') {
    window.addEventListener('beforeunload', () => {
      subscription.unsubscribe();
    });
  }
}

/**
 * Sign in with email and password
 */
export async function signIn(email: string, password: string): Promise<{ error: AuthError | null }> {
  const supabase = getBrowserClient();
  if (!supabase) {
    return { error: { message: 'Supabase not configured', name: 'configuration_error' } as AuthError };
  }

  const { error } = await supabase.auth.signInWithPassword({
    email,
    password
  });

  return { error };
}

/**
 * Sign up with email and password
 */
export async function signUp(email: string, password: string, metadata?: { fullName?: string }): Promise<{ error: AuthError | null }> {
  const supabase = getBrowserClient();
  if (!supabase) {
    return { error: { message: 'Supabase not configured', name: 'configuration_error' } as AuthError };
  }

  const { error } = await supabase.auth.signUp({
    email,
    password,
    options: {
      data: metadata ? { 
        full_name: metadata.fullName
      } : undefined
    }
  });

  return { error };
}

/**
 * Sign out
 */
export async function signOut(): Promise<{ error: AuthError | null }> {
  const supabase = getBrowserClient();
  if (!supabase) {
    return { error: { message: 'Supabase not configured', name: 'configuration_error' } as AuthError };
  }

  const { error } = await supabase.auth.signOut();
  
  if (!error) {
    // Clear session and profile data
    session.set(null);
    userProfile.set(null);
  }

  return { error };
}

/**
 * Reset password
 */
export async function resetPassword(email: string): Promise<{ error: AuthError | null }> {
  const supabase = getBrowserClient();
  if (!supabase) {
    return { error: { message: 'Supabase not configured', name: 'configuration_error' } as AuthError };
  }

  const { error } = await supabase.auth.resetPasswordForEmail(email);
  return { error };
}

/**
 * Get the current access token for API requests
 */
export async function getAccessToken(): Promise<string | null> {
  const supabase = getBrowserClient();
  if (!supabase) return null;

  const { data: { session: currentSession } } = await supabase.auth.getSession();
  return currentSession?.access_token ?? null;
}
