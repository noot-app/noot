import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { writable } from "svelte/store"
import { initAuth } from "./store"

// Mock the dependencies
vi.mock("$app/environment", () => ({
  browser: true,
}))

vi.mock("$app/navigation", () => ({
  invalidateAll: vi.fn(),
}))

vi.mock("$app/stores", () => ({
  navigating: writable(false),
}))

vi.mock("$env/dynamic/public", () => ({
  env: {
    PUBLIC_SUPABASE_URL: "https://test.supabase.co",
    PUBLIC_SUPABASE_ANON_KEY: "test-key",
  },
}))

// Mock Supabase client
const mockSupabaseClient = {
  auth: {
    onAuthStateChange: vi.fn(),
  },
}

vi.mock("@supabase/ssr", () => ({
  createBrowserClient: vi.fn(() => mockSupabaseClient),
}))

describe("Auth Store Navigation Handling", () => {
  let authStateChangeCallback: (event: string, session: any) => void
  let unsubscribeFn: () => void

  beforeEach(() => {
    vi.clearAllMocks()
    
    // Mock the subscription
    unsubscribeFn = vi.fn()
    mockSupabaseClient.auth.onAuthStateChange.mockReturnValue({
      data: { subscription: { unsubscribe: unsubscribeFn } },
    })

    // Capture the auth state change callback
    mockSupabaseClient.auth.onAuthStateChange.mockImplementation((callback) => {
      authStateChangeCallback = callback
      return {
        data: { subscription: { unsubscribe: unsubscribeFn } },
      }
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("should skip invalidation during active navigation", async () => {
    const { invalidateAll } = await import("$app/navigation")
    const { navigating } = await import("$app/stores")
    
    // Initialize auth
    initAuth(null)
    
    // Simulate first auth event (should be skipped)
    await authStateChangeCallback("INITIAL_SESSION", { user: { id: "123" } })
    expect(invalidateAll).not.toHaveBeenCalled()
    
    // Set navigation state to active
    navigating.set({
      from: { url: new URL("http://localhost/record") },
      to: { url: new URL("http://localhost/profile") },
      type: "link",
    })
    
    // Trigger a SIGNED_IN event during navigation
    await authStateChangeCallback("SIGNED_IN", { user: { id: "123" } })
    
    // Should not call invalidateAll during navigation
    expect(invalidateAll).not.toHaveBeenCalled()
  })

  it("should call invalidation when not navigating", async () => {
    const { invalidateAll } = await import("$app/navigation")
    const { navigating } = await import("$app/stores")
    
    // Initialize auth
    initAuth(null)
    
    // Simulate first auth event (should be skipped)
    await authStateChangeCallback("INITIAL_SESSION", { user: { id: "123" } })
    expect(invalidateAll).not.toHaveBeenCalled()
    
    // Ensure navigation state is cleared
    navigating.set(null)
    
    // Trigger a SIGNED_IN event when not navigating
    await authStateChangeCallback("SIGNED_IN", { user: { id: "123" } })
    
    // Should call invalidateAll when not navigating
    expect(invalidateAll).toHaveBeenCalledTimes(1)
  })

  it("should handle invalidation errors gracefully", async () => {
    const { invalidateAll } = await import("$app/navigation")
    const { navigating } = await import("$app/stores")
    
    // Make invalidateAll throw an error
    vi.mocked(invalidateAll).mockRejectedValue(new Error("Network error"))
    
    // Spy on console.error
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {})
    
    // Initialize auth
    initAuth(null)
    
    // Simulate first auth event (should be skipped)
    await authStateChangeCallback("INITIAL_SESSION", { user: { id: "123" } })
    
    // Ensure navigation state is cleared
    navigating.set(null)
    
    // Trigger a SIGNED_IN event
    await authStateChangeCallback("SIGNED_IN", { user: { id: "123" } })
    
    // Should have caught the error and logged it
    expect(consoleSpy).toHaveBeenCalledWith("Error during data invalidation:", expect.any(Error))
    
    consoleSpy.mockRestore()
  })

  it("should prevent concurrent invalidations", async () => {
    const { invalidateAll } = await import("$app/navigation")
    const { navigating } = await import("$app/stores")
    
    // Make invalidateAll take some time
    let resolveInvalidate: () => void
    const invalidatePromise = new Promise<void>((resolve) => {
      resolveInvalidate = resolve
    })
    vi.mocked(invalidateAll).mockReturnValue(invalidatePromise)
    
    // Initialize auth
    initAuth(null)
    
    // Simulate first auth event (should be skipped)
    await authStateChangeCallback("INITIAL_SESSION", { user: { id: "123" } })
    
    // Ensure navigation state is cleared
    navigating.set(null)
    
    // Trigger first SIGNED_IN event
    const promise1 = authStateChangeCallback("SIGNED_IN", { user: { id: "123" } })
    
    // Trigger second SIGNED_IN event before first completes
    const promise2 = authStateChangeCallback("SIGNED_IN", { user: { id: "123" } })
    
    // Complete the first invalidation
    resolveInvalidate!()
    await promise1
    await promise2
    
    // Should only call invalidateAll once (second call should be skipped)
    expect(invalidateAll).toHaveBeenCalledTimes(1)
  })
})