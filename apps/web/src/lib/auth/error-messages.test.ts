import { describe, it, expect } from 'vitest';
import { authErrorMessages, getAuthErrorMessage } from './error-messages';

describe('Auth Error Messages', () => {
  describe('authErrorMessages', () => {
    it('should contain expected error codes', () => {
      const expectedCodes = [
        'email_not_confirmed',
        'invalid_credentials',
        'signup_disabled',
        'email_address_invalid',
        'password_too_short',
        'email_address_not_authorized',
        'too_many_requests',
        'user_not_found',
        'weak_password',
        'email_not_confirmed_retry'
      ];

      expectedCodes.forEach(code => {
        expect(authErrorMessages).toHaveProperty(code);
        expect(typeof authErrorMessages[code]).toBe('string');
        expect(authErrorMessages[code].length).toBeGreaterThan(0);
      });
    });

    it('should provide user-friendly error messages', () => {
      expect(authErrorMessages.invalid_credentials).toContain('Invalid email or password');
      expect(authErrorMessages.email_not_confirmed).toContain('Email not confirmed');
      expect(authErrorMessages.weak_password).toContain('Password is too weak');
    });
  });

  describe('getAuthErrorMessage', () => {
    it('should return mapped message for valid error code', () => {
      const result = getAuthErrorMessage('invalid_credentials', 'fallback message');
      expect(result).toBe(authErrorMessages.invalid_credentials);
      expect(result).toContain('Invalid email or password');
    });

    it('should return fallback message for null error code', () => {
      const fallback = 'Custom fallback message';
      const result = getAuthErrorMessage(null, fallback);
      expect(result).toBe(fallback);
    });

    it('should return fallback message for empty error code', () => {
      const fallback = 'Custom fallback message';
      const result = getAuthErrorMessage('', fallback);
      expect(result).toBe(fallback);
    });

    it('should return fallback message for unknown error code', () => {
      const fallback = 'Unknown error occurred';
      const result = getAuthErrorMessage('unknown_error_code', fallback);
      expect(result).toBe(fallback);
    });

    it('should handle undefined error code gracefully', () => {
      const fallback = 'Undefined error';
      const result = getAuthErrorMessage(undefined as any, fallback);
      expect(result).toBe(fallback);
    });

    it('should prioritize mapped messages over fallback', () => {
      const fallback = 'This should not be returned';
      const result = getAuthErrorMessage('email_not_confirmed', fallback);
      expect(result).not.toBe(fallback);
      expect(result).toBe(authErrorMessages.email_not_confirmed);
    });

    it('should handle edge case with whitespace-only error code', () => {
      const fallback = 'Whitespace fallback';
      const result = getAuthErrorMessage('   ', fallback);
      expect(result).toBe(fallback);
    });
  });
});
