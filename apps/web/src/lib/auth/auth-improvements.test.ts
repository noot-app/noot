import { describe, it, expect, vi, beforeEach } from 'vitest'
import { initAuth } from './store'

// Mock Supabase
vi.mock('$lib/supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn().mockResolvedValue({ data: { session: null }, error: null }),
      onAuthStateChange: vi.fn().mockReturnValue({ 
        data: { subscription: { unsubscribe: vi.fn() } } 
      })
    }
  },
  isSupabaseEnabled: vi.fn().mockReturnValue(true)
}))

// Mock environment
vi.mock('$env/dynamic/public', () => ({
  env: {
    PUBLIC_SUPABASE_URL: 'https://test.supabase.co',
    PUBLIC_SUPABASE_ANON_KEY: 'test-key'
  }
}))

describe('Authentication Improvements', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should initialize auth without server session', async () => {
    // This test validates that initAuth can be called without throwing errors
    await expect(initAuth()).resolves.not.toThrow()
  })

  it('should initialize auth with server session', async () => {
    const mockSession = {
      user: {
        id: 'test-user-id',
        email: 'test@example.com'
      },
      access_token: 'test-token'
    }

    // This test validates that initAuth can accept a server session
    await expect(initAuth(mockSession)).resolves.not.toThrow()
  })

  it('should handle invalid server session gracefully', async () => {
    const invalidSession = { invalid: 'data' }

    // This test validates that initAuth gracefully handles invalid session data
    await expect(initAuth(invalidSession)).resolves.not.toThrow()
  })
})