/**
 * Auth provider interface for Supabase authentication
 */
export interface AuthProvider {
	/** Get the currently authenticated user */
	getCurrentUser(): Promise<User | null>;
}

/**
 * User interface for Supabase authentication
 */
export interface User {
	id: string;
	email: string;
	subscriptionTier: 'free' | 'pro';
	provider?: string;
	subject?: string;
}
