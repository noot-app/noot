/**
 * Error handling utilities for authentication
 * Provides sanitized error messages and secure error handling patterns
 */

/**
 * Sanitize error messages to prevent information leakage in production
 */
export function sanitizeError(error: unknown): { message: string; code?: string } {
  // In development, return more detailed error information
  if (import.meta.env.DEV) {
    if (error && typeof error === 'object') {
      const errorObj = error as { message?: string; code?: string; details?: string };
      return {
        message: errorObj.message || 'An error occurred',
        code: errorObj.code
      };
    }
    return {
      message: String(error) || 'An unknown error occurred'
    };
  }

  // In production, return generic messages
  if (error && typeof error === 'object') {
    const errorObj = error as { message?: string; code?: string };
    
    // Map specific error codes to safe messages
    const safeErrorMessages: Record<string, string> = {
      'invalid_credentials': 'Invalid email or password',
      'email_not_confirmed': 'Please confirm your email address',
      'signup_disabled': 'Account creation is currently unavailable',
      'too_many_requests': 'Too many attempts. Please try again later',
      'weak_password': 'Password does not meet requirements',
      'email_address_invalid': 'Please enter a valid email address',
      'user_not_found': 'Account not found',
      'network_error': 'Network connection error',
      'timeout': 'Request timed out'
    };

    if (errorObj.code && safeErrorMessages[errorObj.code]) {
      return {
        message: safeErrorMessages[errorObj.code],
        code: errorObj.code
      };
    }
  }

  // Generic fallback for production
  return {
    message: 'An error occurred. Please try again.'
  };
}

/**
 * Check if an error indicates a temporary/retryable issue
 */
export function isRetryableError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false;
  
  const errorObj = error as { code?: string; message?: string };
  const retryableCodes = [
    'network_error',
    'timeout',
    'temporary_unavailable',
    'rate_limited',
    'server_error'
  ];

  return retryableCodes.includes(errorObj.code || '') ||
         (errorObj.message || '').toLowerCase().includes('network') ||
         (errorObj.message || '').toLowerCase().includes('timeout');
}

/**
 * Auth-specific error handling
 */
export class AuthError extends Error {
  public code?: string;
  public retryable: boolean;

  constructor(message: string, code?: string, retryable = false) {
    super(message);
    this.name = 'AuthError';
    this.code = code;
    this.retryable = retryable;
  }

  static fromUnknown(error: unknown): AuthError {
    const sanitized = sanitizeError(error);
    const retryable = isRetryableError(error);
    return new AuthError(sanitized.message, sanitized.code, retryable);
  }
}

/**
 * API-specific error handling  
 */
export class ApiError extends Error {
  public status?: number;
  public code?: string;
  public retryable: boolean;

  constructor(message: string, status?: number, code?: string, retryable = false) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.retryable = retryable;
  }

  static fromResponse(response: Response, error?: unknown): ApiError {
    const sanitized = sanitizeError(error);
    const retryable = response.status >= 500 || response.status === 429;
    
    return new ApiError(
      sanitized.message || `Request failed with status ${response.status}`,
      response.status,
      sanitized.code,
      retryable
    );
  }
}