export const DEFAULT_REDIRECT_PATH = "/record"

// Ensure redirects are internal paths only
export function sanitizeRedirect(
  path: string,
  fallback: string = DEFAULT_REDIRECT_PATH,
): string {
  if (!path) return fallback
  // Only allow internal paths starting with '/' but not '//' (protocol-relative URLs)
  if (path.startsWith("/") && !path.startsWith("//")) return path
  return fallback
}

// Get the redirect param from a URL, falling back to the default and sanitizing
export function getRedirectParam(
  url: URL,
  fallback: string = DEFAULT_REDIRECT_PATH,
): string {
  const value = url.searchParams.get("redirect") || ""
  return sanitizeRedirect(value, fallback)
}
