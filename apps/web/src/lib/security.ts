/**
 * Security utilities for safe URL handling and validation
 */

/**
 * Sanitizes a return URL to prevent open redirect attacks.
 * Only allows same-origin relative paths that start with '/'.
 * 
 * @param returnUrl - The URL to validate and sanitize
 * @param defaultUrl - The default URL to return if the input is invalid (defaults to '/')
 * @returns A safe, validated relative path
 */
export function sanitizeReturnUrl(returnUrl: string | null, defaultUrl = '/'): string {
  // Handle null, undefined, or empty strings
  if (!returnUrl || typeof returnUrl !== 'string') {
    return defaultUrl;
  }

  // Trim whitespace
  const trimmed = returnUrl.trim();
  if (!trimmed) {
    return defaultUrl;
  }

  try {
    // If it's an absolute URL, parse it to check if it's same-origin
    if (trimmed.startsWith('http://') || trimmed.startsWith('https://') || trimmed.startsWith('//')) {
      const url = new URL(trimmed);
      
      // Only allow same-origin URLs by checking if we're in a browser context
      // and comparing origins
      if (typeof window !== 'undefined') {
        if (url.origin !== window.location.origin) {
          return defaultUrl; // External URL - reject
        }
        // Return just the pathname + search + hash for same-origin URLs
        return url.pathname + url.search + url.hash;
      } else {
        // Server-side rendering - be conservative and reject absolute URLs
        return defaultUrl;
      }
    }

    // For relative URLs, ensure they start with '/' and don't contain '..'
    if (!trimmed.startsWith('/')) {
      return defaultUrl;
    }

    // Prevent path traversal attacks - check both raw string and URL-decoded
    if (trimmed.includes('../') || trimmed.includes('..\\')) {
      return defaultUrl;
    }
    
    // Also check for URL-encoded path traversal
    const decoded = decodeURIComponent(trimmed);
    if (decoded.includes('../') || decoded.includes('..\\')) {
      return defaultUrl;
    }

    // Additional validation - check for potentially dangerous schemes
    const lowerTrimmed = trimmed.toLowerCase();
    if (lowerTrimmed.startsWith('javascript:') || 
        lowerTrimmed.startsWith('data:') || 
        lowerTrimmed.startsWith('vbscript:') ||
        lowerTrimmed.startsWith('file:') ||
        lowerTrimmed.startsWith('ftp:')) {
      return defaultUrl;
    }

    // Validate that it looks like a reasonable path
    try {
      // Use URL constructor with a dummy base to validate the path structure
      const testUrl = new URL(trimmed, 'http://localhost');
      
      // Return the pathname + search + hash
      return testUrl.pathname + testUrl.search + testUrl.hash;
    } catch {
      // Invalid URL structure
      return defaultUrl;
    }
    
  } catch {
    // Any parsing errors - return default
    return defaultUrl;
  }
}

/**
 * Checks if a URL is safe for internal redirects
 * 
 * @param url - The URL to check
 * @returns true if the URL is safe for redirects
 */
export function isSafeRedirectUrl(url: string): boolean {
  const sanitized = sanitizeReturnUrl(url);
  return sanitized !== '/' || url === '/' || url === '';
}

/**
 * Validates if a request URL should receive authentication headers.
 * Only allows requests to the configured API base URL.
 * 
 * @param requestUrl - The URL being requested
 * @param apiBaseUrl - The trusted API base URL
 * @returns true if auth headers should be attached
 */
export function shouldAttachAuthHeader(requestUrl: string, apiBaseUrl: string): boolean {
  if (!requestUrl || !apiBaseUrl) {
    return false;
  }

  try {
    // Handle relative URLs - they should get auth headers since they're same-origin
    if (requestUrl.startsWith('/')) {
      return true;
    }

    const requestUrlObj = new URL(requestUrl);
    const apiBaseUrlObj = new URL(apiBaseUrl);
    
    // Only attach auth headers if the request is going to our trusted API
    return requestUrlObj.origin === apiBaseUrlObj.origin;
    
  } catch {
    // If URL parsing fails, be conservative and don't attach auth headers
    return false;
  }
}