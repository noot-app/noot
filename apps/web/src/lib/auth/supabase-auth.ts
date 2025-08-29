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
	 * Get the currently authenticated user from Supabase (securely using getUser)
	 */
	async getCurrentUser(): Promise<User | null> {
		if (!supabase) return null;

		try {
			// Use getUser() instead of getSession() for security - validates token with server
			const { data: { user }, error } = await supabase.auth.getUser();
			
			if (error) {
				console.error('❌ Failed to get Supabase user:', error);
				return null;
			}

			if (!user) {
				return null;
			}

			// We still need the session for mapping, but get it separately
			const { data: { session }, error: sessionError } = await supabase.auth.getSession();
			
			if (sessionError || !session) {
				console.error('❌ Failed to get session for user mapping:', sessionError);
				return null;
			}

			return await this.mapSupabaseUserToUser(session);
		} catch (error) {
			console.error('❌ Error getting current user:', error);
			return null;
		}
	}

	/**
	 * Sign in with email and password
	 */
	async signIn(email: string, password: string): Promise<{ user: User | null; error: string | null }> {
		if (!supabase) {
			console.error('❌ Supabase client not available in signIn');
			return { user: null, error: 'Supabase not configured' };
		}

		try {
			const { data, error } = await supabase.auth.signInWithPassword({
				email,
				password
			});

			if (error) {
				console.warn('❌ Supabase auth error:', error);
				return { user: null, error: JSON.stringify({ code: error.code || error.name || 'unknown_error', message: error.message }) };
			}

			if (!data.session) {
				console.error('❌ No session created after successful auth');
				return { user: null, error: 'No session created' };
			}

			const user = await this.mapSupabaseUserToUser(data.session);
			return { user, error: null };
		} catch (error) {
			console.error('❌ Exception in SupabaseAuthProvider.signIn:', error);
			return { user: null, error: String(error) };
		}
	}

	/**
	 * Sign up with email and password
	 */
	async signUp(email: string, password: string, metadata?: { fullName?: string; handle?: string }): Promise<{ user: User | null; error: string | null }> {
		if (!supabase) return { user: null, error: JSON.stringify({ code: 'no_supabase', message: 'Supabase not configured' }) };

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
				console.warn('Supabase signup error:', error);
				return { user: null, error: JSON.stringify({ code: error.code || error.name || 'unknown_error', message: error.message }) };
			}

			const user = data.session ? await this.mapSupabaseUserToUser(data.session) : null;
			return { user, error: null };
		} catch (error) {
			return { user: null, error: JSON.stringify({ code: 'signup_error', message: String(error) }) };
		}
	}

	/**
	 * Sign out
	 */
	async signOut(): Promise<{ error: string | null }> {
		if (!supabase) return { error: JSON.stringify({ code: 'no_supabase', message: 'Supabase not configured' }) };

		try {
			const { error } = await supabase.auth.signOut();
			if (error) {
				console.warn('Supabase signout error:', error);
				return { error: JSON.stringify({ code: error.code || error.name || 'unknown_error', message: error.message }) };
			}
			return { error: null };
		} catch (error) {
			return { error: JSON.stringify({ code: 'signout_error', message: String(error) }) };
		}
	}

	/**
	 * Reset password
	 */
	async resetPassword(email: string): Promise<{ error: string | null }> {
		if (!supabase) return { error: JSON.stringify({ code: 'no_supabase', message: 'Supabase not configured' }) };

		try {
			const { error } = await supabase.auth.resetPasswordForEmail(email);
			if (error) {
				console.warn('Supabase reset password error:', error);
				return { error: JSON.stringify({ code: error.code || error.name || 'unknown_error', message: error.message }) };
			}
			return { error: null };
		} catch (error) {
			return { error: JSON.stringify({ code: 'reset_password_error', message: String(error) }) };
		}
	}

	/**
	 * Subscribe to auth state changes
	 */
	onAuthStateChange(callback: (user: User | null) => void) {
		if (!supabase) return () => {};

		const { data: { subscription } } = supabase.auth.onAuthStateChange(async (event, session) => {
			if (session?.user) {
				// Verify the user is authentic using getUser() instead of trusting session directly
				try {
					const { data: { user }, error } = await supabase.auth.getUser();
					if (error || !user) {
						callback(null);
						return;
					}
					// User is verified, now we can safely use the session for mapping
					const mappedUser = await this.mapSupabaseUserToUser(session);
					callback(mappedUser);
				} catch (error) {
					console.error('❌ Error verifying user in auth state change:', error);
					callback(null);
				}
			} else {
				callback(null);
			}
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
				.from('profiles')
				.select('id, email, subscription_tier')
				.eq('id', supabaseUser.id)
				.single();
			
			if (error) {
				console.warn('⚠️ Failed to fetch user data from Supabase:', error);
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
			console.warn('⚠️ Failed to fetch user data from database, falling back to session data:', error);
		}

		// Fallback to session data if database query fails
		return {
			id: supabaseUser.id,
			email: supabaseUser.email || '',
			subscriptionTier: 'free' as const, // Default fallback
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
			// First verify the user is authentic
			const { data: { user }, error: userError } = await supabase.auth.getUser();
			
			if (userError || !user) {
				return null;
			}

			// Now get the session for the access token
			const { data: { session }, error: sessionError } = await supabase.auth.getSession();
			
			if (sessionError || !session) {
				return null;
			}

			return session.access_token;
		} catch (error) {
			console.error('Failed to get access token:', error);
			return null;
		}
	}
}
