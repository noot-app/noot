import { dev } from '$app/environment';
import { env } from '$env/dynamic/public';

/** @type {import('./$types').LayoutServerLoad} */
export async function load({ locals }) {
  // Server-side auth state management - pass session to client to prevent hydration mismatch
  
  return {
    // Pass environment info to client
    isDevMode: dev,
    supabaseEnabled: !!(env.PUBLIC_SUPABASE_URL && env.PUBLIC_SUPABASE_ANON_KEY && 
                       !env.PUBLIC_SUPABASE_URL.includes('REPLACE_ME') && 
                       !env.PUBLIC_SUPABASE_ANON_KEY.includes('REPLACE_ME')),
    // Pass session to client to prevent auth hydration mismatch
    session: locals.session
  };
}
