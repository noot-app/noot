import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock environment variables to enable Supabase
vi.mock('$env/dynamic/public', () => ({
  env: {
    PUBLIC_SUPABASE_URL: 'http://localhost:3000',
    PUBLIC_SUPABASE_ANON_KEY: 'test-key'
  }
}));

// Mock Supabase client - declare functions directly in the mock
vi.mock('@supabase/ssr', () => ({
  createBrowserClient: vi.fn(() => ({
    auth: {
      getUser: vi.fn(),
      signInWithPassword: vi.fn(),
      signUp: vi.fn(),
      signOut: vi.fn(),
      resetPasswordForEmail: vi.fn(),
      refreshSession: vi.fn(),
      onAuthStateChange: vi.fn(),
      getSession: vi.fn()
    },
    from: vi.fn(() => ({
      select: vi.fn(() => ({
        eq: vi.fn(() => ({
          single: vi.fn()
        }))
      }))
    }))
  }))
}));

// Mock the supabase module
vi.mock('$lib/supabase', () => ({
  supabase: {
    auth: {
      getUser: vi.fn(),
      signInWithPassword: vi.fn(),
      signUp: vi.fn(),
      signOut: vi.fn(),
      resetPasswordForEmail: vi.fn(),
      refreshSession: vi.fn(),
      onAuthStateChange: vi.fn(),
      getSession: vi.fn()
    },
    from: vi.fn(() => ({
      select: vi.fn(() => ({
        eq: vi.fn(() => ({
          single: vi.fn()
        }))
      }))
    }))
  },
  isSupabaseEnabled: vi.fn(() => true)
}));

import { SupabaseAuthProvider } from './supabase-auth';
import { supabase } from '$lib/supabase';

describe('SupabaseAuthProvider - Simple Tests', () => {
  let provider: SupabaseAuthProvider;
  let mockSupabase: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockSupabase = supabase;
    provider = new SupabaseAuthProvider();
  });

  describe('basic functionality', () => {
    it('should create provider successfully', () => {
      expect(provider).toBeInstanceOf(SupabaseAuthProvider);
    });

    it('should handle successful sign in', async () => {
      const mockSession = {
        user: {
          id: 'test-id',
          email: 'test@example.com'
        }
      };

      mockSupabase.auth.signInWithPassword.mockResolvedValue({
        data: { session: mockSession },
        error: null
      });

      const result = await provider.signIn('test@example.com', 'password');

      expect(result.user).toBeTruthy();
      expect(result.user?.email).toBe('test@example.com');
      expect(result.error).toBe(null);
    });

    it('should handle sign in errors', async () => {
      mockSupabase.auth.signInWithPassword.mockResolvedValue({
        data: { session: null },
        error: { message: 'Invalid credentials', code: 'invalid_credentials' }
      });

      const result = await provider.signIn('test@example.com', 'wrongpassword');

      expect(result.user).toBe(null);
      expect(result.error).toContain('Invalid credentials');
    });

    it('should get current user', async () => {
      const mockUser = {
        id: 'test-id',
        email: 'test@example.com'
      };

      mockSupabase.auth.getUser.mockResolvedValue({
        data: { user: mockUser },
        error: null
      });

      // Mock the database query
      mockSupabase.from.mockReturnValue({
        select: vi.fn().mockReturnValue({
          eq: vi.fn().mockReturnValue({
            single: vi.fn().mockResolvedValue({
              data: { email: 'test@example.com', tier: 'pro' },
              error: null
            })
          })
        })
      });

      const user = await provider.getCurrentUser();

      expect(user).toBeTruthy();
      expect(user?.email).toBe('test@example.com');
    });

    it('should handle sign out', async () => {
      mockSupabase.auth.signOut.mockResolvedValue({
        error: null
      });

      const result = await provider.signOut();

      expect(result.error).toBe(null);
      expect(mockSupabase.auth.signOut).toHaveBeenCalled();
    });
  });
});
