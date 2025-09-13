import { describe, it, expect, vi, beforeEach } from 'vitest'
import { load } from './+page.server.js'

// Mock the server client
const mockServerApiClient = {
  GET: vi.fn()
}

vi.mock('$lib/api/server-client', () => ({
  createServerApiClientFromSession: vi.fn(() => mockServerApiClient)
}))

describe('Consumption Page Server', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('load function', () => {
    const mockLocals = {
      getSession: vi.fn(),
      supabase: {} as any // Mock supabase property
    }

    it('should format string errors properly', async () => {
      mockLocals.getSession.mockResolvedValue(null)
      mockServerApiClient.GET.mockResolvedValue({
        error: 'Not found',
        data: null
      })

      const result = await load({
        params: { id: 'test-id' },
        locals: mockLocals
      } as any)

      expect(result.error).toBe('Not found')
      expect(result.isUnauthenticated).toBe(true)
    })

    it('should format object errors with message property', async () => {
      mockLocals.getSession.mockResolvedValue(null)
      mockServerApiClient.GET.mockResolvedValue({
        error: { message: 'Access denied' },
        data: null
      })

      const result = await load({
        params: { id: 'test-id' },
        locals: mockLocals
      } as any)

      expect(result.error).toBe('Access denied')
    })

    it('should format object errors with error property', async () => {
      mockLocals.getSession.mockResolvedValue(null)
      mockServerApiClient.GET.mockResolvedValue({
        error: { error: 'Forbidden' },
        data: null
      })

      const result = await load({
        params: { id: 'test-id' },
        locals: mockLocals
      } as any)

      expect(result.error).toBe('Forbidden')
    })

    it('should use fallback message for unknown error types', async () => {
      mockLocals.getSession.mockResolvedValue(null)
      mockServerApiClient.GET.mockResolvedValue({
        error: { someOtherProperty: 'value' },
        data: null
      })

      const result = await load({
        params: { id: 'test-id' },
        locals: mockLocals
      } as any)

      expect(result.error).toBe("This consumption doesn't exist or you don't have permission to view it.")
    })

    it('should handle null/undefined errors with fallback', async () => {
      mockLocals.getSession.mockResolvedValue(null)
      mockServerApiClient.GET.mockResolvedValue({
        error: null,
        data: null
      })

      const result = await load({
        params: { id: 'test-id' },
        locals: mockLocals
      } as any)

      expect(result.error).toBe("This consumption doesn't exist or you don't have permission to view it.")
    })

    it('should successfully load consumption when no error', async () => {
      const mockSession = { access_token: 'token' }
      const mockConsumption = { id: 'test-id', title: 'Test Consumption' }
      
      mockLocals.getSession.mockResolvedValue(mockSession)
      mockServerApiClient.GET.mockResolvedValue({
        error: null,
        data: mockConsumption
      })

      const result = await load({
        params: { id: 'test-id' },
        locals: mockLocals
      } as any)

      expect(result.consumption).toEqual(mockConsumption)
      expect(result.isUnauthenticated).toBe(false)
      expect(result.error).toBeUndefined()
    })
  })
})
