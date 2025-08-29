import type { AuthProvider, User, AuthResult, ErrorResult, UserMetadata } from './provider';
import { supabase, isSupabaseEnabled } from '$lib/supabase';
import type { Session, User as SupabaseUser } from '@supabase/supabase-js';
import { authLogger } from '$lib/utils/logger';
import { userProfileCache } from '$lib/utils/user-cache';
import { AuthError, sanitizeError } from '$lib/utils/errors';
import { tokenRefreshLimiter, withRateLimit } from '$lib/utils/rate-limit';

// Type for minimal session-like object we create from verified user
type VerifiedSession = {
	user: SupabaseUser;
	access_token: string;
	refresh_token: string;
	expires_in: number;
	expires_at: number;
	token_type: string;
};

/**
 * Supabase auth provider for production authentication
 * Implements the AuthProvider interface
 */
export class SupabaseAuthProvider implements AuthProvider {
	private cachedAccessToken: string | null = null;
	private tokenExpiryTime: number = 0;
	private lastProcessedUserId: string | null = null;
	private lastProcessedEvent: string | null = null;
	private lastProcessedTime: number = 0;

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

		authLogger.debug('Getting current user from Supabase...');
		try {
			// Use getUser() instead of getSession() for security - validates token with server
			const { data: { user }, error } = await supabase.auth.getUser();
			
			if (error) {
				// Don't log session missing errors as errors since they're expected when not logged in
				if (error.message?.includes('Auth session missing')) {
					authLogger.debug('No active auth session (user not logged in)');
				} else {
					authLogger.error('Failed to get Supabase user:', error);
				}
				return null;
			}

			if (!user) {
				authLogger.debug('No user found in Supabase session');
				return null;
			}

			authLogger.debug('Supabase user found, mapping to app user...');
			// We have a verified user, create a minimal session-like object for mapping
			const verifiedSession = {
				user: user,
				access_token: '', // Not needed for user mapping
				refresh_token: '',
				expires_in: 0,
				expires_at: 0,
				token_type: 'bearer' as const
			};

			const mappedUser = await this.mapSupabaseUserToUser(verifiedSession);
			authLogger.debug('Successfully mapped Supabase user:', mappedUser.email);
			return mappedUser;
		} catch (error) {
			authLogger.error('Error getting current user:', error);
			return null;
		}
	}

	/**
	 * Sign in with email and password
	 */
	async signIn(email: string, password: string): Promise<AuthResult> {
		if (!supabase) {
			authLogger.error('Supabase client not available in signIn');
			return { user: null, error: 'Supabase not configured' };
		}

		try {
			const { data, error } = await supabase.auth.signInWithPassword({
				email,
				password
			});

			if (error) {
				authLogger.warn('Supabase auth error:', error);
				const sanitized = sanitizeError(error);
				return { 
					user: null, 
					error: JSON.stringify({ 
						code: error.code || error.name || 'unknown_error', 
						message: sanitized.message 
					}) 
				};
			}

			if (!data.session) {
				authLogger.error('No session created after successful auth');
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
			const sanitized = sanitizeError(error);
			authLogger.error('Exception in SupabaseAuthProvider.signIn:', sanitized.message);
			return { user: null, error: sanitized.message };
		}
	}

	/**
	 * Sign up with email and password
	 */
	async signUp(email: string, password: string, metadata?: UserMetadata): Promise<AuthResult> {
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
				authLogger.warn('Supabase signup error:', error);
				const sanitized = sanitizeError(error);
				return { 
					user: null, 
					error: JSON.stringify({ 
						code: error.code || error.name || 'unknown_error', 
						message: sanitized.message 
					}) 
				};
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
			const sanitized = sanitizeError(error);
			return { user: null, error: JSON.stringify({ code: 'signup_error', message: sanitized.message }) };
		}
	}

	/**
	 * Sign out
	 */
	async signOut(): Promise<ErrorResult> {
		if (!supabase) return { error: JSON.stringify({ code: 'no_supabase', message: 'Supabase not configured' }) };

		try {
			const { error } = await supabase.auth.signOut();
			if (error) {
				authLogger.warn('Supabase signout error:', error);
				const sanitized = sanitizeError(error);
				return { 
					error: JSON.stringify({ 
						code: error.code || error.name || 'unknown_error', 
						message: sanitized.message 
					}) 
				};
			}
			return { error: null };
		} catch (error) {
			const sanitized = sanitizeError(error);
			return { error: JSON.stringify({ code: 'signout_error', message: sanitized.message }) };
		}
	}

	/**
	 * Reset password
	 */
	async resetPassword(email: string): Promise<ErrorResult> {
		if (!supabase) return { error: JSON.stringify({ code: 'no_supabase', message: 'Supabase not configured' }) };

		try {
			const { error } = await supabase.auth.resetPasswordForEmail(email);
			if (error) {
				authLogger.warn('Supabase reset password error:', error);
				const sanitized = sanitizeError(error);
				return { 
					error: JSON.stringify({ 
						code: error.code || error.name || 'unknown_error', 
						message: sanitized.message 
					}) 
				};
			}
			return { error: null };
		} catch (error) {
			const sanitized = sanitizeError(error);
			return { error: JSON.stringify({ code: 'reset_password_error', message: sanitized.message }) };
		}
	}

	/**
	 * Subscribe to auth state changes
	 */
	onAuthStateChange(callback: (user: User | null) => void) {
		if (!supabase) return () => {};

		authLogger.debug('Setting up Supabase auth state change listener');
		const { data: { subscription } } = supabase.auth.onAuthStateChange(async (event, session) => {
			authLogger.debug('Supabase auth event:', event);
			
			// Deduplication: Skip processing if this is the same event for the same user within a short time window
			const currentTime = Date.now();
			const currentUserId = session?.user?.id || null;
			
			if (
				this.lastProcessedEvent === event &&
				this.lastProcessedUserId === currentUserId &&
				currentTime - this.lastProcessedTime < 1000 // 1 second window
			) {
				authLogger.debug('Skipping duplicate auth event processing');
				return;
			}
			
			// Update deduplication tracking
			this.lastProcessedEvent = event;
			this.lastProcessedUserId = currentUserId;
			this.lastProcessedTime = currentTime;
			
			// Cache the access token when we get a valid session from auth events
			if (session?.access_token && event !== 'TOKEN_REFRESHED') {
				this.cachedAccessToken = session.access_token;
				this.tokenExpiryTime = session.expires_at ? session.expires_at * 1000 : 0;
				authLogger.debug('Cached access token from auth event');
			}
			
			if (session?.user && supabase) {
				// Verify the user is authentic using getUser() instead of trusting session directly
				try {
					authLogger.debug('Verifying user authenticity after auth state change...');
					const { data: { user }, error } = await supabase.auth.getUser();
					if (error || !user) {
						authLogger.debug('User verification failed, setting to null');
						this.cachedAccessToken = null;
						this.tokenExpiryTime = 0;
						callback(null);
						return;
					}
					// User is verified, now we can safely use the session for mapping
					const verifiedSessionObj: VerifiedSession = {
						user: user,
						access_token: session.access_token || '',
						refresh_token: session.refresh_token || '',
						expires_in: session.expires_in || 0,
						expires_at: session.expires_at || 0,
						token_type: session.token_type || 'bearer'
					};
					const mappedUser = await this.mapSupabaseUserToUser(verifiedSessionObj);
					authLogger.debug('Auth state change verified, user:', mappedUser.email);
					callback(mappedUser);
				} catch (error) {
					authLogger.error('Error verifying user in auth state change:', error);
					this.cachedAccessToken = null;
					this.tokenExpiryTime = 0;
					callback(null);
				}
			} else {
				authLogger.debug('Auth state change: no session or user, setting to null');
				this.cachedAccessToken = null;
				this.tokenExpiryTime = 0;
				// Reset deduplication state on sign out
				this.lastProcessedUserId = null;
				this.lastProcessedEvent = null;
				this.lastProcessedTime = 0;
				// Clear user profile cache on sign out
				userProfileCache.clear();
				callback(null);
			}
		});

		return () => {
			authLogger.debug('Unsubscribing from Supabase auth state changes');
			subscription.unsubscribe();
		};
	}

	/**
	 * Map Supabase session to our User interface with real database data
	 * Uses caching to avoid redundant database queries
	 */
	private async mapSupabaseUserToUser(session: Session | VerifiedSession): Promise<User> {
		const supabaseUser = session.user;
		authLogger.debug('Mapping Supabase user to app user:', supabaseUser.email);
		
		// Check cache first
		const cachedUser = userProfileCache.get(supabaseUser.id);
		if (cachedUser) {
			authLogger.debug('Using cached user profile for:', supabaseUser.email);
			return cachedUser;
		}
		
		try {
			// Fetch real user data directly from Supabase database using RLS
			if (!supabase) {
				throw new Error('Supabase client not available');
			}

			authLogger.debug('Fetching user profile from database...');
			const { data, error } = await supabase
				.from('profiles')
				.select('id, email, subscription_tier')
				.eq('id', supabaseUser.id)
				.single();
			
			if (error) {
				authLogger.warn('Failed to fetch user data from Supabase:', error);
			} else if (data) {
				// Use data directly from database with RLS protection
				authLogger.debug('User profile loaded from database:', {
					email: data.email,
					tier: data.subscription_tier
				});
				const user: User = {
					id: data.id,
					email: data.email,
					subscriptionTier: data.subscription_tier as 'free' | 'pro',
					provider: 'supabase',
					subject: supabaseUser.id
				};
				
				// Cache the user profile
				userProfileCache.set(supabaseUser.id, user);
				return user;
			}
		} catch (error) {
			authLogger.warn('Failed to fetch user data from database, falling back to session data:', error);
		}

		// Fallback to session data if database query fails
		authLogger.debug('Using session data fallback for user mapping');
		const fallbackUser: User = {
			id: supabaseUser.id,
			email: supabaseUser.email || '',
			subscriptionTier: 'free' as const, // Default fallback
			provider: 'supabase',
			subject: supabaseUser.id
		};
		
		// Cache the fallback user for a shorter duration to retry database query later
		userProfileCache.set(supabaseUser.id, fallbackUser);
		return fallbackUser;
	}

	/**
	 * Get the current JWT token for API requests
	 * Uses cached token from auth events to avoid getSession() calls
	 */
	async getAccessToken(): Promise<string | null> {
		if (!supabase) return null;

		try {
			// First verify the user is still authentic
			const { data: { user }, error: userError } = await supabase.auth.getUser();
			
			if (userError || !user) {
				authLogger.debug('No verified user for access token');
				this.cachedAccessToken = null;
				this.tokenExpiryTime = 0;
				return null;
			}

			// Check if we have a valid cached token
			const now = Date.now();
			const bufferTime = 60000; // 1 minute buffer before expiry
			
			if (this.cachedAccessToken && this.tokenExpiryTime > (now + bufferTime)) {
				authLogger.debug('Using cached access token');
				return this.cachedAccessToken;
			}

			// Token is expired or missing, we need to refresh
			authLogger.debug('Access token expired/missing, refreshing...');
			
			try {
				// Use rate-limited token refresh
				const refreshResult = await withRateLimit(
					tokenRefreshLimiter,
					'token_refresh',
					() => supabase!.auth.refreshSession() // Use non-null assertion since we checked above
				);
				
				if (refreshResult.error || !refreshResult.data.session) {
					authLogger.debug('Failed to refresh session for access token');
					this.cachedAccessToken = null;
					this.tokenExpiryTime = 0;
					return null;
				}

				// Cache the new token
				this.cachedAccessToken = refreshResult.data.session.access_token;
				this.tokenExpiryTime = refreshResult.data.session.expires_at 
					? refreshResult.data.session.expires_at * 1000 
					: 0;
				
				authLogger.debug('Access token refreshed and cached');
				return this.cachedAccessToken;
			} catch (rateLimitError) {
				authLogger.warn('Token refresh rate limited:', rateLimitError);
				// Return cached token even if expired rather than failing completely
				return this.cachedAccessToken;
			}
		} catch (error) {
			authLogger.error('Failed to get access token:', error);
			this.cachedAccessToken = null;
			this.tokenExpiryTime = 0;
			return null;
		}
	}
}
