/**
 * Mapping of Supabase auth error codes to user-friendly error messages
 */
export const authErrorMessages: Record<string, string> = {
  email_not_confirmed: 'Email not confirmed. Please check your email for the confirmation link.',
  invalid_credentials: 'Invalid email or password. Please try again.',
  signup_disabled: 'New account creation is currently disabled. Please contact support.',
  email_address_invalid: 'Please enter a valid email address.',
  password_too_short: 'Password must be at least 8 characters long.',
  email_address_not_authorized: 'This email address is not authorized to sign up.',
  too_many_requests: 'Too many login attempts. Please wait a moment and try again.',
  user_not_found: 'No account found with this email address.',
  weak_password: 'Password is too weak. Please choose a stronger password.',
  email_not_confirmed_retry: 'Email not confirmed. Please check your email for a confirmation and then try again.',
  // OAuth generic errors
  missing_oauth_code: 'Authentication failed. Please try again.',
  oauth_exchange_failed: 'Authentication failed during sign-in. Please try again.',
  oauth_callback_failed: 'An error occurred during authentication. Please try again.',
  server_error: 'Authentication failed. Please try again.',
  unexpected_failure: 'Authentication failed. Please try again.',
};

/**
 * Get a user-friendly error message for an auth error code
 * @param errorCode - The error code from Supabase
 * @param fallbackMessage - The original error message to use if no mapping exists
 * @returns User-friendly error message
 */
export function getAuthErrorMessage(errorCode: string | null, fallbackMessage: string): string {
  if (!errorCode || !authErrorMessages[errorCode]) {
    return fallbackMessage;
  }
  return authErrorMessages[errorCode];
}
