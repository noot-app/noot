// src/hooks.server.ts
import type { Handle } from "@sveltejs/kit"

// Simple pass-through handle for noot
// No authentication needed for the PoC
export const handle: Handle = async ({ event, resolve }) => {
  return resolve(event)
}
