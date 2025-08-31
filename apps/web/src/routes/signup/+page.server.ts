import { redirect } from '@sveltejs/kit';

export const load = async ({ locals, url }) => {
  const session = await locals.getSession();
  if (session) {
    const target = url.searchParams.get('redirect') || '/';
    throw redirect(303, target);
  }
  return {};
};
