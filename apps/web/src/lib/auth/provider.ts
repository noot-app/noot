/**
 * Auth provider interface for Supabase authentication
 * Defines the contract that all auth providers must implement
 */
export interface AuthProvider {
	/** Get the currently authenticated user */
	getCurrentUser(): Promise<User | null>;
	
	/** Sign in with email and password */
	signIn?(email: string, password: string): Promise<AuthResult>;
	
	/** Sign up with email and password */
	signUp?(email: string, password: string, metadata?: UserMetadata): Promise<AuthResult>;
	
	/** Sign out the current user */
	signOut?(): Promise<ErrorResult>;
	
	/** Reset password for email */
	resetPassword?(email: string): Promise<ErrorResult>;
	
	/** Get access token for API requests */
	getAccessToken?(): Promise<string | null>;
	
	/** Subscribe to auth state changes */
	onAuthStateChange?(callback: (user: User | null) => void): () => void;
}

/**
 * User interface for authenticated users
 */
export interface User {
	/** Unique user identifier */
	id: string;
	/** User's email address */
	email: string;
	/** User's subscription tier */
	subscriptionTier: 'free' | 'pro';
	/** Auth provider identifier */
	provider?: string;
	/** Subject identifier (usually same as id) */
	subject?: string;
}

/**
 * Result of authentication operations
 */
export interface AuthResult {
	/** The authenticated user, null if failed */
	user: User | null;
	/** Error message if operation failed */
	error: string | null;
}

/**
 * Result of operations that can fail
 */
export interface ErrorResult {
	/** Error message if operation failed */
	error: string | null;
}

/**
 * Metadata for user registration
 */
export interface UserMetadata {
	/** User's full name */
	fullName?: string;
	/** User's handle/username */
	handle?: string;
}

/**
 * Auth state for reactive components
 */
export interface AuthState {
	/** Current user, null if not authenticated */
	user: User | null;
	/** Whether auth is currently loading */
	loading: boolean;
	/** Whether user is authenticated */
	isAuthenticated: boolean;
	/** Whether user has pro subscription */
	isPro: boolean;
}
