/**
 * Auth provider interface for abstracting authentication methods
 * Implemented by both DevAuthProvider and SupabaseAuthProvider
 */
export interface AuthProvider {
	/** Get the currently authenticated user */
	getCurrentUser(): Promise<User | null>;
	
	/** Switch to a different user (development only) */
	switchUser?(userId: string): Promise<void>;
	
	/** Check if the provider supports user switching (dev mode only) */
	supportsUserSwitching(): boolean;
}

/**
 * User interface compatible with both dev and Supabase implementations
 * Extended with Supabase user properties
 */
export interface User {
	id: string;
	email: string;
	subscriptionTier: 'free' | 'pro';
	provider?: string;
	subject?: string;
}

/**
 * Available dev users for switching
 * TODO: Remove or make this dynamic when implementing Supabase auth
 */
export const DEV_USERS = [
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