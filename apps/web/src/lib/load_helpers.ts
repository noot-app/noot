import { isBrowser } from "@supabase/ssr"
import type { Session, SupabaseClient } from "@supabase/supabase-js"
import type { Database } from "../DatabaseDefinitions.js"
import { authLogger } from "./utils/logger.js"

export const load_helper = async (
  server_session: Session | null,
  supabase: SupabaseClient<Database>,
) => {
  // On server, use the session populated by authGuard hook
  let session = server_session
  
  if (isBrowser()) {
    // On client-side, only verify authentication if we don't have a valid server session with user
    if (!server_session?.user) {
      authLogger.debug('No server session user, verifying client-side authentication');
      // Use getUser() for secure authentication instead of getSession()
      const { data: { user }, error } = await supabase.auth.getUser();
      if (error || !user) {
        authLogger.debug('Client-side authentication failed or no user');
        return {
          session: null,
          user: null,
        }
      }
      
      // Create a minimal session-like object with verified user (consistent with server)
      session = {
        access_token: '',
        refresh_token: '',
        expires_in: 0,
        expires_at: 0,
        token_type: 'bearer',
        user: user
      } as Session;
    } else {
      authLogger.debug('Using existing server session for client-side');
    }
  }
  
  if (!session?.user) {
    return {
      session: null,
      user: null,
    }
  }

  // Return the session and user data consistently
  return {
    session,
    user: session.user,
  }
}
