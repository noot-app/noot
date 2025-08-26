import { dev } from '$app/environment';

/**
 * Parse an error from various formats into a user-friendly string
 */
export function parseErrorMessage(error: any): string {
  // If it's a plain string, return it
  if (typeof error === 'string') {
    return error;
  }

  // If it has a message property, use that
  if (error?.message && typeof error.message === 'string') {
    return error.message;
  }

  // If it's an Error object, use its message
  if (error instanceof Error) {
    return error.message;
  }

  // Try to extract meaningful error from API response
  if (error?.error && typeof error.error === 'string') {
    return error.error;
  }

  // Fallback to JSON representation
  return JSON.stringify(error);
}

/**
 * Format error message for user display with optional translations
 */
export function formatErrorForUser(rawError: any, translations?: Record<string, string>): string {
  const errorMessage = parseErrorMessage(rawError);
  
  // Default error translations
  const defaultTranslations: Record<string, string> = {
    'At least one override must be provided': 'At least one nutrition target override must be provided',
    'Invalid request': 'Please check your input and try again',
    'Unauthorized': 'You need to be logged in to perform this action',
    'Forbidden': 'You don\'t have permission to perform this action'
  };

  const allTranslations = { ...defaultTranslations, ...translations };
  
  // Check if we have a translation for this exact message
  const translation = allTranslations[errorMessage];
  if (translation) {
    return dev ? `${translation} (Dev: ${errorMessage})` : translation;
  }

  // Return original message, with dev info if in development
  return dev ? `${errorMessage} (Raw: ${JSON.stringify(rawError)})` : errorMessage;
}

/**
 * Common error handling wrapper for API calls
 */
export async function handleApiCall<T>(
  apiCall: () => Promise<T>,
  errorMessage: string = 'An error occurred'
): Promise<{ data?: T; error?: string }> {
  try {
    const data = await apiCall();
    return { data };
  } catch (error) {
    return { error: formatErrorForUser(error) };
  }
}