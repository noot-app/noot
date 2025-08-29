import { isBrowser } from "@supabase/ssr"
import type { Session, SupabaseClient } from "@supabase/supabase-js"
import type { Database } from "../DatabaseDefinitions.js"

export const load_helper = async (
  server_session: Session | null,
  supabase: SupabaseClient<Database>,
) => {
  // on server populated on server by LayoutData, using authGuard hook
  let session = server_session
  if (isBrowser()) {
    // Use getUser() for secure authentication instead of getSession()
    const { data: { user }, error } = await supabase.auth.getUser();
    if (error || !user) {
      return {
        session: null,
        user: null,
      }
    }
    
    // Only get session if user is verified (for token access if needed)
    const getSessionResponse = await supabase.auth.getSession()
    session = getSessionResponse.data.session
  }
  if (!session) {
    return {
      session: null,
      user: null,
    }
  }

  // We already verified the user above with getUser(), so no need to do it again
  // Just return the session and user data
  return {
    session,
    user: session.user,
  }
}
