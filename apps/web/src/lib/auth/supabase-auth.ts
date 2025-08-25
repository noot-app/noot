import type { AuthProvider, User } from './provider';
import { supabase, isSupabaseEnabled } from '$lib/supabase';
import type { Session } from '@supabase/supabase-js';

/**
 * Supabase auth provider for production authentication
 * Implements the AuthProvider interface for seamless switching between dev and prod auth
 */
export class SupabaseAuthProvider implements AuthProvider {
	constructor() {
		if (!isSupabaseEnabled()) {
			throw new Error('Supabase is not properly configured');
		}
	}

	/**
	 * Get the currently authenticated user from Supabase session
	 */
	async getCurrentUser(): Promise<User | null> {
		if (!supabase) return null;

		try {
			const { data: { session }, error } = await supabase.auth.getSession();
			
			if (error) {
				console.error('Failed to get Supabase session:', error);
				return null;
			}

			if (!session?.user) {
				return null;
			}

			return this.mapSupabaseUserToUser(session);
		} catch (error) {
			console.error('Error getting current user:', error);
			return null;
		}
	}

	/**
	 * User switching is not supported in production Supabase auth
	 */
	supportsUserSwitching(): boolean {
		return false;
	}

	/**
	 * Sign in with email and password
	 */
	async signIn(email: string, password: string): Promise<{ user: User | null; error: Error | null }> {
		if (!supabase) return { user: null, error: new Error('Supabase not configured') };

		try {
			const { data, error } = await supabase.auth.signInWithPassword({ email, password });
			
			if (error) {
				return { user: null, error };
			}

			const user = data.session ? this.mapSupabaseUserToUser(data.session) : null;
			return { user, error: null };
		} catch (error) {
			return { user: null, error: error as Error };
		}
	}

	/**
	 * Sign up with email and password
	 */
	async signUp(email: string, password: string, metadata?: { fullName?: string }): Promise<{ user: User | null; error: Error | null }> {
		if (!supabase) return { user: null, error: new Error('Supabase not configured') };

		try {
			const { data, error } = await supabase.auth.signUp({
				email,
				password,
				options: {
					data: metadata ? { full_name: metadata.fullName } : undefined
				}
			});
			
			if (error) {
				return { user: null, error };
			}

			const user = data.session ? this.mapSupabaseUserToUser(data.session) : null;
			return { user, error: null };
		} catch (error) {
			return { user: null, error: error as Error };
		}
	}

	/**
	 * Sign out
	 */
	async signOut(): Promise<{ error: Error | null }> {
		if (!supabase) return { error: new Error('Supabase not configured') };

		try {
			const { error } = await supabase.auth.signOut();
			return { error };
		} catch (error) {
			return { error: error as Error };
		}
	}

	/**
	 * Reset password
	 */
	async resetPassword(email: string): Promise<{ error: Error | null }> {
		if (!supabase) return { error: new Error('Supabase not configured') };

		try {
			const { error } = await supabase.auth.resetPasswordForEmail(email);
			return { error };
		} catch (error) {
			return { error: error as Error };
		}
	}

	/**
	 * Subscribe to auth state changes
	 */
	onAuthStateChange(callback: (user: User | null) => void) {
		if (!supabase) return () => {};

		const { data: { subscription } } = supabase.auth.onAuthStateChange((event, session) => {
			const user = session ? this.mapSupabaseUserToUser(session) : null;
			callback(user);
		});

		return () => subscription.unsubscribe();
	}

	/**
	 * Map Supabase session to our User interface
	 */
	private mapSupabaseUserToUser(session: Session): User {
		const supabaseUser = session.user;
		
		// For now, default to free tier - in production you'd query your subscription service
		// TODO: Implement proper subscription tier detection from database or Stripe
		const subscriptionTier: 'free' | 'pro' = 'free';

		return {
			id: supabaseUser.id,
			email: supabaseUser.email || '',
			subscriptionTier,
			provider: 'supabase',
			subject: supabaseUser.id
		};
	}

	/**
	 * Get the current JWT token for API requests
	 */
	async getAccessToken(): Promise<string | null> {
		if (!supabase) return null;

		try {
			const { data: { session }, error } = await supabase.auth.getSession();
			
			if (error || !session) {
				return null;
			}

			return session.access_token;
		} catch (error) {
			console.error('Failed to get access token:', error);
			return null;
		}
	}
}