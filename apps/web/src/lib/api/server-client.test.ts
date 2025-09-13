import { describe, it, expect, vi } from "vitest"

// Mock modules before importing
vi.mock("openapi-fetch")
vi.mock("$env/dynamic/public", () => ({
  env: {
    PUBLIC_API_BASE_URL: "https://localhost/api/v1"
  }
}))

// Import after mocking
import createClient from "openapi-fetch"
import { createServerApiClient, createServerApiClientFromSession } from "./server-client"

const mockCreateClient = vi.mocked(createClient)

describe("server-client", () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe("createServerApiClient", () => {
    it("should call createClient with correct base URL and no headers when no token", () => {
      createServerApiClient()

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {}
      })
    })

    it("should call createClient with Authorization header when token provided", () => {
      const token = "test-jwt-token"
      
      createServerApiClient(token)

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
    })

    it("should handle empty token by not adding Authorization header", () => {
      createServerApiClient("")

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {}
      })
    })

    it("should handle undefined token by not adding Authorization header", () => {
      createServerApiClient(undefined)

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {}
      })
    })
  })

  describe("createServerApiClientFromSession", () => {
    it("should call createServerApiClient with access_token from session", () => {
      const session = { access_token: "session-token-123" }
      
      createServerApiClientFromSession(session)

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {
          'Authorization': 'Bearer session-token-123'
        }
      })
    })

    it("should handle null session by creating client without auth", () => {
      createServerApiClientFromSession(null)

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {}
      })
    })

    it("should handle session without access_token", () => {
      const session = {} as any
      
      createServerApiClientFromSession(session)

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {}
      })
    })

    it("should handle session with empty access_token", () => {
      const session = { access_token: "" }
      
      createServerApiClientFromSession(session)

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {}
      })
    })
  })

  describe("Security and Edge Cases", () => {
    it("should handle malformed tokens safely", () => {
      const maliciousTokens = [
        "token\n\rX-Evil: header",
        'token"; DROP TABLE users; --',
        "token<script>alert('xss')</script>",
        "token${process.env.SECRET}",
        "a".repeat(10000) // Very long token
      ]

      maliciousTokens.forEach(token => {
        expect(() => {
          createServerApiClient(token)
        }).not.toThrow()

        expect(mockCreateClient).toHaveBeenCalledWith({
          baseUrl: "https://localhost/api/v1",
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })
      })
    })

    it("should handle malformed session objects gracefully", () => {
      const malformedSessions = [
        { access_token: null },
        { access_token: 123 as any },
        { access_token: {} as any },
        "not-an-object" as any,
        42 as any
      ]

      malformedSessions.forEach(session => {
        expect(() => {
          createServerApiClientFromSession(session)
        }).not.toThrow()
      })
    })

    it("should handle session with additional properties correctly", () => {
      const session = {
        access_token: "token-123",
        refresh_token: "refresh-456",
        user: { id: "user-789" },
        expires_at: 1234567890
      }
      
      createServerApiClientFromSession(session)

      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: "https://localhost/api/v1",
        headers: {
          'Authorization': 'Bearer token-123'
        }
      })
    })
  })

  describe("Performance", () => {
    it("should handle multiple rapid client creations", () => {
      const startTime = Date.now()
      
      for (let i = 0; i < 100; i++) {
        createServerApiClient(`token-${i}`)
      }
      
      const duration = Date.now() - startTime
      
      expect(duration).toBeLessThan(100) // Should be very fast
      expect(mockCreateClient).toHaveBeenCalledTimes(100)
    })

    it("should handle concurrent session-based client creation", () => {
      const sessions = Array(50).fill(0).map((_, i) => ({
        access_token: `token-${i}`
      }))

      expect(() => {
        sessions.forEach(session => {
          createServerApiClientFromSession(session)
        })
      }).not.toThrow()

      expect(mockCreateClient).toHaveBeenCalledTimes(50)
    })
  })
})
