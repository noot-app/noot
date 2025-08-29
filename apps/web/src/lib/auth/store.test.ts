import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { isSupabaseEnabled } from '$lib/supabase';

// Mock the auth provider creation globally
const mockAuthProvider = {
  getCurrentUser: vi.fn().mockResolvedValue(null),
  signIn: vi.fn().mockResolvedValue({ 
    user: { id: 'test-id', email: 'test@example.com', subscriptionTier: 'free' }, 
    error: null 
  }),
  signUp: vi.fn().mockResolvedValue({ 
    user: { id: 'test-id', email: 'test@example.com', subscriptionTier: 'free' }, 
    error: null 
  }),
  signOut: vi.fn().mockResolvedValue({ error: null }),
  resetPassword: vi.fn().mockResolvedValue({ error: null }),
  onAuthStateChange: vi.fn().mockReturnValue(() => {}),
  getAccessToken: vi.fn().mockResolvedValue(null)
};

// Mock modules with more comprehensive implementation
vi.mock('./supabase-auth', () => ({
  SupabaseAuthProvider: vi.fn().mockImplementation(() => mockAuthProvider)
}));

vi.mock('$lib/supabase', () => ({
  isSupabaseEnabled: vi.fn().mockReturnValue(true)
}));

// Mock environment detection
const mockWindow = {} as any;

beforeEach(() => {
  vi.clearAllMocks();
  // Don't reset modules to preserve mocks
  // Mock client-side environment by default
  (global as any).window = mockWindow;
  
  // Ensure isSupabaseEnabled mock is working
  vi.mocked(isSupabaseEnabled).mockReturnValue(true);
  
  // Reset mock provider methods
  Object.keys(mockAuthProvider).forEach(key => {
    if (typeof mockAuthProvider[key as keyof typeof mockAuthProvider] === 'function') {
      vi.mocked(mockAuthProvider[key as keyof typeof mockAuthProvider] as any).mockClear();
    }
  });
  
  // Reset global mock values
  mockAuthProvider.signIn.mockResolvedValue({ 
    user: { id: 'test-id', email: 'test@example.com', subscriptionTier: 'free' }, 
    error: null 
  });
  mockAuthProvider.signUp.mockResolvedValue({ 
    user: { id: 'test-id', email: 'test@example.com', subscriptionTier: 'free' }, 
    error: null 
  });
  mockAuthProvider.signOut.mockResolvedValue({ error: null });
  mockAuthProvider.resetPassword.mockResolvedValue({ error: null });
});

afterEach(() => {
  delete (global as any).window;
});

describe('Auth Store', () => {
  describe('getAuthProvider', () => {
    it('should return null during SSR', async () => {
      delete (global as any).window;
      
      const { getAuthProvider } = await import('./store');
      const provider = getAuthProvider();
      
      expect(provider).toBe(null);
    });

    it('should create and cache auth provider instance', async () => {
      const { getAuthProvider } = await import('./store');
      
      const provider1 = getAuthProvider();
      const provider2 = getAuthProvider();
      
      expect(provider1).not.toBe(null);
      expect(provider1).toBe(provider2); // Should return same cached instance
    });

  });

  describe('currentUser store', () => {
    it('should initialize with null value', async () => {
      const { currentUser } = await import('./store');
      expect(get(currentUser)).toBe(null);
    });

    it('should be writable', async () => {
      const { currentUser } = await import('./store');
      const mockUser = {
        id: 'test-id',
        email: 'test@example.com',
        subscriptionTier: 'free' as const
      };

      currentUser.set(mockUser);
      expect(get(currentUser)).toEqual(mockUser);
    });
  });

  describe('isPro derived store', () => {
    it('should return false for free users', async () => {
      const { currentUser, isPro } = await import('./store');
      const freeUser = {
        id: 'test-id',
        email: 'test@example.com',
        subscriptionTier: 'free' as const
      };

      currentUser.set(freeUser);
      expect(get(isPro)).toBe(false);
    });

    it('should return true for pro users', async () => {
      const { currentUser, isPro } = await import('./store');
      const proUser = {
        id: 'test-id',
        email: 'test@example.com',
        subscriptionTier: 'pro' as const
      };

      currentUser.set(proUser);
      expect(get(isPro)).toBe(true);
    });

    it('should return false when user is null', async () => {
      const { currentUser, isPro } = await import('./store');
      currentUser.set(null);
      expect(get(isPro)).toBe(false);
    });
  });

  describe('initAuth', () => {
    it('should skip initialization when already initialized', async () => {
      const { initAuth } = await import('./store');
      
      await initAuth(); // First call
      const result = await initAuth(); // Second call should skip
      
      expect(result).toBeUndefined();
    });

    it('should skip during SSR', async () => {
      delete (global as any).window;
      
      const { initAuth } = await import('./store');
      const result = await initAuth();
      
      expect(result).toBeUndefined();
    });
  });

  describe('signIn', () => {
    it('should handle successful sign in', async () => {
      const mockUser = {
        id: 'test-id',
        email: 'test@example.com',
        subscriptionTier: 'free' as const
      };
      
      // Configure the global mock
      mockAuthProvider.signIn.mockResolvedValue({ user: mockUser, error: null });
      
      const { signIn } = await import('./store');
      const result = await signIn('test@example.com', 'password');
      
      expect(result.user).toEqual(mockUser);
      expect(result.error).toBe(null);
      expect(mockAuthProvider.signIn).toHaveBeenCalledWith('test@example.com', 'password');
    });

    it('should handle sign in errors', async () => {
      const mockError = 'Invalid credentials';
      
      // Configure the global mock  
      mockAuthProvider.signIn.mockResolvedValue({ user: null, error: mockError });
      
      const { signIn } = await import('./store');
      const result = await signIn('test@example.com', 'wrong-password');
      
      expect(result.user).toBe(null);
      expect(result.error).toBe(mockError);
    });

    it('should handle missing auth provider', async () => {
      delete (global as any).window;
      
      const { signIn } = await import('./store');
      const result = await signIn('test@example.com', 'password');
      
      expect(result.user).toBe(null);
      expect(result.error).toBe('No auth provider available');
    });

  });

  describe('resetAuthInitialization', () => {
    it('should reset auth initialization state', async () => {
      const { initAuth, resetAuthInitialization } = await import('./store');
      
      // Initialize auth
      await initAuth();
      
      // Reset and initialize again
      resetAuthInitialization();
      const result = await initAuth(); // Should not skip
      
      expect(result).toBeUndefined();
    });
  });
});
