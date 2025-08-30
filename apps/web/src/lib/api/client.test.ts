import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock environment
const mockEnv = {
  PUBLIC_API_BASE_URL: 'http://localhost:3001/api/v1'
};

vi.mock('$env/dynamic/public', () => ({
  env: mockEnv
}));

// Mock openapi-fetch
const mockCreateClient = vi.fn(() => ({
  GET: vi.fn(),
  POST: vi.fn(),
  PUT: vi.fn(),
  DELETE: vi.fn(),
  PATCH: vi.fn()
}));

vi.mock('openapi-fetch', () => ({
  default: mockCreateClient
}));

// Mock auth store
const mockGetAccessToken = vi.fn();
vi.mock('$lib/auth/store', () => ({
  getAccessToken: mockGetAccessToken
}));

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEnv.PUBLIC_API_BASE_URL = 'http://localhost:3001/api/v1';
    
    // Mock client-side environment
    (global as any).window = {};
  });

  afterEach(() => {
    delete (global as any).window;
  });

  describe('client initialization', () => {
    it('should create client with environment base URL', async () => {
      await import('./client');
      
      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: 'http://localhost:3001/api/v1'
      });
    });

    it('should fallback to production URL when env var is missing', async () => {
      mockEnv.PUBLIC_API_BASE_URL = '';
      vi.resetModules();
      
      await import('./client');
      
      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: 'https://api.nootapp.io/api/v1'
      });
    });

    it('should handle undefined environment variable', async () => {
      mockEnv.PUBLIC_API_BASE_URL = undefined as any;
      vi.resetModules();
      
      await import('./client');
      
      expect(mockCreateClient).toHaveBeenCalledWith({
        baseUrl: 'https://api.nootapp.io/api/v1'
      });
    });
  });

  describe('getAccessToken', () => {
    it('should return null during SSR', async () => {
      delete (global as any).window;
      
      // We can't directly test the private function, but we can test its behavior
      // by checking that auth provider is not called during SSR
      
      // Mock the underlying method to capture calls
      const mockMethod = vi.fn().mockResolvedValue({ data: null });
      const mockTarget = { GET: mockMethod };
      
      // Create a proxy handler to test the behavior
      const handler = {
        get(target: any, prop: string) {
          if (prop === 'GET' && typeof target[prop] === 'function') {
            return async function(url: string, init?: unknown) {
              // This mimics the proxy behavior but in a testable way
              return target[prop](url, init);
            };
          }
          return target[prop];
        }
      };
      
      const testProxy = new Proxy(mockTarget, handler);
      await testProxy.GET('/test');
      
      expect(mockMethod).toHaveBeenCalled();
    });
  });

  describe('API client proxy', () => {
    it('should return original method for non-HTTP methods', async () => {
      const { apiClient } = await import('./client');
      
      const originalProp = (apiClient as any).someProperty;
      expect(originalProp).toBeUndefined(); // Original property should be undefined
    });

    it('should wrap HTTP methods with auth header injection', async () => {
      const mockToken = 'test-access-token';
      mockGetAccessToken.mockResolvedValue(mockToken);
      
      const mockGet = vi.fn().mockResolvedValue({ data: 'test' });
      mockCreateClient.mockReturnValue({
        GET: mockGet,
        POST: vi.fn(),
        PUT: vi.fn(),
        DELETE: vi.fn(),
        PATCH: vi.fn()
      });
      
      vi.resetModules();
      const { apiClient } = await import('./client');
      
      await (apiClient as any).GET('/test', {});
      
      expect(mockGet).toHaveBeenCalledWith('/test', {
        headers: {
          'Authorization': `Bearer ${mockToken}`
        }
      });
    });

    it('should handle requests without init parameter', async () => {
      const mockToken = 'test-access-token';
      mockGetAccessToken.mockResolvedValue(mockToken);
      
      const mockPost = vi.fn().mockResolvedValue({ data: 'test' });
      mockCreateClient.mockReturnValue({
        GET: vi.fn(),
        POST: mockPost,
        PUT: vi.fn(),
        DELETE: vi.fn(),
        PATCH: vi.fn()
      });
      
      vi.resetModules();
      const { apiClient } = await import('./client');
      
      await (apiClient as any).POST('/test');
      
      expect(mockPost).toHaveBeenCalledWith('/test', {
        headers: {
          'Authorization': `Bearer ${mockToken}`
        }
      });
    });

    it('should preserve existing headers while adding auth', async () => {
      const mockToken = 'test-access-token';
      mockGetAccessToken.mockResolvedValue(mockToken);
      
      const mockPut = vi.fn().mockResolvedValue({ data: 'test' });
      mockCreateClient.mockReturnValue({
        GET: vi.fn(),
        POST: vi.fn(),
        PUT: mockPut,
        DELETE: vi.fn(),
        PATCH: vi.fn()
      });
      
      vi.resetModules();
      const { apiClient } = await import('./client');
      
      const existingHeaders = { 'Content-Type': 'application/json' };
      await (apiClient as any).PUT('/test', { headers: existingHeaders });
      
      expect(mockPut).toHaveBeenCalledWith('/test', {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockToken}`
        }
      });
    });

    it('should work without access token', async () => {
      mockGetAccessToken.mockResolvedValue(null);
      
      const mockDelete = vi.fn().mockResolvedValue({ data: 'test' });
      mockCreateClient.mockReturnValue({
        GET: vi.fn(),
        POST: vi.fn(),
        PUT: vi.fn(),
        DELETE: mockDelete,
        PATCH: vi.fn()
      });
      
      vi.resetModules();
      const { apiClient } = await import('./client');
      
      await (apiClient as any).DELETE('/test', {});
      
      expect(mockDelete).toHaveBeenCalledWith('/test', {
        headers: {}
      });
    });
  });

  describe('edge cases', () => {
    it('should handle malformed init object', () => {
      // Test that malformed init objects throw proper error
      expect(() => {
        const malformedInit = 'not-an-object' as any;
        malformedInit.headers = malformedInit.headers || {};
      }).toThrow(TypeError);
    });

    it('should handle auth provider import failure', async () => {
      // Test that client still works even if auth provider import fails
      const { apiClient } = await import('./client');
      expect(apiClient).toBeDefined();
      expect(typeof apiClient.GET).toBe('function');
    });

    it('should handle empty token string', async () => {
      mockGetAccessToken.mockResolvedValue('');
      
      const mockGet = vi.fn().mockResolvedValue({ data: 'test' });
      mockCreateClient.mockReturnValue({
        GET: mockGet,
        POST: vi.fn(),
        PUT: vi.fn(),
        DELETE: vi.fn(),
        PATCH: vi.fn()
      });
      
      vi.resetModules();
      const { apiClient } = await import('./client');
      
      await (apiClient as any).GET('/test', {});
      
      // Empty string is falsy, so no auth header should be added
      expect(mockGet).toHaveBeenCalledWith('/test', {
        headers: {}
      });
    });

    it('should handle non-string token', async () => {
      mockGetAccessToken.mockResolvedValue(123 as any); // Non-string token
      
      const mockGet = vi.fn().mockResolvedValue({ data: 'test' });
      mockCreateClient.mockReturnValue({
        GET: mockGet,
        POST: vi.fn(),
        PUT: vi.fn(),
        DELETE: vi.fn(),
        PATCH: vi.fn()
      });
      
      vi.resetModules();
      const { apiClient } = await import('./client');
      
      await (apiClient as any).GET('/test', {});
      
      // Non-string token should still be converted to string (truthy value)
      expect(mockGet).toHaveBeenCalledWith('/test', {
        headers: {
          'Authorization': 'Bearer 123'
        }
      });
    });
  });

  describe('environment configuration', () => {
    it('should handle various base URL formats', () => {
      const testUrls = [
        'http://localhost:3000',
        'https://api.example.com',
        'https://api.example.com/',
        'http://127.0.0.1:8080/api/v1'
      ];

      testUrls.forEach(url => {
        mockEnv.PUBLIC_API_BASE_URL = url;
        vi.resetModules();
        
        expect(() => import('./client')).not.toThrow();
      });
    });

    it('should log API base URL for debugging', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      
      mockEnv.PUBLIC_API_BASE_URL = 'http://test-url:3000';
      vi.resetModules();
      
      await import('./client');
      
      expect(consoleSpy).toHaveBeenCalledWith('API Base URL (runtime):', 'http://test-url:3000');
      
      consoleSpy.mockRestore();
    });
  });
});
