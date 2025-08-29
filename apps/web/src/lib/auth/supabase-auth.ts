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

		console.debug('🔍 Getting current user from Supabase...');
		try {
			// Use getUser() instead of getSession() for security - validates token with server
			const { data: { user }, error } = await supabase.auth.getUser();
			
			if (error) {
				// Don't log session missing errors as errors since they're expected when not logged in
				if (error.message?.includes('Auth session missing')) {
					console.debug('🔓 No active auth session (user not logged in)');
				} else {
					console.error('❌ Failed to get Supabase user:', error);
				}
				return null;
			}

			if (!user) {
				console.debug('👤 No user found in Supabase session');
				return null;
			}

			console.debug('🔍 Supabase user found, getting session for mapping...');
			// We still need the session for mapping, but get it separately
			const { data: { session }, error: sessionError } = await supabase.auth.getSession();
			
			if (sessionError || !session) {
				console.error('❌ Failed to get session for user mapping:', sessionError);
				return null;
			}

			const mappedUser = await this.mapSupabaseUserToUser(session);
			console.debug('✅ Successfully mapped Supabase user:', mappedUser.email);
			return mappedUser;
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

			// Note: Don't call mapSupabaseUserToUser here to avoid duplicate mapping
			// The onAuthStateChange listener will handle user mapping automatically
			// Return a basic user object for the function response
			const basicUser: User = {
				id: data.session.user.id,
				email: data.session.user.email || '',
				subscriptionTier: 'free', // Will be updated by onAuthStateChange with real data
				provider: 'supabase',
				subject: data.session.user.id
			};
			
			return { user: basicUser, error: null };
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

			// Note: Don't call mapSupabaseUserToUser here to avoid duplicate mapping
			// The onAuthStateChange listener will handle user mapping automatically  
			const basicUser = data.session ? {
				id: data.session.user.id,
				email: data.session.user.email || '',
				subscriptionTier: 'free' as const, // Will be updated by onAuthStateChange with real data
				provider: 'supabase',
				subject: data.session.user.id
			} : null;
			
			return { user: basicUser, error: null };
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

		console.debug('👂 Setting up Supabase auth state change listener');
		const { data: { subscription } } = supabase.auth.onAuthStateChange(async (event, session) => {
			console.debug('🔄 Supabase auth event:', event);
			
			if (session?.user && supabase) {
				// Verify the user is authentic using getUser() instead of trusting session directly
				try {
					console.debug('🔍 Verifying user authenticity after auth state change...');
					const { data: { user }, error } = await supabase.auth.getUser();
					if (error || !user) {
						console.debug('❌ User verification failed, setting to null');
						callback(null);
						return;
					}
					// User is verified, now we can safely use the session for mapping
					const mappedUser = await this.mapSupabaseUserToUser(session);
					console.debug('✅ Auth state change verified, user:', mappedUser.email);
					callback(mappedUser);
				} catch (error) {
					console.error('❌ Error verifying user in auth state change:', error);
					callback(null);
				}
			} else {
				console.debug('🔓 Auth state change: no session or user, setting to null');
				callback(null);
			}
		});

		return () => {
			console.debug('🔌 Unsubscribing from Supabase auth state changes');
			subscription.unsubscribe();
		};
	}

	/**
	 * Map Supabase session to our User interface with real database data
	 */
	private async mapSupabaseUserToUser(session: Session): Promise<User> {
		const supabaseUser = session.user;
		console.debug('🗂️ Mapping Supabase user to app user:', supabaseUser.email);
		
		try {
			// Fetch real user data directly from Supabase database using RLS
			if (!supabase) {
				throw new Error('Supabase client not available');
			}

			console.debug('🔍 Fetching user profile from database...');
			const { data, error } = await supabase
				.from('profiles')
				.select('id, email, subscription_tier')
				.eq('id', supabaseUser.id)
				.single();
			
			if (error) {
				console.warn('⚠️ Failed to fetch user data from Supabase:', error);
			} else if (data) {
				// Use data directly from database with RLS protection
				console.debug('✅ User profile loaded from database:', {
					email: data.email,
					tier: data.subscription_tier
				});
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
		console.debug('📋 Using session data fallback for user mapping');
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
