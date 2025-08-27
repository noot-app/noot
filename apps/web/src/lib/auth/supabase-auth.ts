import type { AuthProvider, User } from './provider';
import { supabase, isSupabaseEnabled } from '$lib/supabase';
import type { Session } from '@supabase/supabase-js';

/**
 * Supabase auth provider for production authentication
 * Implements the AuthProvider interface
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

			return await this.mapSupabaseUserToUser(session);
		} catch (error) {
			console.error('Error getting current user:', error);
			return null;
		}
	}

	/**
	 * Sign in with email and password
	 */
	async signIn(email: string, password: string): Promise<{ user: User | null; error: string | null }> {
		if (!supabase) {
			return { user: null, error: 'Supabase not configured' };
		}

		try {
			const { data, error } = await supabase.auth.signInWithPassword({
				email,
				password
			});

			if (error) {
				return { user: null, error: error.message };
			}

			if (!data.session) {
				return { user: null, error: 'No session created' };
			}

			const user = await this.mapSupabaseUserToUser(data.session);
			return { user, error: null };
		} catch (error) {
			return { user: null, error: String(error) };
		}
	}

	/**
	 * Sign up with email and password
	 */
	async signUp(email: string, password: string, metadata?: { fullName?: string; handle?: string }): Promise<{ user: User | null; error: Error | null }> {
		if (!supabase) return { user: null, error: new Error('Supabase not configured') };

		try {
			const { data, error } = await supabase.auth.signUp({
				email,
				password,
				options: {
					data: metadata ? { 
						full_name: metadata.fullName,
						handle: metadata.handle 
					} : undefined
				}
			});
			
			if (error) {
				return { user: null, error };
			}

			const user = data.session ? await this.mapSupabaseUserToUser(data.session) : null;
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

		const { data: { subscription } } = supabase.auth.onAuthStateChange(async (event, session) => {
			const user = session ? await this.mapSupabaseUserToUser(session) : null;
			callback(user);
		});

		return () => subscription.unsubscribe();
	}

	/**
	 * Map Supabase session to our User interface with real database data
	 */
	private async mapSupabaseUserToUser(session: Session): Promise<User> {
		const supabaseUser = session.user;
		
		try {
			// Fetch real user data directly from Supabase database using RLS
			if (!supabase) {
				throw new Error('Supabase client not available');
			}

			const { data, error } = await supabase
				.from('users')
				.select('id, email, subscription_tier')
				.eq('id', supabaseUser.id)
				.single();
			
			if (error) {
				console.warn('Failed to fetch user data from Supabase:', error);
			} else if (data) {
				// Use data directly from database with RLS protection
				return {
					id: data.id,
					email: data.email,
					subscriptionTier: data.subscription_tier as 'free' | 'pro',
					provider: 'supabase',
					subject: supabaseUser.id
				};
			}
		} catch (error) {
			console.warn('Failed to fetch user data from database, falling back to session data:', error);
		}

		// Fallback to session data if database query fails
		return {
			id: supabaseUser.id,
			email: supabaseUser.email || '',
			subscriptionTier: 'free', // Default fallback
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
