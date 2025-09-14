import { describe, it, expect } from 'vitest'
import { DEFAULT_REDIRECT_PATH, sanitizeRedirect, getRedirectParam } from './redirect'

describe('Redirect Utilities', () => {
  describe('sanitizeRedirect', () => {
    it('should return fallback for empty path', () => {
      expect(sanitizeRedirect('')).toBe(DEFAULT_REDIRECT_PATH)
      expect(sanitizeRedirect('', '/custom')).toBe('/custom')
    })

    it('should return path for valid internal paths', () => {
      expect(sanitizeRedirect('/dashboard')).toBe('/dashboard')
      expect(sanitizeRedirect('/record')).toBe('/record')
      expect(sanitizeRedirect('/users/123')).toBe('/users/123')
      expect(sanitizeRedirect('/api/v1/test')).toBe('/api/v1/test')
    })

    it('should return fallback for external URLs', () => {
      expect(sanitizeRedirect('http://example.com')).toBe(DEFAULT_REDIRECT_PATH)
      expect(sanitizeRedirect('https://malicious.com')).toBe(DEFAULT_REDIRECT_PATH)
      // Fixed: Protocol-relative URLs now properly blocked for security
      expect(sanitizeRedirect('//example.com')).toBe(DEFAULT_REDIRECT_PATH)
      expect(sanitizeRedirect('javascript:alert("xss")')).toBe(DEFAULT_REDIRECT_PATH)
    })

    it('should return fallback for relative paths without leading slash', () => {
      expect(sanitizeRedirect('dashboard')).toBe(DEFAULT_REDIRECT_PATH)
      expect(sanitizeRedirect('record/123')).toBe(DEFAULT_REDIRECT_PATH)
    })

    it('should use custom fallback when provided', () => {
      const customFallback = '/home'
      expect(sanitizeRedirect('', customFallback)).toBe(customFallback)
      expect(sanitizeRedirect('http://evil.com', customFallback)).toBe(customFallback)
    })

    it('should handle null and undefined', () => {
      expect(sanitizeRedirect(null as any)).toBe(DEFAULT_REDIRECT_PATH)
      expect(sanitizeRedirect(undefined as any)).toBe(DEFAULT_REDIRECT_PATH)
    })
  })

  describe('getRedirectParam', () => {
    it('should extract valid redirect param', () => {
      const url = new URL('http://localhost:3000/login?redirect=/dashboard')
      expect(getRedirectParam(url)).toBe('/dashboard')
    })

    it('should sanitize invalid redirect param', () => {
      const url = new URL('http://localhost:3000/login?redirect=http://evil.com')
      expect(getRedirectParam(url)).toBe(DEFAULT_REDIRECT_PATH)
    })

    it('should return fallback for missing redirect param', () => {
      const url = new URL('http://localhost:3000/login')
      expect(getRedirectParam(url)).toBe(DEFAULT_REDIRECT_PATH)
    })

    it('should use custom fallback', () => {
      const url = new URL('http://localhost:3000/login')
      expect(getRedirectParam(url, '/custom-home')).toBe('/custom-home')
    })

    it('should handle empty redirect param', () => {
      const url = new URL('http://localhost:3000/login?redirect=')
      expect(getRedirectParam(url)).toBe(DEFAULT_REDIRECT_PATH)
    })

    it('should handle complex internal paths', () => {
      const url = new URL('http://localhost:3000/login?redirect=/dashboard/profile%3Ftab%3Dsettings')
      expect(getRedirectParam(url)).toBe('/dashboard/profile?tab=settings')
    })

    it('should handle protocol-relative URLs (security fix)', () => {
      const url = new URL('http://localhost:3000/login?redirect=//evil.com')
      // Fixed: Protocol-relative URLs now properly blocked for security
      expect(getRedirectParam(url)).toBe(DEFAULT_REDIRECT_PATH)
    })

    it('should handle multiple query parameters', () => {
      const url = new URL('http://localhost:3000/login?foo=bar&redirect=/profile&baz=qux')
      expect(getRedirectParam(url)).toBe('/profile')
    })
  })

  describe('DEFAULT_REDIRECT_PATH', () => {
    it('should be a valid internal path', () => {
      expect(DEFAULT_REDIRECT_PATH).toBe('/record')
      expect(DEFAULT_REDIRECT_PATH.startsWith('/')).toBe(true)
    })
  })
})