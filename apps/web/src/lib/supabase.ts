import { createBrowserClient } from '@supabase/ssr';
import { env } from '$env/dynamic/public';

// Supabase client configuration
const supabaseUrl = env.PUBLIC_SUPABASE_URL;
const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY;

/**
 * Supabase client instance for browser (SSR-compatible)
 * Returns null if credentials are not configured
 * Uses @supabase/ssr for proper cookie handling between client/server
 */
export const supabase = (() => {
	if (!supabaseUrl || !supabaseAnonKey) {
		console.warn('supabase public variables not configured');
		return null;
	}

	// Only create browser client on client-side
	if (typeof window !== 'undefined') {
		return createBrowserClient(supabaseUrl, supabaseAnonKey);
	}
	
	return null;
})();

/**
 * Check if Supabase is available and configured
 */
export const isSupabaseEnabled = (): boolean => {
	return supabase !== null;
};
