import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock @supabase/ssr
vi.mock('@supabase/ssr', () => ({
  isBrowser: vi.fn()
}));

import { load_helper } from './load_helpers';
import { isBrowser } from '@supabase/ssr';

const mockIsBrowser = vi.mocked(isBrowser);

describe('Load Helper', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('server-side behavior', () => {
    beforeEach(() => {
      mockIsBrowser.mockReturnValue(false);
    });

    it('should return server session when provided', async () => {
      const mockSession = {
        access_token: 'server-token',
        refresh_token: 'refresh',
        expires_in: 3600,
        expires_at: Date.now() / 1000 + 3600,
        token_type: 'bearer',
        user: {
          id: 'user-id',
          email: 'test@example.com',
          created_at: '2023-01-01'
        }
      } as any;

      const mockSupabase = {} as any; // Not used on server side

      const result = await load_helper(mockSession, mockSupabase);

      expect(result.session).toEqual(mockSession);
      expect(result.user).toEqual(mockSession.user);
    });

    it('should return null when no server session provided', async () => {
      const mockSupabase = {} as any;

      const result = await load_helper(null, mockSupabase);

      expect(result.session).toBe(null);
      expect(result.user).toBe(null);
    });

    it('should handle undefined server session', async () => {
      const mockSupabase = {} as any;

      const result = await load_helper(undefined as any, mockSupabase);

      expect(result.session).toBe(null);
      expect(result.user).toBe(null);
    });
  });

  describe('client-side behavior', () => {
    beforeEach(() => {
      mockIsBrowser.mockReturnValue(true);
    });

    it('should verify user with getUser() and create session object', async () => {
      const mockUser = {
        id: 'user-id',
        email: 'test@example.com',
        created_at: '2023-01-01'
      };

      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: mockUser },
            error: null
          })
        }
      } as any;

      const result = await load_helper(null, mockSupabase);

      expect(mockSupabase.auth.getUser).toHaveBeenCalled();
      expect(result.session).toEqual({
        access_token: '',
        refresh_token: '',
        expires_in: 0,
        expires_at: 0,
        token_type: 'bearer',
        user: mockUser
      });
      expect(result.user).toEqual(mockUser);
    });

    it('should return null when getUser() returns error', async () => {
      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: null },
            error: { message: 'No user found' }
          })
        }
      } as any;

      const result = await load_helper(null, mockSupabase);

      expect(mockSupabase.auth.getUser).toHaveBeenCalled();
      expect(result.session).toBe(null);
      expect(result.user).toBe(null);
    });

    it('should return null when getUser() returns no user', async () => {
      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: null },
            error: null
          })
        }
      } as any;

      const result = await load_helper(null, mockSupabase);

      expect(mockSupabase.auth.getUser).toHaveBeenCalled();
      expect(result.session).toBe(null);
      expect(result.user).toBe(null);
    });

    it('should use server session on client side when available', async () => {
      const mockServerSession = {
        access_token: 'old-token',
        user: { id: 'old-user' }
      } as any;

      const mockClientUser = {
        id: 'new-user',
        email: 'new@example.com'
      };

      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: mockClientUser },
            error: null
          })
        }
      } as any;

      const result = await load_helper(mockServerSession, mockSupabase);

      // Should NOT call getUser if we have a valid server session
      expect(mockSupabase.auth.getUser).not.toHaveBeenCalled();
      expect(result.user).toEqual({ id: 'old-user' });
      expect(result.session?.user).toEqual({ id: 'old-user' });
    });

    it('should handle getUser() throwing an exception', async () => {
      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockRejectedValue(new Error('Network error'))
        }
      } as any;

      // The function doesn't have explicit error handling, so it should throw
      await expect(load_helper(null, mockSupabase)).rejects.toThrow('Network error');
    });
  });

  describe('edge cases', () => {
    it('should handle invalid supabase client', async () => {
      mockIsBrowser.mockReturnValue(true);
      
      const invalidSupabase = null as any;

      // This should throw when trying to access auth property
      await expect(load_helper(null, invalidSupabase)).rejects.toThrow();
    });

    it('should handle supabase client without auth', async () => {
      mockIsBrowser.mockReturnValue(true);
      
      const mockSupabase = {} as any; // No auth property

      await expect(load_helper(null, mockSupabase)).rejects.toThrow();
    });

    it('should handle malformed user data from getUser()', async () => {
      mockIsBrowser.mockReturnValue(true);
      
      const malformedUser = { 
        // Missing required fields
        some_field: 'value' 
      };

      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: malformedUser },
            error: null
          })
        }
      } as any;

      const result = await load_helper(null, mockSupabase);

      expect(result.user).toEqual(malformedUser);
      expect(result.session?.user).toEqual(malformedUser);
    });

    it('should handle empty user object', async () => {
      mockIsBrowser.mockReturnValue(true);
      
      const emptyUser = {};

      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: emptyUser },
            error: null
          })
        }
      } as any;

      const result = await load_helper(null, mockSupabase);

      expect(result.user).toEqual(emptyUser);
      expect(result.session?.user).toEqual(emptyUser);
    });

    it('should create proper session-like object structure', async () => {
      mockIsBrowser.mockReturnValue(true);
      
      const mockUser = {
        id: 'test-id',
        email: 'test@example.com'
      };

      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: mockUser },
            error: null
          })
        }
      } as any;

      const result = await load_helper(null, mockSupabase);

      expect(result.session).toMatchObject({
        access_token: expect.any(String),
        refresh_token: expect.any(String),
        expires_in: expect.any(Number),
        expires_at: expect.any(Number),
        token_type: expect.any(String),
        user: expect.any(Object)
      });
    });

    it('should handle server session with missing user', async () => {
      mockIsBrowser.mockReturnValue(false);
      
      const sessionWithoutUser = {
        access_token: 'token',
        refresh_token: 'refresh',
        expires_in: 3600,
        expires_at: Date.now() / 1000 + 3600,
        token_type: 'bearer'
        // user property missing
      } as any;

      const result = await load_helper(sessionWithoutUser, {} as any);

      // Should return null for invalid session (no user)
      expect(result.session).toBe(null);
      expect(result.user).toBe(null);
    });

    it('should handle server session with null user', async () => {
      mockIsBrowser.mockReturnValue(false);
      
      const sessionWithNullUser = {
        access_token: 'token',
        user: null
      } as any;

      const result = await load_helper(sessionWithNullUser, {} as any);

      // Should return null for invalid session (null user)
      expect(result.session).toBe(null);
      expect(result.user).toBe(null);
    });
  });

  describe('isBrowser detection edge cases', () => {
    it('should handle isBrowser throwing an error', async () => {
      mockIsBrowser.mockImplementation(() => {
        throw new Error('isBrowser detection failed');
      });

      // Should throw when trying to determine environment
      await expect(load_helper(null, {} as any)).rejects.toThrow('isBrowser detection failed');
    });

    it('should handle isBrowser returning non-boolean', async () => {
      mockIsBrowser.mockReturnValue('true' as any); // String instead of boolean

      // Truthy string should be treated as client-side
      const mockSupabase = {
        auth: {
          getUser: vi.fn().mockResolvedValue({
            data: { user: null },
            error: null
          })
        }
      } as any;

      const result = await load_helper(null, mockSupabase);

      expect(mockSupabase.auth.getUser).toHaveBeenCalled();
      expect(result.session).toBe(null);
    });

    it('should handle isBrowser returning undefined', async () => {
      mockIsBrowser.mockReturnValue(undefined as any);

      // Falsy value should be treated as server-side
      const mockSession = { user: { id: 'test' } } as any;
      const result = await load_helper(mockSession, {} as any);

      expect(result.session).toEqual(mockSession);
    });
  });
});
