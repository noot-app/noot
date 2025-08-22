import { dev } from '$app/environment';
import { apiClient } from '$lib/api/client';
import type { AuthProvider, User, DEV_USERS } from './provider';

/**
 * Development auth provider that handles user switching via headers
 * TODO: Replace this with SupabaseAuthProvider when implementing Supabase auth
 */
export class DevAuthProvider implements AuthProvider {
	private readonly STORAGE_KEY = 'dev-selected-user';
	
	constructor() {
		// Only allow in development mode
		if (!dev) {
			throw new Error('DevAuthProvider can only be used in development mode');
		}
	}

	/**
	 * Get the currently selected dev user
	 * TODO: When implementing Supabase, replace this with JWT token validation
	 */
	async getCurrentUser(): Promise<User | null> {
		if (!dev) return null;
		
		const selectedUserId = this.getSelectedUserId();
		
		try {
			// Make API call with dev user header to get user info
			const response = await apiClient.GET('/health', {
				headers: selectedUserId ? { 'X-Dev-User-ID': selectedUserId } : {}
			});
			
			if (response.error) {
				console.warn('Failed to fetch current user:', response.error);
				return this.getDefaultUser();
			}
			
			// For now, return user info from DEV_USERS constant
			// TODO: When implementing Supabase, return actual user from API response
			return this.getUserFromId(selectedUserId) || this.getDefaultUser();
		} catch (error) {
			console.warn('Error fetching current user:', error);
			return this.getDefaultUser();
		}
	}

	/**
	 * Switch to a different dev user
	 */
	async switchUser(userId: string): Promise<void> {
		if (!dev) {
			throw new Error('User switching is only available in development mode');
		}

		const user = this.getUserFromId(userId);
		if (!user) {
			throw new Error(`Invalid user ID: ${userId}`);
		}

		// Store selection in localStorage for persistence
		localStorage.setItem(this.STORAGE_KEY, userId);
		
		// Trigger a page reload to ensure all components get the new user context
		// TODO: When implementing Supabase, consider using reactive stores instead of reload
		window.location.reload();
	}

	/**
	 * Check if user switching is supported (always true for dev provider)
	 */
	supportsUserSwitching(): boolean {
		return dev;
	}

	/**
	 * Get the currently selected user ID from localStorage
	 */
	private getSelectedUserId(): string {
		if (typeof localStorage === 'undefined') return 'monalisa';
		return localStorage.getItem(this.STORAGE_KEY) || 'monalisa';
	}

	/**
	 * Get user info from user ID
	 */
	private getUserFromId(userId: string): User | null {
		const devUser = (DEV_USERS as readonly any[]).find(u => u.id === userId);
		if (!devUser) return null;

		return {
			id: devUser.id,
			email: devUser.email,
			subscriptionTier: devUser.subscriptionTier,
			provider: 'dev',
			subject: devUser.id
		};
	}

	/**
	 * Get default user (monalisa)
	 */
	private getDefaultUser(): User {
		return {
			id: 'monalisa',
			email: 'monalisa@birki.io', 
			subscriptionTier: 'pro',
			provider: 'dev',
			subject: 'monalisa'
		};
	}
}