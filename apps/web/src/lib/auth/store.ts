import { writable, derived } from 'svelte/store';
import { dev } from '$app/environment';
import { env } from '$env/dynamic/public';
import type { AuthProvider, User } from './provider';
import { DevAuthProvider } from './dev-auth';
import { SupabaseAuthProvider } from './supabase-auth';
import { isSupabaseEnabled } from '$lib/supabase';

// Local copy of available dev users to avoid circular imports  
const DEV_USERS = [
	{
		id: 'monalisa',
		email: 'monalisa@birki.io',
		subscriptionTier: 'pro' as const,
		displayName: 'Monalisa (Pro)'
	},
	{
		id: 'alice', 
		email: 'alice@birki.io',
		subscriptionTier: 'free' as const,
		displayName: 'Alice (Free)'
	}
] as const;

/**
 * Current auth provider instance
 * Switches between DevAuthProvider (development) and SupabaseAuthProvider (production)
 * based on environment variables and configuration
 */
function createAuthProvider(): AuthProvider | null {
	// Check if we should force dev auth even in production (for testing)
	const forceDevAuth = env.PUBLIC_FORCE_DEV_AUTH === 'true';
	
	if (dev || forceDevAuth) {
		try {
			return new DevAuthProvider();
		} catch (error) {
			console.warn('Failed to initialize DevAuthProvider:', error);
			return null;
		}
	}
	
	// In production, try to use Supabase auth if configured
	if (isSupabaseEnabled()) {
		try {
			return new SupabaseAuthProvider();
		} catch (error) {
			console.warn('Failed to initialize SupabaseAuthProvider:', error);
			return null;
		}
	}
	
	console.warn('No auth provider available - neither dev auth nor Supabase is properly configured');
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
 * Derived store to check if user switching is available
 */
export const canSwitchUsers = derived([currentUser], () => 
	authProvider?.supportsUserSwitching() ?? false
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
 * Switch to a different user (development only)
 */
export async function switchUser(userId: string): Promise<void> {
	if (!authProvider?.supportsUserSwitching()) {
		throw new Error('User switching is not supported in this environment');
	}

	try {
		await authProvider.switchUser!(userId);
		// Note: DevAuthProvider triggers a page reload, so this line may not execute
		const user = await authProvider.getCurrentUser();
		currentUser.set(user);
	} catch (error) {
		console.error('Failed to switch user:', error);
		throw error;
	}
}

/**
 * Get available dev users for switching
 * Only available when using DevAuthProvider
 */
export function getAvailableDevUsers() {
	if (!authProvider?.supportsUserSwitching()) {
		return [];
	}

	return DEV_USERS;
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
export async function signUp(email: string, password: string, metadata?: { fullName?: string }): Promise<{ user: User | null; error: Error | null }> {
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
