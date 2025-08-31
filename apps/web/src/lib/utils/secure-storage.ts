/**
 * Simple localStorage utilities for SvelteKit
 */

/**
 * Safely get a value from localStorage
 */
export function getStorageItem(key: string): string | null {
  if (typeof window === "undefined") return null

  try {
    return localStorage.getItem(key)
  } catch (error) {
    console.warn(`Failed to get localStorage item ${key}:`, error)
    return null
  }
}

/**
 * Safely set a value in localStorage
 */
export function setStorageItem(key: string, value: string): boolean {
  if (typeof window === "undefined") return false

  try {
    localStorage.setItem(key, value)
    return true
  } catch (error) {
    console.warn(`Failed to set localStorage item ${key}:`, error)
    return false
  }
}

/**
 * Safely remove a value from localStorage
 */
export function removeStorageItem(key: string): boolean {
  if (typeof window === "undefined") return false

  try {
    localStorage.removeItem(key)
    return true
  } catch (error) {
    console.warn(`Failed to remove localStorage item ${key}:`, error)
    return false
  }
}

/**
 * Get and parse a JSON value from localStorage safely
 */
export function getStorageJSON<T>(key: string, defaultValue: T): T {
  const value = getStorageItem(key)
  if (!value) return defaultValue

  try {
    return JSON.parse(value) as T
  } catch (error) {
    console.warn(`Failed to parse JSON from localStorage ${key}:`, error)
    return defaultValue
  }
}

/**
 * Set a JSON value to localStorage safely
 */
export function setStorageJSON<T>(key: string, value: T): boolean {
  try {
    const jsonValue = JSON.stringify(value)
    return setStorageItem(key, jsonValue)
  } catch (error) {
    console.warn(`Failed to stringify value for localStorage ${key}:`, error)
    return false
  }
}
