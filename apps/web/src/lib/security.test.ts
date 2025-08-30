import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { sanitizeReturnUrl, isSafeRedirectUrl, shouldAttachAuthHeader } from './security';

describe('Security utilities', () => {
  describe('sanitizeReturnUrl', () => {
    // Mock window.location for tests
    const mockLocation = {
      origin: 'https://example.com'
    };

    beforeEach(() => {
      // @ts-expect-error - mocking window for tests
      global.window = {
        location: mockLocation
      };
    });

    afterEach(() => {
      // @ts-expect-error - cleaning up mock
      delete global.window;
    });

    it('should return default URL for null input', () => {
      expect(sanitizeReturnUrl(null)).toBe('/');
      expect(sanitizeReturnUrl(null, '/home')).toBe('/home');
    });

    it('should return default URL for empty string', () => {
      expect(sanitizeReturnUrl('')).toBe('/');
      expect(sanitizeReturnUrl('   ')).toBe('/');
    });

    it('should allow valid relative paths', () => {
      expect(sanitizeReturnUrl('/dashboard')).toBe('/dashboard');
      expect(sanitizeReturnUrl('/profile/settings')).toBe('/profile/settings');
      expect(sanitizeReturnUrl('/dashboard?tab=overview')).toBe('/dashboard?tab=overview');
      expect(sanitizeReturnUrl('/profile#section')).toBe('/profile#section');
      expect(sanitizeReturnUrl('/search?q=test&page=2#results')).toBe('/search?q=test&page=2#results');
    });

    it('should reject paths that do not start with /', () => {
      expect(sanitizeReturnUrl('dashboard')).toBe('/');
      expect(sanitizeReturnUrl('profile/settings')).toBe('/');
      expect(sanitizeReturnUrl('../admin')).toBe('/');
    });

    it('should reject path traversal attempts', () => {
      expect(sanitizeReturnUrl('/../../etc/passwd')).toBe('/');
      expect(sanitizeReturnUrl('/../admin')).toBe('/');
      expect(sanitizeReturnUrl('/profile/../admin')).toBe('/');
      expect(sanitizeReturnUrl('/safe/path/../../../danger')).toBe('/');
      expect(sanitizeReturnUrl('/path\\..\\admin')).toBe('/');
    });

    it('should reject dangerous URL schemes', () => {
      expect(sanitizeReturnUrl('javascript:alert(1)')).toBe('/');
      expect(sanitizeReturnUrl('JavaScript:alert(1)')).toBe('/');
      expect(sanitizeReturnUrl('data:text/html,<script>alert(1)</script>')).toBe('/');
      expect(sanitizeReturnUrl('vbscript:msgbox(1)')).toBe('/');
      expect(sanitizeReturnUrl('file:///etc/passwd')).toBe('/');
      expect(sanitizeReturnUrl('ftp://malicious.com/file')).toBe('/');
    });

    it('should handle same-origin absolute URLs', () => {
      expect(sanitizeReturnUrl('https://example.com/dashboard')).toBe('/dashboard');
      expect(sanitizeReturnUrl('https://example.com/profile?tab=settings')).toBe('/profile?tab=settings');
      expect(sanitizeReturnUrl('https://example.com/search#results')).toBe('/search#results');
    });

    it('should reject external absolute URLs', () => {
      expect(sanitizeReturnUrl('https://evil.com/dashboard')).toBe('/');
      expect(sanitizeReturnUrl('http://malicious.example.com/profile')).toBe('/');
      expect(sanitizeReturnUrl('https://subdomain.evil.com/admin')).toBe('/');
      expect(sanitizeReturnUrl('https://example.com.evil.com/fake')).toBe('/');
    });

    it('should handle malformed URLs gracefully', () => {
      expect(sanitizeReturnUrl('https://')).toBe('/');
      expect(sanitizeReturnUrl('http://')).toBe('/');
      expect(sanitizeReturnUrl('://')).toBe('/');
      expect(sanitizeReturnUrl('not-a-url')).toBe('/');
      expect(sanitizeReturnUrl('https://[invalid')).toBe('/');
    });

    it('should work in server-side rendering context', () => {
      // @ts-expect-error - removing window to simulate SSR
      delete global.window;

      // In SSR, absolute URLs should be rejected even if same-origin
      expect(sanitizeReturnUrl('https://example.com/dashboard')).toBe('/');
      expect(sanitizeReturnUrl('/dashboard')).toBe('/dashboard');
    });

    it('should handle edge cases', () => {
      expect(sanitizeReturnUrl('/')).toBe('/');
      expect(sanitizeReturnUrl('//')).toBe('/'); // Double slash gets normalized to single slash
      expect(sanitizeReturnUrl('/?')).toBe('/'); // Empty query parameter gets removed
      expect(sanitizeReturnUrl('/#')).toBe('/'); // Empty fragment gets removed
    });

    it('should preserve query parameters and fragments', () => {
      expect(sanitizeReturnUrl('/dashboard?tab=overview&filter=active'))
        .toBe('/dashboard?tab=overview&filter=active');
      expect(sanitizeReturnUrl('/profile#settings'))
        .toBe('/profile#settings');
      expect(sanitizeReturnUrl('/search?q=test%20query&page=2#results'))
        .toBe('/search?q=test%20query&page=2#results');
    });

    it('should handle special characters in paths', () => {
      expect(sanitizeReturnUrl('/caf%C3%A9')).toBe('/caf%C3%A9');
      expect(sanitizeReturnUrl('/search?q=hello%20world')).toBe('/search?q=hello%20world');
      expect(sanitizeReturnUrl('/user/@username')).toBe('/user/@username');
    });
  });

  describe('isSafeRedirectUrl', () => {
    beforeEach(() => {
      // @ts-expect-error - mocking window for tests
      global.window = {
        location: { origin: 'https://example.com' }
      };
    });

    afterEach(() => {
      // @ts-expect-error - cleaning up mock
      delete global.window;
    });

    it('should return true for safe URLs', () => {
      expect(isSafeRedirectUrl('/dashboard')).toBe(true);
      expect(isSafeRedirectUrl('/profile/settings')).toBe(true);
      expect(isSafeRedirectUrl('/')).toBe(true);
    });

    it('should return false for dangerous URLs', () => {
      expect(isSafeRedirectUrl('https://evil.com/malicious')).toBe(false);
      expect(isSafeRedirectUrl('javascript:alert(1)')).toBe(false);
      expect(isSafeRedirectUrl('/../admin')).toBe(false);
    });

    it('should handle empty string as safe', () => {
      expect(isSafeRedirectUrl('')).toBe(true);
    });
  });

  describe('shouldAttachAuthHeader', () => {
    it('should return true for relative URLs', () => {
      expect(shouldAttachAuthHeader('/api/users', 'https://api.example.com')).toBe(true);
      expect(shouldAttachAuthHeader('/api/posts', 'https://api.example.com')).toBe(true);
    });

    it('should return true for same-origin API requests', () => {
      const apiBase = 'https://api.example.com/v1';
      expect(shouldAttachAuthHeader('https://api.example.com/v1/users', apiBase)).toBe(true);
      expect(shouldAttachAuthHeader('https://api.example.com/v2/posts', apiBase)).toBe(true);
      expect(shouldAttachAuthHeader('https://api.example.com/', apiBase)).toBe(true);
    });

    it('should return false for external URLs', () => {
      const apiBase = 'https://api.example.com/v1';
      expect(shouldAttachAuthHeader('https://evil.com/api/steal', apiBase)).toBe(false);
      expect(shouldAttachAuthHeader('https://malicious.example.com/api', apiBase)).toBe(false);
      expect(shouldAttachAuthHeader('https://api.example.com.evil.com/fake', apiBase)).toBe(false);
    });

    it('should return false for empty or null inputs', () => {
      expect(shouldAttachAuthHeader('', 'https://api.example.com')).toBe(false);
      expect(shouldAttachAuthHeader('/api/users', '')).toBe(false);
      expect(shouldAttachAuthHeader('', '')).toBe(false);
    });

    it('should handle malformed URLs gracefully', () => {
      const apiBase = 'https://api.example.com/v1';
      expect(shouldAttachAuthHeader('not-a-url', apiBase)).toBe(false);
      expect(shouldAttachAuthHeader('https://', apiBase)).toBe(false);
      expect(shouldAttachAuthHeader('/api/users', 'not-a-url')).toBe(true); // Relative URL should still work
    });

    it('should handle different protocols', () => {
      const httpApiBase = 'http://localhost:3000/api';
      const httpsApiBase = 'https://api.example.com/v1';
      
      expect(shouldAttachAuthHeader('http://localhost:3000/api/users', httpApiBase)).toBe(true);
      expect(shouldAttachAuthHeader('https://api.example.com/v1/posts', httpsApiBase)).toBe(true);
      
      // Different protocols should not match
      expect(shouldAttachAuthHeader('http://localhost:3000/api/users', httpsApiBase)).toBe(false);
      expect(shouldAttachAuthHeader('https://api.example.com/v1/posts', httpApiBase)).toBe(false);
    });

    it('should handle ports correctly', () => {
      const apiBase = 'http://localhost:3000/api';
      expect(shouldAttachAuthHeader('http://localhost:3000/api/users', apiBase)).toBe(true);
      expect(shouldAttachAuthHeader('http://localhost:3001/api/users', apiBase)).toBe(false);
      expect(shouldAttachAuthHeader('http://localhost/api/users', apiBase)).toBe(false);
    });
  });
});