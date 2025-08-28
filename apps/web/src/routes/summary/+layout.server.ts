// Server-side load function for summary route
// User/session is already validated by hooks.server.ts
export async function load({ locals }) {
  return {
    user: locals.user,
    session: locals.session
  };
}