import { createClient } from '@supabase/supabase-js';
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

	return createClient(supabaseUrl, supabaseAnonKey, {
		auth: {
			autoRefreshToken: true,
			persistSession: true,
			detectSessionInUrl: true
		}
	});
})();

/**
 * Check if Supabase is available and configured
 */
export const isSupabaseEnabled = (): boolean => {
	return supabase !== null;
};
