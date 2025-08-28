/**
 * Secure localStorage utilities with validation and sanitization
 */

import { browser } from '$app/environment';

// Allowed keys for localStorage - whitelist approach
const ALLOWED_KEYS = [
  'noot-units',
  'noot-show-age',
] as const;

type AllowedKey = typeof ALLOWED_KEYS[number];

/**
 * Safely get a value from localStorage with validation
 */
export function getSecureItem(key: AllowedKey): string | null {
  if (!browser) return null;
  
  if (!ALLOWED_KEYS.includes(key)) {
    console.warn(`Attempted to access unauthorized localStorage key: ${key}`);
    return null;
  }
  
  try {
    const value = localStorage.getItem(key);
    return value;
  } catch (error) {
    console.warn(`Failed to get localStorage item ${key}:`, error);
    return null;
  }
}

/**
 * Safely set a value in localStorage with validation
 */
export function setSecureItem(key: AllowedKey, value: string): boolean {
  if (!browser) return false;
  
  if (!ALLOWED_KEYS.includes(key)) {
    console.warn(`Attempted to set unauthorized localStorage key: ${key}`);
    return false;
  }
  
  // Basic sanitization - prevent XSS in stored values
  const sanitizedValue = sanitizeStorageValue(value);
  
  try {
    localStorage.setItem(key, sanitizedValue);
    return true;
  } catch (error) {
    console.warn(`Failed to set localStorage item ${key}:`, error);
    return false;
  }
}

/**
 * Safely remove a value from localStorage
 */
export function removeSecureItem(key: AllowedKey): boolean {
  if (!browser) return false;
  
  if (!ALLOWED_KEYS.includes(key)) {
    console.warn(`Attempted to remove unauthorized localStorage key: ${key}`);
    return false;
  }
  
  try {
    localStorage.removeItem(key);
    return true;
  } catch (error) {
    console.warn(`Failed to remove localStorage item ${key}:`, error);
    return false;
  }
}

/**
 * Get and parse a JSON value from localStorage safely
 */
export function getSecureJSON<T>(key: AllowedKey, defaultValue: T): T {
  const value = getSecureItem(key);
  if (!value) return defaultValue;
  
  try {
    return JSON.parse(value) as T;
  } catch (error) {
    console.warn(`Failed to parse JSON from localStorage ${key}:`, error);
    return defaultValue;
  }
}

/**
 * Set a JSON value to localStorage safely
 */
export function setSecureJSON<T>(key: AllowedKey, value: T): boolean {
  try {
    const jsonValue = JSON.stringify(value);
    return setSecureItem(key, jsonValue);
  } catch (error) {
    console.warn(`Failed to stringify value for localStorage ${key}:`, error);
    return false;
  }
}

/**
 * Basic sanitization for localStorage values
 * Prevents basic XSS attempts in stored data
 */
function sanitizeStorageValue(value: string): string {
  // Remove any script tags or javascript: URLs
  return value
    .replace(/<script[^>]*>.*?<\/script>/gi, '')
    .replace(/javascript:/gi, '')
    .replace(/on\w+\s*=/gi, ''); // Remove event handlers like onclick=
}