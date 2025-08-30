import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';

// Mock browser environment
const mockWindow = {} as any;

// Mock Supabase client
const mockSupabaseClient = {
  auth: {
    onAuthStateChange: vi.fn().mockReturnValue({ 
      data: { subscription: { unsubscribe: vi.fn() } } 
    }),
    signInWithPassword: vi.fn(),
    signUp: vi.fn(), 
    signOut: vi.fn(),
    resetPasswordForEmail: vi.fn(),
    getSession: vi.fn().mockResolvedValue({ data: { session: null } })
  }
};

// Mock Supabase SSR
vi.mock('@supabase/ssr', () => ({
  createBrowserClient: vi.fn().mockReturnValue(mockSupabaseClient),
  isBrowser: vi.fn().mockReturnValue(true)
}));

// Mock environment
vi.mock('$env/dynamic/public', () => ({
  env: {
    PUBLIC_SUPABASE_URL: 'http://localhost:54321',
    PUBLIC_SUPABASE_ANON_KEY: 'test-anon-key'
  }
}));

// Mock app navigation
vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$app/navigation', () => ({
  invalidateAll: vi.fn().mockResolvedValue(undefined)
}));

describe('New Auth Store', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock client-side environment
    (global as any).window = mockWindow;
  });

  afterEach(() => {
    delete (global as any).window;
  });

  describe('stores initialization', () => {
    it('should initialize session with null', async () => {
      const { session } = await import('./store');
      expect(get(session)).toBe(null);
    });

    it('should initialize user as derived from session', async () => {
      const { user, session } = await import('./store');
      expect(get(user)).toBe(null);
      
      // Mock session
      session.set({
        access_token: 'test-token',
        refresh_token: 'refresh-token',
        expires_at: Date.now() + 3600,
        expires_in: 3600,
        token_type: 'bearer',
        user: {
          id: 'test-id',
          email: 'test@example.com',
          aud: 'authenticated',
          created_at: '',
          app_metadata: {},
          user_metadata: {},
          is_anonymous: false
        }
      });
      
      expect(get(user)).toEqual({
        id: 'test-id',
        email: 'test@example.com',
        aud: 'authenticated',
        created_at: '',
        app_metadata: {},
        user_metadata: {},
        is_anonymous: false
      });
    });

    it('should initialize isAuthenticated as derived from session', async () => {
      vi.resetModules(); // Reset modules to ensure fresh state
      const { isAuthenticated, session } = await import('./store');
      
      // Reset session to null to ensure clean state
      session.set(null);
      expect(get(isAuthenticated)).toBe(false);
      
      // Mock session
      session.set({
        access_token: 'test-token',
        refresh_token: 'refresh-token',
        expires_at: Date.now() + 3600,
        expires_in: 3600,
        token_type: 'bearer',
        user: {
          id: 'test-id',
          email: 'test@example.com',
          aud: 'authenticated',
          created_at: '',
          app_metadata: {},
          user_metadata: {},
          is_anonymous: false
        }
      });
      
      expect(get(isAuthenticated)).toBe(true);
    });
  });

  describe('auth actions', () => {
    it('should handle sign in', async () => {
      mockSupabaseClient.auth.signInWithPassword.mockResolvedValue({
        data: { 
          user: { id: 'test-id', email: 'test@example.com' },
          session: { access_token: 'token' }
        },
        error: null
      });

      const { signIn } = await import('./store');
      const result = await signIn('test@example.com', 'password');
      
      expect(result.error).toBe(null);
      expect(mockSupabaseClient.auth.signInWithPassword).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password'
      });
    });

    it('should handle sign in errors', async () => {
      const authError = { message: 'Invalid credentials', name: 'AuthError' };
      mockSupabaseClient.auth.signInWithPassword.mockResolvedValue({
        data: { user: null, session: null },
        error: authError
      });

      const { signIn } = await import('./store');
      const result = await signIn('test@example.com', 'wrong-password');
      
      expect(result.error).toEqual(authError);
    });

    it('should handle sign up', async () => {
      mockSupabaseClient.auth.signUp.mockResolvedValue({
        data: { 
          user: { id: 'test-id', email: 'test@example.com' },
          session: null // Email confirmation required
        },
        error: null
      });

      const { signUp } = await import('./store');
      const result = await signUp('test@example.com', 'password', { fullName: 'Test User' });
      
      expect(result.error).toBe(null);
      expect(mockSupabaseClient.auth.signUp).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password',
        options: {
          data: {
            full_name: 'Test User'
          }
        }
      });
    });

    it('should handle sign out', async () => {
      mockSupabaseClient.auth.signOut.mockResolvedValue({
        error: null
      });

      const { signOut } = await import('./store');
      const result = await signOut();
      
      expect(result.error).toBe(null);
      expect(mockSupabaseClient.auth.signOut).toHaveBeenCalled();
    });

    it('should handle reset password', async () => {
      mockSupabaseClient.auth.resetPasswordForEmail.mockResolvedValue({
        error: null
      });

      const { resetPassword } = await import('./store');
      const result = await resetPassword('test@example.com');
      
      expect(result.error).toBe(null);
      expect(mockSupabaseClient.auth.resetPasswordForEmail).toHaveBeenCalledWith('test@example.com');
    });
  });

  describe('getAccessToken', () => {
    it('should return access token from session', async () => {
      mockSupabaseClient.auth.getSession.mockResolvedValue({
        data: { 
          session: {
            access_token: 'test-access-token',
            user: { id: 'test-id' }
          }
        }
      });

      const { getAccessToken } = await import('./store');
      const token = await getAccessToken();
      
      expect(token).toBe('test-access-token');
    });

    it('should return null when no session', async () => {
      mockSupabaseClient.auth.getSession.mockResolvedValue({
        data: { session: null }
      });

      const { getAccessToken } = await import('./store');
      const token = await getAccessToken();
      
      expect(token).toBe(null);
    });
  });
});
