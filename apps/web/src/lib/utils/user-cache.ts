/**
 * User profile cache utility to reduce database queries
 * Caches user profile data for a short period to avoid redundant queries
 */

import type { User } from '$lib/auth/provider';

interface CachedUserProfile {
  user: User;
  timestamp: number;
  expiryTime: number;
}

class UserProfileCache {
  private cache = new Map<string, CachedUserProfile>();
  private readonly CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

  /**
   * Set user profile in cache
   */
  set(userId: string, user: User): void {
    const now = Date.now();
    this.cache.set(userId, {
      user: { ...user }, // Shallow clone to prevent mutations
      timestamp: now,
      expiryTime: now + this.CACHE_DURATION
    });
  }

  /**
   * Get user profile from cache if still valid
   */
  get(userId: string): User | null {
    const cached = this.cache.get(userId);
    if (!cached) return null;

    const now = Date.now();
    if (now >= cached.expiryTime) {
      this.cache.delete(userId);
      return null;
    }

    return { ...cached.user }; // Return shallow clone
  }

  /**
   * Check if user profile is cached and valid
   */
  has(userId: string): boolean {
    return this.get(userId) !== null;
  }

  /**
   * Remove user profile from cache
   */
  delete(userId: string): void {
    this.cache.delete(userId);
  }

  /**
   * Clear all cached profiles
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * Clean up expired entries
   */
  cleanup(): void {
    const now = Date.now();
    for (const [userId, cached] of this.cache.entries()) {
      if (now >= cached.expiryTime) {
        this.cache.delete(userId);
      }
    }
  }

  /**
   * Get cache statistics (useful for debugging)
   */
  getStats(): { size: number; entries: Array<{ userId: string; age: number; expiresIn: number }> } {
    const now = Date.now();
    const entries = Array.from(this.cache.entries()).map(([userId, cached]) => ({
      userId,
      age: now - cached.timestamp,
      expiresIn: Math.max(0, cached.expiryTime - now)
    }));

    return {
      size: this.cache.size,
      entries
    };
  }
}

// Export singleton instance
export const userProfileCache = new UserProfileCache();

// Set up periodic cleanup (only in browser environment)
if (typeof window !== 'undefined') {
  setInterval(() => {
    userProfileCache.cleanup();
  }, 10 * 60 * 1000); // Cleanup every 10 minutes
}