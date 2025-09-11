import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  getStorageItem,
  setStorageItem,
  removeStorageItem,
  getStorageJSON,
  setStorageJSON
} from './secure-storage'

// Mock localStorage
const mockLocalStorage = (() => {
  let store: Record<string, string> = {}
  
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    })
  }
})()

// Mock console.warn to test error handling
const mockConsoleWarn = vi.fn()

describe('Secure Storage Utilities', () => {
  beforeEach(() => {
    // Reset mocks
    mockLocalStorage.clear()
    mockConsoleWarn.mockClear()
    
    // Set up global mocks
    Object.defineProperty(global, 'localStorage', {
      value: mockLocalStorage,
      writable: true
    })
    Object.defineProperty(global, 'console', {
      value: { ...console, warn: mockConsoleWarn },
      writable: true
    })
    Object.defineProperty(global, 'window', {
      value: global,
      writable: true
    })
  })

  describe('getStorageItem', () => {
    it('should get item from localStorage', () => {
      mockLocalStorage.setItem('test-key', 'test-value')
      
      const result = getStorageItem('test-key')
      
      expect(result).toBe('test-value')
      expect(mockLocalStorage.getItem).toHaveBeenCalledWith('test-key')
    })

    it('should return null for non-existent key', () => {
      const result = getStorageItem('non-existent')
      
      expect(result).toBe(null)
      expect(mockLocalStorage.getItem).toHaveBeenCalledWith('non-existent')
    })

    it('should return null in SSR environment', () => {
      // Mock server-side rendering (no window) by temporarily setting window to undefined
      Object.defineProperty(global, 'window', {
        value: undefined,
        writable: true,
        configurable: true
      })
      
      const result = getStorageItem('test-key')
      
      expect(result).toBe(null)
      
      // Restore window
      Object.defineProperty(global, 'window', {
        value: global,
        writable: true,
        configurable: true
      })
    })

    it('should handle localStorage errors gracefully', () => {
      mockLocalStorage.getItem.mockImplementationOnce(() => {
        throw new Error('Storage access denied')
      })
      
      const result = getStorageItem('error-key')
      
      expect(result).toBe(null)
      expect(mockConsoleWarn).toHaveBeenCalledWith(
        'Failed to get localStorage item error-key:',
        expect.any(Error)
      )
    })
  })

  describe('setStorageItem', () => {
    it('should set item in localStorage', () => {
      const result = setStorageItem('test-key', 'test-value')
      
      expect(result).toBe(true)
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('test-key', 'test-value')
    })

    it('should return false in SSR environment', () => {
      Object.defineProperty(global, 'window', {
        value: undefined,
        writable: true,
        configurable: true
      })
      
      const result = setStorageItem('test-key', 'test-value')
      
      expect(result).toBe(false)
      
      // Restore window
      Object.defineProperty(global, 'window', {
        value: global,
        writable: true,
        configurable: true
      })
    })

    it('should handle localStorage errors gracefully', () => {
      mockLocalStorage.setItem.mockImplementationOnce(() => {
        throw new Error('Storage quota exceeded')
      })
      
      const result = setStorageItem('error-key', 'error-value')
      
      expect(result).toBe(false)
      expect(mockConsoleWarn).toHaveBeenCalledWith(
        'Failed to set localStorage item error-key:',
        expect.any(Error)
      )
    })
  })

  describe('removeStorageItem', () => {
    it('should remove item from localStorage', () => {
      const result = removeStorageItem('test-key')
      
      expect(result).toBe(true)
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('test-key')
    })

    it('should return false in SSR environment', () => {
      Object.defineProperty(global, 'window', {
        value: undefined,
        writable: true,
        configurable: true
      })
      
      const result = removeStorageItem('test-key')
      
      expect(result).toBe(false)
      
      // Restore window
      Object.defineProperty(global, 'window', {
        value: global,
        writable: true,
        configurable: true
      })
    })

    it('should handle localStorage errors gracefully', () => {
      mockLocalStorage.removeItem.mockImplementationOnce(() => {
        throw new Error('Storage access denied')
      })
      
      const result = removeStorageItem('error-key')
      
      expect(result).toBe(false)
      expect(mockConsoleWarn).toHaveBeenCalledWith(
        'Failed to remove localStorage item error-key:',
        expect.any(Error)
      )
    })
  })

  describe('getStorageJSON', () => {
    it('should parse JSON from localStorage', () => {
      const testData = { name: 'John', age: 30 }
      mockLocalStorage.setItem('json-key', JSON.stringify(testData))
      
      const result = getStorageJSON('json-key', {})
      
      expect(result).toEqual(testData)
    })

    it('should return default value for non-existent key', () => {
      const defaultValue = { default: true }
      
      const result = getStorageJSON('non-existent', defaultValue)
      
      expect(result).toEqual(defaultValue)
    })

    it('should return default value for invalid JSON', () => {
      mockLocalStorage.setItem('invalid-json', 'not-valid-json')
      const defaultValue = { error: true }
      
      const result = getStorageJSON('invalid-json', defaultValue)
      
      expect(result).toEqual(defaultValue)
      expect(mockConsoleWarn).toHaveBeenCalledWith(
        'Failed to parse JSON from localStorage invalid-json:',
        expect.any(Error)
      )
    })

    it('should handle complex nested objects', () => {
      const complexData = {
        user: { id: 123, name: 'Jane' },
        preferences: { theme: 'dark', notifications: true },
        lastLogin: '2023-01-01T00:00:00Z'
      }
      mockLocalStorage.setItem('complex-json', JSON.stringify(complexData))
      
      const result = getStorageJSON('complex-json', {})
      
      expect(result).toEqual(complexData)
    })

    it('should work with arrays', () => {
      const arrayData = [1, 2, 3, { name: 'test' }]
      mockLocalStorage.setItem('array-json', JSON.stringify(arrayData))
      
      const result = getStorageJSON('array-json', [])
      
      expect(result).toEqual(arrayData)
    })
  })

  describe('setStorageJSON', () => {
    it('should stringify and store JSON', () => {
      const testData = { name: 'Alice', score: 95 }
      
      const result = setStorageJSON('json-key', testData)
      
      expect(result).toBe(true)
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'json-key',
        JSON.stringify(testData)
      )
    })

    it('should handle primitive values', () => {
      const result1 = setStorageJSON('string-key', 'hello')
      const result2 = setStorageJSON('number-key', 42)
      const result3 = setStorageJSON('boolean-key', true)
      
      expect(result1).toBe(true)
      expect(result2).toBe(true)
      expect(result3).toBe(true)
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('string-key', '"hello"')
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('number-key', '42')
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('boolean-key', 'true')
    })

    it('should handle arrays', () => {
      const arrayData = ['a', 'b', { c: 3 }]
      
      const result = setStorageJSON('array-key', arrayData)
      
      expect(result).toBe(true)
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'array-key',
        JSON.stringify(arrayData)
      )
    })

    it('should handle JSON stringify errors', () => {
      // Create an object that can't be stringified (circular reference)
      const circularObj: any = { name: 'test' }
      circularObj.self = circularObj
      
      const result = setStorageJSON('circular-key', circularObj)
      
      expect(result).toBe(false)
      expect(mockConsoleWarn).toHaveBeenCalledWith(
        'Failed to stringify value for localStorage circular-key:',
        expect.any(Error)
      )
    })

    it('should propagate localStorage set errors', () => {
      mockLocalStorage.setItem.mockImplementationOnce(() => {
        throw new Error('Storage quota exceeded')
      })
      
      const result = setStorageJSON('error-key', { data: 'test' })
      
      expect(result).toBe(false)
    })
  })

  describe('Integration tests', () => {
    it('should handle full get/set/remove cycle', () => {
      const testData = { user: 'john', preferences: { theme: 'dark' } }
      
      // Set
      const setResult = setStorageJSON('integration-test', testData)
      expect(setResult).toBe(true)
      
      // Get
      const getResult = getStorageJSON('integration-test', {})
      expect(getResult).toEqual(testData)
      
      // Remove
      const removeResult = removeStorageItem('integration-test')
      expect(removeResult).toBe(true)
      
      // Verify removal
      const afterRemove = getStorageJSON('integration-test', { default: true })
      expect(afterRemove).toEqual({ default: true })
    })
  })
})