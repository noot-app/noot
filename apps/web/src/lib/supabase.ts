import { createClient } from '@supabase/supabase-js';
import { dev } from '$app/environment';
import { env } from '$env/dynamic/public';

// Supabase client configuration
// Only initialize if we have valid Supabase credentials and not in dev mode with dev auth enabled
const supabaseUrl = env.PUBLIC_SUPABASE_URL;
const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY;

/**
 * Supabase client instance
 * Returns null if in development mode or if credentials are not configured
 */
export const supabase = (() => {
	// In development mode, we might want to disable Supabase entirely
	// and use only the dev auth system
	const enableSupabaseInDev = env.PUBLIC_ENABLE_SUPABASE_IN_DEV === 'true';
	
	if (dev && !enableSupabaseInDev) {
		console.log('Supabase disabled in development mode');
		return null;
	}

	if (!supabaseUrl || !supabaseAnonKey || 
		supabaseUrl.includes('REPLACE_ME') || 
		supabaseAnonKey.includes('REPLACE_ME')) {
		console.warn('Supabase credentials not configured');
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