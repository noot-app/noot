import { writable, derived } from 'svelte/store';
import type { AuthProvider, User } from './provider';
import { SupabaseAuthProvider } from './supabase-auth';
import { isSupabaseEnabled } from '$lib/supabase';

/**
 * Current auth provider instance - uses Supabase authentication
 */
function createAuthProvider(): AuthProvider | null {
	// Check if Supabase is enabled and properly configured
	if (isSupabaseEnabled()) {
		try {
			console.log('Using SupabaseAuthProvider');
			return new SupabaseAuthProvider();
		} catch (error) {
			console.warn('Failed to initialize SupabaseAuthProvider:', error);
			return null;
		}
	}
	
	console.warn('No auth provider available - Supabase is not properly configured');
	return null;
}

export const authProvider = createAuthProvider();

/**
 * Reactive store for current user state
 * TODO: When implementing Supabase, ensure this store updates on auth state changes
 */
export const currentUser = writable<User | null>(null);

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
export async function initAuth(): Promise<void> {
	if (!authProvider) {
		console.warn('No auth provider available');
		return;
	}

	try {
		const user = await authProvider.getCurrentUser();
		currentUser.set(user);

		// If using SupabaseAuthProvider, set up auth state change listener
		if ('onAuthStateChange' in authProvider && typeof authProvider.onAuthStateChange === 'function') {
			authProvider.onAuthStateChange((user: User | null) => {
				currentUser.set(user);
			});
		}
	} catch (error) {
		console.error('Failed to initialize auth:', error);
		currentUser.set(null);
	}
}

/**
 * Sign in with email and password (Supabase only)
 */
export async function signIn(email: string, password: string): Promise<{ user: User | null; error: Error | null }> {
	if (!authProvider) {
		return { user: null, error: new Error('No auth provider available') };
	}

	// Check if the provider supports sign in (Supabase auth)
	if ('signIn' in authProvider && typeof authProvider.signIn === 'function') {
		return await authProvider.signIn(email, password);
	}

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
		return await authProvider.signUp(email, password, metadata);
	}

	return { user: null, error: new Error('Sign up not supported by current auth provider') };
}

/**
 * Sign out (Supabase only)
 */
export async function signOut(): Promise<{ error: Error | null }> {
	if (!authProvider) {
		return { error: new Error('No auth provider available') };
	}

	// Check if the provider supports sign out (Supabase auth)
	if ('signOut' in authProvider && typeof authProvider.signOut === 'function') {
		const result = await authProvider.signOut();
		if (!result.error) {
			currentUser.set(null);
		}
		return result;
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
		return await authProvider.resetPassword(email);
	}

	return { error: new Error('Password reset not supported by current auth provider') };
}
