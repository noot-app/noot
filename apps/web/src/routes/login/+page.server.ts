import { redirect } from "@sveltejs/kit"
import { DEFAULT_REDIRECT_PATH, getRedirectParam } from "$lib/utils/redirect"

export const load = async ({ locals, url }) => {
  const session = await locals.getSession()
  if (session) {
    const target = getRedirectParam(url, DEFAULT_REDIRECT_PATH)
    throw redirect(303, target)
  }
  return {}
}
