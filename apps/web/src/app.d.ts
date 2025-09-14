import { Session, SupabaseClient } from "@supabase/supabase-js"
import { Database } from "./DatabaseDefinitions"

// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
  namespace App {
    interface Locals {
      supabase: SupabaseClient<Database>
      getSession: () => Promise<Session | null>
      cspNonce: string
    }
    interface PageData {
      session: Session | null
      cspNonce?: string
      isDevMode?: boolean
    }
    // interface Error {}
    // interface Platform {}
  }
}

export {}
