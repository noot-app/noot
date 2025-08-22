import { writable, derived } from 'svelte/store';
import { dev } from '$app/environment';
import type { AuthProvider, User } from './provider';
import { DevAuthProvider } from './dev-auth';

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
 * TODO: When implementing Supabase, add logic to switch between DevAuthProvider and SupabaseAuthProvider
 * based on environment variables or feature flags
 */
function createAuthProvider(): AuthProvider | null {
	if (dev) {
		try {
			return new DevAuthProvider();
		} catch (error) {
			console.warn('Failed to initialize DevAuthProvider:', error);
			return null;
		}
	}
	
	// TODO: When implementing Supabase auth, add:
	// if (SUPABASE_ENABLED) {
	//   return new SupabaseAuthProvider();
	// }
	
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
 */
export async function initAuth(): Promise<void> {
	if (!authProvider) {
		console.warn('No auth provider available');
		return;
	}

	try {
		const user = await authProvider.getCurrentUser();
		currentUser.set(user);
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
 * TODO: When implementing Supabase, this should return an empty array or be removed
 */
export function getAvailableDevUsers() {
	if (!authProvider?.supportsUserSwitching()) {
		return [];
	}

	return DEV_USERS;
}