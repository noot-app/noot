import { writable, derived } from 'svelte/store';
import type { AuthProvider, User } from './provider';
import { SupabaseAuthProvider } from './supabase-auth';
import { isSupabaseEnabled } from '$lib/supabase';
import { authLogger } from '$lib/utils/logger';

// Track if auth has been initialized to prevent duplicate calls
let authInitialized = false;
let authProviderInstance: AuthProvider | null = null;

/**
 * Get auth provider instance - lazy-loaded to avoid SSR issues
 */
export function getAuthProvider(): AuthProvider | null {
	// During SSR, always return null since auth happens client-side
	if (typeof window === 'undefined') {
		return null;
	}
	
	// Return existing instance if available
	if (authProviderInstance) {
		return authProviderInstance;
	}
	
	// Create and cache the auth provider (client-side only)
	if (isSupabaseEnabled()) {
		try {
			authLogger.debug('Using SupabaseAuthProvider');
			authProviderInstance = new SupabaseAuthProvider();
			return authProviderInstance;
		} catch (error) {
			authLogger.warn('Failed to initialize SupabaseAuthProvider:', error);
			return null;
		}
	}
	
	authLogger.warn('getAuthProvider() No auth provider available - Supabase is not properly configured');
	return null;
}

/**
 * Reactive store for current user state
 * TODO: When implementing Supabase, ensure this store updates on auth state changes
 */
export const currentUser = writable<User | null>(null);

// Add debugging subscription to currentUser store (only in development)
if (typeof window !== 'undefined' && import.meta.env.DEV) {
	currentUser.subscribe((user) => {
		authLogger.debug('currentUser store updated to:', user ? `${user.email} (${user.id})` : null);
	});
}

/**
 * Derived store to check if user has pro subscription
 */
export const isPro = derived(currentUser, ($user) => 
	$user?.subscriptionTier === 'pro'
);

/**
 * Initialize auth and load current user
 * Call this in your root layout or app initialization
 * Sets up auth state change listener for Supabase if using SupabaseAuthProvider
 */
export async function initAuth(skipIfInitialized: boolean = true): Promise<void> {
	// Prevent duplicate initialization unless explicitly requested
	if (skipIfInitialized && authInitialized) {
		authLogger.debug('Auth already initialized, skipping');
		return;
	}

	authLogger.debug('Initializing auth system...');

	const authProvider = getAuthProvider();
	if (!authProvider) {
		// During SSR, authProvider will be null, which is expected
		if (typeof window === 'undefined') {
			authLogger.debug('Skipping auth initialization during SSR');
			return;
		}
		authLogger.warn('No auth provider available in initAuth');
		return;
	}

	try {
		// Only fetch current user if we don't already have one (to prevent duplicate calls)
		let currentUserValue: User | null = null;
		const unsubscribe = currentUser.subscribe((user: User | null) => {
			currentUserValue = user;
		});
		unsubscribe();

		authLogger.debug('Current user in store:', currentUserValue ? 'user found' : 'none');

		// If we don't have a user yet, get the current user
		if (!currentUserValue) {
			authLogger.debug('No user in store, fetching current user...');
			const user = await authProvider.getCurrentUser();
			authLogger.debug('Fetched user:', user ? `${user.email} (${user.subscriptionTier})` : 'none');
			currentUser.set(user);
		} else {
			authLogger.debug('User already in store, skipping fetch');
		}

		// Set up auth state change listener only if not already done
		if (!authInitialized && 'onAuthStateChange' in authProvider && typeof authProvider.onAuthStateChange === 'function') {
			authLogger.debug('Setting up auth state change listener');
			authProvider.onAuthStateChange((user: User | null) => {
				authLogger.debug('Auth state changed:', user ? `${user.email}` : 'signed out');
				currentUser.set(user);
			});
		}

		authInitialized = true;
		authLogger.debug('Auth initialization complete');
	} catch (error) {
		authLogger.error('Failed to initialize auth:', error);
		currentUser.set(null);
		authInitialized = true; // Mark as initialized even on error to prevent infinite retries
	}
}

/**
 * Sign in with email and password (Supabase only)
 */
export async function signIn(email: string, password: string): Promise<{ user: User | null; error: Error | null }> {
	authLogger.debug('signIn called for email:', email);

	const authProvider = getAuthProvider();
	if (!authProvider) {
		authLogger.warn('No auth provider available');
		return { user: null, error: new Error('No auth provider available') };
	}

	if ('signIn' in authProvider && typeof authProvider.signIn === 'function') {
		const result = await authProvider.signIn(email, password);
		if (result.user && !result.error) {
			authLogger.debug('Sign in successful, user:', result.user.id);
			// currentUser store will be updated via auth state change listener
			return { user: result.user, error: null };
		} else {
			authLogger.debug('Sign in failed:', result.error);
			return { user: null, error: result.error ? new Error(result.error) : new Error('Sign in failed') };
		}
	} else {
		authLogger.warn('signIn method not available on auth provider');
		return { user: null, error: new Error('Sign in method not available') };
	}
}

/**
 * Sign up with email and password (Supabase only)
 */
export async function signUp(email: string, password: string, metadata?: { fullName?: string; handle?: string }): Promise<{ user: User | null; error: Error | null }> {
	const authProvider = getAuthProvider();
	if (!authProvider) {
		return { user: null, error: new Error('No auth provider available') };
	}

	// Check if the provider supports sign up (Supabase auth)
	if ('signUp' in authProvider && typeof authProvider.signUp === 'function') {
		const result = await authProvider.signUp(email, password, metadata);
		
		// Note: Don't manually update currentUser store here since onAuthStateChange will handle it
		// This prevents duplicate user store updates during sign-up
		
		return {
			user: result.user,
			error: result.error ? new Error(result.error) : null
		};
	}

	return { user: null, error: new Error('Sign up not supported by current auth provider') };
}

/**
 * Sign out (Supabase only)
 */
export async function signOut(): Promise<{ error: Error | null }> {
	authLogger.debug('Attempting sign out');
	
	const authProvider = getAuthProvider();
	if (!authProvider) {
		return { error: new Error('No auth provider available') };
	}

	// Check if the provider supports sign out (Supabase auth)
	if ('signOut' in authProvider && typeof authProvider.signOut === 'function') {
		const result = await authProvider.signOut();
		if (!result.error) {
			authLogger.debug('Sign out successful');
			currentUser.set(null);
		} else {
			authLogger.debug('Sign out failed:', result.error);
		}
		return {
			error: result.error ? new Error(result.error) : null
		};
	}

	return { error: new Error('Sign out not supported by current auth provider') };
}

/**
 * Reset password (Supabase only)
 */
export async function resetPassword(email: string): Promise<{ error: Error | null }> {
	const authProvider = getAuthProvider();
	if (!authProvider) {
		return { error: new Error('No auth provider available') };
	}

	// Check if the provider supports password reset (Supabase auth)
	if ('resetPassword' in authProvider && typeof authProvider.resetPassword === 'function') {
		const result = await authProvider.resetPassword(email);
		return {
			error: result.error ? new Error(result.error) : null
		};
	}

	return { error: new Error('Password reset not supported by current auth provider') };
}

/**
 * Reset auth initialization state (useful for testing or manual reinitialize)
 */
export function resetAuthInitialization(): void {
	authInitialized = false;
}
