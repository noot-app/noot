// Server-side load function for record route
// Session is already validated by hooks.server.ts
export async function load({ locals }) {
  const session = await locals.getSession()
  return {
    session,
  }
}
