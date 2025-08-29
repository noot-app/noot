import { describe, it, expect } from 'vitest';
import type { User } from './provider';

describe('AuthProvider Interface', () => {
  it('should define the correct User interface structure', () => {
    const mockUser: User = {
      id: 'test-id',
      email: 'test@example.com',
      subscriptionTier: 'free',
      provider: 'supabase',
      subject: 'test-subject'
    };

    expect(mockUser).toHaveProperty('id');
    expect(mockUser).toHaveProperty('email');
    expect(mockUser).toHaveProperty('subscriptionTier');
    expect(mockUser).toHaveProperty('provider');
    expect(mockUser).toHaveProperty('subject');
  });

  it('should accept pro subscription tier', () => {
    const proUser: User = {
      id: 'test-id',
      email: 'test@example.com',
      subscriptionTier: 'pro'
    };

    expect(proUser.subscriptionTier).toBe('pro');
  });

  it('should accept free subscription tier', () => {
    const freeUser: User = {
      id: 'test-id',
      email: 'test@example.com',
      subscriptionTier: 'free'
    };

    expect(freeUser.subscriptionTier).toBe('free');
  });

  it('should allow optional provider and subject fields', () => {
    const minimalUser: User = {
      id: 'test-id',
      email: 'test@example.com',
      subscriptionTier: 'free'
    };

    expect(minimalUser.provider).toBeUndefined();
    expect(minimalUser.subject).toBeUndefined();
  });
});
