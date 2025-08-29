import { env } from "$env/dynamic/public";

/**
 * Get the application name from environment variables with fallback
 * @returns The app name from PUBLIC_APP_NAME environment variable or 'Noot' as default
 */
export function getAppName(): string {
  return env.PUBLIC_APP_NAME || 'Noot';
}

/**
 * Reactive app name getter for use in Svelte components
 * Returns a derived value that updates when environment changes
 */
export function createAppNameStore() {
  return {
    get appName() {
      return getAppName();
    }
  };
}