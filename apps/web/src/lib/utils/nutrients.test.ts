import { describe, it, expect } from 'vitest'
import { RESTRICTED_NUTRIENTS, isRestrictedNutrient, type RestrictedNutrient } from './nutrients'

describe('nutrients utility', () => {
  describe('RESTRICTED_NUTRIENTS', () => {
    it('should contain all expected restricted nutrients', () => {
      const expected = [
        'added_sugars_g',
        'saturated_fat_g',
        'trans_fat_g',
        'cholesterol_mg',
        'alcohol_g',
        'caffeine_mg'
      ]
      
      expect(RESTRICTED_NUTRIENTS).toEqual(expected)
      expect(RESTRICTED_NUTRIENTS).toHaveLength(6)
    })

    it('should be a const array', () => {
      // TypeScript prevents mutation at compile time with 'as const'
      // This test just verifies the array is properly defined
      expect(Array.isArray(RESTRICTED_NUTRIENTS)).toBe(true)
      expect(RESTRICTED_NUTRIENTS.length).toBeGreaterThan(0)
    })
  })

  describe('isRestrictedNutrient', () => {
    it('should return true for restricted nutrients', () => {
      expect(isRestrictedNutrient('added_sugars_g')).toBe(true)
      expect(isRestrictedNutrient('saturated_fat_g')).toBe(true)
      expect(isRestrictedNutrient('trans_fat_g')).toBe(true)
      expect(isRestrictedNutrient('cholesterol_mg')).toBe(true)
      expect(isRestrictedNutrient('alcohol_g')).toBe(true)
      expect(isRestrictedNutrient('caffeine_mg')).toBe(true)
    })

    it('should return false for non-restricted nutrients', () => {
      expect(isRestrictedNutrient('sodium_mg')).toBe(false)
      expect(isRestrictedNutrient('calories')).toBe(false)
      expect(isRestrictedNutrient('protein_g')).toBe(false)
      expect(isRestrictedNutrient('vitamin_c_mg')).toBe(false)
      expect(isRestrictedNutrient('invalid_nutrient')).toBe(false)
    })

    it('should handle empty strings and undefined gracefully', () => {
      expect(isRestrictedNutrient('')).toBe(false)
      expect(isRestrictedNutrient('undefined')).toBe(false)
    })
  })

  describe('RestrictedNutrient type', () => {
    it('should accept valid restricted nutrient strings', () => {
      // This tests compile-time type checking
      const validNutrient: RestrictedNutrient = 'added_sugars_g'
      expect(validNutrient).toBe('added_sugars_g')
    })
  })
})
