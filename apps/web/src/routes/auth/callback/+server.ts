import { redirect, type RequestHandler } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ url, locals }) => {
  const code = url.searchParams.get('code');
  const rawNext = url.searchParams.get('redirect') || '/';

  // Allow only internal redirects for safety
  const next = rawNext.startsWith('/') ? rawNext : '/';

  if (!code) {
    throw redirect(303, `/login?error=${encodeURIComponent('missing_oauth_code')}`);
  }

  const { error } = await locals.supabase.auth.exchangeCodeForSession(code);

  if (error) {
    throw redirect(303, `/login?error=${encodeURIComponent(error.message)}`);
  }

  throw redirect(303, next);
};