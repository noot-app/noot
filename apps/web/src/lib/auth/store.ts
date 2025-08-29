import { writable, derived } from 'svelte/store';
import type { AuthProvider, User } from './provider';
import { SupabaseAuthProvider } from './supabase-auth';
import { isSupabaseEnabled } from '$lib/supabase';

// Track if auth has been initialized to prevent duplicate calls
let authInitialized = false;

/**
 * Current auth provider instance - uses Supabase authentication
 */
function createAuthProvider(): AuthProvider | null {
	// Check if Supabase is enabled and properly configured
	if (isSupabaseEnabled()) {
		try {
			console.debug('🔐 Using SupabaseAuthProvider');
			return new SupabaseAuthProvider();
		} catch (error) {
			console.warn('Failed to initialize SupabaseAuthProvider:', error);
			return null;
		}
	}
	
	console.warn('createAuthProvider() No auth provider available - Supabase is not properly configured');
	return null;
}

export const authProvider = createAuthProvider();

/**
 * Reactive store for current user state
 * TODO: When implementing Supabase, ensure this store updates on auth state changes
 */
export const currentUser = writable<User | null>(null);

// Add debugging subscription to currentUser store (only in development)
if (typeof window !== 'undefined' && import.meta.env.DEV) {
	currentUser.subscribe((user) => {
		console.debug('📋 currentUser store updated to:', user ? `${user.email} (${user.id})` : null);
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
		console.debug('🔄 Auth already initialized, skipping');
		return;
	}

	console.debug('🚀 Initializing auth system...');

	if (!authProvider) {
		console.warn('❌ No auth provider available in initAuth');
		return;
	}

	try {
		// Only fetch current user if we don't already have one (to prevent duplicate calls)
		const currentUserValue = await new Promise<User | null>((resolve) => {
			const unsubscribe = currentUser.subscribe((user) => {
				resolve(user);
			});
			// Call unsubscribe after the subscription is fully established
			unsubscribe();
		});

		console.debug('📋 Current user in store:', currentUserValue ? `${currentUserValue.email}` : 'none');

		// If we don't have a user yet, get the current user
		if (!currentUserValue) {
			console.debug('🔍 No user in store, fetching current user...');
			const user = await authProvider.getCurrentUser();
			console.debug('👤 Fetched user:', user ? `${user.email} (${user.subscriptionTier})` : 'none');
			currentUser.set(user);
		} else {
			console.debug('✅ User already in store, skipping fetch');
		}

		// Set up auth state change listener only if not already done
		if (!authInitialized && 'onAuthStateChange' in authProvider && typeof authProvider.onAuthStateChange === 'function') {
			console.debug('👂 Setting up auth state change listener');
			authProvider.onAuthStateChange((user: User | null) => {
				console.debug('🔄 Auth state changed:', user ? `${user.email}` : 'signed out');
				currentUser.set(user);
			});
		}

		authInitialized = true;
		console.debug('✅ Auth initialization complete');
	} catch (error) {
		console.error('❌ Failed to initialize auth:', error);
		currentUser.set(null);
		authInitialized = true; // Mark as initialized even on error to prevent infinite retries
	}
}

/**
 * Sign in with email and password (Supabase only)
 */
export async function signIn(email: string, password: string): Promise<{ user: User | null; error: Error | null }> {
	console.debug('🔐 Attempting sign in for:', email);
	
	if (!authProvider) {
		console.error('❌ No auth provider available in signIn');
		return { user: null, error: new Error('No auth provider available') };
	}

	// Check if the provider supports sign in (Supabase auth)
	if ('signIn' in authProvider && typeof authProvider.signIn === 'function') {
		const result = await authProvider.signIn(email, password);
		
		if (result.error) {
			console.debug('❌ Sign in failed:', result.error);
		} else if (result.user) {
			console.debug('✅ Sign in successful:', result.user.email);
		}
		
		// Note: Don't manually update currentUser store here since onAuthStateChange will handle it
		// This prevents duplicate user store updates during sign-in
		
		return {
			user: result.user,
			error: result.error ? new Error(result.error) : null
		};
	}

	console.error('❌ Auth provider does not support signIn');
	return { user: null, error: new Error('Sign in not supported by current auth provider') };
}

/**
 * Sign up with email and password (Supabase only)
 */
export async function signUp(email: string, password: string, metadata?: { fullName?: string; handle?: string }): Promise<{ user: User | null; error: Error | null }> {
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
	console.debug('🚪 Attempting sign out');
	
	if (!authProvider) {
		return { error: new Error('No auth provider available') };
	}

	// Check if the provider supports sign out (Supabase auth)
	if ('signOut' in authProvider && typeof authProvider.signOut === 'function') {
		const result = await authProvider.signOut();
		if (!result.error) {
			console.debug('✅ Sign out successful');
			currentUser.set(null);
		} else {
			console.debug('❌ Sign out failed:', result.error);
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
