import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, locals }) => {
  const code = url.searchParams.get('code');
  const redirectParam = url.searchParams.get('redirect') || '/summary'; // default to /summary
  
  // Validate that code parameter exists
  if (!code) {
    throw redirect(303, '/login?error=missing_oauth_code');
  }
  
  // Sanitize redirect to only allow internal paths (security measure)
  const sanitizedRedirect = redirectParam.startsWith('/') ? redirectParam : '/';
  
  try {
    // Exchange the OAuth code for a session using Supabase SSR
    const { error } = await locals.supabase.auth.exchangeCodeForSession(code);
    
    if (error) {
      console.error('OAuth callback error:', error);
      throw redirect(303, '/login?error=oauth_exchange_failed');
    }
    
    // Successful OAuth - redirect to target location
    throw redirect(303, sanitizedRedirect);
  } catch (err) {
    console.error('OAuth callback exception:', err);
    throw redirect(303, '/login?error=oauth_callback_failed');
  }
};
