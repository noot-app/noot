import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { invalidateAll } from '$app/navigation';
import { createBrowserClient } from '@supabase/ssr';
import { env } from '$env/dynamic/public';
import type { Session, AuthError } from '@supabase/supabase-js';

/**
 * Simple auth store based on session from SSR data and browser client for auth actions.
 * This replaces the complex provider system with a simpler approach following
 * the j4w8n/sveltekit-supabase-ssr pattern.
 */

// Singleton browser client and cached session
let browserClient: ReturnType<typeof createBrowserClient> | null = null;
let sessionPromise: Promise<Session | null> | null = null;
let sawFirstAuthEvent = false;

// Create/get browser client for auth actions (sign in, sign out, etc.)
const getBrowserClient = () => {
  if (!browser) return null;

  if (browserClient) return browserClient;

  const supabaseUrl = env.PUBLIC_SUPABASE_URL;
  const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY;

  if (!supabaseUrl || !supabaseAnonKey) {
    console.warn('Supabase environment variables not configured');
    return null;
  }

  browserClient = createBrowserClient(supabaseUrl, supabaseAnonKey);
  return browserClient;
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
  // cache via session store subscription only
  }
  
  const supabase = getBrowserClient();
  if (!supabase) return;
  
  // Set up auth state change listener
  const { data: { subscription } } = supabase.auth.onAuthStateChange(async (event, newSession) => {
    console.debug('Auth state changed:', event);
    
    // Update session store
    session.set(newSession);
    
    // Invalidate all data to refetch with new auth state
    if (!sawFirstAuthEvent) {
      // Skip the very first auth event to avoid double-loading on initial page mount
      sawFirstAuthEvent = true;
      return;
    }

    if (event === 'SIGNED_IN' || event === 'SIGNED_OUT') {
      await invalidateAll();
    }
  });

  // Keep latestSession in sync for consumers that need a synchronous read
  // No-op; subscription retained if needed later for side-effects

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
 * Sign in with GitHub OAuth
 */
export async function signInWithGitHub(redirectToPath = '/'): Promise<{ error: AuthError | null }> {
  const supabase = getBrowserClient();
  if (!supabase) {
    return { error: { message: 'Supabase not configured', name: 'configuration_error' } as AuthError };
  }

  if (!browser) {
    return { error: { message: 'OAuth only available in browser', name: 'browser_required' } as AuthError };
  }

  // Build callback URL with redirect parameter
  const callbackUrl = `${window.location.origin}/auth/callback?redirect=${encodeURIComponent(redirectToPath)}`;
  
  const { error } = await supabase.auth.signInWithOAuth({
    provider: 'github',
    options: {
      redirectTo: callbackUrl
    }
  });

  return { error };
}

/**
 * Sign in with Google OAuth
 */
export async function signInWithGoogle(redirectToPath = '/'): Promise<{ error: AuthError | null }> {
  const supabase = getBrowserClient();
  if (!supabase) {
    return { error: { message: 'Supabase not configured', name: 'configuration_error' } as AuthError };
  }

  if (!browser) {
    return { error: { message: 'OAuth only available in browser', name: 'browser_required' } as AuthError };
  }

  // Build callback URL with redirect parameter
  const callbackUrl = `${window.location.origin}/auth/callback?redirect=${encodeURIComponent(redirectToPath)}`;

  const { error } = await supabase.auth.signInWithOAuth({
    provider: 'google',
    options: {
      redirectTo: callbackUrl,
      // Request offline access and consent to obtain provider_refresh_token when needed
      queryParams: {
        access_type: 'offline',
        prompt: 'consent'
      }
    }
  });

  return { error };
}

/**
 * Get the current access token for API requests
 */
export async function getAccessToken(): Promise<string | null> {
  // Always consult Supabase for the freshest session; coalesce concurrent calls
  const supabase = getBrowserClient();
  if (!supabase) return null;
  if (!sessionPromise) {
    sessionPromise = supabase.auth.getSession().then(({ data: { session } }) => session);
  }
  const currentSession = await sessionPromise;
  sessionPromise = null;
  return currentSession?.access_token ?? null;
}
