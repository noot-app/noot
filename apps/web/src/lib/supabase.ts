import { createBrowserClient } from '@supabase/ssr';
import { env } from '$env/dynamic/public';

// Supabase client configuration
const supabaseUrl = env.PUBLIC_SUPABASE_URL;
const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY;

/**
 * Supabase client instance
 * Returns null if credentials are not configured
 */
export const supabase = (() => {
	if (!supabaseUrl || !supabaseAnonKey) {
		console.warn('supabase public variables not configured');
		return null;
	}

	// Use createBrowserClient for better SSR support and session synchronization
	return createBrowserClient(supabaseUrl, supabaseAnonKey, {
		auth: {
			// Optimize for longer session persistence
			autoRefreshToken: true,
			persistSession: true,
			detectSessionInUrl: true,
			// Use localStorage for better persistence (30 days by default)
			storage: typeof window !== 'undefined' ? window.localStorage : undefined,
			storageKey: 'noot-supabase-auth-token',
			debug: false
		}
	});
})();

/**
 * Check if Supabase is available and configured
 */
export const isSupabaseEnabled = (): boolean => {
	return supabase !== null;
};
