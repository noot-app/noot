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
    
    // Create a minimal session-like object with verified user (no getSession() call)
    session = {
      access_token: '',
      refresh_token: '',
      expires_in: 0,
      expires_at: 0,
      token_type: 'bearer',
      user: user
    } as Session;
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
