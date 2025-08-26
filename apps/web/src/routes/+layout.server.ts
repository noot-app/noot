import { dev } from '$app/environment';
import { env } from '$env/dynamic/public';

/** @type {import('./$types').LayoutServerLoad} */
export async function load() {
  // Server-side auth state management
  // This will be used for SSR and initial page loads
  
  return {
    // Pass environment info to client
    isDevMode: dev,
    supabaseEnabled: !!(env.PUBLIC_SUPABASE_URL && env.PUBLIC_SUPABASE_ANON_KEY && 
                       !env.PUBLIC_SUPABASE_URL.includes('REPLACE_ME') && 
                       !env.PUBLIC_SUPABASE_ANON_KEY.includes('REPLACE_ME'))
  };
}
